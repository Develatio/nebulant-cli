//go:build windows

// MIT License
//
// Copyright (C) 2021  Develatio Technologies S.L.

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

package util

import (
	"syscall"
	"unsafe"
)

func CommandLineToArgv(cmd string) ([]string, error) {
	ptr, err := syscall.UTF16PtrFromString(cmd)
	if err != nil {
		return nil, err
	}
	var argc int32
	argv, err := syscall.CommandLineToArgv(ptr, &argc)
	if err != nil {
		return nil, err
	}
	defer syscall.LocalFree((syscall.Handle)(unsafe.Pointer(argv)))
	args := make([]string, argc)
	for idx := range args {
		args[idx] = syscall.UTF16ToString(argv[idx][:])
	}
	return args, nil
}
