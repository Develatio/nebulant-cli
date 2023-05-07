// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

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

package subcom

import (
	"flag"
	"fmt"
	"net"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/util"
)

func parseServeFs() (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	config.AddrFlag = fs.String("b", config.SERVER_ADDR+":"+config.SERVER_PORT, "Bind addr:port (ipv4) or [::1]:port (ipv6)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant serve [options]\n")
		fmt.Fprintf(fs.Output(), "\nOptions:\n")
		util.PrintDefaults(fs)
	}
	fs.Parse(flag.Args()[1:])

	var tcpaddr *net.TCPAddr
	var err error
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

func ServeCmd() (int, error) {
	_, err := parseServeFs()
	if err != nil {
		return 1, err
	}

	// Director in server mode
	err = executive.InitDirector(true, false) // Server mode
	if err != nil {
		cast.LogErr(err.Error(), nil)
		panic(err.Error())
	}
	executive.InitServerMode(config.SERVER_ADDR, config.SERVER_PORT)

	executive.ServerWaiter.Wait() // None to wait if server mode is disabled
	if executive.ServerError != nil {
		return 2, executive.ServerError
		// exitCode = 1
		// cast.LogErr(executive.ServerError.Error(), nil)
		// panic(executive.ServerError.Error())
	}
	executive.MDirector.Wait() // None to wait if director has stoped
	return 0, nil
}
