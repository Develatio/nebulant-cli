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

package cast

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/config"
)

// CriticalLevel const
const CriticalLevel = 50

// ErrorLevel const
const ErrorLevel = 40

// WarningLevel const
const WarningLevel = 30

// InfoLevel const
const InfoLevel = 20

// DebugLevel const
const DebugLevel = 10

// NotsetLevel const
const NotsetLevel = 0

// SBusBufferSize const
const SBusBufferSize = 100000

// SBus is a globally shared system bus
var SBus *SystemBus

// Load of the system bus channel
// var BusLoad float64 = 0.0

var BInfo *BusInfo

func init() {
	BInfo = &BusInfo{Load: 0.0}
}

// InitSystemBus func
func InitSystemBus() {
	castWaiter := &sync.WaitGroup{}
	castWaiter.Add(1) // self

	SBus = &SystemBus{
		Executions: make(map[string]bool),
		// feedback
		connect:    make(chan *BusConsumerLink),
		disconnect: make(chan *BusConsumerLink),
		busBuffer:  make(chan *BusData, SBusBufferSize),
		links:      make(map[*BusConsumerLink]bool),
		// waiter
		castWaiter: castWaiter,
		// providers
		providerInitFuncs: make(map[string]base.ProviderInitFunc),
	}
	go SBus.Start()
}

// BusDataType int
type BusDataType int

const (
	// BusDataTypeLog const 0
	BusDataTypeLog BusDataType = iota
	// BusDataTypeEvent const 1
	BusDataTypeEvent
	// BusDataTypeStatus const 2
	BusDataTypeStatus
	// BusDataTypeEOF const 4
	BusDataTypeEOF
)

const (
	// EventDirectorStarting const 0
	EventDirectorStarting int = iota
	// EventDirectorStarted const 1
	EventDirectorStarted
	// EventDirectorPause const 2
	EventDirectorPause
	// EventDirectorOut const 3
	EventDirectorOut
	// EventManagerPrepareBPStart const 4
	EventManagerPrepareBPStart
	// EventManagerPrepareBPEnd const 5
	EventManagerPrepareBPEnd
	// EventManagerPrepareBPEndWithErr const 6
	EventManagerPrepareBPEndWithErr
	// EventManagerStarting const 7
	EventManagerStarting
	// EventManagerResuming const 8
	EventManagerResuming
	// EventManagerStarted const 9
	EventManagerStarted
	// EventManagerPausing const 10
	EventManagerPausing
	// EventManagerPause const 11
	EventManagerPaused
	// EventManagerStopping const 12
	EventManagerStopping
	// EventManagerOut const 13
	EventManagerOut
	// EventRegisteredManager 14
	EventRegisteredManager
	// EventWaitingStatus 15
	EventWaitingForState
	// EventActionUnCaughtKO 16
	EventActionUnCaughtKO
)

// BusData struct
type BusData struct {
	Timestamp int64 `json:"timestamp"`
	//
	// Type of data
	TypeID   BusDataType `json:"type_id"`
	ActionID *string     `json:"action_id,omitempty"`
	// Msg data in bytes
	ThreadID *string `json:"thread_id,omitempty"`
	M        *string `json:"message,omitempty"`
	LogLevel *int    `json:"log_level,omitempty"`
	EOF      bool    `json:"EOF,omitempty"`
	// Event id
	EventID *int `json:"event_id,omitempty"`
	// State id
	LastKnownEventID *int `json:"last_known_event_id,omitempty"`
	// Extra data
	// Be carefully on putting pointers here or
	// race condition may occur
	Extra map[string]interface{} `json:"extra,omitempty"`
	// Manager *executive.Manager
	ExecutionUUID *string `json:"execution_uuid"`
	// Filtered feedback, sent only to client with this UUID
	ClientUUIDFilter *string `json:"-"`
	// Raw data
	Raw bool `json:"raw,omitempty"`
}

// BusConsumerLink struct.
// Used to connect consumer with BusData dispatcher
type BusConsumerLink struct {
	Name            string
	ClientUUID      string
	LogChan         chan *BusData
	CommonChan      chan *BusData
	AllowEventData  bool
	AllowStatusData bool
}

// SystemBus struct
type SystemBus struct {
	mu         sync.Mutex
	Executions map[string]bool
	// FEEDBACK
	connect    chan *BusConsumerLink
	disconnect chan *BusConsumerLink
	links      map[*BusConsumerLink]bool
	busBuffer  chan *BusData
	// Link (connected goroutines) control
	castWaiter *sync.WaitGroup
	// Pproviders
	providerInitFuncs map[string]base.ProviderInitFunc
}

// SetExecutionStatus func. Early status of execution (true/false).
// true indicates that the execution is running. False indicates
// that the execution is stopping or stopped: logs will be dicarded
func (s *SystemBus) SetExecutionStatus(eid string, status bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Executions[eid] = status
}

// ExistsExecution func
func (s *SystemBus) ExistsExecution(eid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Executions[eid]; ok {
		return true
	}
	return false
}

// GetExecutionStatus func
func (s *SystemBus) GetExecutionStatus(eid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if on, ok := s.Executions[eid]; ok && on {
		return true
	}
	return false
}

// Start func
func (s *SystemBus) Start() {

	for {
		// Remember how it works:
		// The select statement lets a goroutine wait on multiple
		// communication operations.

		// A select blocks until one of its cases can run, then it executes
		// that case. It chooses one at random if multiple are ready.

		// Note: Only the sender should close a channel, never the receiver.
		// Sending on a closed channel will cause a panic.

		// Another note: Channels aren't like files; you don't usually need
		// to close them. Closing is only necessary when the receiver must be
		// told there are no more values coming, such as to terminate a
		// range loop.
		select {
		case busConsumerLink := <-s.connect:
			s.links[busConsumerLink] = true
		case busConsumerLink := <-s.disconnect:
			delete(s.links, busConsumerLink)
		case busdata := <-s.busBuffer:
			// Calculate bus buffer load
			e := len(SBus.busBuffer)
			if e > 0 {
				load := (float64(e) / float64(SBusBufferSize)) * 100.0
				BInfo.SetLoad(load)
			}

			// Discard logs as needed
			if busdata.TypeID == BusDataTypeEvent && busdata.ExecutionUUID != nil && busdata.EventID != nil {
				switch *busdata.EventID {
				case EventManagerResuming, EventManagerStarted, EventManagerStarting:
					s.SetExecutionStatus(*busdata.ExecutionUUID, true)
				case EventManagerOut, EventManagerStopping:
					s.SetExecutionStatus(*busdata.ExecutionUUID, false)
				}
			}
			if (busdata.TypeID == BusDataTypeLog || busdata.TypeID == BusDataTypeStatus) && busdata.ExecutionUUID != nil {
				if s.ExistsExecution(*busdata.ExecutionUUID) && !s.GetExecutionStatus(*busdata.ExecutionUUID) {
					continue
				}
			}

			// Dispatch busdata to consumers
			for busConsumerLink := range s.links {
				// apply busdata filter
				if busdata.ClientUUIDFilter != nil && busConsumerLink.ClientUUID != *busdata.ClientUUIDFilter {
					continue
				}

				sent := false                         // always ensure data receipt
				if busdata.TypeID == BusDataTypeLog { // dispatch log data
					if !config.DEBUG && *busdata.LogLevel == DebugLevel {
						continue
					}
					for !sent {
						select {
						case busConsumerLink.LogChan <- busdata:
							sent = true
						default:
							// on dispatch fail, wait and retry
							time.Sleep(100 * time.Millisecond)
						}
					}
				} else { // dispatch non-log data
					// discard just-log links
					if busdata.TypeID == BusDataTypeEvent && !busConsumerLink.AllowEventData {
						continue
					}
					if busdata.TypeID == BusDataTypeStatus && !busConsumerLink.AllowStatusData {
						continue
					}
					for !sent {
						select {
						case busConsumerLink.CommonChan <- busdata:
							sent = true
						default:
							// on dispatch fail, wait and retry
							time.Sleep(100 * time.Millisecond)
						}
					}

				}
			}
		}
	}
}

// RegisterProviderInitFunc func
func (s *SystemBus) RegisterProviderInitFunc(strname string, initfunc base.ProviderInitFunc) {
	s.providerInitFuncs[strname] = initfunc
}

// GetProviderInitFunc func
func (s *SystemBus) GetProviderInitFunc(strname string) (base.ProviderInitFunc, error) {
	if _, exists := s.providerInitFuncs[strname]; exists {
		return s.providerInitFuncs[strname], nil
	}
	return nil, fmt.Errorf("Unkown Cloud Provider: " + strname)
}

// Close func
func (s *SystemBus) Close() *sync.WaitGroup {
	defer s.castWaiter.Done() // self
	bdata := &BusData{
		TypeID: BusDataTypeEOF,
	}
	PushBusData(bdata)
	return s.castWaiter
}

// Log func
func Log(level int, m *string, ei *string, ai *string, ti *string, raw bool) {
	// prevent debug messages on non-debug mode
	if !config.DEBUG && level == DebugLevel {
		return
	}
	bdata := &BusData{
		TypeID:   BusDataTypeLog,
		M:        m,
		LogLevel: &level,
		Raw:      raw,
	}
	if ei != nil {
		eei := *ei
		bdata.ExecutionUUID = &eei
	}
	if ai != nil {
		aai := *ai
		bdata.ActionID = &aai
	}
	if ti != nil {
		tti := *ti
		bdata.ThreadID = &tti
	}
	bdata.Timestamp = time.Now().UTC().UnixMicro()
	PushBusData(bdata)
}

// LogCritical func
func LogCritical(s string, re *string) {
	Log(CriticalLevel, &s, re, nil, nil, false)
}

// LogErr func
func LogErr(s string, re *string) {
	Log(ErrorLevel, &s, re, nil, nil, false)
}

// LogWarn func
func LogWarn(s string, re *string) {
	Log(WarningLevel, &s, re, nil, nil, false)
}

// LogInfo func
func LogInfo(s string, re *string) {
	Log(InfoLevel, &s, re, nil, nil, false)
}

// LogDebug func
func LogDebug(s string, re *string) {
	Log(DebugLevel, &s, re, nil, nil, false)
}

// SBusConnect func
func SBusConnect(fLink *BusConsumerLink) {
	SBus.connect <- fLink
}

// SBusDisconnect func
func SBusDisconnect(fLink *BusConsumerLink) {
	SBus.disconnect <- fLink
}

// PushBusData func
func PushBusData(bdata *BusData) {
L:
	for i := 0; i < 10; i++ {
		select {
		case SBus.busBuffer <- bdata:
			break L
		default:
			// this happends when s.busBuffer buffer is full :(
			// this is really bad :_(
			log.Printf("Problem pushing data to SBus buffer. Maybe Flood ocurr. Bus at %v%% of load. Retrying...\n", BInfo.GetLoad())
			time.Sleep(5000 * time.Millisecond)
		}
	}
}

// PushEvent func
func PushEvent(eid int, re *string) {
	var euuid string
	if re != nil {
		euuid = *re
	}
	bdata := &BusData{
		TypeID:        BusDataTypeEvent,
		EventID:       &eid,
		ExecutionUUID: &euuid,
		Timestamp:     time.Now().UTC().UnixMicro(),
	}
	PushBusData(bdata)
}

// PushEvent func
func PushEventWithExtra(eid int, ei *string, extra map[string]interface{}) {
	bdata := &BusData{
		TypeID:  BusDataTypeEvent,
		EventID: &eid,
		Extra:   extra,
	}
	if ei != nil {
		eei := *ei
		bdata.ExecutionUUID = &eei
	}
	_, ok := extra["action_id"]
	if ok {
		actionid := extra["action_id"].(string)
		bdata.ActionID = &actionid
	}
	bdata.Timestamp = time.Now().UTC().UnixMicro()
	PushBusData(bdata)
}

// PushState func
func PushState(runningIDs []string, state int, ei *string) {
	var s = state
	bdata := &BusData{
		TypeID:           BusDataTypeStatus,
		LastKnownEventID: &s,
	}
	if ei != nil {
		eei := *ei
		bdata.ExecutionUUID = &eei
	}
	bdata.Extra = make(map[string]interface{})
	bdata.Extra["uuids_in_progress"] = runningIDs
	bdata.Timestamp = time.Now().UTC().UnixMicro()
	PushBusData(bdata)
}

// PushFilteredBusData func
func PushFilteredBusData(clientUUIDFilter string, extra map[string]interface{}) {
	bdata := &BusData{
		TypeID:           BusDataTypeEvent,
		ClientUUIDFilter: &clientUUIDFilter,
		Extra:            extra,
		Timestamp:        time.Now().UTC().UnixMicro(),
	}
	PushBusData(bdata)
}

// Logger struct
type Logger struct {
	ExecutionUUID *string
	ActionID      *string
	ThreadID      *string
}

// Duplicate func
func (l *Logger) Duplicate() base.ILogger {
	newLoger := &Logger{}
	if l.ExecutionUUID != nil {
		ei := *l.ExecutionUUID
		newLoger.ExecutionUUID = &ei
	}
	if l.ActionID != nil {
		ai := *l.ActionID
		newLoger.ActionID = &ai
	}
	if l.ThreadID != nil {
		ti := *l.ThreadID
		newLoger.ThreadID = &ti
	}
	return newLoger
}

// SetActionID func
func (l *Logger) SetActionID(ai string) {
	l.ActionID = &ai
}

// SetThreadID func
func (l *Logger) SetThreadID(ti string) {
	l.ThreadID = &ti
}

// LogCritical func
func (l *Logger) LogCritical(s string) {
	Log(CriticalLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, false)
}

// LogErr func
func (l *Logger) LogErr(s string) {
	Log(ErrorLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, false)
}

// ByteLogErr func
func (l *Logger) ByteLogErr(b []byte) {
	// last chance to determine encoding
	s := string(b)
	Log(ErrorLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, true)
}

// LogWarn func
func (l *Logger) LogWarn(s string) {
	Log(WarningLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, false)
}

// LogInfo func
func (l *Logger) LogInfo(s string) {
	Log(InfoLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, false)
}

// ByteLogInfo func
func (l *Logger) ByteLogInfo(b []byte) {
	// last chance to determine encoding
	s := string(b)
	Log(InfoLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, true)
}

// LogDebug func
func (l *Logger) LogDebug(s string) {
	Log(DebugLevel, &s, l.ExecutionUUID, l.ActionID, l.ThreadID, false)
}

type DummyLogger struct {
	ExecutionUUID *string
	ActionID      *string
	ThreadID      *string
}

// Duplicate func
func (l *DummyLogger) Duplicate() base.ILogger {
	newLoger := &DummyLogger{}
	if l.ExecutionUUID != nil {
		ei := *l.ExecutionUUID
		newLoger.ExecutionUUID = &ei
	}
	if l.ActionID != nil {
		ai := *l.ActionID
		newLoger.ActionID = &ai
	}
	if l.ThreadID != nil {
		ti := *l.ThreadID
		newLoger.ThreadID = &ti
	}
	return newLoger
}
func (l *DummyLogger) SetActionID(ai string) {
	l.ActionID = &ai
}
func (l *DummyLogger) SetThreadID(ti string) {
	l.ThreadID = &ti
}
func (l *DummyLogger) LogCritical(s string) {}
func (l *DummyLogger) LogErr(s string)      {}
func (l *DummyLogger) ByteLogErr(b []byte)  {}
func (l *DummyLogger) LogWarn(s string)     {}
func (l *DummyLogger) LogInfo(s string)     {}
func (l *DummyLogger) ByteLogInfo(b []byte) {}
func (l *DummyLogger) LogDebug(s string)    {}

// BusInfo struct
type BusInfo struct {
	mu   sync.Mutex
	Load float64
}

func (b *BusInfo) SetLoad(l float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Load = l
}

func (b *BusInfo) GetLoad() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Load
}
