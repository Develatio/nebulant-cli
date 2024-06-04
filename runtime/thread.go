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
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
)

type ThreadStep int

const (
	ThreadBeforeAction ThreadStep = iota
	ThreadIntoAction
	ThreadAfterAction
	ThreadClose
)

type threadStackCtrl struct {
	confirm chan struct{}
	back    bool
	// skipRun bool
	// reset   bool
}

type Thread struct {
	ThreadStep ThreadStep
	state      base.RuntimeState
	step       chan *threadStackCtrl
	queue      []base.IActionContext
	done       []base.IActionContext
	current    base.IActionContext
	ExitCode   int
	ExitErr    error // uncaught err
	runtime    *Runtime
	elistener  *base.EventListener
}

func (t *Thread) GetQueue() []base.IActionContext {
	return t.queue
}

func (t *Thread) GetDone() []base.IActionContext {
	return t.done
}

func (t *Thread) GetState() base.RuntimeState {
	return t.state
}

func (t *Thread) GetCurrent() base.IActionContext {
	return t.current
}

func (t *Thread) EventListener() *base.EventListener {
	return t.elistener
}

func (t *Thread) Pause() {
	t.step = make(chan *threadStackCtrl)
	t.state = base.RuntimeStateStill
}

func (t *Thread) Play() {
	if t.state == base.RuntimeStatePlay {
		return
	}
	t.state = base.RuntimeStatePlay
	close(t.step)
}

func (t *Thread) Stop() {
	if t.state == base.RuntimeStateStill {
		close(t.step)
	}
	t.state = base.RuntimeStateEnding
}

func (t *Thread) StackUp() (<-chan struct{}, bool) {
	if t.current != nil {
		if t.current.GetRunStatus() == base.RunStatusRunning {
			return nil, false
		}
	}

	stepctrl := &threadStackCtrl{
		confirm: make(chan struct{}),
		back:    true,
	}

	// fill chan and return true on ok
	// this will not wait for chan pop
	select {
	case t.step <- stepctrl:
		return stepctrl.confirm, true
	default:
		// Hey developer!,  what a wonderful day!
		return nil, false
	}
}

func (t *Thread) Step() (<-chan struct{}, bool) {
	if t.current != nil {
		if t.current.GetRunStatus() == base.RunStatusRunning {
			return nil, false
		}
	}

	stepctrl := &threadStackCtrl{
		confirm: make(chan struct{}),
		back:    false,
	}

	// fill chan and return true on ok
	// this will not wait for chan pop
	select {
	case t.step <- stepctrl:
		return stepctrl.confirm, true
	default:
		// Hey developer!,  what a wonderful day!
		return nil, false
	}
}

// return true if this thread is handling or has handled actx
func (t *Thread) hasActionContext(actx base.IActionContext) bool {
	if t.current != nil && t.current == actx {
		return true
	}
	for _, qactx := range t.queue {
		if qactx == actx {
			return true
		}
	}
	for _, dactx := range t.done {
		if dactx == actx {
			return true
		}
	}
	return false
}

func (t *Thread) _loadPrev() bool {
	t.ThreadStep = ThreadBeforeAction
	if t.current == nil {
		// cannot prev in unkown action
		return false
	}

	parents := t.current.Parents()
	if len(parents) <= 0 {
		// no more parents
		// cannot prev without parents
		return false
	}

	if len(parents) > 1 {
		// more than one parent
		// cannot (yet) prev with many parents
		// we need more info: wich parent?
		return false
	}

	// empty the queue!, we are in a new bifurcation mc fly!
	t.queue = []base.IActionContext{}
	t.current = parents[0]
	t.current.SetRunStatus(base.RunStatusArranging)
	return true
}

func (t *Thread) _loadNext() bool {
	t.ThreadStep = ThreadBeforeAction
	if len(t.queue) <= 0 {
		t.current = nil
		return false
	}

	t.current = t.queue[0]
	t.current.SetRunStatus(base.RunStatusArranging)
	t.queue = t.queue[1:]
	return true
}

// func (t *Thread) _pdbg() {
// 	prnt := fmt.Sprintf("\n****%p*****\n", t)
// 	for _, actx := range t.done {
// 		prnt = prnt + fmt.Sprintf("[%s]->", actx.GetAction().ActionName)
// 	}
// 	if t.current != nil {
// 		prnt = prnt + fmt.Sprintf("[%s]<-", t.current.GetAction().ActionName)
// 	}
// 	if t.state == base.RuntimeStateEnd {
// 		prnt = prnt + "END"
// 	}
// 	prnt = prnt + "\n*********\n"
// 	fmt.Print(prnt)
// }

func (t *Thread) _runCurrent() {
	// t._pdbg()
	// defer t._pdbg()
	if t.current == nil {
		panic("hey dev, this is your fault")
	}
	actx := t.current
	action := actx.GetAction()
	//	store := actx.GetStore()
	// store.GetLogger().SetThreadID(fmt.Sprintf("%p", t))
	// store.GetLogger().SetActionID(action.ActionID)
	var nexts []*blueprint.Action

	if actx.IsThreadPoint() {
		t.runtime.switchContext(actx)
		for _, fkactx := range actx.Children() {
			t.runtime.NewThread(fkactx)
		}
		// t.runtime.deactivateContext(actx)
		t.done = append(t.done, actx)
		// stop exec, no more actx stored in queue
		// so this should be the last actx
		return
	}

	t.runtime.switchContext(actx) // deactivate parent, activate self (actx)

	t.ThreadStep = ThreadIntoAction
	defer func() {
		t.ThreadStep = ThreadAfterAction
	}()

	if action.JoinThreadsPoint {
		tt := t.runtime.cjoiner.Join(actx)
		if tt == nil {
			// destroy this thread, there is already
			// a thread handling the join point
			t.runtime._deactivateContext(actx)
			return
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if t.state == base.RuntimeStateEnding || t.state == base.RuntimeStateEnd {
				break
			}
			if t.runtime.hasRunningParents(actx) {
				continue
			}
			break
		}
	}

	aout, aerr := actx.RunAction()

	// recopilate nexts
	if aerr != nil {
		provider, err := actx.GetStore().GetProvider(action.Provider)
		if err != nil {
			// hey dev, this is your fault. Only an internal action has no
			// provider and you should handle these actions before this
			log.Panic(errors.Join(err, fmt.Errorf("cannot obtain provider %s", action.Provider)))
		}
		nexts, err = provider.OnActionErrorHook(aout)
		// update action err on provider err hook err (nil will be ignored)

		aerr = errors.Join(fmt.Errorf("%s %s KO", action.ActionID, action.ActionName), aerr, err)
		if nexts == nil {
			nexts = action.NextAction.NextKo
		}
		// will be reset if KO port is handled
		t.ExitCode = 1
		t.ExitErr = aerr
	} else if action.NextAction.ConditionalNext {
		if aout.Records[0].Value.(bool) {
			nexts = action.NextAction.NextOkTrue
		} else {
			nexts = action.NextAction.NextOkFalse
		}
	} else {
		nexts = action.NextAction.NextOk
	}

	switch len(nexts) {
	case 0:
		// no more actions, errs and exit code
		// keep as setted before
		t.done = append(t.done, actx)
		if t.ExitCode > 0 {
			cast.LogErr(t.ExitErr.Error(), t.runtime.irb.ExecutionUUID)
		}
		return
	case 1:
		// reset exit code and err if exists
		if t.ExitCode > 0 && config.DEBUG {
			cast.LogWarn(t.ExitErr.Error(), t.runtime.irb.ExecutionUUID)
		}
		t.ExitCode = 0
		t.ExitErr = nil
		// NewAContext will create the JoinThreadContext,
		// right before the run of that context
		nactx := t.runtime.NewAContext(actx, nexts[0])
		t.done = append(t.done, actx)
		// put new actx at first to prevent fails on t.Back()
		t.queue = append([]base.IActionContext{nactx}, t.queue...)
	default:
		// reset exit code and err if exists
		t.ExitCode = 0
		t.ExitErr = nil
		// more than one, new threads needed
		t.done = append(t.done, actx)
		threadactx := t.runtime.NewAContextThread(actx, nexts)
		t.queue = append([]base.IActionContext{threadactx}, t.queue...)
	}
}

// closes the thread, his llops and his event listeners
func (t *Thread) close() {
	t.ThreadStep = ThreadClose
	t.state = base.RuntimeStateEnding

	if t.current != nil {
		// deactivate current action id
		t.runtime._deactivateContext(t.current)
		t.current.Cancel(nil)
	} else if len(t.done) > 0 {
		// deactivate the last action id
		t.runtime._deactivateContext(t.done[len(t.done)-1])
	}

	// remove thread t
	t.runtime.finishThread(t)
	cast.LogDebug("Thread finished", t.runtime.irb.ExecutionUUID)
	t.state = base.RuntimeStateEnd
	// TODO: close t.elistener?
}

func (t *Thread) GetLastRun() base.IActionContext {
	if len(t.done) <= 0 {
		return nil
	}
	return t.done[len(t.done)-1]
}

// commonly called by go Init()
func (t *Thread) Init() {
	var waitcount int
	defer func() {
		t.close()
		// t._pdbg()
	}()

	// define uninitialized (nil) "confirm" chan
	// var confirm chan struct{} // nil chan
	var stpctrl *threadStackCtrl = &threadStackCtrl{back: false}
	var more bool

	for {
		// este control del load me gusta, lo recupero y lo
		// dejo por aquí
		if cast.BInfo.GetLoad() > 10.0 {
			waitcount++
			// reduce run speed on high bus load
			var fa time.Duration = time.Duration(cast.BInfo.GetLoad() * 10.0)
			// for debug:
			// fmt.Println("Bus load up 10:", cast.BusLoad, "Sleeping", fa)
			time.Sleep(fa * time.Millisecond)
			if waitcount > 100 {
				fmt.Println("High bus load detected. Waiting load reduction...")
				waitcount = 0
			}
			continue
		}
		waitcount = 0

	preload:
		// load action after all
		if stpctrl != nil && stpctrl.back {
			// in back mode, here we load previous actx
			t._loadPrev()
			more = t.current != nil
		} else {
			more = t._loadNext()
		}

	prerun:
		// uninitialized channel confirm is nil
		// a return from closed t.step is also nil
		// this confirms the previous after-run t.step
		// if exists
		if stpctrl != nil && stpctrl.confirm != nil {
			stpctrl.confirm <- struct{}{}
		}

		// run  action. Note that there is no steps
		// awaiting. The t.Step() cannot be called if
		// there is one action running
		if !more {
			return
		}

		if t.state == base.RuntimeStateEnding {
			return
		}
		// uninitialized step is nil
		// closed step is not nil
		// this is the step and confirm before-run
		// allow step-ing only if the thread state
		// is in pause
		if t.state == base.RuntimeStateStill && t.step != nil {
			// getting from closed step, returns nil
			stpctrl = <-t.step
			if stpctrl != nil && stpctrl.back {
				// backing from here means that we should
				// leave all ready to re-run the same actx
				// it is possible that current is nil
				if t.current != nil {
					t.current.SetRunStatus(base.RunStatusReady)
				}
				goto preload
			}
			// set the run status before step confirm
			// to protect thread from t.Step() calls
			// in this state
			if t.current != nil {
				t.current.SetRunStatus(base.RunStatusRunning)
			}
			// on close step, confirm is nil
			if stpctrl != nil && stpctrl.confirm != nil {
				stpctrl.confirm <- struct{}{}
			}
		}

		t._runCurrent()
		t.current.SetRunStatus(base.RunStatusDone)

		if t.state == base.RuntimeStateEnding {
			return
		}
		// closed step is not nil
		// allow step-ing only if the thread state
		// is in pause
		if t.state == base.RuntimeStateStill && t.step != nil {
			// closed step returns nil on get
			// awaiting chan also returns nil on close
			stpctrl = <-t.step
			if stpctrl != nil && stpctrl.back {
				// backing from here means we should goto
				// to afterload actx and waiting step to
				// re-run same actx
				// set the status ready to leave all ready
				// on goto prerun to re-run actx
				t.current.SetRunStatus(base.RunStatusArranging)
				goto prerun
			}
		}
		// end thread on no more actions
	}
}
