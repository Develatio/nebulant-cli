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

package base

import (
	"io"

	"github.com/develatio/nebulant-cli/blueprint"
)

type ActionContextRunStatus int

type ContextType int

const (
	ContextTypeRegular ContextType = iota // one parent, one child
	ContextTypeThread                     // one parent, many children
	ContextTypeJoin                       // many parents, one child
)

const (
	RunStatusReady ActionContextRunStatus = iota
	RunStatusArranging
	RunStatusRunning
	RunStatusDone
)

// ActionContext interface
type IActionContext interface {
	Type() ContextType
	Parent(IActionContext)
	Parents() []IActionContext
	Child(IActionContext)
	Children() []IActionContext
	IsThreadPoint() bool
	IsJoinPoint() bool
	GetAction() *blueprint.Action
	SetStore(IStore)
	GetStore() IStore
	// SetProvider(IProvider)
	// GetProvider() IProvider
	Done() <-chan struct{}
	WithCancelCause()
	WithEventListener(*EventListener)
	EventListener() *EventListener
	Cancel(error)
	WithRunFunc(func() (*ActionOutput, error))
	RunAction() (*ActionOutput, error)
	SetRunStatus(ActionContextRunStatus)
	GetRunStatus() ActionContextRunStatus
	//
	GetSluvaFD() io.ReadWriteCloser
	GetMustarFD() io.ReadWriteCloser
	//
	// the func that runs on DebugInit call
	WithDebugInitFunc(func())
	// returns the fd in wich the action can
	// read/write to interact with  user,
	// commonly the sluva FD of vpty
	DebugInit()
}

// IActor interface
type IActor interface {
	RunAction(action *blueprint.Action) (*ActionOutput, error)
}

// Actor struct
type Actor struct {
	Provider IProvider
}

// ActionOutput struct
type ActionOutput struct {
	Action  *blueprint.Action
	Records []*StorageRecord
}

// NewActionOutput func.
func NewActionOutput(action *blueprint.Action, storageRecordValue interface{}, storageRecordValueID *string) *ActionOutput {
	aout := &ActionOutput{
		Action: action,
	}
	record := &StorageRecord{
		Action:    action,
		RawSource: storageRecordValue,
		Value:     storageRecordValue,
		// Recursive reference. Inside aout there is also
		// reference to this StorageRecord object.
		Aout: aout,
	}
	if storageRecordValueID != nil {
		record.ValueID = *storageRecordValueID
	}
	if action.Output != nil {
		record.RefName = *action.Output
	}
	aout.Records = append(aout.Records, record)
	return aout
}
