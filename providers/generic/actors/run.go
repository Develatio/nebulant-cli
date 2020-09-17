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

package actors

// Considerations:
// - Only one instance of runActor per script or cmd. Keep in mind that for each
// execution there must be an output and it must be stored, so the functionality
// of executing multiple scripts with an instance of runActor should not be
// implemented.
//

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/develatio/nebulant-cli/base"
)

type runScriptParameters struct {
	Target *string `json:"target" validate:"required"`
	// Unnused:
	// Username       *string `json:"username", validate:"required"`
	// PrivateKeyPath *string `json:"keyfile"`
	// Password       *string `json:"password"`
	// Port           *string `json:"port"`
	// ScriptPath     *string `json:"scriptPath"`
	// ScriptText     *string `json:"script"`
	// Command        *string `json:"command"`
}

func RunScript(ctx *ActionContext) (*base.ActionOutput, error) {
	p := &runScriptParameters{}
	jsonErr := json.Unmarshal(ctx.Action.Parameters, p)
	if jsonErr != nil {
		return nil, jsonErr
	}

	if p.Target == nil {
		ctx.Logger.LogDebug("Running remote script")
		return nil, fmt.Errorf("target cannot be empty")
	}

	if strings.ToLower(*p.Target) == "local" {
		ctx.Logger.LogDebug("Running local script")
		return RunLocalScript(ctx)
	}
	return RunRemoteScript(ctx)
}
