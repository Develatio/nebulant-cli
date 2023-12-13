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
	"errors"
	"io"
	"sync"
)

type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error {
	return nil
}

func NopWriter(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

func NewVirtPTY() *VPTY2 {
	vpty2 := &VPTY2{}

	s := NewPort()
	// wrap in_w to call vpty ticker on every in_w write
	// writes comes from shell/app
	swticker := &WriteCloserTicker{w: s.in_w, ticker: vpty2.sluvaWriteTick}
	s.in_w = swticker

	m := NewPort()
	// wrap in_w to call vpty ticker on every in_w write
	// writes comes from term
	mwticker := &WriteCloserTicker{w: m.in_w, ticker: vpty2.mustarWriteTick}
	m.in_w = mwticker

	vpty2.sluva = s
	vpty2.mustar = m
	return vpty2
}

func NewPort() *Port {
	// var inff []byte
	// var outff []byte

	// inbuff := bytes.NewBuffer(inff)
	// outbuff := bytes.NewBuffer(outff)

	// inr := bufio.NewReader(inbuff)
	// inw := bufio.NewWriter(inbuff)

	// outr := bufio.NewReader(outbuff)
	// outw := bufio.NewWriter(outbuff)

	inr, inw := NewPipe()
	outr, outw := NewPipe()

	return &Port{
		in_r:  inr,
		in_w:  inw,
		out_r: outr,
		out_w: outw,
	}
}

type Ldisc interface {
	SetMustarFD(io.ReadWriteCloser)
	SetSluvaFD(io.ReadWriteCloser)
	// Call it to disable this ldisc
	Close() error
	// interface to control ldisc, WIP
	IOctl()
	// To be called from vpty on
	// mustar/sluva writes
	ReceiveMustarBuff(int)
	ReceiveSluvaBuff(int)
}

type WriteCloserTicker struct {
	w      io.WriteCloser
	ticker func(n int)
}

func (t *WriteCloserTicker) Write(p []byte) (n int, err error) {
	// fmt.Println("write tick")
	n, err = t.w.Write(p)
	t.ticker(n)
	return n, err
}

func (t *WriteCloserTicker) Close() error {
	return t.w.Close()
}

type PortFD struct {
	name string
	w    io.WriteCloser
	r    io.ReadCloser
}

func (fd *PortFD) Read(p []byte) (n int, err error) {
	// fmt.Println("fd read", fd.name)
	return fd.r.Read(p)
}

func (fd *PortFD) Write(p []byte) (n int, err error) {
	// fmt.Println("fd write", fd.name)
	return fd.w.Write(p)
}

func (fd *PortFD) Close() error {
	err := fd.r.Close()
	err2 := fd.w.Close()
	return errors.Join(err, err2)
}

type Port struct {
	// commonly used internally
	// ie: bring it to ldisc
	in_r  io.ReadCloser  // <- in_w
	out_w io.WriteCloser // -> out_r

	// commonly used externally
	// ie: bring it to app or term
	in_w  io.WriteCloser // -> in_r
	out_r io.ReadCloser  // <- out_w
}

type VPTY2 struct {
	mu sync.Mutex
	// in theory, sluva should translate in/out
	sluva  *Port
	ldisc  Ldisc
	mustar *Port
	errors []error
}

func (v *VPTY2) sluvaWriteTick(n int) {
	v.ldisc.ReceiveSluvaBuff(n)
}

func (v *VPTY2) mustarWriteTick(n int) {
	v.ldisc.ReceiveMustarBuff(n)
}

func (v *VPTY2) SetLDisc(ldisc Ldisc) {
	// TODO: disable previous ldisc
	v.ldisc = ldisc
	mfd := &PortFD{
		name: "Mustar FD for ldisc (mustar in_r out_w)",
		r:    v.mustar.in_r,  // <- v.mustar.in_w
		w:    v.mustar.out_w, // -> v.mustar.out_r
	}
	sfd := &PortFD{
		name: "Sluva FD for ldisc (sluva in_r out_w)",
		r:    v.sluva.in_r, //
		w:    v.sluva.out_w,
	}
	ldisc.SetMustarFD(mfd)
	ldisc.SetSluvaFD(sfd)
	// fmt.Println("setted ldisc")
}

// SluvaFD file descriptor for shell/app
func (v *VPTY2) SluvaFD() *PortFD {
	return &PortFD{
		name: "Sluva FD for Shell/app (sluva out_r in_w)",
		r:    v.sluva.out_r, // <- v.sluva.out_w
		w:    v.sluva.in_w,  // -> v.sluva.in_r
	}
}

// MustarFD file descriptor for term
// write here keyboard input (os.Stdin)
// read from here the term data (and write
// to os.Stdout)
func (v *VPTY2) MustarFD() *PortFD {
	return &PortFD{
		name: "Mustar FD for TERM (mustar out_r in_w)",
		r:    v.mustar.out_r, // <- v.mustar.out_w
		w:    v.mustar.in_w,  // -> v.mustar.in_r
	}
}

func (v *VPTY2) addErr(err error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.errors = append(v.errors, err)
}
