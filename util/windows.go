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
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"syscall"
	"time"
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

// due to https://github.com/golang/go/issues/31247
// and due to https://github.com/golang/go/issues/31247
// and because robustio is available only from go 1.20.5
// and we use go 1.19 (use robustio on go update)
// https://pkg.go.dev/cmd/go/internal/robustio#Rename
func RenameFile(oldpath, newpath string) error {
	var errno syscall.Errno
	var retries int
	var maxRetries int = 1000
	for {
		retries++
		err := os.Rename(oldpath, newpath)
		if errors.As(err, &errno) {
			if retries > maxRetries {
				return err
			}
			switch errno {
			case syscall.ERROR_ACCESS_DENIED,
				syscall.ERROR_FILE_NOT_FOUND,
				windows.ERROR_SHARING_VIOLATION:
				time.Sleep(1 * time.Millisecond)
			default:
				return err
			}
		} else {
			return err
		}
	}
}
