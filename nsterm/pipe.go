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
	"io"
)

var ErrClosedPipe = io.ErrClosedPipe

type pipe struct {
	// buff  *bytes.Buffer
	done   chan struct{}
	buff   chan []byte
	remain chan []byte
}

func (pp *pipe) Write(p []byte) (n int, err error) {
	pp.buff <- p
	return len(p), nil
}

func (pp *pipe) Close() error {
	close(pp.done)
	return nil
}

func (pp *pipe) Read(p []byte) (n int, err error) {
	select {
	case r := <-pp.remain:
		if len(r) > len(p) {
			nn := copy(p, r)
			pp.remain <- r[nn:]
			return nn, nil
		}
		return copy(p, r), nil
	case <-pp.done:
		return 0, io.EOF
	default:
		b := <-pp.buff
		if len(b) > len(p) {
			nn := copy(p, b)
			pp.remain <- b[nn:]
			return nn, nil
		}
		return copy(p, b), nil

	}
}

func NewPipe() (io.ReadCloser, io.WriteCloser) {
	pp := &pipe{
		done:   make(chan struct{}),
		buff:   make(chan []byte, 250),
		remain: make(chan []byte, 1),
	}
	return pp, pp
}
