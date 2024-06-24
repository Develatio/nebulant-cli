// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/tui/uimenu"
	"github.com/develatio/nebulant-cli/tuicmd"
)

var waiter *sync.WaitGroup

// sessionState is used to track which model is focused
type sessionState uint

const (
	defaultTime              = time.Minute
	logState    sessionState = iota
	menuState
	emptyState
	quitState
)

var (
	progressMsg   = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	choiceStyle   = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("222"))
	quitViewStyle = lipgloss.NewStyle().Padding(1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170"))
)

type progressInfo struct {
	info string
	size int64
	// writed   int64
	progress progress.Model
}

type mainModel struct {
	state sessionState
	// timer timer.Model
	// table   table.Model
	progress     map[string]*progressInfo
	progessslice []*progressInfo // same as above, but for iterate in order
	// available spinners:
	// spinner.Line,
	// spinner.Dot,
	// spinner.MiniDot,
	// spinner.Jump,
	// spinner.Pulse,
	// spinner.Points,
	// spinner.Globe,
	// spinner.Moon,
	// spinner.Monkey,
	spinner spinner.Model
	uimenu  *uimenu.Menu
	//
	width    int
	height   int
	threads  map[string]bool
	thfilter string
	lk       *cast.BusConsumerLink
	dbglevel int
}

func frontUIModel(timeout time.Duration, l *cast.BusConsumerLink) mainModel {
	m := mainModel{
		state:    logState,
		lk:       l,
		threads:  make(map[string]bool),
		thfilter: "all",
		progress: make(map[string]*progressInfo),
		dbglevel: 6,
		uimenu:   uimenu.New(),
	}

	// spinner
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	m.spinner = s

	// form

	return m
}

type YesQuitMsg struct{}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.spinner.Tick, readCastBusCmd(m.lk))
}

func (m mainModel) updateQuitView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case YesQuitMsg:
		select {
		case m.lk.Off <- struct{}{}:
			// ok
		default:
			fmt.Println("warn: cannot stop TUI logger :O")
		}
		select {
		case base.InterruptSignalChannel <- os.Interrupt:
			// ok
		default:
			fmt.Println("warn: cannot send interrupt signal :O")
		}
		m.state = emptyState
		cmds = append(cmds, shutdownUI())
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			if m.state == quitState {
				return m, confirmQuit()
			}
		case "n":
			if m.state == quitState {
				m.state = logState
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// var vpCmd tea.Cmd

	// m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case *cast.BusData:
		cmds = append(cmds, readCastBusCmd(m.lk))
		switch msg.TypeID {
		case cast.BusDataTypeLog:
			s := cast.FormatConsoleLogMsg(msg)
			if s != "" {
				cmds = append(cmds, tea.Printf(s))
			}
		case cast.BusDataTypeEOF:
			select {
			case m.lk.Off <- struct{}{}:
				// ok
			default:
				fmt.Println("warn: cannot stop TUI logger :O")
			}
			m.lk.Degraded = true
			cmds = append(cmds, shutdownUI())
		case cast.BusDataTypeEvent:
			// a nil pointer err here is a dev fail
			// never send event type without event id
			switch *msg.EventID {
			case cast.EventNewThread:
				m.threads[*msg.ThreadID] = true
				mm := cast.FormatConsoleLogMsg(msg)
				if mm != "" {
					cmds = append(cmds, tea.Printf(mm))
				}
			case cast.EventProgressStart:
				npr := &progressInfo{
					progress: progress.New(
						progress.WithDefaultGradient(),
						progress.WithWidth(40),
						// progress.WithoutPercentage(),
					),
					size: msg.Extra["size"].(int64),
					info: msg.Extra["info"].(string),
				}
				m.progress[msg.Extra["progressid"].(string)] = npr
				m.progessslice = append(m.progessslice, npr)
				cmds = append(cmds, npr.progress.SetPercent(0))
			case cast.EventProgressTick:
				size := msg.Extra["size"].(int64)
				writed := msg.Extra["writed"].(int64)
				pid := msg.Extra["progressid"].(string)
				pinfo := m.progress[pid]
				if writed <= 0 {
					cmds = append(cmds, pinfo.progress.SetPercent(0.0))
				} else {
					cmds = append(cmds, pinfo.progress.SetPercent(float64(writed)/float64(size)))
				}
			case cast.EventProgressEnd:
				// TODO: add cmd?
				pid := msg.Extra["progressid"].(string)
				npr := m.progress[pid]
				for ii, _npr := range m.progessslice {
					if npr == _npr {
						// index found
						m.progessslice = append(m.progessslice[:ii], m.progessslice[ii+1:]...)
					}
				}
				delete(m.progress, pid)
			case cast.EventInteractiveMenuStart:
				m.state = menuState
			}
		}
		// TODO: do things with cmd chann
	case uimenu.QuitMenuMsg:
		m.state = logState
	case tea.QuitMsg:
		m.state = emptyState
		return m, tea.Quit
	case tea.KeyMsg:
		if m.state == logState {
			switch msg.String() {
			case "ctrl+c", "q":
				m.state = quitState
			case "b":
				cmds = append(cmds, tuicmd.OpenBuilderCmd())
			case "6":
				config.DEBUG = true
			case "enter":
				// nothing yet
			}
		}
		// switch m.state?
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	// case timer.TickMsg:
	// 	m.timer, cmd = m.timer.Update(msg)
	// 	cmds = append(cmds, cmd)
	case progress.FrameMsg:
		for _, pinf := range m.progress {
			newModel, cmd := pinf.progress.Update(msg)
			if newModel, ok := newModel.(progress.Model); ok {
				pinf.progress = newModel
			}
			cmds = append(cmds, cmd)
		}
		return m, cmd

	}

	switch m.state {
	case menuState:
		_, cmdss := m.uimenu.Update(msg)
		cmds = append(cmds, cmdss)
	case quitState:
		_, cmdss := m.updateQuitView(msg)
		cmds = append(cmds, cmdss)
	case logState:
		//
	}

	return m, tea.Batch(cmds...)
}

// waitForActivity
func readCastBusCmd(cc *cast.BusConsumerLink) tea.Cmd {
	return func() tea.Msg {
		select {
		case ff := <-cc.CommonChan:
			return ff
		case ff := <-cc.LogChan:
			return ff
		case <-cc.Off:
			return nil
		}
	}
}

func confirmQuit() tea.Cmd {
	return func() tea.Msg {
		return YesQuitMsg{}
	}
}

func shutdownUI() tea.Cmd {
	return tea.Quit
}

func renderHelp() string {
	return helpStyle.Render("\nb: open builder • p: open pannel • l: switch log level • q: exit\n")
}

func (m mainModel) quitView() string {
	text := lipgloss.JoinHorizontal(lipgloss.Top, "Are you sure you want to leave Nebulant CLI?", choiceStyle.Render("[yn]"))
	return quitViewStyle.Render(text)
}

func (m mainModel) logView() string {
	return renderHelp()
}

func (m mainModel) View() string {
	// progress will always be rendered
	vv := ""
	for _, pinf := range m.progessslice {
		spin := m.spinner.View() + " "
		prog := pinf.progress.View()
		cellsAvail := max(0, m.width-lipgloss.Width(spin+prog))
		prgInfo := progressMsg.Render(pinf.info)
		info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(prgInfo)
		cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog))
		gap := strings.Repeat(" ", cellsRemaining)
		vv = vv + spin + info + gap + prog + "\n"
	}

	// render specific-state-view
	var s string
	switch m.state {
	case emptyState:
		s = ""
	case menuState:
		s = m.uimenu.View()
	case logState:
		s = m.logView()
	case quitState:
		s = m.quitView()
	}

	return vv + s
}

func StartUI(lk *cast.BusConsumerLink) (tea.Model, error) {
	waiter.Add(1)
	defer waiter.Done()
	defer waiter.Done()
	m := frontUIModel(defaultTime, lk)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		// stop cli?
		return m, err
	}
	// stop cli?
	return m, nil
}

func Wait() {
	waiter.Wait()
}

func init() {
	waiter = &sync.WaitGroup{}
	waiter.Add(1)
}
