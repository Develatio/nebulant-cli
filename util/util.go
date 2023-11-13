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

package util

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"runtime"

	"github.com/go-playground/validator/v10"
)

// UIStore interface
type UIStore interface {
	GetValueByReference(reference *string) (interface{}, error)
}

// UnmarshalValidJSON func
func UnmarshalValidJSON(data []byte, v interface{}) error {
	jsonErr := json.Unmarshal(data, v)
	if jsonErr != nil {
		return jsonErr
	}

	// json tag validator
	validate := validator.New()
	vErr := validate.Struct(v)
	if vErr != nil {
		return vErr
	}

	// Validate()
	vvv := reflect.TypeOf(v)
	_, validable := vvv.MethodByName("Validate")
	if validable {
		ret := reflect.ValueOf(v).MethodByName("Validate").Call([]reflect.Value{})
		switch ret[0].Interface().(type) {
		case error:
			return ret[0].Interface().(error)
		}
	}

	return nil
}

// DeepCopy func
func DeepCopy(src interface{}, dst interface{}) error {
	enc, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return UnmarshalValidJSON(enc, dst)
}

func OpenUrl(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		exec.Command("xdg-open", url)
	}
	return cmd.Run()
}

type PanicData struct {
	PanicValue interface{}
	PanicTrace []byte
}
