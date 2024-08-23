// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

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

package executive

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/ipc"
	"github.com/develatio/nebulant-cli/runtime"
	"github.com/develatio/nebulant-cli/storage"
	"github.com/develatio/nebulant-cli/util"
)

// ManagerStatusID int
type ManagerStatusID int

// NewManager func
func NewManager(serverMode bool) *Manager {
	mng := &Manager{serverMode: serverMode}
	// init vars calling reset
	mng.reset()
	return mng
}

// Manager struct
type Manager struct {
	Runtime          *runtime.Runtime
	runtimeStopEvent chan struct{}
	// The uuid generated by the builder or uuid asked to backend
	ExecutionUUID *string
	IRB           *blueprint.IRBlueprint
	serverMode    bool
	Logger        *cast.Logger
	Stats         *stats
	mu            sync.Mutex
}

type stats struct {
	stages  int
	actions int
}

func (m *Manager) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Runtime = nil
	m.ExecutionUUID = nil
	m.Stats = &stats{}
	m.Logger = &cast.Logger{}
	m.Logger.SetThreadID("manager")
}

// GetLogger func
func (m *Manager) GetLogger() base.ILogger {
	return m.Logger
}

// PrepareIRB func
func (m *Manager) PrepareIRB(irb *blueprint.IRBlueprint) {
	// Reset all and prepare for new blueprint
	m.reset()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IRB = irb
	m.Runtime = runtime.NewRuntime(irb, m.serverMode)
	eventlistener := m.Runtime.NewEventListener()
	// quick register end event waiter. Do this just after runtime creation
	// to prevent race condition of start-end/end-start events order
	m.runtimeStopEvent = eventlistener.WaitUntilChan([]base.EventCode{base.RuntimeEndEvent})
	m.ExecutionUUID = irb.ExecutionUUID
	if m.ExecutionUUID != nil {
		re := *m.ExecutionUUID
		m.Logger.ExecutionUUID = &re
	}
}

// Run func
func (m *Manager) Run() error {
	exit := false
	defer func() {
		exit = true
		if r := recover(); r != nil {
			switch r.(type) {
			case *util.PanicData:
				panic(r)
			}
			panic(&util.PanicData{
				PanicValue: r,
				PanicTrace: debug.Stack(),
			})
		}
	}()
	if exit {
		return nil
	}

	m.Logger.LogDebug("[Manager] Starting...")

	cast.LogDebug("Sending EventRuntimeStarting", nil)
	cast.PushEvent(cast.EventRuntimeStarting, m.ExecutionUUID)
	if m.IRB.StartAction == nil {
		return fmt.Errorf("[Manager] First action id not found")
	}

	// conf ipcs server
	ipcs, err := ipc.NewIPCServer()
	if err != nil {
		return err
	}
	go func() {
		err := ipcs.Accept()
		if err != nil {
			m.Logger.LogErr(err.Error())
		}
	}()
	defer ipcs.Close()

	// init store
	st := storage.NewStore()

	// set ipcs into store
	st.SetPrivateVar("IPCS", ipcs)

	// set vars from cli args
	for _, irbarg := range m.IRB.Args {
		if st.ExistsRefName(irbarg.Name) {
			// this is an stack var wich can acumulate values
			err := st.Push(&base.StorageRecord{
				RefName: irbarg.Name,
				Aout:    nil,
				Value:   irbarg.Value,
				Action:  nil,
			}, "")
			if err != nil {
				return err
			}
		} else {
			err := st.Insert(&base.StorageRecord{
				RefName: irbarg.Name,
				Aout:    nil,
				Value:   irbarg.Value,
				Literal: true,
				Action:  nil,
			}, "")
			if err != nil {
				return err
			}
		}
	}

	m.Logger.ParanoidLogDebug(fmt.Sprintf("[Manager] Setting %s as start point", m.IRB.StartAction.ActionName))

	startActionContext := m.Runtime.NewAContext(nil, m.IRB.StartAction)
	m.Logger.ParanoidLogDebug("after set context")
	st.SetLogger(m.GetLogger().Duplicate())
	startActionContext.SetStore(st)
	m.Logger.ParanoidLogDebug("after set store")

	// for run stats
	startTime := time.Now()

	// start to run
	if m.Runtime.NewThread(startActionContext) {
		cast.LogDebug("Sending EventRuntimeStarted", nil)
		cast.PushEvent(cast.EventRuntimeStarted, m.ExecutionUUID)
		m.Logger.ParanoidLogDebug("after push event")
		m.Logger.ParanoidLogDebug("after set manager state")
		m.Logger.LogDebug("[Manager] Ready")

		// freezes until runtime ends
		<-m.runtimeStopEvent
	} else {
		cast.LogDebug("cannot start new run. Probably a previous stop received", m.ExecutionUUID)
	}

	m.Logger.ParanoidLogDebug("after wait until")

	m.mu.Lock()
	defer m.mu.Unlock()

	m.Logger.LogInfo(fmt.Sprintf("Blueprint runtime finished with exit code %v", m.Runtime.ExitCode()))
	err = m.Runtime.Error()
	if err != nil {
		errs := m.Runtime.Errors()
		cerr := len(errs)
		lerr, errs := errs[len(errs)-1], errs[:len(errs)-1]
		cast.LogInfo(fmt.Sprintf("%v unhandled errors found. Summary:\n***\n", cerr), m.ExecutionUUID)
		for _, rr := range errs {
			cast.LogErr(rr.Error(), m.ExecutionUUID)
		}
		cast.LogErr(fmt.Sprintf("%s\n\n***", lerr.Error()), m.ExecutionUUID)
	}

	elapsedTime := time.Since(startTime).String()
	m.Logger.LogInfo("[Manager] stats: " + strconv.Itoa(m.Stats.actions) + " actions executed by " + strconv.Itoa(m.Stats.stages) + " stages in " + elapsedTime)

	// m.internalRegistry.SetManagerState(cast.EventRuntimeOut)
	// m.ExternalRegistry.SetManagerState(cast.EventRuntimeOut)
	m.Logger.LogInfo("[Manager] out")
	return nil
}
