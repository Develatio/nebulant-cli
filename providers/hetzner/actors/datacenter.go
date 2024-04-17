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

type DatacenterListResponseWithMeta struct {
	*schema.DatacenterListResponse
	Meta schema.Meta `json:"meta"`
}

func FindDatacenters(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.DatacenterListOpts{}

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
	_, response, err := ctx.HClient.Datacenter.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &DatacenterListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneDatacenter(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindDatacenters(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no datacenter found")
	}
	raw := aout.Records[0].Value.(*schema.DatacenterListResponse)
	found := len(raw.Datacenters)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no datacenter found")
	}
	id := fmt.Sprintf("%v", raw.Datacenters[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.Datacenters[0], &id)
	return aout, nil
}
