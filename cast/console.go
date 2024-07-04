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
	"errors"
	"fmt"
	"log"

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/term"
)

var counter int = 0

var prefxmap map[int]string = map[int]string{
	CriticalLevel: "!CRITICAL ",
	ErrorLevel:    "!ERROR ",
	WarningLevel:  "!WARNING ",
	InfoLevel:     "",
	DebugLevel:    "DEBUG ",
	NotsetLevel:   "",
}

var threadcolor map[string]string

// ConsoleLogger struct
type ConsoleLogger struct {
	fLink *BusConsumerLink
}

func FormatConsoleLogMsg(fback *BusData) string {
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
		format = fmt.Sprintf("%s %s\t |%s %%s %%s", color, aname, term.Reset)
	}

	if fback.EventID != nil {
		switch *fback.EventID {
		case EventActionInit:
			return fmt.Sprintf(format, prefxmap[*fback.LogLevel], "start")
		case EventActionKO:
			return fmt.Sprintf(format, prefxmap[*fback.LogLevel], *fback.M)
		case EventActionUnCaughtKO:
			return fmt.Sprintf(format, prefxmap[*fback.LogLevel], *fback.M)
		case EventActionOK:
			return fmt.Sprintf(format, prefxmap[*fback.LogLevel], "done")
		case EventActionUnCaughtOK:
			return fmt.Sprintf(format, prefxmap[*fback.LogLevel], "done")
		case EventThreadDestroyed:
			delete(threadcolor, *fback.ThreadID)
		default:
			return ""
		}
	}

	if fback.Raw {
		if fback.M != nil {
			return term.Reset + *fback.M + term.Reset
		} else {
			return term.Reset
		}
	} else {
		if fback.M != nil {
			if config.DEBUG {
				counter++
				return fmt.Sprintf(format, prefxmap[*fback.LogLevel], fmt.Sprintf("[%v] %v", counter, *fback.M))
			} else {
				return fmt.Sprintf(format, prefxmap[*fback.LogLevel], *fback.M)
			}
		} else {
			return fmt.Sprintf(format, prefxmap[*fback.LogLevel], "")
		}
	}
}

func (c *ConsoleLogger) printMessage(fback *BusData) bool {
	if !config.DEBUG && fback.LogLevel != nil && *fback.LogLevel == DebugLevel {
		return false
	}

	msg := FormatConsoleLogMsg(fback)

	if fback.Raw {
		var err error
		_, err = term.Print(msg)
		if err != nil {
			return false
		}
	} else {
		log.Println(msg)
	}
	return false
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
		CriticalLevel:      " " + term.White + term.BGRed + " " + term.EmojiSet["FaceScreamingInFear"] + " CRITICAL ERROR " + term.Reset,
		ErrorLevel:         " " + term.White + term.BGRed + " " + term.EmojiSet["PoliceCarLight"] + " ERROR " + term.Reset,
		WarningLevel:       " " + term.Black + term.BGYellow + " " + term.EmojiSet["Construction"] + " WARNING " + term.Reset,
		InfoLevel:          " " + term.Blue + "»" + term.Magenta + "»" + term.Reset,
		DebugLevel:         " " + term.Black + term.BGMagenta + " " + term.EmojiSet["Wrench"] + " DEBUG " + term.Reset,
		ParanoicDebugLevel: " " + term.Black + term.BGMagenta + " " + term.EmojiSet["Wrench"] + term.EmojiSet["Wrench"] + " DEBUG " + term.Reset,
		NotsetLevel:        "( · ) ",
	}
}

// InitConsoleLogger func
func InitConsoleLogger(upgrader func(*BusConsumerLink) error) {
	threadcolor = make(map[string]string)
	fLink := &BusConsumerLink{
		Name:           "Console",
		LogChan:        make(chan *BusData, 100),
		CommonChan:     make(chan *BusData, 100),
		Off:            make(chan struct{}),
		AllowEventData: true,
	}
	clogger := &ConsoleLogger{fLink: fLink}
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
