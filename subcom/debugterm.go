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
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
)

func newBar() {
	line := term.AppendLine()
	defer line.Close()
	bar, err := line.GetProgressBar(int64(100), "Testing bar", false)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		err := bar.Add(1)
		if err != nil {
			panic(err)
		}
		time.Sleep(60000 * time.Microsecond)
	}
}

func parseTestsFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("debugterm", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant debugterm [command]\n")
		fmt.Fprintf(fs.Output(), "\nCommands:\n")
		fmt.Fprintf(fs.Output(), "  scanln\t\tTest scanln while log\n")
		fmt.Fprintf(fs.Output(), "  ansi\t\tTest ansi codes\n")
		fmt.Fprintf(fs.Output(), "  raw\t\tSet raw term mode\n")
		fmt.Fprintf(fs.Output(), "  unraw\t\tdisable raw term mode\n")
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func DebugtermCmd(nblc *subsystem.NBLcommand) (int, error) {
	cmdline := nblc.CommandLine()
	fs, err := parseTestsFs(cmdline)
	if err != nil {
		return 1, err
	}

	// subsubcmd := fs.Arg(0)
	subsubcmd := cmdline.Arg(1)
	switch subsubcmd {
	case "scanln":
		return testScanln()
	case "ansi":
		return testAnsiCodes()
	case "raw":
		cast.LogInfo("Not implemented yet", nil)
		return 0, nil
	case "unraw":
		cast.LogInfo("Not implemented yet", nil)
		return 0, nil
	default:
		fs.Usage()
		return 1, fmt.Errorf("please provide some subcommand to auth")
	}

}

func wrtAnsiCodes(w io.Writer) {
	fmt.Fprintf(w, "%s\n", term.CursorToColZero)
	fmt.Fprintf(w, "%s\n", term.EraseDisplay)
	fmt.Fprintf(w, "%sRED RED RED RED RED RED RED RED%s\n", term.Red, term.Reset)
	fmt.Fprintf(w, "%sBGRED BGRED BGRED BGRED BGRED BGRED BGRED BGRED%s\n", term.BGRed, term.Reset)
	fmt.Fprintf(w, "%sBGRED BGRED BGRED BGRED BGRED BGRED BGRED BGRED%s\n", term.BGRed, term.Reset)
}

func testAnsiCodes() (int, error) {
	fmt.Fprint(term.GenuineOsStdout, "Genuine STDOUT:\n")
	wrtAnsiCodes(term.GenuineOsStdout)

	fmt.Fprint(term.Stdout, "\nANSI STDOUT:\n")
	wrtAnsiCodes(term.Stdout)

	fmt.Fprint(term.Stdout, "\nEnabling ansi support:\n")
	err := term.EnableColorSupport()
	if err != nil {
		fmt.Fprint(term.Stdout, err.Error())
	}

	fmt.Fprint(term.GenuineOsStdout, "Genuine STDOUT:\n")
	wrtAnsiCodes(term.GenuineOsStdout)

	fmt.Fprint(term.Stdout, "\nANSI STDOUT:\n")
	wrtAnsiCodes(term.Stdout)

	return 0, nil
}

func testScanln() (int, error) {
	go func() {
		counter1 := 0
		for {
			counter1++
			cast.LogInfo("blabla blablabla blabla bla bla bla blabla bla bla bla bla ... "+strings.Repeat(" ", counter1)+strconv.Itoa(counter1), nil)
			time.Sleep(600000 * time.Microsecond)
			if counter1 == 9 {
				counter1 = 0
			}
		}
	}()

	counter2 := 0
	lin := term.AppendLine()
	defer lin.Close()
L:
	for {
		counter2++

		var vv string
		_, err := lin.Scanln("Nebulant "+strconv.Itoa(counter2)+"> ", nil, &vv)
		if err != nil {
			return 1, err
		}
		switch vv {
		case "exit":
			break L
		case "addbar":
			go newBar()
		}
	}
	return 0, nil
}
