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
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	nebulantssh "github.com/develatio/nebulant-cli/netproto/ssh"
	"golang.org/x/crypto/ssh"
)

type runRemoteParameters struct {
	Target               *string           `json:"target" validate:"required"`
	Username             *string           `json:"username" validate:"required"`
	PrivateKey           *string           `json:"privkey"`
	PrivateKeyPath       *string           `json:"privkeyPath"`
	PrivateKeyPassphrase *string           `json:"passphrase"`
	Password             *string           `json:"password"`
	Port                 uint16            `json:"port"`
	ScriptPath           *string           `json:"scriptPath"`
	ScriptText           *string           `json:"script"`
	Command              *string           `json:"command"`
	Vars                 map[string]string `json:"vars"`
	VarsTargets          []string          `json:"vars_targets"`
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
	err = json.Unmarshal(ctx.Action.Parameters, p)
	if err != nil {
		return nil, err
	}

	err = ctx.Store.Interpolate(p.Target)
	if err != nil {
		return nil, err
	}

	var sshRunErr interface{}
	combineOut := true

	sshConfig := &ssh.ClientConfig{
		User: *p.Username,
		//#nosec G106 -- Allow config this? Hacker comunity feedback needed.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if p.PrivateKeyPath != nil {
		key, err := ioutil.ReadFile(*p.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		// Create the Signer for this private key.
		var signer ssh.Signer
		if p.PrivateKeyPassphrase != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(*p.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else if p.PrivateKey != nil {
		// Create the Signer for this private key.
		var signer ssh.Signer
		if p.PrivateKeyPassphrase != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(*p.PrivateKey), []byte(*p.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(*p.PrivateKey))
		}
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else if p.Password != nil {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*p.Password),
		}
	} else {
		// Use ssh agent for auth
		sshAgent, err := nebulantssh.GetSSHAgentClient()
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(sshAgent.Signers),
		}
	}

	port := "22"
	if p.Port != 0 {
		port = fmt.Sprintf("%d", p.Port)
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

	sshClient := nebulantssh.NewSSHClient(sshOut, sshErr)
	sshClient.Env = p.Vars

	addr := *p.Target + ":" + port
	ctx.Logger.LogDebug("Connecting to addr " + addr)
	connerr := sshClient.Connect(addr, sshConfig)
	if connerr != nil {
		return nil, connerr
	}
	defer sshClient.Disconnect()

	for _, vt := range p.VarsTargets {
		var f *os.File
		var dst string
		switch vt {
		case "bash", "zsh":
			f, err = ctx.Store.DumpValuesToShellFile()
			if err != nil {
				return nil, err
			}
			dst = "/tmp/" + vt + filepath.Base(f.Name())
			if vt == "zsh" {
				sshClient.Env["NEBULANT_BASH_VARIABLES_PATH"] = dst
			} else {
				sshClient.Env["NEBULANT_ZSH_VARIABLES_PATH"] = dst
			}
		case "json":
			f, err = ctx.Store.DumpValuesToJSONFile()
			if err != nil {
				return nil, err
			}
			dst = "/tmp/" + vt + filepath.Base(f.Name())
			sshClient.Env["NEBULANT_JSON_VARIABLES_PATH"] = dst
		default:
			return nil, fmt.Errorf("unknown var dump type")
		}
		defer os.Remove(f.Name())
		src := f.Name()
		params := scpCopyParameters{
			Target:               p.Target,
			Username:             p.Username,
			Port:                 p.Port,
			PrivateKey:           p.PrivateKey,
			PrivateKeyPath:       p.PrivateKeyPath,
			PrivateKeyPassphrase: p.PrivateKeyPassphrase,
			Password:             p.Password,
			Paths: []scpCopyParametersPath{
				{
					Dst: &dst,
					Src: &src,
				},
			},
		}
		d, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		ac := &ActionContext{
			Action: &blueprint.Action{
				Parameters: d,
			},
			Store:  ctx.Store,
			Logger: ctx.Logger,
		}
		_, err = ScpCopy(ac)
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
