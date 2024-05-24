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
	if err := json.Unmarshal(ctx.Action.Parameters, p); err != nil {
		return nil, err
	}

	if p.Target == nil {
		return nil, fmt.Errorf("target cannot be empty")
	}

	if strings.ToLower(*p.Target) == "local" {
		if !ctx.Rehearsal {
			ctx.Logger.LogDebug("Running local script")
		}
		return RunLocalScript(ctx)
	}
	if !ctx.Rehearsal {
		ctx.Logger.LogDebug("Running remote script")
	}
	return RunRemoteScript(ctx)
}
