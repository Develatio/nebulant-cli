// MIT License
//
// Copyright (C) 2022  Develatio Technologies S.L.

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

package term

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"math/rand"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/config"
	x_term "golang.org/x/term"
)

var Reset string = "\033[0m"
var Red string = "\033[31m"
var BGRed string = "\033[41m"
var BGBrightRed string = "\033[101m"

var Green string = "\033[32m"
var BGBrightGreen string = "\033[102m"
var Yellow string = "\033[33m"
var BGYellow string = "\033[43m"
var BGBrightYellow string = "\033[103m"

var Blue string = "\033[34m"
var BGBlue string = "\033[44m"
var Black string = "\033[30m"
var BGBlack string = "\033[40m"
var Magenta string = "\033[35m"
var BGMagenta string = "\033[45m"
var BGBrightMagenta string = "\033[105m"

var Cyan string = "\033[36m"
var BGCyan string = "\033[46m"

var Gray string = "\033[97m"

var White string = "\033[37m"

var Bold string = "\033[1m"

var CursorToColZero = "\033[0G"
var CursorUp string = "\033[1A"
var CursorDown string = "\033[1B"
var CursorLeft string = "\033[1D"

var SaveCursor string = "\033[s"
var RestoreCursor string = "\033[u"

var HideCursor string = "\033[?25l"
var ShowCursor string = "\033[?25h"

var EraseLine string = "\033[K"

var EraseLineFromCursor string = "\033[0K"
var EraseEntireLine string = "\033[2K"
var EraseDisplay string = "\033[2J"

var IdentifyDevice string = "\033Z"

// var mls *MultilineStdout = nil

var colors []string
var used_colors []string

func GetNewColor() string {
	if len(colors) <= 0 {
		colors = used_colors
		used_colors = []string{}
	}

	i := 0
	color := colors[i]
	colors = append(colors[:i], colors[i+1:]...)
	used_colors = append(used_colors, color)

	return color
}

type OSPTY interface {
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Wait(ctx context.Context) (int64, error)
}

type OSPTYConf struct {
	Shell string
	Env   []string
}

// https://github.com/manifoldco/promptui/issues/49
type noBellStdout struct{}

func (n *noBellStdout) Write(p []byte) (int, error) {
	if len(p) == 1 && p[0] == CharBell {
		return 0, nil
	}
	return Stdout.Write(p)
}

func (n *noBellStdout) Close() error {
	return Stdout.Close()
}

var NoBellStdout = &noBellStdout{}

// IsTerminal returns false if no real term on stdout has
// dettected or if NoTerm flag (-n) has been setted
func IsTerminal() bool {
	if config.NoTermFlag != nil && *config.NoTermFlag {
		return false
	}
	return x_term.IsTerminal(int(os.Stdout.Fd()))
}

// PrintInfo func
func PrintInfo(s string) {
	fmt.Fprintf(Stdout, Gray+s+Reset)
}

// PrintWarn func
func PrintWarn(s string) {
	fmt.Fprintf(Stdout, Yellow+s+Reset)
}

// PrintErr func
func PrintErr(s string) {
	fmt.Fprintf(Stdout, Red+s+Reset)
}

func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(Stdout, a...)
}

func Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(Stdout, a...)
}

// UpgradeTerm func sets advanced ANSI supoprt, colors and
// multiline StdOut
func UpgradeTerm() error {
	// return nil
	var err error
	if config.LOGLEVEL > base.DebugLevel {
		log.SetFlags(0)
	}

	err = EnableColorSupport()
	if err != nil {
		return errors.Join(fmt.Errorf("cannot enable colors"), err)
	}
	log.SetOutput(Stdout)
	//
	// uses Stdout (term.Stdout in os.go)
	// it can be equal to readline.Stdout
	// as default, or can be os.Stdout if
	// os cannot support colors or if has
	// been disabled manually.
	// openMultilineStdout()
	return nil
}

func init() {
	for i := 0; i < 256; i++ {
		colors = append(colors, fmt.Sprintf("\033[38;5;%vm", i))
	}
	for i := len(colors) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		colors[i], colors[j] = colors[j], colors[i]
	}
}
