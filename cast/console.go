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

// ConsoleLogger struct
type ConsoleLogger struct {
	fLink *BusConsumerLink
}

func FormatConsoleLogMsg(fback *BusData) string {
	if fback.EventID != nil {
		switch *fback.EventID {
		case EventActionInit:
			return fmt.Sprintf(" %s %s (%s in thread %s)", term.EmojiSet["Sparkles"]+term.EmojiSet["Dizzy"], *fback.ActionName, *fback.ActionID, *fback.ThreadID)
		case EventActionKO:
			a := fmt.Sprintf(" %s %s (%s in thread %s)", term.EmojiSet["CrossMarkButton"], *fback.ActionName, *fback.ActionID, *fback.ThreadID)
			b := fmt.Sprintf(" %s %s %s", prefxmap[*fback.LogLevel], *fback.M, term.Reset)
			return a + b
		case EventActionUnCaughtKO:
			a := fmt.Sprintf(" ERROR | %s (%s in thread %s)", *fback.ActionName, *fback.ActionID, *fback.ThreadID)
			b := fmt.Sprintf(" %s %s %s", prefxmap[*fback.LogLevel], *fback.M, term.Reset)
			return a + b
		case EventActionOK:
			return fmt.Sprintf(" %s %s (%s in thread %s)", term.EmojiSet["CheckMarkButton"], *fback.ActionName, *fback.ActionID, *fback.ThreadID)
		case EventActionUnCaughtOK:
			return fmt.Sprintf(" %s %s (%s in thread %s)", term.EmojiSet["CheckMarkButton"], *fback.ActionName, *fback.ActionID, *fback.ThreadID)
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
		aname := ""
		if fback.ActionName != nil {
			aname = fmt.Sprintf("[%s]", *fback.ActionName)
		}
		if fback.M != nil {
			if config.DEBUG {
				counter++
				// log.Println(prefxmap[*fback.LogLevel] + " " + fmt.Sprintf("[%v] %v", counter, *fback.M) + term.Reset)
				return fmt.Sprintf(" %s %s [%v] %s%s", prefxmap[*fback.LogLevel], aname, counter, *fback.M, term.Reset)
			} else {
				// log.Println(prefxmap[*fback.LogLevel] + " " + *fback.M + term.Reset)
				return fmt.Sprintf(" %s %s %s %s", prefxmap[*fback.LogLevel], aname, *fback.M, term.Reset)
			}
		} else {
			// log.Println(prefxmap[*fback.LogLevel] + "" + term.Reset)
			return fmt.Sprintf("%s%s", prefxmap[*fback.LogLevel], term.Reset)
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
		// if fback.M != nil {
		// 	// log will use term.Output
		// 	// because log.SetOutput(Stdout)
		// 	if config.DEBUG {
		// 		counter++
		// 		log.Println(prefxmap[*fback.LogLevel] + " " + fmt.Sprintf("[%v] %v", counter, *fback.M) + term.Reset)
		// 	} else {
		// 		log.Println(prefxmap[*fback.LogLevel] + " " + *fback.M + term.Reset)
		// 	}
		// } else {
		// 	log.Println(prefxmap[*fback.LogLevel] + "" + term.Reset)
		// }
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
func InitConsoleLogger() {
	fLink := &BusConsumerLink{
		Name:           "Console",
		LogChan:        make(chan *BusData, 100),
		CommonChan:     make(chan *BusData, 100),
		AllowEventData: true,
	}
	clogger := &ConsoleLogger{fLink: fLink}
	clogger.setDefaultTheme()
	SBus.connect <- fLink
	SBus.castWaiter.Add(1)

	// TODO: if interactive/no term
	// go clogger.readCastBus()

	StartUI(fLink)
}
