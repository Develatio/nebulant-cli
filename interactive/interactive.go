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

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
	"github.com/manifoldco/promptui"
)

type menuItem struct {
	Name        string
	Description string
	Cmd         string
}

func Loop(nblc *subsystem.NBLcommand) error {
	menuItems := []*menuItem{
		{Name: "Serve ", Description: "Start server mode at " + net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT), Cmd: "serve"},
		{Name: "Build ", Description: "Open builder app and start server mode", Cmd: "build"},
		{Name: "Browse", Description: "Brwose and run the blueprints stored in your account", Cmd: "browse"},
		{Name: "Path", Description: "Manually indicates the path to a blueprint", Cmd: "path"},
		// {Name: "Config", Description: "Configuration stuff", Cmd: "noim"},
		// {Name: "Args  ", Description: "Print available commandline args", Cmd: "args"},
		{Name: "Exit  ", Description: "Exit Nebulant CLI", Cmd: "exit"},
	}
	templates := &promptui.SelectTemplates{
		Label:    " ",
		Active:   term.EmojiSet["BackhandIndexPointingRight"] + " {{ .Name | magenta }} \t\t {{ .Description }}",
		Inactive: "   {{ .Name | cyan }} \t\t {{ .Description | faint }}",
		Selected: "> {{ .Name | magenta }}",
	}
L:
	for {
		prompt := promptui.Select{
			Items:     menuItems,
			Size:      len(menuItems),
			Templates: templates,
			Stdout:    nblc.Stdout,
			Stdin:     nblc.Stdin,
		}
		i, _, err := prompt.Run()
		if err != nil {
			return err
		}
		input := menuItems[i].Cmd

		switch input {
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
			err := Path(nblc)
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
			err := Browser(nblc)
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
