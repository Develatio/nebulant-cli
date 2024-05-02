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
	"net/netip"
	"strconv"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
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

type hcLoadBalancerCreateOptsServiceHealthCheckHTTPWrap struct {
	*hcloud.LoadBalancerCreateOptsServiceHealthCheckHTTP
	TLS *bool
}

func (v *hcLoadBalancerCreateOptsServiceHealthCheckHTTPWrap) unwrap() (*hcloud.LoadBalancerCreateOptsServiceHealthCheckHTTP, error) {
	out := &hcloud.LoadBalancerCreateOptsServiceHealthCheckHTTP{
		Domain:      v.Domain,
		Path:        v.Path,
		Response:    v.Response,
		StatusCodes: v.StatusCodes,
		TLS:         v.TLS,
	}
	return out, nil
}

type hcLoadBalancerCreateOptsServiceHealthCheckWrap struct {
	*hcloud.LoadBalancerCreateOptsServiceHealthCheck
	Protocol *string
	Port     *string
	Interval *string
	Timeout  *string
	Retries  *string
	HTTP     *hcLoadBalancerCreateOptsServiceHealthCheckHTTPWrap
}

func (v *hcLoadBalancerCreateOptsServiceHealthCheckWrap) unwrap() (*hcloud.LoadBalancerCreateOptsServiceHealthCheck, error) {
	out := &hcloud.LoadBalancerCreateOptsServiceHealthCheck{}
	if v.HTTP != nil {
		hh, err := v.HTTP.unwrap()
		if err != nil {
			return nil, err
		}
		out.HTTP = hh
	}
	if v.Protocol != nil {
		out.Protocol = hcloud.LoadBalancerServiceProtocol(*v.Protocol)
	}
	if v.Port != nil {
		intPort, err := strconv.ParseInt(*v.Port, 10, 32)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use port '%v' as int", *v.Port), err)
		}
		ii := int(intPort)
		out.Port = &ii
	}
	if v.Retries != nil {
		intRetries, err := strconv.ParseInt(*v.Retries, 10, 32)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use retries '%v' as int", *v.Retries), err)
		}
		ii := int(intRetries)
		out.Retries = &ii
	}
	if v.Timeout != nil {
		intTimeout, err := strconv.ParseInt(*v.Timeout, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use timeout '%v' as int seconds", *v.Timeout), err)
		}
		d := time.Duration(intTimeout) * time.Second
		out.Timeout = &d
	}
	if v.Interval != nil {
		intInterval, err := strconv.ParseInt(*v.Interval, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use interval '%v' as int seconds", *v.Interval), err)
		}
		d := time.Duration(intInterval) * time.Second
		out.Interval = &d
	}
	return out, nil
}

type hcLoadBalancerCreateOptsServiceHTTPWrap struct {
	*hcloud.LoadBalancerCreateOptsServiceHTTP
	CookieLifetime *string
}

func (v *hcLoadBalancerCreateOptsServiceHTTPWrap) unwrap() (*hcloud.LoadBalancerCreateOptsServiceHTTP, error) {
	out := &hcloud.LoadBalancerCreateOptsServiceHTTP{
		CookieName: v.CookieName,
		// CookieLifetime: v.CookieLifetime,
		Certificates:   v.Certificates,
		RedirectHTTP:   v.RedirectHTTP,
		StickySessions: v.StickySessions,
	}
	if v.CookieLifetime != nil {
		intCookieLifetime, err := strconv.ParseInt(*v.CookieLifetime, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use CookieLifetime '%v' as int seconds", *v.CookieLifetime), err)
		}
		d := time.Duration(intCookieLifetime) * time.Second
		out.CookieLifetime = &d
	}
	return out, nil
}

type hcLoadBalancerCreateOptsServiceWrap struct {
	*hcloud.LoadBalancerCreateOptsService
	Proxyprotocol   *bool
	Protocol        *string
	ListenPort      *string
	DestinationPort *string
	HealthCheck     *hcLoadBalancerCreateOptsServiceHealthCheckWrap
	HTTP            *hcLoadBalancerCreateOptsServiceHTTPWrap
}

func (v *hcLoadBalancerCreateOptsServiceWrap) unwrap() (*hcloud.LoadBalancerCreateOptsService, error) {
	out := &hcloud.LoadBalancerCreateOptsService{
		// Protocol:      v.Protocol,
		// Proxyprotocol: v.Proxyprotocol,
		// HTTP:          v.HTTP,
	}
	if v.Proxyprotocol != nil {
		out.Proxyprotocol = v.Proxyprotocol
	}
	if v.Protocol != nil {
		out.Protocol = hcloud.LoadBalancerServiceProtocol(*v.Protocol)
	}
	if v.HTTP != nil {
		hhttp, err := v.HTTP.unwrap()
		if err != nil {
			return nil, err
		}
		out.HTTP = hhttp
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
	if v.HealthCheck != nil {
		hh, err := v.HealthCheck.unwrap()
		if err != nil {
			return nil, err
		}
		out.HealthCheck = hh
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
	output := &schema.LoadBalancerCreateResponse{}

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

	_, response, err := ctx.HClient.LoadBalancer.Create(context.Background(), *opts)
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
				err = ctx.WaitForAndLog(output.Action, "Waiting for load balancer %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
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
		return nil, HCloudErrResponse(err, response)
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
		hnet, response, err := ctx.HClient.Network.GetByID(context.Background(), hnetID)
		if err != nil {
			return nil, HCloudErrResponse(err, response)
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
				return nil, HCloudErrResponse(err, response)
			}
			ctx.Logger.LogInfo(fmt.Sprintf("valid ip %s found for subnet %s", opts.IP.String(), ipnet.String()))
			output := &schema.LoadBalancerActionAttachToNetworkResponse{}
			aout, err := GenericHCloudOutput(ctx, response, output)
			if err != nil {
				return nil, err
			}
			if internalparams.Waiters != nil {
				for _, wnam := range internalparams.Waiters {
					if wnam == "success" {
						err = ctx.WaitForAndLog(output.Action, "Waiting for lb attach %v%...")
						if err != nil {
							return nil, err
						}
					}
				}
			}
			return aout, err
		}
	}
	_, response, err := ctx.HClient.LoadBalancer.AttachToNetwork(context.Background(), hlb, *opts)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	output := &schema.LoadBalancerActionAttachToNetworkResponse{}
	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for lb attach %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func DetachLoadBalancerFromNetwork(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerDetachFromNetworkParameters{}
	output := &schema.LoadBalancerActionDetachFromNetworkResponse{}

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
		return nil, HCloudErrResponse(err, response)
	}

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for lb detach %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func AddTargetToLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerAddTargetParameters{}
	output := &schema.LoadBalancerActionAddTargetResponse{}

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

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for target addition %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}

func RemoveTargetFromLoadBalancer(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &loadbalancerRemoveTargetParameters{}
	output := &schema.LoadBalancerActionRemoveTargetResponse{}

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

	aout, err := GenericHCloudOutput(ctx, response, output)
	if err != nil {
		return nil, err
	}
	if internalparams.Waiters != nil {
		for _, wnam := range internalparams.Waiters {
			if wnam == "success" {
				err = ctx.WaitForAndLog(output.Action, "Waiting for target rm from lb %v%...")
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return aout, err
}
