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
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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
	logState sessionState = iota
	menuState
	emptyState
	promptState
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

type promptInfo struct {
	b *cast.BusData
	m tea.Model
}

type mainModel struct {
	state sessionState
	// state before the quitState: the state that should
	// be restored after quit "n" answer
	quitRejectState sessionState
	// timer timer.Model
	// table   table.Model
	progress     map[string]*progressInfo
	progessslice []*progressInfo // same as above, but for iterate in order
	prompts      []*promptInfo   // a list of pending prompts
	prompt       *promptInfo     // the current active prompt
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
	width   int
	height  int
	threads map[string]bool
	// link to receive bus msg and events
	lk *cast.BusConsumerLink
	// setted debug level
	dbglevel int
}

func frontUIModel(l *cast.BusConsumerLink) mainModel {
	m := mainModel{
		state:           logState,
		quitRejectState: logState,
		lk:              l,
		threads:         make(map[string]bool),
		progress:        make(map[string]*progressInfo),
		dbglevel:        6,
		uimenu:          uimenu.New(),
	}

	// spinner
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	m.spinner = s

	// form

	return m
}

type YesQuitMsg struct{}
type NextPromptMsg struct{}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.spinner.Tick, readCastBusCmd(m.lk))
}

func newPrompt(b *cast.BusData) (*huh.Form, tea.Cmd) {
	var f *huh.Form
	var cmd tea.Cmd
	switch b.EPO.Type {
	case cast.EventPromptTypeBool:
		def := false
		if b.EPO.DefaultValue == "true" {
			def = true
		}
		cnf := huh.NewConfirm().
			Key("value").
			Title(b.EPO.PromptTitle).
			Affirmative("true").
			Negative("false").
			Value(&def)
		cmd = cnf.Focus()
		f = huh.NewForm(
			huh.NewGroup(cnf))
	case cast.EventPromptTypeInput:
		def := b.EPO.DefaultValue
		inp := huh.NewInput().
			Key("value").
			Title(b.EPO.PromptTitle).
			Value(&def)
		v := func(val string) error {
			if b.EPO.Validate.ValueType == "int" {
				_, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				return nil
			}
			// assume type string below this line
			if !b.EPO.Validate.AllowEmpty && val == "" {
				return fmt.Errorf("empty value not allowed")
			}
			return nil
		}
		inp.Validate(v)
		cmd = inp.Focus()
		f = huh.NewForm(
			huh.NewGroup(inp))
	case cast.EventPromptTypeSelect:
		var options []huh.Option[string]
		for v, l := range b.EPO.Options {
			opt := huh.NewOption(l, v)
			options = append(options, opt)
		}
		sel := huh.NewSelect[string]().
			Key("value").
			Title(b.EPO.PromptTitle).
			Options(options...)

		v := func(val string) error {
			if !b.EPO.Validate.AllowEmpty && val == "" {
				return fmt.Errorf("empty value not allowed")
			}
			return nil
		}
		sel.Validate(v)
		cmd = sel.Focus()
		f = huh.NewForm(
			huh.NewGroup(sel)).WithShowHelp(true)
	default:
		return nil, nil
	}
	return f, cmd
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
		case "y", "enter":
			return m, confirmQuit()
		case "n":
			fmt.Println(m.state, m.quitRejectState)
			m.state = m.quitRejectState
		default:
			tea.Println(msg)
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
			if s != nil {
				cmds = append(cmds, tea.Printf(*s))
			}
		case cast.BusDataTypeEOF:
			select {
			case m.lk.Off <- struct{}{}:
				// ok
			default:
				// hey dev:
				// on default case probably the Off chan has not been
				// init yet, so there is no need to off the logger
				// fmt.Println("warn: cannot stop TUI logger :O")
			}
			m.lk.Degraded = true
			m.state = emptyState
			cmds = append(cmds, tickCmd(), shutdownUI())
		case cast.BusDataTypeEvent:
			// a nil pointer err here is a dev fail
			// never send event type without event id
			switch *msg.EventID {
			case cast.EventNewThread:
				// m.threads[*msg.ThreadID] = true
				// mm := cast.FormatConsoleLogMsg(msg)
				// if mm != nil {
				// 	cmds = append(cmds, tea.Printf(*mm))
				// }
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
			case cast.EventPrompt:
				m.prompts = append(m.prompts, &promptInfo{b: msg})
				cmds = append(cmds, nextPromptCmd())
			case cast.EventPromptDone:
				if m.prompt != nil && m.prompt.b.EPO.UUID == msg.EPO.UUID {
					// this prompt has already answered by some other UI,
					// advance for the next prompt
					cmds = append(cmds, nextPromptCmd())
				} else {
					var prompts []*promptInfo
					for _, p := range m.prompts {
						if p.b.EPO.UUID == msg.EPO.UUID {
							// a prompt in the qeue has been already answered, skip
							continue
						}
						prompts = append(prompts, p)
					}
					m.prompts = prompts
				}
			case cast.EventInteractiveMenuStart:
				m.state = menuState
			}
		}
		// TODO: do things with cmd chann
	case NextPromptMsg:
		if len(m.prompts) == 0 {
			m.state = logState
		} else {
			m.state = promptState
			s := len(m.prompts)
			p := m.prompts[s-1]
			m.prompts = m.prompts[:s-1] // Remove the last value from the slice.
			p.m, cmd = newPrompt(p.b)
			cmds = append(cmds, cmd)
			if p.m != nil {
				m.prompt = p
			} else {
				// TODO: err?
				m.prompt = nil
				m.state = logState
				cmds = append(cmds, nextPromptCmd())
			}
		}
	case uimenu.QuitMenuMsg:
		m.state = logState
	case tuicmd.RunRemoteBPCmdStartMsg:
		m.state = logState
	case tuicmd.RunRemoteBPCmdENDMsg:
		m.state = menuState
	case tea.QuitMsg:
		m.state = emptyState
		return m, tea.Quit
	case tuicmd.QuitStateMsg:
		m.quitRejectState = m.state
		m.state = quitState
	case tea.KeyMsg:
		if m.state == logState {
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitRejectState = m.state
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
		return m, cmd // ???

	}

	switch m.state {
	case promptState:
		form, cmd := m.prompt.m.Update(msg)
		cmds = append(cmds, cmd)
		if f, ok := form.(*huh.Form); ok {
			m.prompt.m = f
			if f.State == huh.StateCompleted {
				kl := f.GetString("value")
				if m.prompt.b.EPO.Type == cast.EventPromptTypeBool {
					kb := f.GetBool("value")
					kl = "false"
					if kb {
						kl = "true"
					}
				}
				cast.AnswerPrompt(m.prompt.b, kl)
				cmds = append(cmds, nextPromptCmd())
			} else if f.State == huh.StateAborted {
				cast.AnswerPrompt(m.prompt.b, "")
				cmds = append(cmds, nextPromptCmd())
			}
		}
	case menuState:
		_, cmdss := m.uimenu.Update(msg)
		cmds = append(cmds, cmdss)
	case quitState:
		_m, cmdss := m.updateQuitView(msg)
		m = _m.(mainModel)
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

func tickCmd() tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

func nextPromptCmd() tea.Cmd {
	return func() tea.Msg {
		return NextPromptMsg{}
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
	// return helpStyle.Render("\nb: open builder • p: open pannel • l: switch log level • q: exit\n")
	return helpStyle.Render("\nb: open builder • p: open pannel • q: exit\n")
}

func (m mainModel) quitView() string {
	text := lipgloss.JoinHorizontal(lipgloss.Top, "Are you sure you want to leave Nebulant CLI?", choiceStyle.Render("[Yn]"))
	return quitViewStyle.Render(text)
}

func (m mainModel) logView() string {
	return renderHelp()
}

func (m mainModel) promptView() string {
	if m.prompt != nil {
		return m.prompt.m.View()
	}
	return ""
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
	case promptState:
		s = m.promptView()
	case quitState:
		s = m.quitView()
	}

	return vv + s
}

func StartUI(lk *cast.BusConsumerLink) (tea.Model, error) {
	waiter.Add(1)
	defer waiter.Done()
	defer waiter.Done()

	m := frontUIModel(lk)
	p := tea.NewProgram(m)
	defer func() {
		p.ReleaseTerminal()
	}()
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
