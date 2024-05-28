// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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
