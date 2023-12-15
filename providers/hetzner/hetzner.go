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

package hetzner

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	hook_providers "github.com/develatio/nebulant-cli/hook/providers"
	"github.com/develatio/nebulant-cli/providers/hetzner/actors"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func ActionValidator(action *blueprint.Action) error {
	if action.Provider != "hetznerCloud" {
		return nil
	}
	al, exists := actors.ActionFuncMap[action.ActionName]
	if !exists {
		return fmt.Errorf("hetzner: invalid action name " + action.ActionName)
	}
	if al.N == actors.NextOK && len(action.NextAction.NextKo) > 0 {
		return fmt.Errorf("hetzner: action " + action.ActionName + " has no KO port")
	}

	if al.N == actors.NextKO && len(action.NextAction.NextOk) > 0 {
		return fmt.Errorf("hetzner: action " + action.ActionName + " has no OK port")
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
	client := p.store.GetPrivateVar("hetznerClient")
	if client != nil {
		newClient := client.(*hcloud.Client)
		freshStore.SetPrivateVar("hetznerClient", newClient)
	}
}

// HandleAction func
func (p *Provider) HandleAction(action *blueprint.Action) (*base.ActionOutput, error) {
	p.Logger.LogDebug("HETZNER: Received action " + action.ActionName)
	err := p.touchSession()
	if err != nil {
		return nil, err
	}

	if al, exists := actors.ActionFuncMap[action.ActionName]; exists {
		client := p.store.GetPrivateVar("hetznerClient").(*hcloud.Client)
		return al.F(actors.NewActionContext(client, action, p.store, p.Logger))
	}
	return nil, fmt.Errorf("HETZNER: Unknown action: " + action.ActionName)
}

// OnActionErrorHook func
func (p *Provider) OnActionErrorHook(aout *base.ActionOutput) ([]*blueprint.Action, error) {

	// retry on net err, skip others
	if _, ok := aout.Records[0].Error.(net.Error); ok {
		phcontext := &hook_providers.ProviderHookContext{
			Logger: p.Logger,
			Store:  p.store,
		}
		return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
	}
	return nil, nil
}

func (p *Provider) touchSession() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.store.GetPrivateVar("hetznerClient") != nil {
		return nil
	}

	p.Logger.LogInfo("Initializing Hetzner client...")

	hct := os.Getenv("HETZNER_CLIENT_AUTH_TOKEN")
	if len(hct) <= 0 {
		return &base.ProviderAuthError{Err: fmt.Errorf("cannot found hetzner client auth token. Please set HETZNER_CLIENT_AUTH_TOKEN env var")}
	}

	client := hcloud.NewClient(hcloud.WithToken(hct))
	p.store.SetPrivateVar("hetznerClient", client)

	// All credential parameters has been provided, but not validated
	return nil
}
