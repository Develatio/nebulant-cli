// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package executive

import (
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/storage"
	"github.com/develatio/nebulant-cli/util"
)

// ManagerStatusID int
type ManagerStatusID int

// NewManager func
func NewManager() *Manager {
	mng := &Manager{
		internalRegistry: &Registry{},
		ExternalRegistry: &Registry{},
	}
	// init vars calling reset
	mng.reset()
	return mng
}

// Manager struct
type Manager struct {
	// The uuid generated by the builder or uuid asked to backend
	ExecutionUUID *string
	IRB           *blueprint.IRBlueprint
	// Coumunicate instruction to manager
	execInstruction chan *ExecCtrlInstruction
	// Logger instance with remote execution uuid configured
	Logger *cast.Logger
	// Comunicate Action exec
	StageReport chan *stageReport
	// List of active stages
	stages map[*Stage]string
	// Place to store stages before bein pushed to run
	queue []*Stage
	// A map of threads tracked by &action -> &store
	syncThreads map[*blueprint.Action]base.IStore
	//
	internalRegistry *Registry
	ExternalRegistry *Registry
	//
	Stats *stats
	mu    sync.Mutex
}

type stats struct {
	stages  int
	actions int
}

func (m *Manager) reset() {
	m.ExecutionUUID = nil
	m.StageReport = make(chan *stageReport, 1000)
	m.execInstruction = make(chan *ExecCtrlInstruction, 10)
	m.Stats = &stats{}
	m.Logger = &cast.Logger{}
	m.stages = make(map[*Stage]string)
	m.syncThreads = make(map[*blueprint.Action]base.IStore)
	m.internalRegistry.reset()
	m.ExternalRegistry.reset()
	m.internalRegistry.Logger = m.Logger
	m.ExternalRegistry.Logger = m.Logger
}

// GetExecutionUUID func
func (m *Manager) GetExecutionUUID() *string {
	return m.ExecutionUUID
}

// GetLogger func
func (m *Manager) GetLogger() base.ILogger {
	return m.Logger
}

// PrepareIRB func
func (m *Manager) PrepareIRB(irb *blueprint.IRBlueprint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Reset all and prepare for new blueprint
	m.reset()
	m.IRB = irb
	m.ExecutionUUID = irb.ExecutionUUID
	m.internalRegistry.ExecutionUUID = irb.ExecutionUUID
	m.ExternalRegistry.ExecutionUUID = irb.ExecutionUUID
	if m.ExecutionUUID != nil {
		re := *m.ExecutionUUID
		m.Logger.ExecutionUUID = &re
	}
	m.internalRegistry.SetManagerState(cast.EventManagerPrepareBPEnd)
	m.ExternalRegistry.SetManagerState(cast.EventManagerPrepareBPEnd)
}

// RunStages func
func (m *Manager) RunStages(stages []*Stage) {
	if m.internalRegistry.IsInStop() {
		// prevent run new stages on manager stop process
		return
	}
	for _, newStage := range stages {
		stageaddr := fmt.Sprintf("%p", newStage)
		m.Logger.LogDebug("[Manager] Starting stage[" + stageaddr + "] for action " + newStage.StartAction.ActionID)
		if _, exists := m.stages[newStage]; exists {
			log.Panic("[Manager] Already running stage")
		}

		m.internalRegistry.QueueAction(newStage.StartAction.ActionID, newStage)
		m.ExternalRegistry.QueueAction(newStage.StartAction.ActionID, newStage)

		m.stages[newStage] = newStage.StartAction.ActionID
		m.Stats.stages++
		newStage.SetStageID(strconv.Itoa(m.Stats.stages))
		if m.internalRegistry.IsInPause() {
			newStage.execInstruction <- &ExecCtrlInstruction{
				Instruction: ExecPause,
			}
		} else {
			newStage.execInstruction <- &ExecCtrlInstruction{
				Instruction: ExecStart,
			}
		}

		go newStage.Init()
	}
}

func (m *Manager) EmancipateStages() {
	for stage := range m.stages {
		select {
		case stage.execInstruction <- &ExecCtrlInstruction{
			Instruction: ExecEmancipation,
		}:
		default:
			// Hey developer!,  what a wonderful day!
		}
	}
}

// Run func
func (m *Manager) Run() error {
	exit := false
	defer func() {
		cast.PushEvent(cast.EventManagerOut, m.ExecutionUUID)
		m.internalRegistry.SetManagerState(cast.EventManagerOut)
		m.ExternalRegistry.SetManagerState(cast.EventManagerOut)
		m.EmancipateStages()
		exit = true
		if r := recover(); r != nil {
			switch r.(type) {
			case *util.PanicData:
				panic(r)
			}
			panic(&util.PanicData{
				PanicValue: r,
				PanicTrace: debug.Stack(),
			})
		}
	}()
	if exit {
		return nil
	}

	m.Logger.LogDebug("[Manager] Starting...")

	cast.PushEvent(cast.EventManagerStarting, m.ExecutionUUID)
	m.internalRegistry.SetManagerState(cast.EventManagerStarting)
	m.ExternalRegistry.SetManagerState(cast.EventManagerStarting)
	if m.IRB.StartAction == nil {
		return fmt.Errorf("[Manager] First action id not found")
	}
	m.RunStages([]*Stage{NewStage(m, storage.NewStore(), m.IRB.StartAction)})
	cast.PushEvent(cast.EventManagerStarted, m.ExecutionUUID)
	m.internalRegistry.SetManagerState(cast.EventManagerStarted)
	m.ExternalRegistry.SetManagerState(cast.EventManagerStarted)
	m.Logger.LogDebug("[Manager] Ready")
	// for run stats
	startTime := time.Now()

	// Just for debug purposes
	// go func() {
	// 	for {
	// 		log.Println("len: " + fmt.Sprint(len(m.StageReport)))
	// 		time.Sleep(50 * time.Millisecond)
	// 	}
	// }()

L:
	for { // Infine loop until break L
		select { // Loop until a case ocurrs.
		case instruction := <-m.execInstruction:
			if instruction.ExecutionUUID != nil && *m.ExecutionUUID != *instruction.ExecutionUUID {
				continue
			}
			switch instruction.Instruction {
			case ExecStart:
				// m.managerStatus = ManagerStatusRunning
			case ExecStop:
				// Perform kill?
				cast.PushEvent(cast.EventManagerStopping, m.ExecutionUUID)
				m.internalRegistry.SetManagerState(cast.EventManagerStopping)
				m.ExternalRegistry.SetManagerState(cast.EventManagerStopping)
				m.ExternalRegistry.publishStatus()
			case ExecPause:
				cast.PushEvent(cast.EventManagerPausing, m.ExecutionUUID)
				m.internalRegistry.SetManagerState(cast.EventManagerPausing)
				m.ExternalRegistry.SetManagerState(cast.EventManagerPausing)
				m.ExternalRegistry.publishStatus()
			case ExecResume:
				cast.PushEvent(cast.EventManagerResuming, m.ExecutionUUID)
				m.internalRegistry.SetManagerState(cast.EventManagerResuming)
				m.ExternalRegistry.SetManagerState(cast.EventManagerResuming)
				m.ExternalRegistry.publishStatus()
			case ExecState:
				m.ExternalRegistry.publishStatus()
				continue // do not propagate state to stage
			default:
				m.Logger.LogErr("[Manager] Don't know how to handle this instruction")
			}
			// Propagate instruction
			for stage := range m.stages {
				select {
				case stage.execInstruction <- instruction:
				default:
					// Hey developer!,  what a wonderful day!
				}
			}
		case sr := <-m.StageReport:
			stage := sr.stage
			switch sr.reportReason {
			case StageActionReport:
				if sr.actionRunStatus == ActionRunning {
					// action is currently running
					m.Stats.actions++
				}
				m.internalRegistry.HandleActionReport(sr)
				m.ExternalRegistry.HandleActionReport(sr)
				m.ExternalRegistry.publishStatus()
				continue // prevent stage deletion
			case StagePause:
				allPaused := true
				for stage := range m.stages {
					// TODO: maybe datarace on write/read: check it
					if stage.stageStatus != StageStatusPaused {
						allPaused = false
						break
					}
				}
				if allPaused {
					cast.PushEvent(cast.EventManagerPaused, m.ExecutionUUID)
					m.internalRegistry.SetManagerState(cast.EventManagerPaused)
					m.ExternalRegistry.SetManagerState(cast.EventManagerPaused)
					m.ExternalRegistry.publishStatus()
				}
				continue // prevent stage deletion
			case StageResume:
				allResumed := true
				for stage := range m.stages {
					if stage.stageStatus != StageStatusRunning {
						allResumed = false
						break
					}
				}
				if allResumed {
					cast.PushEvent(cast.EventManagerStarted, m.ExecutionUUID)
					m.internalRegistry.SetManagerState(cast.EventManagerStarted)
					m.ExternalRegistry.SetManagerState(cast.EventManagerStarted)
					m.ExternalRegistry.publishStatus()
				}
				continue // prevent stage deletion
			case StageEndByDivision:
				stages := stage.Divide(sr.Next)
				m.queue = append(m.queue, stages...)
			case StageEndByJoin:
				newStore := stage.GetStore()
				action := sr.LastAction
				if existentStore, exists := m.syncThreads[action]; exists {
					m.Logger.LogDebug("[Manager] Join found. Merge store")
					existentStore.Merge(newStore)
				} else {
					m.Logger.LogDebug("[Manager] Join found")
					m.syncThreads[action] = newStore
					m.internalRegistry.AddWaitingID(action.ActionID)
					m.ExternalRegistry.AddWaitingID(action.ActionID)
				}
			case StageEndByRunDone:
				if sr.ExitCode > 0 {
					m.internalRegistry.ExitCode = sr.ExitCode
					m.ExternalRegistry.ExitCode = sr.ExitCode
				}
				if sr.Error != nil {
					extra := make(map[string]interface{})
					extra["action_id"] = sr.LastAction.ActionID
					extra["action_error"] = sr.Error.Error()
					cast.PushEventWithExtra(cast.EventActionUnCaughtKO, m.ExecutionUUID, extra)
				}
				if sr.Panic {
					m.Logger.LogDebug("[Manager] Panic in stage, exiting...")
					inst := &ExecCtrlInstruction{
						Instruction:   ExecStop,
						ExecutionUUID: m.ExecutionUUID,
					}
					for stage := range m.stages {
						select {
						case stage.execInstruction <- inst:
						default:
							// Hey developer!,  what a wonderful day!
						}
					}
					panic(&util.PanicData{
						PanicValue: sr.PanicValue,
						PanicTrace: sr.PanicTrace,
					})
				}
				m.Logger.LogDebug("[Manager] Stage done")
			}
			rpt := &stageReport{
				stage:           stage,
				actionRunStatus: ActionNotRunning,
				LastAction:      sr.LastAction,
			}
			if sr.LastAction != nil {
				lastActionID := sr.LastAction.ActionID
				m.Logger.LogDebug("[Manager] Destroying stage with last id " + lastActionID)
				rpt.actionID = lastActionID
			} else {
				m.Logger.LogDebug("[Manager] Destroying stage ...")
			}

			delete(m.stages, stage)
			m.internalRegistry.HandleActionReport(rpt)
			m.ExternalRegistry.HandleActionReport(rpt)
		default:

			// Push stages to run from queue.
			// Allow it even in pause mode. The Stages will
			// start in pause status
			if len(m.queue) > 0 && len(m.StageReport) < 500 {
				stages := m.queue
				m.RunStages(stages)
				m.queue = []*Stage{}
			}

			// Check join
			// Allow it even in pause mode. The Stages will
			// start in pause status
			m.checkSyncThreads()

			if m.internalRegistry.IsInPause() {
				// add extra sleep
				time.Sleep(1 * time.Second)
			} else {
				m.Logger.LogDebug("[Manager] " + strconv.Itoa(len(m.stages)) + " stages running")
				m.Logger.LogDebug("[Manager] " + strconv.Itoa(len(m.syncThreads)) + " stages waiting join")
			}

			time.Sleep(1 * time.Second)
			if len(m.stages) <= 0 && len(m.syncThreads) <= 0 {
				m.Logger.LogDebug("[Manager] No more stages left. Exiting Manager...")
				break L
			}
		}
	}

	m.Logger.LogInfo("[Manager] out")
	elapsedTime := time.Since(startTime).String()
	m.Logger.LogInfo("[Manager] stats: " + strconv.Itoa(m.Stats.actions) + " actions executed by " + strconv.Itoa(m.Stats.stages) + " stages in " + elapsedTime)
	if m.ExecutionUUID != nil {
		m.Logger.LogDebug("Sending EventManagerOut for UUID" + *m.ExecutionUUID)
	} else {
		m.Logger.LogDebug("Sending EventManagerOut for no UUID")
	}

	cast.PushEvent(cast.EventManagerOut, m.ExecutionUUID)
	m.internalRegistry.SetManagerState(cast.EventManagerOut)
	m.ExternalRegistry.SetManagerState(cast.EventManagerOut)
	m.Logger.LogDebug("[Manager] out")
	return nil
}

func (m *Manager) checkSyncThreads() {
	syncReady := true
	if len(m.syncThreads) > 0 {
		var readyActions []*blueprint.Action
		for aa, st := range m.syncThreads {
			syncReady = true
			wp := aa.KnowParentIDs
			if m.internalRegistry.IsRunningSomeIDonMap(wp) {
				syncReady = false
				break
			}
			for waa := range m.syncThreads {
				if _, exists := wp[waa.ActionID]; exists {
					syncReady = false
					break
				}
			}
			if syncReady {
				readyActions = append(readyActions, aa)
				var stages []*Stage
				for _, newAction := range aa.NextAction.NextOk {
					store := st.Duplicate()
					stage := NewStage(m, store, newAction)
					stages = append(stages, stage)
				}
				if stages != nil {
					m.RunStages(stages)
				}
			}
		}
		for _, action := range readyActions {
			delete(m.syncThreads, action)
			m.internalRegistry.DelWaitingID(action.ActionID)
			m.ExternalRegistry.DelWaitingID(action.ActionID)
		}
	}
}
