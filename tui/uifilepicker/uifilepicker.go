// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

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

package uifilepicker

import (
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/tui/theme"
	"github.com/develatio/nebulant-cli/tuicmd"
)

type formState uint

const (
	defaultTime           = time.Minute
	emptyState  formState = iota
	pickingState
	runningState
	waitenterState
)

type QuitFilePickerMsg struct{}

type FilePickerForm struct {
	state            formState
	pickermodel      *huh.Form
	currentPickerDir string
}

func emptyForm() (*huh.Form, formState) {
	return huh.NewForm(
		huh.NewGroup(huh.NewNote())), emptyState
}

func newPickerForm(currentdir string) *huh.Form {
	var err error
	if currentdir == "" {
		currentdir, err = os.Getwd()
		if err != nil {
			currentdir, err = os.UserHomeDir()
			if err != nil {
				currentdir = "/"
			}
		}
	}

	filepicker := huh.NewFilePicker().
		Key("value").
		Title("File Picker").
		Description("Select blueprint file").Picking(true).CurrentDirectory(currentdir).ShowHidden(true)
	fp := huh.NewForm(
		huh.NewGroup(filepicker).WithShowHelp(true),
	).WithHeight(25).WithTheme(theme.HuhTheme()).WithKeyMap(&huh.KeyMap{
		FilePicker: huh.FilePickerKeyMap{
			GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first"), key.WithDisabled()),
			GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last"), key.WithDisabled()),
			PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up"), key.WithDisabled()),
			PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down"), key.WithDisabled()),
			Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("←", "back"), key.WithDisabled()),
			Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"), key.WithDisabled()),
			Up:       key.NewBinding(key.WithKeys("up", "k", "ctrl+k", "ctrl+p"), key.WithHelp("↑", "up"), key.WithDisabled()),
			Down:     key.NewBinding(key.WithKeys("down", "j", "ctrl+j", "ctrl+n"), key.WithHelp("↓", "down"), key.WithDisabled()),
			Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("→", "join")),
			Close:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close"), key.WithDisabled()),
			Prev:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back"), key.WithDisabled()),
			Next:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
			Submit:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		},
	})
	return fp
}

func newModel() (*FilePickerForm, error) {
	return &FilePickerForm{
		state:       pickingState,
		pickermodel: newPickerForm(""),
	}, nil
}

func (m *FilePickerForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := m.pickermodel.Update(msg)
	cmds = append(cmds, cmd)
	if f, ok := form.(*huh.Form); ok {
		if f.State == huh.StateCompleted {
			kl := f.GetString("value")
			m.currentPickerDir = filepath.Dir(kl)
			m.pickermodel = newPickerForm(m.currentPickerDir)
			cmds = append(cmds, m.pickermodel.Init())
			m.state = runningState
			cmds = append(cmds, msgRunLocalStartCmd(), tuicmd.RunLocalBPCmd(kl))
		} else if f.State == huh.StateAborted {
			m.pickermodel, m.state = emptyForm()
			cmds = append(cmds, QuitFilePickerCmd())
		} else {
			m.pickermodel = f
		}
	}

	switch msg := msg.(type) {
	case tuicmd.RunLocalBPCmdENDMsg:
		if msg.Err != nil {
			cast.LogErr(msg.Err.Error(), nil)
		}
		m.state = waitenterState
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.pickermodel, m.state = emptyForm()
			cmds = append(cmds, QuitFilePickerCmd())
		case "h":
		case "enter":
			if m.state == waitenterState {
				m.pickermodel = newPickerForm(m.currentPickerDir)
				cmds = append(cmds, m.pickermodel.Init())
				m.state = pickingState
			}
		default:
		}
	default:
	}

	return m, tea.Batch(cmds...)
}

func msgRunLocalStartCmd() tea.Cmd {
	return func() tea.Msg {
		return tuicmd.RunLocalBPCmdStartMsg{}
	}
}

func (m *FilePickerForm) View() string {
	if m.state == emptyState || m.state == runningState {
		return ""
	}
	if m.state == waitenterState {
		return lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().
				Width(24).
				Align(lipgloss.Center).Render("Run finished. Press enter to continue..."),
		)

	}
	return m.pickermodel.View()
}

func QuitFilePickerCmd() tea.Cmd {
	return func() tea.Msg {
		return QuitFilePickerMsg{}
	}
}

func (m *FilePickerForm) Init() tea.Cmd {
	return m.pickermodel.Init()
}

func New() (*FilePickerForm, error) {
	m, err := newModel()
	if err != nil {
		return nil, err
	}
	return m, nil
}
