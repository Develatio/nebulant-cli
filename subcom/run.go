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

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/executive"
)

func parseRunFs() (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant run [path or nebulant:// protocol] [--varname=varvalue --varname=varvalue]\n")
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(flag.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func RunCmd() (int, error) {
	fs, err := parseRunFs()
	if err != nil {
		return 1, err
	}
	bluePrintFilePath := fs.Arg(0)
	if bluePrintFilePath == "" {
		fs.Usage()
		return 1, fmt.Errorf("please provide file path or nebulant:// protocol")
	}

	cast.LogInfo("Processing blueprint...", nil)
	irbConf := &blueprint.IRBGenConfig{}
	args := fs.Args()
	if len(args) > 1 {
		irbConf.Args = args[1:]
	}
	irb, err := blueprint.NewIRBFromAny(bluePrintFilePath, irbConf)
	if err != nil {
		return 1, err
	}
	// Director in one run mode
	err = executive.InitDirector(false, false)
	if err != nil {
		return 1, err
	}
	executive.MDirector.HandleIRB <- irb
	executive.MDirector.Wait()
	return 0, nil
}
