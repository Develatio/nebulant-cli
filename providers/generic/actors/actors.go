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

import (
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
}

// ActionFuncMap map
var ActionFuncMap map[string]*ActionLayout = map[string]*ActionLayout{
	"run_script":       {F: RunScript, N: NextOKKO},
	"define_envs":      {F: DefineEnvs, N: NextOKKO},
	"define_variables": {F: DefineVars, N: NextOKKO},
	"upload_files":     {F: RemoteCopy, N: NextOKKO},
	"download_files":   {F: RemoteCopy, N: NextOKKO},
	"condition":        {F: ConditionParse, N: NextOKKO},
	"start":            {F: Start, N: NextOKKO},
	"stop":             {F: Stop, N: NextOKKO},
	"sleep":            {F: Sleep, N: NextOKKO},
	"ok/ko":            {F: OKKO, N: NextOKKO},
	"log":              {F: Log, N: NextOKKO},
	"noop":             {F: NOOP, N: NextOK},
	"panic":            {F: Panic, N: NextOKKO},
	"send_mail":        {F: SendMail, N: NextOKKO},
	// handled by core stage
	"join_threads": {F: NOOP, N: NextOK},
}
