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

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"

	// hey hacker:
	// uncomment for profiling
	// _ "net/http/pprof"
	// grmon "github.com/bcicen/grmon/agent"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/interactive"
	"github.com/develatio/nebulant-cli/providers/aws"
	"github.com/develatio/nebulant-cli/providers/azure"
	"github.com/develatio/nebulant-cli/providers/cloudflare"
	"github.com/develatio/nebulant-cli/providers/generic"
	"github.com/develatio/nebulant-cli/providers/hetzner"
	"github.com/develatio/nebulant-cli/subcom"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

type sctype int

const (
	secmain sctype = iota
	secruntime
	sechidden
)

type nbcommand struct {
	upgradeTerm   bool
	welcomeMsg    bool
	initProviders bool
	help          string
	sec           sctype
	run           func()
}

var commands map[string]*nbcommand = map[string]*nbcommand{
	"serve": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: true,
		help:          "  serve\t\t\t" + term.EmojiSet["TridentEmblem"] + " Start server mode\n",
		sec:           secmain,
		run: func() {
			exitCode, err := subcom.ServeCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
				os.Exit(exitCode)
			}
		},
	},
	"run": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: true,
		help:          "  run\t\t\t" + term.EmojiSet["RunningShoe"] + " Run blueprint form file or net\n",
		sec:           secmain,
		run: func() {
			exitCode, err := subcom.RunCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
				os.Exit(exitCode)
			}
		},
	},
	"assets": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  assets\t\t" + term.EmojiSet["Squid"] + " Handle cli assets\n",
		sec:           secmain,
		run: func() {
			exitCode, err := subcom.AssetsCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
				os.Exit(exitCode)
			}
		},
	},
	"interactive": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: true,
		help:          "  interactive\t\t" + term.EmojiSet["Television"] + " Start interactive menu\n",
		sec:           secmain,
		run: func() {
			// Interactive mode
			err := interactive.Loop()
			if err != nil {
				if err == term.ErrInterrupt {
					fmt.Println("^C")
					cast.SBus.Close().Wait()
					os.Exit(0)
				}
				if err == term.ErrEOF {
					fmt.Println("^D")
					cast.SBus.Close().Wait()
					os.Exit(0)
				}
				exitCode := 1
				cast.LogErr(err.Error(), nil)
				os.Exit(exitCode)
			}
			cast.SBus.Close().Wait()
			os.Exit(0)
		},
	},
	"auth": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  auth\t\t\t" + term.EmojiSet["Key"] + " Server authentication\n",
		sec:           secmain,
		run: func() {
			exitCode, err := subcom.AuthCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
				os.Exit(exitCode)
			}
		},
	},
	"debugterm": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "",
		sec:           sechidden,
		run: func() {
			exitCode, err := subcom.DebugtermCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
				os.Exit(exitCode)
			}
		},
	},
	"update": {
		upgradeTerm:   true,
		welcomeMsg:    true,
		initProviders: false,
		help:          "  update\t\t" + term.EmojiSet["Squid"] + " Update the cli to the latest version\n",
		sec:           secmain,
		run: func() {
			exitCode, err := subcom.UpdateCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
				os.Exit(exitCode)
			}
		},
	},
	"readvar": {
		upgradeTerm:   false,
		welcomeMsg:    false,
		initProviders: false,
		help:          "  readvar\t\t" + term.EmojiSet["FaceWithMonocle"] + " Read blueprint variable value during runtime\n",
		sec:           secruntime,
		run: func() {
			exitCode, err := subcom.ReadvarCmd()
			if err != nil {
				cast.LogErr(err.Error(), nil)
				cast.SBus.Close().Wait()
			}
			os.Exit(exitCode)
		},
	},
}

func main() {
	var err error
	exitCode := 0
	// System init
	cast.InitSystemBus()
	defer func() {
		if r := recover(); r != nil {
			exitCode = 1
			switch r := r.(type) {
			case *util.PanicData:
				v := fmt.Sprintf("%v", r.PanicValue)
				cast.LogErr("If you think this is a bug,", nil)
				cast.LogErr("please consider posting stack trace as a GitHub", nil)
				cast.LogErr("issue (https://github.com/develatio/nebulant-cli/issues)", nil)
				cast.LogErr("Stack Trace:", nil)
				cast.LogErr(v, nil)
				cast.LogErr(string(r.PanicTrace), nil)
			default:
				cast.LogErr("Panic", nil)
				cast.LogErr("If you think this is a bug,", nil)
				cast.LogErr("please consider posting stack trace as a GitHub", nil)
				cast.LogErr("issue (https://github.com/develatio/nebulant-cli/issues)", nil)
				cast.LogErr("Stack Trace:", nil)
				v := fmt.Sprintf("%v", r)
				cast.LogErr(v, nil)
				cast.LogErr(string(debug.Stack()), nil)
			}
		}
		// Heredate exit cocde from director
		if exitCode == 0 && executive.MDirector != nil {
			exitCode = executive.MDirector.ExitCode
		}
		if exitCode > 0 {
			cast.LogErr("Done with status "+strconv.Itoa(exitCode), nil)
		} else {
			cast.LogInfo("Done with status "+strconv.Itoa(exitCode), nil)
		}
		cast.SBus.Close().Wait()
		os.Exit(exitCode)
	}()

	config.VersionFlag = flag.Bool("v", false, "Show version and exit.")
	config.DebugFlag = flag.Bool("x", false, "Enable debug.")
	config.Ipv6Flag = flag.Bool("6", false, "Force ipv6")
	config.DisableColorFlag = flag.Bool("c", false, "Disable colors.")
	config.ForceTerm = flag.Bool("ft", false, "Force terminal. Bypass no-term detection.")

	flag.Usage = func() {
		var runtimecmds []string
		var orderedcmdtxt []string
		for cmdtxt, cmd := range commands {
			if cmd.sec == sechidden {
				continue
			}
			orderedcmdtxt = append(orderedcmdtxt, cmdtxt)
		}
		sort.Strings(orderedcmdtxt)
		fmt.Fprint(flag.CommandLine.Output(), "\nUsage: nebulant [flags] [command]\n")
		fmt.Fprint(flag.CommandLine.Output(), "\nFlags:\n")
		subcom.PrintDefaults(flag.CommandLine)
		fmt.Fprint(flag.CommandLine.Output(), "\nCommands:\n")
		for _, cmdtxt := range orderedcmdtxt {
			cmd := commands[cmdtxt]
			if cmd.sec == secruntime {
				runtimecmds = append(runtimecmds, cmd.help)
				continue
			}
			fmt.Fprint(flag.CommandLine.Output(), cmd.help)
		}
		fmt.Fprint(flag.CommandLine.Output(), "\n\nRuntime commands:\n")
		for _, hh := range runtimecmds {
			fmt.Fprint(flag.CommandLine.Output(), hh)
		}
		// fmt.Fprint(flag.CommandLine.Output(), "  readvar\t\t"+term.EmojiSet["Key"]+" Read blueprint variable value during runtime\n")
		fmt.Fprint(flag.CommandLine.Output(), "\n\nrun nebulant [command] --help to show help for a command\n")
	}

	flag.Parse()

	// Version and exit
	if *config.VersionFlag {
		fmt.Println("v" + config.Version)
		fmt.Println("Build date: " + config.VersionDate)
		fmt.Println("Build commit: " + config.VersionCommit)
		fmt.Println("Compiler version: " + config.VersionGo)
		os.Exit(0)
	}

	// Debug
	if *config.DebugFlag {
		config.DEBUG = true
	}

	sc := flag.Arg(0)
	if sc == "" {
		sc = "interactive"
	}
	if cmd, exists := commands[sc]; exists {
		if cmd.upgradeTerm {
			// Init Term
			err = term.UpgradeTerm()
			if err != nil {
				cast.SBus.Close().Wait()
				fmt.Println("cannot init term :(")
				fmt.Println(err.Error())
				os.Exit(1)
			}
			term.ConfigColors()
		}

		// Init console logger
		cast.InitConsoleLogger()
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

		// finally run command
		cmd.run()
	} else {
		flag.Usage()
		cast.LogErr("Unknown command", nil)
		cast.SBus.Close().Wait()
		os.Exit(1)
	}

	// None to wait if director hasn't been started
	if executive.MDirector != nil {
		executive.MDirector.Wait() // None to wait if director has stoped
	}
	//
	// Please don't print anything here, SBus is still closing (because defer)
	// there are still messages in the logger buffer that have to be processed.
}
