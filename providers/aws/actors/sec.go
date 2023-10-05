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

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

// FindSecurityGroups func
func FindSecurityGroups(ctx *ActionContext) (*base.ActionOutput, error) {
	awsinput := new(ec2.DescribeSecurityGroupsInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	region := ctx.AwsSess.Config.Region
	ctx.Logger.LogInfo("Looking for seg in region " + *region)

	svc := ctx.NewEC2Client()
	result, err := svc.DescribeSecurityGroups(awsinput)
	if err != nil {
		aerr := err.(awserr.Error)
		if aerr.Code() == "InvalidGroup.NotFound" {
			result = &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: nil,
			}
		} else {
			return nil, err
		}
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneSecurityGroup func
func FindOneSecurityGroup(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindSecurityGroups(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("security Group Not Found")
	}
	raw := aout.Records[0].Value.(*ec2.DescribeSecurityGroupsOutput)
	// none results not allowed
	if raw.SecurityGroups == nil {
		return nil, fmt.Errorf("not Found")
	}
	resultCount := len(raw.SecurityGroups)
	// no zero allowed
	if resultCount <= 0 {
		return nil, fmt.Errorf("security Group Not Found")
	}
	// only one result allowed
	if resultCount > 1 {
		return nil, fmt.Errorf("too many results")
	}
	aout = base.NewActionOutput(ctx.Action, raw.SecurityGroups[0], raw.SecurityGroups[0].GroupId)
	return aout, nil
}

// DeleteSecurityGroup func
func DeleteSecurityGroup(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DeleteSecurityGroupInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
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
	result, err := svc.DeleteSecurityGroup(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}
