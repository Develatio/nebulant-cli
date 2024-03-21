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

package cli

import (
	"flag"
	"fmt"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/subcom"
	"github.com/develatio/nebulant-cli/subsystem"
)

func Start() (errcode int) {

	// Init console logger
	cast.InitConsoleLogger()
	subcom.RegisterSubcommands()

	subsystem.ConfArgs(flag.CommandLine)
	flag.Parse()

	// Version and exit
	if *config.VersionFlag {
		fmt.Println("v" + config.Version)
		fmt.Println("Build date: " + config.VersionDate)
		fmt.Println("Build commit: " + config.VersionCommit)
		fmt.Println("Compiler version: " + config.VersionGo)
		return 0
		// os.Exit(0)
	}

	// Debug
	if *config.DebugFlag {
		config.DEBUG = true
	}
	if *config.ParanoicDebugFlag {
		config.DEBUG = true
		config.PARANOICDEBUG = true
	}

	sc := flag.Arg(0)
	if sc == "" {
		sc = "interactive"
	}

	exitcode, err := subsystem.Run(sc)
	if err != nil {
		cast.LogErr(err.Error(), nil)
	}
	return exitcode
}
