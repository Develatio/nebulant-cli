//go:build !windows

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

package term

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func getCursorPosition() (width, height int, err error) {
	oldState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)
	if _, err = os.Stdout.Write([]byte("\033[6n")); err != nil {
		return 0, 0, err
	}
	if _, err = fmt.Fscanf(os.Stdin, "\033[%d;%d", &width, &height); err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func EnableColorSupport() error { return nil }
