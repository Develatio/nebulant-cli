// MIT License
//
// Copyright (C) 2023 Develatio Technologies S.L.

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

package actors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcPrimaryIPCreateOptsWrap struct {
	hcloud.PrimaryIPCreateOpts
	AssigneeID *string
}

func (v *hcPrimaryIPCreateOptsWrap) unwrap() (*hcloud.PrimaryIPCreateOpts, error) {
	output := &hcloud.PrimaryIPCreateOpts{
		AssigneeType: v.AssigneeType,
		AutoDelete:   v.AutoDelete,
		Datacenter:   v.Datacenter,
		Labels:       v.Labels,
		Name:         v.Name,
		Type:         v.Type,
	}
	if v.AssigneeID == nil {
		return output, nil
	}
	int64aid, err := strconv.ParseInt(*v.AssigneeID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.AssigneeID), err)
	}
	return &hcloud.PrimaryIPCreateOpts{AssigneeID: &int64aid}, nil
}

type hcPrimaryIPAssignOptsWrap struct {
	hcloud.PrimaryIPAssignOpts
	ID         *string `validate:"required"`
	AssigneeID *string `validate:"required"`
}

func (v *hcPrimaryIPAssignOptsWrap) unwrap() (*hcloud.PrimaryIPAssignOpts, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	int64aid, err := strconv.ParseInt(*v.AssigneeID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 AssigneeID", *v.AssigneeID), err)
	}
	return &hcloud.PrimaryIPAssignOpts{ID: int64id, AssigneeID: int64aid}, nil
}

type hcPrimaryIPWrap struct {
	hcloud.PrimaryIP
	ID *string `validate:"required"`
}

func (v *hcPrimaryIPWrap) unwrap() (*hcloud.PrimaryIP, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.PrimaryIP{ID: int64id}, nil
}

type unassignPrimaryIPParameters struct {
	ID string `json:"id" validate:"required"`
}

type PrimaryIPListResultWithMeta struct {
	schema.PrimaryIPListResult
	Meta schema.Meta `json:"meta"`
}

func CreatePrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcPrimaryIPCreateOptsWrap{}
	output := &schema.PrimaryIPCreateResponse{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	internalparams := &blueprint.InternalParameters{}
	err = json.Unmarshal(ctx.Action.Parameters, internalparams)
	if err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	hipcreateopts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.PrimaryIP.Create(context.Background(), *hipcreateopts)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	err = UnmarshallHCloudToSchema(response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" && output.Action != nil {
				err = ctx.WaitForAndLog(*output.Action, "Waiting for primary ip")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	id := fmt.Sprintf("%v", output.PrimaryIP.ID)
	return base.NewActionOutput(ctx.Action, output, &id), nil
}

func DeletePrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only PrimaryIP.ID attr are really used
	input := &hcPrimaryIPWrap{}

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

	hip, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, err = ctx.HClient.PrimaryIP.Delete(context.Background(), hip)
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

	err := ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.PrimaryIP.List(context.Background(), *input)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	output := &PrimaryIPListResultWithMeta{}
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
	raw := aout.Records[0].Value.(*PrimaryIPListResultWithMeta)
	found := len(raw.PrimaryIPs)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no primary ip found")
	}
	output := &schema.PrimaryIPGetResult{}
	output.PrimaryIP = raw.PrimaryIPs[0]
	id := fmt.Sprintf("%v", raw.PrimaryIPs[0].ID)
	aout = base.NewActionOutput(ctx.Action, output, &id)
	return aout, nil
}

func AssignPrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcPrimaryIPAssignOptsWrap{}
	// ok to use hcloud instead scheme here
	output := &hcloud.PrimaryIPAssignResult{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	internalparams := &blueprint.InternalParameters{}
	err := json.Unmarshal(ctx.Action.Parameters, internalparams)
	if err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	hipassignopts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.PrimaryIP.Assign(context.Background(), *hipassignopts)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for primary ip assignation")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func UnassignPrimaryIP(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &unassignPrimaryIPParameters{}
	output := &hcloud.PrimaryIPAssignResult{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	internalparams := &blueprint.InternalParameters{}
	err := json.Unmarshal(ctx.Action.Parameters, internalparams)
	if err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	int64id, err := strconv.ParseInt(input.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", input.ID), err)
	}

	_, response, err := ctx.HClient.PrimaryIP.Unassign(context.Background(), int64id)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for primary ip unassignation")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}
