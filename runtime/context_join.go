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
	"io"

	"github.com/develatio/nebulant-cli/base"
)

// actioncontext type
type joinPointContext struct {
	_dbgname  string
	runStatus base.ActionContextRunStatus
	store     base.IStore
	action    *base.Action
	parents   []base.IActionContext
	child     base.IActionContext
	elistener *base.EventListener
}

func (j *joinPointContext) SetStore(s base.IStore) {
	j.store = s
}

func (j *joinPointContext) GetStore() base.IStore {
	return j.store
}

// func (j *joinPointContext) SetProvider(p base.IProvider) {}
// func (j *joinPointContext) GetProvider() base.IProvider  { return nil }

func (j *joinPointContext) Parent(p base.IActionContext) {
	if p != nil {
		j.parents = append(j.parents, p)
	}
}

func (j *joinPointContext) Parents() []base.IActionContext {
	return j.parents
}

func (j *joinPointContext) Child(c base.IActionContext) {
	if c != nil {
		j.child = c
	}
}

func (j *joinPointContext) Children() []base.IActionContext {
	if j.child == nil {
		return nil
	}
	return []base.IActionContext{j.child}
}

func (j *joinPointContext) Type() base.ContextType { return base.ContextTypeJoin }
func (j *joinPointContext) IsThreadPoint() bool    { return false }
func (j *joinPointContext) IsJoinPoint() bool      { return true }

func (j *joinPointContext) GetAction() *base.Action {
	return j.action
}

func (j *joinPointContext) WithRunFunc(f func() (*base.ActionOutput, error)) {}
func (j *joinPointContext) RunAction() (*base.ActionOutput, error) {
	return nil, nil
	// return nil, fmt.Errorf("this kind of actx has no executable action") // maybe in a future
}

func (j *joinPointContext) SetRunStatus(s base.ActionContextRunStatus) {
	j.runStatus = s
}
func (j *joinPointContext) GetRunStatus() base.ActionContextRunStatus {
	return j.runStatus
}

func (j *joinPointContext) Done() <-chan struct{} { return nil }

func (j *joinPointContext) WithEventListener(el *base.EventListener) {
	j.elistener = el
}

func (j *joinPointContext) EventListener() *base.EventListener {
	return j.elistener
}

func (j *joinPointContext) WithCancelCause() {}
func (j *joinPointContext) Cancel(e error)   {}

func (j *joinPointContext) WithDebugInitFunc(f func()) {}
func (j *joinPointContext) DebugInit()                 {}

func (j *joinPointContext) GetSluvaFD() io.ReadWriteCloser  { return nil }
func (j *joinPointContext) GetMustarFD() io.ReadWriteCloser { return nil }
