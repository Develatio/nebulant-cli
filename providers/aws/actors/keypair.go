// MIT License
//
// Copyright (C) 2021  Develatio Technologies S.L.

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

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

// FindKeyPairs func
func FindKeyPairs(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeKeyPairsInput)
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

	region := ctx.AwsSess.Config.Region
	ctx.Logger.LogInfo("Looking for key pairs in region " + *region)

	svc := ctx.NewEC2Client()
	result, err := svc.DescribeKeyPairs(awsinput)
	if err != nil {
		aerr := err.(awserr.Error)
		if aerr.Code() == "InvalidKeyPair.NotFound" {
			result = &ec2.DescribeKeyPairsOutput{
				KeyPairs: nil,
			}
		} else {
			return nil, err
		}
	}
	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneKeyPair func
func FindOneKeyPair(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindKeyPairs(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("key Pair Not Found")
	}
	raw := aout.Records[0].Value.(*ec2.DescribeKeyPairsOutput)
	// none results not allowed
	if raw.KeyPairs == nil {
		return nil, fmt.Errorf("not Found")
	}
	resultCount := len(raw.KeyPairs)
	// no zero allowed
	if resultCount <= 0 {
		return nil, fmt.Errorf("key Pair Not Found")
	}
	// only one result allowed
	if resultCount > 1 {
		return nil, fmt.Errorf("too many results")
	}
	aout = base.NewActionOutput(ctx.Action, raw.KeyPairs[0], raw.KeyPairs[0].KeyPairId)
	return aout, nil
}

// DeleteKeyPair func
func DeleteKeyPair(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DeleteKeyPairInput)
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
	result, err := svc.DeleteKeyPair(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}
