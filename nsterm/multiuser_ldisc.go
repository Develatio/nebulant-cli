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
	fmt.Printf("Sluva count: %v\n", len(m.sluvaFD))
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
