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

package nsterm

// func run(vpty *VPTY2, s string, stdin io.ReadCloser, stdout io.WriteCloser) (int, error) {
// 	argslice, err := util.CommandLineToArgv(s)
// 	if err != nil {
// 		stdout.Write([]byte(err.Error()))
// 	}

// 	cmdline := flag.NewFlagSet(argslice[0], flag.ContinueOnError)
// 	cmdline.SetOutput(stdout)
// 	subsystem.ConfArgs(cmdline)
// 	err = cmdline.Parse(argslice)
// 	if err != nil {
// 		return 1, err
// 	}
// 	sc := cmdline.Arg(0)

// 	if cmd, exists := subsystem.NBLCommands[sc]; exists {
// 		// TODO: implement raw requirement per command
// 		// and set raw only if needed
// 		prev_ldisc := vpty.GetLDisc()
// 		vpty.SetLDisc(NewRawLdisc())
// 		defer vpty.SetLDisc(prev_ldisc)

// 		cmd.UpgradeTerm = false // prevent set raw term
// 		cmd.WelcomeMsg = false  // prevent welcome msg
// 		err := subsystem.PrepareCmd(cmd)
// 		if err != nil {
// 			return 1, err
// 		}
// 		cmd.Stdin = stdin
// 		cmd.Stdout = stdout
// 		// finally run command
// 		return cmd.Run(cmdline)
// 	}
// 	return 127, fmt.Errorf(fmt.Sprintf("?? unknown cmd %v\n", cmdline.Arg(0)))
// }

// func NSShell(vpty *VPTY2, stdin io.ReadCloser, stdout io.WriteCloser) (int, error) {
// 	prmpt := NewPrompt(vpty, stdin, stdout)
// 	for {
// 		s, err := prmpt.ReadLine()
// 		if err != nil {
// 			return 1, err
// 		}
// 		if s == nil {
// 			// no command
// 			continue
// 		}
// 		scmd := *s

// 		// built in :)
// 		if scmd == "exit" {
// 			return 0, nil
// 		}

// 		// built in ;)
// 		if scmd == "help" {
// 			scmd = "--help"
// 		}
// 		ecode, err := run(vpty, scmd, stdin, stdout)
// 		if err != nil {
// 			stdout.Write([]byte(err.Error()))
// 			stdout.Write([]byte(fmt.Sprintf("Exitcode %v", ecode)))
// 		}
// 	}
// }
