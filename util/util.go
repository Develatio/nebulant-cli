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
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	validate := validator.New()
	vErr := validate.Struct(v)
	if vErr != nil {
		return vErr
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
		cmd = exec.Command("start", url)
	default:
		exec.Command("xdg-open", url)
	}
	return cmd.Run()
}

type PanicData struct {
	PanicValue interface{}
	PanicTrace []byte
}

func PrintUsage(err error) {
	if err != nil {
		fmt.Println("nebulant:", err.Error())
	}
	fmt.Println("Usage: nebulant [-options] [file.json | nebulant://UUID]")
	fmt.Println("Nebulant options:")
	flag.PrintDefaults()
	os.Exit(1)
}
