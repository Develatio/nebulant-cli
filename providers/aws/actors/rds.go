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

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/develatio/nebulant-cli/base"
)

// FindDatabases func
func FindDatabases(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	awsinput := new(rds.DescribeDBInstancesInput)
	err = json.Unmarshal(ctx.Action.Parameters, awsinput)
	if err != nil {
		return nil, err
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
	err = json.Unmarshal(ctx.Action.Parameters, awsinput)
	if err != nil {
		return nil, err
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
	err = json.Unmarshal(ctx.Action.Parameters, awsinput)
	if err != nil {
		return nil, err
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
	err = json.Unmarshal(ctx.Action.Parameters, awsinput)
	if err != nil {
		return nil, err
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
