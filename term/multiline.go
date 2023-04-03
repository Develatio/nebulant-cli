package term

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/config"
	"github.com/schollz/progressbar/v3"
)

type oneLineWriteCloser struct {
	MainStdout *MultilineStdout
	P          []byte
}

func (s *oneLineWriteCloser) Write(p []byte) (int, error) {
	if *config.NoTermFlag {
		return s.MainStdout.Write(p)
	}
	s.P = p
	return s.MainStdout.RePaintLines()
}

func (s *oneLineWriteCloser) Close() error {
	if *config.NoTermFlag {
		return nil
	}
	s.P = []byte("")
	_, err := s.MainStdout.RePaintLines()
	if err != nil {
		return err
	}
	return s.MainStdout.DeleteLine(s)
}

type alwaysReturnWrapWritCloser struct {
	stdout io.WriteCloser
}

func (a *alwaysReturnWrapWritCloser) Write(p []byte) (int, error) {
	if bytes.HasSuffix(p, []byte("\r")) {
		p[len(p)-1] = 10
	}
	if bytes.HasPrefix(p, []byte("\r")) {
		p = p[1:]
	}
	if !bytes.HasSuffix(p, []byte("\n")) {
		p = append(p, 10)
	}
	return a.stdout.Write(p)
}

func (a *alwaysReturnWrapWritCloser) Close() error {
	return a.stdout.Close()
}

func (s *oneLineWriteCloser) GetProgressBar(max int64, description string, showbytes bool) *progressbar.ProgressBar {
	if *config.NoTermFlag {
		arwc := &alwaysReturnWrapWritCloser{
			stdout: s,
		}
		return progressbar.NewOptions64(max,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSetWriter(arwc),
			progressbar.OptionShowBytes(showbytes),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(1*time.Second),
			progressbar.OptionShowCount(),
			progressbar.OptionUseANSICodes(true),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
		)
	}
	return progressbar.NewOptions64(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(s),
		progressbar.OptionShowBytes(showbytes),
		progressbar.OptionEnableColorCodes(!*config.ColorFlag),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

func (m *oneLineWriteCloser) Print(s string) {
	if *config.NoTermFlag {
		if !strings.HasSuffix(s, "\n") {
			s = s + "\n"
		}
	}
	_, err := m.Write([]byte(s))
	if err != nil {
		panic(err)
	}
}

type MultilineStdout struct {
	MainStdout io.WriteCloser
	Lines      []*oneLineWriteCloser
}

func (m *MultilineStdout) Write(p []byte) (int, error) {
	if *config.NoTermFlag {
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
	if *config.NoTermFlag {
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
