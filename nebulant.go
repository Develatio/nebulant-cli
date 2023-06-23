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
	"github.com/develatio/nebulant-cli/providers/generic"
	"github.com/develatio/nebulant-cli/subcom"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

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

	// Init Term
	term.InitTerm()

	config.VersionFlag = flag.Bool("v", false, "Show version and exit.")
	config.DebugFlag = flag.Bool("x", false, "Enable debug.")
	config.Ipv6Flag = flag.Bool("6", false, "Force ipv6")
	config.DisableColorFlag = flag.Bool("c", false, "Disable colors.")
	config.ForceTerm = flag.Bool("ft", false, "Force terminal. Bypass no-term detection.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage: nebulant [options] [command]\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nGlobal options:\n")
		subcom.PrintDefaults(flag.CommandLine)
		fmt.Fprintf(flag.CommandLine.Output(), "\nCommands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  serve\t\t\t"+term.EmojiSet["TridentEmblem"]+" Start server mode\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  run\t\t\t"+term.EmojiSet["RunningShoe"]+" Run blueprint form file or net\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  assets\t\t"+term.EmojiSet["Squid"]+" Handle cli assets\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  interactive\t\t"+term.EmojiSet["Television"]+" Start interactive menu\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n\nrun nebulant [command] --help to show help for a command\n")
	}

	flag.Parse()
	term.ConfigColors()

	_, err = term.Println(term.Magenta+"Nebulant CLI"+term.Reset, "- A cloud builder by", term.Blue+"develat.io"+term.Reset)
	if err != nil {
		fmt.Println("Nebulant CLI - A cloud builder by develat.io")
	}
	_, err = term.Println(term.Gray+" Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler, term.Reset)
	if err != nil {
		fmt.Println("Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler)
	}

	// Version and exit
	if *config.VersionFlag {
		os.Exit(0)
	}

	// Debug
	if *config.DebugFlag {
		config.DEBUG = true
	}

	term.PrintInfo(" Welcome :)\n")

	// Init console logger
	cast.InitConsoleLogger()
	if config.DEBUG {
		cast.LogDebug("Debug mode activated. Testing message levels...", nil)
		cast.LogInfo("Info message", nil)
		cast.LogWarn("Warning message", nil)
		cast.LogErr("Error message", nil)
		cast.LogCritical("Critical message", nil)
	}

	// Init Providers
	cast.SBus.RegisterProviderInitFunc("aws", aws.New)
	cast.SBus.RegisterProviderInitFunc("azure", azure.New)
	cast.SBus.RegisterProviderInitFunc("generic", generic.New)
	blueprint.ActionValidators["providerValidator"] = func(action *blueprint.Action) error {
		if _, err := cast.SBus.GetProviderInitFunc(action.Provider); err != nil {
			return err
		}
		return nil
	}
	blueprint.ActionValidators["awsValidator"] = aws.ActionValidator
	blueprint.ActionValidators["azureValidator"] = azure.ActionValidator
	blueprint.ActionValidators["genericsValidator"] = generic.ActionValidator

	sc := flag.Arg(0)
	switch sc {
	case "serve":
		exitCode, err = subcom.ServeCmd()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cast.SBus.Close().Wait()
			os.Exit(exitCode)
		}
	case "run":
		exitCode, err = subcom.RunCmd()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cast.SBus.Close().Wait()
			os.Exit(exitCode)
		}
	case "assets":
		exitCode, err = subcom.AssetsCmd()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cast.SBus.Close().Wait()
			os.Exit(exitCode)
		}
	case "auth":
		exitCode, err = subcom.AuthCmd()
		if err != nil {
			cast.LogErr(err.Error(), nil)
			cast.SBus.Close().Wait()
			os.Exit(exitCode)
		}
	case "", "interactive":
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
			exitCode = 1
			cast.LogErr(err.Error(), nil)
			os.Exit(exitCode)
		}
		cast.SBus.Close().Wait()
		os.Exit(0)
	default:
		flag.Usage()
		cast.LogErr("Unknown command", nil)
		cast.SBus.Close().Wait()
		os.Exit(1)
	}

	// hey hacker:
	// uncomment for profiling
	// if config.PROFILING {
	// 	go func() {
	// 		cast.LogInfo("Starting profiling at localhost:6060", nil)
	// 		http.ListenAndServe("localhost:6060", nil)
	// 	}()
	// 	grmon.Start()
	// }

	// None to wait if director hasn't been started
	if executive.MDirector != nil {
		executive.MDirector.Wait() // None to wait if director has stoped
	}
	//
	// Please don't print anything here, SBus is still closing (because defer)
	// there are still messages in the logger buffer that have to be processed.
}
