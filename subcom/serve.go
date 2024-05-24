// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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

package subcom

import (
	"flag"
	"fmt"
	"net"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
)

func parseServeFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	config.AddrFlag = fs.String("b", config.SERVER_ADDR+":"+config.SERVER_PORT, "Bind addr:port (ipv4) or [::1]:port (ipv6)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant serve [options]\n")
		fmt.Fprintf(fs.Output(), "\nOptions:\n")
		subsystem.PrintDefaults(fs)
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}

	var tcpaddr *net.TCPAddr
	if *config.Ipv6Flag {
		tcpaddr, err = net.ResolveTCPAddr("tcp6", *config.AddrFlag)
		if err != nil {
			return fs, err
		}
	} else {
		tcpaddr, err = net.ResolveTCPAddr("tcp", *config.AddrFlag)
		if err != nil {
			return fs, err
		}
	}
	host, port, err := net.SplitHostPort(tcpaddr.String())
	if err != nil {
		return fs, err
	}

	config.SERVER_ADDR = host
	config.SERVER_PORT = port

	return fs, nil
}

func ServeCmd(nblc *subsystem.NBLcommand) (int, error) {
	_, err := parseServeFs(nblc.CommandLine())
	if err != nil {
		return 1, err
	}

	// Director in server mode
	err = executive.InitDirector(true, false) // Server mode
	if err != nil {
		cast.LogErr(err.Error(), nil)
		panic(err.Error())
	}
	errc := executive.InitServerMode()
	err = <-errc
	if err != nil {
		return 2, err
	}
	executive.MDirector.Wait() // None to wait if director has stoped
	return 0, nil
}
