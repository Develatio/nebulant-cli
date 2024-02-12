// // Nebulant
// // Copyright (C) 2020  Develatio Technologies S.L.

// // This program is free software: you can redistribute it and/or modify
// // it under the terms of the GNU Affero General Public License as published by
// // the Free Software Foundation, either version 3 of the License, or
// // (at your option) any later version.

// // This program is distributed in the hope that it will be useful,
// // but WITHOUT ANY WARRANTY; without even the implied warranty of
// // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// // GNU Affero General Public License for more details.

// // You should have received a copy of the GNU Affero General Public License
// // along with this program.  If not, see <https://www.gnu.org/licenses/>.

package executive

// import (
// 	"sync"

// 	"github.com/develatio/nebulant-cli/base"
// 	"github.com/develatio/nebulant-cli/cast"
// )

// // NewRegistry func
// func NewRegistry() *Registry {
// 	mng := &Registry{}
// 	mng.reset()
// 	return mng
// }

// // Registry struct
// type Registry struct {
// 	ExecutionUUID *string
// 	managerState  int
// 	runningIDs    map[string]map[*Stage]ActionRunStatus
// 	byStageID     map[*Stage]string
// 	waitingIDs    map[string]bool
// 	Logger        base.ILogger
// 	mu            sync.Mutex
// 	SavedOutputs  []*base.ActionOutput
// 	ExitCode      int
// }

// func (r *Registry) reset() {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	r.ExecutionUUID = nil
// 	r.managerState = 0
// 	r.runningIDs = make(map[string]map[*Stage]ActionRunStatus)
// 	r.byStageID = make(map[*Stage]string)
// 	r.waitingIDs = make(map[string]bool)
// }

// // SetManagerState func
// func (r *Registry) SetManagerState(state int) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	r.managerState = state
// }

// // QueueAction func
// func (r *Registry) QueueAction(actionID string, stage *Stage) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	if _, exists := r.runningIDs[actionID]; !exists {
// 		r.runningIDs[actionID] = make(map[*Stage]ActionRunStatus)
// 	}
// 	r.runningIDs[actionID][stage] = ActionQueued
// 	r.byStageID[stage] = actionID
// }

// // AddWaitingID func
// func (r *Registry) AddWaitingID(actionID string) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	r.waitingIDs[actionID] = true
// }

// // DelWaitingID func
// func (r *Registry) DelWaitingID(actionID string) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	delete(r.waitingIDs, actionID)
// }

// // IsRunningSomeIDonMap func
// func (r *Registry) IsRunningSomeIDonMap(knowParents map[string]bool) bool {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	for runningID := range r.runningIDs {
// 		if _, exists := knowParents[runningID]; exists {
// 			r.Logger.LogDebug("Know parent with id " + runningID + " is running")
// 			return true
// 		}
// 	}
// 	return false
// }

// // IsInPause func
// func (r *Registry) IsInPause() bool {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	return r.managerState == cast.EventManagerPausing || r.managerState == cast.EventManagerPaused
// }

// // IsInStop func
// func (r *Registry) IsInStop() bool {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	return r.managerState == cast.EventManagerStopping || r.managerState == cast.EventManagerOut
// }

// // HandleActionReport func
// // func (r *Registry) HandleActionReport(srr *stageReport) {
// // 	r.mu.Lock()
// // 	defer r.mu.Unlock()
// // 	if _, exists := r.runningIDs[srr.actionID]; !exists {
// // 		r.runningIDs[srr.actionID] = make(map[*Stage]ActionRunStatus)
// // 	}
// // 	if srr.actionRunStatus == ActionNotRunning {
// // 		r.Logger.LogDebug("Registry: Deleting references for ID " + srr.actionID)
// // 		delete(r.runningIDs[srr.actionID], srr.stage)
// // 		delete(r.byStageID, srr.stage)
// // 		if len(r.runningIDs[srr.actionID]) <= 0 {
// // 			delete(r.runningIDs, srr.actionID)
// // 		}
// // 		if srr.LastAction != nil && srr.LastAction.SaveRawResults {
// // 			store := srr.stage.GetStore()
// // 			output, err := store.GetActionOutputByActionID(&srr.actionID)
// // 			if err != nil {
// // 				return
// // 			}
// // 			r.SavedOutputs = append(r.SavedOutputs, output)
// // 		}
// // 	} else {
// // 		r.Logger.LogDebug("Registry: Storing action ID " + srr.actionID)
// // 		r.runningIDs[srr.actionID][srr.stage] = srr.actionRunStatus
// // 		r.byStageID[srr.stage] = srr.actionID
// // 	}
// // }
