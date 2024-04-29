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
	"net/netip"
	"strconv"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcLoadBalancerWrap struct {
	*hcloud.LoadBalancer
	ID *string `validate:"required"`
}

func (v *hcLoadBalancerWrap) unwrap() (*hcloud.LoadBalancer, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.LoadBalancer{ID: int64id}, nil
}

type hcLoadBalancerAttachToNetworkOptsWrap struct {
	*hcloud.LoadBalancerAttachToNetworkOpts
	Network           *hcNetworkWrap `json:"network"`
	IP                *string        `json:"ip"`
	lookupAvailableIP *net.IPNet
}

func (v *hcLoadBalancerAttachToNetworkOptsWrap) unwrap() (*hcloud.LoadBalancerAttachToNetworkOpts, error) {
	out := &hcloud.LoadBalancerAttachToNetworkOpts{}
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
			_, nn, err := net.ParseCIDR(*v.IP)
			if err != nil {
				return nil, err
			}
			v.lookupAvailableIP = nn
		}
		out.IP = ip
	}
	return out, nil
}

type hcLoadBalancerDetachFromNetworkOptsWrap struct {
	*hcloud.LoadBalancerDetachFromNetworkOpts
	Network *hcNetworkWrap `json:"network"`
}

func (v *hcLoadBalancerDetachFromNetworkOptsWrap) unwrap() (*hcloud.LoadBalancerDetachFromNetworkOpts, error) {
	out := &hcloud.LoadBalancerDetachFromNetworkOpts{}
	if v.Network != nil {
		net, err := v.Network.unwrap()
		if err != nil {
			return nil, err
		}
		out.Network = net
	}
	return out, nil
}

type loadbalancerAttachToNetworkParameters struct {
	AttachOpts   *hcLoadBalancerAttachToNetworkOptsWrap `json:"opts" validate:"required"`
	LoadBalancer *hcLoadBalancerWrap                    `json:"load_balancer" validate:"required"` // only LoadBalancer.ID is really used
}

type loadbalancerDetachFromNetworkParameters struct {
	DetachOpts   *hcLoadBalancerDetachFromNetworkOptsWrap `json:"opts" validate:"required"`
	LoadBalancer *hcLoadBalancerWrap                      `json:"load_balancer" validate:"required"` // only LoadBalancer.ID is really used
}

type LoadBalancerListResponseWithMeta struct {
	*schema.LoadBalancerListResponse
	Meta schema.Meta `json:"meta"`
}

type hcLoadBalancerAddIPTargetOptsWrap struct {
	*hcloud.LoadBalancerAddIPTargetOpts
	IP string `json:"ip" validate:"required"`
}

func (v *hcLoadBalancerAddIPTargetOptsWrap) unwrap() (*hcloud.LoadBalancerAddIPTargetOpts, error) {
	ip := net.ParseIP(v.IP)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip addr")
	}
	return &hcloud.LoadBalancerAddIPTargetOpts{IP: ip}, nil
}

type hcLoadBalancerAddServerTargetOptsWrap struct {
	*hcloud.LoadBalancerAddServerTargetOpts
	Server *hcServerWrap `validate:"required"`
}

type loadbalancerAddTargetParameters struct {
	Type              string                                         `json:"type" validate:"required"`
	IPOpts            *hcLoadBalancerAddIPTargetOptsWrap             `json:"ip_opts"`
	LabelSelectorOpts *hcloud.LoadBalancerAddLabelSelectorTargetOpts `json:"label_selector_opts"`
	ServerOpts        *hcLoadBalancerAddServerTargetOptsWrap         `json:"server_opts"`
	LoadBalancer      *hcLoadBalancerWrap                            `json:"load_balancer" validate:"required"` // only LoadBalancer.ID is really used
}

type loadbalancerRemoveTargetParameters struct {
	Type          string              `json:"type" validate:"required"`
	IP            *string             `json:"ip"`
	LabelSelector *string             `json:"label_selector"`
	Server        *hcServerWrap       `json:"server"`
	LoadBalancer  *hcLoadBalancerWrap `json:"load_balancer" validate:"required"` // only LoadBalancer.ID is really used
}

type hcLoadBalancerCreateOptsServiceWrap struct {
	*hcloud.LoadBalancerCreateOptsService
	ListenPort      *string
	DestinationPort *string
}

func (v *hcLoadBalancerCreateOptsServiceWrap) unwrap() (*hcloud.LoadBalancerCreateOptsService, error) {
	out := &hcloud.LoadBalancerCreateOptsService{
		Protocol:      v.Protocol,
		Proxyprotocol: v.Proxyprotocol,
		HTTP:          v.HTTP,
		HealthCheck:   v.HealthCheck,
	}
	if v.ListenPort != nil {
		intPort, err := strconv.ParseInt(*v.ListenPort, 10, 32)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use '%v' as int", *v.ListenPort), err)
		}
		ii := int(intPort)
		out.ListenPort = &ii
	}
	if v.DestinationPort != nil {
		intPort, err := strconv.ParseInt(*v.DestinationPort, 10, 32)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use '%v' as int", *v.DestinationPort), err)
		}
		ii := int(intPort)
		out.DestinationPort = &ii
	}
	return out, nil
}

type hcLoadBalancerCreateOptsWrap struct {
	*hcloud.LoadBalancerCreateOpts
	Network  *hcNetworkWrap `json:"network"`
	Services []hcLoadBalancerCreateOptsServiceWrap
}

func (v *hcLoadBalancerCreateOptsWrap) unwrap() (*hcloud.LoadBalancerCreateOpts, error) {
	out := &hcloud.LoadBalancerCreateOpts{
		Name:             v.Name,
		LoadBalancerType: v.LoadBalancerType,
		Algorithm:        v.Algorithm,
		Location:         v.Location,
		NetworkZone:      v.NetworkZone,
		Labels:           v.Labels,
		Targets:          v.Targets,
		PublicInterface:  v.PublicInterface,
	}
	for _, ss := range v.Services {
		hsv, err := ss.unwrap()
		if err != nil {
			return nil, err
		}
		out.Services = append(out.Services, *hsv)
	}
	if v.Network != nil {
		net, err := v.Network.unwrap()
		if err != nil {
			return nil, err
		}
		out.Network = net
	}
	return out, nil
}

func CreateLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcLoadBalancerCreateOptsWrap{}

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

	opts, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.LoadBalancer.Create(context.Background(), *opts)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only LoadBalancer.ID attr is really used
	input := &hcLoadBalancerWrap{}

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

	hlb, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	_, err = ctx.HClient.LoadBalancer.Delete(context.Background(), hlb)
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

	err := ctx.Store.DeepInterpolation(input)
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.LoadBalancer.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &LoadBalancerListResponseWithMeta{}
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
	raw := aout.Records[0].Value.(*LoadBalancerListResponseWithMeta)
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

func AttachLoadBalancerToNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerAttachToNetworkParameters{}

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

	hlb, err := input.LoadBalancer.unwrap()
	if err != nil {
		return nil, err
	}
	opts, err := input.AttachOpts.unwrap()
	if err != nil {
		return nil, err
	}
	if opts.Network == nil {
		return nil, fmt.Errorf("please, provide network id")
	}

	if input.AttachOpts.lookupAvailableIP != nil {
		ipnet := input.AttachOpts.lookupAvailableIP
		hnetID := opts.Network.ID
		hnet, _, err := ctx.HClient.Network.GetByID(context.Background(), hnetID)
		if err != nil {
			return nil, err
		}
		found := false
		for _, ss := range hnet.Subnets {
			ctx.Logger.LogDebug(fmt.Sprintf("found subnet %s", ss.IPRange.String()))
			if ss.IPRange.String() == ipnet.String() {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cannot found subnet %s", ipnet.String())
		}
		p, err := netip.ParsePrefix(ipnet.String())
		if err != nil {
			return nil, err
		}
		p = p.Masked()
		addr := p.Addr()
		for {
			time.Sleep(500 * time.Millisecond)
			ctx.Logger.LogDebug(fmt.Sprintf("looking for ip %s", addr.String()))
			if addr.As4()[3] == 0 {
				ctx.Logger.LogDebug(fmt.Sprintf("skip mask ip %s", addr.String()))
				addr = addr.Next()
				continue
			}
			if addr.As4()[3] == 255 {
				ctx.Logger.LogDebug(fmt.Sprintf("skip bcast ip %s", addr.String()))
				addr = addr.Next()
				continue
			}
			if !p.Contains(addr) {
				return nil, fmt.Errorf("cannot determine a valid ip for subnet %s", ipnet.String())
			}
			opts.IP = net.ParseIP(addr.String())
			_, response, err := ctx.HClient.LoadBalancer.AttachToNetwork(context.Background(), hlb, *opts)
			if herr, ok := err.(hcloud.Error); ok {
				if herr.Code == hcloud.ErrorCodeIPNotAvailable {
					// ok, already used ip, keep trying
					ctx.Logger.LogWarn(fmt.Sprintf("ip %s already used. Still looking for a valid ip...", opts.IP.String()))
					addr = addr.Next()
					continue
				}
			} else if err != nil {
				return nil, err
			}
			ctx.Logger.LogInfo(fmt.Sprintf("valid ip %s found for subnet %s", opts.IP.String(), ipnet.String()))
			output := &schema.LoadBalancerActionAttachToNetworkResponse{}
			return GenericHCloudOutput(ctx, response, output)
		}
	}
	_, response, err := ctx.HClient.LoadBalancer.AttachToNetwork(context.Background(), hlb, *opts)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerActionAttachToNetworkResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DetachLoadBalancerFromNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerDetachFromNetworkParameters{}

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

	hlb, err := input.LoadBalancer.unwrap()
	if err != nil {
		return nil, err
	}
	opts, err := input.DetachOpts.unwrap()
	if err != nil {
		return nil, err
	}

	_, response, err := ctx.HClient.LoadBalancer.DetachFromNetwork(context.Background(), hlb, *opts)
	if err != nil {
		return nil, err
	}

	output := &schema.LoadBalancerActionDetachFromNetworkResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func AddTargetToLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerAddTargetParameters{}

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

	hlb, err := input.LoadBalancer.unwrap()
	if err != nil {
		return nil, err
	}

	var response *hcloud.Response
	switch input.Type {
	case "server":
		if input.ServerOpts == nil {
			return nil, fmt.Errorf("please set server opts (server_opts)")
		}
		hsrv, err := input.ServerOpts.Server.unwrap()
		if err != nil {
			return nil, err
		}
		opts := &hcloud.LoadBalancerAddServerTargetOpts{
			Server: hsrv,
		}
		if input.ServerOpts != nil && input.ServerOpts.UsePrivateIP != nil {
			opts.UsePrivateIP = input.ServerOpts.UsePrivateIP
		}
		_, response, err = ctx.HClient.LoadBalancer.AddServerTarget(context.Background(), hlb, *opts)
		if err != nil {
			return nil, err
		}
	case "ip":
		if input.IPOpts == nil {
			return nil, fmt.Errorf("please set ip opts (op_opts)")
		}
		opts, err := input.IPOpts.unwrap()
		if err != nil {
			return nil, err
		}
		_, response, err = ctx.HClient.LoadBalancer.AddIPTarget(context.Background(), hlb, *opts)
		if err != nil {
			return nil, err
		}
	case "label_selector":
		if input.LabelSelectorOpts == nil {
			return nil, fmt.Errorf("please, set label selector opts (label_selector_opts)")
		}
		_, response, err = ctx.HClient.LoadBalancer.AddLabelSelectorTarget(context.Background(), hlb, *input.LabelSelectorOpts)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid resource type")

	}

	output := &schema.LoadBalancerActionAddTargetResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func RemoveTargetFromLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerRemoveTargetParameters{}

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

	hlb, err := input.LoadBalancer.unwrap()
	if err != nil {
		return nil, err
	}

	var response *hcloud.Response
	switch input.Type {
	case "server":
		if input.Server == nil {
			return nil, fmt.Errorf("please, provide server data")
		}
		hsrv, err := input.Server.unwrap()
		if err != nil {
			return nil, err
		}
		_, response, err = ctx.HClient.LoadBalancer.RemoveServerTarget(context.Background(), hlb, hsrv)
		if err != nil {
			return nil, err
		}
	case "ip":
		if input.IP == nil {
			return nil, fmt.Errorf("please provide ip data")
		}
		ip := net.ParseIP(*input.IP)
		if ip == nil {
			return nil, fmt.Errorf("invalid ip addr")
		}
		_, response, err = ctx.HClient.LoadBalancer.RemoveIPTarget(context.Background(), hlb, ip)
		if err != nil {
			return nil, err
		}
	case "label_selector":
		_, response, err = ctx.HClient.LoadBalancer.RemoveLabelSelectorTarget(context.Background(), hlb, *input.LabelSelector)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid resource type")

	}

	output := &schema.LoadBalancerActionRemoveTargetResponse{}
	return GenericHCloudOutput(ctx, response, output)
}
