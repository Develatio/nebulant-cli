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
	"github.com/develatio/nebulant-cli/blueprint"
)

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
