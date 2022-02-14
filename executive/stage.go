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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/util"
)

// StageStatusID int
type StageStatusID int

const (
	// StageStatusStopped const
	StageStatusStopped StageStatusID = iota
	// StageStatusRunning const
	StageStatusRunning
	// StageStatusPaused const
	StageStatusPaused
)

// NewStage func
func NewStage(manager *Manager, store base.IStore, startAction *blueprint.Action) *Stage {
	logger := manager.GetLogger()
	store.SetLogger(logger)
	stage := &Stage{
		// On new Stage, set new Store. To get a duplicated store
		// call Duplicate method on this stage
		store:           store,
		logger:          logger,
		StartAction:     startAction,
		stageStatus:     StageStatusStopped,
		CurrentAction:   startAction,
		manager:         manager,
		execInstruction: make(chan *ExecCtrlInstruction, 10),
	}
	return stage
}

// Stage struct
type Stage struct {
	recoveringStatus bool
	panicErr         string
	store            base.IStore
	logger           base.ILogger
	manager          *Manager
	execInstruction  chan *ExecCtrlInstruction
	StartAction      *blueprint.Action
	CurrentAction    *blueprint.Action
	LastAction       *blueprint.Action
	LastActionError  error
	//
	stageStatus StageStatusID
	//
	stageID        string
	stageParentsID string
}

// Divide func
func (s *Stage) Divide(actions []*blueprint.Action) []*Stage {
	var stages []*Stage
	var parentsStringSeparator string
	if s.stageParentsID != "" {
		parentsStringSeparator = "|"
	}

	for _, action := range actions {
		store := s.store.Duplicate()
		stage := &Stage{
			store:           store,
			logger:          store.GetLogger(),
			StartAction:     action,
			stageStatus:     StageStatusStopped,
			CurrentAction:   action,
			manager:         s.manager,
			execInstruction: make(chan *ExecCtrlInstruction, 10),
			stageID:         "",
			stageParentsID:  s.stageParentsID + parentsStringSeparator + s.stageID,
		}

		stageaddr := fmt.Sprintf("%p", stage)
		s.logger.LogDebug(s.lpfx() + "New division at addr [" + stageaddr + "]")
		stages = append(stages, stage)
	}

	for _, newstage := range stages {
		stageaddr := fmt.Sprintf("%p", newstage)
		s.logger.LogDebug(s.lpfx() + "Division: returning stage[" + stageaddr + "]")
		newstage.execInstruction <- &ExecCtrlInstruction{
			Instruction: ExecStart,
		}
	}

	return stages
}

// GetStore struct
func (s *Stage) GetStore() base.IStore {
	return s.store
}

func (s *Stage) lpfx() string {
	if config.DEBUG {
		return "( Thread )-[" + s.stageParentsID + "]-(" + s.stageID + ")> "
	}
	if s.stageParentsID == "" {
		return "( Main )> "
	}
	return "( Thread )> "
}

// SetStageID func
func (s *Stage) SetStageID(stageID string) {
	s.stageID = stageID
	var parentsStringSeparator string
	if s.stageParentsID != "" {
		parentsStringSeparator = "|"
	}
	s.logger.SetThreadID(s.stageParentsID + parentsStringSeparator + stageID)
}

// GetProvider func
func (s *Stage) GetProvider(providerName string) (base.IProvider, error) {
	if s.store.ExistsProvider(providerName) {
		provider, err := s.store.GetProvider(providerName)
		if err != nil {
			return nil, err
		}
		return provider, nil
	}
	providerInitFunc, initfErr := cast.SBus.GetProviderInitFunc(providerName)
	if initfErr != nil {
		return nil, initfErr
	}
	provider, providerErr := providerInitFunc(s.store)
	if providerErr != nil {
		return nil, providerErr
	}
	s.store.StoreProvider(providerName, provider)
	return provider, nil
}

func (s *Stage) chanStageReport(sr *stageReport) {
	if s.manager == nil {
		return
	}
	sr.stage = s
	sr.LastAction = s.LastAction
	for i := 0; i <= 10; i++ {
		select {
		case s.manager.StageReport <- sr:
			return
		default:
			for {
				if len(s.manager.StageReport) < 500 {
					break
				}
				time.Sleep(100 * time.Microsecond)
			}
		}
	}
	// this is really bad. A not resolved flood means broken
	// execution for some unknown reason.
	panic(s.lpfx() + "Cannot report to manager due to flood messages. Please reduce the amount of actions.")
}

func (s *Stage) setNextAction(action *blueprint.Action) {
	s.LastActionError = nil
	s.CurrentAction = action
}

// Init func
func (s *Stage) Init() {
	s.logger.LogDebug(s.lpfx() + "Stage init")
	s.stageStatus = StageStatusStopped
	exit := false

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			s.recoveringStatus = true
			sr := &stageReport{
				reportReason: StageEndByRunDone,
				ExitCode:     1,
				Panic:        true,
			}
			switch r := r.(type) {
			case *util.PanicData:
				exit = true
				sr.PanicValue = r.PanicValue
				sr.PanicTrace = r.PanicTrace
				s.chanStageReport(sr)
			default:
				s.panicErr = fmt.Sprintf("%v", r)
				s.panicErr = s.panicErr + string(debug.Stack())
				s.logger.LogDebug(s.lpfx() + " Recovering [" + s.CurrentAction.ActionName + "] [" + s.CurrentAction.ActionID + "] from panic...")
				s.execInstruction <- &ExecCtrlInstruction{
					Instruction: ExecResume,
				}
				go s.Init()
				exit = true
			}
		}
	}()
	if exit {
		return
	}
L:
	for { // Infine loop until break L
		s.logger.LogDebug(s.lpfx() + "  Loop iteration++")
		select { // Loop until a case event ocurrs.

		// Command
		case eitrn := <-s.execInstruction:
			switch eitrn.Instruction {

			// Start stage
			case ExecStart:
				s.logger.LogDebug(s.lpfx() + "Received stage start command")
				if s.CurrentAction == nil {
					s.CurrentAction = s.StartAction
				}
				s.stageStatus = StageStatusRunning

			// Stop stage
			case ExecStop:
				s.logger.LogInfo(s.lpfx() + "Stop received. [" + s.CurrentAction.ActionName + "] [" + s.CurrentAction.ActionID + "] ...")
				s.chanStageReport(&stageReport{reportReason: StageEndByRunDone})
				break L

			// Pause stage
			case ExecPause:
				s.logger.LogInfo(s.lpfx() + "Pause received. [" + s.CurrentAction.ActionName + "] [" + s.CurrentAction.ActionID + "] ...")
				s.stageStatus = StageStatusPaused
				s.chanStageReport(&stageReport{reportReason: StagePause})

			// Resume stage
			case ExecResume:
				s.logger.LogInfo(s.lpfx() + "Resume received. [" + s.CurrentAction.ActionName + "] [" + s.CurrentAction.ActionID + "] ...")
				s.stageStatus = StageStatusRunning
				s.chanStageReport(&stageReport{reportReason: StageResume})

			// Manager out, be free
			case ExecEmancipation:
				s.manager = nil

			// Unknown command
			default:
				s.logger.LogErr(s.lpfx() + "Don't know how to handle this instruction")
			}

		// Normal flow
		default:
			// After panic, try to recover
			if s.recoveringStatus {
				s.recoveringStatus = false
				if next := s.PostAction(nil, fmt.Errorf(s.panicErr)); !next {
					break L
				}
				s.panicErr = ""
			}

			// No action loaded or not started stage
			if s.stageStatus != StageStatusRunning || s.CurrentAction == nil {
				time.Sleep(2 * time.Second)
				continue
			}

			// Running action Feedback
			if s.CurrentAction.Output != nil {
				s.logger.LogInfo(s.lpfx() + "Running [" + s.CurrentAction.ActionName + " => " + *s.CurrentAction.Output + "] [" + s.CurrentAction.ActionID + "] ...")
			} else {
				s.logger.LogInfo(s.lpfx() + "Running [" + s.CurrentAction.ActionName + "] [" + s.CurrentAction.ActionID + "] ...")
			}

			// Running action comunication to manager
			s.chanStageReport(&stageReport{
				reportReason:    StageActionReport,
				actionID:        s.CurrentAction.ActionID,
				actionRunStatus: ActionRunning,
			})

			// Join threads. Stop exec and return to manager
			if s.CurrentAction.JoinThreadsPoint {
				s.LastAction = s.CurrentAction
				s.CurrentAction = nil
				s.chanStageReport(&stageReport{reportReason: StageEndByJoin})
				break L
			}

			// Get provider for current action
			provider, err := s.GetProvider(s.CurrentAction.Provider)
			if err != nil {
				// Missing provider.
				log.Panic(err.Error())
			}

			// Run action
			actionOutput, actionErr := provider.HandleAction(s.CurrentAction)
			// On HandleAction panic, the execution ends here and instead
			// calling PostAction here, a PostAction is called at start
			// of this block during recover
			if next := s.PostAction(actionOutput, actionErr); !next {
				break L
			}
		}
	}

	s.logger.LogDebug(s.lpfx() + "Stage out for id " + s.StartAction.ActionID)
}

func (s *Stage) PostAction(actionOutput *base.ActionOutput, actionErr error) bool {
	// Debug provides raw output to stdout
	if s.CurrentAction.DebugNetwork && actionOutput != nil {
		for _, record := range actionOutput.Records {
			var out bytes.Buffer
			// encode to
			enc, err := json.Marshal(record.Value)
			if err != nil {
				s.logger.LogDebug(s.lpfx() + "[" + s.LastAction.ActionID + "] " + err.Error())
			}
			if err = json.Indent(&out, enc, "", "\t"); err != nil {
				s.logger.LogDebug(s.lpfx() + "[" + s.LastAction.ActionID + "] " + err.Error())
			}
			// Do you want to use json data with s.logger?
			// use out.String() to obtain string value of Buffer bytes
			if _, err = out.WriteTo(os.Stdout); err != nil {
				s.logger.LogDebug(s.lpfx() + "[" + s.LastAction.ActionID + "] " + err.Error())
			}
		}
	}

	// Save action output into the store
	if actionOutput != nil {
		for idx := 0; idx < len(actionOutput.Records); idx++ {
			err := s.store.Insert(actionOutput.Records[idx], s.CurrentAction.Provider)
			if err != nil {
				log.Panic(err.Error())
			}
		}
	}
	if actionErr != nil { // Action error
		aerr := actionOutput
		if actionOutput == nil {
			aerr = base.NewActionOutput(s.CurrentAction, nil, nil)
			aerr.Records[0].Fail = true
			aerr.Records[0].Error = actionErr
			aerr.Records[0].Value = actionErr.Error()
		}
		for idx := 0; idx < len(aerr.Records); idx++ {
			aerr.Records[idx].Fail = true
			aerr.Records[idx].Error = actionErr
			err := s.store.Insert(aerr.Records[idx], s.CurrentAction.Provider)
			if err != nil {
				log.Panic(err.Error())
			}
		}
	}

	// Set last action executed data.
	s.LastAction = s.CurrentAction
	s.CurrentAction = nil
	s.LastActionError = actionErr
	var actions []*blueprint.Action
	if actionErr != nil { // Action error
		s.logger.LogDebug(s.lpfx() + "Stage: Action finish KO")
		actions = s.LastAction.NextAction.NextKo
	} else if s.LastAction.NextAction.ConditionalNext {
		s.logger.LogDebug(s.lpfx() + "Stage: Action finish OK")
		if actionOutput.Records[0].Value.(bool) {
			actions = s.LastAction.NextAction.NextOkTrue
		} else {
			actions = s.LastAction.NextAction.NextOkFalse
		}
	} else { // No action error
		s.logger.LogDebug(s.lpfx() + "Stage: Action finish OK")
		actions = s.LastAction.NextAction.NextOk
	}

	switch len(actions) {

	// No more actions -> instruct exit
	case 0:
		sr := &stageReport{reportReason: StageEndByRunDone}
		s.logger.LogDebug(s.lpfx() + "[" + s.LastAction.ActionID + "] Stop ---> X")
		if s.LastActionError != nil {
			sr.ExitCode = 1
			s.logger.LogErr(s.LastActionError.Error())
		}
		s.logger.LogDebug(s.lpfx() + "No more actions, stop this stage")
		s.chanStageReport(sr)
		// end
		return false

	// One action received -> continue execution
	case 1:
		s.chanStageReport(&stageReport{
			reportReason:    StageActionReport,
			actionID:        s.LastAction.ActionID,
			actionRunStatus: ActionNotRunning,
			Next:            []*blueprint.Action{actions[0]},
		})
		if s.LastActionError != nil {
			s.logger.LogWarn("Caught KO error: " + s.LastActionError.Error())
		}
		s.logger.LogDebug(s.lpfx() + "[" + s.LastAction.ActionID + "] Next ---> " + actions[0].ActionID)
		s.setNextAction(actions[0])

	// Thread init detected -> back to manager
	default:
		if s.LastActionError != nil {
			s.logger.LogWarn("Caught KO error: " + s.LastActionError.Error())
		}
		s.logger.LogDebug(s.lpfx() + "Thread detected")
		// Inform to manager to split exec
		s.chanStageReport(&stageReport{
			reportReason: StageEndByDivision,
			Next:         actions,
		})
		// end
		return false
	}

	// Continue
	return true
}
