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
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
)

func newBar() {
	bar := cast.NewProgress(&cast.ProgressConf{
		Size:    int64(100),
		Info:    "Testing bar",
		Autoend: true,
	})

	for i := 0; i < 100; i++ {
		bar.Add(1)
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
		fmt.Fprintf(fs.Output(), "  bar\t\tTest progress bar\n")
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
		if errors.Is(err, flag.ErrHelp) {
			return 0, nil
		}
		return 1, err
	}

	// subsubcmd := fs.Arg(0)
	subsubcmd := cmdline.Arg(1)
	switch subsubcmd {
	case "scanln":
		return testScanln()
	case "ansi":
		return testAnsiCodes()
	case "bar":
		return testBar()
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

func testBar() (int, error) {
	go newBar()
	time.Sleep(1 * time.Second)
	go newBar()
	time.Sleep(1 * time.Second)
	go newBar()
	time.Sleep(1 * time.Second)
	go newBar()
	time.Sleep(1 * time.Second)
	go newBar()
	time.Sleep(1 * time.Second)
	newBar()
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
	// lin := term.AppendLine()
	// defer lin.Close()
L:
	for {
		counter2++
		vc, err := cast.PromptInput("> ", true, "")
		if err != nil {
			return 1, err
		}
		vv := <-vc
		switch vv {
		case "exit":
			break L
		case "addbar":
			go newBar()
		}
	}
	return 0, nil
}
