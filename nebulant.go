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

	"sync"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/providers/aws"
	"github.com/develatio/nebulant-cli/providers/azure"
	"github.com/develatio/nebulant-cli/providers/generic"
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
	var versionFlag = flag.Bool("v", false, "Show version and exit")
	var debugFlag = flag.Bool("vv", false, "Enable debug")
	var mFlag = flag.Bool("m", false, "Disable colors.")

	flag.Parse()
	args := flag.Args()
	bluePrintFilePath := flag.Arg(0)

	if *versionFlag {
		fmt.Println("Nebulant - A cloud builder by develat.io", config.Version, config.VersionDate)
		os.Exit(0)
	}

	// Need at least a file config or server mode
	if len(args) <= 0 && !*serverModeFlag {
		fmt.Println("Nebulant - A cloud builder by develat.io", config.Version, config.VersionDate)
		fmt.Println("")
		fmt.Println("Usage: nebulant [-options] <nebulantblueprint.json>")
		fmt.Println("")
		flag.PrintDefaults()
		return
	}

	if *debugFlag {
		config.DEBUG = true
	}

	// Providers
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

	// Init console logger
	cast.InitConsoleLogger(!*mFlag)
	cast.LogInfo("Welcome to Nebulant :)", nil)

	// if "c" flag, start new blueprint and send it to Director
	var bp *blueprint.Blueprint
	if bluePrintFilePath != "" {
		var err error

		if len(bluePrintFilePath) > 11 && bluePrintFilePath[:11] == "nebulant://" {
			bp, err = blueprint.NewFromBackend(bluePrintFilePath[11:])
			if err != nil {
				cast.LogErr(err.Error(), nil)
				exitCode = 1
				return // To keep running on server mode even if bp fails, comment this
			}
		} else {
			bp, err = blueprint.NewFromFile(bluePrintFilePath)
			if err != nil {
				cast.LogErr(err.Error(), nil)
				exitCode = 1
				return // To keep running on server mode even if bp fails, comment this
			}
		}
	}

	if !*serverModeFlag && bp == nil {
		cast.LogErr("Nothing to do, exiting...", nil)
		return
	}

	serverWaiter := &sync.WaitGroup{}
	if !*serverModeFlag {
		// Director
		executive.InitDirector(false) // NO server mode: run one bp and exit
	} else {
		// Director
		executive.InitDirector(true) // Server mode
		serverWaiter.Add(1)          // Append +1 waiter
		go func() {
			defer serverWaiter.Done() // Append -1 waiter
			srv := &executive.Httpd{}
			addr := "localhost:15678"
			err := srv.Serve(&addr) // serve to address serverModeFlag
			if err != nil {
				cast.LogErr(err.Error(), nil)
				panic(err.Error())
			}
		}()
	}

	if bp != nil {
		cast.LogInfo("Running blueprint file...", nil)
		irb, irbErr := blueprint.GenerateIRB(bp, &blueprint.IRBGenConfig{})
		if irbErr != nil {
			cast.LogErr(irbErr.Error(), nil)
			if !*serverModeFlag {
				cast.LogErr(irbErr.Error(), nil)
				panic(irbErr.Error())
			}
		} else {
			executive.MDirector.HandleIRB <- irb
		}
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

	serverWaiter.Wait()        // None to wait if server mode is disabled
	executive.MDirector.Wait() // None to wait if director has stoped
	//
	// Please don't print anything here, SBus is still closing (because defer)
	// there are still messages in the logger buffer that have to be processed.
}
