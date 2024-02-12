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
	"log"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
)

type RunState int

const (
	RuntimePlay RunState = iota
	RuntimeStill
	RuntimeEnd
)

type runtimeEvent struct {
	ecode base.EventCode
}

func (r *runtimeEvent) EventCode() base.EventCode { return r.ecode }
func (r *runtimeEvent) String() string            { return fmt.Sprintf("runtime event: %v", r.ecode) }

type breakPoint struct {
	suscribers []chan struct{}
	actx       base.IActionContext
	err        error
}

func (b *breakPoint) Subscribe() <-chan struct{} {
	s := make(chan struct{})
	b.suscribers = append(b.suscribers, s)
	return s
}

// WIP quizás End podría aceptar un subscriber para
// cerrar sólo ese
func (b *breakPoint) End() {
	for i := 0; i < len(b.suscribers); i++ {
		close(b.suscribers[i])
	}
}

func (b *breakPoint) GetActionContext() base.IActionContext {
	return b.actx
}

func NewRuntime(irb *blueprint.IRBlueprint) *Runtime {
	return &Runtime{
		irb:                irb,
		actionContextStack: make([]base.IActionContext, 0, 1),
		activeActionIDs:    make(map[string]int),
		activeJoinPoints:   make(map[string]base.IActionContext),
		activeBreakPoints:  make(map[*breakPoint]bool),
		activeThreads:      make(map[*Thread]bool),
		eventListeners:     make(map[*base.EventListener]bool),
		exitCode:           0,
	}
}

// actioncontext type
type joinPointContext struct {
	_dbgname  string
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
func (j *joinPointContext) IsThreadPoint() bool    { return true }
func (j *joinPointContext) IsJoinPoint() bool      { return false }

func (j *joinPointContext) GetAction() *blueprint.Action {
	return j.action
}

func (j *joinPointContext) WithRunFunc(f func() (*base.ActionOutput, error)) {}
func (j *joinPointContext) RunAction() (*base.ActionOutput, error) {
	return nil, fmt.Errorf("this kind of actx has no executable action") // maybe in a future
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

// actioncontext type
type threadPointContext struct {
	_dbgname  string
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
	return nil
}

func (t *threadPointContext) WithRunFunc(f func() (*base.ActionOutput, error)) {}
func (t *threadPointContext) RunAction() (*base.ActionOutput, error) {
	return nil, fmt.Errorf("this kind of actx has no action")
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

type actionContext struct {
	_dbgname string
	ctx      context.Context
	cancel   func(error)
	initfunc func() (*base.ActionOutput, error)
	store    base.IStore
	action   *blueprint.Action
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

func (a *actionContext) GetAction() *blueprint.Action {
	return a.action
}

type Thread struct {
	queue     []base.IActionContext
	done      []base.IActionContext
	ExitCode  int
	ExitErr   error // uncaught err
	runtime   *Runtime
	elistener *base.EventListener
}

func (t *Thread) GetQueue() []base.IActionContext {
	return t.queue
}

func (t *Thread) EventListener() *base.EventListener {
	return t.elistener
}

// checks for [pause/still] event and waits
// until pause status ends. Returns true if
// exec should continue and false if exec
// should end
func (t *Thread) checkEvents() bool {
	if t.elistener.ReadUntil(base.RuntimeStillEvent) {
		ec := t.elistener.WaitUntil([]base.EventCode{base.RuntimePlayEvent, base.RuntimeEndEvent})
		switch ec {
		case base.RuntimePlayEvent:
			return true
		case base.RuntimeEndEvent:
			return false
		}
	}
	return true
}

// commonly called by go Init()
func (t *Thread) Init() {
	// WIP: quizás sea necesario lanzar un evento para atrás de thread
	// inicializado
	var waitcount int
	defer func() {
		t.runtime.finishThread(t)
	}()
	for {
		if len(t.queue) <= 0 {
			// WIP: quizás avisar a Runtime, mediante algún
			// chanel que el propio runtime cree y setee en el
			// thread, que este thread ha terminado
			break
		}

		/////////////////////////////////////
		if !t.checkEvents() { ///////////////
			return //////////////////////////
		} ///////////////////////////////////
		/////////////////////////////////////

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

		actx := t.queue[0]
		action := actx.GetAction()
		var nexts []*blueprint.Action
		if action.JoinThreadsPoint {
			t.runtime.activateContext(actx)
			if t.runtime.IsRunningParentsOf(actx) {
				// WIP: quizás informar a runtime
				// que este thread ha terminado
				// WIP: quizás mover esto a un WithRunFunc
				// y tratar aquí la caja de JoinTreads como
				// caja normal: si el WithRunFunc detecta
				// que él mismo es el último de los threads
				// que devuelva las "nexts" actions
				//
				// de momento sólo hacemos skip aquí si
				// aún hay un parent trabajando
				// de ser el último thread, el resto
				// del código debería poder gestionar
				// este action
				return
			}
			t.runtime.deactivateContext(actx)
		}

		/////////////////////////////////////
		if !t.checkEvents() { ///////////////
			return //////////////////////////
		} ///////////////////////////////////
		/////////////////////////////////////

		// WIP: testear si el action a ejecutar es un breakpoint
		// o es un JoinThreads porque en ese caso hay que actuar
		// de forma distinta.
		// JoinThreads: el thread actual debe terminar darle el
		// aviso a runtime para que arranque un nuevo thread
		// en caso necesario, del mismo modo que lo hace ahora
		// el manager cuando le llegan stageReports de JoinThreadsPoint
		// aout and aerr are self stored by RunAction
		t.runtime.activateContext(actx)
		aout, aerr := actx.RunAction()
		t.runtime.deactivateContext(actx)

		/////////////////////////////////////
		if !t.checkEvents() { ///////////////
			return //////////////////////////
		} ///////////////////////////////////
		/////////////////////////////////////

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
			aerr = errors.Join(aerr, err)
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

		/////////////////////////////////////
		if !t.checkEvents() { ///////////////
			return //////////////////////////
		} ///////////////////////////////////
		/////////////////////////////////////

		switch len(nexts) {
		case 0:
			// no more actions, errs and exit code
			// keep as setted before
			// WIP: aquí sería un buen lugar para devolver info
			// a runtime
			return
		case 1:
			// reset exit code and err if exists
			t.ExitCode = 0
			t.ExitErr = nil
			nactx := t.runtime.NewAContext(actx, nexts[0])
			t.done = append(t.done, actx)
			t.queue = append(t.queue, nactx)
			t.queue = t.queue[1:] // rm current actx wich is in index 0
		default:
			// reset exit code and err if exists
			t.ExitCode = 0
			t.ExitErr = nil
			// more than one, new threads needed
			actxs := t.runtime.NewAContextThread(actx, nexts)
			for _, actx := range actxs {
				t.runtime.NewThread(actx)
			}
			// WIP: aquí quizas informar a runtime de que el thread ha terminado?
			return
		}
	}
}

// WIP quizás sería intersante añadir aquí un channel, o un sistema
// tipo cast, donde Stage, Manager y demás se pudieran suscribir para
// recibir eventos de stop y demás
type Runtime struct {
	mu                 sync.Mutex
	state              RunState
	activeThreads      map[*Thread]bool
	eventListeners     map[*base.EventListener]bool
	irb                *blueprint.IRBlueprint
	actionContextStack []base.IActionContext
	activeActionIDs    map[string]int
	// join points
	activeJoinPoints map[string]base.IActionContext
	// WIP: de momento bool sin usar, se deja para un
	// uso posterior
	activeBreakPoints map[*breakPoint]bool
	//
	exitCode int
	exitErrs []error // uncaught err
	//
	savedActionOutputs []*base.ActionOutput
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

func (r *Runtime) NewEventListener() *base.EventListener {
	el := base.NewEventListener()
	r.eventListeners[el] = true
	return el
}

func (r *Runtime) Dispatch(e base.IEvent) {
	switch e.EventCode() {
	case base.RuntimePlayEvent:
		r.state = RuntimePlay
	case base.RuntimeEndEvent:
		r.state = RuntimeEnd
	case base.RuntimeStillEvent:
		r.state = RuntimeStill
	}
	for el := range r.eventListeners {
		el.EventChan() <- e
	}
}

func (r *Runtime) NewThread(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	el := r.NewEventListener()
	th := &Thread{runtime: r, elistener: el}
	th.queue = append(th.queue, actx)
	r.activeThreads[th] = true
	// WIP:
	// aquí quizás un
	// go th.Init() ????
	r.Dispatch(&runtimeEvent{ecode: base.RuntimePlayEvent})
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
		r.Dispatch(&runtimeEvent{ecode: base.RuntimeEndEvent})
	}

	el := th.EventListener()
	delete(r.eventListeners, el)

	fmt.Println("threads:", len(r.activeThreads))
}

// func (r *Runtime) RuntimeOn() error {
// 	r.activity = ActivityPlay
// 	return nil
// }

// func (r *Runtime) RuntimeOff() error {
// 	r.activity = ActivityPause
// 	return nil
// }

// func (r *Runtime) RuntimeStop() error {
// 	r.activity = ActivityStop
// 	return nil
// }

func (r *Runtime) GetBreakPoints() map[*breakPoint]bool {
	return r.activeBreakPoints
}

func (r *Runtime) GetThreads() map[*Thread]bool {
	return r.activeThreads
}

// func (r *Runtime) Subscribe() <-chan int {
// 	return make(<-chan int)
// }

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

func (r *Runtime) _setRunBreakpointFunc(actx base.IActionContext) {
	actx.WithRunFunc(func() (*base.ActionOutput, error) {
		// WIP: si ya hay un server arrancado, evitar arrancar otro
		r.activateContext(actx)
		defer r.deactivateContext(actx)

		r.Dispatch(&runtimeEvent{ecode: base.RuntimeStillEvent})

		dbg := NewDebugger(r)
		go dbg.Serve()
		// WIP: añadir breakpoints globales, en lugar de
		// crear uno nuevo, suscribirse al existente
		breakpoint := &breakPoint{actx: actx}
		r.activeBreakPoints[breakpoint] = true
		// WIP: guardar el channel en el actx, así luego
		// podemos desuscribir de forma individual, cuando
		// la feature se implemente
		<-breakpoint.Subscribe()
		r.Dispatch(&runtimeEvent{ecode: base.RuntimePlayEvent})
		return nil, breakpoint.err
	})
}

func (r *Runtime) setRunFunc(actx base.IActionContext) {
	action := actx.GetAction()
	if action.BreakPoint {
		r._setRunBreakpointFunc(actx)
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
			// WIP: ahora que tenemos runtime, parece más
			// razonable inicializar el provider en el action
			// context y que los hijos hereden el provider
			// del actioncontext
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
		if aout != nil && aerr != nil {
			panic("Hey dev!, this is your fault!")
		}
		if aout != nil {
			for idx := 0; idx < len(aout.Records); idx++ {
				err := store.Insert(aout.Records[idx], action.Provider)
				if err != nil {
					log.Panic(err.Error())
				}
			}
		} else if aerr != nil {
			aout = base.NewActionOutput(action, nil, nil)
			aout.Records[0].Fail = true
			aout.Records[0].Error = aerr
			aout.Records[0].Value = aerr.Error()
			err := store.Insert(aout.Records[0], action.Provider)
			if err != nil {
				log.Panic(err.Error())
			}
		}

		if action.SaveRawResults {
			r.saveActionOutput(aout)
		}

		return aout, aerr
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

// WIP: cuando se mueva el actual report.managerState
// hasta aquí, recordar que es necesario hacer cast.PushState
// cuando cambie el estado de la ejecución
func (r *Runtime) GetActiveActionIds() []string {
	var cc []string
	for actnID, howMany := range r.activeActionIDs {
		if howMany > 0 {
			cc = append(cc, actnID)
		}
	}
	return cc
}

// WIP: quizás este func pueda inicializar una serie
// de canales dentro de actionContext, que podrían ser
// listeneados por el for(select) de stage cuando
// se llama a HandleAction() en lugar de esperar el
// return de ese HandleAction(), de ese modo el
// stage no se quedaría bloqueado
func (r *Runtime) activateContext(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// WIP: el estado que se manda en PushState es el que
	// actualmente se almacena en report.go, se trata
	// del estado actual del manager, mover el estado
	// de la ejecución a Runtime
	defer cast.PushState(r.GetActiveActionIds(), cast.EventManagerStarted, r.irb.ExecutionUUID)
	// r.runningContexts[actx.(*actionContext)] = true

	ev := r.NewEventListener()
	actx.WithEventListener(ev)
	if actx.Type() == base.ContextTypeRegular {
		r.setRunFunc(actx)
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

// WIP: mover este func a func interna?
func (r *Runtime) deactivateContext(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer cast.PushState(r.GetActiveActionIds(), cast.EventManagerStarted, r.irb.ExecutionUUID)

	el := actx.EventListener()
	delete(r.eventListeners, el)

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
	}
	r.pushActionContext(ac)
	return ac
}

func (r *Runtime) NewAContextThread(parent base.IActionContext, actions []*blueprint.Action) []base.IActionContext {
	newcontexts := make([]base.IActionContext, len(actions))
	// thread point parent aka MIM context
	threadpp := &threadPointContext{
		_dbgname: "threadPointContext",
		parent:   parent,
	}
	r.pushActionContext(threadpp)
	for i := 0; i < len(actions); i++ {
		// add itered child to thread point parent
		newchild := &actionContext{
			_dbgname: "actionContext",
			action:   actions[i],
			parent:   threadpp,
			store:    parent.GetStore().Duplicate(),
		}
		threadpp.Child(newchild)
		r.pushActionContext(newchild)
		newcontexts[i] = newchild
	}
	return newcontexts
}

func (r *Runtime) NewAContextJoin(parent base.IActionContext, action *blueprint.Action) base.IActionContext {
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
