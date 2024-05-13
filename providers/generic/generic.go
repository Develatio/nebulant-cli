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

package generic

import (
	"fmt"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	hook_providers "github.com/develatio/nebulant-cli/hook/providers"
	"github.com/develatio/nebulant-cli/providers/generic/actors"
	"github.com/develatio/nebulant-cli/util"
)

func ActionValidator(action *blueprint.Action) error {
	if action.Provider != "generic" {
		return nil
	}
	al, exists := actors.ActionFuncMap[action.ActionName]
	if !exists {
		return fmt.Errorf("generic: invalid action name " + action.ActionName)
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
		Logger:    &cast.DummyLogger{},
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
}

// DumpPrivateVars func
func (p *Provider) DumpPrivateVars(freshStore base.IStore) {}

// HandleAction func
func (p *Provider) HandleAction(actx base.IActionContext) (*base.ActionOutput, error) {
	action := actx.GetAction()
	p.Logger.LogDebug("GENERIC: Received action " + action.ActionName)
	if al, exists := actors.ActionFuncMap[action.ActionName]; exists {
		l := p.Logger.Duplicate()
		l.SetActionID(action.ActionID)
		return al.F(&actors.ActionContext{
			Action: action,
			Store:  p.store,
			Logger: l,
			Actx:   actx,
		})
	}
	return nil, fmt.Errorf("GENERIC: Unknown action: " + action.ActionName)
}

func (p *Provider) OnActionErrorHook(aout *base.ActionOutput) ([]*blueprint.Action, error) {
	al, exists := actors.ActionFuncMap[aout.Action.ActionName]
	if !exists {
		return nil, nil
	}

	// skip not retriable action
	if !al.R {
		return nil, nil
	}

	// retry on net err, skip others
	if util.IsNetError(aout.Records[0].Error) {
		phcontext := &hook_providers.ProviderHookContext{
			Logger: p.Logger,
			Store:  p.store,
		}
		return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
	}
	return nil, nil
}
