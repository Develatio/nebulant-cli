package term

import (
	"fmt"
	"io"
	"time"

	"github.com/schollz/progressbar/v3"
)

type oneLineWriteCloser struct {
	MainStdout *MultilineStdout
	P          []byte
}

func (s *oneLineWriteCloser) Write(p []byte) (int, error) {
	s.P = p
	return s.MainStdout.RePaintLines()
}

func (s *oneLineWriteCloser) Close() error {
	s.P = []byte("")
	_, err := s.MainStdout.RePaintLines()
	if err != nil {
		return err
	}
	return s.MainStdout.DeleteLine(s)
}

func (s *oneLineWriteCloser) GetProgressBar(max int64, description string, showbytes bool) *progressbar.ProgressBar {
	return progressbar.NewOptions64(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(s),
		progressbar.OptionShowBytes(showbytes),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			s.Write([]byte("DONE"))
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

func (m *oneLineWriteCloser) Print(s string) {
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
	return Stdout.Close()
}

func (m *MultilineStdout) RePaintLines() (int, error) {
	// Make space for Lines
	for i := 0; i < len(m.Lines); i++ {
		m.MainStdout.Write([]byte("\n"))
		m.MainStdout.Write([]byte(EraseLine))
	}

	// Back
	for i := 0; i < len(m.Lines); i++ {
		m.MainStdout.Write([]byte(CursorUp))
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
			return 0, nil
		}
		// Line down
		m.MainStdout.Write([]byte("\n"))
	}

	// Back again
	for i := 0; i < len(m.Lines); i++ {
		m.MainStdout.Write([]byte(CursorUp))
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
