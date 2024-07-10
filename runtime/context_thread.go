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

// actioncontext type

package runtime

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/develatio/nebulant-cli/base"
)

type threadPointContext struct {
	_dbgname  string
	runStatus base.ActionContextRunStatus
	parent    base.IActionContext
	children  []base.IActionContext
	elistener *base.EventListener
	a         *base.Action
}

func (t *threadPointContext) SetStore(s base.IStore) {}

func (t *threadPointContext) GetStore() base.IStore {
	return nil
}

// func (t *threadPointContext) SetProvider(p base.IProvider) {}
// func (t *threadPointContext) GetProvider() base.IProvider  { return nil }

func (t *threadPointContext) Parent(p base.IActionContext) {
	if p != nil {
		t.parent = p
	}
}

func (t *threadPointContext) Parents() []base.IActionContext {
	if t.parent == nil {
		return nil
	}
	return []base.IActionContext{t.parent}
}

func (t *threadPointContext) Child(c base.IActionContext) {
	if c != nil {
		t.children = append(t.children, c)
	}
}

func (t *threadPointContext) Children() []base.IActionContext {
	return t.children
}

func (t *threadPointContext) Type() base.ContextType { return base.ContextTypeThread }
func (t *threadPointContext) IsThreadPoint() bool    { return true }
func (t *threadPointContext) IsJoinPoint() bool      { return false }

func (t *threadPointContext) GetAction() *base.Action {
	if t.a == nil {
		t.a = &base.Action{
			ActionName: "forky",
			ActionID:   fmt.Sprintf("%d", rand.Int()), // #nosec G404 -- Weak random is OK here
		}
	}
	return t.a
}

func (t *threadPointContext) WithRunFunc(f func() (*base.ActionOutput, error)) {}
func (t *threadPointContext) RunAction() (*base.ActionOutput, error) {
	return nil, fmt.Errorf("this kind of actx has no action")
}

func (t *threadPointContext) SetRunStatus(s base.ActionContextRunStatus) {
	t.runStatus = s
}
func (t *threadPointContext) GetRunStatus() base.ActionContextRunStatus {
	return t.runStatus
}

func (t *threadPointContext) Done() <-chan struct{} { return nil }

func (t *threadPointContext) WithEventListener(el *base.EventListener) {
	t.elistener = el
}

func (t *threadPointContext) EventListener() *base.EventListener {
	return t.elistener
}

func (t *threadPointContext) WithCancelCause() {}
func (t *threadPointContext) Cancel(e error)   {}

func (t *threadPointContext) WithDebugInitFunc(f func())      {}
func (t *threadPointContext) DebugInit()                      {}
func (t *threadPointContext) GetSluvaFD() io.ReadWriteCloser  { return nil }
func (t *threadPointContext) GetMustarFD() io.ReadWriteCloser { return nil }
