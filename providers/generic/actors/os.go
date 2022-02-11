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

package actors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/joho/godotenv"
)

type defineEnvsParameters struct {
	Vars  map[string]string `json:"vars" validate:"required"`
	Files []string          `json:"files"`
}

type runLocalScriptOutput struct {
	RawStdout *bytes.Buffer `json:"-"`
	RawStderr *bytes.Buffer `json:"-"`
	Stdout    string        `json:"stdout"`
	Stderr    string        `json:"stderr"`
	ExitCode  string        `json:"exit_code"`
}

type runLocalParameters struct {
	Target *string `json:"target" validate:"required"`
	// Username       *string `json:"username", validate:"required"`
	// PrivateKeyPath *string `json:"keyfile"`
	// Password       *string `json:"password"`
	// Port           *string `json:"port"`
	Vars             map[string]string `json:"vars"`
	VarsTargets      []string          `json:"vars_targets"`
	ScriptText       *string           `json:"script"`
	ScriptParameters *string           `json:"scriptParameters"`
	Command          *string           `json:"command"`
	Entrypoint       *string           `json:"entrypoint"`
}

// RunLocalScript func
func RunLocalScript(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error

	p := &runLocalParameters{}
	jsonErr := json.Unmarshal(ctx.Action.Parameters, p)
	if jsonErr != nil {
		return nil, jsonErr
	}

	var cmd *exec.Cmd
	if p.Command != nil { // run cmd
		stin := *p.Command
		if p.Entrypoint != nil {
			stin = *p.Entrypoint + " " + stin
		}
		argv, err := util.CommandLineToArgv(stin)
		if err != nil {
			return nil, err
		}
		if len(argv) > 1 {
			cmd = exec.Command(argv[0], argv[1:]...) //#nosec G204 -- Tainted arguments here are the responsibility of the user
		} else {
			cmd = exec.Command(argv[0]) //#nosec G204 -- Tainted arguments here are the responsibility of the user
		}
	} else if p.ScriptText != nil {
		// If dir is the empty string, CreateTemp uses
		// the default directory for temporary files,
		// as returned by TempDir.
		// If pattern includes a "*", the random
		// string replaces the last "*"
		f, err := os.CreateTemp("", "nblscript.*")
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		if _, err := f.Write([]byte(*p.ScriptText)); err != nil {
			if err2 := f.Close(); err2 != nil {
				return nil, fmt.Errorf(err.Error() + " " + err2.Error())
			}
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
		//#nosec G302 -- Here +x is needed
		if err := os.Chmod(f.Name(), 0755); err != nil {
			return nil, err
		}
		stin := f.Name()
		if p.Entrypoint != nil {
			stin = *p.Entrypoint + " " + stin
		}
		if p.ScriptParameters != nil {
			stin = stin + *p.ScriptParameters
		}
		argv, err := util.CommandLineToArgv(stin)
		if err != nil {
			return nil, err
		}
		if len(argv) > 1 {
			cmd = exec.Command(argv[0], argv[1:]...) //#nosec G204 -- Tainted arguments here are the responsibility of the user
		} else {
			cmd = exec.Command(argv[0]) //#nosec G204 -- Tainted arguments here are the responsibility of the user
		}
	} else {
		return nil, fmt.Errorf("no command nor embedded script provided")
	}

	envVars := os.Environ()
	for varname := range p.Vars {
		varvalue := p.Vars[varname]
		err := ctx.Store.Interpolate(&varvalue)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, varname+"="+varvalue)
	}

	for _, vt := range p.VarsTargets {
		switch vt {
		case "bash":
			f, err := ctx.Store.DumpValuesToShellFile()
			if err != nil {
				return nil, err
			}
			defer os.Remove(f.Name())
			envVars = append(envVars, "NEBULANT_BASH_VARIABLES_PATH="+f.Name())
		case "zsh":
			f, err := ctx.Store.DumpValuesToShellFile()
			if err != nil {
				return nil, err
			}
			defer os.Remove(f.Name())
			envVars = append(envVars, "NEBULANT_ZSH_VARIABLES_PATH="+f.Name())
		case "json":
			f, err := ctx.Store.DumpValuesToJSONFile()
			if err != nil {
				return nil, err
			}
			defer os.Remove(f.Name())
			envVars = append(envVars, "NEBULANT_JSON_VARIABLES_PATH="+f.Name())
		}
	}

	cmd.Env = envVars
	result := &runLocalScriptOutput{}
	var cmdOut io.Writer
	var cmdErr io.Writer
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "local"
	}
	result.RawStdout = new(bytes.Buffer)
	result.RawStderr = new(bytes.Buffer)
	cmdOut = io.MultiWriter(result.RawStdout, &logWriter{
		Log:       ctx.Logger.ByteLogInfo,
		LogPrefix: []byte(hostname + "> "),
	})
	cmdErr = io.MultiWriter(result.RawStderr, &logWriter{
		Log:       ctx.Logger.ByteLogErr,
		LogPrefix: []byte(hostname + "> "),
	})
	cmd.Stdout = cmdOut
	cmd.Stderr = cmdErr
	cmdRunError := cmd.Run()
	result.Stdout = result.RawStdout.String()
	result.Stderr = result.RawStderr.String()

	if cmdRunError == nil {
		result.ExitCode = "0"
	} else if _, isExitError := cmdRunError.(*exec.ExitError); isExitError {
		// It is cast to a string because the conditional evaluation of
		// this result is defined from the graphical app and in it the
		// value to compare will always be a string.
		result.ExitCode = strconv.Itoa(cmdRunError.(*exec.ExitError).ExitCode())
		err = fmt.Errorf("Exit status != 0 (" + result.ExitCode + ")")
	} else if _, isNetError := cmdRunError.(*net.OpError); isNetError {
		err = cmdRunError.(*net.OpError).Err
	} else {
		err = cmdRunError
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, err
}

func DefineEnvs(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(defineEnvsParameters)
	jsonErr := util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if jsonErr != nil {
		return nil, jsonErr
	}
	for varname := range params.Vars {
		varvalue := params.Vars[varname]
		ctx.Logger.LogInfo("Setting env var " + varname)
		err := ctx.Store.Interpolate(&varvalue)
		if err != nil {
			return nil, err
		}
		err = os.Setenv(varname, varvalue)
		if err != nil {
			return nil, err
		}
	}
	for _, file := range params.Files {
		err := godotenv.Load(file)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
