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

func (c *ConsoleLogger) printMessage(fback *BusData) bool {
	if !config.DEBUG && fback.LogLevel != nil && *fback.LogLevel == DebugLevel {
		return false
	}

	if fback.Raw {
		var err error
		if fback.M != nil {
			_, err = term.Print(term.Reset + *fback.M + term.Reset)
		} else {
			_, err = term.Print(term.Reset)
		}
		if err != nil {
			return false
		}
	} else {
		if fback.M != nil {
			// log will use term.Output
			// because log.SetOutput(Stdout)
			if config.DEBUG {
				counter++
				log.Println(prefxmap[*fback.LogLevel] + " " + fmt.Sprintf("[%v] %v", counter, *fback.M) + term.Reset)
			} else {
				log.Println(prefxmap[*fback.LogLevel] + " " + *fback.M + term.Reset)
			}
		} else {
			log.Println(prefxmap[*fback.LogLevel] + "" + term.Reset)
		}
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
		Name:       "Console",
		LogChan:    make(chan *BusData, 100),
		CommonChan: make(chan *BusData, 100),
	}
	clogger := &ConsoleLogger{fLink: fLink}
	clogger.setDefaultTheme()
	SBus.connect <- fLink
	SBus.castWaiter.Add(1)
	go clogger.readCastBus()
}
