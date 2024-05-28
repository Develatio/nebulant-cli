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

package ws

import (
	"errors"
	"fmt"
	"io"

	"github.com/gorilla/websocket"
)

func NewWebSocketReadWriteCloser(c *websocket.Conn) io.ReadWriteCloser {
	wsrwc := &WebSocketReadWriteCloser{conn: c}
	c.SetPingHandler(wsrwc.handlePing)
	return wsrwc
}

type WebSocketReadWriteCloser struct {
	conn  *websocket.Conn
	rr    io.Reader
	ww    io.WriteCloser
	close bool
	err   error
}

func (w *WebSocketReadWriteCloser) Write(p []byte) (n int, err error) {
	ww, err := w.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	w.ww = ww
	defer w.ww.Close()
	return w.ww.Write(p)
}

func (w *WebSocketReadWriteCloser) Close() error {
	return w.ww.Close()
}

func (w *WebSocketReadWriteCloser) Read(p []byte) (n int, err error) {
	if w.close {
		return 0, errors.Join(w.err, fmt.Errorf("reading on closed socket"))
	}
	for {
		if w.rr == nil {
			// start new nextreader
			// from doc: Applications must break out of the
			// application's read loop when this method returns
			// a non-nil error value. Errors returned from this
			// method are permanent. Once this method returns a
			// non-nil error, all subsequent calls to this method
			// return the same error.
			mt, rr, err := w.conn.NextReader()
			if err != nil {
				err2 := w.conn.Close()
				w.close = true
				w.err = errors.Join(err, err2)
				return 0, w.err
			}
			switch mt {
			case websocket.CloseMessage:
				w.close = true
				// force new reader
				continue
			case websocket.PingMessage, websocket.PongMessage:
				// force new reader
				continue
			}
			// set reader to .Read()
			w.rr = rr
		}
		// from doc:
		// An instance of this general case is that a Reader returning a non-zero number
		// of bytes at the end of the input stream may return either err == EOF or err == nil.
		// The next Read should return 0, EOF.
		n, err = w.rr.Read(p)
		if n == 0 && err == io.EOF {
			// end of this reader force new
			w.rr = nil
			// until if this ws is in close process
			if w.close {
				return 0, io.EOF
			}
			continue
		}
		// eof err but n > 0, hide
		// err and return to keep
		// this Reader interface
		// alwais open
		if err == io.EOF {
			err = nil
		}
		return
	}
}

func (w *WebSocketReadWriteCloser) handlePing(appData string) error {
	return w.conn.WriteMessage(websocket.PongMessage, []byte(appData))
}
