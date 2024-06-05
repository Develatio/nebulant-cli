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

package cast

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/term"
)

/*
This example assumes an existing understanding of commands and messages. If you
haven't already read our tutorials on the basics of Bubble Tea and working
with commands, we recommend reading those first.

Find them at:
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
*/

// sessionState is used to track which model is focused
type sessionState uint
type fillbuff struct {
	bd []*BusData
}

const (
	defaultTime              = time.Minute
	tableView   sessionState = iota
	loggerView
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	modelStyle = lipgloss.NewStyle().
			Width(25).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(25).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type mainModel struct {
	state    sessionState
	timer    timer.Model
	table    table.Model
	spinner  spinner.Model
	viewport viewport.Model
	// messages []string
	thmsg    map[string][]string
	thfilter string
	lk       *BusConsumerLink
	index    int
}

func frontUIModel(timeout time.Duration, l *BusConsumerLink) mainModel {
	m := mainModel{
		state:    tableView,
		lk:       l,
		thmsg:    make(map[string][]string),
		thfilter: "all",
	}
	m.timer = timer.New(timeout)
	// m.spinner = spinner.New()

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Nebulant CLI UI`)
	m.viewport = vp

	columns := []table.Column{
		{Title: "ThID", Width: 15},
		{Title: "S", Width: 2},
	}

	rows := []table.Row{
		//{"all", ""},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	m.table = t

	return m
}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.timer.Init(), m.spinner.Tick, readCastBusCmd(m.lk))
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// var vpCmd tea.Cmd

	// m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight
	case *BusData:
		cmds = append(cmds, readCastBusCmd(m.lk))
		switch msg.TypeID {
		case BusDataTypeLog:
			s := FormatConsoleLogMsg(msg)
			if s != "" {
				m.thmsg["all"] = append(m.thmsg["all"], s)
				if msg.ThreadID != nil {
					m.thmsg[*msg.ThreadID] = append(m.thmsg[*msg.ThreadID], s)
				}
				// TODO: set content of the selected thread id or all msgs
				// if none selected
				m.viewport.SetContent(strings.Join(m.thmsg[m.thfilter], "\n"))
				m.viewport.GotoBottom()
			}
		case BusDataTypeEOF:
			return m, tea.Quit
		case BusDataTypeEvent:
			// a nil pointer err here is a dev fail
			// never send event type without event id
			switch *msg.EventID {
			case EventNewThread:
				if _, exists := m.thmsg[*msg.ThreadID]; !exists {
					m.thmsg[*msg.ThreadID] = make([]string, 0)
				}
				rows := []table.Row{}
				for thid, _ := range m.thmsg {
					rows = append(rows, table.Row{thid, term.EmojiSet["Rocket"]})
				}
				m.table.SetRows(rows)
			}
			// mm := FormatConsoleLogMsg(msg)
			// if mm != "" {
			// 	m.messages = append(m.messages, mm)
			// }
			// m.viewport.SetContent(strings.Join(m.messages, "\n"))
			// m.viewport.GotoBottom()
		}
		// TODO: do things with cmd chann
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.state == tableView {
				m.state = loggerView
			} else {
				m.state = tableView
			}
		case "6":
			config.DEBUG = true
		case "enter":
			if m.state == tableView {
				m.thfilter = m.table.SelectedRow()[0]
			}
			m.viewport.SetContent(strings.Join(m.thmsg[m.thfilter], "\n"))
			m.viewport.GotoBottom()
		}
		switch m.state {
		// update whichever model is focused
		case loggerView:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		default:
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
		}
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// waitForActivity
func readCastBusCmd(cc *BusConsumerLink) tea.Cmd {
	return func() tea.Msg {
		select {
		case ff := <-cc.CommonChan:
			return ff
		case ff := <-cc.LogChan:
			return ff
		}
	}
}

func (m mainModel) View() string {
	var s string
	// model := m.currentFocusedModel()
	if m.state == tableView {
		s += lipgloss.JoinHorizontal(lipgloss.Top, focusedModelStyle.Render(fmt.Sprintf("%4s", tableStyle.Render(m.table.View()))), m.viewport.View())
	} else {
		s += lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(fmt.Sprintf("%4s", tableStyle.Render(m.table.View()))), m.viewport.View())
	}

	s += helpStyle.Render("\ntab: focus next • 6: debug • q: exit\n")
	return s
}

func (m mainModel) headerView() string {
	title := titleStyle.Render("Nebulant UI :)")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m mainModel) currentFocusedModel() string {
	if m.state == tableView {
		return "timer"
	}
	return "spinner"
}

func (m *mainModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *mainModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func StartUI(lk *BusConsumerLink) tea.Model {
	m := frontUIModel(defaultTime, lk)
	p := tea.NewProgram(m)
	go func() {
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	return m
}
