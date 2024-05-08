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
	"regexp"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcImageWrap struct {
	hcloud.Image
	ID *string `validate:"required"`
}

func (v *hcImageWrap) unwrap() (*hcloud.Image, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.Image{ID: int64id}, nil
}

type ImageListResponseWithMeta struct {
	schema.ImageListResponse
	Meta   schema.Meta    `json:"meta"`
	Images []schema.Image `json:"images"`
}

type hcImageListOptsWrap struct {
	hcloud.ImageListOpts
	ID          *string `json:"id"` // for GetByID
	Description *string `json:"description"`
}

func (v *hcImageListOptsWrap) unwrap() (*hcloud.ImageListOpts, error) {
	return &hcloud.ImageListOpts{
		Type:              v.Type,
		BoundTo:           v.BoundTo,
		Name:              v.Name,
		Sort:              v.Sort,
		Status:            v.Status,
		IncludeDeprecated: v.IncludeDeprecated,
		Architecture:      v.Architecture,
	}, nil
}

func DeleteImage(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only Image.ID attr are really used
	input := &hcImageWrap{}

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

	himg, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	response, err := ctx.HClient.Image.Delete(context.Background(), himg)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	// delete returns scheme like {image:{}}, same as update
	// so here is OK to use update response as delete response
	output := &schema.ImageUpdateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindImages(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcImageListOptsWrap{}

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

	var response *hcloud.Response
	if input.Description != nil {
		var images []schema.Image
		opts.Page = 1     // min allowed (0 means no page)
		opts.PerPage = 50 // max allowed
		r := regexp.MustCompile(`(?i)` + *input.Description + ``)
		for {
			_, _rsp, err := ctx.HClient.Image.List(context.Background(), *opts)
			if err != nil {
				return nil, HCloudErrResponse(err, _rsp)
			}
			_v := &ImageListResponseWithMeta{}
			err = UnmarshallHCloudToSchema(_rsp, _v)
			if err != nil {
				return nil, err
			}
			if len(_v.Images) <= 0 {
				break
			}
			for _, im := range _v.Images {
				if r.MatchString(im.Description) {
					images = append(images, im)
				}
			}
			if _rsp.Meta.Pagination.NextPage == 0 {
				break
			}
			opts.Page = _rsp.Meta.Pagination.NextPage
		}
		output := &ImageListResponseWithMeta{
			Images: images,
			Meta: schema.Meta{
				Pagination: &schema.MetaPagination{
					Page:         1,
					LastPage:     1,
					NextPage:     0,
					TotalEntries: len(images),
				},
			},
		}
		return base.NewActionOutput(ctx.Action, output, nil), nil
	}

	if input.ID != nil {
		int64id, err := strconv.ParseInt(*input.ID, 10, 64)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *input.ID), err)
		}
		_, response, err = ctx.HClient.Image.GetByID(context.Background(), int64id)
		if err != nil {
			return nil, HCloudErrResponse(err, response)
		}
		imgsch := &schema.ImageGetResponse{}
		err = UnmarshallHCloudToSchema(response, imgsch)
		if err != nil {
			return nil, err
		}
		output := &ImageListResponseWithMeta{
			Images: []schema.Image{imgsch.Image},
			Meta: schema.Meta{
				Pagination: &schema.MetaPagination{
					Page:         1,
					LastPage:     1,
					NextPage:     0,
					TotalEntries: 1,
				},
			},
		}
		return base.NewActionOutput(ctx.Action, output, nil), nil
	}

	// normal list
	_, response, err = ctx.HClient.Image.List(context.Background(), *opts)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}
	return GenericHCloudOutput(ctx, response, &ImageListResponseWithMeta{})
}

func FindOneImage(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindImages(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no image found (E1)")
	}
	raw := aout.Records[0].Value.(*ImageListResponseWithMeta)
	found := len(raw.Images)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no image found (E2)")
	}
	output := &schema.ImageGetResponse{}
	output.Image = raw.Images[0]
	id := fmt.Sprintf("%v", raw.Images[0].ID)
	aout = base.NewActionOutput(ctx.Action, output, &id)
	return aout, nil
}
