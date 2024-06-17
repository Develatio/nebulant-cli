//go:build !js

// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package interactive

import (
	"errors"
	"net"
	"net/http"

	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

func LoopV2(nblc *subsystem.NBLcommand) error {
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
						huh.NewOption("Browse\tBrowse and run the blueprints stored in your account", "browse"),
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
				if !errors.Is(err, http.ErrServerClosed) {
					term.PrintErr(err.Error() + "\n")
				}
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
				if !errors.Is(err, http.ErrServerClosed) {
					term.PrintErr(err.Error() + "\n")
				}
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
