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
