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

func NewMultiUserLdisc() *MultiUserLdisc {
	return &MultiUserLdisc{users: make(map[io.ReadWriteCloser]bool)}
}

type MultiUserLdisc struct {
	// internal mustar
	mustarFD []io.ReadWriteCloser
	// internal sluva
	sluvaFD []io.ReadWriteCloser
	// a map of sluvas to write from mustarFD
	users map[io.ReadWriteCloser]bool
	errs  []error

	// mustaropen bool
	// sluvaopen  bool
}

func (m *MultiUserLdisc) SetMustarFD(fd io.ReadWriteCloser) {
	for _, f := range m.mustarFD {
		if f == fd {
			return
		}
	}
	m.mustarFD = append(m.mustarFD, fd)
}

func (m *MultiUserLdisc) SetSluvaFD(fd io.ReadWriteCloser) {
	for _, f := range m.sluvaFD {
		if f == fd {
			return
		}
	}
	m.sluvaFD = append(m.sluvaFD, fd)
}

func (m *MultiUserLdisc) ReceiveMustarBuff(n int, mInFD *PortFD) {
	// fmt.Println("receive mustar ", mInFD)
	p := make([]byte, n)
	_, err := mInFD.Read(p)
	if err != nil {
		m.errs = append(m.errs, err)
	}
	for _, slv := range m.sluvaFD {
		_, err := slv.Write(p)
		if err != nil {
			m.errs = append(m.errs, err)
		}
	}
}

func (m *MultiUserLdisc) ReceiveSluvaBuff(n int, sInFD *PortFD) {
	// fmt.Println("receive sluva ", sInFD)
	p := make([]byte, n)
	_, err := sInFD.Read(p)
	if err != nil {
		m.errs = append(m.errs, err)
	}
	for _, mst := range m.mustarFD {
		_, err := mst.Write(p)
		if err != nil {
			m.errs = append(m.errs, err)
		}
	}
}

func (m *MultiUserLdisc) AddUser(fd io.ReadWriteCloser) {
	m.users[fd] = true
}

func (m *MultiUserLdisc) DelUser(fd io.ReadWriteCloser) {
	delete(m.users, fd)
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
	return nil
}
