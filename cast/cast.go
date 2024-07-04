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

package cast

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/config"
	"github.com/google/uuid"
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

// DebugLevel const
const ParanoicDebugLevel = 5

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
	// EventRuntimeStarting const 7
	EventRuntimeStarting
	// EventRuntimeResuming const 8
	EventRuntimeResuming
	// EventRuntimeStarted const 9
	EventRuntimeStarted
	// EventRuntimePausing const 10
	EventRuntimePausing
	// EventManagerPause const 11
	EventRuntimePaused
	// EventRuntimeStopping const 12
	EventRuntimeStopping
	// EventRuntimeOut const 13
	EventRuntimeOut
	// EventRegisteredManager const 14
	EventRegisteredManager
	// EventWaitingStatus const 15
	EventWaitingForState
	// EventActionUnCaughtKO const 16
	EventActionUnCaughtKO
	// EventActionUnCaughtOK const 17
	EventActionUnCaughtOK
	// EventActionInit const 18
	EventActionInit
	// EventActionKO Handled KO const 19
	EventActionKO
	// EventActionOK const 20
	EventActionOK
	// ThreadCreated const 21
	EventNewThread
	// ThreadDestroyed const 22
	EventThreadDestroyed
	//
	EventProgressStart
	EventProgressTick
	EventProgressEnd
	EventInteractiveMenuStart
	EventPrompt
	EventPromptDone
)

// EventPromptType int
type EventPromptType int

const (
	EventPromptTypeInput EventPromptType = iota
	EventPromptTypeSelect
	EventPromptTypeBool
)

type EventPromptOptsValidateOpts struct {
	ValueType  string // int, string ...
	AllowNull  bool
	AllowEmpty bool
}

type EventPromptOpts struct {
	UUID         string            `json:"uuid"`
	Type         EventPromptType   `json:"type"`
	PromptTitle  string            `json:"prompt_title"`
	DefaultValue string            `json:"default_value"`
	Options      map[string]string `json:"options"` // for select
	// for responses
	value chan string
	//
	Validate *EventPromptOptsValidateOpts `json:"validate"`
}

// BusData struct
type BusData struct {
	Timestamp int64 `json:"timestamp"`
	//
	// Type of data
	TypeID BusDataType `json:"type_id"`
	// id of the action sending the data
	ActionID *string `json:"action_id,omitempty"`
	// name of the action sending the data
	ActionName *string `json:"action_name,omitempty"`
	// id of the thread sending the data
	ThreadID *string `json:"thread_id,omitempty"`
	// Msg data in bytes
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
	Raw bool             `json:"raw,omitempty"`
	EPO *EventPromptOpts `json:"epo,omitempty"`
}

// BusConsumerLink struct.
// Used to connect consumer with BusData dispatcher
type BusConsumerLink struct {
	Name       string
	ClientUUID string
	LogChan    chan *BusData
	CommonChan chan *BusData
	// tells a busreader that should stop read. This
	// is usefull to switch readers using the same
	// link
	Off chan struct{}
	// EOF msg received, so this link should not used
	// to read any more logs
	Degraded        bool
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
				case EventRuntimeResuming, EventRuntimeStarted, EventRuntimeStarting:
					s.SetExecutionStatus(*busdata.ExecutionUUID, true)
				case EventRuntimeOut, EventRuntimeStopping:
					s.SetExecutionStatus(*busdata.ExecutionUUID, false)
				}
			}
			if (busdata.TypeID == BusDataTypeLog || busdata.TypeID == BusDataTypeStatus) && busdata.ExecutionUUID != nil {
				// WIP: el filtro de get execution status debería existir?
				// ocurre un problema que no mola nada: cuando hay un panic
				// el execution se pone a false y los mensajes con este
				// execution uuid se mandan a parla
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

// WIP: esto lo mismo podríamos moverlo a runtime
// RegisterProviderInitFunc func
func (s *SystemBus) RegisterProviderInitFunc(strname string, initfunc base.ProviderInitFunc) {
	s.providerInitFuncs[strname] = initfunc
}

// WIP: esto lo mismo podríamos moverlo a runtime
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
	_log(_logmsg{
		level: level,
		m:     m,
		ei:    ei,
		ai:    ai,
		ti:    ti,
		raw:   raw,
	})
}

type _logmsg struct {
	level int     // level
	m     *string // msg
	ei    *string // execution id
	ai    *string // action id
	an    *string // action name
	ti    *string // thread id
	raw   bool    // raw or not
}

func _log(_l _logmsg) {
	// prevent debug messages on non-debug mode
	if !config.DEBUG && _l.level == DebugLevel {
		return
	}
	if !config.PARANOICDEBUG && _l.level == ParanoicDebugLevel {
		return
	}

	bdata := &BusData{
		TypeID:   BusDataTypeLog,
		M:        _l.m,
		LogLevel: &_l.level,
		Raw:      _l.raw,
	}
	if _l.ei != nil {
		eei := *_l.ei
		bdata.ExecutionUUID = &eei
	}
	if _l.ai != nil {
		aai := *_l.ai
		bdata.ActionID = &aai
	}
	if _l.an != nil {
		aan := *_l.an
		bdata.ActionName = &aan
	}
	if _l.ti != nil {
		tti := *_l.ti
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

// PushMixedLogEventBusData send a two copies of bdata
// to bus buffer overriding bdata.TypeID and setting it
// to cast.BusDataTypeEvent and cast.BusDataTypeLog.
// This is usefull on init/end of action or something
// to fine-log init events of an action before the
// output of that action.
func PushMixedLogEventBusData(bdata *BusData) {
	bdata.TypeID = BusDataTypeEvent
	PushBusData(bdata)
	bbdata := BusData(*bdata)
	bbdata.TypeID = BusDataTypeLog
	PushBusData(&bbdata)
}

func EP(e int) *int {
	return &e
}

func SEP(s string) *string {
	return &s
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

// TODO: look for a better way to do this
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

// WIP: esto debería ser un sistema de eventos: crear un struct
// tipo Event aquí en cast y en lugar de usar Push, usar algo así
// como dispatchEvent(e *Event)
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
	ActionName    *string
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

// SetActionName func
func (l *Logger) SetActionName(an string) {
	l.ActionName = &an
}

// SetThreadID func
func (l *Logger) SetThreadID(ti string) {
	l.ThreadID = &ti
}

// LogCritical func
func (l *Logger) LogCritical(s string) {
	_log(_logmsg{
		level: CriticalLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   false,
	})
}

// LogErr func
func (l *Logger) LogErr(s string) {
	_log(_logmsg{
		level: ErrorLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   false,
	})
}

// ByteLogErr func
func (l *Logger) ByteLogErr(b []byte) {
	// last chance to determine encoding
	s := string(b)
	_log(_logmsg{
		level: ErrorLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   true,
	})
}

// LogWarn func
func (l *Logger) LogWarn(s string) {
	_log(_logmsg{
		level: WarningLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   false,
	})
}

// LogInfo func
func (l *Logger) LogInfo(s string) {
	_log(_logmsg{
		level: InfoLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   false,
	})
}

// ByteLogInfo func
func (l *Logger) ByteLogInfo(b []byte) {
	// last chance to determine encoding
	s := string(b)
	_log(_logmsg{
		level: InfoLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   true,
	})
}

// LogDebug func
func (l *Logger) LogDebug(s string) {
	_log(_logmsg{
		level: DebugLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   false,
	})
}

// ParanoicLogDebug func
func (l *Logger) ParanoicLogDebug(s string) {
	_log(_logmsg{
		level: ParanoicDebugLevel,
		m:     &s,
		ei:    l.ExecutionUUID,
		ai:    l.ActionID,
		an:    l.ActionName,
		ti:    l.ThreadID,
		raw:   false,
	})
}

type DummyLogger struct {
	ExecutionUUID *string
	ActionID      *string
	ActionName    *string
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
func (l *DummyLogger) SetActionName(an string) {
	l.ActionName = &an
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

func PromptInput(title string, def string) (chan string, error) {
	uid7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	bdata := &BusData{
		TypeID:    BusDataTypeEvent,
		EventID:   EP(EventPrompt),
		Timestamp: time.Now().UTC().UnixMicro(),
		EPO: &EventPromptOpts{
			UUID:         uid7.String(),
			Type:         EventPromptTypeInput,
			PromptTitle:  title,
			DefaultValue: def,
			value:        make(chan string, 1),
			Validate: &EventPromptOptsValidateOpts{
				ValueType:  "string",
				AllowNull:  false,
				AllowEmpty: false,
			},
		},
	}
	PushBusData(bdata)
	return bdata.EPO.value, nil
}

func PromptInt(title string, def string) (chan string, error) {
	uid7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	bdata := &BusData{
		TypeID:    BusDataTypeEvent,
		EventID:   EP(EventPrompt),
		Timestamp: time.Now().UTC().UnixMicro(),
		EPO: &EventPromptOpts{
			UUID:         uid7.String(),
			Type:         EventPromptTypeInput,
			PromptTitle:  title,
			DefaultValue: def,
			value:        make(chan string, 1),
			Validate: &EventPromptOptsValidateOpts{
				ValueType:  "int",
				AllowNull:  false,
				AllowEmpty: false,
			},
		},
	}
	PushBusData(bdata)
	return bdata.EPO.value, nil
}

func PromptSelect(title string, options map[string]string) (chan string, error) {
	uid7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	bdata := &BusData{
		TypeID:    BusDataTypeEvent,
		EventID:   EP(EventPrompt),
		Timestamp: time.Now().UTC().UnixMicro(),
		EPO: &EventPromptOpts{
			UUID:        uid7.String(),
			Type:        EventPromptTypeSelect,
			PromptTitle: title,
			Options:     options,
			value:       make(chan string, 1),
			Validate: &EventPromptOptsValidateOpts{
				ValueType:  "string",
				AllowNull:  false,
				AllowEmpty: false,
			},
		},
	}
	PushBusData(bdata)
	return bdata.EPO.value, nil
}

func PromptBool(title string) (chan string, error) {
	uid7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	bdata := &BusData{
		TypeID:    BusDataTypeEvent,
		EventID:   EP(EventPrompt),
		Timestamp: time.Now().UTC().UnixMicro(),
		EPO: &EventPromptOpts{
			UUID:        uid7.String(),
			Type:        EventPromptTypeBool,
			PromptTitle: title,
			value:       make(chan string, 1),
			Validate: &EventPromptOptsValidateOpts{
				ValueType:  "bool",
				AllowNull:  false,
				AllowEmpty: false,
			},
		},
	}
	PushBusData(bdata)
	return bdata.EPO.value, nil
}

func AnswerPrompt(b *BusData, v string) {
	bdata := &BusData{
		TypeID:    BusDataTypeEvent,
		EventID:   EP(EventPromptDone),
		Timestamp: time.Now().UTC().UnixMicro(),
		EPO: &EventPromptOpts{
			UUID: b.EPO.UUID,
		},
	}
	b.EPO.value <- v
	PushBusData(bdata)
}
