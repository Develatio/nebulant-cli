// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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

package providers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"

	"github.com/develatio/nebulant-cli/base"
)

type ProviderHookContext struct {
	Logger     base.ILogger
	Store      base.IStore
	MaxRetries *int
}

// DefaultOnActionErrorHook func
func DefaultOnActionErrorHook(ctx *ProviderHookContext, aout *base.ActionOutput) ([]*base.Action, error) {
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
		// rand.Seed(time.Now().UnixNano())
		randIntString := fmt.Sprintf("%d", rand.Int()) // #nosec G404 -- Weak random is OK here
		aout.Action.RetryCount = retries_count
		actions := []*base.Action{
			{
				ActionID: sleepIdPrefix + randIntString,
				SafeID:   &randIntString,
				Provider: "generic",
				NextAction: base.NextAction{
					NextOk: []*base.Action{aout.Action},
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
