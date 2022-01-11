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
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

func main() {
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
		cast.LogErr("Done with status "+strconv.Itoa(exitCode), nil)
		cast.SBus.Close().Wait()
		os.Exit(exitCode)
	}()

	var serverModeFlag = flag.Bool("d", false, "Enable server mode at localhost:15678 to use within Nebulant Pipeline Builder.")
	var versionFlag = flag.Bool("v", false, "Show version and exit.")
	var debugFlag = flag.Bool("vv", false, "Enable debug.")
	var mFlag = flag.Bool("m", false, "Disable colors.")

	flag.Parse()
	args := flag.Args()
	bluePrintFilePath := flag.Arg(0)

	// Version and exit
	if *versionFlag {
		fmt.Println("Nebulant CLI - A cloud builder by develat.io", config.Version, config.VersionDate)
		os.Exit(0)
	}

	// Debug
	if *debugFlag {
		config.DEBUG = true
	}

	// Init Term
	term.InitTerm(!*mFlag)

	// Init console logger
	cast.InitConsoleLogger(!*mFlag)

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

	term.PrintInfo("Welcome to Nebulant :)\n")

	if bluePrintFilePath != "" {
		cast.LogInfo("Obtaining blueprint...", nil)
		irb, err := blueprint.NewIRBFromAny(bluePrintFilePath)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			exitCode = 1
			panic(err.Error())
		}
		// Director in one run mode
		err = executive.InitDirector(false, false)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			exitCode = 1
			panic(err.Error())
		}
		executive.MDirector.HandleIRB <- irb
	} else if *serverModeFlag {
		// Director in server mode
		err := executive.InitDirector(true, false) // Server mode
		if err != nil {
			cast.LogErr(err.Error(), nil)
			panic(err.Error())
		}
		executive.InitServerMode("15678")
	} else if len(args) <= 0 {
		// Interactive mode
		err := interactive.Loop()
		if err != nil {
			exitCode = 1
			cast.LogErr(err.Error(), nil)
			panic(err.Error())
		}
		os.Exit(0)
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

	executive.ServerWaiter.Wait() // None to wait if server mode is disabled
	if executive.ServerError != nil {
		exitCode = 1
		cast.LogErr(executive.ServerError.Error(), nil)
		panic(executive.ServerError.Error())
	}
	executive.MDirector.Wait() // None to wait if director has stoped
	//
	// Please don't print anything here, SBus is still closing (because defer)
	// there are still messages in the logger buffer that have to be processed.
}
