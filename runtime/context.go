// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

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

package runtime

import (
	"context"
	"io"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nsterm"
)

type actionContext struct {
	_dbgname  string
	ctx       context.Context
	runStatus base.ActionContextRunStatus
	cancel    func(error)
	initfunc  func() (*base.ActionOutput, error)
	dbgfunc   func()
	vpty      *nsterm.VPTY2
	store     base.IStore
	action    *base.Action
	// provider base.IProvider
	parent base.IActionContext
	child  base.IActionContext
	//
	elistener *base.EventListener
}

func (a *actionContext) Done() <-chan struct{} {
	if a.ctx == nil {
		return nil
	}
	return a.ctx.Done()
}

func (a *actionContext) WithEventListener(el *base.EventListener) {
	a.elistener = el
}

func (a *actionContext) EventListener() *base.EventListener {
	return a.elistener
}

func (a *actionContext) WithCancelCause() {
	ctx, cancel := context.WithCancelCause(context.Background())
	a.ctx = ctx
	a.cancel = cancel
}

func (a *actionContext) Cancel(e error) {
	if a.cancel != nil {
		a.cancel(e)
	}
}

// the func that runs on DebugInit call
func (a *actionContext) WithDebugInitFunc(f func()) {
	a.dbgfunc = f
}

// returns the fd in wich the action can
// read/write to interact with  user,
// commonly the sluva FD of vpty
func (a *actionContext) DebugInit() {
	a.dbgfunc()
}

func (a *actionContext) GetSluvaFD() io.ReadWriteCloser {
	if a.vpty == nil {
		vpty := nsterm.NewVirtPTY()
		vpty.SetLDisc(nsterm.NewRawLdisc())
		a.vpty = vpty
	}
	return a.vpty.SluvaFD()
}
func (a *actionContext) GetMustarFD() io.ReadWriteCloser {
	if a.vpty == nil {
		vpty := nsterm.NewVirtPTY()
		vpty.SetLDisc(nsterm.NewRawLdisc())
		a.vpty = vpty
	}
	return a.vpty.MustarFD()
}

func (a *actionContext) SetStore(s base.IStore) {
	a.store = s
}

func (a *actionContext) GetStore() base.IStore {
	return a.store
}

// func (a *actionContext) SetProvider(p base.IProvider) { a.provider = p }
// func (a *actionContext) GetProvider() base.IProvider  { return a.provider }

func (a *actionContext) Type() base.ContextType { return base.ContextTypeRegular }
func (a *actionContext) IsThreadPoint() bool    { return false }
func (a *actionContext) IsJoinPoint() bool      { return false }
func (t *actionContext) Parents() []base.IActionContext {
	if t.parent == nil {
		return nil
	}
	return []base.IActionContext{t.parent}
}

func (a *actionContext) Parent(p base.IActionContext) {
	if p != nil {
		a.parent = p
	}
}

func (a *actionContext) Child(c base.IActionContext) {
	if c != nil {
		a.child = c
	}
}

func (a *actionContext) Children() []base.IActionContext {
	if a.child == nil {
		return nil
	}
	return []base.IActionContext{a.child}
}

func (a *actionContext) WithRunFunc(f func() (*base.ActionOutput, error)) {
	a.initfunc = f
}
func (a *actionContext) RunAction() (*base.ActionOutput, error) {
	return a.initfunc()
}

func (a *actionContext) SetRunStatus(s base.ActionContextRunStatus) {
	a.runStatus = s
}
func (a *actionContext) GetRunStatus() base.ActionContextRunStatus {
	return a.runStatus
}

func (a *actionContext) GetAction() *base.Action {
	return a.action
}
