// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

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
	events  chan IEvent
	discard chan IEvent
	// when ReadUntil or WaitUntil are
	// called. If not, this reading is
	// false and all events sended to
	// this listener will be discarded
	reading bool
}

func (e *EventListener) EventChan() chan IEvent {
	if !e.reading {
		go func() {
			<-e.discard
		}()
		return e.discard
	}
	return e.events
}

// Len: returns the count of events
// awaiting to be readed
// func (e *EventListener) Len() int {
// 	return len(e.events)
// }

// Read all events until EventCode are found, return
// true if EventCode gets found. Return false if events
// chan gets empty without any ocurrence of EventCode
func (e *EventListener) ReadUntil(ec EventCode) bool {
	if e.reading {
		panic("hey dev, this is your fault, never call listener two times!")
	}
	e.reading = true
	defer func() { e.reading = false }()
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
	if e.reading {
		panic("hey dev, this is your fault, never call listener two times!")
	}
	e.reading = true
	defer func() { e.reading = false }()
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
		events:  make(chan IEvent, 10),
		discard: make(chan IEvent, 10),
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
