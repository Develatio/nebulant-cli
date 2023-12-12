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

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
)

func parseAuthFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant auth [command] [options]\n")
		fmt.Fprintf(fs.Output(), "\nCommands:\n")
		fmt.Fprintf(fs.Output(), "  newtoken\t\tNegotiate and save new backend token\n")
		// fmt.Fprintf(fs.Output(), "  login\t\tLogin\n")
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func AuthCmd(cmdline *flag.FlagSet) (int, error) {
	fs, err := parseAuthFs(cmdline)
	if err != nil {
		return 1, err
	}

	// subsubcmd := fs.Arg(0)
	subsubcmd := cmdline.Arg(1)
	switch subsubcmd {
	case "newtoken":
		err := config.RequestToken()
		if err != nil {
			return 1, err
		}
		cast.LogInfo("token sucefully saved", nil)
	// case "login":
	default:
		fs.Usage()
		return 1, fmt.Errorf("please provide some subcommand to auth")
	}
	return 0, nil
}
