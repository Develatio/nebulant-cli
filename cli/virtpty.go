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

type LDisc interface {
	// Call from Shell/Program:
	Open(*VPTY)
	Close()
	// read from vpty (master/keyboard/buff)
	Read()
	//
	// Sent data to tty device (stdout)
	Write()
	//
	IOctl()

	// Call from vpty
	ReceiveBuff()
}

/*
struct tty_ldisc_ops {
	char	*name;
	int	num;

	//  The following routines are called from above.
	//  (i.e., by the program or module accessing the TTY device):

	 // The function open() is called as soon as the TTY device switches to this line discipline.
	int	(*open)(struct tty_struct *tty);

	// The function close() is called when the current TTY line discipline is deactivated.
	// This happens when a TTY device switches from this line discipline into another one
	// (where the device is first reset to the standard line discipline N_TTX by the Linux
	// kernel) and when the TTY device itself is closed.
	void	(*close)(struct tty_struct *tty);

	void	(*flush_buffer)(struct tty_struct *tty);

	// The function read() is called when a program wants to read data from the TTY device.
	ssize_t	(*read)(struct tty_struct *tty, struct file *file, u8 *buf, size_t nr, void **cookie, unsigned long offset);

	// The function write() is called when a program wants to send data to the TTY device.
	ssize_t	(*write)(struct tty_struct *tty, struct file *file, const u8 *buf, size_t nr);

	// The function ioctl() is called when a program uses the system call ioctl() to change the
	// configuration of the TTY line discipline or of the actual TTY device, but only provided
	// that the higher-layer generic driver for TTY devices was unable to process the ioctl()
	// call (as is the case, for example, when the device switches to another TTY line discipline).
	int	(*ioctl)(struct tty_struct *tty, unsigned int cmd, unsigned long arg);

	int	(*compat_ioctl)(struct tty_struct *tty, unsigned int cmd, unsigned long arg);
	void	(*set_termios)(struct tty_struct *tty, const struct ktermios *old);
	__poll_t (*poll)(struct tty_struct *tty, struct file *file, struct poll_table_struct *wait);
	void	(*hangup)(struct tty_struct *tty);


	//  The following routines are called from below.
	//  (i.e., from the actual device driver of the TTY device)

	// The function receive_buf() is called when the device driver has received data and wants to
	// forward this data to the higher-layer program (i.e., to the driver of the TTY line discipline
	// in this case). The parameters passed include the address and length of data.
	void	(*receive_buf)(struct tty_struct *tty, const u8 *cp, const u8 *fp, size_t count);

	// The function write_wakeup() optionally can be called by the device driver as soon as it has
	// finished sending a data block and is ready to accept more data. However, this happens only
	// provided that it has been explicitly requested by the flag TTY_DO_WRITE_WAKEUP
	void	(*write_wakeup)(struct tty_struct *tty);

	void	(*dcd_change)(struct tty_struct *tty, bool active);
	size_t	(*receive_buf2)(struct tty_struct *tty, const u8 *cp, const u8 *fp, size_t count);
	void	(*lookahead_buf)(struct tty_struct *tty, const u8 *cp, const u8 *fp, size_t count);

	struct  module *owner;
};
*/

// func LFtoCRLF(data []byte) []byte {
// 	p := data
// 	m := bytes.Count(data, []byte{10})
// 	if m > 0 {
// 		l := len(data)
// 		p = make([]byte, l+m)
// 		e := -1
// 		for i := 0; i < len(data); i++ {
// 			e++
// 			p[e] = data[i]
// 			if data[i] != 10 {
// 				continue
// 			}
// 			// data_i == 10
// 			if i == 0 || (i > 0 && data[i-1] != 13) {
// 				p[e] = 13
// 				e++
// 				p[e] = 10
// 				continue
// 			}
// 		}
// 		p = p[:e+1]
// 	}
// 	return p
// }

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
	ldisc  LDisc
	mustar *port
	errors []error
}

func (v *VPTY) SetLDisc(ldisc LDisc) {
	v.ldisc = ldisc
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

func (v *VPTY) startLDisc() {
	go func() {
		_, err := io.Copy(v.mustar.out, v.sluva.in)
		if err != nil {
			v.addErr(err)
		}
	}()

	// prompt buff
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
