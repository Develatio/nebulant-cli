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

package aws

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/providers/aws/actors"
	generic_actors "github.com/develatio/nebulant-cli/providers/generic/actors"
)

func ActionValidator(action *blueprint.Action) error {
	if action.Provider != "aws" {
		return nil
	}
	al, exists := actors.ActionFuncMap[action.ActionName]
	if !exists {
		return fmt.Errorf("aws: invalid action name " + action.ActionName)
	}
	if al.N == actors.NextOK && len(action.NextAction.NextKo) > 0 {
		return fmt.Errorf("generic: action " + action.ActionName + " has no KO port")
	}

	if al.N == actors.NextKO && len(action.NextAction.NextOk) > 0 {
		return fmt.Errorf("generic: action " + action.ActionName + " has no OK port")
	}

	ac := &actors.ActionContext{
		Rehearsal: true,
		Action:    action,
	}
	_, err := al.F(ac)
	if err != nil {
		return err
	}

	return nil
}

// New var
var New base.ProviderInitFunc = func(store base.IStore) (base.IProvider, error) {
	prov := &Provider{
		store:  store,
		Logger: store.GetLogger(),
	}
	return prov, nil
}

// Provider struct
type Provider struct {
	store  base.IStore
	Logger base.ILogger
	mu     sync.Mutex
}

// DumpPrivateVars func
func (p *Provider) DumpPrivateVars(freshStore base.IStore) {
	sess := p.store.GetPrivateVar("awsSess")
	if sess != nil {
		newSess := sess.(*session.Session).Copy()
		freshStore.SetPrivateVar("awsSess", newSess)
	}
}

// HandleAction func
func (p *Provider) HandleAction(action *blueprint.Action) (*base.ActionOutput, error) {
	p.Logger.LogDebug("AWS: Received action " + action.ActionName)
	err := p.touchSession()
	if err != nil {
		return nil, err
	}

	if al, exists := actors.ActionFuncMap[action.ActionName]; exists {
		sess := p.store.GetPrivateVar("awsSess").(*session.Session)
		return al.F(actors.NewActionContext(sess, action, p.store, p.Logger))
	}
	return nil, fmt.Errorf("AWS: Unknown action: " + action.ActionName)
}

func (p *Provider) OnActionErrorHook(aout *base.ActionOutput) ([]*blueprint.Action, error) {
	if aerr, ok := aout.Records[0].Error.(awserr.Error); ok {
		// TODO: skip retry on obvious user err like *.Malformed?
		if reqErr, ok := aout.Records[0].Error.(awserr.Error).(awserr.RequestFailure); ok {
			p.Logger.LogDebug(fmt.Sprintf("AWS: Retry Hook. AWS Service errCode:%v - StatusCode:%v - RequestID:%v", aerr.Code(), reqErr.StatusCode(), reqErr.RequestID()))
		}
	}
	retries_count := 0
	retries_count_id := fmt.Sprintf("retries_count_%v", aout.Action.ActionID)
	retries_count_interface := p.store.GetPrivateVar(retries_count_id)
	if retries_count_interface != nil {
		retries_count = retries_count_interface.(int)
	}
	if retries_count < *aout.Action.MaxRetries {
		if retries_count > *aout.Action.MaxRetries || (retries_count > 0 && aout.Action.NextAction.NextKoLoop) {
			p.Logger.LogDebug("Ending retry process...")
			// retry process end, reset retry status
			p.store.SetPrivateVar(retries_count_id, 0)
			// returning nil, nil to stop hook callback
			return nil, nil
		}
		retries_count++
		p.store.SetPrivateVar(retries_count_id, retries_count)
		seconds := int64(10.0 * math.Pow(float64(retries_count), (1.0/4.0)))
		sleep_parameters := generic_actors.SleepParameters{
			Seconds: seconds,
		}

		if aout.Action.NextAction.NextKoLoop {
			p.Logger.LogWarn(fmt.Sprintf("Action Error. Retrying after %vs. There is a loop over KO, limiting the retries to only one...", seconds))
		} else {
			p.Logger.LogWarn(fmt.Sprintf("Action Error. Retrying after %vs (Retry %d/%d)...", seconds, retries_count, *aout.Action.MaxRetries))
		}

		// TODO: move the hot-sleep-action-creation
		// to something like providers.generic.NewSleepAction()?
		// NOTE to my future self: hahaha you know this is going to
		// stay here forever hahahaha, laugh with me..., I mean with
		// you..., I mean with both..., sh*t I'm going crazy.
		param, err := json.Marshal(sleep_parameters)
		if err != nil {
			return nil, err
		}
		rand.Seed(time.Now().UnixNano())
		randIntString := fmt.Sprintf("%d", rand.Int()) //#nosec G404 -- Weak random is OK here
		actions := []*blueprint.Action{
			{
				ActionID: "internal-aws-retry-control-" + randIntString,
				SafeID:   &randIntString,
				Provider: "generic",
				NextAction: blueprint.NextAction{
					NextOk: []*blueprint.Action{aout.Action},
				},
				ActionName: "sleep",
				Parameters: param,
			},
		}
		// retry
		return actions, nil
	}
	// do not retry
	return nil, nil
}

func (p *Provider) touchSession() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.store.GetPrivateVar("awsSess") != nil {
		return nil
	}

	p.Logger.LogInfo("Initializing AWS session...")

	// Init session
	// NewSessionWithOptions + SharedConfigState + SharedConfigEnable = use
	// credentials and config from ~/.aws/config and ~/.aws/credentials
	sess, serr := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if serr != nil {
		return &base.ProviderAuthError{Err: serr}
	}

	// Save session into store, this struct and his values are ephemeral
	p.store.SetPrivateVar("awsSess", sess)

	// Check that the credentials have been provided. Here its validity
	// is not checked, only its existence.
	credentials, err := sess.Config.Credentials.Get()
	if err != nil {
		return &base.ProviderAuthError{Err: err}
	}

	p.Logger.LogInfo("AWS: Using access key id: " + credentials.AccessKeyID[:3] + "..." + credentials.AccessKeyID[len(credentials.AccessKeyID)-3:])

	// All credential parameters has been provided, but not validated
	return nil
}
