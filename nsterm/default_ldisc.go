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
	mustarFD io.ReadWriteCloser
	sluvaFD  io.ReadWriteCloser
	errs     []error
	ERR      chan error
	// LineBuff     []byte
	RuneBuff     []rune
	CursorOffset int
	ESC          chan string
	ECO          bool
	//
	ESCSet map[string]string
}

func (d *DefaultLdisc) SetMustarFD(fd io.ReadWriteCloser) {
	d.mustarFD = fd
}

func (d *DefaultLdisc) SetSluvaFD(fd io.ReadWriteCloser) {
	d.sluvaFD = fd
}

func (d *DefaultLdisc) SetBuff(s string) {
	r := []rune(s)
	d.RuneBuff = r
	d.CursorOffset = len(r)
}

// Called from vpty on mustar port write.
// Read -> process -> Write to sluva FD
func (d *DefaultLdisc) ReceiveMustarBuff(n int) {
	// _, err := io.CopyN(d.sluvaFD, d.mustarFD, int64(n))
	// if err != nil {
	// 	d.errs = append(d.errs, err)
	// }
	// fmt.Println("reading from mustar")

	if n == 0 {
		return
	}

	data_b := make([]byte, n)
	n, err := io.ReadFull(d.mustarFD, data_b)
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
		_, err = d.sluvaFD.Write([]byte(string(d.RuneBuff)))
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
				d.mustarFD.Write(data_b)
			}
		case CursorRight:
			if d.CursorOffset < len(d.RuneBuff) {
				d.CursorOffset++
				// eco cursor right
				d.mustarFD.Write(data_b)
			}
		case CursorHome:
			if d.CursorOffset == 0 {
				// do nothing
				break
			}
			d.mustarFD.Write(bytes.Repeat([]byte(CursorLeft), d.CursorOffset))
			d.CursorOffset = 0
		case CursorEnd:
			if d.CursorOffset == len(d.RuneBuff) {
				// do nothing
				break
			}
			d.mustarFD.Write(bytes.Repeat([]byte(CursorRight), len(d.RuneBuff)-d.CursorOffset))
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
				d.mustarFD.Write([]byte("\b"))
				d.mustarFD.Write([]byte(string(d.RuneBuff[d.CursorOffset:])))
				d.mustarFD.Write([]byte(" \b"))
				d.mustarFD.Write(bytes.Repeat([]byte(CursorLeft), len(nbr[d.CursorOffset:])))

				// fmt.Println("\r\n", len(d.RuneBuff), string(d.RuneBuff))
				break
			}

			// cursor at end, rm last rune
			d.RuneBuff = d.RuneBuff[0 : len(d.RuneBuff)-1]
			d.CursorOffset--
			// eco cursor right
			d.mustarFD.Write([]byte("\b \b"))
		case CtrlC:
			d.mustarFD.Write([]byte("^C"))
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
		nbr := make([]rune, 0)
		nbr = append(nbr, d.RuneBuff[0:d.CursorOffset]...)
		nbr = append(nbr, data_r)
		nbr = append(nbr, d.RuneBuff[d.CursorOffset:]...)
		d.RuneBuff = nbr
		d.CursorOffset = d.CursorOffset + 1 // data_r len == 1
		//eco
		d.mustarFD.Write(data_b)
		// complete line
		d.mustarFD.Write([]byte(string(nbr[d.CursorOffset:])))
		// bring term cursor back after line completion
		d.mustarFD.Write(bytes.Repeat([]byte(CursorLeft), len(nbr[d.CursorOffset:])))
		// d.mustarFD.Write([]byte("-"))
		return
	}

	d.RuneBuff = append(d.RuneBuff, data_r)
	d.CursorOffset = d.CursorOffset + len(data_b)
	// eco
	d.mustarFD.Write(data_b)

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
	// 	d.mustarFD.Write(data_b)
	// 	// complete line
	// 	d.mustarFD.Write(nbf[d.CursorOffset:])
	// 	// bring term cursor back after line completion
	// 	d.mustarFD.Write(bytes.Repeat([]byte(CursorLeft), len(nbf[d.CursorOffset:])))
	// 	// d.mustarFD.Write([]byte("-"))
	// 	return
	// }

	// d.LineBuff = append(d.LineBuff, data_b...)
	// d.CursorOffset = d.CursorOffset + len(data_b)
	// // eco
	// d.mustarFD.Write(data_b)

}

func (d *DefaultLdisc) ReceiveSluvaBuff(n int) {
	// fmt.Println("reading from sluva")

	data := make([]byte, n)
	_, err := io.ReadFull(d.sluvaFD, data)
	if err != nil {
		d.errs = append(d.errs, err)
		return
	}
	data = LFtoCRLF(data)
	// data = append([]byte("d;"), data...)
	_, err = d.mustarFD.Write(data)
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
