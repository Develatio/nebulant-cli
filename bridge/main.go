// Nebulant
// Copyright (C) 2024  Develatio Technologies S.L.

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
	"log"
	"net"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/develatio/nebulant-cli/bridge/views"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
)

var exitCode = 0

func main() {
	cast.InitSystemBus()
	cast.InitConsoleLogger()
	defer func() {
		if r := recover(); r != nil {
			exitCode = 1
			cast.LogErr("Panic", nil)
			cast.LogErr("If you think this is a bug,", nil)
			cast.LogErr("please consider posting stack trace as a GitHub", nil)
			cast.LogErr("issue (https://github.com/develatio/nebulant-cli/issues)", nil)
			cast.LogErr("Stack Trace:", nil)
			v := fmt.Sprintf("%v", r)
			cast.LogErr(v, nil)
			cast.LogErr(string(debug.Stack()), nil)
		}
		if exitCode > 0 {
			cast.LogErr("exit with status "+strconv.Itoa(exitCode), nil)
		} else {
			cast.LogInfo("exit with status "+strconv.Itoa(exitCode), nil)
		}
		cast.SBus.Close().Wait()
		os.Exit(exitCode)
	}()

	// dflag := flag.NewFlagSet("ndbgbroker", flag.ExitOnError)

	config.BridgeAddrFlag = flag.String("b", net.JoinHostPort(config.BRIDGE_ADDR, config.BRIDGE_PORT), "Bind addr:port (ipv4) or [::1]:port (ipv6).")
	config.VersionFlag = flag.Bool("v", false, "Show version and exit.")
	config.DebugFlag = flag.Bool("x", false, "Enable debug.")
	config.BridgeSecretFlag = flag.String("bs", config.BRIDGE_SECRET, "Auth secret string (overrides env NEBULANT_BRIDGE_SECRET).")
	config.BridgeOriginFlag = flag.String("o", "*", "Access-Control-Allow-Origin header.")
	config.BridgeCertPathFlag = flag.String("c", "", "https/wss cert file path")
	config.BridgeKeyPathFlag = flag.String("k", "", "https/wss key file path")

	if *config.DebugFlag {
		config.DEBUG = true
	}

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(*config.BridgeSecretFlag) <= 0 {
		flag.Usage()
		cast.LogErr("Please provide -bs flag or NEBULANT_BRIDGE_SECRET env", nil)
		exitCode = 1
		return
	}

	if *config.VersionFlag {
		fmt.Println("v" + config.Version)
		fmt.Println("Build date: " + config.VersionDate)
		fmt.Println("Build commit: " + config.VersionCommit)
		fmt.Println("Compiler version: " + config.VersionGo)
		exitCode = 0
		return
	}

	err := views.Serve()
	if err != nil {
		log.Println(err.Error())
		exitCode = 1
		return
	}
}
