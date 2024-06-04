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

type contextJoinerPoint struct {
	t chan struct{}
	p base.IActionContext
}

type contextJoiner struct {
	mu sync.Mutex
	pt map[string]*contextJoinerPoint
}

func (j *contextJoiner) Lock() {
	j.mu.Lock()
}

func (j *contextJoiner) Unlock() {
	j.mu.Unlock()
}

func (j *contextJoiner) Join(actx base.IActionContext) chan struct{} {
	j.mu.Lock()
	defer j.mu.Unlock()
	action := actx.GetAction()
	if jpoint, exists := j.pt[action.ActionID]; exists {
		// jpctx := jpoint.p
		// If all goes ok, actx should has one or zero parents
		// because the merge should be done here
		// TODO: unit test for this behavior
		prs := actx.Parents()
		if len(prs) > 1 {
			panic("hey dev, this is your fault :*")
		}
		jpoint.p.GetStore().Merge(actx.GetStore())
		return nil
	}
	ticker := make(chan struct{})
	j.pt[action.ActionID] = &contextJoinerPoint{
		t: ticker,
		p: actx,
	}
	return ticker
}

func (j *contextJoiner) Destroy(actx base.IActionContext) {
	j.mu.Lock()
	defer j.mu.Unlock()
	action := actx.GetAction()
	if _, exists := j.pt[action.ActionID]; exists {
		close(j.pt[action.ActionID].t)
		delete(j.pt, action.ActionID)
	}
}
