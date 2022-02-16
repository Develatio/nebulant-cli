//go:build js

// Nebulant
// Copyright (C) 2022  Develatio Technologies S.L.

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
)

var Stdout = os.Stdout
var Stderr = os.Stderr
var CharBell = []byte(fmt.Sprintf("%c", 7))[0]

// Jus for bypass build, not really used
var ErrInterrupt = fmt.Errorf("^C")
var ErrEOF = fmt.Errorf("^D")
