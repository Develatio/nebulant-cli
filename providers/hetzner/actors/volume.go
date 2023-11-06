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

type volumeAttachParameters struct {
	AttachOpts hcloud.VolumeAttachOpts `json:"attach_opts" validate:"required"`
	Volume     *hcloud.Volume          `json:"volume" validate:"required"` // only Volume.ID is really used
}

func CreateVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.VolumeCreateOpts{}

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

	_, response, err := ctx.HClient.Volume.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.VolumeCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Volume.ID attr are really used
	input := &hcloud.Volume{}

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

	_, err = ctx.HClient.Volume.Delete(context.Background(), input)
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

	_, response, err := ctx.HClient.Volume.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.VolumeListResponse{}
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
	raw := aout.Records[0].Value.(*schema.VolumeListResponse)
	found := len(raw.Volumes)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no volume found")
	}
	id := fmt.Sprintf("%v", raw.Volumes[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.Volumes[0], &id)
	return aout, nil
}

func AttachVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &volumeAttachParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.Volume.AttachWithOpts(context.Background(), input.Volume, input.AttachOpts)
	if err != nil {
		return nil, err
	}

	output := &schema.VolumeActionAttachVolumeResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DetachVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	// only volume.ID is really used
	input := &hcloud.Volume{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.Volume.Detach(context.Background(), input)
	if err != nil {
		return nil, err
	}

	output := &schema.VolumeActionDetachVolumeResponse{}
	return GenericHCloudOutput(ctx, response, output)
}