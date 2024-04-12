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
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/util"
)

// MDirector var
var MDirector *Director

// InitDirector func
func InitDirector(serverMode bool, interactiveMode bool) error {
	if MDirector != nil {
		return fmt.Errorf("director already running")
	}
	directorWaiter := &sync.WaitGroup{}
	directorWaiter.Add(1) // self
	MDirector = &Director{
		managers:          make(map[*Manager]*blueprint.IRBlueprint),
		managersByIRB:     make(map[*blueprint.IRBlueprint]*Manager),
		HandleIRB:         make(chan *HandleIRBConfig, 10),
		ExecInstruction:   make(chan *ExecCtrlInstruction, 10),
		UnregisterManager: make(chan *Manager, 10),
		directorWaiter:    directorWaiter,
		serverMode:        serverMode,
		interactiveMode:   interactiveMode,
	}
	go MDirector.startDirector()
	return nil
}

type HandleIRBConfig struct {
	Manager *Manager
	IRB     *blueprint.IRBlueprint
}

// Director struct
type Director struct {
	ExecInstruction   chan *ExecCtrlInstruction
	serverMode        bool
	interactiveMode   bool
	HandleIRB         chan *HandleIRBConfig
	UnregisterManager chan *Manager
	StopDirector      chan int
	directorWaiter    *sync.WaitGroup
	managers          map[*Manager]*blueprint.IRBlueprint
	managersByIRB     map[*blueprint.IRBlueprint]*Manager
	ExitCode          int
}

// Wait func
func (d *Director) Wait() {
	d.directorWaiter.Wait()
}

func (d *Director) Clean() {
	MDirector = nil
}

// startDirector func
func (d *Director) startDirector() error {
	defer d.directorWaiter.Done()
	cast.PushEvent(cast.EventDirectorStarting, nil)
	cast.PushEvent(cast.EventDirectorStarted, nil)

L:
	for { // Infine loop until break L
		select { // Loop until a case ocurrs.
		case instr := <-d.ExecInstruction:
			cast.LogInfo("[Director] Received instruction with id "+*instr.ExecutionUUID, nil)
			if len(d.managers) <= 0 {
				cast.LogInfo("[Director] No managers available", nil)
				cast.PushEvent(cast.EventRuntimeOut, instr.ExecutionUUID)
				continue
			}
			managerFound := false
			for manager := range d.managers {
				if instr.ExecutionUUID != nil && *manager.ExecutionUUID == *instr.ExecutionUUID {
					managerFound = true
					cast.LogInfo("[Director] Sending Instruction to Runtime", nil)
					switch instr.Instruction {
					case ExecStop:
						manager.Runtime.Stop()
					case ExecStart:
						manager.Runtime.Play()
					case ExecPause:
						manager.Runtime.Pause()
					case ExecResume:
						manager.Runtime.Play()
					case ExecState:
						cast.LogInfo("NOOP instruction", nil)
					case ExecEmancipation:
						cast.LogInfo("NOOP instruction", nil)
					}
					break
				}
			}
			if !managerFound {
				cast.LogInfo("[Director] No manager found for instruction", nil)
				cast.PushEvent(cast.EventRuntimeOut, instr.ExecutionUUID)
			}
		case hirbcfg := <-d.HandleIRB:
			irb := hirbcfg.IRB
			manager := hirbcfg.Manager

			// if manager is nil, hirbcfg.IRB should be configured
			if manager == nil {
				manager = NewManager(d.serverMode)
				manager.PrepareIRB(irb)
			}

			// if irb is nil, manager.IRB should be configured
			if irb == nil {
				irb = manager.IRB
			}

			if irb.BP.BuilderWarnings > 0 {
				cast.LogWarn("This blueprint has "+fmt.Sprintf("%v", irb.BP.BuilderWarnings)+" warnings from the builder", irb.BP.ExecutionUUID)
			}
			d.managersByIRB[irb] = manager
			d.managers[manager] = irb
			extra := make(map[string]interface{})
			extra["manager"] = manager
			cast.PushEventWithExtra(cast.EventRegisteredManager, irb.BP.ExecutionUUID, extra)
			cast.LogDebug("[Director] sending irb to manager...", nil)
			go func() {
				exit := false
				defer func() {
					exit = true
					if r := recover(); r != nil {
						cast.LogErr("Unrecoverable error found. Feel free to send us feedback", irb.BP.ExecutionUUID)
						switch r := r.(type) {
						case *util.PanicData:
							v := fmt.Sprintf("%v", r.PanicValue)
							cast.LogErr(v, irb.BP.ExecutionUUID)
							cast.LogErr(string(r.PanicTrace), irb.BP.ExecutionUUID)
						default:
							cast.LogErr("Panic", irb.BP.ExecutionUUID)
							cast.LogErr("If you think this is a bug,", irb.BP.ExecutionUUID)
							cast.LogErr("please consider posting stack trace as a GitHub", irb.BP.ExecutionUUID)
							cast.LogErr("issue (https://github.com/develatio/nebulant-cli/issues)", irb.BP.ExecutionUUID)
							cast.LogErr("Stack Trace:", irb.BP.ExecutionUUID)
							cast.LogErr("Panic", irb.BP.ExecutionUUID)
							v := fmt.Sprintf("%v", r)
							cast.LogErr(v, irb.BP.ExecutionUUID)
							cast.LogErr(string(debug.Stack()), irb.BP.ExecutionUUID)
						}
						if !d.serverMode && !d.interactiveMode {
							os.Exit(1)
						}
					}
				}()
				if exit {
					return
				}

				defer func() {
					d.UnregisterManager <- manager
				}()
				cast.LogDebug("Starting manager loop...", nil)
				err := manager.Run()
				if err != nil {
					cast.LogErr(err.Error(), nil)
				}
				cast.LogDebug("[Director] manager has end Run()", nil)
			}()
		case manager := <-d.UnregisterManager:
			cast.LogInfo("[Director] Unregistering Manager", nil)
			// exitCode := manager.ExternalRegistry.ExitCode
			exitCode := manager.Runtime.ExitCode()

			manager.reset()
			irb := d.managers[manager]
			delete(d.managers, manager)
			delete(d.managersByIRB, irb)
			if len(d.managers) <= 0 && !d.serverMode {
				d.Clean()
				d.ExitCode = exitCode
				break L
			}
		default:
			time.Sleep(2 * time.Second)
		}
	}

	cast.PushEvent(cast.EventDirectorOut, nil)
	cast.LogInfo("Director out", nil)
	return nil
}
