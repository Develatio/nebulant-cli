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
	"fmt"

	"github.com/develatio/nebulant-cli/interactive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
)

func RegisterSubcommands() {
	subsystem.NBLCommands = map[string]*subsystem.NBLcommand{
		"serve": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: true,
			Help:          "  serve\t\t\t" + term.EmojiSet["TridentEmblem"] + " Start server mode\n",
			Sec:           subsystem.SecMain,
			Call:          ServeCmd,
		},
		"run": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: true,
			Help:          "  run\t\t\t" + term.EmojiSet["RunningShoe"] + " Run blueprint form file or net\n",
			Sec:           subsystem.SecMain,
			Call:          RunCmd,
		},
		"assets": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "  assets\t\t" + term.EmojiSet["Squid"] + " Handle cli assets\n",
			Sec:           subsystem.SecMain,
			Call:          AssetsCmd,
		},
		"interactive": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: true,
			Help:          "  interactive\t\t" + term.EmojiSet["Television"] + " Start interactive menu\n",
			Sec:           subsystem.SecMain,
			Call: func(nblc *subsystem.NBLcommand) (int, error) {
				// Interactive mode
				err := interactive.LoopV2(nblc)
				if err != nil {
					if err == term.ErrInterrupt {
						fmt.Println("^C")
						// cast.SBus.Close().Wait()
						return 0, nil
						// os.Exit(0)
					}
					if err == term.ErrEOF {
						fmt.Println("^D")
						// cast.SBus.Close().Wait()
						return 0, nil
						// os.Exit(0)
					}
					return 1, err
					// exitCode := 1
					// cast.LogErr(err.Error(), nil)
					// os.Exit(exitCode)
				}
				// cast.SBus.Close().Wait()
				return 0, nil
				// os.Exit(0)
			},
		},
		"auth": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "  auth\t\t\t" + term.EmojiSet["Key"] + " Server authentication\n",
			Sec:           subsystem.SecMain,
			Call:          AuthCmd,
		},
		"help": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "  help\t\t\t" + term.EmojiSet["Ambulance"] + " shows this help msg\n",
			Sec:           subsystem.SecMain,
			Call: func(nblc *subsystem.NBLcommand) (int, error) {
				nblc.CommandLine().Usage()
				return 0, nil
			},
		},
		"debugterm": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "",
			Sec:           subsystem.SecHidden,
			Call:          DebugtermCmd,
		},
		"update": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "  update\t\t" + term.EmojiSet["Squid"] + " Update the cli to the latest version\n",
			Sec:           subsystem.SecMain,
			Call:          UpdateCmd,
		},
		"readvar": {
			UpgradeTerm:   false,
			WelcomeMsg:    false,
			InitProviders: false,
			Help:          "  readvar\t\t" + term.EmojiSet["FaceWithMonocle"] + " Read blueprint variable value during runtime\n",
			Sec:           subsystem.SecRuntime,
			Call:          ReadvarCmd,
		},
		"debugger": {
			UpgradeTerm:   true,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "  debugger\t\t" + term.EmojiSet["FaceWithMonocle"] + " connect to running debugger\n",
			Sec:           subsystem.SecRuntime,
			Call:          DebuggerCmd,
		},
		// "shell": {
		// 	UpgradeTerm:   false,
		// 	WelcomeMsg:    true,
		// 	InitProviders: false,
		// 	Help:          "  shell\t\t" + term.EmojiSet["FaceWithMonocle"] + " run interactive shell\n",
		// 	Sec:           subsystem.SecRuntime,
		// 	Call:          NSTerm,
		// },
	}
}
