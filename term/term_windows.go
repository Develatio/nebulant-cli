//go:build windows

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
	"os"

	"github.com/Azure/go-ansiterm/winterm"
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
