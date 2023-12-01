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

package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

var cursorUp string = "\033[A"
var cursorDown string = "\033[B"

// var cursorRight string = "\033[C"
// var cursorLeft string = "\033[D"

func NSShell(stdin io.ReadCloser, stdout io.WriteCloser) (int, error) {
	// override term.Stdout (readline.Stdout) by stdout arg
	// this term.Stdout is used by log package and term.Print*
	// the cast.ConsoleLogger also uses term.Print* and log
	// package.

	prev_term_stdout := term.Stdout
	defer func() { term.Stdout = prev_term_stdout }()
	term.Stdout = stdout

	prev_term_stderr := term.Stderr
	defer func() { term.Stderr = prev_term_stderr }()
	term.Stderr = stdout

	log.SetOutput(stdout)

	defer func() { config.ForceNoTerm = false }()
	config.ForceNoTerm = true

	term.UpgradeTerm()

	reader := bufio.NewReader(stdin)
	var lin, prompt string
	var err error

	var shellhistory []string
	var shellhistoryidx int = -1
	PS1 := []byte("NBShell> ")

L:
	for {
		stdout.Write(PS1)
	L2:
		for {
			lin = ""
			for {
				lin, err = reader.ReadString('\n')
				if err != nil {
					stdout.Write([]byte("out of for"))
					return 1, err
				}
				// stdout.Write([]byte("readed line: " + lin))
				break
			}

			// 27 = esc
			if lin[0] == 27 || lin == "\n" {
				esc := strings.TrimSuffix(lin, "\n")
				switch esc {
				case cursorUp: // ESC arrow up
					shellhistoryidx--
					if shellhistoryidx < 0 {
						shellhistoryidx = 0
					}
					prompt = shellhistory[shellhistoryidx]
					stdout.Write([]byte("\r" + term.EraseEntireLine))
					stdout.Write(PS1)
					stdout.Write([]byte(prompt))
					continue
				case cursorDown: // ESC arrow up
					// stdout.Write([]byte("history up\n"))
					shellhistoryidx++
					if shellhistoryidx > len(shellhistory)-1 {
						shellhistoryidx = len(shellhistory) - 1
						stdout.Write([]byte("\r" + term.EraseEntireLine))
						break L2
					}
					prompt = shellhistory[shellhistoryidx]
					stdout.Write([]byte("\r" + term.EraseEntireLine))
					stdout.Write(PS1)
					stdout.Write([]byte(prompt))
					continue
				case "":
					stdout.Write([]byte("running history command\n"))
				default:
					stdout.Write([]byte(fmt.Sprintf("unknown special instruction %v\n", []byte(esc))))
				}
			} else {
				prompt = strings.TrimSuffix(lin, "\n")
			}

			shellhistory = append(shellhistory, prompt)
			shellhistoryidx = len(shellhistory)

			stdout.Write([]byte("> " + prompt + ";\n"))
			// internal commands
			switch prompt {
			case "exit":
				stdout.Write([]byte("should exit from shell\n"))
				break L
			case "\n":
				stdout.Write([]byte("\n"))
			default:
				// ln := lin[:len(lin)-1] // rm \n
				argslice, err := util.CommandLineToArgv(prompt)
				if err != nil {
					stdout.Write([]byte(err.Error()))
				}

				cmdline := flag.NewFlagSet(argslice[0], flag.ContinueOnError)
				cmdline.SetOutput(stdout)
				ConfArgs(cmdline)
				cmdline.Parse(argslice)
				sc := cmdline.Arg(0)

				if cmd, exists := NBLCommands[sc]; exists {
					cmd.upgradeTerm = false // prevent set raw term
					cmd.welcomeMsg = false  // prevent welcome msg
					err := PrepareCmd(cmd)
					if err != nil {
						stdout.Write([]byte(err.Error()))
						break L2
					}
					// finally run command
					exitcode, err := cmd.run(cmdline)
					if err != nil {
						stdout.Write([]byte(err.Error()))
						// cast.SBus.Close().Wait()
					}
					stdout.Write([]byte(fmt.Sprintf("out with exitcode %d\n", exitcode)))
					break L2
				}
				stdout.Write([]byte(fmt.Sprintf("unknown cmd %v\n", []byte(prompt))))
			}
			break L2
		}
	}
	return 0, nil
}

func init() {
	NBLCommands["shell"].run = NSTerm
}
