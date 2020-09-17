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

// SBus is a globally shared system bus
var SBus *SystemBus

// InitSystemBus func
func InitSystemBus() {
	castWaiter := &sync.WaitGroup{}
	castWaiter.Add(1) // self

	SBus = &SystemBus{
		// feedback
		connect:          make(chan *FeedBackLink),
		disconnect:       make(chan *FeedBackLink),
		dispatchFeedBack: make(chan *FeedBack, 100000),
		links:            make(map[*FeedBackLink]bool),
		// waiter
		castWaiter: castWaiter,
		// providers
		providerInitFuncs: make(map[string]base.ProviderInitFunc),
	}
	go SBus.Start()
}

// FeedBackType int
type FeedBackType int

const (
	// FeedBackLog const 0
	FeedBackLog FeedBackType = iota
	// FeedBackEvent const 1
	FeedBackEvent
	// FeedBackStatus const 2
	FeedBackStatus
	// FeedBackFiltered const used in conjuction with ClientUUIDFilter to send
	// messages only to one log/event listener: commonly the websocket logger
	// to command to join into uuid channel using Extra["join"]=uuid
	FeedBackFiltered
	// FeedBackEOF const 4
	FeedBackEOF
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
)

// FeedBack struct
type FeedBack struct {
	// Type of feedback
	TypeID FeedBackType `json:"type_id"`
	// Msg data in bytes
	B        []byte `json:"log_bytes"`
	LogLevel *int   `json:"log_level"`
	EOF      bool
	// Event id
	EventID *int `json:"event_id"`
	// State id
	LastKnownEventID *int `json:"last_known_event_id"`
	// Extra data
	Extra map[string]interface{} `json:"extra"`
	// Manager *executive.Manager
	ExecutionUUID *string `json:"execution_uuid"`
	// Filtered feedback, sent only to client with this UUID
	ClientUUIDFilter *string
	// Raw data
	Raw bool `json:"raw"`
}

// FeedBackLink struct.
// Used to connect consumer with FeedBack dispatcher
type FeedBackLink struct {
	ClientUUID  string
	FeedBackBus chan *FeedBack
}

// SystemBus struct
type SystemBus struct {
	// FEEDBACK
	connect          chan *FeedBackLink
	disconnect       chan *FeedBackLink
	links            map[*FeedBackLink]bool
	dispatchFeedBack chan *FeedBack
	// Link (connected goroutines) control
	castWaiter *sync.WaitGroup
	// Pproviders
	providerInitFuncs map[string]base.ProviderInitFunc
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
		case fLink := <-s.connect:
			s.links[fLink] = true
		case fLink := <-s.disconnect:
			delete(s.links, fLink)
		case fback := <-s.dispatchFeedBack:
			for fLink := range s.links {
				if fback.ClientUUIDFilter != nil && fLink.ClientUUID != *fback.ClientUUIDFilter {
					continue
				}
				select {
				case fLink.FeedBackBus <- fback:
				default:
					// Hi developer! :)
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
	fback := &FeedBack{
		TypeID: FeedBackEOF,
	}
	PublishFeedBack(fback)
	return s.castWaiter
}

// Log func
func Log(level int, b []byte, re *string, raw bool) {
	fback := &FeedBack{
		TypeID:        FeedBackLog,
		B:             b,
		LogLevel:      &level,
		ExecutionUUID: re,
		Raw:           raw,
	}
	PublishFeedBack(fback)
}

// LogCritical func
func LogCritical(s string, re *string) {
	Log(CriticalLevel, []byte(s), re, false)
}

// LogErr func
func LogErr(s string, re *string) {
	Log(ErrorLevel, []byte(s), re, false)
}

// LogWarn func
func LogWarn(s string, re *string) {
	Log(WarningLevel, []byte(s), re, false)
}

// LogInfo func
func LogInfo(s string, re *string) {
	Log(InfoLevel, []byte(s), re, false)
}

// LogDebug func
func LogDebug(s string, re *string) {
	Log(DebugLevel, []byte(s), re, false)
}

// SBusConnect func
func SBusConnect(fLink *FeedBackLink) {
	SBus.connect <- fLink
}

// SBusDisconnect func
func SBusDisconnect(fLink *FeedBackLink) {
	SBus.disconnect <- fLink
}

// PublishFeedBack func
func PublishFeedBack(fback *FeedBack) {
L:
	for i := 0; i < 10; i++ {
		select {
		case SBus.dispatchFeedBack <- fback:
			break L
		default:
			// this happends when s.dispatchFeedBack buffer is full
			log.Println("Problem publishing feedback data across SBus. Maybe Flood ocurr. Retrying...")
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// PublishEvent func
func PublishEvent(eid int, re *string) {
	var euuid string
	if re != nil {
		euuid = *re
	}
	fback := &FeedBack{
		TypeID:        FeedBackEvent,
		EventID:       &eid,
		ExecutionUUID: &euuid,
	}
	PublishFeedBack(fback)
}

// PublishEvent func
func PublishEventWithExtra(eid int, re *string, extra map[string]interface{}) {
	fback := &FeedBack{
		TypeID:        FeedBackEvent,
		EventID:       &eid,
		ExecutionUUID: re,
		Extra:         extra,
	}
	PublishFeedBack(fback)
}

// PublishState func
func PublishState(runningIDs []string, state int, re *string) {
	var s = state
	fback := &FeedBack{
		TypeID:           FeedBackStatus,
		LastKnownEventID: &s,
		ExecutionUUID:    re,
	}
	fback.Extra = make(map[string]interface{})
	fback.Extra["uuids_in_progress"] = runningIDs
	PublishFeedBack(fback)
}

// PublishFiltered func
func PublishFiltered(clientUUIDFilter string, extra map[string]interface{}) {
	fback := &FeedBack{
		TypeID:           FeedBackFiltered,
		ClientUUIDFilter: &clientUUIDFilter,
		Extra:            extra,
	}
	PublishFeedBack(fback)
}

// Logger struct
type Logger struct {
	ExecutionUUID *string
}

// LogCritical func
func (l *Logger) LogCritical(s string) {
	Log(CriticalLevel, []byte(s), l.ExecutionUUID, false)
}

// LogErr func
func (l *Logger) LogErr(s string) {
	Log(ErrorLevel, []byte(s), l.ExecutionUUID, false)
}

// ByteLogErr func
func (l *Logger) ByteLogErr(b []byte) {
	Log(ErrorLevel, b, l.ExecutionUUID, true)
}

// LogWarn func
func (l *Logger) LogWarn(s string) {
	Log(WarningLevel, []byte(s), l.ExecutionUUID, false)
}

// LogInfo func
func (l *Logger) LogInfo(s string) {
	Log(InfoLevel, []byte(s), l.ExecutionUUID, false)
}

// ByteLogInfo func
func (l *Logger) ByteLogInfo(b []byte) {
	Log(InfoLevel, b, l.ExecutionUUID, true)
}

// LogDebug func
func (l *Logger) LogDebug(s string) {
	Log(DebugLevel, []byte(s), l.ExecutionUUID, false)
}
