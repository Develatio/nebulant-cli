// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

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

package actors

import (
	"io"
	"sync"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
)

type LogFunc func(b []byte)

type logWriter struct {
	mu        sync.Mutex
	Log       LogFunc
	LogPrefix []byte
	v         bool
}

func (l *logWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := make([]byte, len(p))
	copy(b, p)
	if !l.v {
		l.Log(append(l.LogPrefix, []byte("\n")...))
		l.v = true
	}
	l.Log(b)
	return len(p), nil
}

// ActionContext struct
type ActionContext struct {
	Rehearsal bool
	Action    *blueprint.Action
	Store     base.IStore
	Logger    base.ILogger
	Actx      base.IActionContext
}

func (a *ActionContext) DebugInit() {
	a.Actx.DebugInit()
}

func (a *ActionContext) GetSluvaFD() io.ReadWriteCloser {
	return a.Actx.GetSluvaFD()
}

func (a *ActionContext) GetMustarFD() io.ReadWriteCloser {
	return a.Actx.GetMustarFD()
}

// ActionFunc func
type ActionFunc func(ctx *ActionContext) (*base.ActionOutput, error)

// NextType int
type NextType int

const (
	// NextOKKO const 0
	NextOKKO NextType = iota
	// NextOK const 1
	NextOK
	// NextKO const
	NextKO
)

// ActionLayout struct
type ActionLayout struct {
	F ActionFunc
	N NextType
	// Retriable or not
	R bool
}

// ActionFuncMap map
var ActionFuncMap map[string]*ActionLayout = map[string]*ActionLayout{
	"run_script":       {F: RunScript, N: NextOKKO, R: true},
	"define_envs":      {F: DefineEnvs, N: NextOKKO, R: false},
	"define_variables": {F: DefineVars, N: NextOKKO, R: false},
	"upload_files":     {F: RemoteCopy, N: NextOKKO, R: true},
	"download_files":   {F: RemoteCopy, N: NextOKKO, R: true},
	"condition":        {F: ConditionParse, N: NextOKKO, R: false},
	"start":            {F: DefineVars, N: NextOKKO, R: false},
	"group":            {F: NOOP, N: NextOKKO, R: false},
	"stop":             {F: Stop, N: NextOKKO, R: false},
	"end":              {F: NOOP, N: NextOKKO, R: false},
	"sleep":            {F: Sleep, N: NextOKKO, R: false},
	"ok/ko":            {F: OKKO, N: NextOKKO, R: false},
	"log":              {F: Log, N: NextOKKO, R: false},
	"noop":             {F: NOOP, N: NextOK, R: false},
	"panic":            {F: Panic, N: NextOKKO, R: false},
	"send_mail":        {F: SendMail, N: NextOKKO, R: true},
	"send_email":       {F: SendMail, N: NextOKKO, R: true},
	"http_request":     {F: HttpRequest, N: NextOKKO, R: true},
	"read_file":        {F: ReadFile, N: NextOKKO, R: false},
	"write_file":       {F: WriteFile, N: NextOKKO, R: false},
	// handled by core stage
	"join_threads": {F: NOOP, N: NextOK, R: false},
	"debug":        {F: NOOP, N: NextOK, R: false},
}
