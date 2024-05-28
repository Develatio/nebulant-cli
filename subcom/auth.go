// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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

package subcom

import (
	"flag"
	"fmt"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/subsystem"
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

func AuthCmd(nblc *subsystem.NBLcommand) (int, error) {
	cmdline := nblc.CommandLine()
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
