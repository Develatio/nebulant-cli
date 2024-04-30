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
	"net"
	"strconv"

	"encoding/json"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcServerWrap struct {
	*hcloud.Server
	ID *string `validate:"required"`
}

func (v *hcServerWrap) unwrap() (*hcloud.Server, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.Server{ID: int64id}, nil
}

type findOneServerParameters struct {
	hcloud.ServerListOpts
	ID *string `json:"id"`
}

type ServerListResponseWithMeta struct {
	*schema.ServerListResponse
	Meta schema.Meta `json:"meta"`
}

type hcServerAttachToNetworkOptsWrap struct {
	*hcloud.ServerAttachToNetworkOpts
	Server   *hcServerWrap  `json:"server"`
	Network  *hcNetworkWrap `json:"network"`
	IP       *string        `json:"ip"`
	AliasIPs []*string      `json:"alias_ips"`
}

func (v *hcServerAttachToNetworkOptsWrap) unwrap() (*hcloud.ServerAttachToNetworkOpts, error) {
	out := &hcloud.ServerAttachToNetworkOpts{}
	if v.Network != nil {
		n, err := v.Network.unwrap()
		if err != nil {
			return nil, err
		}
		out.Network = n
	}
	if v.IP != nil {
		ip := net.ParseIP(*v.IP)
		if ip == nil {
			return nil, fmt.Errorf("invalid ip addr %s", *v.IP)
		}
		out.IP = ip
	}
	for _, al := range v.AliasIPs {
		ip := net.ParseIP(*al)
		if ip == nil {
			return nil, fmt.Errorf("invalid alias ip addr %s", *al)
		}
		out.AliasIPs = append(out.AliasIPs, ip)
	}
	return out, nil
}

type hcServerDetachFromNetworkOptsWrap struct {
	*hcloud.ServerDetachFromNetworkOpts
	Server  *hcServerWrap  `json:"server"`
	Network *hcNetworkWrap `json:"network"`
}

func (v *hcServerDetachFromNetworkOptsWrap) unwrap() (*hcloud.ServerDetachFromNetworkOpts, error) {
	out := &hcloud.ServerDetachFromNetworkOpts{}
	if v.Network != nil {
		n, err := v.Network.unwrap()
		if err != nil {
			return nil, err
		}
		out.Network = n
	}
	return out, nil
}

type hcServerCreateImageOptsWrap struct {
	*hcloud.ServerCreateImageOpts
	Server *hcServerWrap `json:"server"`
}

func (v *hcServerCreateImageOptsWrap) unwrap() (*hcloud.ServerCreateImageOpts, error) {
	return &hcloud.ServerCreateImageOpts{
		Type:        v.Type,
		Description: v.Description,
		Labels:      v.Labels,
	}, nil
}

func CreateServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.ServerCreateOpts{}
	output := &schema.ServerCreateResponse{}

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

	_, response, err := ctx.HClient.Server.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func DeleteServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Server.ID attr are really used
	input := &hcServerWrap{}
	output := &schema.ServerDeleteResponse{}

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

	hsrv, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.DeleteWithResult(context.Background(), hsrv)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server delete %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func FindServers(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.ServerListOpts{}

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
		err := ctx.Store.DeepInterpolation(input)
		if err != nil {
			return nil, err
		}
		int64id, err := strconv.ParseInt(*input.ID, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *input.ID), err)
		}
		_, response, err := ctx.HClient.Server.GetByID(context.Background(), int64id)
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

		sid := fmt.Sprintf("%v", raw.Servers[0].ID)
		return base.NewActionOutput(ctx.Action, raw.Servers[0], &sid), nil
	}
}

func PowerOnServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Server.ID are really used
	input := &hcServerWrap{}
	output := &schema.ServerActionPoweronResponse{}

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

	hsrv, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.Poweron(context.Background(), hsrv)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server power on %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func PowerOffServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Server.ID are really used
	input := &hcServerWrap{}
	output := &schema.ServerActionPoweroffResponse{}

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

	hsrv, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.Poweroff(context.Background(), hsrv)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server power off %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func AttachServerToNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcServerAttachToNetworkOptsWrap{}
	output := &schema.ServerActionAttachToNetworkResponse{}

	if err := json.Unmarshal(ctx.Action.Parameters, input); err != nil {
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

	opts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	hsrv, err := input.Server.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.AttachToNetwork(context.Background(), hsrv, *opts)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server attach to net %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func DetachServerFromNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcServerDetachFromNetworkOptsWrap{}
	output := &schema.ServerActionDetachFromNetworkResponse{}

	if err := json.Unmarshal(ctx.Action.Parameters, input); err != nil {
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

	opts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	hsrv, err := input.Server.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.DetachFromNetwork(context.Background(), hsrv, *opts)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server detach from net %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func CreateImageFromServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcServerCreateImageOptsWrap{}
	output := &schema.ServerActionCreateImageResponse{}

	if err := json.Unmarshal(ctx.Action.Parameters, input); err != nil {
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

	opts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	hsrv, err := input.Server.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.CreateImage(context.Background(), hsrv, opts)
	if err != nil {
		return nil, err
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for image creation %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}
