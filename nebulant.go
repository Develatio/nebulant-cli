// MIT License
//
// Copyright (C) 2021  Develatio Technologies S.L.

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
	"fmt"
	"os"
	"runtime/debug"
	"strconv"

	// hey hacker:
	// uncomment for profiling
	// _ "net/http/pprof"
	// grmon "github.com/bcicen/grmon/agent"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/cli"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/nhttpd"
	"github.com/develatio/nebulant-cli/util"
)

var exitCode = 0

func main() {
	// exitCode := 0
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
			cast.LogErr("exit with status "+strconv.Itoa(exitCode), nil)
		} else {
			cast.LogInfo("exit with status "+strconv.Itoa(exitCode), nil)
		}
		executive.RemoveDirector()
		cast.SBus.Close().Wait()
		os.Exit(exitCode)
	}()

	go func() {
		count := 0
		for {
			<-base.InterruptSignalChannel
			if count > 0 {
				fmt.Println(" Force quit")
				os.Exit(1)
			}
			count++
			if executive.MDirector != nil {
				select {
				case executive.MDirector.ExecInstruction <- &executive.ExecCtrlInstruction{Instruction: executive.ExecShutdown}:
					cast.LogInfo("gracefully shutdown started...", nil)
					nhttpd.GetServer().Shutdown()
				default:
					cast.LogErr("cannot gracefully shutdown cli", nil)
				}
			}
		}
	}()

	exitCode = cli.Start()

	// None to wait if director hasn't been started
	if executive.MDirector != nil {
		executive.MDirector.Wait() // None to wait if director has stoped
	}
	//
	// Please don't print anything here, SBus is still closing (because defer)
	// there are still messages in the logger buffer that have to be processed.
	//

	// defer func() at start of this main() should be the last defer to exec
}
