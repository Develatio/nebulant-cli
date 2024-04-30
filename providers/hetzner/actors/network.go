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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcNetworkWrap struct {
	*hcloud.Network
	ID *string `validate:"required"`
}

func (v *hcNetworkWrap) unwrap() (*hcloud.Network, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.Network{ID: int64id}, nil
}

type hcNetworkRouteWrap struct {
	*hcloud.NetworkRoute
	Gateway     *string `json:"gateway"`
	Destination *string `json:"destination"`
}

func (v *hcNetworkRouteWrap) unwrap() (*hcloud.NetworkRoute, error) {
	out := &hcloud.NetworkRoute{}
	if v.Destination != nil {
		_, nn, err := net.ParseCIDR(*v.Destination)
		if err != nil {
			return nil, err
		}
		out.Destination = nn
	}
	if v.Gateway != nil {
		ip := net.ParseIP(*v.Gateway)
		if ip == nil {
			return nil, fmt.Errorf("invalid gateway addr")
		}
		out.Gateway = ip
	}
	return out, nil
}

type hcNetworkSubnetWrap struct {
	*hcloud.NetworkSubnet
	IPRange     *string `json:"ip_range"`
	Gateway     *string `json:"gateway"`
	VSwitchID   *string `json:"vswitch_id"`
	Type        *string `json:"type"`
	NetworkZone *string `json:"network_zone"`
}

func (v *hcNetworkSubnetWrap) unwrap() (*hcloud.NetworkSubnet, error) {
	out := &hcloud.NetworkSubnet{}
	if v.Type != nil {
		out.Type = hcloud.NetworkSubnetType(*v.Type)
	}
	if v.NetworkZone != nil {
		out.NetworkZone = hcloud.NetworkZone(*v.NetworkZone)
	}
	if v.VSwitchID != nil {
		int64id, err := strconv.ParseInt(*v.VSwitchID, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.VSwitchID), err)
		}
		out.VSwitchID = int64id
	}
	if v.IPRange != nil {
		_, nn, err := net.ParseCIDR(*v.IPRange)
		if err != nil {
			return nil, err
		}
		out.IPRange = nn
	}
	if v.Gateway != nil {
		ip := net.ParseIP(*v.Gateway)
		if ip == nil {
			return nil, fmt.Errorf("invalid gateway addr")
		}
		out.Gateway = ip
	}
	return out, nil
}

type hcNetworkCreateOptsWrap struct {
	*hcloud.NetworkCreateOpts
	IPRange *string                `json:"ip_range"`
	Subnets []*hcNetworkSubnetWrap `json:"subnets"`
	Routes  []*hcNetworkRouteWrap  `json:"routes"`
}

func (v *hcNetworkCreateOptsWrap) unwrap() (*hcloud.NetworkCreateOpts, error) {
	out := &hcloud.NetworkCreateOpts{
		Name:   v.Name,
		Labels: v.Labels,
	}
	if v.IPRange != nil {
		_, nn, err := net.ParseCIDR(*v.IPRange)
		if err != nil {
			return nil, err
		}
		if nn == nil {
			return nil, fmt.Errorf("invalid ip value %s: nil result", *v.IPRange)
		}
		if nn.String() == "" {
			return nil, fmt.Errorf("invalid ip value %s: empty result", *v.IPRange)
		}
		out.IPRange = nn
	}
	for _, s := range v.Subnets {
		ss, err := s.unwrap()
		if err != nil {
			return nil, err
		}
		out.Subnets = append(out.Subnets, *ss)
	}
	for _, s := range v.Routes {
		ss, err := s.unwrap()
		if err != nil {
			return nil, err
		}
		out.Routes = append(out.Routes, *ss)
	}
	return out, nil
}

type NetworkListResponseWithMeta struct {
	*schema.NetworkListResponse
	Meta schema.Meta `json:"meta"`
}

type hcNetworkAddSubnetOptsWrap struct {
	*hcloud.NetworkAddSubnetOpts
	Subnet  *hcNetworkSubnetWrap `json:"subnet" validate:"required"`
	Network *hcNetworkWrap       `json:"network" validate:"required"`
}

func (v *hcNetworkAddSubnetOptsWrap) unwrap() (*hcloud.NetworkAddSubnetOpts, error) {
	if v.Subnet == nil {
		return nil, fmt.Errorf("no subnet data")
	}
	hnet, err := v.Subnet.unwrap()
	if err != nil {
		return nil, err
	}
	return &hcloud.NetworkAddSubnetOpts{
		Subnet: *hnet,
	}, nil
}

type hcNetworkDeleteSubnetOptsWrap struct {
	*hcloud.NetworkDeleteSubnetOpts
	Subnet  *hcNetworkSubnetWrap `json:"subnet" validate:"required"`
	Network *hcNetworkWrap       `json:"network" validate:"required"`
}

func (v *hcNetworkDeleteSubnetOptsWrap) unwrap() (*hcloud.NetworkDeleteSubnetOpts, error) {
	if v.Subnet == nil {
		return nil, fmt.Errorf("no subnet data")
	}
	hnet, err := v.Subnet.unwrap()
	if err != nil {
		return nil, err
	}
	return &hcloud.NetworkDeleteSubnetOpts{
		Subnet: *hnet,
	}, nil
}

func CreateNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcNetworkCreateOptsWrap{}

	if err := json.Unmarshal(ctx.Action.Parameters, input); err != nil {
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

	_, response, err := ctx.HClient.Network.Create(context.Background(), *opts)
	if err != nil {
		return nil, err
	}

	output := &schema.NetworkCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Network.ID attr are really used
	input := &hcNetworkWrap{}

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

	hnet, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, err = ctx.HClient.Network.Delete(context.Background(), hnet)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindNetworks(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.NetworkListOpts{}

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

	_, response, err := ctx.HClient.Network.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &NetworkListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindNetworks(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no network found")
	}
	raw := aout.Records[0].Value.(*NetworkListResponseWithMeta)
	found := len(raw.Networks)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no network found")
	}
	id := fmt.Sprintf("%v", raw.Networks[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.Networks[0], &id)
	return aout, nil
}

func AddSubnetToNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcNetworkAddSubnetOptsWrap{}
	output := &schema.NetworkActionAddSubnetResponse{}

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

	opts, err := input.unwrap()
	if err != nil {
		return nil, err
	}
	hnet, err := input.Network.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Network.AddSubnet(context.Background(), hnet, *opts)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for subnet addition to net %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func DeleteSubnetFromNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcNetworkDeleteSubnetOptsWrap{}
	output := &schema.NetworkActionDeleteSubnetResponse{}

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

	opts, err := input.unwrap()
	if err != nil {
		return nil, err
	}
	hnet, err := input.Network.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.Network.DeleteSubnet(context.Background(), hnet, *opts)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for subnet addition to net %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}
