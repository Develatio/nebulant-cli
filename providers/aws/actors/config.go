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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

type setRegionParameters struct {
	Region *string `validate:"required"`
}

// SetRegion func
func SetRegion(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(setRegionParameters)
	jsonErr := util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if jsonErr != nil {
		return nil, jsonErr
	}

	ctx.Logger.LogInfo("Setting new region to " + *params.Region)
	newSess := ctx.AwsSess.Copy(&aws.Config{Region: aws.String(*params.Region)})
	ctx.Store.SetPrivateVar("awsSess", newSess)

	return nil, nil
}
