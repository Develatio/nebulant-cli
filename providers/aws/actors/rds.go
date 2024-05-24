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

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

// FindDatabases func
func FindDatabases(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(rds.DescribeDBInstancesInput)
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

	svc := rds.New(ctx.AwsSess)
	result, err := svc.DescribeDBInstances(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, nil
}

// FindOneDatabase func
func FindOneDatabase(ctx *ActionContext) (*base.ActionOutput, error) {
	aout, err := FindDatabases(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if len(aout.Records) <= 0 {
		return nil, fmt.Errorf("no database found")
	}
	raw := aout.Records[0].Value.(*rds.DescribeDBInstancesOutput)
	found := len(raw.DBInstances)
	if found > 1 {
		return nil, fmt.Errorf("too many results")
	}
	if found <= 0 {
		return nil, fmt.Errorf("no database found")
	}
	aout = base.NewActionOutput(ctx.Action, raw.DBInstances[0], raw.DBInstances[0].DBInstanceIdentifier)
	return aout, nil
}

// CreateDatabase func
func CreateDatabase(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(rds.CreateDBInstanceInput)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, awsinput); err != nil {
		return nil, err
	}
	if ctx.Rehearsal {
		return nil, nil
	}

	svc := rds.New(ctx.AwsSess)
	result, err := svc.CreateDBInstance(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, result.DBInstance, result.DBInstance.DBInstanceIdentifier)
	return aout, nil
}

// DeleteDatabase func
func DeleteDatabase(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(rds.DeleteDBInstanceInput)
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

	svc := rds.New(ctx.AwsSess)
	_, err = svc.DeleteDBInstance(awsinput)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// CreateSnapshotDatabase func
func CreateSnapshotDatabase(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(rds.CreateDBSnapshotInput)
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

	svc := rds.New(ctx.AwsSess)
	awsout, err := svc.CreateDBSnapshot(awsinput)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, awsout.DBSnapshot, awsout.DBSnapshot.DBInstanceIdentifier)
	return aout, nil
}

// RestoreSnapshotDatabase func
func RestoreSnapshotDatabase(ctx *ActionContext) (*base.ActionOutput, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}
