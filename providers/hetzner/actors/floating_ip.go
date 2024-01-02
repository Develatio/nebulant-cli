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

type hcFloatingIPWrap struct {
	*hcloud.FloatingIP
	ID *string `validate:"required"`
}

func (v *hcFloatingIPWrap) unwrap() (*hcloud.FloatingIP, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.FloatingIP{ID: int64id}, nil
}

type findOneFloatingIPParameters struct {
	IdOrName *string `json:"id_or_name" validate:"required"`
}

type assignFloatingIPParameters struct {
	// Only FloatingIP.ID is really ussed
	FloatingIP *hcFloatingIPWrap `json:"floating_ip" validate:"required"`
	// only Server.ID is really ussed
	Server *hcloud.Server `json:"server" validate:"required"`
}

type unassignFloatingIPParameters struct {
	// Only FloatingIP.ID is really ussed
	FloatingIP *hcFloatingIPWrap `json:"floating_ip" validate:"required"`
}

// CreateFloatingIP func
func CreateFloatingIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.FloatingIPCreateOpts{}

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

	_, response, err := ctx.HClient.FloatingIP.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.FloatingIPCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteFloatingIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only FloatingIP.ID attr are really used
	// https://github.com/hetznercloud/hcloud-go/blob/v2.3.0/hcloud/floating_ip.go#L279
	input := &hcFloatingIPWrap{}

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

	hfip, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, err = ctx.HClient.FloatingIP.Delete(context.Background(), hfip)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindOneFloatingIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &findOneFloatingIPParameters{}

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

	_, response, err := ctx.HClient.FloatingIP.Get(context.Background(), *input.IdOrName)
	if err != nil {
		return nil, err
	}

	output := &schema.FloatingIPGetResponse{}
	err = UnmarshallHCloudToSchema(response, output)
	if err != nil {
		return nil, err
	}
	sid := fmt.Sprintf("%v", output.FloatingIP.ID)
	return base.NewActionOutput(ctx.Action, output.FloatingIP, &sid), nil
}

func AssignFloatingIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &assignFloatingIPParameters{}

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

	hfip, err := input.FloatingIP.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.FloatingIP.Assign(context.Background(), hfip, input.Server)
	if err != nil {
		return nil, err
	}

	output := &schema.FloatingIPActionAssignResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func UnassignFloatingIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &unassignFloatingIPParameters{}

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

	hfip, err := input.FloatingIP.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.FloatingIP.Unassign(context.Background(), hfip)
	if err != nil {
		return nil, err
	}

	output := &schema.FloatingIPActionUnassignRequest{}
	return GenericHCloudOutput(ctx, response, output)
}
