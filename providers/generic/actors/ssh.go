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
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	nebulantssh "github.com/develatio/nebulant-cli/netproto/ssh"
	"github.com/develatio/nebulant-cli/nsterm"
	"github.com/develatio/nebulant-cli/util"
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

// type generateKeypairParameters struct {
// }

type runRemoteParameters struct {
	nebulantssh.ClientConfigParameters
	// Proxies     []*nebulantssh.ClientConfigParameters `json:"proxies"`
	ScriptPath *string           `json:"scriptPath"`
	ScriptText *string           `json:"script"`
	Command    *string           `json:"command"`
	Vars       map[string]string `json:"vars"`
	// VarsTargets []string          `json:"vars_targets"`
	DumpJSON            *bool `json:"dump_json"`
	OpenDbgShellAfter   bool  `json:"open_dbg_shell_after"`
	OpenDbgShellBefore  bool  `json:"open_dbg_shell_before"`
	OpenDbgShellOnerror bool  `json:"open_dbg_shell_onerror"`
}

type runRemoteScriptOutput struct {
	Stdout   *bytes.Buffer `json:"stdout"`
	Stderr   *bytes.Buffer `json:"stderr"`
	Error    error         `json:"error"`
	ExitCode string        `json:"exit_code"`
}

func newSSHDebugShell(ctx *ActionContext, sshClient *nebulantssh.SSHClient) error {
	mst := ctx.GetMustarFD()
	svu := ctx.GetSluvaFD()
	// ctx.Logger.LogErr(errors.Join(fmt.Errorf("remote exec fail"), sshRunErr.(error)).Error())
	dbgsession, err := sshClient.NewSessionPthShellWithOpts(&nebulantssh.SessOpts{
		Stdin:  svu,
		Stdout: svu,
		Stderr: svu,
	})
	if err != nil {
		return err
	}
	ctx.DebugInit()
	ctx.Logger.LogInfo("waiting for debug session to finish")
	err = dbgsession.Wait()
	ctx.Logger.LogInfo("debug session finished")
	// svu.Close()
	mst.Close() // force close of mustar
	// ctx.Logger.LogInfo("debugger finished")
	return err
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

	err = ctx.Store.DeepInterpolation(p)
	if err != nil {
		return nil, err
	}

	var sshRunErr interface{}
	combineOut := true

	remoteAddress := p.Target
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
		// closing main client
		// will close subclients
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
					ctx.Logger.LogInfo(fmt.Sprintf("SSH Dialing %v...", addr))
				}
				if evt.Type == nebulantssh.SSHClientEventClosed {
					ctx.Logger.LogDebug(fmt.Sprintf("SSH Closing %v...", addr))
				}
				if evt.Type == nebulantssh.SSHClientEventError {
					if evt.Error != io.EOF {
						ctx.Logger.LogWarn(evt.Error.Error())
					}
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

	// the remote ipc server will be closed
	// automaticallly on sshClient.Close()
	ipcc, err := sshClient.StartIPC()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("start of remote IPC fail"), err)
	}
	ctx.Logger.LogDebug("Exposing vars to remote unix socket...")
	outexpose := ipcc.ExposeStoreVars(ctx.Store)
	defer func() {
		// close the unix sock requests dispatcher
		outexpose <- true
	}()

	result := &runRemoteScriptOutput{}
	var sshOut io.Writer
	var sshErr io.Writer
	sshVpty := nsterm.NewVirtPTY()
	sshVpty.SetLDisc(nsterm.NewRawLdisc())
	sshmfd := sshVpty.MustarFD()
	sshStdin := sshVpty.SluvaFD()

	result.Stdout = new(bytes.Buffer)
	result.Stderr = new(bytes.Buffer)
	if combineOut {
		result.Stdout = new(bytes.Buffer)
		result.Stderr = result.Stdout
	}

	logStdoutSwch := util.NewSwitchableWriter(&logWriter{
		Log:       ctx.Logger.ByteLogInfo,
		LogPrefix: []byte(*p.Target + ":ssh> "),
	})
	resultStdoutSwch := util.NewSwitchableWriter(result.Stdout)
	logStderrSwch := util.NewSwitchableWriter(&logWriter{
		Log:       ctx.Logger.ByteLogErr,
		LogPrefix: []byte(*p.Target + ":ssh> "),
	})
	resultStderrSwch := util.NewSwitchableWriter(result.Stderr)

	// start writers off

	logStdoutSwch.Off()
	resultStdoutSwch.Off()
	logStderrSwch.Off()
	resultStderrSwch.Off()

	sshOut = io.MultiWriter(result.Stdout, &logWriter{
		Log:       ctx.Logger.ByteLogInfo,
		LogPrefix: []byte(*p.Target + ":ssh> "),
	})
	sshErr = io.MultiWriter(result.Stderr, &logWriter{
		Log:       ctx.Logger.ByteLogErr,
		LogPrefix: []byte(*p.Target + ":ssh> "),
	})

	session, err := sshClient.NewSessionPthShellWithOpts(&nebulantssh.SessOpts{
		Stdout: sshOut,
		Stderr: sshErr,
		Stdin:  sshStdin,
	})
	if err != nil {
		return nil, err
	}

	ctx.Logger.LogInfo("Setting env vars...")
	// WIP: prevent log sensible vars
	if p.Vars != nil {
		for key, val := range p.Vars {
			vv := val
			err = ctx.Store.Interpolate(&vv)
			if err != nil {
				return nil, err
			}
			// too asumptions?
			sshmfd.Write([]byte(fmt.Sprintf("export %s=$( cat <<EOF\n%s\nEOF\n)\n", key, vv)))
			// sshClient.Env[key] = vv
		}
	}

	if p.DumpJSON != nil && *p.DumpJSON {
		ctx.Logger.LogInfo("Uploading a dump of json vars...")
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

	// if p.OpenDbgShellBefore {
	// 	ctx.Logger.LogInfo("Opening debug before run...")
	// 	ctx.DebugInit()
	// }

	if p.Command != nil { // run cmd
		sshmfd.Write([]byte(*p.Command))
		sshmfd.Write([]byte("\n"))
		sshmfd.Write([]byte("exit $?"))
		sshmfd.Write([]byte("\n"))
	} else if p.ScriptPath != nil { // upload local script and run
		// TODO: configurable dst path
		scriptpath := fmt.Sprintf("/tmp/nscript_%d.sh", rand.Int())
		ctx.Logger.LogInfo(fmt.Sprintf("Uploading local:%s -> remote:%s script...", *p.ScriptPath, scriptpath))
		scpClient, err := sshClient.NewSCPClientFromExistingSSH()
		if err != nil {
			ctx.Logger.LogErr("errr1" + err.Error())
			return nil, err
		}
		err = scpClient.CopyFileToRemote(*p.ScriptPath, scriptpath, &scp.FileTransferOption{Perm: 0700})
		if err != nil {
			ctx.Logger.LogErr("errr2" + err.Error())
			return nil, err
		}
		sshmfd.Write([]byte(scriptpath))
		sshmfd.Write([]byte("\n"))
		sshmfd.Write([]byte("exit $?"))
		sshmfd.Write([]byte("\n"))
	} else if p.ScriptText != nil {
		// TODO: configurable dst path
		scriptpath := fmt.Sprintf("/tmp/nscript_%d.sh", rand.Int())
		ctx.Logger.LogInfo(fmt.Sprintf("Uploading remote:%s script...", scriptpath))
		scpClient, err := sshClient.NewSCPClientFromExistingSSH()
		if err != nil {
			ctx.Logger.LogErr("errr3" + err.Error())
			return nil, err
		}
		rr := bytes.NewReader([]byte(*p.ScriptText))
		err = scpClient.CopyToRemote(rr, scriptpath, &scp.FileTransferOption{Perm: 0700})
		if err != nil {
			ctx.Logger.LogErr("errr4" + err.Error())
			return nil, err
		}
		sshmfd.Write([]byte(scriptpath))
		sshmfd.Write([]byte("\n"))
		sshmfd.Write([]byte("exit $?"))
		sshmfd.Write([]byte("\n"))
	} else {
		return nil, fmt.Errorf("no script provided")
	}

	// TODO: capture ^D to exit debug
	// and keep shell alive
	// if p.OpenDbgShellAfter {
	// 	ctx.Logger.LogInfo("Opening debug after shell...")
	// 	ctx.DebugInit()
	// }

	// before

	ctx.Logger.LogInfo("Waiting shell to finish...")
	sshRunErr = session.Wait()
	ctx.Logger.LogInfo("Finished")
	fmt.Println(p)
	if sshRunErr != nil && p.OpenDbgShellOnerror {
		ctx.Logger.LogErr(errors.Join(fmt.Errorf("remote exec fail"), sshRunErr.(error)).Error())
		ctx.Logger.LogInfo("waiting for debug session to finish")
		sshRunErr = newSSHDebugShell(ctx, sshClient)
		ctx.Logger.LogInfo("debugger finished")
		// if err != nil {
		// 	return nil, err
		// }
		// ctx.DebugInit()
		// ctx.Logger.LogInfo("waiting for debug session to finish")
		// sshRunErr = dbgsession.Wait()
		// ctx.Logger.LogInfo("debug session finished")

		// sshRunErr =

		// mst := ctx.GetMustarFD()
		// svu := ctx.GetSluvaFD()

		// dbgsession, err := sshClient.NewSessionPthShellWithOpts(&nebulantssh.SessOpts{
		// 	Stdin:  svu,
		// 	Stdout: svu,
		// 	Stderr: svu,
		// })
		// if err != nil {
		// 	return nil, err
		// }

		// svu.Close()
		//mst.Close() // force close of mustar

	}

	// after
	// ctx.DebugInit()

	// if p.Command != nil { // run cmd
	// 	sshRunErr = sshClient.RunCmd(*p.Command)
	// } else if p.ScriptPath != nil { // upload local script and run
	// 	sshRunErr = sshClient.RunScriptFromLocalPath(*p.ScriptPath)
	// } else if p.ScriptText != nil {
	// 	sshRunErr = sshClient.RunScriptFromText(p.ScriptText)
	// } else {
	// 	return nil, fmt.Errorf("no script provided")
	// }
	ctx.Logger.ByteLogInfo([]byte("\n----------\n"))

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

	ctx.Logger.ByteLogInfo([]byte("\nout of remotescript actionn"))
	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, err
}

// TODO:
// func GenerateKeyPair(ctx *ActionContext) (*base.ActionOutput, error) {
// 	var err error
// 	input := &generateKeypairParameters{}
// 	if err = json.Unmarshal(ctx.Action.Parameters, input); err != nil {
// 		return nil, err
// 	}

// 	if ctx.Rehearsal {
// 		return nil, nil
// 	}

// 	err = ctx.Store.DeepInterpolation(input)
// 	if err != nil {
// 		return nil, err
// 	}
// }
