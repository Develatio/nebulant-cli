// MIT License
//
// Copyright (C) 2021  Develatio Technologies S.L.

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

package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/subcom"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/tui"
)

func Start() (errcode int) {
	if err := subsystem.ConfArgs(flag.CommandLine, os.Args[1:]); err != nil {
		cast.LogErr(err.Error(), nil)
		return 1
	}

	// Init console logger
	ff := func(fLink *cast.BusConsumerLink) error {
		_, err := tui.StartUI(fLink)
		return err
	}
	cast.InitConsoleLogger(ff)
	subcom.RegisterSubcommands()

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
