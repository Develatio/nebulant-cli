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

type loadbalancerAttachToNetworkParameters struct {
	AttachOpts   hcloud.LoadBalancerAttachToNetworkOpts `json:"opts" validate:"required"`
	LoadBalancer *hcloud.LoadBalancer                   `json:"load_balancer" validate:"required"` // only LoadBalancer.ID is really used
}

type loadbalancerDetachFromNetworkParameters struct {
	DetachOpts   hcloud.LoadBalancerDetachFromNetworkOpts `json:"opts" validate:"required"`
	LoadBalancer *hcloud.LoadBalancer                     `json:"load_balancer" validate:"required"` // only LoadBalancer.ID is really used
}

func CreateLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.LoadBalancerCreateOpts{}

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

	_, response, err := ctx.HClient.LoadBalancer.Create(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only LoadBalancer.ID attr is really used
	input := &hcloud.LoadBalancer{}

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

	_, err = ctx.HClient.LoadBalancer.Delete(context.Background(), input)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindLoadBalancers(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.LoadBalancerListOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.LoadBalancer.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerListResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindLoadBalancers(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no load balancer found")
	}
	raw := aout.Records[0].Value.(*schema.LoadBalancerListResponse)
	found := len(raw.LoadBalancers)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no load balancer found")
	}
	id := fmt.Sprintf("%v", raw.LoadBalancers[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.LoadBalancers[0], &id)
	return aout, nil
}

func AttachToNetworkLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerAttachToNetworkParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.LoadBalancer.AttachToNetwork(context.Background(), input.LoadBalancer, input.AttachOpts)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerActionAttachToNetworkResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DetachFromNetworkLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerDetachFromNetworkParameters{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.LoadBalancer.DetachFromNetwork(context.Background(), input.LoadBalancer, input.DetachOpts)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerActionDetachFromNetworkResponse{}
	return GenericHCloudOutput(ctx, response, output)
}
