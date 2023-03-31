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
	"log"
	"strconv"

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/term"
)

var failprefx string = "[!] "
var okprefx string = "(·) "

var colormap map[int]string = map[int]string{
	CriticalLevel: term.Red,
	ErrorLevel:    term.Red,
	WarningLevel:  term.Yellow,
	InfoLevel:     term.Blue,
	DebugLevel:    term.Purple,
	NotsetLevel:   term.Blue,
}

var prefxmap map[int]string = map[int]string{
	CriticalLevel: failprefx,
	ErrorLevel:    failprefx,
	WarningLevel:  failprefx,
	InfoLevel:     okprefx,
	DebugLevel:    okprefx,
	NotsetLevel:   okprefx,
}

// ConsoleLogger struct
type ConsoleLogger struct {
	fLink  *BusConsumerLink
	colors bool
}

func (c *ConsoleLogger) printMessage(fback *BusData) bool {
	if !config.DEBUG && fback.LogLevel != nil && *fback.LogLevel == DebugLevel {
		return false
	}

	color := ""
	prefx := strconv.Itoa(*fback.LogLevel) + prefxmap[*fback.LogLevel]
	if c.colors {
		color = colormap[*fback.LogLevel]
	}

	if fback.Raw {
		var err error
		if fback.M != nil {
			_, err = term.Print(color + *fback.M + term.Reset)
		} else {
			_, err = term.Print(color + "" + term.Reset)
		}
		if err != nil {
			return false
		}
	} else {
		if fback.M != nil {
			log.Println(color + prefx + term.Reset + *fback.M + term.Reset)
		} else {
			log.Println(color + prefx + term.Reset + "" + term.Reset)
		}
	}
	return false
}

func (c *ConsoleLogger) readCastBus() {
	defer SBus.castWaiter.Done()
	power := true
L:
	for {
		select {
		case fback := <-c.fLink.CommonChan:
			if fback.TypeID == BusDataTypeEOF {
				// entering shutdown mode
				power = false
			}
		case fback := <-c.fLink.LogChan:
			c.printMessage(fback)
		default:
			// perform exit
			if !power {
				break L
			}
		}
	}
}

// InitConsoleLogger func
func InitConsoleLogger(colors bool) {
	fLink := &BusConsumerLink{
		Name:       "Console",
		LogChan:    make(chan *BusData, 100),
		CommonChan: make(chan *BusData, 100),
	}
	clogger := &ConsoleLogger{fLink: fLink, colors: colors}
	SBus.connect <- fLink
	SBus.castWaiter.Add(1)
	go clogger.readCastBus()
}
