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

package uiauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/tui/theme"
)

type formState uint

var spinnerTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
var keyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#979797"))
var descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

const (
	defaultTime           = time.Minute
	rootState   formState = iota
	detailtokenState
	runningState
	emptyState
)

type QuitAuthMsg struct{}

type SetTokenAsDefaultResultMsg struct{ Err error }
type RemoveTokenResultMsg struct{ Err error }
type RequestNewTokenResultMsg struct{ Err error }
type LoginResultMsg struct{ Err error }

type AuthForm struct {
	state             formState
	selectedTokenName string
	mainForm          tea.Model
	spinner           spinner.Model
	spinnerText       string
	runningCancelFunc context.CancelFunc
}

func emptyForm() (*huh.Form, formState) {
	return huh.NewForm(
		huh.NewGroup(huh.NewNote())), emptyState
}

func rootForm() (*huh.Form, formState, error) {
	var options []huh.Option[string]
	options = append(options, huh.NewOption("...Back", ".."))
	options = append(options, huh.NewOption("Request New Token...", "new-token-cmd"))
	crs, err := config.ReadCredentialsFile()
	if err != nil {
		return nil, rootState, err
	}
	for n, cr := range crs.Credentials {
		tit := *cr.Access
		if n == crs.ActiveProfile {
			tit = fmt.Sprintf("* %s", *cr.Access)
		}
		opt := huh.NewOption(tit, fmt.Sprintf("profile-%s", n))
		options = append(options, opt)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title("Manage your local credentials").
				Options(options...),
		)).WithShowHelp(true).WithTheme(theme.HuhTheme())
	return form, rootState, nil
}

func detailTokenForm(tokenName string) (*huh.Form, formState, error) {
	var options []huh.Option[string]
	options = append(options, huh.NewOption("...Back", "..."))
	options = append(options, huh.NewOption("Set as default", "set-as-default-cmd"))
	options = append(options, huh.NewOption("Login", "login-cmd"))
	options = append(options, huh.NewOption("Remove", "remove-cmd"))

	var cr *config.Credential
	crs, err := config.ReadCredentialsFile()
	if err != nil {
		return nil, detailtokenState, err
	}
	for n, _cr := range crs.Credentials {
		if n == tokenName {
			cr = &_cr
		}
	}
	if cr == nil {
		return nil, detailtokenState, fmt.Errorf("cannot found token")
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title(*cr.Access).
				Options(options...),
		)).WithShowHelp(true).WithTheme(theme.HuhTheme())
	return form, detailtokenState, nil
}

func newModel() (AuthForm, error) {
	f, s, err := rootForm()
	return AuthForm{
		state:    s,
		mainForm: f,
	}, err
}

func (m *AuthForm) resetSpinner(text string) tea.Cmd {
	m.spinner = spinner.New()
	m.spinnerText = text
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	return m.spinner.Tick
}

func (m *AuthForm) backToRoot() tea.Cmd {
	var err error
	m.mainForm, m.state, err = rootForm()
	if err != nil {
		m.mainForm, m.state = emptyForm()
		cast.LogErr(err.Error(), nil)
		return QuitAuthCmd()
	}
	return nil
}

func (m *AuthForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.state == runningState {
		switch msg := msg.(type) {
		case SetTokenAsDefaultResultMsg:
			if msg.Err != nil {
				cast.LogErr(msg.Err.Error(), nil)
			}
			cmd := m.backToRoot()
			cmds = append(cmds, cmd)
		case RemoveTokenResultMsg:
			if msg.Err != nil {
				if errors.Is(msg.Err, context.Canceled) {
					cast.LogErr("remove token request aborted by user", nil)
				} else {
					cast.LogErr(msg.Err.Error(), nil)
				}
			}
			cmd := m.backToRoot()
			cmds = append(cmds, cmd)
		case RequestNewTokenResultMsg:
			if msg.Err != nil {
				if errors.Is(msg.Err, context.Canceled) {
					cast.LogErr("token request aborted by user", nil)
				} else {
					cast.LogErr(msg.Err.Error(), nil)
				}
			}
			cmd := m.backToRoot()
			cmds = append(cmds, cmd)
		case LoginResultMsg:
			if msg.Err != nil {
				cast.LogErr(msg.Err.Error(), nil)
			}
			cmd := m.backToRoot()
			cmds = append(cmds, cmd)
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				m.runningCancelFunc()
			default:
			}
		}
		return m, tea.Batch(cmds...)
	}

	form, cmd := m.mainForm.Update(msg)
	cmds = append(cmds, cmd)
	if f, ok := form.(*huh.Form); ok {
		if f.State == huh.StateCompleted {
			kl := f.GetString("value")
			switch kl {
			case "new-token-cmd":
				m.mainForm, m.state = emptyForm()
				cmd := m.resetSpinner("Requesting new token...")
				ctx, cancel := context.WithCancel(context.Background())
				m.runningCancelFunc = cancel
				cmd2 := RequestNewTokenCmd(ctx)
				cmds = append(cmds, cmd, cmd2)
				m.state = runningState
			case "set-as-default-cmd":
				m.mainForm, m.state = emptyForm()
				cmd := m.resetSpinner("Setting token as default...")
				cmd2 := SetTokenAsDefaultCmd(m.selectedTokenName)
				cmds = append(cmds, cmd, cmd2)
				m.state = runningState
			case "remove-cmd":
				m.mainForm, m.state = emptyForm()
				cmd := m.resetSpinner("Removing token...")
				cmd2 := RemoveTokenCmd(m.selectedTokenName)
				cmds = append(cmds, cmd, cmd2)
				m.state = runningState
			case "login-cmd":
				m.mainForm, m.state = emptyForm()
				cmd := m.resetSpinner("Login with token...")
				ctx, cancel := context.WithCancel(context.Background())
				m.runningCancelFunc = cancel
				cmd2 := LoginCmd(ctx, m.selectedTokenName)
				cmds = append(cmds, cmd, cmd2)
				m.state = runningState
			case "", "..":
				m.mainForm, m.state = emptyForm()
				cmds = append(cmds, QuitAuthCmd())
			case "...":
				cmd := m.backToRoot()
				cmds = append(cmds, cmd)
			default:
				var err error
				tname := strings.TrimPrefix(kl, "profile-")
				m.selectedTokenName = tname
				m.mainForm, m.state, err = detailTokenForm(tname)
				if err != nil {
					cast.LogErr(err.Error(), nil)
					cmd := m.backToRoot()
					cmds = append(cmds, cmd)
				}
			}
			// m.mainForm, m.state = rootForm()
		} else if f.State == huh.StateAborted {
			// maybe return to uiconsole or launch tea.quit?
			cmd := m.backToRoot()
			cmds = append(cmds, cmd)
		} else {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "d":
					// set profile as default
					// cmds = append(cmds, SetDefaultProfile())
				case "r":
					// remove profile
				case "l":
					// login?
				case "esc", "q", "backspace", "ctr+c":
					m.mainForm, m.state = emptyForm()
					cmds = append(cmds, QuitAuthCmd())
				default:
					m.mainForm = f
				}
			default:
				m.mainForm = f
			}
		}
		// f.State = huh.StateNormal
	} else {
		// non-form, here is the browser model
	}

	switch msg.(type) {
	case QuitAuthMsg:
		m.mainForm, m.state = emptyForm()
	}

	return m, tea.Batch(cmds...)
}

func (m *AuthForm) View() string {
	if m.state == runningState {
		return lipgloss.JoinVertical(
			lipgloss.Top,
			fmt.Sprintf("\n %s%s\n\n", m.spinner.View(), spinnerTextStyle(m.spinnerText)),
			keyStyle.Render("\n\nctrl+c")+descStyle.Render(" cancel\n"))
	}

	return m.mainForm.View()
}

func QuitAuthCmd() tea.Cmd {
	return func() tea.Msg {
		return QuitAuthMsg{}
	}
}

func LoginCmd(ctx context.Context, tokenName string) tea.Cmd {
	return func() tea.Msg {
		_, err := config.LoginWithCredentialName(ctx, tokenName)
		return LoginResultMsg{Err: err}
	}
}

func RequestNewTokenCmd(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		err := config.RequestToken(ctx)
		return RequestNewTokenResultMsg{Err: err}
	}
}

func SetTokenAsDefaultCmd(tokenName string) tea.Cmd {
	return func() tea.Msg {
		err := config.SetTokenAsDefault(tokenName)
		return SetTokenAsDefaultResultMsg{Err: err}
	}
}

func RemoveTokenCmd(tokenName string) tea.Cmd {
	return func() tea.Msg {
		err := config.RemoveToken(tokenName)
		return RemoveTokenResultMsg{Err: err}
	}
}

func (m *AuthForm) Init() tea.Cmd {
	return m.mainForm.Init()
}

func New() (*AuthForm, error) {
	m, err := newModel()
	return &m, err
}
