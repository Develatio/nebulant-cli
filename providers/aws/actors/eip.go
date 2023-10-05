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

// AllocateAddress func
func AllocateAddress(ctx *ActionContext) (*base.ActionOutput, error) {
	awsinput := new(ec2.AllocateAddressInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	svc := ctx.NewEC2Client()
	result, err := svc.AllocateAddress(awsinput)
	if err != nil {
		return nil, err
	}

	// The result should be ec2.Address to maintain
	// compatibility.
	// Here the IP is unknown but will be captured later
	addr := &ec2.Address{}
	dcErr := util.DeepCopy(result, addr)
	if dcErr != nil {
		return nil, dcErr
	}
	aout := base.NewActionOutput(ctx.Action, addr, result.AllocationId)

	return aout, nil
}

// FindAddresses func
func FindAddresses(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeAddressesInput)
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
	result, err := svc.DescribeAddresses(awsinput)
	if err != nil {
		return nil, err
	}
	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneAddress func
func FindOneAddress(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindAddresses(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no address found")
	}
	raw := aout.Records[0].Value.(*ec2.DescribeAddressesOutput)
	found := len(raw.Addresses)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no address found")
	}
	aout = base.NewActionOutput(ctx.Action, raw.Addresses[0], raw.Addresses[0].AllocationId)
	return aout, nil
}

// AttachAddress func
func AttachAddress(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.AssociateAddressInput)
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
	result, err := svc.AssociateAddress(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, result.AssociationId)
	return aout, nil
}

// DetachAddress func
// ec2.DisassociateAddressInput
//
// AssociationId *string `type:"string"`
// DryRun *bool `locationName:"dryRun" type:"boolean"`
// PublicIp *string `type:"string"`
func DetachAddress(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DisassociateAddressInput)
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
	result, err := svc.DisassociateAddress(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// ReleaseAddress func
func ReleaseAddress(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.ReleaseAddressInput)
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
	result, err := svc.ReleaseAddress(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}
