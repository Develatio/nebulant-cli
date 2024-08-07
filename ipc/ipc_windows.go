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

package ipc

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func (p *IPC) listen() (net.Listener, error) {
	path := `\\.\pipe\` + "ipc_" + p.uuid

	l, err := winio.ListenPipe(path, nil)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func Read(ipsid string, ipcid string, msg string) (string, error) {
	path := `\\.\pipe\` + "ipc_" + ipsid
	c, err := winio.DialPipe(path, nil)
	if err != nil {
		return "", err
	}
	defer c.Close()

	_, err = c.Write([]byte(ipsid + " " + ipcid + " " + msg))
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)

	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}
	return string(buf[0:n]), nil
}
