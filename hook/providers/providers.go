// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

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

package providers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
)

type ProviderHookContext struct {
	Logger     base.ILogger
	Store      base.IStore
	MaxRetries *int
}

// DefaultOnActionErrorHook func
func DefaultOnActionErrorHook(ctx *ProviderHookContext, aout *base.ActionOutput) ([]*blueprint.Action, error) {
	skipSleep := false
	sleepIdPrefix := "internal-default-retry-control-"
	if ctx.MaxRetries == nil {
		ctx.MaxRetries = new(int)
		*ctx.MaxRetries = 5

	}

	if strings.HasPrefix(aout.Action.ActionID, sleepIdPrefix) {
		// prevent infinite deep loop
		return nil, nil
	}

	if neterr, ok := aout.Records[0].Error.(net.Error); ok {
		// Net err
		if neterr.Timeout() {
			// prevent sleep
			skipSleep = true
			ctx.Logger.LogWarn("Network error. Timeout error: sleep before retry will be skipped")
		}
	}
	retries_count := 0
	retries_count_id := fmt.Sprintf("retries_count_%v", aout.Action.ActionID)
	retries_count_interface := ctx.Store.GetPrivateVar(retries_count_id)
	if retries_count_interface != nil {
		retries_count = retries_count_interface.(int)
	}

	if aout.Action.MaxRetries != nil {
		*ctx.MaxRetries = *aout.Action.MaxRetries
	}
	if retries_count < *ctx.MaxRetries {
		retries_count++
		ctx.Store.SetPrivateVar(retries_count_id, retries_count)
		seconds := int64(1)
		if !skipSleep {
			seconds = int64(10.0 * math.Pow(float64(retries_count), (1.0/4.0)))
		}
		ctx.Logger.LogWarn(fmt.Sprintf("Action Error. Retrying after %vs (Retry %d/%d)...", seconds, retries_count, *ctx.MaxRetries))

		// TODO: move the hot-sleep-action-creation
		// to something like providers.generic.NewSleepAction()?
		// NOTE to my future self: hahaha you know this is going to
		// stay here forever hahahaha, laugh with me..., I mean with
		// you..., I mean with both..., sh*t I'm going crazy.
		rand.Seed(time.Now().UnixNano())
		randIntString := fmt.Sprintf("%d", rand.Int()) //#nosec G404 -- Weak random is OK here
		actions := []*blueprint.Action{
			{
				ActionID: sleepIdPrefix + randIntString,
				SafeID:   &randIntString,
				Provider: "generic",
				NextAction: blueprint.NextAction{
					NextOk: []*blueprint.Action{aout.Action},
				},
				ActionName: "sleep",
				Parameters: json.RawMessage([]byte(fmt.Sprintf("{\"seconds\": %v}", seconds))),
			},
		}
		// retry
		return actions, nil
	}
	// do not retry
	return nil, nil
}
