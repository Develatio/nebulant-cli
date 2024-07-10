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
	"time"

	"github.com/develatio/nebulant-cli/base"
)

func NewProgress(size int64, info, actionid, actionname, threadid, euuid string) *Progress {
	p := &Progress{
		size:       size,
		info:       info,
		actionid:   &actionid,
		actionname: &actionname,
		threadid:   &threadid,
		euuid:      &euuid,
	}
	p.id = fmt.Sprintf("%p", p)
	PushMixedLogEventBusData(&BusData{
		EventID:       EP(EventProgressStart),
		ActionID:      p.actionid,
		ActionName:    p.actionname,
		LogLevel:      EP(base.InfoLevel),
		ThreadID:      p.threadid,
		ExecutionUUID: p.euuid,
		Timestamp:     time.Now().UTC().UnixMicro(),
		Extra: map[string]interface{}{
			"progressid": p.id,
			"size":       p.size,
			"writed":     p.writed,
			"info":       p.info,
		},
	})
	return p
}

type Progress struct {
	id         string
	info       string
	size       int64
	writed     int64
	euuid      *string
	actionid   *string
	threadid   *string
	actionname *string
}

func (g *Progress) Add64(n int64) {
	g.writed = g.writed + int64(n)
	if g.writed == g.size {
		PushMixedLogEventBusData(&BusData{
			EventID:       EP(EventProgressEnd),
			ActionID:      g.actionid,
			ActionName:    g.actionname,
			LogLevel:      EP(base.InfoLevel),
			ThreadID:      g.threadid,
			ExecutionUUID: g.euuid,
			Timestamp:     time.Now().UTC().UnixMicro(),
			Extra: map[string]interface{}{
				"progressid": g.id,
				"size":       g.size,
				"writed":     g.writed,
				"info":       g.info,
			},
		})
		return
	}
	PushMixedLogEventBusData(&BusData{
		EventID:       EP(EventProgressTick),
		ActionID:      g.actionid,
		ActionName:    g.actionname,
		LogLevel:      EP(base.InfoLevel),
		ThreadID:      g.threadid,
		ExecutionUUID: g.euuid,
		Timestamp:     time.Now().UTC().UnixMicro(),
		Extra: map[string]interface{}{
			"progressid": g.id,
			"size":       g.size,
			"writed":     g.writed,
			"info":       g.info,
		},
	})
}

func (g *Progress) Add(n int) {
	g.Add64(int64(n))
}

func (g *Progress) Write(p []byte) (int, error) {
	g.writed = g.writed + int64(len(p))
	if g.writed == g.size {
		PushMixedLogEventBusData(&BusData{
			EventID:       EP(EventProgressEnd),
			ActionID:      g.actionid,
			ActionName:    g.actionname,
			LogLevel:      EP(base.InfoLevel),
			ThreadID:      g.threadid,
			ExecutionUUID: g.euuid,
			Timestamp:     time.Now().UTC().UnixMicro(),
			Extra: map[string]interface{}{
				"progressid": g.id,
				"size":       g.size,
				"writed":     g.writed,
				"info":       g.info,
			},
		})
		return len(p), nil
	}
	PushMixedLogEventBusData(&BusData{
		EventID:       EP(EventProgressTick),
		ActionID:      g.actionid,
		ActionName:    g.actionname,
		LogLevel:      EP(base.InfoLevel),
		ThreadID:      g.threadid,
		ExecutionUUID: g.euuid,
		Timestamp:     time.Now().UTC().UnixMicro(),
		Extra: map[string]interface{}{
			"progressid": g.id,
			"size":       g.size,
			"writed":     g.writed,
			"info":       g.info,
		},
	})
	return len(p), nil
}
