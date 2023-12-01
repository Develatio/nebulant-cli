// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

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

package subcom

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/term"
)

func newBar() {
	line := term.AppendLine()
	defer line.Close()
	bar, err := line.GetProgressBar(int64(100), "Testing bar", false)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		err := bar.Add(1)
		if err != nil {
			panic(err)
		}
		time.Sleep(60000 * time.Microsecond)
	}
}

func parseTestsFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("debugterm", flag.ExitOnError)
	fs.SetOutput(cmdline.Output())
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func DebugtermCmd(cmdline *flag.FlagSet) (int, error) {
	_, err := parseTestsFs(cmdline)
	if err != nil {
		return 1, err
	}

	go func() {
		counter1 := 0
		for {
			counter1++
			cast.LogInfo("blabla blablabla blabla bla bla bla blabla bla bla bla bla ... "+strings.Repeat(" ", counter1)+strconv.Itoa(counter1), nil)
			time.Sleep(600000 * time.Microsecond)
			if counter1 == 9 {
				counter1 = 0
			}
		}
	}()

	counter2 := 0
	lin := term.AppendLine()
	defer lin.Close()
L:
	for {
		counter2++

		var vv string
		_, err = lin.Scanln("Nebulant "+strconv.Itoa(counter2)+"> ", nil, &vv)
		if err != nil {
			return 1, err
		}
		switch vv {
		case "exit":
			break L
		case "addbar":
			go newBar()
		}
	}
	return 0, nil
}
