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
		return nil, err
	}

	output := &schema.SSHKeyCreateResponse{}
	return GenericHCloudOutput(ctx, response, output)
}

func DeleteSSHKey(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	// only SSHKey.ID attr is really used
	input := &hcloud.SSHKey{}

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

	_, err = ctx.HClient.SSHKey.Delete(context.Background(), input)
	if err != nil {
		return nil, err
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

	_, response, err := ctx.HClient.SSHKey.List(context.Background(), *input)
	if err != nil {
		return nil, err
	}

	output := &schema.SSHKeyListResponse{}
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
	raw := aout.Records[0].Value.(*schema.SSHKeyListResponse)
	found := len(raw.SSHKeys)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no ssh key found")
	}
	id := fmt.Sprintf("%v", raw.SSHKeys[0].ID)
	aout = base.NewActionOutput(ctx.Action, raw.SSHKeys[0], &id)
	return aout, nil
}
