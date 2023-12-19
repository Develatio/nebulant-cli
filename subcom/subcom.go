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
				err := interactive.Loop(nblc)
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
		"shell": {
			UpgradeTerm:   false,
			WelcomeMsg:    true,
			InitProviders: false,
			Help:          "  debugger\t\t" + term.EmojiSet["FaceWithMonocle"] + " connect to running debugger\n",
			Sec:           subsystem.SecRuntime,
			Call:          NSTerm,
		},
	}
}
