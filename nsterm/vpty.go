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
	"errors"
	"fmt"
	"io"
	"sync"
)

func NewFD(name string, r io.ReadCloser, w io.WriteCloser) io.ReadWriteCloser {
	return &PortFD{
		name: name,
		r:    r,
		w:    w,
	}
}

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
	vpty2 := &VPTY2{
		sluvas:  make(map[*Port]bool),
		mustars: make(map[*Port]bool),
	}

	// s := NewPort()
	// // wrap in_w to call vpty ticker on every in_w write
	// // writes comes from shell/app
	// swticker := &WriteCloserTicker{w: s.in_w, ticker: vpty2.sluvaWriteTick}
	// s.in_w = swticker

	// m := NewPort()
	// // wrap in_w to call vpty ticker on every in_w write
	// // writes comes from term
	// mwticker := &WriteCloserTicker{w: m.in_w, ticker: vpty2.mustarWriteTick}
	// m.in_w = mwticker

	vpty2.NewSluvaPort()
	vpty2.NewMustarPort()

	return vpty2
}

// func NewPort() *Port {
// 	// var inff []byte
// 	// var outff []byte

// 	// inbuff := bytes.NewBuffer(inff)
// 	// outbuff := bytes.NewBuffer(outff)

// 	// inr := bufio.NewReader(inbuff)
// 	// inw := bufio.NewWriter(inbuff)

// 	// outr := bufio.NewReader(outbuff)
// 	// outw := bufio.NewWriter(outbuff)

// 	// in pipe
// 	inr, inw := NewPipe()
// 	// out pipe
// 	outr, outw := NewPipe()

// 	// mst / slv   mst / slv
// 	// outFD     | inFD
// 	// commonly  | commonly
// 	// to app    | to ldisc
// 	//           |
// 	// in_w   ---|--> in_r
// 	// out_r  <--|--- out_w
// 	//           |
// 	return &Port{
// 		inFD: &PortFD{
// 			r: inr,
// 			w: outw, // <-- writes here
// 		},
// 		outFD: &PortFD{
// 			r: outr, // <-- can readed here
// 			w: inw,
// 		},
// 		// in pipe
// 		// in_r: inr,
// 		// in_w: inw,
// 		// out pipe
// 		// out_r: outr,
// 		// out_w: outw,
// 	}
// }

type Ldisc interface {
	SetMustarFD(io.ReadWriteCloser)
	SetSluvaFD(io.ReadWriteCloser)
	// Call it to disable this ldisc
	Close() error
	// interface to control ldisc, WIP
	IOctl()
	// To be called from vpty on
	// mustar/sluva writes
	ReceiveMustarBuff(int, *PortFD)
	ReceiveSluvaBuff(int, *PortFD)
	//
	ReadRuneBuff() []rune
	SetBuff(s string)
	GetESC() chan string
}

type WriteCloserTicker struct {
	w io.WriteCloser // write part of pipe
	// r      io.ReadCloser
	infd   *PortFD
	ticker func(n int, p *PortFD)
}

func (t *WriteCloserTicker) Write(p []byte) (n int, err error) {
	// fmt.Println("write tick", bytes.TrimRight(p, "\x00"))
	n, err = t.w.Write(p)
	t.ticker(n, t.infd)
	return n, err
}

func (t *WriteCloserTicker) Close() error {
	return t.w.Close()
}

// PortFD the internal or external
// part of the two pipes of *Port
type PortFD struct {
	name string
	w    io.WriteCloser
	r    io.ReadCloser
}

func (fd *PortFD) Read(p []byte) (n int, err error) {
	return fd.r.Read(p)
}

func (fd *PortFD) Write(p []byte) (n int, err error) {
	return fd.w.Write(p)
}

func (fd *PortFD) Close() error {
	err := fd.r.Close()
	err2 := fd.w.Close()
	return errors.Join(err, err2)
}

func (fd *PortFD) GetRawR() io.ReadCloser {
	return fd.r
}

func (fd *PortFD) GetRawW() io.WriteCloser {
	return fd.w
}

// Port a pair of pipes
type Port struct {
	// commonly used internally
	// ie: bring it to ldisc
	inFD *PortFD
	// in_r  io.ReadCloser  // <- in_w
	// out_w io.WriteCloser // -> out_r

	// commonly used externally
	// ie: bring it to app or term
	outFD *PortFD
	// in_w  io.WriteCloser // -> in_r
	// out_r io.ReadCloser  // <- out_w
}

func (p *Port) InFD() *PortFD {
	return p.inFD
}

func (p *Port) OutFD() *PortFD {
	return p.outFD
}

type VPTY2 struct {
	// in theory, sluva should translate in/out
	// sluva represent a port, pair of pipes
	mu      sync.Mutex
	sluva   *Port
	ldisc   Ldisc
	mustar  *Port
	sluvas  map[*Port]bool
	mustars map[*Port]bool
	errors  []error
	// external FD
	// sluvaFD  *PortFD
	// mustarFD *PortFD
}

func (v *VPTY2) _newPort() *Port {

	// in pipe
	inr, inw := NewPipe()
	// out pipe
	outr, outw := NewPipe()

	// inFD := &PortFD{
	// 	r: inr,
	// 	w: outw, // <-- writes here
	// }
	// outFD := &PortFD{
	// 	r: outr, // <-- can readed here
	// 	w: inw,
	// }

	// port := &Port{
	// 	inFD:  inFD,
	// 	outFD: outFD,
	// }

	return &Port{
		inFD: &PortFD{
			r: inr,
			w: outw, // <-- writes here
		},
		outFD: &PortFD{
			r: outr, // <-- can readed here
			w: inw,
		},
	}
	// 	// in pipe
	// 	// in_r: inr,
	// 	// in_w: inw,
	// 	// out pipe
	// 	// out_r: outr,
	// 	// out_w: outw,
	// }
}

func (v *VPTY2) _newTickerPort() *Port {
	port := v._newPort()
	// replace out fd writer
	// with ticker so  a write
	// to outFD will tick the
	// inFD to vpty
	wticker := &WriteCloserTicker{
		w:      port.outFD.w,
		infd:   port.inFD,
		ticker: nil,
	}
	port.outFD.w = wticker
	return port
}

// NewSluvaPort create new sluva port, it
// will stored
func (v *VPTY2) NewSluvaPort() *Port {
	v.mu.Lock()
	defer v.mu.Unlock()
	port := v._newTickerPort()
	port.outFD.w.(*WriteCloserTicker).ticker = v.sluvaWriteTick
	port.inFD.name = "Sluva inFD for ldisc (sluva in_r out_w)"
	port.outFD.name = "Sluva outFD for Shell/app (sluva out_r in_w)"
	v.sluvas[port] = true
	if v.sluva == nil {
		v.sluva = port
	}
	return port
}

func (v *VPTY2) NewMustarPort() *Port {
	v.mu.Lock()
	defer v.mu.Unlock()
	port := v._newTickerPort()
	port.outFD.w.(*WriteCloserTicker).ticker = v.mustarWriteTick
	port.inFD.name = "Mustar inFD for ldisc (mustar in_r out_w)"
	port.outFD.name = "Mustar outFD for TERM (mustar out_r in_w)"
	v.mustars[port] = true
	if v.mustar == nil {
		v.mustar = port
	}
	return port
}

func (v *VPTY2) sluvaWriteTick(n int, inFD *PortFD) {
	v.ldisc.ReceiveSluvaBuff(n, inFD)
}

func (v *VPTY2) mustarWriteTick(n int, inFD *PortFD) {
	v.ldisc.ReceiveMustarBuff(n, inFD)
}

func (v *VPTY2) CursorSluva(p *Port) error {
	if _, exists := v.sluvas[p]; !exists {
		return fmt.Errorf("unknown port")
	}
	v.sluva = p
	v.ldisc.SetSluvaFD(p.inFD)
	return nil
}

func (v *VPTY2) CursorMustar(p *Port) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, exists := v.mustars[p]; !exists {
		return fmt.Errorf("unknown port")
	}
	v.mustar = p
	v.ldisc.SetMustarFD(p.inFD)
	return nil
}

func (v *VPTY2) DestroyPort(p *Port) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	var portset map[*Port]bool
	var pport **Port
	if _, exists := v.mustars[p]; exists {
		portset = v.mustars
		pport = &v.mustar
	} else if _, exists := v.sluvas[p]; exists {
		portset = v.sluvas
		pport = &v.sluva
	} else {
		return fmt.Errorf("unknown port")
	}

	if len(portset) == 1 {
		return fmt.Errorf("cannot destry the last mustar port")
	}

	delete(portset, p)
	if *pport == p {
		for np := range portset {
			*pport = np
			break
		}
	}
	return nil
}

func (v *VPTY2) SetLDisc(ldisc Ldisc) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ldisc != nil {
		err := v.ldisc.Close()
		if err != nil {
			v.errors = append(v.errors, err)
		}
	}

	mInFD := v.mustar.inFD
	sInFD := v.sluva.inFD

	// notice to ldisc all mustars and sluvas
	for p := range v.mustars {
		p.inFD.name = "Mustar inFD for ldisc (mustar in_r out_w)"
		ldisc.SetMustarFD(p.inFD)
	}

	for p := range v.sluvas {
		p.inFD.name = "Sluva inFD for ldisc (sluva in_r out_w)"
		ldisc.SetSluvaFD(p.inFD)
	}

	// finally set the current mustar and sluva
	ldisc.SetMustarFD(mInFD)
	ldisc.SetSluvaFD(sInFD)

	v.ldisc = ldisc
}

func (v *VPTY2) GetLDisc() (ldisc Ldisc) {
	return v.ldisc
}

// SluvaFD file descriptor for shell/app
func (v *VPTY2) SluvaFD() *PortFD {
	sOutFD := v.sluva.outFD
	sOutFD.name = "Sluva outFD for Shell/app (sluva out_r in_w)"
	return sOutFD
	// return &PortFD{
	// 	name: "Sluva FD for Shell/app (sluva out_r in_w)",
	// 	r:    v.sluva.out_r, // pipe <- v.sluva.out_w
	// 	w:    v.sluva.in_w,  // pipe -> v.sluva.in_r
	// }
}

// MustarFD file descriptor for term
// write here keyboard input (os.Stdin)
// read from here the term data (and write
// to os.Stdout)
func (v *VPTY2) MustarFD() *PortFD {
	mOutFD := v.mustar.outFD
	mOutFD.name = "Mustar outFD for TERM (mustar out_r in_w)"
	return mOutFD
	// return &PortFD{
	// 	name: "Mustar FD for TERM (mustar out_r in_w)",
	// 	r:    v.mustar.out_r, // <- v.mustar.out_w
	// 	w:    v.mustar.in_w,  // -> v.mustar.in_r
	// }
}

func (v *VPTY2) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	for slv := range v.sluvas {
		slv.inFD.Close()
		slv.outFD.Close()
	}
	for mst := range v.mustars {
		mst.inFD.Close()
		mst.outFD.Close()
	}
	return v.ldisc.Close()
}

// func (v *VPTY2) addErr(err error) {
// 	// v.mu.Lock()
// 	// defer v.mu.Unlock()
// 	v.errors = append(v.errors, err)
// }
