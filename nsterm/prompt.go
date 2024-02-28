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
	"strings"

	"github.com/develatio/nebulant-cli/term"
)

// Prompt struct. A helper to write
// shell prompts asking for a input
// abstracting the origin of the app
// and the input source. This means
// that the source and/or the app
// can be local or remote. This should
// be used with VPTY2 that handles all
// the complex stuff.
// This pieze allows to have a PS
// prompt, line edit and history.
type Prompt struct {
	vpty       *VPTY2
	stdin      io.ReadCloser
	stdout     io.WriteCloser
	history    []string
	historyidx int
	PS1        []byte
	ldisc      Ldisc
}

func NewPrompt(vpty *VPTY2, stdin io.ReadCloser, stdout io.WriteCloser) *Prompt {
	return &Prompt{
		vpty:   vpty,
		stdin:  stdin,
		stdout: stdout,
		ldisc:  NewDefaultLdisc(),
		PS1:    []byte("> "),
	}
}

func (p *Prompt) SetPS1(ps1 string) {
	p.PS1 = []byte(ps1)
}

func (p *Prompt) ReadLine() (*string, error) {
	p.vpty.SetLDisc(p.ldisc)
	ESC := p.ldisc.GetESC()
	for {
		p.stdout.Write([]byte(term.CursorToColZero + term.EraseEntireLine))
		p.stdout.Write(p.PS1)
		p.stdout.Write([]byte(string(p.ldisc.ReadRuneBuff())))
	L2:
		for {
			esc := <-ESC
			switch string(esc) {
			case CursorUp: // ESC arrow up
				p.historyidx--
				if p.historyidx < 0 {
					p.historyidx = 0
				}
				tcmd := p.history[p.historyidx]
				p.ldisc.SetBuff(tcmd)
				break L2
			case CursorDown: // ESC arrow down
				p.historyidx++
				if p.historyidx > len(p.history)-1 {
					p.historyidx = len(p.history) - 1
					p.ldisc.SetBuff("")
					break L2 // force re-prompt PS1
				}
				tcmd := p.history[p.historyidx]
				p.ldisc.SetBuff(tcmd)
				break L2
			case CarriageReturn:
				p.stdout.Write([]byte("\n"))
				pb := make([]byte, len(p.ldisc.ReadRuneBuff()))
				p.ldisc.SetBuff("")
				n, err := p.stdin.Read(pb)
				if err != nil {
					return nil, err
				}

				// just \n or \r, skip
				if n <= 1 {
					break L2
				}

				s := string(pb)
				s = strings.TrimSuffix(s, "\n")
				s = strings.TrimSuffix(s, "\r")

				p.history = append(p.history, s)
				p.historyidx = len(p.history)

				return &s, nil
				// run command
				// ecode, err := run(vpty, s, stdin, stdout)

				// if err != nil {
				// 	stdout.Write([]byte(err.Error()))
				// 	stdout.Write([]byte(fmt.Sprintf("Exitcode %v", ecode)))
				// }
				// end of run command

				// break L2 // force re-prompt PS1
			}
		}
	}
}
