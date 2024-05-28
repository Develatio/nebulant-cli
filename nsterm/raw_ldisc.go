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
