//go:build !js

// Nebulant
// Copyright (C) 2022  Develatio Technologies S.L.

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

package interactive

import (
	"net"

	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

func LoopV2(nblc *subsystem.NBLcommand) error {

	// menuItems := []*menuItem{
	// 	{Name: "Serve ", Description: "Start server mode at " + net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT), Cmd: "serve"},
	// 	{Name: "Build ", Description: "Open builder app and start server mode", Cmd: "build"},
	// 	{Name: "Browse", Description: "Brwose and run the blueprints stored in your account", Cmd: "browse"},
	// 	{Name: "Path", Description: "Manually indicates the path to a blueprint", Cmd: "path"},
	// 	// {Name: "Config", Description: "Configuration stuff", Cmd: "noim"},
	// 	// {Name: "Args  ", Description: "Print available commandline args", Cmd: "args"},
	// 	{Name: "Exit  ", Description: "Exit Nebulant CLI", Cmd: "exit"},
	// }
	// templates := &promptui.SelectTemplates{
	// 	Label:    " ",
	// 	Active:   term.EmojiSet["BackhandIndexPointingRight"] + " {{ .Name | magenta }} \t\t {{ .Description }}",
	// 	Inactive: "   {{ .Name | cyan }} \t\t {{ .Description | faint }}",
	// 	Selected: "> {{ .Name | magenta }}",
	// }
L:
	for {
		var Cmd string

		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Main menu. What do you want to do?").
					Options(
						huh.NewOption("Serve\tStart server mode at "+net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT), "serve"),
						huh.NewOption("Build\tOpen builder app and start server mode", "build"),
						huh.NewOption("Path\tManually indicates the path to a blueprint", "path"),
						huh.NewOption("Browse\tBrwose and run the blueprints stored in your account", "browse"),
						huh.NewOption("Exit\tExit Nebulant CLI", "exit"),
						// huh.NewOption("Brazil", "BR"),
						// huh.NewOption("Canada", "CA"),
					).
					Value(&Cmd),
			)).Run()
		if err != nil {
			return err
		}
		switch Cmd {
		// case "args":
		// 	util.PrintUsage(nil)
		case "exit":
			break L
		case "serve":
			term.PrintInfo("Starting server mode...\n")
			term.PrintInfo("You can also start server mode using -s argument\n")
			err := executive.InitDirector(true, true) // Server mode
			if err != nil {
				term.PrintErr(err.Error() + "\n")
				continue
			}
			errc := executive.InitServerMode()
			err = <-errc
			if err != nil {
				term.PrintErr(err.Error() + "\n")
				continue
			}
			executive.MDirector.Wait()
		case "path":
			err := PathV2(nblc)
			if err != nil {
				term.PrintErr(err.Error() + "\n")
				continue
			}
		case "build":
			err := util.OpenUrl(config.FrontUrl)
			if err != nil {
				term.PrintWarn("Warning: " + err.Error() + "\n")
				term.PrintWarn("You can still run \"serve\" command and open " + config.FrontUrl + " manually.\n")
				continue
			}
			err = executive.InitDirector(true, true) // Server mode
			if err != nil {
				term.PrintErr(err.Error() + "\n")
				continue
			}
			errc := executive.InitServerMode()
			err = <-errc
			if err != nil {
				term.PrintErr(err.Error() + "\n")
				continue
			}
			executive.MDirector.Wait()
			break L
		case "browse":
			err := Browserv2(nblc)
			if err != nil {
				term.PrintErr(err.Error() + "\n")
				continue
			}
		case "noim":
			term.PrintErr("Not implemented.\n")
		case "":
			continue
		default:
			term.PrintErr("Error: Unknown command.\n")
		}
	}
	return nil
}
