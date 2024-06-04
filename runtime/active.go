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
	"sync"

	"github.com/develatio/nebulant-cli/base"
)

type activeAction struct {
	ctxs  []base.IActionContext
	count int
}

type activeActionsID struct {
	mu sync.Mutex
	a  map[string]*activeAction
}

// func (a *activeActionsID) _pdbg() {
// 	prnt := "\n*********\n"
// 	for _, active := range a.a {
// 		prnt = prnt + fmt.Sprintf("[%s]->%v\n", active.ctxs[0].GetAction().ActionName, active.count)
// 	}
// 	prnt = prnt + "\n*********\n"
// 	fmt.Print(prnt)
// }

func (a *activeActionsID) More(actx base.IActionContext) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := actx.GetAction().ActionID
	if _, exists := a.a[id]; !exists {
		a.a[id] = &activeAction{
			count: 0,
			ctxs:  []base.IActionContext{actx},
		}
	}
	a.a[id].count++
}

func (a *activeActionsID) Less(actx base.IActionContext) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := actx.GetAction().ActionID
	if _, exists := a.a[id]; !exists {
		return
	}
	a.a[id].count--
}

func (a *activeActionsID) Exists(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	activ, exists := a.a[id]
	return exists && activ.count > 0
}

func (a *activeActionsID) ExistsAny(ids map[string]bool) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for id := range ids {
		if activ, exists := a.a[id]; exists {
			if activ.count > 0 {
				return true
			}
			continue
		}
	}
	return false
}

func (a *activeActionsID) Slice() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var cc []string
	for actnID, activ := range a.a {
		if activ.count > 0 {
			cc = append(cc, actnID)
		}
	}
	return cc
}
