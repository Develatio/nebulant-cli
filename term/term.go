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

package term

import (
	"fmt"
	"log"
	"os"

	"github.com/develatio/nebulant-cli/config"
	"golang.org/x/term"
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

var White string = "\033[97m"

var Bold string = "\033[1m"

var CorsorToColZero = "\033[0G"
var CursorUp string = "\033[1F"

var HideCursor string = "\033[?25l"
var ShowCursor string = "\033[?25h"

var EraseLine string = "\033[K"

var mls *MultilineStdout = nil

var statusBarLine *oneLineWriteCloser = nil

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

func isTerminal() bool {
	if *config.ForceTerm {
		return true
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func OpenStatusBar() *oneLineWriteCloser {
	if statusBarLine != nil {
		return statusBarLine
	}
	statusBarLine = mls.AppendLine()
	return statusBarLine
}

func AppendLine() *oneLineWriteCloser {
	return mls.AppendLine()
}

func Selectable(prompt string, options []string) (int, error) {
	return mls.SelectTest(prompt, options)
}

func DeleteLine(line *oneLineWriteCloser) error {
	return mls.DeleteLine(line)
}

func OpenMultilineStdout() {
	if mls == nil {
		mls = &MultilineStdout{
			// WARN: this sould be called AFTER InitTerm
			MainStdout: Stdout,
		}
		log.SetOutput(mls)
	}
}

func CloseStatusBar() error {
	if statusBarLine != nil {
		err := mls.DeleteLine(statusBarLine)
		if err != nil {
			return err
		}
		statusBarLine = nil
	}
	return nil
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

func InitTerm() {
	if *config.DisableColorFlag {
		Stdout = os.Stdout
		Stderr = os.Stderr
		// Reset = ""
		Red = ""
		BGRed = ""
		BGBrightRed = ""
		Green = ""
		BGBrightGreen = ""
		Yellow = ""
		BGYellow = ""
		BGBrightYellow = ""
		Blue = ""
		Black = ""
		BGBlack = ""
		Magenta = ""
		BGBrightMagenta = ""
		Cyan = ""
		Gray = ""
		White = ""
		Bold = ""
		// CursorUp = ""
		// EraseLine = ""
	} else {
		log.SetOutput(Stdout)
	}
	if !config.DEBUG {
		log.SetFlags(0)
	}
	//
	// uses Stdout (term.Stdout in os.go)
	// it can be equal to readline.Stdout
	// as default, or can be os.Stdout if
	// os cannot support colors or if has
	// been disabled manually.
	OpenMultilineStdout()
}
