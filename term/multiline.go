package term

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/config"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

type threadSafeInput struct {
	mu sync.Mutex
}

var tsi = &threadSafeInput{}

type stdinEcoWriter struct {
	mu sync.Mutex
	// where to write cumulated eco
	stdout io.WriteCloser
	p      []byte
	close  chan bool
	tick   bool
}

func (s *stdinEcoWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p = append(s.p, p...)
	s.eco(true)
	return 0, nil
	// return s.stdout.Write(spc)
}

func (s *stdinEcoWriter) eco(tick bool) {
	var spc []byte
	if tick && s.tick {
		spc = append(s.p, []byte(Blue+"\u2588"+Reset)...)
		s.tick = !s.tick
	} else if tick && !s.tick {
		spc = append(s.p, []byte(Magenta+"\u2588"+Reset)...)
		s.tick = !s.tick
	} else {
		spc = append(s.p, []byte(Magenta+" "+Reset)...)
	}
	// write s.p + cursor to onelineWriteCloser
	s.stdout.Write([]byte(EraseEntireLine)) //#nosec G104 -- Unhandle is OK here
	s.stdout.Write(spc)                     //#nosec G104 -- Unhandle is OK here
}

func (s *stdinEcoWriter) Init() {
	s.close = make(chan bool, 1)
	go func() {
		ticker := time.NewTicker((1 * time.Second) / 2)
		tick := true
	L:
		for {
			select {
			case <-s.close:
				break L
			case <-ticker.C:
				s.mu.Lock()
				s.eco(tick)
				s.mu.Unlock()
				tick = !tick
			}
		}
		ticker.Stop()
	}()
}

func (s *stdinEcoWriter) Stop() {
	select {
	case s.close <- true:
	default:
		// Hey developer!,  what a wonderful day!
	}
}
func (s *stdinEcoWriter) Backspace() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p = s.p[:len(s.p)-1]
	s.eco(true)
}

type alwaysReturnWrapWritCloser struct {
	stdout io.WriteCloser
}

func (a *alwaysReturnWrapWritCloser) Write(p []byte) (int, error) {
	if bytes.HasSuffix(p, []byte("\r")) {
		// 10 == ascii line feed
		p[len(p)-1] = 10
	}
	if bytes.HasPrefix(p, []byte("\r")) {
		p = p[1:]
	}
	if !bytes.HasSuffix(p, []byte("\n")) {
		// 10 == ascii line feed
		p = append(p, 10)
	}
	return a.stdout.Write(p)
}

func (a *alwaysReturnWrapWritCloser) Close() error {
	return a.stdout.Close()
}

type oneLineWriteCloser struct {
	mu       sync.Mutex
	closed   bool
	MLStdout *MultilineStdout
	olP      []byte
}

func (o *oneLineWriteCloser) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !isTerminal() {
		return o.MLStdout.Write(p)
	}
	o.olP = p
	o.MLStdout.Repaint()
	return 0, nil
}

func (o *oneLineWriteCloser) IsClosed() bool {
	return o.closed
}

func (o *oneLineWriteCloser) Read() []byte {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.olP
}

func (o *oneLineWriteCloser) Close() error {
	if !isTerminal() {
		return nil
	}
	o.olP = []byte(EraseEntireLine + Reset)
	o.closed = true
	o.MLStdout.Repaint()
	return nil
}

func (o *oneLineWriteCloser) GetProgressBar(max int64, description string, showbytes bool) (*progressbar.ProgressBar, error) {
	description = " " + description
	if !isTerminal() {
		arwc := &alwaysReturnWrapWritCloser{
			stdout: o,
		}
		_, err := arwc.Write([]byte(description))
		if err != nil {
			return nil, err
		}

		return progressbar.NewOptions64(max,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSetWriter(arwc),
			progressbar.OptionShowBytes(showbytes),
			progressbar.OptionSetWidth(20),
			progressbar.OptionThrottle(30*time.Second),
			progressbar.OptionShowCount(),
			progressbar.OptionSpinnerCustom([]string{}),
		), nil
	}

	return progressbar.NewOptions64(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(o),
		progressbar.OptionShowBytes(showbytes),
		progressbar.OptionEnableColorCodes(!*config.DisableColorFlag),
		progressbar.OptionSetWidth(20),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionClearOnFinish(),
	), nil
}

func (o *oneLineWriteCloser) Print(s string) {
	if !isTerminal() {
		if !strings.HasSuffix(s, "\n") {
			s = s + "\n"
		}
	}
	_, err := o.Write([]byte(s))
	if err != nil {
		panic(err)
	}
}

func (o *oneLineWriteCloser) Scanln(prompt string, def []byte, a ...any) (n int, err error) {
	// to prevent interactively ask many
	// options at a time
	tsi.mu.Lock()
	defer tsi.mu.Unlock()
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)

	// hide cursor, windows friendly
	err = SetHideCursor()
	if err != nil {
		return n, err
	}
	defer SetShowCursor()

	// hide cursor, ANSI friendly
	n, err = o.MLStdout.Write([]byte(HideCursor + "\r"))
	if err != nil {
		return n, err
	}
	defer o.MLStdout.Write([]byte(ShowCursor + "\r")) //#nosec G104 -- Unhandle is OK here

	var buff bytes.Buffer
	eco := &stdinEcoWriter{stdout: o}
	eco.Init()
	defer eco.Stop()
	_, err = eco.Write([]byte(EraseEntireLine + prompt))
	if err != nil {
		return -1, err
	}
	reader := bufio.NewReader(os.Stdin)
	if def != nil {
		_, err := buff.Write(def)
		if err != nil {
			return 0, err
		}
		_, err = eco.Write(def)
		if err != nil {
			return 0, err
		}
	}

	for {
		char, size, err := reader.ReadRune()
		if err != nil {
			return size, err
		}

		// ascii codes:
		// 127 for del
		// 3 for ^C
		// -1 for eof
		if char == 13 || char == -1 {
			_, err := fmt.Fscan(bytes.NewReader(buff.Bytes()), a...)
			if err != nil {
				return -1, err
			}
			return 0, nil
		}
		if char == 127 || char == 8 {
			if buff.Len() <= 0 {
				continue
			}
			eco.Backspace()
			buff.Truncate(buff.Len() - 1)
			continue
		}

		_, err = eco.Write([]byte(string(char)))
		if err != nil {
			return -1, err
		}
		buff.WriteRune(char)
		// var buf []byte
		// utf8.EncodeRune(buf, char)
		// buff.Write(buf)

	}
	// return 0, io.EOF
}

type MultilineStdout struct {
	mu         sync.Mutex
	mainStdout io.WriteCloser
	Lines      []*oneLineWriteCloser
	close      chan bool
	repaint    chan bool
}

func (m *MultilineStdout) Init() {
	m.repaint = make(chan bool, 1)
	m.close = make(chan bool, 1)
	go func() {
		ticker := time.NewTicker((1 * time.Second) / 2)
		for {
			select {
			case <-m.close:
				return
			case <-m.repaint:
				m.rePaintLines(nil) //#nosec G104 -- Unhandle is OK here
			case <-ticker.C:
				m.rePaintLines(nil) //#nosec G104 -- Unhandle is OK here
			}
		}
	}()
}

func (m *MultilineStdout) Stop() {
	select {
	case m.close <- true:
	default:
		// Hey developer!,  what a wonderful day!
	}
}

func (m *MultilineStdout) SetMainStdout(mstdout io.WriteCloser) {
	m.mainStdout = mstdout
}

func (m *MultilineStdout) Repaint() {
	select {
	case m.repaint <- true:
	default:
		// Hey developer!,  what a wonderful day!
	}
}

func (m *MultilineStdout) Write(p []byte) (int, error) {
	if !isTerminal() {
		return m.mainStdout.Write(p)
	}
	n, err2 := m.rePaintLines(p)
	if err2 != nil {
		return n, err2
	}
	return n, err2
}

func (m *MultilineStdout) Close() error {
	m.close <- true
	return m.mainStdout.Close()
}

func (m *MultilineStdout) rePaintLines(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !isTerminal() {
		return 0, nil
	}

	lenlines := len(m.Lines)
	// Make space for Lines
	for i := 0; i < lenlines; i++ {
		// debug:
		// time.Sleep(1 * time.Second)
		//_, err := m.mainStdout.Write([]byte(CursorDown))
		_, err := m.mainStdout.Write([]byte("\r\n"))
		if err != nil {
			return i, err
		}
	}

	// Back
	for i := 0; i < lenlines; i++ {
		// debug:
		// time.Sleep(1 * time.Second)
		_, err := m.mainStdout.Write([]byte(CursorUp))
		if err != nil {
			return i, err
		}
		// _, err = m.mainStdout.Write([]byte(CursorToColZero + "\r" + Reset))
		// if err != nil {
		// 	return i, err
		// }
	}

	// write log line
	if p != nil {
		// debug:
		// time.Sleep(1 * time.Second)
		_, err := m.mainStdout.Write([]byte(EraseEntireLine + "\r" + Reset))
		if err != nil {
			return 0, nil
		}
		// debug:
		// time.Sleep(1 * time.Second)
		_, err = m.mainStdout.Write(p)
		if err != nil {
			return 0, nil
		}
	}

	// Write lines
	for i := 0; i < lenlines; i++ {
		// debug:
		// time.Sleep(1 * time.Second)
		// TODO: prevent carriage return \n
		_, err := m.mainStdout.Write([]byte(EraseEntireLine + "\r" + Reset))
		if err != nil {
			return 0, nil
		}

		_, err = m.mainStdout.Write(m.Lines[i].Read())
		if err != nil {
			return i, nil
		}

		// Line down
		_, err = m.mainStdout.Write([]byte("\n"))
		if err != nil {
			return i, err
		}
	}

	// Back again
	for i := 0; i < lenlines; i++ {
		// debug:
		// time.Sleep(1 * time.Second)
		_, err := m.mainStdout.Write([]byte(CursorUp + Reset))
		if err != nil {
			return i, err
		}
		// _, err = m.mainStdout.Write([]byte(CursorToColZero + "\r" + "zero2" + Reset))
		// if err != nil {
		// 	return i, err
		// }

		// debug
		// _, err = m.mainStdout.Write([]byte("--" + strconv.Itoa(i) + "/" + strconv.Itoa(lenlines) + "--"))
		// if err != nil {
		// 	return i, err
		// }
	}

	// Pop closed lines
	var lines []*oneLineWriteCloser
	for i := 0; i < lenlines; i++ {
		if m.Lines[i].IsClosed() {
			continue
		}
		lines = append(lines, m.Lines[i])
	}
	m.Lines = lines

	return lenlines, nil
}

func (m *MultilineStdout) AppendLine() *oneLineWriteCloser {
	m.mu.Lock()
	defer m.mu.Unlock()
	awc := &oneLineWriteCloser{
		MLStdout: m,
	}
	m.Lines = append(m.Lines, awc)
	return awc
}

func (m *MultilineStdout) SelectTest(prompt string, options []string) (int, error) {
	tsi.mu.Lock()
	defer tsi.mu.Unlock()
	oldState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)
	_, err = m.mainStdout.Write([]byte(HideCursor))
	if err != nil {
		return -1, err
	}
	defer m.mainStdout.Write([]byte(ShowCursor)) //#nosec G104 -- Unhandle is OK here

	var buff bytes.Buffer
	reader := bufio.NewReader(os.Stdin)
	selected := 0
	var lines []*oneLineWriteCloser

	// prompt
	nl := m.AppendLine()
	err = nl.Close()
	if err != nil {
		return 0, err
	}
	lines = append(lines, nl)

	// helper text
	nl = m.AppendLine()
	err = nl.Close()
	if err != nil {
		return 0, err
	}
	lines = append(lines, nl)

	// options
	for i := 0; i < len(options); i++ {
		nl := m.AppendLine()
		err = nl.Close()
		if err != nil {
			return 0, err
		}
		lines = append(lines, nl)
	}

	for {
		// prompt
		_, err := lines[0].Write([]byte(CursorToColZero + EraseLine + prompt))
		if err != nil {
			return -1, err
		}

		// helper text
		_, err = lines[1].Write([]byte(CursorToColZero + EraseLine + "use arrows :)"))
		if err != nil {
			return -1, err
		}
		for i := 0; i < len(options); i++ {
			if i == selected {
				_, err := lines[i+2].Write([]byte(CursorToColZero + Reset + EraseLine + Blue + "»" + Magenta + "» " + Reset + options[i] + Reset))
				if err != nil {
					return -1, err
				}
			} else {
				_, err := lines[i+2].Write([]byte(CursorToColZero + Reset + EraseLine + "   " + options[i] + Reset))
				if err != nil {
					return -1, err
				}
			}
		}

		// read stdin
		char, size, err := reader.ReadRune()
		if err != nil {
			return size, err
		}

		if char == 13 || char == 10 {
			for i := 0; i < len(lines); i++ {
				_, err := lines[i].Write([]byte(CursorToColZero + Reset + EraseLine))
				if err != nil {
					return -1, err
				}
			}
			_, err := lines[0].Write([]byte("  ✓ " + options[selected]))
			if err != nil {
				return -1, err
			}
			return selected, nil
		}
		if char == -1 {
			return -1, nil
		}

		// append stdin
		buff.WriteRune(char)

		// listen just for escape seq
		if buff.Bytes()[0] != 27 {
			buff.Reset()
			continue
		}

		// skip unformed scape seq
		if len(buff.Bytes()) < 3 {
			continue
		}

		// match up
		if bytes.Equal(buff.Bytes(), []byte{27, 91, 65}) {
			// fmt.Println(buff.Bytes())
			if selected == 0 {
				buff.Reset()
				continue
			}
			selected--
			buff.Reset()
			continue
		}

		// match down
		if bytes.Equal(buff.Bytes(), []byte{27, 91, 66}) {
			if selected == len(options)-1 {
				buff.Reset()
				continue
			}
			selected++
			buff.Reset()
			continue
		}
		// skip else
		buff.Reset()
	}
	// return -1, io.EOF
}
