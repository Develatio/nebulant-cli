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

package azure

import (
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	hook_providers "github.com/develatio/nebulant-cli/hook/providers"
)

func ActionValidator(action *blueprint.Action) error {
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
}

// DumpPrivateVars func
func (p *Provider) DumpPrivateVars(freshStore base.IStore) {}

// HandleAction func
func (p *Provider) HandleAction(action *blueprint.Action) (*base.ActionOutput, error) {
	p.Logger.LogDebug("AZURE: Received action " + action.ActionName)
	return nil, nil
}

func (p *Provider) OnActionErrorHook(aout *base.ActionOutput) ([]*blueprint.Action, error) {
	phcontext := &hook_providers.ProviderHookContext{
		Logger: p.Logger,
		Store:  p.store,
	}
	return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
}
