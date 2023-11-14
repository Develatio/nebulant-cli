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
		CriticalLevel: " " + term.White + term.BGRed + " " + term.EmojiSet["FaceScreamingInFear"] + " CRITICAL ERROR " + term.Reset,
		ErrorLevel:    " " + term.White + term.BGRed + " " + term.EmojiSet["PoliceCarLight"] + " ERROR " + term.Reset,
		WarningLevel:  " " + term.Black + term.BGYellow + " " + term.EmojiSet["Construction"] + " WARNING " + term.Reset,
		InfoLevel:     " " + term.Blue + "»" + term.Magenta + "»" + term.Reset,
		DebugLevel:    " " + term.Black + term.BGMagenta + " " + term.EmojiSet["Wrench"] + " DEBUG " + term.Reset,
		NotsetLevel:   "( · ) ",
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
