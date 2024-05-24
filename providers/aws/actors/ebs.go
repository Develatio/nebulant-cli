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
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
)

// FindVolumes func
func FindVolumes(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeVolumesInput)
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
	ctx.Logger.LogInfo("Looking for volumes in region " + *region)

	svc := ctx.NewEC2Client()
	result, err := svc.DescribeVolumes(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneVolume func
func FindOneVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindVolumes(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	raw := aout.Records[0].Value.(*ec2.DescribeVolumesOutput)
	found := len(raw.Volumes)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no volume found")
	}

	aout = base.NewActionOutput(ctx.Action, raw.Volumes[0], raw.Volumes[0].VolumeId)
	return aout, nil
}

// CreateVolume func
func CreateVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	awsinput := new(ec2.CreateVolumeInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
		return nil, err
	}
	internalparams := new(blueprint.InternalParameters)
	err := json.Unmarshal(ctx.Action.Parameters, internalparams)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	// awsinput.AvailabilityZone can be nil because its a pointer
	if awsinput.AvailabilityZone == nil {
		defaultAvZone := *ctx.AwsSess.Config.Region + "a"
		awsinput.AvailabilityZone = &defaultAvZone
	}

	svc := ctx.NewEC2Client()
	vol, err := svc.CreateVolume(awsinput) // ec2.Volume
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			return nil, err
		}
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{vol.VolumeId},
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilVolumeAvailable":
				ctx.Logger.LogInfo("Waiting for volume to be available...")
				err = svc.WaitUntilVolumeAvailable(waitinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	aout := base.NewActionOutput(ctx.Action, vol, vol.VolumeId)
	return aout, nil
}

// AttachVolume func
func AttachVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	// perform attach
	var err error
	svc := ctx.NewEC2Client()

	awsinput := new(ec2.AttachVolumeInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
		return nil, err
	}
	internalparams := new(blueprint.InternalParameters)
	err = json.Unmarshal(ctx.Action.Parameters, internalparams)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(awsinput)
	if err != nil {
		return nil, err
	}

	result, err := svc.AttachVolume(awsinput)
	if err != nil {
		return nil, err
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{awsinput.VolumeId},
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilVolumeInUse":
				ctx.Logger.LogInfo("Waiting for volume to be attached...")
				err = svc.WaitUntilVolumeInUse(waitinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	// Return nil, nil?
	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// DetachVolume func
func DetachVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DetachVolumeInput)
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
	result, err := svc.DetachVolume(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// DeleteVolume func
func DeleteVolume(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DeleteVolumeInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
		return nil, err
	}
	internalparams := new(blueprint.InternalParameters)
	err = json.Unmarshal(ctx.Action.Parameters, internalparams)
	if err != nil {
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
	_, err = svc.DeleteVolume(awsinput)
	if err != nil {
		return nil, err
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{awsinput.VolumeId},
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilVolumeDeleted":
				ctx.Logger.LogInfo("Waiting for volume to be deleted...")
				err = svc.WaitUntilVolumeDeleted(waitinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	return nil, nil
}
