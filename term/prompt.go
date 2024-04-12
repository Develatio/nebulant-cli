package term

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

type Prompt struct {
	mu     sync.Mutex
	PS1    string
	Stdin  io.Reader
	Stdout io.Writer
}

func (p *Prompt) Cursor(stop chan struct{}) {
	ticker := time.NewTicker((1 * time.Second) / 2)
	step := 0
	cursorstep := [][]byte{
		[]byte(Blue + "\u2588" + Reset + CursorLeft),
		[]byte(Reset + " " + Reset + CursorLeft),
		[]byte(Magenta + "\u2588" + Reset + CursorLeft),
		[]byte(Reset + " " + Reset + CursorLeft),
	}
L:
	for {
		select {
		case <-stop:
			p.mu.Lock()
			_, err := p.Stdout.Write(cursorstep[1])
			if err != nil {
				fmt.Println(err)
			}
			p.mu.Unlock()
			break L
		case <-ticker.C:
			p.mu.Lock()
			if step == len(cursorstep) {
				step = 0
			}
			_, err := p.Stdout.Write(cursorstep[step])
			if err != nil {
				fmt.Println(err)
			}
			step++
			p.mu.Unlock()
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	ticker.Stop()
}

func (p *Prompt) Write(c []byte) (n int, err error) {
	cc := append([]byte("\b"), c...)
	return p.Stdout.Write(cc)
}

func (p *Prompt) Prompt(buff *bytes.Buffer) (n int, err error) {
	// to prevent interactively ask many
	// options at a time
	tsi.mu.Lock()
	defer tsi.mu.Unlock()

	// raw term
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)

	// hide system cursor, windows friendly
	err = SetHideCursor()
	if err != nil {
		return n, err
	}
	defer SetShowCursor()

	// hide system cursor, ANSI friendly
	n, err = p.Stdout.Write([]byte(HideCursor + "\r"))
	if err != nil {
		return n, err
	}
	defer p.Stdout.Write([]byte(ShowCursor + "\r")) // #nosec G104 -- Unhandle is OK here

	n, err = p.Stdout.Write([]byte("\n" + CursorToColZero + EraseLine + p.PS1))
	if err != nil {
		return n, err
	}

	// custom cursor
	stopcursor := make(chan struct{})
	defer close(stopcursor)
	go p.Cursor(stopcursor)
	defer p.Stdout.Write([]byte(CursorToColZero + EraseLine)) // #nosec G104 -- Unhandle is OK here

	reader := bufio.NewReader(p.Stdin)
	for {
		char, size, err := reader.ReadRune()
		if err != nil {
			return size, err
		}

		// ascii codes:
		// 127 for del
		// 3 for ^C
		// -1 for eof
		// 13 carriage return
		if char == 13 || char == -1 {
			if buff.Len() <= 0 {
				return 0, io.EOF
			}
			return buff.Len(), nil
		}

		// del
		if char == 127 || char == 8 {
			if buff.Len() <= 0 {
				continue
			}
			p.mu.Lock()
			n, err := p.Stdout.Write([]byte(" \b\b"))
			if err != nil {
				return n, err
			}
			buff.Truncate(buff.Len() - 1)
			p.mu.Unlock()
			continue
		}

		// ^C
		if char == 3 {
			return -1, EOT
		}

		p.mu.Lock()
		_, err = p.Stdout.Write([]byte(string(char)))
		if err != nil {
			return -1, err
		}
		buff.WriteRune(char)
		p.mu.Unlock()
	}
	// return 0, io.EOF
}
