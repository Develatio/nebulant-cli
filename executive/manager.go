// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

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

package executive

import (
	"fmt"
	"math/rand"
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
	Runtime *runtime.Runtime
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
	m.Logger.SetThreadID(fmt.Sprintf("%d", rand.Int()))
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
		cast.PushEvent(cast.EventRuntimeOut, m.ExecutionUUID)
	}()
	if exit {
		return nil
	}

	m.Logger.LogDebug("[Manager] Starting...")

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
				Action:  nil,
			}, "")
			if err != nil {
				return err
			}
		}
	}

	m.Logger.ParanoicLogDebug(fmt.Sprintf("[Manager] Setting %s as start point", m.IRB.StartAction.ActionName))

	startActionContext := m.Runtime.NewAContext(nil, m.IRB.StartAction)
	m.Logger.ParanoicLogDebug("after set context")
	st.SetLogger(m.GetLogger())
	startActionContext.SetStore(st)
	m.Logger.ParanoicLogDebug("after set store")

	eventlistener := m.Runtime.NewEventListener()
	// start to run
	m.Runtime.NewThread(startActionContext)

	cast.PushEvent(cast.EventRuntimeStarted, m.ExecutionUUID)
	m.Logger.ParanoicLogDebug("after push event")
	m.Logger.ParanoicLogDebug("after set manager state")
	m.Logger.LogDebug("[Manager] Ready")
	// for run stats
	startTime := time.Now()

	// freezes until runtime end
	eventlistener.WaitUntil([]base.EventCode{base.RuntimeEndEvent})

	m.mu.Lock()
	defer m.mu.Unlock()

	err = m.Runtime.Error()
	if err != nil {
		errs := m.Runtime.Errors()
		cerr := len(errs)
		lerr, errs := errs[len(errs)-1], errs[:len(errs)-1]
		cast.LogErr(fmt.Sprintf("%v unhandled errors found. Summary:\n***\n", cerr), m.ExecutionUUID)
		for _, rr := range errs {
			cast.LogErr(rr.Error(), m.ExecutionUUID)
		}
		cast.LogErr(fmt.Sprintf("%s\n\n***", lerr.Error()), m.ExecutionUUID)
	}

	m.Logger.LogInfo(fmt.Sprintf("Blueprint runtime finished with exit code %v", m.Runtime.ExitCode()))
	m.Logger.LogInfo("[Manager] out")

	elapsedTime := time.Since(startTime).String()
	m.Logger.LogInfo("[Manager] stats: " + strconv.Itoa(m.Stats.actions) + " actions executed by " + strconv.Itoa(m.Stats.stages) + " stages in " + elapsedTime)
	if m.ExecutionUUID != nil {
		m.Logger.LogDebug("Sending EventRuntimeOut for UUID" + *m.ExecutionUUID)
	} else {
		m.Logger.LogDebug("Sending EventRuntimeOut for no UUID")
	}

	cast.PushEvent(cast.EventRuntimeOut, m.ExecutionUUID)
	// m.internalRegistry.SetManagerState(cast.EventRuntimeOut)
	// m.ExternalRegistry.SetManagerState(cast.EventRuntimeOut)
	m.Logger.LogDebug("[Manager] out")
	return nil
}
