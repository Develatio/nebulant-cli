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
	"io"
	"reflect"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// ActionContext struct
type ActionContext struct {
	Rehearsal bool
	HClient   *hcloud.Client
	Action    *blueprint.Action
	Store     base.IStore
	Logger    base.ILogger
}

// CleanInput func unmarshall action.Parameters into input interface{}
// and calls input.Validate() methods if exists
func CleanInput(action *blueprint.Action, input interface{}) error {
	err := util.UnmarshalValidJSON(action.Parameters, input)
	if err != nil {
		return err
	}
	vvv := reflect.TypeOf(input)
	_, validable := vvv.MethodByName("Validate")
	if validable {
		ret := reflect.ValueOf(input).MethodByName("Validate").Call([]reflect.Value{})
		switch ret[0].Interface().(type) {
		case error:
			return ret[0].Interface().(error)
		}
	}
	return nil
}

func UnmarshallHCloudToSchema(response *hcloud.Response, v interface{}) error {
	var body []byte
	body, err := io.ReadAll(response.Response.Body)
	if err != nil {
		return err
	}
	err = util.UnmarshalValidJSON(body, v)
	if err != nil {
		return err
	}
	return nil
}

var NewActionContext = func(client *hcloud.Client, action *blueprint.Action, store base.IStore, logger base.ILogger) *ActionContext {
	l := logger.Duplicate()
	l.SetActionID(action.ActionID)
	return &ActionContext{
		HClient: client,
		Action:  action,
		Store:   store,
		Logger:  l,
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
	"create_floating_ip": {F: CreateFloatingIP, N: NextOKKO},
}
