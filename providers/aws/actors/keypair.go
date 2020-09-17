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
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
)

// FindKeyPairs func
func FindKeyPairs(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeKeyPairsInput)
	err = json.Unmarshal(ctx.Action.Parameters, awsinput)
	if err != nil {
		return nil, err
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
	err = json.Unmarshal(ctx.Action.Parameters, awsinput)
	if err != nil {
		return nil, err
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
