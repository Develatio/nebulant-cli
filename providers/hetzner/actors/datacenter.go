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
	"fmt"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type DatacenterListResponseWithMeta struct {
	schema.DatacenterListResponse
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
		return nil, HCloudErrResponse(err, response)
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
