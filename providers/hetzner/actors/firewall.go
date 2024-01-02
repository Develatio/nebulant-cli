// Nebulant
// Copyright (C) 2023 Develatio Technologies S.L.

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

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcFirewallWrap struct {
	*hcloud.Firewall
	ID *string `validate:"required"`
}

func (v *hcFirewallWrap) unwrap() (*hcloud.Firewall, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.Firewall{ID: int64id}, nil
}

type applyResourcesParameters struct {
	Resources []hcloud.FirewallResource `json:"resources" validate:"required"`
	Firewall  *hcFirewallWrap           `json:"firewall" validate:"required"` // only Firewall.ID is really used
}

type removeResourcesParameters struct {
	Resources []hcloud.FirewallResource `json:"resources" validate:"required"`
	Firewall  *hcFirewallWrap           `json:"firewall" validate:"required"` // only Firewall.ID is really used
}

type setRulesParameters struct {
	Opts     hcloud.FirewallSetRulesOpts `json:"opts" validate:"required"`
	Firewall *hcFirewallWrap             `json:"firewall" validate:"required"` // only Firewall.ID is really used
}

type FirewallListResponseWithMeta struct {
	*schema.FirewallListResponse
	Meta schema.Meta `json:"meta"`
}

func CreateFirewall(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.FirewallCreateOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Firewall.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.FirewallCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteFirewall(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Firewall.ID attr are really used
	input := &hcFirewallWrap{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	hfwall, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, err = ctx.HClient.Firewall.Delete(context.Background(), hfwall)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindFirewalls(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.FirewallListOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.Firewall.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &FirewallListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneFirewall(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindFirewalls(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no firewall found")
	}
	raw := aout.Records[0].Value.(*FirewallListResponseWithMeta)
	found := len(raw.Firewalls)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no firewall found")
	}
	id := fmt.Sprintf("%v", raw.Firewalls[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.Firewalls[0], &id)
	return aout, nil
}

func ApplyToResourcesFirewall(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &applyResourcesParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err := ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	hfwall, err := input.Firewall.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Firewall.ApplyResources(context.Background(), hfwall, input.Resources)
	if err != nil {
		return nil, err
	}

	output := &schema.FirewallActionApplyToResourcesResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func RemoveFromResourcesFirewall(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &removeResourcesParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err := ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	hfwall, err := input.Firewall.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Firewall.RemoveResources(context.Background(), hfwall, input.Resources)
	if err != nil {
		return nil, err
	}

	output := &schema.FirewallActionRemoveFromResourcesRequest{}
	return GenericHCloudOutput(ctx, response, output)
}

func SetRulesFirewall(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &setRulesParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err := ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	hfwall, err := input.Firewall.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Firewall.SetRules(context.Background(), hfwall, input.Opts)
	if err != nil {
		return nil, err
	}

	output := &schema.FirewallActionSetRulesResponse{}
	return GenericHCloudOutput(ctx, response, output)
}
