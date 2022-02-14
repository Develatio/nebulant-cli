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
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/providers/aws/actors"
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
		return serr
	}

	// Save session into store, this struct and his values are ephemeral
	p.store.SetPrivateVar("awsSess", sess)

	// Check that the credentials have been provided. Here its validity
	// is not checked, only its existence.
	credentials, err := sess.Config.Credentials.Get()
	if err != nil {
		return err
	}

	p.Logger.LogInfo("AWS: Using access key id: " + credentials.AccessKeyID[:3] + "..." + credentials.AccessKeyID[len(credentials.AccessKeyID)-3:])

	// All credential parameters has been provided, but not validated
	return nil
}
