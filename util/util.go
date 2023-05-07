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
	"os/exec"
	"runtime"
	"strings"

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

// func PrintUsage(err error) {
// 	if err != nil {
// 		fmt.Println("nebulant:", err.Error())
// 	}
// 	fmt.Println("\nUsage: nebulant [options] [command]")
// 	fmt.Println("Nebulant options:")
// 	flag.PrintDefaults()
// 	os.Exit(1)
// }

func PrintDefaults(f *flag.FlagSet) {
	f.VisitAll(func(ff *flag.Flag) {
		var b strings.Builder
		fmt.Fprintf(&b, "  -%s ", ff.Name)
		name, usage := flag.UnquoteUsage(ff)
		if len(name) > 0 {
			b.WriteString(name)
		}
		l := 25 - (len(b.String()) + len(name))
		for i := 0; i < l; i++ {
			b.WriteString(" ")
		}
		b.WriteString(usage)
		if ff.DefValue != "" && ff.DefValue != "false" {
			fmt.Fprintf(&b, " (default %v)", ff.DefValue)
		}
		fmt.Fprint(f.Output(), b.String(), "\n")
	})
}
