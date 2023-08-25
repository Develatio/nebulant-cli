// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

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

package actors

// Considerations:
// - Only one instance of runActor per script or cmd. Keep in mind that for each
// execution there must be an output and it must be stored, so the functionality
// of executing multiple scripts with an instance of runActor should not be
// implemented.
//

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	nebulantssh "github.com/develatio/nebulant-cli/netproto/ssh"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"
)

// type ConnConfig struct {
// 	Target               *string `json:"target" validate:"required"`
// 	Username             *string `json:"username" validate:"required"`
// 	PrivateKey           *string `json:"privkey"`
// 	PrivateKeyPath       *string `json:"privkeyPath"`
// 	PrivateKeyPassphrase *string `json:"passphrase"`
// 	Password             *string `json:"password"`
// 	Port                 uint16  `json:"port"`
// }

type runRemoteParameters struct {
	nebulantssh.ClientConfigParameters
	// Proxies     []*nebulantssh.ClientConfigParameters `json:"proxies"`
	ScriptPath *string           `json:"scriptPath"`
	ScriptText *string           `json:"script"`
	Command    *string           `json:"command"`
	Vars       map[string]string `json:"vars"`
	// VarsTargets []string          `json:"vars_targets"`
	DumpJSON *bool `json:"dump_json"`
}

type runRemoteScriptOutput struct {
	Stdout   *bytes.Buffer `json:"stdout"`
	Stderr   *bytes.Buffer `json:"stderr"`
	ExitCode string        `json:"exit_code"`
}

// RunRemoteScript func
func RunRemoteScript(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	p := &runRemoteParameters{}
	if err = json.Unmarshal(ctx.Action.Parameters, p); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	if err = ctx.Store.Interpolate(p.Target); err != nil {
		return nil, err
	}

	var sshRunErr interface{}
	combineOut := true

	remoteAddress := p.Target
	err = ctx.Store.Interpolate(remoteAddress)
	if err != nil {
		return nil, err
	}
	if strings.Trim(*remoteAddress, " ") == "" {
		return nil, fmt.Errorf("the target addr is empty. Please provide one")
	}
	var proxies []*nebulantssh.ClientConfigParameters
	for _, prx := range p.Proxies {
		raddr := prx.Target
		err = ctx.Store.Interpolate(raddr)
		if err != nil {
			return nil, err
		}
		if strings.Trim(*raddr, " ") == "" {
			return nil, fmt.Errorf("the proxy target addr is empty. Please provide one")
		}
		proxies = append(proxies, &nebulantssh.ClientConfigParameters{
			Target:               raddr,
			Port:                 prx.Port,
			Username:             prx.Username,
			PrivateKey:           prx.PrivateKey,
			PrivateKeyPath:       prx.PrivateKeyPath,
			PrivateKeyPassphrase: prx.PrivateKeyPassphrase,
			Password:             prx.Password,
			Proxies:              prx.Proxies,
		})
	}

	sshClient := nebulantssh.NewSSHClient()
	mainclient := sshClient
	out := make(chan bool)
	defer func() {
		err := mainclient.Disconnect()
		if err != nil {
			ctx.Logger.LogWarn(err.Error())
		}
		out <- true
	}()
	mainClientEvents := sshClient.Events
	go func() {
	L1:
		for {
			select {
			case evt := <-mainClientEvents:
				addr := evt.SSHClient.DialAddr
				if evt.Type == nebulantssh.SSHClientEventMasterClosed {
					ctx.Logger.LogDebug(fmt.Sprintf("SSH Closing %v...", addr))
					break L1
				}
				if evt.Type == nebulantssh.SSHClientEventDialing {
					ctx.Logger.LogDebug(fmt.Sprintf("SSH Dialing %v...", addr))
				}
				if evt.Type == nebulantssh.SSHClientEventClosed {
					ctx.Logger.LogDebug(fmt.Sprintf("SSH Closing %v...", addr))
				}
			case <-out:
				break L1
			default:
				ctx.Logger.LogDebug("Waiting ssh event...")
				time.Sleep(200000 * time.Microsecond)
			}
		}
	}()
	sshClient, err = sshClient.DialWithProxies(&nebulantssh.ClientConfigParameters{
		Target:               remoteAddress,
		Port:                 p.Port,
		Username:             p.Username,
		PrivateKey:           p.PrivateKey,
		PrivateKeyPath:       p.PrivateKeyPath,
		PrivateKeyPassphrase: p.PrivateKeyPassphrase,
		Password:             p.Password,
		Proxies:              proxies,
	})
	if err != nil {
		return nil, err
	}

	result := &runRemoteScriptOutput{}
	var sshOut io.Writer
	var sshErr io.Writer

	result.Stdout = new(bytes.Buffer)
	result.Stderr = new(bytes.Buffer)
	if combineOut {
		result.Stdout = new(bytes.Buffer)
		result.Stderr = result.Stdout
	}
	sshOut = io.MultiWriter(result.Stdout, &logWriter{
		Log:       ctx.Logger.ByteLogInfo,
		LogPrefix: []byte(*p.Target + ":ssh> "),
	})
	sshErr = io.MultiWriter(result.Stderr, &logWriter{
		Log:       ctx.Logger.ByteLogErr,
		LogPrefix: []byte(*p.Target + ":ssh> "),
	})

	sshClient.Stderr = sshErr
	sshClient.Stdout = sshOut
	if p.Vars != nil {
		for key, val := range p.Vars {
			vv := val
			err = ctx.Store.Interpolate(&vv)
			if err != nil {
				return nil, err
			}
			sshClient.Env[key] = vv
		}
	}

	if p.DumpJSON != nil && *p.DumpJSON {
		f, err := ctx.Store.DumpValuesToJSONFile()
		if err != nil {
			return nil, err
		}
		dst := "/tmp/json" + filepath.Base(f.Name())

		sshClient.Env["NEBULANT_JSON_VARIABLES_PATH"] = dst
		defer os.Remove(f.Name())
		src := f.Name()
		scpClient, err := sshClient.NewSCPClientFromExistingSSH()
		if err != nil {
			return nil, err
		}
		err = scpClient.CopyFileToRemote(src, dst, &scp.FileTransferOption{})
		if err != nil {
			return nil, err
		}
	}

	if p.Command != nil { // run cmd
		sshRunErr = sshClient.RunCmd(*p.Command)
	} else if p.ScriptPath != nil { // upload local script and run
		sshRunErr = sshClient.RunScriptFromLocalPath(*p.ScriptPath)
	} else if p.ScriptText != nil {
		sshRunErr = sshClient.RunScriptFromText(p.ScriptText)
	} else {
		return nil, fmt.Errorf("no script provided")
	}

	if sshRunErr == nil {
		result.ExitCode = "0"
	} else if _, isExitError := sshRunErr.(*ssh.ExitError); isExitError {
		// It is cast to a string because the conditional evaluation of
		// this result is defined from the graphical app and in it the
		// value to compare will always be a string.
		result.ExitCode = strconv.Itoa(sshRunErr.(*ssh.ExitError).ExitStatus())
		err = fmt.Errorf("Exit status != 0 (" + result.ExitCode + ")")
	} else if _, isNetError := sshRunErr.(*net.OpError); isNetError {
		err = sshRunErr.(*net.OpError).Err
	} else {
		err = sshRunErr.(error)
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, err
}
