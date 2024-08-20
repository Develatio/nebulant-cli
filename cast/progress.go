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
	"io"
	"math"
	"time"

	"github.com/develatio/nebulant-cli/base"
)

type ProgressConf struct {
	Size       int64
	Info       string
	ActionId   string
	ActionName string
	ThreadId   string
	EuuId      string
	Autoend    bool
}

func NewProgress(cnf *ProgressConf) *Progress {
	// func NewProgress(size int64, info, actionid, actionname, threadid, euuid string, autoend bool) *Progress {
	p := &Progress{
		size:       cnf.Size,
		info:       cnf.Info,
		actionid:   &cnf.ActionId,
		actionname: &cnf.ActionName,
		threadid:   &cnf.ThreadId,
		euuid:      &cnf.EuuId,
		autoend:    cnf.Autoend,
		start:      time.Now(),
		lastUpdate: time.Now(),
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
	autoend    bool
	start      time.Time
	lastUpdate time.Time
	err        error
	speedRate  float64 // bytes/s
	rwrap      io.Reader
}

func (g *Progress) Add64(n int64) {
	g.writed = g.writed + int64(n)
	g.Tick()
}

func (g *Progress) Add(n int) {
	g.Add64(int64(n))
}

func (g *Progress) refreshRate(total bool) {

	now := time.Now()
	var elapsed time.Duration
	if total {
		elapsed = now.Sub(g.start)
	} else {
		elapsed = now.Sub(g.lastUpdate)
		if elapsed.Milliseconds() < 300 {
			return
		}
	}

	g.speedRate = float64(g.writed) / elapsed.Seconds()

	g.lastUpdate = time.Now()
}

func (g *Progress) Rate() float64 {
	return g.speedRate
}

func (g *Progress) FormattedRate() string {
	bysec := g.speedRate * 100
	u := byte(' ')

	if bysec < 100*1000 {
		u = 'K'
		bysec = (bysec + 512) / 1024
	} else {
		unit := " KMGT"
		i := 0
		for ; bysec >= 100*1000 && unit[i] != byte('T'); i++ {
			bysec = (bysec + 512) / 1024
		}
		u = unit[i]
	}

	n := (bysec + 5) / 100
	r := math.Mod((bysec+5)/10, 10)
	return fmt.Sprintf("%3d.%1d%sB/s", int64(n), int64(r), string(u))
}

func (g *Progress) Tick() {
	g.refreshRate(false)
	if g.writed == g.size && g.autoend {
		g.End()
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
			"frate":      g.FormattedRate(),
		},
	})
}

func (g *Progress) End() {
	g.refreshRate(true)
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
			"error":      g.err,
			"frate":      g.FormattedRate(),
		},
	})
}

func (g *Progress) Write(p []byte) (int, error) {
	g.writed = g.writed + int64(len(p))
	g.Tick()
	return len(p), nil
}

func (g *Progress) Set(n int64) {
	g.writed = n
	g.Tick()
}

func (g *Progress) SetErr(err error) {
	g.err = err
}

func (g *Progress) WrapRead(r io.Reader) {
	g.rwrap = r
}

func (g *Progress) Read(p []byte) (n int, err error) {
	n, err = g.rwrap.Read(p)
	g.Add(n)
	return n, err
}
