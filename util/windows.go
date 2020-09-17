//go:build windows

// Nebulant
// Copyright (C) 2021  Develatio Technologies S.L.

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
