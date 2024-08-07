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
	"net"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcNetworkWrap struct {
	hcloud.Network
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
	hcloud.NetworkRoute
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
	hcloud.NetworkSubnet
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
	hcloud.NetworkCreateOpts
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
	schema.NetworkListResponse
	Meta schema.Meta `json:"meta"`
}

type hcNetworkAddSubnetOptsWrap struct {
	hcloud.NetworkAddSubnetOpts
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
	hcloud.NetworkDeleteSubnetOpts
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

type hcNetworkAddRouteOptsWrap struct {
	hcloud.NetworkAddRouteOpts
	Route   *hcNetworkRouteWrap
	Network *hcNetworkWrap
}

func (v *hcNetworkAddRouteOptsWrap) unwrap() (*hcloud.NetworkAddRouteOpts, error) {
	out := &hcloud.NetworkAddRouteOpts{}
	if v.Route != nil {
		hroutes, err := v.Route.unwrap()
		if err != nil {
			return nil, err
		}
		out.Route = *hroutes
	}
	return out, nil
}

type hcNetworkDeleteRouteOptsWrap struct {
	hcloud.NetworkDeleteRouteOpts
	Route   *hcNetworkRouteWrap
	Network *hcNetworkWrap
}

func (v *hcNetworkDeleteRouteOptsWrap) unwrap() (*hcloud.NetworkDeleteRouteOpts, error) {
	if v.Route == nil {
		return nil, fmt.Errorf("no route data")
	}
	hroute, err := v.Route.unwrap()
	if err != nil {
		return nil, err
	}
	return &hcloud.NetworkDeleteRouteOpts{
		Route: *hroute,
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
		return nil, HCloudErrResponse(err, response)
	}

	output := &schema.NetworkCreateResponse{}
	err = UnmarshallHCloudToSchema(response, output)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%v", output.Network.ID)
	return base.NewActionOutput(ctx.Action, output, &id), nil
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
		return nil, HCloudErrResponse(err, response)
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
	output := &schema.NetworkGetResponse{}
	output.Network = raw.Networks[0]
	id := fmt.Sprintf("%v", raw.Networks[0].ID)
	aout = base.NewActionOutput(ctx.Action, output, &id)
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for subnet addition to net")
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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for subnet deletion from net")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func AddRouteToNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcNetworkAddRouteOptsWrap{}
	output := &schema.NetworkActionAddRouteResponse{}

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

	_, response, err := ctx.HClient.Network.AddRoute(context.Background(), hnet, *opts)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for route addition to net")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func DeleteRouteFromNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcNetworkDeleteRouteOptsWrap{}
	output := &schema.NetworkActionDeleteRouteResponse{}

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

	_, response, err := ctx.HClient.Network.DeleteRoute(context.Background(), hnet, *opts)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for route deletion from net")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}
