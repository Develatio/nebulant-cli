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

func parseReadVar() (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("readvar", flag.ExitOnError)
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
	ipccid := os.Getenv("NEBULANT_IPCCID")
	varname := flag.Arg(1)
	val, err := ipc.Read(ipcsid, ipccid, "readvar "+varname)
	if err != nil {
		return 1, err
	}
	if val == "\x20" {
		fmt.Print()
		return 0, nil
	}
	fmt.Print(val)
	return 0, nil
}
