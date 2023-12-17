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

type findOneServerParameters struct {
	hcloud.ServerListOpts
	ID *int64 `json:"id"`
}

type ServerListResponseWithMeta struct {
	*schema.ServerListResponse
	Meta schema.Meta `json:"meta"`
}

func CreateServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.ServerCreateOpts{}

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

	_, response, err := ctx.HClient.Server.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.ServerCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Server.ID attr are really used
	input := &hcloud.Server{}

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

	_, response, err := ctx.HClient.Server.DeleteWithResult(context.Background(), input)
	if err != nil {
		return nil, err
	}

	output := &schema.ServerDeleteResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindServers(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.ServerListOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.Server.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &ServerListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneServer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &findOneServerParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if input.ID != nil {
		if ctx.Rehearsal {
			return nil, nil
		}
		_, response, err := ctx.HClient.Server.GetByID(context.Background(), *input.ID)
		if err != nil {
			return nil, err
		}
		output := &schema.ServerGetResponse{}
		err = UnmarshallHCloudToSchema(response, output)
		if err != nil {
			return nil, err
		}
		sid := fmt.Sprintf("%v", output.Server.ID)
		return base.NewActionOutput(ctx.Action, output.Server, &sid), nil
	} else {
		aout, err := FindServers(ctx)
		if err != nil {
			return nil, err
		}

		if ctx.Rehearsal {
			return nil, nil
		}

		if len(aout.Records) <= 0 {
			return nil, fmt.Errorf("no server found")
		}

		raw := aout.Records[0].Value.(*ServerListResponseWithMeta)
		found := len(raw.Servers)

		if found > 1 {
			return nil, fmt.Errorf("too many results")
		}

		if found <= 0 {
			return nil, fmt.Errorf("no server found")
		}

		output := &schema.ServerGetResponse{}
		output.Server = raw.Servers[0]
		sid := fmt.Sprintf("%v", output.Server.ID)
		return base.NewActionOutput(ctx.Action, output.Server, &sid), nil
	}
}

func PowerOnServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Server.ID are really used
	input := &hcloud.Server{}

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

	_, response, err := ctx.HClient.Server.Poweron(context.Background(), input)
	if err != nil {
		return nil, err
	}

	output := &schema.ServerActionPoweronResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func PowerOffServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Server.ID are really used
	input := &hcloud.Server{}

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

	_, response, err := ctx.HClient.Server.Poweroff(context.Background(), input)
	if err != nil {
		return nil, err
	}

	output := &schema.ServerActionPoweroffResponse{}
	return GenericHCloudOutput(ctx, response, output)
}
