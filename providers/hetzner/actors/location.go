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

type LocationListResponseWithMeta struct {
	schema.LocationListResponse
	Meta schema.Meta `json:"meta"`
}

func FindLocations(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.LocationListOpts{}

	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, input); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	_, response, err := ctx.HClient.Location.List(context.Background(), *input)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	output := &LocationListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneLocation(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindLocations(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no location found")
	}
	raw := aout.Records[0].Value.(*LocationListResponseWithMeta)
	found := len(raw.Locations)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no location found")
	}
	output := &schema.LocationGetResponse{}
	output.Location = raw.Locations[0]
	id := fmt.Sprintf("%v", raw.Locations[0].ID)
	aout = base.NewActionOutput(ctx.Action, output, &id)
	return aout, nil
}
