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
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty"
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

func SetHideCursor() error { return nil }

func SetShowCursor() error { return nil }

type nixPTY struct {
	wrap *os.File
	cmd  *exec.Cmd
}

func (n *nixPTY) Close() error                { return n.wrap.Close() }
func (n *nixPTY) Read(p []byte) (int, error)  { return n.wrap.Read(p) }
func (n *nixPTY) Write(p []byte) (int, error) { return n.wrap.Write(p) }
func (n *nixPTY) Wait(ctx context.Context) (int64, error) {
	exitErr := n.cmd.Wait()
	if eerr, ok := exitErr.(*exec.ExitError); ok {
		return int64(eerr.ExitCode()), eerr
	}
	if exitErr != nil {
		return 1, exitErr
	}
	return 0, nil
}

func GetOSPTY(shell string) (OSPTY, error) {
	cmd := exec.Command(shell)
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &nixPTY{
		wrap: f,
		cmd:  cmd,
	}, nil
}
