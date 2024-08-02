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
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/develatio/nebulant-cli/ipc"
	"github.com/develatio/nebulant-cli/subsystem"
)

// strict *bool by default is false
//
// false: Hide err msgs and always return
// empty string. The exitcode will be 0
// or 1 on sucessful or error.
//
// true: all errors are printed
var strict *bool

func parseReadVar(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("readvar", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	strict = fs.Bool("strict", false, "Force err msg instead empty string")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "\nUsage: nebulant readvar [variable name] [flags]\n")
		subsystem.PrintDefaults(fs)
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func ReadvarCmd(nblc *subsystem.NBLcommand) (int, error) {
	_, err := parseReadVar(nblc.CommandLine())
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, nil
		}
		return 1, err
	}

	ipcsid := os.Getenv("NEBULANT_IPCSID")
	if ipcsid == "" {
		return 1, fmt.Errorf("cannot found IPC server ID")
	}
	ipccid := os.Getenv("NEBULANT_IPCCID")
	if ipccid == "" {
		return 1, fmt.Errorf("cannot found IPC consumer ID")
	}
	varname := flag.Arg(1)
	val, err := ipc.Read(ipcsid, ipccid, "readvar "+varname)
	if err != nil {
		return 1, err
	}
	if val == "\x10" {
		if *strict {
			return 1, fmt.Errorf("undefined var")
		}
		fmt.Print()
		return 1, nil
	}
	fmt.Print(val)
	return 0, nil
}
