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

package ws

import (
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
	for {
		if w.rr == nil {
			// start new nextreader
			mt, rr, err := w.conn.NextReader()
			if err != nil {
				return 0, err
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
