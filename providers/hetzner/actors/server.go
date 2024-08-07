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
	hcloud.Server
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
	schema.ServerListResponse
	Meta schema.Meta `json:"meta"`
}

type hcServerAttachToNetworkOptsWrap struct {
	hcloud.ServerAttachToNetworkOpts
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
	hcloud.ServerDetachFromNetworkOpts
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
	hcloud.ServerCreateImageOpts
	Server *hcServerWrap `json:"server"`
}

func (v *hcServerCreateImageOptsWrap) unwrap() (*hcloud.ServerCreateImageOpts, error) {
	return &hcloud.ServerCreateImageOpts{
		Type:        v.Type,
		Description: v.Description,
		Labels:      v.Labels,
	}, nil
}

type hcServerCreatePublicNetWrap struct {
	hcloud.ServerCreatePublicNet
	IPv4 *hcPrimaryIPWrap
	IPv6 *hcPrimaryIPWrap
}

func (v *hcServerCreatePublicNetWrap) unwrap() (*hcloud.ServerCreatePublicNet, error) {
	out := &hcloud.ServerCreatePublicNet{
		EnableIPv4: v.EnableIPv4,
		EnableIPv6: v.EnableIPv6,
	}
	if v.IPv4 != nil {
		hip4, err := v.IPv4.unwrap()
		if err != nil {
			return nil, err
		}
		out.IPv4 = hip4
	}
	if v.IPv6 != nil {
		hip6, err := v.IPv6.unwrap()
		if err != nil {
			return nil, err
		}
		out.IPv6 = hip6
	}
	return out, nil
}

type hcServerCreateOptsWrap struct {
	hcloud.ServerCreateOpts
	Image     *hcImageWrap
	PublicNet *hcServerCreatePublicNetWrap
	Networks  []*hcNetworkWrap
	SSHKeys   []*hcSSHKeyWrap
}

func (v *hcServerCreateOptsWrap) unwrap() (*hcloud.ServerCreateOpts, error) {
	out := &hcloud.ServerCreateOpts{
		Name:             v.Name,
		ServerType:       v.ServerType,
		Location:         v.Location,
		Datacenter:       v.Datacenter,
		UserData:         v.UserData,
		StartAfterCreate: v.StartAfterCreate,
		Labels:           v.Labels,
		Automount:        v.Automount,
		Volumes:          v.Volumes,
		Firewalls:        v.Firewalls,
		PlacementGroup:   v.PlacementGroup,
	}
	if v.Image != nil {
		him, err := v.Image.unwrap()
		if err != nil {
			return nil, err
		}
		out.Image = him
	}
	if v.PublicNet != nil {
		hpn, err := v.PublicNet.unwrap()
		if err != nil {
			return nil, err
		}
		out.PublicNet = hpn
	}
	for _, s := range v.Networks {
		hnet, err := s.unwrap()
		if err != nil {
			return nil, err
		}
		out.Networks = append(out.Networks, hnet)
	}
	for _, s := range v.SSHKeys {
		hssh, err := s.unwrap()
		if err != nil {
			return nil, err
		}
		out.SSHKeys = append(out.SSHKeys, hssh)
	}
	return out, nil
}

func CreateServer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcServerCreateOptsWrap{}
	output := &schema.ServerCreateResponse{}

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

	hopts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Server.Create(context.Background(), *hopts)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	err = UnmarshallHCloudToSchema(response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server")
				if err != nil {
					return nil, err
				}
				err = ctx.WaitForManyAndLog(output.NextActions, "Waiting for actions post server creation")
				if err != nil {
					return nil, err
				}
				if output.Server.ID == 0 {
					out := &schema.ActionGetResponse{}
					_, rsp, err := ctx.HClient.Action.GetByID(context.Background(), output.Action.ID)
					if err != nil {
						return nil, err
					}
					err = UnmarshallHCloudToSchema(rsp, out)
					if err != nil {
						return nil, err
					}
					for _, rr := range out.Action.Resources {
						if rr.Type == "server" {
							output.Server.ID = rr.ID
						}
					}
				}
			}
		}
	}

	if output.Server.ID == 0 {
		ctx.Logger.LogWarn("cannot dettermite server ID")
	}
	id := fmt.Sprintf("%v", output.Server.ID)
	return base.NewActionOutput(ctx.Action, output, &id), nil
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server delete")
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
		return nil, HCloudErrResponse(err, response)
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
			return nil, HCloudErrResponse(err, response)
		}
		output := &schema.ServerGetResponse{}
		err = UnmarshallHCloudToSchema(response, output)
		if err != nil {
			return nil, err
		}
		sid := fmt.Sprintf("%v", output.Server.ID)
		return base.NewActionOutput(ctx.Action, output, &sid), nil
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
		output := &schema.ServerGetResponse{}
		output.Server = raw.Servers[0]
		return base.NewActionOutput(ctx.Action, output, &sid), nil
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server power on")
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server power off")
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server attach to net")
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for server detach from net")
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
		return nil, HCloudErrResponse(err, response)
	}

	err = UnmarshallHCloudToSchema(response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for image creation")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	id := fmt.Sprintf("%v", output.Image.ID)
	return base.NewActionOutput(ctx.Action, output, &id), nil
}
