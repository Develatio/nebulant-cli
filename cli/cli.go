// Nebulant
// Copyright (C) 2021  Develatio Technologies S.L.

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
	"errors"
	"flag"
	"fmt"
	"runtime"
	"sort"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/interactive"
	"github.com/develatio/nebulant-cli/providers/aws"
	"github.com/develatio/nebulant-cli/providers/azure"
	"github.com/develatio/nebulant-cli/providers/cloudflare"
	"github.com/develatio/nebulant-cli/providers/generic"
	"github.com/develatio/nebulant-cli/providers/hetzner"
	"github.com/develatio/nebulant-cli/subcom"
	"github.com/develatio/nebulant-cli/term"
)

type SCType int

const (
	SecMain SCType = iota
	SecRuntime
	SecHidden
)

type NBLcommand struct {
	upgradeTerm   bool
	welcomeMsg    bool
	initProviders bool
	help          string
	sec           SCType
	run           func(*flag.FlagSet) (int, error)
}

var NBLCommands map[string]*NBLcommand = map[string]*NBLcommand{
	"serve": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: true,
		help:          "  serve\t\t\t" + term.EmojiSet["TridentEmblem"] + " Start server mode\n",
		sec:           SecMain,
		run:           subcom.ServeCmd,
	},
	"run": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: true,
		help:          "  run\t\t\t" + term.EmojiSet["RunningShoe"] + " Run blueprint form file or net\n",
		sec:           SecMain,
		run:           subcom.RunCmd,
	},
	"assets": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  assets\t\t" + term.EmojiSet["Squid"] + " Handle cli assets\n",
		sec:           SecMain,
		run:           subcom.AssetsCmd,
	},
	"interactive": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: true,
		help:          "  interactive\t\t" + term.EmojiSet["Television"] + " Start interactive menu\n",
		sec:           SecMain,
		run: func(cmdline *flag.FlagSet) (int, error) {
			// Interactive mode
			err := interactive.Loop()
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
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  auth\t\t\t" + term.EmojiSet["Key"] + " Server authentication\n",
		sec:           SecMain,
		run:           subcom.AuthCmd,
	},
	"debugterm": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "",
		sec:           SecHidden,
		run:           subcom.DebugtermCmd,
	},
	"update": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  update\t\t" + term.EmojiSet["Squid"] + " Update the cli to the latest version\n",
		sec:           SecMain,
		run:           subcom.UpdateCmd,
	},
	"readvar": {
		upgradeTerm:   false,
		welcomeMsg:    false,
		initProviders: false,
		help:          "  readvar\t\t" + term.EmojiSet["FaceWithMonocle"] + " Read blueprint variable value during runtime\n",
		sec:           SecRuntime,
		run:           subcom.ReadvarCmd,
	},
	"debugger": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  debugger\t\t" + term.EmojiSet["FaceWithMonocle"] + " connect to running debugger\n",
		sec:           SecRuntime,
		run:           subcom.DebuggerCmd,
	},
	"shell": {
		upgradeTerm:   false,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  debugger\t\t" + term.EmojiSet["FaceWithMonocle"] + " connect to running debugger\n",
		sec:           SecRuntime,
		// will be setted by nshell.go.init()
		run: nil,
	},
}

func PrepareCmd(cmd *NBLcommand) error {
	var err error
	if cmd.upgradeTerm {
		// Init Term
		err = term.UpgradeTerm()
		if err != nil {
			// cast.SBus.Close().Wait()
			// fmt.Println("cannot init term :(")
			// fmt.Println(err.Error())
			// os.Exit(1)
			return errors.Join(fmt.Errorf("cannot init term :("), err)
		}
		term.ConfigColors()
	}

	if config.DEBUG {
		cast.LogDebug("Debug mode activated. Testing message levels...", nil)
		cast.LogInfo("Info message", nil)
		cast.LogWarn("Warning message", nil)
		cast.LogErr("Error message", nil)
		cast.LogCritical("Critical message", nil)
	}

	if cmd.welcomeMsg {
		_, err = term.Println(term.Magenta+"Nebulant CLI"+term.Reset, "- A cloud builder by", term.Blue+"develat.io"+term.Reset)
		if err != nil {
			fmt.Println("Nebulant CLI - A cloud builder by develat.io")
		}
		_, err = term.Println(term.Gray+" Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler, term.Reset)
		if err != nil {
			fmt.Println("Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		}
		term.PrintInfo(" Welcome :)\n")
	}

	// Init Providers
	if cmd.initProviders {
		cast.SBus.RegisterProviderInitFunc("aws", aws.New)
		cast.SBus.RegisterProviderInitFunc("azure", azure.New)
		cast.SBus.RegisterProviderInitFunc("generic", generic.New)
		cast.SBus.RegisterProviderInitFunc("hetzner", hetzner.New)
		cast.SBus.RegisterProviderInitFunc("cloudflare", cloudflare.New)
		blueprint.ActionValidators["providerValidator"] = func(action *blueprint.Action) error {
			if _, err := cast.SBus.GetProviderInitFunc(action.Provider); err != nil {
				return err
			}
			return nil
		}
		blueprint.ActionValidators["awsValidator"] = aws.ActionValidator
		blueprint.ActionValidators["azureValidator"] = azure.ActionValidator
		blueprint.ActionValidators["genericsValidator"] = generic.ActionValidator
		blueprint.ActionValidators["hetznerValidator"] = hetzner.ActionValidator
		blueprint.ActionValidators["cloudflareValidator"] = cloudflare.ActionValidator
	}

	return nil
}

func ConfArgs(flag *flag.FlagSet) {
	// var err error
	config.VersionFlag = flag.Bool("v", false, "Show version and exit.")
	config.DebugFlag = flag.Bool("x", false, "Enable debug.")
	config.ParanoicDebugFlag = flag.Bool("xx", false, "Enable paranoic debug.")
	config.Ipv6Flag = flag.Bool("6", false, "Force ipv6")
	config.DisableColorFlag = flag.Bool("c", false, "Disable colors.")
	config.ForceTerm = flag.Bool("ft", false, "Force terminal. Bypass no-term detection.")

	flag.Usage = func() {
		var runtimecmds []string
		var orderedcmdtxt []string
		for cmdtxt, cmd := range NBLCommands {
			if cmd.sec == SecHidden {
				continue
			}
			orderedcmdtxt = append(orderedcmdtxt, cmdtxt)
		}
		sort.Strings(orderedcmdtxt)
		fmt.Fprint(flag.Output(), "\nUsage: nebulant [flags] [command]\n")
		fmt.Fprint(flag.Output(), "\nFlags:\n")
		subcom.PrintDefaults(flag)
		fmt.Fprint(flag.Output(), "\nCommands:\n")
		for _, cmdtxt := range orderedcmdtxt {
			cmd := NBLCommands[cmdtxt]
			if cmd.sec == SecRuntime {
				runtimecmds = append(runtimecmds, cmd.help)
				continue
			}
			fmt.Fprint(flag.Output(), cmd.help)
		}
		fmt.Fprint(flag.Output(), "\n\nRuntime commands:\n")
		for _, hh := range runtimecmds {
			fmt.Fprint(flag.Output(), hh)
		}
		// fmt.Fprint(flag.Output(), "  readvar\t\t"+term.EmojiSet["Key"]+" Read blueprint variable value during runtime\n")
		fmt.Fprint(flag.Output(), "\n\nrun nebulant [command] --help to show help for a command\n")
	}
}

func Start() (errcode int) {

	// Init console logger
	cast.InitConsoleLogger()

	ConfArgs(flag.CommandLine)
	flag.Parse()

	// Version and exit
	if *config.VersionFlag {
		fmt.Println("v" + config.Version)
		fmt.Println("Build date: " + config.VersionDate)
		fmt.Println("Build commit: " + config.VersionCommit)
		fmt.Println("Compiler version: " + config.VersionGo)
		return 0
		// os.Exit(0)
	}

	// Debug
	if *config.DebugFlag {
		config.DEBUG = true
	}
	if *config.ParanoicDebugFlag {
		config.DEBUG = true
		config.PARANOICDEBUG = true
	}

	sc := flag.Arg(0)
	if sc == "" {
		sc = "interactive"
	}
	if cmd, exists := NBLCommands[sc]; exists {
		// prepare cmd
		err := PrepareCmd(cmd)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			// cast.SBus.Close().Wait()
			return 1
		}
		// finally run command
		exitcode, err := cmd.run(flag.CommandLine)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			// cast.SBus.Close().Wait()
		}
		return exitcode
	} else {
		flag.Usage()
		cast.LogErr("Unknown command", nil)
		// cast.SBus.Close().Wait()
		return 1
		// os.Exit(1)
	}
}
