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
				// do nothing
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

func ParseKeyboard(data []byte) []byte {
	r, _ := utf8.DecodeRune(data)
	if r == 13 {
		return []byte("\r\n")
	}
	if r == 127 {
		return []byte{8, 127}
	}
	return data
}

// type inTranslator struct {
// 	on bool
// 	w  io.WriteCloser
// }

// type outTranslator struct {
// 	on bool
// 	w  io.WriteCloser
// }

// func (i *inTranslator) Close() error {
// 	return i.w.Close()
// }

// func (i *inTranslator) CloseWithError(err error) error {
// 	if p, ok := i.w.(*io.PipeWriter); ok {
// 		return p.CloseWithError(err)
// 	}
// 	return nil
// }

// func (i *inTranslator) Write(data []byte) (int, error) {
// 	if !i.on {
// 		return i.w.Write(data)
// 	}
// 	//
// 	r, _ := utf8.DecodeRune(data)
// 	p := data
// 	e := 0
// 	if r == 13 {
// 		p = []byte("\r\n")
// 		//
// 		e++
// 	}
// 	if r == 127 {
// 		p = []byte{8, 127}
// 		e++
// 	}
// 	_n, err := i.w.Write(p)
// 	return _n - e, err
// }

// func (o *outTranslator) Close() error {
// 	return o.w.Close()
// }

// func (o *outTranslator) CloseWithError(err error) error {
// 	if p, ok := o.w.(*io.PipeWriter); ok {
// 		return p.CloseWithError(err)
// 	}
// 	return nil
// }

// func (o *outTranslator) Write(data []byte) (n int, err error) {
// 	if !o.on {
// 		return o.w.Write(data)
// 	}
// 	p := LFtoCRLF(data)
// 	// for debug:
// 	// p = append([]byte("; "), p...)
// 	return o.w.Write(p)
// }

// type port struct {
// 	in  io.ReadCloser
// 	out io.WriteCloser
// }

// func (p *port) Set(in io.ReadCloser, out io.WriteCloser) {
// 	p.in = in
// 	p.out = out
// }

// type VPTY struct {
// 	mu sync.Mutex
// 	// in theory, sluva should translate in/out
// 	sluva  *port
// 	ldisc  LDisc
// 	mustar *port
// 	errors []error
// }

// func (v *VPTY) SetLDisc(ldisc LDisc) {
// 	v.ldisc = ldisc
// }

// func (v *VPTY) addErr(err error) {
// 	v.mu.Lock()
// 	defer v.mu.Unlock()
// 	v.errors = append(v.errors, err)
// }

// // OpenSluva func
// //
// //	term            mr   pty  sl     shell
// //
// // ┌──────────┐     ┌──────────┐     ┌──────────┐
// // │       in ◄──┬──►in     in ◄──┬──►in        │
// // │          │  │  │          │  │  │          │
// // │          │  │  │          │  │  │          │
// // │       out├──┴──┤out    out├──┴──┤out       │
// // └──────────┘     └──────────┘     └──────────┘
// func (v *VPTY) OpenSluva(raw bool) (stdin io.ReadCloser, stdout io.WriteCloser) {
// 	stdinr, stdinw := io.Pipe()
// 	stdoutr, stdoutw := io.Pipe()

// 	// in translator (from term), traduce \r\n to \n
// 	trnsl_stdinw := &inTranslator{w: stdinw, on: true}
// 	// out translator (to term), traduce \n to \r\n
// 	trnsl_stdoutw := &outTranslator{w: stdoutw, on: true}

// 	v.sluva.Set(stdoutr, trnsl_stdinw)
// 	if raw {
// 		go v.startRaw()
// 	} else {
// 		go v.startLDisc()
// 	}
// 	// in/out to use in shell
// 	return stdinr, trnsl_stdoutw
// }

// func (v *VPTY) startRaw() {
// 	go func() {
// // 		// fmt.Println("copying sluva out to mustar in")
// 		_, err := io.Copy(v.sluva.out, v.mustar.in)
// 		if err != nil {
// 			v.addErr(err)
// 		}
// 		// fmt.Println("end of copying sluva out to mustar in")
// 	}()
// 	// fmt.Println("copying mustar out to sluva in")
// 	_, err := io.Copy(v.mustar.out, v.sluva.in)
// 	if err != nil {
// 		v.addErr(err)
// 	}
// 	// fmt.Println("end of copying mustar out to sluva in")
// 	defer v.mustar.in.Close()
// 	defer v.sluva.in.Close()
// }

// func (v *VPTY) startLDisc() {
// 	go func() {
// 		_, err := io.Copy(v.mustar.out, v.sluva.in)
// 		if err != nil {
// 			v.addErr(err)
// 		}
// 	}()

// 	// prompt buff
// 	var bff []byte
// 	buff := bytes.NewBuffer(bff)

// 	reader := bufio.NewReader(v.mustar.in)

// 	_c := make(chan rune, 10)
// 	_e := make(chan error, 10)
// 	go func() {
// 		for {
// 			ccc, _, eee := reader.ReadRune()
// 			if eee != nil {
// 				_e <- eee
// 				continue
// 			}
// 			_c <- ccc
// 		}
// 	}()

// 	// esc_tim := time.Now()
// 	for {
// 		// char, _, err := reader.ReadRune()
// 		// if err != nil {
// 		// 	v.mustar.out.Write([]byte(errors.Join(fmt.Errorf("term read err"), err).Error()))
// 		// }

// 		select {
// 		case char := <-_c:

// 			// ESC sequence
// 			if char == 27 {
// 				// collect
// 				time.Sleep(512 * time.Microsecond)
// 				esq_seq_size := len(_c)
// 				if esq_seq_size > 0 {
// 					esc_seq := []rune{27}
// 					for i := 0; i < esq_seq_size; i++ {
// 						esc_seq = append(esc_seq, <-_c)
// 					}
// 					esc_seq = append(esc_seq, []rune("\n")...)
// 					v.sluva.out.Write([]byte(string(esc_seq)))
// 					continue
// 				}
// 				v.sluva.out.Write([]byte(string(char)))
// 				continue
// 			}

// 			// del
// 			if char == 127 {
// 				if buff.Len() <= 0 {
// 					continue
// 				}
// 				buff.Truncate(buff.Len() - 1)
// 				v.mustar.out.Write([]byte("\b \b"))
// 				continue
// 			}

// 			// intro
// 			if char == 13 {
// 				v.mustar.out.Write([]byte("\r\n"))
// 				buff.Write([]byte("\n"))
// 				_, err := v.sluva.out.Write(buff.Bytes())
// 				if err != nil {
// 					v.mustar.out.Write([]byte(errors.Join(fmt.Errorf("app write err"), err).Error()))
// 				}
// 				buff.Reset()
// 				continue
// 			}

// 			buff.WriteRune(char)
// 			v.mustar.out.Write([]byte(string(char)))

// 		case _err := <-_e:
// 			v.mustar.out.Write([]byte(errors.Join(fmt.Errorf("term read err"), _err).Error()))
// 		default:
// 			<-time.After(100 * time.Microsecond)
// 		}

// 	}

// 	// fmt.Println("copying mustar out to sluva in")
// 	// _, err := io.Copy(v.mustar.out, v.sluva.in)
// 	// if err != nil {
// 	// 	v.addErr(err)
// 	// }
// 	// fmt.Println("end of copying mustar out to sluva in")
// 	// defer v.mustar.in.Close()
// 	// defer v.sluva.in.Close()
// }

// func (v *VPTY) OpenMustar(in io.ReadCloser, out io.WriteCloser) {
// 	v.mustar.Set(in, out)
// }

// func (v *VPTY) Close() error {
// 	return errors.Join(
// 		v.mustar.in.Close(),
// 		v.sluva.in.Close(),
// 		v.sluva.out.Close(),
// 		v.mustar.out.Close(),
// 	)
// }

// func NewVirtpty() *VPTY {
// 	return &VPTY{
// 		sluva:  &port{},
// 		mustar: &port{},
// 	}
// }
