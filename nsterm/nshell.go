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

package nsterm

import (
	"flag"
	"fmt"
	"io"

	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/util"
)

func run(vpty *VPTY2, s string, stdin io.ReadCloser, stdout io.WriteCloser) (int, error) {
	argslice, err := util.CommandLineToArgv(s)
	if err != nil {
		stdout.Write([]byte(err.Error()))
	}

	cmdline := flag.NewFlagSet(argslice[0], flag.ContinueOnError)
	cmdline.SetOutput(stdout)
	subsystem.ConfArgs(cmdline)
	err = cmdline.Parse(argslice)
	if err != nil {
		return 1, err
	}
	sc := cmdline.Arg(0)

	if cmd, exists := subsystem.NBLCommands[sc]; exists {
		// TODO: implement raw requirement per command
		// and set raw only if needed
		prev_ldisc := vpty.GetLDisc()
		vpty.SetLDisc(NewRawLdisc())
		defer vpty.SetLDisc(prev_ldisc)

		cmd.UpgradeTerm = false // prevent set raw term
		cmd.WelcomeMsg = false  // prevent welcome msg
		err := subsystem.PrepareCmd(cmd)
		if err != nil {
			return 1, err
		}
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		// finally run command
		return cmd.Run(cmdline)
	}
	return 127, fmt.Errorf(fmt.Sprintf("?? unknown cmd %v\n", cmdline.Arg(0)))
}

func NSShell(vpty *VPTY2, stdin io.ReadCloser, stdout io.WriteCloser) (int, error) {
	prmpt := NewPrompt(vpty, stdin, stdout)
	for {
		s, err := prmpt.ReadLine()
		if err != nil {
			return 1, err
		}
		if s == nil {
			// no command
			continue
		}
		scmd := *s

		// built in :)
		if scmd == "exit" {
			return 0, nil
		}

		// built in ;)
		if scmd == "help" {
			scmd = "--help"
		}
		ecode, err := run(vpty, scmd, stdin, stdout)
		if err != nil {
			stdout.Write([]byte(err.Error()))
			stdout.Write([]byte(fmt.Sprintf("Exitcode %v", ecode)))
		}
	}
}
