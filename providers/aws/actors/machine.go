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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
)

// RunInstance func
func RunInstance(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.RunInstancesInput)
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
	// force only one
	awsinput.MaxCount = aws.Int64(1)
	awsinput.MinCount = aws.Int64(1)
	runResult, err := svc.RunInstances(awsinput)
	if err != nil {
		return nil, err
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeInstancesInput{}
		waitstatusinput := &ec2.DescribeInstanceStatusInput{}
		// empty the array
		waitinput.InstanceIds = nil
		waitstatusinput.InstanceIds = nil
		// INFO: here the code waits for all created instances, so multi
		// instance are handled
		for _, instance := range runResult.Instances {
			waitinput.InstanceIds = append(waitinput.InstanceIds, instance.InstanceId)
			waitstatusinput.InstanceIds = append(waitinput.InstanceIds, instance.InstanceId)
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilInstanceRunning":
				ctx.Logger.LogInfo("Waiting for instances to become ready...")
				err = svc.WaitUntilInstanceRunning(waitinput)
				if err != nil {
					return nil, err
				}
			case "WaitUntilInstanceStatusOk":
				ctx.Logger.LogInfo("Waiting for instances to become status OK...")
				err = svc.WaitUntilInstanceStatusOk(waitstatusinput)
				if err != nil {
					return nil, err
				}
			case "WaitUntilInstanceExists":
				ctx.Logger.LogInfo("Waiting for instances to exist...")
				err = svc.WaitUntilInstanceExists(waitinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	aout := base.NewActionOutput(
		ctx.Action,
		runResult.Instances[0],
		runResult.Instances[0].InstanceId,
	)
	return aout, nil
}

// DeleteInstance func
func DeleteInstance(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.TerminateInstancesInput)
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
	terminateResult, err := svc.TerminateInstances(awsinput)
	if err != nil {
		return nil, err
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeInstancesInput{
			InstanceIds: awsinput.InstanceIds,
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilInstanceTerminated":
				ctx.Logger.LogInfo("Waiting for instances to be terminated...")
				err = svc.WaitUntilInstanceTerminated(waitinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	aout := base.NewActionOutput(ctx.Action, terminateResult, nil)
	return aout, nil
}

// StopInstance func
func StopInstance(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.StopInstancesInput)
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
	stopResult, err := svc.StopInstances(awsinput)
	if err != nil {
		return nil, err
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeInstancesInput{
			InstanceIds: awsinput.InstanceIds,
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilInstanceStopped":
				ctx.Logger.LogInfo("Waiting for instances to be stopped...")
				err = svc.WaitUntilInstanceStopped(waitinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	aout := base.NewActionOutput(ctx.Action, stopResult, nil)
	return aout, nil
}

// StartInstance func
func StartInstance(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.StartInstancesInput)
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
	startResult, err := svc.StartInstances(awsinput)
	if err != nil {
		return nil, err
	}

	if internalparams.Waiters != nil {
		waitinput := &ec2.DescribeInstancesInput{
			InstanceIds: awsinput.InstanceIds,
		}
		waitstatusinput := &ec2.DescribeInstanceStatusInput{
			InstanceIds: awsinput.InstanceIds,
		}
		for _, waitername := range internalparams.Waiters {
			switch waitername {
			case "WaitUntilInstanceRunning":
				ctx.Logger.LogInfo("Waiting for instances to become ready ....")
				err = svc.WaitUntilInstanceRunning(waitinput)
				if err != nil {
					return nil, err
				}
			case "WaitUntilInstanceStatusOk":
				ctx.Logger.LogInfo("Waiting for instances to become status OK ....")
				err = svc.WaitUntilInstanceStatusOk(waitstatusinput)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unkown waiter")
			}
		}
	}

	aout := base.NewActionOutput(ctx.Action, startResult, nil)
	return aout, nil
}

// FindInstances func
func FindInstances(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(ec2.DescribeInstancesInput)
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
	result, err := svc.DescribeInstances(awsinput)

	if err != nil {
		return nil, err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			ctx.Logger.LogInfo("Found Instance " + *instance.InstanceId + " with status " + *instance.State.Name)
		}
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneInstance func
func FindOneInstance(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindInstances(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no machine found")
	}
	found := 0
	var first *ec2.Instance
	for _, reservation := range aout.Records[0].Value.(*ec2.DescribeInstancesOutput).Reservations {
		for _, instance := range reservation.Instances {
			if found == 0 {
				first = instance
			}
			found++
		}
	}
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no machine found")
	}
	aout = base.NewActionOutput(ctx.Action, first, first.InstanceId)
	return aout, nil
}
