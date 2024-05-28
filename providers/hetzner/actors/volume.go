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

type hcVolumeWrap struct {
	hcloud.Volume
	ID *string `validate:"required"`
}

func (v *hcVolumeWrap) unwrap() (*hcloud.Volume, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.Volume{ID: int64id}, nil
}

type volumeAttachParameters struct {
	AttachOpts hcloud.VolumeAttachOpts `json:"attach_opts" validate:"required"`
	Volume     *hcVolumeWrap           `json:"volume" validate:"required"` // only Volume.ID is really used
	Server     *hcServerWrap           `json:"server" validate:"required"` // only Server.ID is really used
}

type VolumeListResponseWithMeta struct {
	schema.VolumeListResponse
	Meta schema.Meta `json:"meta"`
}

func CreateVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.VolumeCreateOpts{}
	output := &schema.VolumeCreateResponse{}

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

	_, response, err := ctx.HClient.Volume.Create(context.Background(), *input)
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
				err = ctx.WaitForAndLog(*output.Action, "Waiting for volume")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	id := fmt.Sprintf("%v", output.Volume.ID)
	return base.NewActionOutput(ctx.Action, output, &id), nil
}

func DeleteVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Volume.ID attr are really used
	input := &hcVolumeWrap{}

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

	hvol, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, err = ctx.HClient.Volume.Delete(context.Background(), hvol)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindVolumes(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.VolumeListOpts{}

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

	_, response, err := ctx.HClient.Volume.List(context.Background(), *input)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	output := &VolumeListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindVolumes(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no volume found")
	}
	raw := aout.Records[0].Value.(*VolumeListResponseWithMeta)
	found := len(raw.Volumes)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no volume found")
	}
	output := &schema.VolumeGetResponse{}
	output.Volume = raw.Volumes[0]
	id := fmt.Sprintf("%v", raw.Volumes[0].ID)
	aout = base.NewActionOutput(ctx.Action, output, &id)
	return aout, nil
}

func AttachVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &volumeAttachParameters{}
	output := &schema.VolumeActionAttachVolumeResponse{}

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

	hvol, err := input.Volume.unwrap()
	if err != nil {
		return nil, err
	}

	hsrv, err := input.Server.unwrap()
	if err != nil {
		return nil, err
	}
	input.AttachOpts.Server = hsrv

	_, response, err := ctx.HClient.Volume.AttachWithOpts(context.Background(), hvol, input.AttachOpts)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for volume attach")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func DetachVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	// only volume.ID is really used
	input := &hcVolumeWrap{}
	output := &schema.VolumeActionDetachVolumeResponse{}

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

	hvol, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Volume.Detach(context.Background(), hvol)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for volume detach")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}
