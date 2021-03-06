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
)

var Reset string = "\033[0m"
var Red string = "\033[31m"

var Green string = "\033[32m"
var Yellow string = "\033[33m"

var Blue string = "\033[34m"
var Purple string = "\033[35m"

var Cyan string = "\033[36m"
var Gray string = "\033[97m"

// var White string = "\033[97m"

var Bold string = "\033[1m"

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

func InitTerm(colors bool) {
	if !colors {
		Stdout = os.Stdout
		Stderr = os.Stderr
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		// Cyan = ""
		Gray = ""
		// White = ""
		Bold = ""
	} else {
		log.SetOutput(Stdout)
	}

}
