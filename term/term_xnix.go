//go:build !windows && !js

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

func GetOSPTY(cfg *OSPTYConf) (OSPTY, error) {
	cmd := exec.Command(cfg.Shell)
	cmd.Env = append(cmd.Env, cfg.Env...)
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &nixPTY{
		wrap: f,
		cmd:  cmd,
	}, nil
}
