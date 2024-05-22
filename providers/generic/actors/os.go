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
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/ipc"
	"github.com/develatio/nebulant-cli/term"
	nebulant_term "github.com/develatio/nebulant-cli/term"
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
	Vars                map[string]string `json:"vars"`
	DumpJSON            *bool             `json:"dump_json"`
	ScriptText          *string           `json:"script"`
	ScriptParameters    *string           `json:"scriptParameters"`
	ScriptName          string            `json:"scriptName"`
	Command             *string           `json:"command"`
	CommandAsSingleArg  bool              `json:"pass_to_entrypoint_as_single_param"`
	Entrypoint          *string           `json:"entrypoint"`
	OpenDbgShellAfter   bool              `json:"open_dbg_shell_after"`
	OpenDbgShellBefore  bool              `json:"open_dbg_shell_before"`
	OpenDbgShellOnerror bool              `json:"open_dbg_shell_onerror"`
}

type readFileParameters struct {
	FilePath    *string `json:"file_path" validate:"required"`
	Interpolate bool    `json:"interpolate"`
}

type writeFileParameters struct {
	FilePath    *string `json:"file_path" validate:"required"`
	Content     *string `json:"content" validate:"required"`
	Interpolate bool    `json:"interpolate"`
}

func newLocalDebugShell(ctx *ActionContext, failed *exec.Cmd) error {
	shell, err := nebulant_term.DetermineOsShell()
	if err != nil {
		ctx.Logger.LogErr(err.Error())
		return err
	}

	ctx.DebugInit()

	mst := ctx.GetMustarFD()
	defer mst.Close()
	svu := ctx.GetSluvaFD() // bring sluva to tty

	ctx.Logger.LogInfo(fmt.Sprintf("original exec: %s", strings.Join(failed.Args, " ")))

	f, err := nebulant_term.GetOSPTY(&nebulant_term.OSPTYConf{Shell: shell, Env: failed.Env})
	if err != nil {
		return err
	}

	// copy sluva to os pty
	go func() { _, _ = io.Copy(f, svu) }()

	// copy os pty to sluva
	_, err = io.Copy(svu, f)
	if err != nil {
		return err
	}
	return nil
}

// RunLocalScript func
func RunLocalScript(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error

	p := &runLocalParameters{}
	if err := json.Unmarshal(ctx.Action.Parameters, p); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(p)
	if err != nil {
		return nil, err
	}

	if p.ScriptText != nil {
		// If dir is the empty string, CreateTemp uses
		// the default directory for temporary files,
		// as returned by TempDir.
		// If pattern includes a "*", the random
		// string replaces the last "*"
		f, err := os.CreateTemp("", "nblscript.*"+p.ScriptName)
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		if _, err := f.Write([]byte(*p.ScriptText)); err != nil {
			if err2 := f.Close(); err2 != nil {
				return nil, errors.Join(err, err2)
			}
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
		// #nosec G302 -- Here +x is needed
		if err := os.Chmod(f.Name(), 0755); err != nil {
			return nil, err
		}
		stin := f.Name()
		if p.ScriptParameters != nil {
			stin = stin + " " + *p.ScriptParameters
		}
		p.Command = &stin
	} else if p.Command != nil {
		if p.ScriptParameters != nil {
			*p.Command = *p.Command + " " + *p.ScriptParameters
		}
	} else {
		return nil, fmt.Errorf("embedded script or command are needed")
	}

	if p.Entrypoint == nil || p.Entrypoint != nil && len(strings.Replace(*p.Entrypoint, " ", "", -1)) <= 0 {
		shell, err := term.DetermineOsShell()
		if err != nil {
			return nil, err
		}

		p.Entrypoint = &shell
		p.CommandAsSingleArg = true
	}

	var cmd *exec.Cmd
	var argv []string

	if p.CommandAsSingleArg {
		argv, err = util.CommandLineToArgv(*p.Entrypoint)
		if err != nil {
			return nil, err
		}
		argv = append(argv, *p.Command)
	} else {
		argv, err = util.CommandLineToArgv(*p.Entrypoint + " " + *p.Command)
		if err != nil {
			return nil, err
		}
	}

	ctx.Logger.LogInfo("Running cmd [" + strings.Join(argv, ", ") + "]")
	cmd = exec.Command(argv[0], argv[1:]...) // #nosec G204 -- allowed here

	envVars := os.Environ()
	for varname := range p.Vars {
		varvalue := p.Vars[varname]
		err := ctx.Store.Interpolate(&varvalue)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, varname+"="+varvalue)
	}

	if p.DumpJSON != nil && *p.DumpJSON {
		f, err := ctx.Store.DumpValuesToJSONFile()
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		envVars = append(envVars, "NEBULANT_JSON_VARIABLES_PATH="+f.Name())
	}

	execpath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, "NEBULANT_CLI_PATH="+execpath)

	// Conf IPCS Consumer
	ipcs := ctx.Store.GetPrivateVar("IPCS").(*ipc.IPC)
	envVars = append(envVars, "NEBULANT_IPCSID="+ipcs.GetUUID())
	// out := make(chan bool)
	ipccid := fmt.Sprintf("%d", rand.Int()) // #nosec G404 -- Weak random is OK here
	ipcc := &ipc.IPCConsumer{
		ID:     ipccid,
		Stream: make(chan *ipc.PipeData),
	}
	ipcs.AppendConsumer(ipcc)
	envVars = append(envVars, "NEBULANT_IPCCID="+ipcc.ID)
	out := ipcc.ExposeStoreVars(ctx.Store)
	defer func() {
		out <- true
		ipcs.OutConsumer(ipcc)
	}()

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

	if err != nil && p.OpenDbgShellOnerror {
		ctx.Logger.LogErr(errors.Join(fmt.Errorf("exec fail"), err.(error)).Error())
		ctx.Logger.LogInfo("waiting for debug session to finish")
		dbgErr := newLocalDebugShell(ctx, cmd)
		if dbgErr != nil {
			err = errors.Join(dbgErr, err)
		}
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, err
}

func DefineEnvs(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(defineEnvsParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
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

func ReadFile(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(readFileParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err := ctx.Store.Interpolate(params.FilePath)
	if err != nil {
		return nil, err
	}

	finfo, err := os.Stat(*params.FilePath)
	if err != nil {
		return nil, err
	}

	if finfo.Size() > 100*1024*1024 {
		ctx.Logger.LogWarn("reading a file larger than 100MB, be carefully")
	}

	bcontent, err := os.ReadFile(*params.FilePath)
	if err != nil {
		return nil, err
	}
	scontent := string(bcontent)

	if params.Interpolate {
		err := ctx.Store.Interpolate(&scontent)
		if err != nil {
			return nil, err
		}
	}

	aout := base.NewActionOutput(ctx.Action, scontent, nil)
	aout.Records[0].Literal = true
	return aout, nil
}

func WriteFile(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(writeFileParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	var scontent = *params.Content
	if params.Interpolate {
		err := ctx.Store.Interpolate(&scontent)
		if err != nil {
			return nil, err
		}
	}

	err := ctx.Store.Interpolate(params.FilePath)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(*params.FilePath, []byte(scontent), 0600)
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}
