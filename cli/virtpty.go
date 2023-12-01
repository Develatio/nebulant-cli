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

package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
	"unicode/utf8"
)

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

type inTranslator struct {
	on bool
	w  io.WriteCloser
}

type outTranslator struct {
	on bool
	w  io.WriteCloser
}

func (i *inTranslator) Close() error {
	return i.w.Close()
}

func (i *inTranslator) CloseWithError(err error) error {
	if p, ok := i.w.(*io.PipeWriter); ok {
		return p.CloseWithError(err)
	}
	return nil
}

func (i *inTranslator) Write(data []byte) (int, error) {
	if !i.on {
		return i.w.Write(data)
	}
	//
	r, _ := utf8.DecodeRune(data)
	p := data
	e := 0
	if r == 13 {
		p = []byte("\r\n")
		//
		e++
	}
	if r == 127 {
		p = []byte{8, 127}
		e++
	}
	_n, err := i.w.Write(p)
	return _n - e, err
}

func (o *outTranslator) Close() error {
	return o.w.Close()
}

func (o *outTranslator) CloseWithError(err error) error {
	if p, ok := o.w.(*io.PipeWriter); ok {
		return p.CloseWithError(err)
	}
	return nil
}

func (o *outTranslator) Write(data []byte) (n int, err error) {
	if !o.on {
		return o.w.Write(data)
	}
	p := LFtoCRLF(data)
	// for debug:
	// p = append([]byte("; "), p...)
	return o.w.Write(p)
}

type port struct {
	in  io.ReadCloser
	out io.WriteCloser
}

func (p *port) Set(in io.ReadCloser, out io.WriteCloser) {
	p.in = in
	p.out = out
}

type VPTY struct {
	mu sync.Mutex
	// in theory, sluva should translate in/out
	sluva  *port
	mustar *port
	errors []error
}

func (v *VPTY) addErr(err error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.errors = append(v.errors, err)
}

// OpenSluva func
//
//	term            mr   pty  sl     shell
//
// ┌──────────┐     ┌──────────┐     ┌──────────┐
// │       in ◄──┬──►in     in ◄──┬──►in        │
// │          │  │  │          │  │  │          │
// │          │  │  │          │  │  │          │
// │       out├──┴──┤out    out├──┴──┤out       │
// └──────────┘     └──────────┘     └──────────┘
func (v *VPTY) OpenSluva(raw bool) (stdin io.ReadCloser, stdout io.WriteCloser) {
	stdinr, stdinw := io.Pipe()
	stdoutr, stdoutw := io.Pipe()

	// in translator (from term), traduce \r\n to \n
	trnsl_stdinw := &inTranslator{w: stdinw, on: true}
	// out translator (to term), traduce \n to \r\n
	trnsl_stdoutw := &outTranslator{w: stdoutw, on: true}

	v.sluva.Set(stdoutr, trnsl_stdinw)
	if raw {
		go v.startRaw()
	} else {
		go v.startLDisc()
	}
	// in/out to use in shell
	return stdinr, trnsl_stdoutw
}

func (v *VPTY) startRaw() {
	go func() {
		// fmt.Println("copying sluva out to mustar in")
		_, err := io.Copy(v.sluva.out, v.mustar.in)
		if err != nil {
			v.addErr(err)
		}
		// fmt.Println("end of copying sluva out to mustar in")
	}()
	// fmt.Println("copying mustar out to sluva in")
	_, err := io.Copy(v.mustar.out, v.sluva.in)
	if err != nil {
		v.addErr(err)
	}
	// fmt.Println("end of copying mustar out to sluva in")
	defer v.mustar.in.Close()
	defer v.sluva.in.Close()
}

// func escCollect(esc_time time.Time, reader *bufio.Reader) ([]rune, rune, error) {
// 	var esc_seq []rune = make([]rune, 0)
// 	// first read should be 91 [
// 	char, _, err := reader.ReadRune()
// 	if err != nil {
// 		return nil, char, err
// 	}
// 	if char == 91 && time.Since(esc_time).Microseconds() < 512 {
// 		esc_seq = append(esc_seq, char)
// 	} else {
// 		return esc_seq, char, nil
// 	}

// 	for {
// 		char, _, err := reader.ReadRune()
// 		if err != nil {
// 			return nil, char, err
// 		}
// 		if time.Since(esc_time).Microseconds() < 512 {
// 			esc_seq = append(esc_seq, char)
// 		} else {
// 			return esc_seq, char, nil
// 		}
// 	}
// }

// func escCollect(esc_time time.Time, reader *bufio.Reader) ([]rune, rune, error) {
// 	reader.Peek()
// 	var esc_seq []rune = make([]rune, 0)
// 	select {
// 	case char, _, err := <-reader.ReadRune():
// 		//
// 	default:
// 	}
// }

func (v *VPTY) groundEscParse(esc_seq []rune) []rune {
	switch string(esc_seq) {
	case "[?25l":
		// make cursor invisible
		// do nothing
		return make([]rune, 0)
	default:
		return esc_seq
	}
}

func (v *VPTY) startLDisc() {
	go func() {
		_, err := io.Copy(v.mustar.out, v.sluva.in)
		if err != nil {
			v.addErr(err)
		}
	}()

	var bff []byte
	buff := bytes.NewBuffer(bff)

	reader := bufio.NewReader(v.mustar.in)

	_c := make(chan rune, 10)
	_e := make(chan error, 10)
	go func() {
		for {
			ccc, _, eee := reader.ReadRune()
			if eee != nil {
				_e <- eee
				continue
			}
			_c <- ccc
		}
	}()

	// esc_tim := time.Now()
	for {
		// char, _, err := reader.ReadRune()
		// if err != nil {
		// 	v.mustar.out.Write([]byte(errors.Join(fmt.Errorf("term read err"), err).Error()))
		// }

		select {
		case char := <-_c:

			// ESC sequence
			if char == 27 {
				// collect
				time.Sleep(512 * time.Microsecond)
				esq_seq_size := len(_c)
				if esq_seq_size > 0 {
					esc_seq := []rune{27}
					for i := 0; i < esq_seq_size; i++ {
						esc_seq = append(esc_seq, <-_c)
					}
					esc_seq = append(esc_seq, []rune("\n")...)
					v.sluva.out.Write([]byte(string(esc_seq)))
					continue
				}
				v.sluva.out.Write([]byte(string(char)))
				continue
			}

			// del
			if char == 127 {
				if buff.Len() <= 0 {
					continue
				}
				buff.Truncate(buff.Len() - 1)
				v.mustar.out.Write([]byte("\b \b"))
				continue
			}

			// intro
			if char == 13 {
				v.mustar.out.Write([]byte("\r\n"))
				buff.Write([]byte("\n"))
				_, err := v.sluva.out.Write(buff.Bytes())
				if err != nil {
					v.mustar.out.Write([]byte(errors.Join(fmt.Errorf("app write err"), err).Error()))
				}
				buff.Reset()
				continue
			}

			buff.WriteRune(char)
			v.mustar.out.Write([]byte(string(char)))

		case _err := <-_e:
			v.mustar.out.Write([]byte(errors.Join(fmt.Errorf("term read err"), _err).Error()))
		default:
			<-time.After(100 * time.Microsecond)
		}

	}

	// fmt.Println("copying mustar out to sluva in")
	// _, err := io.Copy(v.mustar.out, v.sluva.in)
	// if err != nil {
	// 	v.addErr(err)
	// }
	// fmt.Println("end of copying mustar out to sluva in")
	// defer v.mustar.in.Close()
	// defer v.sluva.in.Close()
}

func (v *VPTY) OpenMustar(in io.ReadCloser, out io.WriteCloser) {
	v.mustar.Set(in, out)
}

func (v *VPTY) Close() error {
	return errors.Join(
		v.mustar.in.Close(),
		v.sluva.in.Close(),
		v.sluva.out.Close(),
		v.mustar.out.Close(),
	)
}

func NewVirtpty() *VPTY {
	return &VPTY{
		sluva:  &port{},
		mustar: &port{},
	}
}
