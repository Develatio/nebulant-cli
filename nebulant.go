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
	"net"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	// hey hacker:
	// uncomment for profiling
	// _ "net/http/pprof"
	// grmon "github.com/bcicen/grmon/agent"

	"github.com/develatio/nebulant-cli/assets"
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
		if exitCode > 0 {
			cast.LogErr("Done with status "+strconv.Itoa(exitCode), nil)
		} else {
			cast.LogInfo("Done with status "+strconv.Itoa(exitCode), nil)
		}
		cast.SBus.Close().Wait()
		os.Exit(exitCode)
	}()

	config.ServerModeFlag = flag.Bool("s", false, "Server mode to be used within Nebulant Pipeline Builder.")
	config.AddrFlag = flag.String("b", config.SERVER_ADDR+":"+config.SERVER_PORT, "Bind addr:port (ipv4) or [::1]:port (ipv6) for server mode.")
	config.VersionFlag = flag.Bool("v", false, "Show version and exit.")
	config.DebugFlag = flag.Bool("x", false, "Enable debug.")
	config.Ipv6Flag = flag.Bool("6", false, "Force ipv6")
	config.ColorFlag = flag.Bool("c", false, "Disable colors.")
	config.UpgradeAssetsFlag = flag.Bool("u", false, "Upgrade assets from remote location and exit. Build search index as needed.")
	config.ForceUpgradeAssetsFlag = flag.Bool("uu", false, "Force upgrade assets.")
	config.LookupAssetFlag = flag.String("l", "", "Test asset search and exit. Use assetid:searchterm:offset:limit:sort syntax. Ej. nebulant -l \"aws_images:linux x86:10:5:-$.Name\"")
	config.NoTermFlag = flag.Bool("nt", false, "Disable term capabilities. This also disables color.")

	flag.Parse()
	args := flag.Args()
	bluePrintFilePath := flag.Arg(0)

	if *config.NoTermFlag {
		*config.ColorFlag = true
	}

	var tcpaddr *net.TCPAddr
	var err error
	if *config.Ipv6Flag {
		tcpaddr, err = net.ResolveTCPAddr("tcp6", *config.AddrFlag)
	} else {
		tcpaddr, err = net.ResolveTCPAddr("tcp", *config.AddrFlag)
	}
	if err != nil {
		util.PrintUsage(err)
		os.Exit(1)
	}
	if (*config.UpgradeAssetsFlag || *config.ForceUpgradeAssetsFlag) && *config.ServerModeFlag {
		util.PrintUsage(fmt.Errorf("server mode and force asset upgrading are incompatible flags. Set only one of both"))
		os.Exit(1)
	}

	host, port, err := net.SplitHostPort(tcpaddr.String())
	if err != nil {
		util.PrintUsage(err)
		os.Exit(1)
	}

	config.SERVER_ADDR = host
	config.SERVER_PORT = port

	// Version and exit
	if *config.VersionFlag {
		fmt.Println("Nebulant CLI - A cloud builder by develat.io", "v"+config.Version, config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		os.Exit(0)
	}

	// Debug
	if *config.DebugFlag {
		config.DEBUG = true
	}

	// Init Term
	term.InitTerm()

	_, err = term.Println(term.Purple+"Nebulant CLI"+term.Reset, "- A cloud builder by", term.Cyan+"develat.io"+term.Reset)
	if err != nil {
		fmt.Println("Nebulant CLI - A cloud builder by develat.io")
	}
	_, err = term.Println(term.Gray+"Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler, term.Reset)
	if err != nil {
		fmt.Println("Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler)
	}

	// Init console logger
	cast.InitConsoleLogger()
	if *config.UpgradeAssetsFlag || *config.ForceUpgradeAssetsFlag {
		err := assets.UpgradeAssets(*config.ForceUpgradeAssetsFlag)
		if err != nil {
			util.PrintUsage(err)
			os.Exit(1)
		}
		cast.SBus.Close().Wait()
		os.Exit(0)
	}

	if len(*config.LookupAssetFlag) > 0 {
		cut := strings.Split(*config.LookupAssetFlag, ":")
		if len(cut) <= 1 {
			util.PrintUsage(fmt.Errorf("invalid search syntax "))
			os.Exit(1)
		}
		assetid := cut[0]
		term := cut[1]
		assetdef, ok := assets.AssetsDefinition[assetid]
		if !ok {
			util.PrintUsage(fmt.Errorf("unknown asset id"))
			os.Exit(1)
		}
		cast.LogInfo("Looking for "+term+" in "+assetid, nil)
		srq := &assets.SearchRequest{SearchTerm: term}
		if len(cut) > 2 {
			srq.Offset, err = strconv.Atoi(cut[2])
			if err != nil {
				util.PrintUsage(fmt.Errorf("invalid search pagination offset"))
				os.Exit(1)
			}
		}
		if len(cut) > 3 {
			srq.Limit, err = strconv.Atoi(cut[3])
			if err != nil {
				util.PrintUsage(fmt.Errorf("invalid search pagination limit"))
				os.Exit(1)
			}
		}
		if len(cut) > 4 {
			srq.Sort = cut[4]
		}

		cast.LogDebug("lookup "+fmt.Sprintf("%v", srq), nil)
		searchres, err := assets.Search(srq, assetdef)
		if err != nil {
			util.PrintUsage(err)
			os.Exit(1)
		}
		cast.LogInfo("Found "+fmt.Sprintf("%v", searchres.Count)+" items", nil)

		for e, item := range searchres.Results {
			cast.LogInfo(fmt.Sprintf("Result %v / %v -> %v", e, searchres.Count, item), nil)
			if e >= 10 {
				cast.LogInfo("[...]", nil)
				break
			}
		}
		cast.LogInfo("Done.", nil)
		cast.SBus.Close().Wait()
		os.Exit(0)
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

	term.PrintInfo("Welcome :)\n")

	if bluePrintFilePath != "" {
		cast.LogInfo("Processing blueprint...", nil)
		irb, err := blueprint.NewIRBFromAny(bluePrintFilePath)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			exitCode = 1
			return
		}
		// Director in one run mode
		err = executive.InitDirector(false, false)
		if err != nil {
			cast.LogErr(err.Error(), nil)
			exitCode = 1
			panic(err.Error())
		}
		executive.MDirector.HandleIRB <- irb
	} else if *config.ServerModeFlag {
		// Director in server mode
		err := executive.InitDirector(true, false) // Server mode
		if err != nil {
			cast.LogErr(err.Error(), nil)
			panic(err.Error())
		}
		executive.InitServerMode(config.SERVER_ADDR, config.SERVER_PORT)
	} else if len(args) <= 0 {
		// Interactive mode
		err := interactive.Loop()
		if err != nil {
			if err == term.ErrInterrupt {
				fmt.Println("^C")
				os.Exit(0)
			}
			if err == term.ErrEOF {
				fmt.Println("^D")
				os.Exit(0)
			}
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
