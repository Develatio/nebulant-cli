// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

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
	"github.com/develatio/nebulant-cli/util"
)

// FindNetworkInterfaces func
func FindNetworkInterfaces(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeNetworkInterfacesInput)
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
	outresult, err := svc.DescribeNetworkInterfaces(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, outresult, nil)
	return aout, nil
}

// FindNetworkInterface func
func FindNetworkInterface(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindNetworkInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no interface found")
	}
	raw := aout.Records[0].Value.(*ec2.DescribeNetworkInterfacesOutput)
	found := len(raw.NetworkInterfaces)
	if found > 1 {
		return nil, fmt.Errorf("too many interface results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no interface found")
	}
	aout = base.NewActionOutput(ctx.Action, raw.NetworkInterfaces[0], raw.NetworkInterfaces[0].NetworkInterfaceId)
	return aout, nil
}

// DeleteNetworkInterface func
func DeleteNetworkInterface(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DeleteNetworkInterfaceInput)
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
	result, err := svc.DeleteNetworkInterface(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}
