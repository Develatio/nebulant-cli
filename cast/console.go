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
	"strconv"

	"github.com/develatio/nebulant-cli/config"
)

var reset string = "\033[0m"
var red string = "\033[31m"

// var green string = "\033[32m"
var yellow string = "\033[33m"

var blue string = "\033[34m"
var purple string = "\033[35m"

// var cyan string = "\033[36m"
// var gray string = "\033[37m"
// var white string = "\033[97m"

var failprefx string = "[!] "
var okprefx string = "(·) "

var colormap map[int]string = map[int]string{
	CriticalLevel: red,
	ErrorLevel:    red,
	WarningLevel:  yellow,
	InfoLevel:     blue,
	DebugLevel:    purple,
	NotsetLevel:   blue,
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
	fLink  *FeedBackLink
	colors bool
}

func (c *ConsoleLogger) printMessage(fback *FeedBack) bool {
	if !config.DEBUG && fback.LogLevel != nil && *fback.LogLevel == DebugLevel {
		return false
	}

	color := ""
	prefx := strconv.Itoa(*fback.LogLevel) + prefxmap[*fback.LogLevel]
	if c.colors {
		color = colormap[*fback.LogLevel]
	}

	if fback.Raw {
		fmt.Print(color + string(fback.B) + reset)
	} else {
		log.Println(color + prefx + reset + string(fback.B) + reset)
	}
	return false
}

func (c *ConsoleLogger) readCastBus() {
	defer SBus.castWaiter.Done()
	for fback := range c.fLink.FeedBackBus {
		if fback.TypeID == FeedBackEOF {
			break
		}
		if fback.TypeID != FeedBackLog {
			continue
		}
		c.printMessage(fback)
	}
}

// InitConsoleLogger func
func InitConsoleLogger(colors bool) {
	fLink := &FeedBackLink{
		FeedBackBus: make(chan *FeedBack, 100),
	}
	clogger := &ConsoleLogger{fLink: fLink, colors: colors}
	SBus.connect <- fLink
	SBus.castWaiter.Add(1)
	go clogger.readCastBus()
}
