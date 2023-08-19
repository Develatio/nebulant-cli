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

package subcom

import (
	"flag"
	"fmt"
	"os"

	"github.com/develatio/nebulant-cli/ipc"
)

// strict *bool by default is false
//
// false: Hide err msgs and always return
// empty string. The exitcode will be 0
// or 1 on sucessful or error.
//
// true: all errors are printed
var strict *bool

func parseReadVar() (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("readvar", flag.ExitOnError)
	strict = fs.Bool("strict", false, "Force err msg instead empty string")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "\nUsage: nebulant readvar [variable name] [flags]\n")
		PrintDefaults(fs)
	}
	err := fs.Parse(flag.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func ReadvarCmd() (int, error) {
	_, err := parseReadVar()
	if err != nil {
		return 1, err
	}

	ipcsid := os.Getenv("NEBULANT_IPCSID")
	if ipcsid == "" {
		return 1, fmt.Errorf("cannot found IPC server ID")
	}
	ipccid := os.Getenv("NEBULANT_IPCCID")
	if ipccid == "" {
		return 1, fmt.Errorf("cannot found IPC consumer ID")
	}
	varname := flag.Arg(1)
	val, err := ipc.Read(ipcsid, ipccid, "readvar "+varname)
	if err != nil {
		if *strict {
			return 1, err
		}
		return 1, nil
	}
	if val == "\x20" {
		if *strict {
			return 1, fmt.Errorf("undefined var")
		}
		fmt.Print()
		return 1, nil
	}
	fmt.Print(val)
	return 0, nil
}
