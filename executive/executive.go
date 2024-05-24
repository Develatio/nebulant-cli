// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
