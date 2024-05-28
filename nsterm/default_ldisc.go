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
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

var CursorUp string = "\033[A"
var CursorDown string = "\033[B"
var CursorRight string = "\033[C"
var CursorLeft string = "\033[D"
var CursorHome string = "\033[H"
var CursorEnd string = "\033[F"
var LineFeed string = "\n"
var CarriageReturn string = "\r"
var Backspace string = string([]byte{127})
var Delete string = "\033[3~"
var CtrlC string = string([]byte{3})

var DefaultEscSet map[string]string = map[string]string{
	CursorUp:    "",
	CursorDown:  "",
	CursorLeft:  "",
	CursorRight: "",
	Backspace:   "",
	CtrlC:       "",
	CursorHome:  "",
	CursorEnd:   "",
	Delete:      "",
}

func NewDefaultLdisc() *DefaultLdisc {
	return &DefaultLdisc{
		// LineBuff:     make([]byte, 0),
		RuneBuff:     make([]rune, 0),
		ERR:          make(chan error, 10),
		ESC:          make(chan string, 10),
		ECO:          true,
		CursorOffset: 0,
		ESCSet:       DefaultEscSet,
	}
}

type DefaultLdisc struct {
	mInFD io.ReadWriteCloser
	sInFD io.ReadWriteCloser

	errs []error
	ERR  chan error
	// LineBuff     []byte
	RuneBuff     []rune
	CursorOffset int
	ESC          chan string
	ECO          bool
	//
	ESCSet map[string]string
}

func (d *DefaultLdisc) GetESC() chan string {
	return d.ESC
}

func (d *DefaultLdisc) ReadRuneBuff() []rune {
	return d.RuneBuff
}

func (d *DefaultLdisc) SetMustarFD(fd io.ReadWriteCloser) {
	d.mInFD = fd
}

func (d *DefaultLdisc) SetSluvaFD(fd io.ReadWriteCloser) {
	d.sInFD = fd
}

func (d *DefaultLdisc) SetBuff(s string) {
	r := []rune(s)
	d.RuneBuff = r
	d.CursorOffset = len(r)
}

// Called from vpty on mustar port write.
// Read -> process -> Write to sluva FD
func (d *DefaultLdisc) ReceiveMustarBuff(n int, mInFD *PortFD) {

	// _, err := io.CopyN(d.sInFD, d.mustarFD, int64(n))
	// if err != nil {
	// 	d.errs = append(d.errs, err)
	// }
	// fmt.Println("reading from mustar")

	if n == 0 {
		return
	}

	data_b := make([]byte, n)
	n, err := io.ReadFull(mInFD, data_b)
	if err != nil {
		d.errs = append(d.errs, err)
		d.ERR <- err
		return
	}

	// sometimes a term uses maxread to store
	// data and send here mostly empty data_b
	data_b = bytes.TrimRight(data_b, "\x00")

	data_s := string(data_b)
	data_r, _ := utf8.DecodeRune(data_b)

	if n == 0 {
		return
	}

	// handle "\n"
	if data_s == CarriageReturn {
		d.RuneBuff = append(d.RuneBuff, data_r)
		_, err = d.sInFD.Write([]byte(string(d.RuneBuff)))
		if err != nil {
			d.errs = append(d.errs, err)
			d.ERR <- err
		}
		d.ESC <- data_s
		return
	}

	// handle knowed ESC sequences
	if _, exists := d.ESCSet[data_s]; exists {
		switch data_s {
		case CursorLeft:
			if d.CursorOffset > 0 {
				d.CursorOffset--
				// eco cursor left
				mInFD.Write(data_b)
			}
		case CursorRight:
			if d.CursorOffset < len(d.RuneBuff) {
				d.CursorOffset++
				// eco cursor right
				mInFD.Write(data_b)
			}
		case CursorHome:
			if d.CursorOffset == 0 {
				// do nothing
				break
			}
			mInFD.Write(bytes.Repeat([]byte(CursorLeft), d.CursorOffset))
			d.CursorOffset = 0
		case CursorEnd:
			if d.CursorOffset == len(d.RuneBuff) {
				// do nothing
				break
			}
			mInFD.Write(bytes.Repeat([]byte(CursorRight), len(d.RuneBuff)-d.CursorOffset))
			d.CursorOffset = len(d.RuneBuff)
		case Delete:
			if d.CursorOffset == len(d.RuneBuff) {
				// do nothing (TODO: do something xD)
			}
		case Backspace:
			if d.CursorOffset == 0 {
				// do nothing
				break
			}

			// line edit
			if d.CursorOffset < len(d.RuneBuff) {
				nbr := make([]rune, 0)
				nbr = append(nbr, d.RuneBuff[0:d.CursorOffset-1]...)
				nbr = append(nbr, d.RuneBuff[d.CursorOffset:]...)

				d.RuneBuff = nbr
				d.CursorOffset--

				// eco line edit
				mInFD.Write([]byte("\b"))
				mInFD.Write([]byte(string(d.RuneBuff[d.CursorOffset:])))
				mInFD.Write([]byte(" \b"))
				mInFD.Write(bytes.Repeat([]byte(CursorLeft), len(nbr[d.CursorOffset:])))

				// fmt.Println("\r\n", len(d.RuneBuff), string(d.RuneBuff))
				break
			}

			// cursor at end, rm last rune
			d.RuneBuff = d.RuneBuff[0 : len(d.RuneBuff)-1]
			d.CursorOffset--
			// eco cursor right
			mInFD.Write([]byte("\b \b"))
		case CtrlC:
			mInFD.Write([]byte("^C"))
		}
		d.ESC <- data_s
		return
	}

	// handle not knowed sequences
	if data_b[0] == 27 {
		fmt.Println(data_b, string(data_b[1:]))
		d.ESC <- data_s
		return
	}

	// if len(data_b) > 1 {
	// 	// probably out-ascii char
	// 	r, _ := utf8.DecodeRune(data_b)
	// 	fmt.Println("holi", data_b, string(data_b), r, string(r))
	// }

	// write to (line edit)
	if d.CursorOffset < len(d.RuneBuff) {
		rr := []rune(string(data_b))
		nbr := make([]rune, 0)
		nbr = append(nbr, d.RuneBuff[0:d.CursorOffset]...)
		nbr = append(nbr, rr...)
		nbr = append(nbr, d.RuneBuff[d.CursorOffset:]...)
		d.RuneBuff = nbr
		d.CursorOffset = d.CursorOffset + len(rr)
		//eco
		mInFD.Write(data_b)
		// complete line
		mInFD.Write([]byte(string(nbr[d.CursorOffset:])))
		// bring term cursor back after line completion
		mInFD.Write(bytes.Repeat([]byte(CursorLeft), len(nbr[d.CursorOffset:])))
		// mInFD.Write([]byte("-"))
		return
	}

	rr := []rune(string(data_b))
	d.RuneBuff = append(d.RuneBuff, rr...)
	d.CursorOffset = d.CursorOffset + len(data_b)
	// eco
	mInFD.Write(data_b)

	/// byte version:

	// write to linebuff
	// if d.CursorOffset < len(d.LineBuff) {
	// 	nbf := make([]byte, 0)
	// 	nbf = append(nbf, d.LineBuff[0:d.CursorOffset]...)
	// 	nbf = append(nbf, data_b...)
	// 	nbf = append(nbf, d.LineBuff[d.CursorOffset:]...)

	// 	// fmt.Println("\r\n")
	// 	// fmt.Println("concatenate", string(d.LineBuff[0:d.CursorOffset]), "with", string(data_b), "with", string(d.LineBuff[d.CursorOffset:]))
	// 	// fmt.Println("resulting in", nbf)

	// 	d.LineBuff = nbf
	// 	d.CursorOffset = d.CursorOffset + len(data_b)
	// 	//eco
	// 	mInFD.Write(data_b)
	// 	// complete line
	// 	mInFD.Write(nbf[d.CursorOffset:])
	// 	// bring term cursor back after line completion
	// 	mInFD.Write(bytes.Repeat([]byte(CursorLeft), len(nbf[d.CursorOffset:])))
	// 	// mInFD.Write([]byte("-"))
	// 	return
	// }

	// d.LineBuff = append(d.LineBuff, data_b...)
	// d.CursorOffset = d.CursorOffset + len(data_b)
	// // eco
	// mInFD.Write(data_b)

}

func (d *DefaultLdisc) ReceiveSluvaBuff(n int, sInFD *PortFD) {
	// fmt.Println("reading from sluva")

	data := make([]byte, n)
	_, err := io.ReadFull(sInFD, data)
	if err != nil {
		d.errs = append(d.errs, err)
		return
	}
	data = LFtoCRLF(data)
	// data = append([]byte("d;"), data...)
	_, err = d.mInFD.Write(data)
	if err != nil {
		d.errs = append(d.errs, err)
	}
}

func (d *DefaultLdisc) IOctl() {

}

func (d *DefaultLdisc) Close() error {
	return nil
}

// commonly from app to term (from slave to master)
func LFtoCRLF(data []byte) []byte {
	p := data
	m := bytes.Count(data, []byte{10})
	if m > 0 {
		l := len(data)
		p = make([]byte, l+m)
		e := -1
		for i := 0; i < len(data); i++ {
			e++
			p[e] = data[i]
			if data[i] != 10 {
				continue
			}
			// data_i == 10
			if i == 0 || (i > 0 && data[i-1] != 13) {
				p[e] = 13
				e++
				p[e] = 10
				continue
			}
		}
		p = p[:e+1]
	}
	return p
}
