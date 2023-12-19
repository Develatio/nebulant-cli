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
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/update"
)

var forceupdate *bool

func parseUpdate(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	forceupdate = fs.Bool("f", false, "force update")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant update [options]\n")
		fmt.Fprintf(cmdline.Output(), "\nOptions:\n")
		subsystem.PrintDefaults(fs)
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func UpdateCmd(nblc *subsystem.NBLcommand) (int, error) {
	_, err := parseUpdate(nblc.CommandLine())
	if err != nil {
		return 1, err
	}

	out, err := update.UpdateCLI("latest", *forceupdate)
	if err != nil {
		if _, ok := err.(*update.AlreadyUpToDateError); ok {
			cast.LogInfo("Already up to date", nil)
			return 0, nil
		}
		return 1, err
	}
	if out != nil {
		cast.LogInfo(fmt.Sprintf("Updated to version: %s (%s) ", out.NewVersion.Version, out.NewVersion.Date), nil)
	}
	return 0, nil
}
