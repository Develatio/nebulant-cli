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
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/term"
	"github.com/schollz/progressbar/v3"
)

var counter int = 0
var span int = 10

var prefxmap map[int]string = map[int]string{
	base.CriticalLevel: "!CRITICAL ",
	base.ErrorLevel:    "!ERROR ",
	base.WarningLevel:  "!WARNING ",
	base.InfoLevel:     "",
	base.DebugLevel:    "DEBUG ",
	base.NotsetLevel:   "",
}

var threadcolor map[string]string

var progress map[string]*progressInfo

type noreturnFilterWriter struct{}

func (a noreturnFilterWriter) Write(p []byte) (int, error) {
	var pp []byte
	var nn = []byte("\n")
	var rr = []byte("\r")
	var n = len(p)
	pp = append(pp, p...)

	pp = bytes.TrimPrefix(pp, rr)
	pp = bytes.TrimSuffix(pp, rr)

	pp = bytes.TrimSpace(pp)
	if len(pp) <= 0 {
		return n, nil
	}

	pp = append(pp, nn...)

	s, err := os.Stdout.Write(pp)
	if err != nil {
		return s, err
	}
	return n, nil
}

type progressInfo struct {
	info     string
	size     int64
	writed   int64
	progress *progressbar.ProgressBar
}

func (p *progressInfo) Update(w int64) error {
	add := w - p.writed
	if add <= 0 {
		return nil
	}
	p.writed = w
	return p.progress.Add64(add)
}

func (p *progressInfo) Finish() error {
	return p.progress.Finish()
}

func newProgress(size int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(size,
		progressbar.OptionEnableColorCodes(false),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(description),
		progressbar.OptionUseANSICodes(false),
		progressbar.OptionSetWriter(&noreturnFilterWriter{}),
	)
}

// ConsoleLogger struct
type ConsoleLogger struct {
	logLevelFilter int
	fLink          *BusConsumerLink
}

func p(s string) *string {
	return &s
}

func FormatConsoleLogMsg(fback *BusData, verbose bool) *string {
	color := ""
	if fback.ThreadID != nil {
		if _, exists := threadcolor[*fback.ThreadID]; !exists {
			threadcolor[*fback.ThreadID] = term.GetNewColor()
		}
		color = threadcolor[*fback.ThreadID]
	}

	format := "%s %s"
	aname := ""
	if fback.ActionName != nil {
		aname = *fback.ActionName
		space := span - len(aname)
		if space <= 1 {
			span = len(aname) + 2
			space = 2
		}
		format = fmt.Sprintf("%s %s%s|%s %%s %%s", color, aname, strings.Repeat(" ", space), term.Reset)
	}

	if fback.EventID != nil {
		switch *fback.EventID {
		case EventActionInit:
			return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], "start"))
		case EventActionKO:
			return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], *fback.M))
		case EventActionUnCaughtKO:
			return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], *fback.M))
		case EventActionOK:
			return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], "done"))
		case EventActionUnCaughtOK:
			return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], "done"))
		case EventThreadDestroyed:
			delete(threadcolor, *fback.ThreadID)
		default:
			return nil
		}
	}

	if fback.Raw {
		if fback.M != nil {
			return p(term.Reset + *fback.M + term.Reset)
		} else {
			return nil
		}
	} else {
		if fback.M != nil {
			if verbose {
				counter++
				return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], fmt.Sprintf("[%v] %v", counter, *fback.M)))
			} else {
				return p(fmt.Sprintf(format, prefxmap[*fback.LogLevel], *fback.M))
			}
		} else {
			return nil
		}
	}
}

func (c *ConsoleLogger) _print(raw bool, msg string) {
	if raw {
		var err error
		_, err = term.Print(msg)
		if err != nil {
			// fback err?
		}
	} else {
		log.Println(msg)
	}
}

func (c *ConsoleLogger) printMessage(fback *BusData) bool {
	switch fback.TypeID {
	case BusDataTypeLog:
		if fback.LogLevel != nil && *fback.LogLevel >= c.logLevelFilter {
			s := FormatConsoleLogMsg(fback, c.logLevelFilter <= base.DebugLevel)
			if s != nil {
				c._print(fback.Raw, *s)
			}
		}
	case BusDataTypeEvent:
		// a nil pointer err here is a dev fail
		// never send event type without event id
		switch *fback.EventID {
		case EventNewThread,
			EventActionInit,
			EventActionKO,
			EventActionUnCaughtKO,
			EventActionOK,
			EventActionUnCaughtOK,
			EventThreadDestroyed:
			s := FormatConsoleLogMsg(fback, c.logLevelFilter <= base.DebugLevel)
			if s != nil {
				c._print(fback.Raw, *s)
			}
		case EventProgressStart:
			pid := fback.Extra["progressid"].(string)
			size := fback.Extra["size"].(int64)
			info := fback.Extra["info"].(string)
			progress[pid] = &progressInfo{
				size:     size,
				info:     info,
				progress: newProgress(size, info),
			}
		case EventProgressTick:
			writed := fback.Extra["writed"].(int64)
			pid := fback.Extra["progressid"].(string)
			pinfo := progress[pid]
			pinfo.Update(writed)
		case EventProgressEnd:
			// TODO: add cmd?
			pid := fback.Extra["progressid"].(string)
			npr := progress[pid]
			npr.Finish()
			delete(progress, pid)
			c._print(true, "\n")
		}
	}
	return true

}

func (c *ConsoleLogger) readCastBus() {
	defer SBus.castWaiter.Done()
	power := true
L:
	for {
		if !power && len(c.fLink.LogChan) <= 0 {
			break L
		}
		select {
		case fback := <-c.fLink.CommonChan:
			if fback.TypeID == BusDataTypeEOF {
				// entering shutdown mode
				power = false
			}
		case fback := <-c.fLink.LogChan:
			c.printMessage(fback)
		}
	}
}

func (c *ConsoleLogger) setDefaultTheme() {
	prefxmap = map[int]string{
		base.CriticalLevel:      " " + term.White + term.BGRed + " " + term.EmojiSet["FaceScreamingInFear"] + " CRITICAL ERROR " + term.Reset,
		base.ErrorLevel:         " " + term.White + term.BGRed + " " + term.EmojiSet["PoliceCarLight"] + " ERROR " + term.Reset,
		base.WarningLevel:       " " + term.Black + term.BGYellow + " " + term.EmojiSet["Construction"] + " WARNING " + term.Reset,
		base.InfoLevel:          " " + term.Blue + "»" + term.Magenta + "»" + term.Reset,
		base.DebugLevel:         " " + term.Black + term.BGMagenta + " " + term.EmojiSet["Wrench"] + " DEBUG " + term.Reset,
		base.ParanoicDebugLevel: " " + term.Black + term.BGMagenta + " " + term.EmojiSet["Wrench"] + term.EmojiSet["Wrench"] + " DEBUG " + term.Reset,
		base.NotsetLevel:        "( · ) ",
	}
}

// InitConsoleLogger func
func InitConsoleLogger(upgrader func(*BusConsumerLink) error) {
	threadcolor = make(map[string]string)
	progress = make(map[string]*progressInfo)
	fLink := &BusConsumerLink{
		Name:           "Console",
		LogChan:        make(chan *BusData, 100),
		CommonChan:     make(chan *BusData, 100),
		Off:            make(chan struct{}),
		AllowEventData: true,
	}
	clogger := &ConsoleLogger{fLink: fLink, logLevelFilter: config.LOGLEVEL}
	clogger.setDefaultTheme()
	SBus.connect <- fLink
	SBus.castWaiter.Add(1)

	if !term.IsTerminal() {
		go clogger.readCastBus()
	} else {
		go func() {
			if upgrader != nil {
				err := upgrader(fLink)
				if err != nil {
					LogErr(errors.Join(fmt.Errorf("a problem in TUI ocurred. Downgroading to non-interactive console mode"), err).Error(), nil)
				}
			}

			// on TUI exit (gracefully or with err, start basic logger to print last shutdown msgs)
			if !fLink.Degraded {
				clogger.readCastBus()
			} else {
				SBus.castWaiter.Done()
			}
		}()
	}
}
