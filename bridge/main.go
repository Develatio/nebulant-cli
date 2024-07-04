// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

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
	cast.InitConsoleLogger(nil)
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
	config.BridgeCertPathFlag = flag.String("c", "", "https/wss cert file path.")
	config.BridgeKeyPathFlag = flag.String("k", "", "https/wss key file path.")
	config.BridgeXtermRootPath = flag.String("w", "", "webroot path for custom xtermjs implementation")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *config.DebugFlag {
		config.DEBUG = true
	}

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
