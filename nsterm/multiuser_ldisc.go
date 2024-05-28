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
	"fmt"
	"io"
)

func NewMultiUserLdisc() *MultiUserLdisc {
	return &MultiUserLdisc{
		mustarFD: make(map[io.ReadWriteCloser]bool),
		sluvaFD:  make(map[io.ReadWriteCloser]bool),
	}
}

type MultiUserLdisc struct {
	// internal mustar
	mustarFD map[io.ReadWriteCloser]bool
	// internal sluva
	sluvaFD map[io.ReadWriteCloser]bool
	// a map of sluvas to write from mustarFD
	errs             []error
	MustarWriteCount int
	MustarReadCount  int

	// mustaropen bool
	// sluvaopen  bool
}

func (m *MultiUserLdisc) SetMustarFD(fd io.ReadWriteCloser) {
	if _, exists := m.mustarFD[fd]; exists {
		return
	}
	m.mustarFD[fd] = true
}

func (m *MultiUserLdisc) SetSluvaFD(fd io.ReadWriteCloser) {
	if _, exists := m.sluvaFD[fd]; exists {
		return
	}
	m.sluvaFD[fd] = true
}

func (m *MultiUserLdisc) ReceiveMustarBuff(n int, mInFD *PortFD) {
	// m.MustarReadCount = m.MustarReadCount + n
	p := make([]byte, n)
	_, err := mInFD.Read(p)
	if err != nil {
		m.errs = append(m.errs, err)
	}
	for slv := range m.sluvaFD {
		// m.MustarWriteCount = m.MustarWriteCount + n
		_, err := slv.Write(p)
		if err != nil {
			slv.Close()
			delete(m.sluvaFD, slv) // gtfo
			fmt.Println("Deleting sluva")
			m.errs = append(m.errs, err)
		}
	}
	// fmt.Printf("Sluva count: %v\n", len(m.sluvaFD))
	// fmt.Printf("mustar r: %v w: %v\n", m.MustarReadCount, m.MustarWriteCount)
}

func (m *MultiUserLdisc) ReceiveSluvaBuff(n int, sInFD *PortFD) {
	// fmt.Println("receive sluva ", sInFD)
	p := make([]byte, n)
	_, err := sInFD.Read(p)
	if err != nil {
		m.errs = append(m.errs, err)
	}
	for mst := range m.mustarFD {
		_, err := mst.Write(p)
		if err != nil {
			mst.Close()
			m.errs = append(m.errs, err) // gtfo
			delete(m.mustarFD, mst)
		}
	}
}

// NOOP
func (m *MultiUserLdisc) ReadRuneBuff() []rune { return nil }

// NOOP
func (m *MultiUserLdisc) SetBuff(s string) {}

// NOOP
func (m *MultiUserLdisc) GetESC() chan string { return nil }

func (m *MultiUserLdisc) IOctl() {

}

func (m *MultiUserLdisc) Close() error {
	for mst := range m.mustarFD {
		mst.Close()
	}
	for slv := range m.sluvaFD {
		slv.Close()
	}
	m.mustarFD = make(map[io.ReadWriteCloser]bool)
	m.sluvaFD = make(map[io.ReadWriteCloser]bool)
	return nil
}
