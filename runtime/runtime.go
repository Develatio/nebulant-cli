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
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
)

func NewRuntime(irb *blueprint.IRBlueprint, serverMode bool) *Runtime {
	return &Runtime{
		irb:                irb,
		serverMode:         serverMode,
		actionContextStack: make([]base.IActionContext, 0, 1),
		activeActionsID: &activeActionsID{
			a: make(map[string]*activeAction),
		},
		cjoiner: &contextJoiner{
			pt: map[string]*contextJoinerPoint{},
		},
		activeThreads: make(map[*Thread]bool),
		evDispatcher:  base.NewEventDispatcher(),
		exitCode:      0,
	}
}

type Runtime struct {
	mu sync.Mutex
	// ru            sync.Mutex
	serverMode    bool
	state         base.RuntimeState
	activeThreads map[*Thread]bool
	// eventListeners     map[*base.EventListener]bool
	evDispatcher       *base.EventDispatcher
	irb                *blueprint.IRBlueprint
	actionContextStack []base.IActionContext
	activeActionsID    *activeActionsID
	// join points
	cjoiner  *contextJoiner
	exitCode int
	exitErrs []error // uncaught err
	//
	savedActionOutputs []*base.ActionOutput
}

func (r *Runtime) hasRunningParents(actx base.IActionContext) bool {
	r.cjoiner.Lock()
	defer r.cjoiner.Unlock()
	return r.activeActionsID.ExistsAny(actx.GetAction().KnowParentIDs)
}

func (r *Runtime) GetStack() []base.IActionContext {
	return r.actionContextStack
}

func (r *Runtime) IsServerMode() bool {
	return r.serverMode
}

func (r *Runtime) saveActionOutput(aout *base.ActionOutput) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.savedActionOutputs = append(r.savedActionOutputs, aout)
}

func (r *Runtime) SavedActionOutputs() []*base.ActionOutput {
	return r.savedActionOutputs
}

func (r *Runtime) ExitCode() int {
	return r.exitCode
}

func (r *Runtime) Error() error {
	return errors.Join(r.exitErrs...)
}

func (r *Runtime) Errors() []error {
	return r.exitErrs
}

func (r *Runtime) NewEventListener() *base.EventListener {
	return r.evDispatcher.NewEventListener()
}

func (r *Runtime) NewThread(actx base.IActionContext) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state == base.RuntimeStateEnding || r.state == base.RuntimeStateEnd {
		cast.LogDebug(fmt.Sprintf("state ending, prevent start for action %s", actx.GetAction().ActionName), nil)
		return false
	}
	cast.LogDebug("Stat new thread", r.irb.ExecutionUUID)
	// r.mu.Lock()
	// defer r.mu.Unlock()
	el := r.evDispatcher.NewEventListener()
	th := &Thread{
		runtime:   r,
		elistener: el,
		step:      make(chan *threadStackCtrl),
	}
	thid := fmt.Sprintf("%p", th)
	cast.PushBusData(&cast.BusData{
		TypeID:        cast.BusDataTypeEvent,
		EventID:       cast.EP(cast.EventNewThread),
		ThreadID:      &thid,
		ExecutionUUID: r.irb.ExecutionUUID,
		Timestamp:     time.Now().UTC().UnixMicro(),
	})
	actx.GetStore().GetLogger().SetThreadID(thid)
	th.queue = append(th.queue, actx)
	r.activeThreads[th] = true

	// start thread in play or pause mode
	if r.state == base.RuntimeStatePlay {
		th.Play()
	} else if r.state == base.RuntimeStateStill {
		cast.LogDebug(fmt.Sprintf("net thread started in pause mode for action %s", actx.GetAction().ActionName), nil)
		th.Pause()
	}

	cast.LogDebug("Dispatching event", r.irb.ExecutionUUID)
	go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.RuntimePlayEvent})
	cast.LogDebug("Dispatched event", r.irb.ExecutionUUID)
	go th.Init()
	// o quizás el propio runtime debería observar
	// la cola de threads para inicializarlos
	// y así poder tener control y dejar los
	// threads quietos en caso de pause?
	return true
}

func (r *Runtime) finishThread(th *Thread) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.activeThreads, th)
	r.exitCode = r.exitCode + th.ExitCode
	if th.ExitErr != nil {
		r.exitErrs = append(r.exitErrs, th.ExitErr)
	}

	// no threads, no activity
	if len(r.activeThreads) <= 0 {
		go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.RuntimeEndEvent})
	}

	el := th.EventListener()
	r.evDispatcher.DestroyEventListener(el)
}

func (r *Runtime) Play() {
	// TODO: merge both event systems?
	cast.PushEvent(cast.EventRuntimeResuming, r.irb.ExecutionUUID)
	cast.PushEvent(cast.EventRuntimeStarting, r.irb.ExecutionUUID)
	go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.RuntimePlayEvent})
	threads := r.GetThreads()
	for th := range threads {
		th.Play()
	}
	r.state = base.RuntimeStatePlay
	r.DispatchCurrentActiveIdsEvent()
	cast.PushEvent(cast.EventRuntimeStarted, r.irb.ExecutionUUID)
}

func (r *Runtime) Pause() {
	cast.PushEvent(cast.EventRuntimePausing, r.irb.ExecutionUUID)
	go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.RuntimeStillEvent})
	threads := r.GetThreads()
	for th := range threads {
		th.Pause()
	}
	r.state = base.RuntimeStateStill
	r.DispatchCurrentActiveIdsEvent()
	cast.PushEvent(cast.EventRuntimePaused, r.irb.ExecutionUUID)
}

func (r *Runtime) Stop() {
	cast.PushEvent(cast.EventRuntimeStopping, r.irb.ExecutionUUID)
	r.state = base.RuntimeStateEnding
	go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.RuntimeEndEvent})
	threads := r.GetThreads()
	for th := range threads {
		th.Stop()
	}
	r.state = base.RuntimeStateEnd
	r.DispatchCurrentActiveIdsEvent()
	cast.PushEvent(cast.EventRuntimeOut, r.irb.ExecutionUUID)
}

func (r *Runtime) GetThreads() map[*Thread]bool {
	return r.activeThreads
}

func (r *Runtime) _setRunDebugFunc(actx base.IActionContext) {
	actx.WithRunFunc(func() (*base.ActionOutput, error) {
		// Pause exec
		r.Pause()

		// serve debuger
		dbg := NewDebugger(r)
		go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.DebugOnEvent})
		dbg.SetCursor(actx)

		if dbg.running {
			cast.LogInfo("Already running debugger, discarding re-run", r.irb.ExecutionUUID)
			return nil, nil
		}

		// this Start func calls this.IsServerMode
		// to start in local shell or in server mode
		go dbg.Start()
		go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.DebugOffEvent})
		return nil, nil
	})
}

func (r *Runtime) setRunFunc(actx base.IActionContext) {
	action := actx.GetAction()
	if action.DebugPoint {
		r._setRunDebugFunc(actx)
		return
	}
	actx.WithRunFunc(func() (*base.ActionOutput, error) {
		action := actx.GetAction()
		store := actx.GetStore()
		var provider base.IProvider
		var err error
		if !store.ExistsProvider(action.Provider) {
			providerInitFunc, err := cast.SBus.GetProviderInitFunc(action.Provider)
			if err != nil {
				return nil, err
			}
			provider, err = providerInitFunc(store)
			if err != nil {
				return nil, err
			}
			store.StoreProvider(action.Provider, provider)
		} else {
			provider, err = store.GetProvider(action.Provider)
			if err != nil {
				return nil, err
			}
		}

		actx.WithCancelCause()
		defer actx.Cancel(nil)
		// l := store.GetLogger()
		// l.LogInfo(fmt.Sprintf("Running %s", action.ActionName))
		aout, aerr := provider.HandleAction(actx)

		if aerr != nil {
			// ssh run could return non nil aout with
			// result.exitcode > 0 and also aerr non
			// nil with the subyacent err
			if aout == nil {
				aout = base.NewActionOutput(action, aerr.Error(), nil)
			}
			aout.Records[0].Fail = true
			aout.Records[0].Error = aerr
		}

		// aout is nil on action return nil, nil
		if aout != nil {
			for idx := 0; idx < len(aout.Records); idx++ {
				err := store.Insert(aout.Records[idx], action.Provider)
				if err != nil {
					log.Panic(err.Error())
				}
			}
		}

		if action.SaveRawResults {
			r.saveActionOutput(aout)
		}

		return aout, aerr
	})
}

func (r *Runtime) setDebugInitFunc(actx base.IActionContext) {
	actx.WithDebugInitFunc(func() {
		// Pause exec
		r.Pause()

		// serve debuger
		dbg := NewDebugger(r)
		go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.DebugOnEvent})
		dbg.SetCursor(actx)

		// init detached vpty iface
		// vpty := nsterm.NewVirtPTY()
		// vpty.SetLDisc(nsterm.NewRawLdisc())

		must := actx.GetMustarFD()
		dbg.Detach(must)

		// init debugger as needed
		if dbg.running {
			cast.LogInfo("Already running debugger, discarding re-run", r.irb.ExecutionUUID)
		}

		// this Start func calls this.IsServerMode
		// to start in local shell or in server mode
		go dbg.Start()
		go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.DebugOffEvent})
	})
}

func (r *Runtime) pushActionContext(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := len(r.actionContextStack)
	if cap(r.actionContextStack) == idx {
		// grow stack
		newStack := make([]base.IActionContext, idx, 20+cap(r.actionContextStack))
		copy(newStack, r.actionContextStack)
		r.actionContextStack = newStack
	}
	r.actionContextStack = append(r.actionContextStack, actx)
}

func (r *Runtime) DispatchCurrentActiveIdsEvent() {
	cast.PushState(r.activeActionsID.Slice(), cast.EventRuntimeStarted, r.irb.ExecutionUUID)
}

// switchContext activates actx and deactivates his parents
func (r *Runtime) switchContext(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	todeactivate := actx.Parents()
	r._activateContext(actx)
	for _, dactx := range todeactivate {
		r._deactivateContext(dactx)
	}
}

// _activateContext
//   - put actx into active actions registry
//   - initialize actx event listener
//   - if this action is join point, add
//     to active join points
func (r *Runtime) _activateContext(actx base.IActionContext) {
	defer r.DispatchCurrentActiveIdsEvent()
	ev := r.evDispatcher.NewEventListener()
	actx.WithEventListener(ev)
	if actx.Type() == base.ContextTypeRegular {
		r.setRunFunc(actx)
		r.setDebugInitFunc(actx)
	}

	r.activeActionsID.More(actx)
}

// _deactivateContext
//   - rm actx from active actions registry
//   - destroy event listener of the actx
//   - if this actx is a join point, rm from
//     active join points
func (r *Runtime) _deactivateContext(actx base.IActionContext) {
	defer r.DispatchCurrentActiveIdsEvent()

	el := actx.EventListener()
	r.evDispatcher.DestroyEventListener(el)

	r.activeActionsID.Less(actx)
}

// NewActionContext func creates a new base.IActionContext
// (runtime.actionContext), push into the runtime action
// stack and return the newly created action context
func (r *Runtime) NewAContext(parent base.IActionContext, action *base.Action) base.IActionContext {
	if action.JoinThreadsPoint {
		return r.NewAContextJoin(parent, action)
	}
	ac := &actionContext{
		_dbgname: "actionContext",
		parent:   parent,
		action:   action,
	}
	if parent != nil {
		ac.SetStore(parent.GetStore())
		parent.Child(ac)
	}
	r.pushActionContext(ac)
	return ac
}

func (r *Runtime) NewAContextThread(parent base.IActionContext, actions []*base.Action) base.IActionContext {
	newcontexts := make([]base.IActionContext, len(actions))
	// thread point parent aka MIM context
	threadctx := &threadPointContext{
		_dbgname: "threadPointContext",
		parent:   parent,
	}

	// append a fork actx as child of the action that
	// has init a thread
	parent.Child(threadctx)

	r.pushActionContext(threadctx)
	for i := 0; i < len(actions); i++ {
		// add itered child to thread point parent
		newchild := &actionContext{
			_dbgname: "actionContext",
			action:   actions[i],
			parent:   threadctx,
			store:    parent.GetStore().Duplicate(),
		}
		newchild.store.GetLogger().SetThreadID("no-th-id")
		threadctx.Child(newchild)
		r.pushActionContext(newchild)
		newcontexts[i] = newchild
	}
	return threadctx
}

func (r *Runtime) NewAContextJoin(parent base.IActionContext, action *base.Action) base.IActionContext {
	st := parent.GetStore()
	if st == nil {
		panic("asdf")
	}
	jpoint := &joinPointContext{
		_dbgname: "joinPointContext",
		action:   action,
		parents:  make([]base.IActionContext, 1),
		store:    parent.GetStore(), // TODO: dupe instead copy?
	}
	jpoint.parents[0] = parent // this is a new jpoint and has no parents yet
	r.pushActionContext(jpoint)
	return jpoint
}
