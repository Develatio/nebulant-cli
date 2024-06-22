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

package uimenu

import (
	"net"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/config"
)

type sessionState uint

const (
	defaultTime              = time.Minute
	rootState   sessionState = iota
	menuState
	quitState
)

type Menu struct {
	state    sessionState
	mainForm *huh.Form
}

func rootForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title("Main menu. What do you want to do?").
				Options(
					huh.NewOption("Serve\tStart server mode at "+net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT), "serve-cmd"),
					huh.NewOption("Build\tOpen builder app and start server mode", "build-cmd"),
					huh.NewOption("Path\tManually indicates the path to a blueprint", "filepicker-form"),
					huh.NewOption("Browse\tBrowse and run the blueprints stored in your account", "browser-form"),
					huh.NewOption("Exit\tExit Nebulant CLI", "exit-cmd"),
				),
		))
}

func filepickerForm() *huh.Form {
	fp := huh.NewForm(
		huh.NewGroup(
			huh.NewFilePicker().
				Key("value").
				Title("File Picker").
				Description("Select blueprint file"),
		).WithShowHelp(true),
	).WithHeight(25)
	fp.Init()
	return fp
}

func newModel() Menu {
	return Menu{
		state:    menuState,
		mainForm: rootForm(),
	}
}

// type YesQuitMsg struct{}

// func (m *Menu) Init() tea.Cmd {
// 	// start the timer and spinner on program start
// 	return tea.Batch(m.spinner.Tick, readCastBusCmd(m.lk))
// }

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := m.mainForm.Update(msg)
	cmds = append(cmds, cmd)
	if f, ok := form.(*huh.Form); ok {
		if f.State == huh.StateCompleted {
			kl := f.GetString("value")
			cmds = append(cmds, tea.Println(kl))

			switch kl {
			case "":
				cmds = append(cmds, tea.Println("none select, come back?"))
			case "filepicker-form":
				m.mainForm = filepickerForm()
			default:
				m.mainForm = rootForm()
			}
			// m.mainForm = rootForm()
		} else {
			m.mainForm = f
		}
		// f.State = huh.StateNormal
	}

	return m, tea.Batch(cmds...)
}

func (m *Menu) View() string {

	// return lipgloss.NewStyle().
	// 	SetString("What’s for lunch?").
	// 	Height(32).
	// 	Foreground(lipgloss.Color("63")).Render(m.mainForm.View())

	return m.mainForm.View()

	// body := lipgloss.JoinHorizontal(lipgloss.Top, m.mainForm.View())
	// return body
	// // footer := m.mainForm.Help().ShortHelpView(m.mainForm.KeyBinds())

	// return body + "\n\n" + footer
}

func (m *Menu) Init() tea.Cmd {
	return m.mainForm.Init()
}

func New() *Menu {
	m := newModel()
	return &m
}
