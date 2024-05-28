// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

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
