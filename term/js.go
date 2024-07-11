//go:build js

// MIT License
//
// Copyright (C) 2022  Develatio Technologies S.L.

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

package term

import (
	"fmt"
	"os"
)

var GenuineOsStdout *os.File = os.Stdout
var GenuineOsStderr *os.File = os.Stderr
var GenuineOsStdin *os.File = os.Stdin

var Stdout = os.Stdout
var Stderr = os.Stderr
var Stdin = os.Stdin
var CharBell = []byte(fmt.Sprintf("%c", 7))[0]

// Jus for bypass build, not really used
var ErrInterrupt = fmt.Errorf("^C")
var ErrEOF = fmt.Errorf("^D")

func GetOSPTY(cfg *OSPTYConf) (OSPTY, error) { return nil, nil }
func EnableColorSupport() error              { return nil }
func SetHideCursor() error                   { return nil }
func SetShowCursor() error                   { return nil }
