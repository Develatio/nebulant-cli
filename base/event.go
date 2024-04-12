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

package base

import "sync"

type EventCode int

const AnyEvent EventCode = -1

const (
	RuntimePlayEvent EventCode = iota
	RuntimeStillEvent
	RuntimeEndEvent
	//
	BreakPointEvent
	DebugOnEvent
	DebugOffEvent
	DebugStepEvent
)

type IEvent interface {
	EventCode() EventCode
	String() string
}

type EventListener struct {
	events chan IEvent
}

func (e *EventListener) EventChan() chan IEvent {
	return e.events
}

// Len: returns the count of events
// awaiting to be readed
func (e *EventListener) Len() int {
	return len(e.events)
}

// Read all events until EventCode are found, return
// true if EventCode gets found. Return false if events
// chan gets empty without any ocurrence of EventCode
func (e *EventListener) ReadUntil(ec EventCode) bool {
	for {
		select {
		case evt := <-e.events:
			if evt.EventCode() == ec {
				return true
			}
			continue
		default:
			return false
		}
	}
}

// Waits for ocurrence of any of given EventCode, returns
// the first EventCode found
func (e *EventListener) WaitUntil(ecs []EventCode) EventCode {
	for {
		evt := <-e.events
		for _, ec := range ecs {
			if evt.EventCode() == ec {
				return ec
			}
		}
	}
}

func NewEventListener() *EventListener {
	return &EventListener{
		events: make(chan IEvent, 10),
	}
}

type EventDispatcher struct {
	mu             sync.Mutex
	eventListeners map[*EventListener]bool
}

func (ev *EventDispatcher) NewEventListener() *EventListener {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	el := NewEventListener()
	ev.eventListeners[el] = true
	return el
}

func (ev *EventDispatcher) DestroyEventListener(e *EventListener) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	delete(ev.eventListeners, e)
}

func (ev *EventDispatcher) Dispatch(e IEvent) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	for el := range ev.eventListeners {
		select {
		case el.EventChan() <- e:
			// ok
		default:
			// also ok
			// cast.LogDebug("event listener full, skip", r.irb.ExecutionUUID)
		}
	}
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{eventListeners: make(map[*EventListener]bool)}
}
