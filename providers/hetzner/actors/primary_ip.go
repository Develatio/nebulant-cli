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
	"fmt"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type unassignPrimaryIPParameters struct {
	ID int64 `json:"id" validate:"required"`
}

func CreatePrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.PrimaryIPCreateOpts{}

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

	_, response, err := ctx.HClient.PrimaryIP.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.PrimaryIPCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeletePrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only PrimaryIP.ID attr are really used
	input := &hcloud.PrimaryIP{}

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

	_, err = ctx.HClient.PrimaryIP.Delete(context.Background(), input)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindPrimaryIPs(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.PrimaryIPListOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.PrimaryIP.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.PrimaryIPListResult{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOnePrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindPrimaryIPs(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no primary ip found")
	}
	raw := aout.Records[0].Value.(*schema.PrimaryIPListResult)
	found := len(raw.PrimaryIPs)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no primary ip found")
	}
	id := fmt.Sprintf("%v", raw.PrimaryIPs[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.PrimaryIPs[0], &id)
	return aout, nil
}

func AssignPrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	input := hcloud.PrimaryIPAssignOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.PrimaryIP.Assign(context.Background(), input)
	if err != nil {
		return nil, err
	}

	// ok to use hcloud instead scheme here
	output := &hcloud.PrimaryIPAssignResult{}
	return GenericHCloudOutput(ctx, response, output)
}

func UnassignPrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &unassignPrimaryIPParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.PrimaryIP.Unassign(context.Background(), input.ID)
	if err != nil {
		return nil, err
	}

	output := &hcloud.PrimaryIPAssignResult{}
	return GenericHCloudOutput(ctx, response, output)
}
