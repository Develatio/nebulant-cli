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
	"golang.org/x/crypto/ssh/terminal"
)

type stdinEcoWriter struct {
	// where to write cumulated eco
	stdout io.WriteCloser
	p      []byte
	close  chan bool
	mu     sync.Mutex
	tick   bool
}

func (s *stdinEcoWriter) Write(p []byte) (int, error) {
	s.p = append(s.p, p...)
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (s *oneLineWriteCloser) Write(p []byte) (int, error) {
	if !isTerminal() {
		return s.MainStdout.Write(p)
	}
	s.P = p
	return s.MainStdout.RePaintLines()
}

func (s *oneLineWriteCloser) Close() error {
	if !isTerminal() {
		return nil
	}
	s.P = []byte("")
	_, err := s.MainStdout.RePaintLines()
	if err != nil {
		return err
	}
	return s.MainStdout.DeleteLine(s)
}

func (s *oneLineWriteCloser) GetProgressBar(max int64, description string, showbytes bool) (*progressbar.ProgressBar, error) {
	description = " " + description
	if !isTerminal() {
		arwc := &alwaysReturnWrapWritCloser{
			stdout: s,
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
		progressbar.OptionSetWriter(s),
		progressbar.OptionShowBytes(showbytes),
		progressbar.OptionEnableColorCodes(!*config.DisableColorFlag),
		progressbar.OptionSetWidth(20),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
	), nil
}

func (m *oneLineWriteCloser) Print(s string) {
	if !isTerminal() {
		if !strings.HasSuffix(s, "\n") {
			s = s + "\n"
		}
	}
	_, err := m.Write([]byte(s))
	if err != nil {
		panic(err)
	}
}

func (m *oneLineWriteCloser) Scanln(prompt string, a ...any) (n int, err error) {
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldState)
	n, err = m.MainStdout.Write([]byte(HideCursor))
	if err != nil {
		return n, err
	}
	defer m.MainStdout.Write([]byte(ShowCursor)) //#nosec G104 -- Unhandle is OK here

	var buff bytes.Buffer
	eco := &stdinEcoWriter{stdout: m}
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
	awc := &oneLineWriteCloser{
		MainStdout: m,
	}
	m.Lines = append(m.Lines, awc)
	return awc
}

func (m *MultilineStdout) DeleteLine(line *oneLineWriteCloser) error {
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
