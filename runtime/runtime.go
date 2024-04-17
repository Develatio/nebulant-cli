// Nebulant
// Copyright (C) 2024  Develatio Technologies S.L.

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

package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/nsterm"
)

type runtimeEvent struct {
	ecode base.EventCode
}

func (r *runtimeEvent) EventCode() base.EventCode { return r.ecode }
func (r *runtimeEvent) String() string            { return fmt.Sprintf("runtime event: %v", r.ecode) }

func NewRuntime(irb *blueprint.IRBlueprint, serverMode bool) *Runtime {
	return &Runtime{
		irb:                irb,
		serverMode:         serverMode,
		actionContextStack: make([]base.IActionContext, 0, 1),
		activeActionIDs:    make(map[string]int),
		activeJoinPoints:   make(map[string]base.IActionContext),
		// activeBreakPoints:  make(map[*breakPoint]bool),
		activeThreads: make(map[*Thread]bool),
		evDispatcher:  base.NewEventDispatcher(),
		exitCode:      0,
	}
}

// actioncontext type
type joinPointContext struct {
	_dbgname  string
	runStatus base.ActionContextRunStatus
	store     base.IStore
	action    *blueprint.Action
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

func (j *joinPointContext) GetAction() *blueprint.Action {
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

// actioncontext type
type threadPointContext struct {
	_dbgname  string
	runStatus base.ActionContextRunStatus
	parent    base.IActionContext
	children  []base.IActionContext
	elistener *base.EventListener
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

func (t *threadPointContext) GetAction() *blueprint.Action {
	return &blueprint.Action{
		ActionName: "forky",
		ActionID:   "no-id-just-fork",
	}
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

type actionContext struct {
	_dbgname  string
	ctx       context.Context
	runStatus base.ActionContextRunStatus
	cancel    func(error)
	initfunc  func() (*base.ActionOutput, error)
	dbgfunc   func()
	vpty      *nsterm.VPTY2
	store     base.IStore
	action    *blueprint.Action
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

func (a *actionContext) GetAction() *blueprint.Action {
	return a.action
}

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

func (t *Thread) _runCurrent() {
	if t.current == nil {
		panic("hey dev, this is your fault")
	}
	actx := t.current
	action := actx.GetAction()
	var nexts []*blueprint.Action

	if actx.IsThreadPoint() {
		t.runtime.activateContext(actx)
		for _, fkactx := range actx.Children() {
			t.runtime.NewThread(fkactx)
		}
		t.runtime.deactivateContext(actx)
		t.done = append(t.done, actx)
		// stop exec, no more actx stored in queue
		// so this should be the last actx
		return
	}

	if action.JoinThreadsPoint {
		t.runtime.activateContext(actx)
		if t.runtime.IsRunningParentsOf(actx) {
			// here the thread should end. Other threads
			// will reexec this action at thread finish.
			// Once there is no running parents, the last
			// thread is reused to continue the run of
			// actions connected to join point.
			return
		}
		t.runtime.deactivateContext(actx)
	}

	t.runtime.activateContext(actx)
	t.ThreadStep = ThreadIntoAction
	aout, aerr := actx.RunAction()
	t.runtime.deactivateContext(actx)

	t.ThreadStep = ThreadAfterAction

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
		t.current.Cancel(nil)
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

type Runtime struct {
	mu            sync.Mutex
	ru            sync.Mutex
	serverMode    bool
	state         base.RuntimeState
	activeThreads map[*Thread]bool
	// eventListeners     map[*base.EventListener]bool
	evDispatcher       *base.EventDispatcher
	irb                *blueprint.IRBlueprint
	actionContextStack []base.IActionContext
	activeActionIDs    map[string]int
	// join points
	activeJoinPoints map[string]base.IActionContext
	exitCode         int
	exitErrs         []error // uncaught err
	//
	savedActionOutputs []*base.ActionOutput
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

func (r *Runtime) NewThread(actx base.IActionContext) {
	if r.state == base.RuntimeStateEnding || r.state == base.RuntimeStateEnd {
		return
	}
	cast.LogDebug("Stat new thread", r.irb.ExecutionUUID)
	r.mu.Lock()
	defer r.mu.Unlock()
	el := r.evDispatcher.NewEventListener()
	th := &Thread{
		runtime:   r,
		elistener: el,
		step:      make(chan *threadStackCtrl),
	}
	th.queue = append(th.queue, actx)
	r.activeThreads[th] = true

	// start thread in play or pause mode
	if r.state == base.RuntimeStatePlay {
		th.Play()
	} else if r.state == base.RuntimeStateStill {
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
	cast.PushEvent(cast.EventRuntimeStarted, r.irb.ExecutionUUID)
	r.DispatchCurrentActiveIdsEvent()
}

func (r *Runtime) Pause() {
	cast.PushEvent(cast.EventRuntimePausing, r.irb.ExecutionUUID)
	go r.evDispatcher.Dispatch(&runtimeEvent{ecode: base.RuntimeStillEvent})
	threads := r.GetThreads()
	for th := range threads {
		th.Pause()
	}
	r.state = base.RuntimeStateStill
	cast.PushEvent(cast.EventRuntimePaused, r.irb.ExecutionUUID)
	r.DispatchCurrentActiveIdsEvent()
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
	cast.PushEvent(cast.EventRuntimeOut, r.irb.ExecutionUUID)
	r.DispatchCurrentActiveIdsEvent()
}

func (r *Runtime) GetThreads() map[*Thread]bool {
	return r.activeThreads
}

func (r *Runtime) IsRunningParentsOf(actx base.IActionContext) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	action := actx.GetAction()
	if action == nil {
		panic("Hey dev, this is your fault. There is an actionContext without action :S")
	}
	for aid := range action.KnowParentIDs {
		if _, exists := r.activeActionIDs[aid]; exists {
			if r.activeActionIDs[aid] > 0 {
				return true
			}
		}
	}
	return false
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
	if actx.Type() == base.ContextTypeJoin {
		action := actx.GetAction()
		r.activeJoinPoints[action.ActionID] = actx
	}
}

func (r *Runtime) GetActiveActionIds() []string {
	var cc []string
	for actnID, howMany := range r.activeActionIDs {
		if howMany > 0 {
			cc = append(cc, actnID)
		}
	}
	return cc
}

func (r *Runtime) DispatchCurrentActiveIdsEvent() {
	cast.PushState(r.GetActiveActionIds(), cast.EventRuntimeStarted, r.irb.ExecutionUUID)
}

func (r *Runtime) activateContext(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.DispatchCurrentActiveIdsEvent()

	ev := r.evDispatcher.NewEventListener()
	actx.WithEventListener(ev)
	if actx.Type() == base.ContextTypeRegular {
		r.setRunFunc(actx)
		r.setDebugInitFunc(actx)
	}

	action := actx.GetAction()
	// already active join action, skip
	if actx.Type() == base.ContextTypeJoin {
		if _, exists := r.activeJoinPoints[action.ActionID]; exists {
			return
		}
		r.activeJoinPoints[action.ActionID] = actx
	}

	if _, exists := r.activeActionIDs[action.ActionID]; !exists {
		r.activeActionIDs[action.ActionID] = 0
	}
	r.activeActionIDs[action.ActionID]++
}

func (r *Runtime) deactivateContext(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.DispatchCurrentActiveIdsEvent()

	el := actx.EventListener()
	r.evDispatcher.DestroyEventListener(el)

	action := actx.GetAction()
	if _, exists := r.activeActionIDs[action.ActionID]; exists {
		r.activeActionIDs[action.ActionID]--
	}
	if actx.Type() == base.ContextTypeJoin {
		delete(r.activeJoinPoints, action.ActionID)
	}
}

func (r *Runtime) GetActiveJoinPoints() map[string]base.IActionContext {
	return r.activeJoinPoints
}

// NewActionContext func creates a new base.IActionContext
// (runtime.actionContext), push into the runtime action
// stack and return the newly created action context
func (r *Runtime) NewAContext(parent base.IActionContext, action *blueprint.Action) base.IActionContext {
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

func (r *Runtime) NewAContextThread(parent base.IActionContext, actions []*blueprint.Action) base.IActionContext {
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
		threadctx.Child(newchild)
		r.pushActionContext(newchild)
		newcontexts[i] = newchild
	}
	return threadctx
}

func (r *Runtime) NewAContextJoin(parent base.IActionContext, action *blueprint.Action) base.IActionContext {
	r.ru.Lock()
	defer r.ru.Unlock()
	if _, exists := r.activeJoinPoints[action.ActionID]; exists {
		jpoint := r.activeJoinPoints[action.ActionID]
		// add new parent to join point
		jpoint.Parent(parent)
		// merge the store of the new parent
		store := jpoint.GetStore()
		store.Merge(parent.GetStore())
		jpoint.SetStore(store)
		return jpoint
	}
	jpoint := &joinPointContext{
		_dbgname: "joinPointContext",
		action:   action,
		parents:  make([]base.IActionContext, 1),
		store:    parent.GetStore(),
	}
	jpoint.parents[0] = parent
	r.pushActionContext(jpoint)
	return jpoint

}
