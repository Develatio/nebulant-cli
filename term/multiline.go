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
	// where to write cumulated eco
	stdout io.WriteCloser
	p      []byte
	close  chan bool
	mu     sync.Mutex
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
	s.stdout.Write(spc)
}

func (s *stdinEcoWriter) Init() {
	go func() {
		ticker := time.NewTicker((1 * time.Second) / 2)
		tick := true
		for {
			select {
			case <-s.close:
				fmt.Println("eco close")
				return
			case <-ticker.C:
				s.mu.Lock()
				s.eco(tick)
				s.mu.Unlock()
				tick = !tick
			}
		}
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
	MainStdout *MultilineStdout
	P          []byte
}

func (o *oneLineWriteCloser) Write(p []byte) (int, error) {
	if !isTerminal() {
		return o.MainStdout.Write(p)
	}
	o.P = p
	return o.MainStdout.RePaintLines()
}

func (o *oneLineWriteCloser) Close() error {
	if !isTerminal() {
		return nil
	}
	o.P = []byte("")
	_, err := o.MainStdout.RePaintLines()
	if err != nil {
		return err
	}
	return o.MainStdout.DeleteLine(o)
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
			progressbar.OptionUseANSICodes(true),
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

func (o *oneLineWriteCloser) Scanln(prompt string, a ...any) (n int, err error) {
	// to prevent interactively ask many
	// options at a time
	tsi.mu.Lock()
	defer tsi.mu.Unlock()
	oldState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)
	n, err = o.MainStdout.Write([]byte(HideCursor))
	if err != nil {
		return n, err
	}
	defer o.MainStdout.Write([]byte(ShowCursor)) //#nosec G104 -- Unhandle is OK here

	var buff bytes.Buffer
	eco := &stdinEcoWriter{stdout: o}
	eco.Init()
	defer eco.Stop()
	eco.Write([]byte(prompt))
	reader := bufio.NewReader(os.Stdin)

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
			fmt.Fscan(bytes.NewReader(buff.Bytes()), a...)
			return 0, nil
		}
		if char == 127 {
			if buff.Len() <= 0 {
				continue
			}
			eco.Backspace()
			buff.Truncate(buff.Len() - 1)
			continue
		}
		eco.Write([]byte(string(char)))
		buff.WriteRune(char)
		// var buf []byte
		// utf8.EncodeRune(buf, char)
		// buff.Write(buf)

	}
	return 0, io.EOF
}

type MultilineStdout struct {
	MainStdout io.WriteCloser
	Lines      []*oneLineWriteCloser
	mu         sync.Mutex
}

func (m *MultilineStdout) Write(p []byte) (int, error) {
	if !isTerminal() {
		return m.MainStdout.Write(p)
	}
	n, err := m.MainStdout.Write([]byte(EraseLine))
	if err != nil {
		return n, err
	}
	n, err = m.MainStdout.Write(p)
	if err != nil {
		return n, err
	}
	n, err2 := m.RePaintLines()
	if err2 != nil {
		return n, err2
	}
	return n, err
}

func (m *MultilineStdout) Close() error {
	return m.MainStdout.Close()
}

func (m *MultilineStdout) RePaintLines() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !isTerminal() {
		return 0, nil
	}
	// Make space for Lines
	for i := 0; i < len(m.Lines); i++ {
		_, err := m.MainStdout.Write([]byte("\n"))
		if err != nil {
			return i, err
		}
		_, err = m.MainStdout.Write([]byte(EraseLine))
		if err != nil {
			return i, err
		}
	}

	// Back
	for i := 0; i < len(m.Lines); i++ {
		_, err := m.MainStdout.Write([]byte(CursorUp))
		if err != nil {
			return i, err
		}
	}

	// Write lines
	for i := 0; i < len(m.Lines); i++ {
		// TODO: prevent carriage return \n
		_, err := m.MainStdout.Write([]byte(EraseLine))
		if err != nil {
			return 0, nil
		}
		_, err = m.MainStdout.Write(m.Lines[i].P)
		if err != nil {
			return i, nil
		}
		// Line down
		_, err = m.MainStdout.Write([]byte("\n"))
		if err != nil {
			return i, err
		}
	}

	// Back again
	for i := 0; i < len(m.Lines); i++ {
		_, err := m.MainStdout.Write([]byte(CursorUp))
		if err != nil {
			return i, err
		}
	}

	return len(m.Lines), nil
}

func (m *MultilineStdout) AppendLine() *oneLineWriteCloser {
	m.mu.Lock()
	defer m.mu.Unlock()
	awc := &oneLineWriteCloser{
		MainStdout: m,
	}
	m.Lines = append(m.Lines, awc)
	return awc
}

func (m *MultilineStdout) DeleteLine(line *oneLineWriteCloser) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	index := -1
	for i := 0; i < len(m.Lines); i++ {
		if m.Lines[i] == line {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("unkown multi line")
	}

	m.Lines = append(m.Lines[:index], m.Lines[index+1:]...)
	return nil
}

func (m *MultilineStdout) SelectTest(prompt string, options []string) (int, error) {
	tsi.mu.Lock()
	defer tsi.mu.Unlock()
	oldState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)
	_, err = m.MainStdout.Write([]byte(HideCursor))
	if err != nil {
		return -1, err
	}
	defer m.MainStdout.Write([]byte(ShowCursor)) //#nosec G104 -- Unhandle is OK here

	var buff bytes.Buffer
	reader := bufio.NewReader(os.Stdin)
	selected := 0
	var lines []*oneLineWriteCloser

	// prompt
	nl := m.AppendLine()
	defer m.DeleteLine(nl)
	lines = append(lines, nl)

	// helper text
	nl = m.AppendLine()
	defer m.DeleteLine(nl)
	lines = append(lines, nl)

	// options
	for i := 0; i < len(options); i++ {
		nl := m.AppendLine()
		defer m.DeleteLine(nl)
		lines = append(lines, nl)
	}

	for {
		// prompt
		lines[0].Write([]byte(CorsorToColZero + EraseLine + prompt))

		// helper text
		lines[1].Write([]byte(CorsorToColZero + EraseLine + "use arrows :)"))
		for i := 0; i < len(options); i++ {
			if i == selected {
				lines[i+2].Write([]byte(CorsorToColZero + Reset + EraseLine + Blue + "»" + Magenta + "» " + Reset + options[i] + Reset))
			} else {
				lines[i+2].Write([]byte(CorsorToColZero + Reset + EraseLine + "   " + options[i] + Reset))
			}
		}

		// read stdin
		char, size, err := reader.ReadRune()
		if err != nil {
			return size, err
		}

		if char == 13 || char == 10 {
			for i := 0; i < len(lines); i++ {
				lines[i].Write([]byte(CorsorToColZero + Reset + EraseLine))
			}
			lines[0].Write([]byte("  ✓ " + options[selected]))
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
	return -1, io.EOF
}
