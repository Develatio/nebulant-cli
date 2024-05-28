//go:build windows

// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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
	"context"
	"os"

	"github.com/Azure/go-ansiterm/winterm"
	"github.com/UserExistsError/conpty"
	"golang.org/x/sys/windows"
)

func getCursorPosition() (width, height int16, err error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(int(os.Stdout.Fd())), &info); err != nil {
		return 0, 0, err
	}

	return info.CursorPosition.X, info.CursorPosition.Y, nil
}

func EnableColorSupport() error {
	var st uint32
	err := windows.GetConsoleMode(windows.Handle(int(os.Stdin.Fd())), &st)
	if err != nil {
		return err
	}

	// https://learn.microsoft.com/en-us/windows/console/setconsolemode
	// ENABLE_VIRTUAL_TERMINAL_PROCESSING 0x0004
	st &^= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	windows.SetConsoleMode(windows.Handle(int(os.Stdin.Fd())), st)
	return nil
}

func SetHideCursor() error {
	var info winterm.CONSOLE_CURSOR_INFO
	handle := uintptr(int(os.Stdout.Fd()))
	err := winterm.GetConsoleCursorInfo(handle, &info)
	if err != nil {
		return err
	}
	info.Visible = 0

	err = winterm.SetConsoleCursorInfo(handle, &info)
	if err != nil {
		return err
	}
	return nil
}

func SetShowCursor() error {
	var info winterm.CONSOLE_CURSOR_INFO
	handle := uintptr(int(os.Stdout.Fd()))
	err := winterm.GetConsoleCursorInfo(handle, &info)
	if err != nil {
		return err
	}
	info.Visible = 1

	err = winterm.SetConsoleCursorInfo(handle, &info)
	if err != nil {
		return err
	}
	return nil
}

type winPTY struct {
	wrap *conpty.ConPty
}

func (n *winPTY) Close() error                { return n.wrap.Close() }
func (n *winPTY) Read(p []byte) (int, error)  { return n.wrap.Read(p) }
func (n *winPTY) Write(p []byte) (int, error) { return n.wrap.Write(p) }
func (n *winPTY) Wait(ctx context.Context) (int64, error) {
	exitCode, err := n.wrap.Wait(ctx)
	return int64(exitCode), err
}

func GetOSPTY(cfg *OSPTYConf) (OSPTY, error) {
	cpty, err := conpty.Start(cfg.Shell)
	if err != nil {
		return nil, err
	}
	return &winPTY{
		wrap: cpty,
	}, nil
}
