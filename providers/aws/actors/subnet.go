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

// FindSubnets func
func FindSubnets(ctx *ActionContext) (*base.ActionOutput, error) {
	awsinput := new(ec2.DescribeSubnetsInput)
	if err := CleanInput(ctx.Action, awsinput); err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	region := ctx.AwsSess.Config.Region
	ctx.Logger.LogInfo("Looking for subnets in region " + *region)

	svc := ctx.NewEC2Client()
	result, err := svc.DescribeSubnets(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneSubnet func
func FindOneSubnet(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindSubnets(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no subnet found")
	}
	raw := aout.Records[0].Value.(*ec2.DescribeSubnetsOutput)
	found := len(raw.Subnets)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no subnet found")
	}
	aout = base.NewActionOutput(ctx.Action, raw.Subnets[0], raw.Subnets[0].SubnetId)
	return aout, nil
}

// DeleteSubnet func
func DeleteSubnet(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DeleteSubnetInput)
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
	result, err := svc.DeleteSubnet(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}
