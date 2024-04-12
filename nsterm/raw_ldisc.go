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

func NewRawLdisc() *RawLdisc {
	return &RawLdisc{}
}

type RawLdisc struct {
	mustarFD io.ReadWriteCloser
	sluvaFD  io.ReadWriteCloser
	errs     []error
	// mustaropen bool
	// sluvaopen  bool
}

func (r *RawLdisc) SetMustarFD(fd io.ReadWriteCloser) {
	r.mustarFD = fd
}

func (r *RawLdisc) SetSluvaFD(fd io.ReadWriteCloser) {
	r.sluvaFD = fd
}

func (r *RawLdisc) ReceiveMustarBuff(n int, mInFD *PortFD) {
	// reading from mustar.in_r
	_, err := io.CopyN(r.sluvaFD, mInFD, int64(n))
	if err != nil {
		r.errs = append(r.errs, err)
	}

	// emulate real raw writing char by char
	// TODO: maybe rune decode is needed
	// for i := 0; i < n; i++ {
	// 	_, err := io.CopyN(r.sluvaFD, mInFD, 1)
	// 	if err != nil {
	// 		r.errs = append(r.errs, err)
	// 	}
	// }
}

func (r *RawLdisc) ReceiveSluvaBuff(n int, sInFD *PortFD) {
	_, err := io.CopyN(r.mustarFD, sInFD, int64(n))
	if err != nil {
		r.errs = append(r.errs, err)
		return
	}
}

// NOOP
func (r *RawLdisc) ReadRuneBuff() []rune { return nil }

// NOOP
func (r *RawLdisc) SetBuff(s string) {}

// NOOP
func (r *RawLdisc) GetESC() chan string { return nil }

func (r *RawLdisc) IOctl() {

}

func (r *RawLdisc) Close() error {
	return nil
}
