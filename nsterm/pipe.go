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

package nsterm

import (
	"fmt"
	"io"
	"time"
)

var ErrClosedPipe = io.ErrClosedPipe

type pipe struct {
	// buff  *bytes.Buffer
	closed bool
	buff   chan []byte
	remain chan []byte
}

func (pp *pipe) Write(p []byte) (n int, err error) {
	// fmt.Println("-> buff", bytes.TrimRight(p, "\x00"))
	if pp.closed {
		return 0, fmt.Errorf("Write to closed pipe")
	}

	// buffsize := len(pp.buff)
	// fmt.Printf("Write: %v, Buff: %v\n", len(p), buffsize)

	sent := false
	failcount := 0
	for !sent {
		select {
		// TODO: WARNING
		// race condition here, chan could be closed
		// just before this line
		case pp.buff <- p:
			sent = true
		default:
			if failcount == 10 {
				return 0, io.ErrShortBuffer
			}
			// TODO: close after some err count to
			// handle buffer overflow
			if pp.closed {
				return 0, io.ErrClosedPipe
			}
			time.Sleep(100 * time.Millisecond)
			failcount++
		}
	}

	return len(p), nil
}

func (pp *pipe) Close() error {
	if pp.closed {
		return io.ErrClosedPipe
	}
	pp.closed = true
	close(pp.buff)
	return nil
}

func (pp *pipe) Read(p []byte) (n int, err error) {
	select {
	case r := <-pp.remain:
		if len(r) > len(p) {
			// fmt.Println("<- buff", bytes.TrimRight(r, "\x00"))
			nn := copy(p, r)
			pp.remain <- r[nn:]
			return nn, nil
		}
		// fmt.Println("<- buff", bytes.TrimRight(r, "\x00"))
		return copy(p, r), nil
	// case <-pp.done:
	// 	return 0, io.EOF
	default:
		// a closed channel with previous data, will
		// send the remain data before close
		b := <-pp.buff
		var err error = nil
		if pp.closed {
			// read can still write to p []byte while
			// return io.EOF data. Only 0, EOF return
			// will stop subsequent readings
			err = io.EOF
		}
		if len(b) > len(p) {
			nn := copy(p, b)
			pp.remain <- b[nn:]
			// fmt.Println("<- buff", bytes.TrimRight(b, "\x00"))
			return nn, err
		}
		// fmt.Println("<- buff", bytes.TrimRight(b, "\x00"))
		return copy(p, b), err
	}
}

func NewPipe() (io.ReadCloser, io.WriteCloser) {
	pp := &pipe{
		// done:   make(chan struct{}),
		buff:   make(chan []byte, 4096),
		remain: make(chan []byte, 1),
	}
	return pp, pp
}
