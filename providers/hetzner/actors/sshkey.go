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
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

type hcSSHKeyWrap struct {
	hcloud.SSHKey
	ID *string `validate:"required"`
}

func (v *hcSSHKeyWrap) unwrap() (*hcloud.SSHKey, error) {
	int64id, err := strconv.ParseInt(*v.ID, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot use '%v' as int64 ID", *v.ID), err)
	}
	return &hcloud.SSHKey{ID: int64id}, nil
}

type SSHKeyListResponseWithMeta struct {
	schema.SSHKeyListResponse
	Meta schema.Meta `json:"meta"`
}

func CreateSSHKey(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	input := &hcloud.SSHKeyCreateOpts{}

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

	_, response, err := ctx.HClient.SSHKey.Create(context.Background(), *input)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	output := &schema.SSHKeyCreateResponse{}
	err = UnmarshallHCloudToSchema(response, output)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%v", output.SSHKey.ID)
	return base.NewActionOutput(ctx.Action, output, &id), nil
}

func DeleteSSHKey(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only SSHKey.ID attr is really used
	input := &hcSSHKeyWrap{}

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

	hsshkey, err := input.unwrap()
	if err != nil {
		return nil, err
	}

	response, err := ctx.HClient.SSHKey.Delete(context.Background(), hsshkey)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}

func FindSSHKeys(ctx *ActionContext) (*base.ActionOutput, error) {
	input := &hcloud.SSHKeyListOpts{}

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

	_, response, err := ctx.HClient.SSHKey.List(context.Background(), *input)
	if err != nil {
		return nil, HCloudErrResponse(err, response)
	}

	output := &SSHKeyListResponseWithMeta{}
	return GenericHCloudOutput(ctx, response, output)
}

func FindOneSSHKey(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindSSHKeys(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no ssh key found")
	}
	raw := aout.Records[0].Value.(*SSHKeyListResponseWithMeta)
	found := len(raw.SSHKeys)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no ssh key found")
	}
	output := &schema.SSHKeyGetResponse{}
	output.SSHKey = raw.SSHKeys[0]
	id := fmt.Sprintf("%v", raw.SSHKeys[0].ID)
	aout = base.NewActionOutput(ctx.Action, output, &id)
	return aout, nil
}
