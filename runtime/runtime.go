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
	"fmt"
	"sync"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
)

func NewRuntime(irb *blueprint.IRBlueprint) *Runtime {
	return &Runtime{
		irb:                irb,
		actionContextStack: make([]base.IActionContext, 0, 1),
		activeActionIDs:    make(map[string]int),
		activeJoinPoints:   make(map[string]base.IActionContext),
	}
}

// actioncontext type
type joinPointContext struct {
	_dbgname string
	store    base.IStore
	action   *blueprint.Action
	parents  []base.IActionContext
	child    base.IActionContext
}

func (j *joinPointContext) SetStore(s base.IStore) {
	j.store = s
}

func (j *joinPointContext) GetStore() base.IStore {
	return j.store
}

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
func (j *joinPointContext) WithCancelCause()      {}
func (j *joinPointContext) Cancel(e error)        {}

// actioncontext type
type threadPointContext struct {
	_dbgname string
	parent   base.IActionContext
	children []base.IActionContext
}

func (t *threadPointContext) SetStore(s base.IStore) {}

func (t *threadPointContext) GetStore() base.IStore {
	return nil
}

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
func (t *threadPointContext) WithCancelCause()      {}
func (t *threadPointContext) Cancel(e error)        {}

type actionContext struct {
	_dbgname string
	ctx      context.Context
	cancel   func(error)
	initfunc func() (*base.ActionOutput, error)
	store    base.IStore
	action   *blueprint.Action
	parent   base.IActionContext
	child    base.IActionContext
}

func (a *actionContext) Done() <-chan struct{} {
	if a.ctx == nil {
		return nil
	}
	return a.ctx.Done()
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

type Runtime struct {
	mu                 sync.Mutex
	irb                *blueprint.IRBlueprint
	actionContextStack []base.IActionContext
	activeActionIDs    map[string]int
	// join points
	activeJoinPoints map[string]base.IActionContext
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

func (r *Runtime) setRunFunc(actx base.IActionContext) {
	actx.WithRunFunc(func() (*base.ActionOutput, error) {
		defer r.Stop(actx)
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
		r.Run(actx)
		aout, err := provider.HandleAction(actx)
		return aout, err
	})
}

func (r *Runtime) PushActionContext(actx base.IActionContext) {
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

	if actx.Type() == base.ContextTypeRegular {
		r.setRunFunc(actx)
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
func (r *Runtime) Run(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// WIP: el estado que se manda en PushState es el que
	// actualmente se almacena en report.go, se trata
	// del estado actual del manager, mover el estado
	// de la ejecución a Runtime
	defer cast.PushState(r.GetActiveActionIds(), cast.EventManagerStarted, r.irb.ExecutionUUID)
	// r.runningContexts[actx.(*actionContext)] = true
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

func (r *Runtime) Stop(actx base.IActionContext) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer cast.PushState(r.GetActiveActionIds(), cast.EventManagerStarted, r.irb.ExecutionUUID)
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
	r.PushActionContext(ac)
	return ac
}

// WIP: me he dado cuenta de que si hacemos esto, el padre del padre sólo puede hacer
// referencia a una de las copias como su hijo, por lo que se pierde la referencia: buscar solución
func (r *Runtime) NewAContextThread(parent base.IActionContext, actions []*blueprint.Action) []base.IActionContext {
	newcontexts := make([]base.IActionContext, len(actions))
	// thread point parent aka MIM context
	threadpp := &threadPointContext{
		_dbgname: "threadPointContext",
		parent:   parent,
	}
	r.PushActionContext(threadpp)
	for i := 0; i < len(actions); i++ {
		// add itered child to thread point parent
		newchild := &actionContext{
			_dbgname: "actionContext",
			action:   actions[i],
			parent:   threadpp,
			store:    parent.GetStore().Duplicate(),
		}
		threadpp.Child(newchild)
		r.PushActionContext(newchild)
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
	r.PushActionContext(jpoint)
	return jpoint

}
