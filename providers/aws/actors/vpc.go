// Nebulant
// Copyright (C) 2021  Develatio Technologies S.L.

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
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
)

// FindVpcs func
func FindVpcs(ctx *ActionContext) (*base.ActionOutput, error) {
	awsinput := new(ec2.DescribeVpcsInput)
	if err := CleanInput(ctx.Action, awsinput); err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	region := ctx.AwsSess.Config.Region
	ctx.Logger.LogInfo("Looking for vpcs in region " + *region)

	svc := ctx.NewEC2Client()
	result, err := svc.DescribeVpcs(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneVpc func
func FindOneVpc(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindVpcs(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no VPC found")
	}
	raw := aout.Records[0].Value.(*ec2.DescribeVpcsOutput)
	found := len(raw.Vpcs)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no VPC found")
	}
	aout = base.NewActionOutput(ctx.Action, raw.Vpcs[0], raw.Vpcs[0].VpcId)
	return aout, nil
}

// DeleteVpc func
func DeleteVpc(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DeleteVpcInput)
	if err := CleanInput(ctx.Action, awsinput); err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(awsinput)
	if err != nil {
		return nil, err
	}

	svc := ctx.NewEC2Client()
	result, err := svc.DeleteVpc(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}
