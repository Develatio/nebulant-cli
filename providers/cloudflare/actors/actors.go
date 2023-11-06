// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
)

type s3ClientFunc func() *s3.Client

// ActionContext struct
type ActionContext struct {
	Rehearsal           bool
	AwsConfig           aws.Config
	Action              *blueprint.Action
	Store               base.IStore
	Logger              base.ILogger
	NewS3Client         s3ClientFunc
	NewCloudFlareClient interface{}
}

var NewActionContext = func(awsConf aws.Config, action *blueprint.Action, store base.IStore, logger base.ILogger) *ActionContext {
	l := logger.Duplicate()
	l.SetActionID(action.ActionID)
	return &ActionContext{
		AwsConfig: awsConf,
		Action:    action,
		Store:     store,
		Logger:    l,
		NewS3Client: func() *s3.Client {
			return s3.NewFromConfig(awsConf)
		},
		NewCloudFlareClient: nil,
	}
}

// ActionFunc func
type ActionFunc func(ctx *ActionContext) (*base.ActionOutput, error)

type NextType int

const (
	// NextOKKO const 0
	NextOKKO NextType = iota
	// NextOK const 1
	NextOK
	// NextKO const
	NextKO
)

type ActionLayout struct {
	F ActionFunc
	N NextType
}

// ActionFuncMap map
var ActionFuncMap map[string]*ActionLayout = map[string]*ActionLayout{
	"r2_upload": {F: R2Upload, N: NextOKKO},
}
