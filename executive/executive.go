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

// ExecCtrlInstructionID int
type ExecCtrlInstructionID int

const (
	// ExecStop const
	ExecStop ExecCtrlInstructionID = iota
	// ExecStart const
	ExecStart
	// ExecPause const
	ExecPause
	// ExecResume const
	ExecResume
	// ExecState const. Cast current state
	ExecState
	// ExecEmancipation const
	ExecEmancipation
)

type ExecCtrlInstruction struct {
	Instruction   ExecCtrlInstructionID
	ExecutionUUID *string
}

// ActionRunStatus int
type ActionRunStatus int

// StageReportReason int
type StageReportReason int

const (
	// StageEndByDivision const
	StageEndByDivision StageReportReason = iota
	// StageEndByJoin const
	StageEndByJoin
	// StageEndByRunDone const
	StageEndByRunDone
	// StageActionReport const
	StageActionReport
	// StagePause const
	StagePause
	// StageResume const
	StageResume
)

const (
	// ActionQueued const
	ActionQueued ActionRunStatus = iota
	// ActionRunning const
	ActionRunning
	// ActionNotRunning const
	ActionNotRunning
)

// type stageReport struct {
// 	stage           *Stage
// 	reportReason    StageReportReason
// 	actionID        string
// 	actionRunStatus ActionRunStatus
// 	// LastAction      *blueprint.Action
// 	LastActionContext base.IActionContext
// 	Next              []*blueprint.Action
// 	ExitCode          int
// 	Error             error
// 	Panic             bool
// 	PanicValue        interface{}
// 	PanicTrace        []byte
// }
