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
	"io"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// ActionContext struct
type ActionContext struct {
	Rehearsal bool
	HClient   *hcloud.Client
	Action    *base.Action
	Store     base.IStore
	Logger    base.ILogger
}

func (a *ActionContext) WaitForAndLog(action schema.Action, msg string) error {
	act := hcloud.ActionFromSchema(action)
	a.Logger.LogDebug(fmt.Sprintf("waiting for action %v", act.ID))
	okCh, errCh := a.HClient.Action.WatchProgress(context.Background(), act)
	var err error
	noprogress_msg := msg + " ... "
	progress_msg := msg + " (%v%%...) "

	errCount := 0
L:
	for {
		select {
		case progress := <-okCh:
			if progress == 0 {
				a.Logger.LogInfo(fmt.Sprint(noprogress_msg))
			} else {
				a.Logger.LogInfo(fmt.Sprintf(progress_msg, progress))
			}
		case err = <-errCh:
			// sometimes hc api ret err even on non
			// failing event try to retry 5 times
			// and exit after all if err persist
			if err != nil {
				// if api did not return any action, maybe the action has finished
				if strings.HasSuffix(err.Error(), "action not returned from API") {
					a.Logger.LogDebug(err.Error())
					err = nil
					break L
				}
				if errCount < 5 {
					errCount++
					time.Sleep(3 * time.Second)
					okCh, errCh = a.HClient.Action.WatchProgress(context.Background(), act)
					continue
				}
				// retry count end, let err fly as free bird
			}
			// on sucess, err is nil
			break L
		}
	}
	return err
}

func (a *ActionContext) WaitForManyAndLog(actions []schema.Action, msg string) error {
	act := hcloud.ActionsFromSchema(actions)
	for _, aa := range act {
		a.Logger.LogDebug(fmt.Sprintf("waiting for action %v", aa.ID))
	}
	okCh, errCh := a.HClient.Action.WatchOverallProgress(context.Background(), act)
	var err error
	noprogress_msg := msg + " ... "
	progress_msg := msg + " (%v%%...) "

	errCount := 0
L:
	for {
		select {
		case progress := <-okCh:
			if progress == 0 {
				a.Logger.LogInfo(fmt.Sprint(noprogress_msg))
			} else {
				a.Logger.LogInfo(fmt.Sprintf(progress_msg, progress))
			}
		case err = <-errCh:
			// sometimes hc api ret err even on non
			// failing event try to retry 5 times
			// and exit after all if err persist
			if err != nil {
				// if api did not return any action, maybe the action has finished
				if strings.HasSuffix(err.Error(), "action not returned from API") {
					a.Logger.LogDebug(err.Error())
					err = nil
					break L
				}
				if errCount < 5 {
					errCount++
					time.Sleep(3 * time.Second)
					okCh, errCh = a.HClient.Action.WatchOverallProgress(context.Background(), act)
					continue
				}
				// retry count end, let err fly as free bird
			}
			// on sucess, err is nil
			break L
		}
	}
	return err
}

func UnmarshallHCloudToSchema(response *hcloud.Response, v interface{}) error {
	var body []byte
	body, err := io.ReadAll(response.Response.Body)
	if err != nil {
		return err
	}
	err = util.UnmarshalValidJSON(body, v)
	if err != nil {
		return err
	}
	return nil
}

var NewActionContext = func(client *hcloud.Client, action *base.Action, store base.IStore, logger base.ILogger) *ActionContext {
	l := logger.Duplicate()
	l.SetActionID(action.ActionID)
	l.SetActionName(action.ActionName)
	return &ActionContext{
		HClient: client,
		Action:  action,
		Store:   store,
		Logger:  l,
	}
}

// ActionFunc func
type ActionFunc func(ctx *ActionContext) (*base.ActionOutput, error)

type NextType int

const (
	// NextOKKO const 0
	NextOKKO NextType = iota
	// NextOK const 1
	NextOK
	// NextKO const
	NextKO
)

type ActionLayout struct {
	F ActionFunc
	N NextType
}

// ActionFuncMap map
var ActionFuncMap map[string]*ActionLayout = map[string]*ActionLayout{
	"create_floating_ip":   {F: CreateFloatingIP, N: NextOKKO},
	"delete_floating_ip":   {F: DeleteFloatingIP, N: NextOKKO},
	"find_floating_ips":    {F: FindFloatingIPs, N: NextOKKO},
	"findone_floating_ip":  {F: FindOneFloatingIP, N: NextOKKO},
	"assign_floating_ip":   {F: AssignFloatingIP, N: NextOKKO},
	"unassign_floating_ip": {F: UnassignFloatingIP, N: NextOKKO},

	"find_images":   {F: FindImages, N: NextOKKO},
	"findone_image": {F: FindOneImage, N: NextOKKO},
	"delete_image":  {F: DeleteImage, N: NextOKKO},

	"create_server":              {F: CreateServer, N: NextOKKO},
	"delete_server":              {F: DeleteServer, N: NextOKKO},
	"find_servers":               {F: FindServers, N: NextOKKO},
	"findone_server":             {F: FindOneServer, N: NextOKKO},
	"start_server":               {F: PowerOnServer, N: NextOKKO},  // poweron server
	"stop_server":                {F: PowerOffServer, N: NextOKKO}, // poweroff server
	"attach_server_to_network":   {F: AttachServerToNetwork, N: NextOKKO},
	"detach_server_from_network": {F: DetachServerFromNetwork, N: NextOKKO},
	"create_image_from_server":   {F: CreateImageFromServer, N: NextOKKO},

	"create_network":             {F: CreateNetwork, N: NextOKKO},
	"delete_network":             {F: DeleteNetwork, N: NextOKKO},
	"find_networks":              {F: FindNetworks, N: NextOKKO},
	"findone_network":            {F: FindOneNetwork, N: NextOKKO},
	"add_subnet_to_network":      {F: AddSubnetToNetwork, N: NextOKKO},
	"delete_subnet_from_network": {F: DeleteSubnetFromNetwork, N: NextOKKO},
	"add_route_to_network":       {F: AddRouteToNetwork, N: NextOKKO},
	"delete_route_from_network":  {F: DeleteRouteFromNetwork, N: NextOKKO},

	"create_volume":  {F: CreateVolume, N: NextOKKO},
	"delete_volume":  {F: DeleteVolume, N: NextOKKO},
	"find_volumes":   {F: FindVolumes, N: NextOKKO},
	"findone_volume": {F: FindOneVolume, N: NextOKKO},
	"attach_volume":  {F: AttachVolume, N: NextOKKO},
	"detach_volume":  {F: DetachVolume, N: NextOKKO},

	"find_datacenters":   {F: FindDatacenters, N: NextOKKO},
	"findone_datacenter": {F: FindOneDatacenter, N: NextOKKO},

	"create_firewall":                {F: CreateFirewall, N: NextOKKO},
	"delete_firewall":                {F: DeleteFirewall, N: NextOKKO},
	"find_firewalls":                 {F: FindFirewalls, N: NextOKKO},
	"findone_firewall":               {F: FindOneFirewall, N: NextOKKO},
	"apply_firewall_to_resources":    {F: ApplyFirewallToResources, N: NextOKKO},
	"remove_firewall_from_resources": {F: RemoveFirewallFromResources, N: NextOKKO},
	"set_rules_firewall":             {F: SetRulesFirewall, N: NextOKKO},

	"find_isos":   {F: FindISOs, N: NextOKKO},
	"findone_iso": {F: FindOneISO, N: NextOKKO},

	"create_load_balancer":               {F: CreateLoadBalancer, N: NextOKKO},
	"delete_load_balancer":               {F: DeleteLoadBalancer, N: NextOKKO},
	"find_load_balancers":                {F: FindLoadBalancers, N: NextOKKO},
	"findone_load_balancer":              {F: FindOneLoadBalancer, N: NextOKKO},
	"attach_load_balancer_to_network":    {F: AttachLoadBalancerToNetwork, N: NextOKKO},
	"dettach_load_balancer_from_network": {F: DetachLoadBalancerFromNetwork, N: NextOKKO},
	"add_target_to_load_balancer":        {F: AddTargetToLoadBalancer, N: NextOKKO},
	"remove_target_from_load_balancer":   {F: RemoveTargetFromLoadBalancer, N: NextOKKO},
	"add_service_to_load_balancer":       {F: AddServiceToLoadBalancer, N: NextOKKO},
	"delete_service_from_load_balancer":  {F: DeleteServiceFromLoadBalancer, N: NextOKKO},

	"find_locations":   {F: FindLocations, N: NextOKKO},
	"findone_location": {F: FindOneLocation, N: NextOKKO},

	"create_primary_ip":   {F: CreatePrimaryIP, N: NextOKKO},
	"delete_primary_ip":   {F: DeletePrimaryIP, N: NextOKKO},
	"find_primary_ips":    {F: FindPrimaryIPs, N: NextOKKO},
	"findone_primary_ip":  {F: FindOnePrimaryIP, N: NextOKKO},
	"assign_primary_ip":   {F: AssignPrimaryIP, N: NextOKKO},
	"unassign_primary_ip": {F: UnassignPrimaryIP, N: NextOKKO},

	"create_ssh_key":  {F: CreateSSHKey, N: NextOKKO},
	"delete_ssh_key":  {F: DeleteSSHKey, N: NextOKKO},
	"find_ssh_keys":   {F: FindSSHKeys, N: NextOKKO},
	"findone_ssh_key": {F: FindOneSSHKey, N: NextOKKO},
}

// GenericHCloudOutput unmarshall response into v and return ActionContext with
// the result
func GenericHCloudOutput(ctx *ActionContext, response *hcloud.Response, v interface{}) (*base.ActionOutput, error) {
	err := UnmarshallHCloudToSchema(response, v)
	if err != nil {
		return nil, err
	}
	aout := base.NewActionOutput(ctx.Action, v, nil)
	return aout, nil
}

func HCloudErrResponse(err error, response *hcloud.Response) error {
	if response == nil {
		return errors.Join(err, fmt.Errorf("no api response"))
	}
	eeb, err1 := io.ReadAll(response.Body)
	if err1 != nil {
		return errors.Join(err, err1)
	}
	return errors.Join(err, fmt.Errorf(string(eeb)))
}
