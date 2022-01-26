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
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
)

// FindVolumes func
func FindVolumes(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeVolumesInput)
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
	if err := CleanInput(ctx.Action, awsinput); err != nil {
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
	if err := CleanInput(ctx.Action, awsinput); err != nil {
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
	if err := CleanInput(ctx.Action, awsinput); err != nil {
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
