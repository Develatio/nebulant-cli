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
func (p *Provider) HandleAction(actx base.IActionContext) (*base.ActionOutput, error) {
	action := actx.GetAction()
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
