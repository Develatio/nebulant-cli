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
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/tui/theme"
	"github.com/develatio/nebulant-cli/tui/uiauth"
	"github.com/develatio/nebulant-cli/tui/uibrowser"
	"github.com/develatio/nebulant-cli/tui/uifilepicker"
	"github.com/develatio/nebulant-cli/tuicmd"
)

type formState uint

const (
	defaultTime           = time.Minute
	rootState   formState = iota
	filepickerState
	browserState
	authState
	emptyState
	pressEnterState
	runningBlueprintState
)

type QuitMenuMsg struct{}

type Menu struct {
	state    formState
	mainForm tea.Model
}

func emptyForm() (*huh.Form, formState) {
	return huh.NewForm(
		huh.NewGroup(huh.NewNote())), emptyState
}

func spanString(s string) string {
	return s + strings.Repeat(" ", 10-len(s))
}

func rootForm() (*huh.Form, formState) {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title("Main menu. What do you want to do?").
				Options(
					huh.NewOption(spanString("Serve")+"Start server mode at "+net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT), "serve-cmd"),
					huh.NewOption(spanString("Build")+"Open builder app into the web browser", "build-cmd"),
					huh.NewOption(spanString("Panel")+"Open panel app into the web browser", "panel-cmd"),
					huh.NewOption(spanString("Path")+"Manually indicates the path to a blueprint", "filepicker-form"),
					huh.NewOption(spanString("Auth")+"Manage tokens", "auth-cmd"),
					huh.NewOption(spanString("Browse")+"Browse and run the blueprints stored in your account", "browser-form"),
					huh.NewOption(spanString("Exit")+"Exit Nebulant CLI", "exit-cmd"),
				),
		)).WithTheme(theme.HuhTheme()), rootState
}

func FilePickerForm() (*uifilepicker.FilePickerForm, formState, error) {
	fp, err := uifilepicker.New()
	if err != nil {
		return nil, emptyState, err
	}
	return fp, filepickerState, nil
}

func BrowserForm() (*uibrowser.BrowserForm, formState, error) {
	br, err := uibrowser.New()
	if err != nil {
		return nil, emptyState, err
	}
	return br, browserState, nil
}

func AuthForm() (*uiauth.AuthForm, formState, error) {
	br, err := uiauth.New()
	if err != nil {
		return nil, emptyState, err
	}
	return br, authState, nil
}

func newModel() Menu {
	f, s := rootForm()
	return Menu{
		state:    s,
		mainForm: f,
	}
}

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := m.mainForm.Update(msg)
	cmds = append(cmds, cmd)
	if f, ok := form.(*huh.Form); ok {
		if f.State == huh.StateCompleted {
			kl := f.GetString("value")
			switch kl {
			case "":
				m.mainForm, m.state = rootForm()
			case "panel-cmd":
				cmds = append(cmds, tuicmd.OpenPannelCmd())
				m.mainForm, m.state = rootForm()
			case "build-cmd":
				cmds = append(cmds, tuicmd.OpenBuilderCmd())
				m.mainForm, m.state = rootForm()
			case "serve-cmd":
				// cmds = append(cmds, tui.StartBuilderCmd())
				cmds = append(cmds, startServerModeCmd(), QuitMenuCmd())
				m.mainForm, m.state = emptyForm()
			case "filepicker-form":
				var err error
				m.mainForm, m.state, err = FilePickerForm()
				cmds = append(cmds, m.mainForm.Init())
				if err != nil {
					cast.LogErr(err.Error(), nil)
					m.mainForm, m.state = rootForm()
				}
			case "browser-form":
				var err error
				m.mainForm, m.state, err = BrowserForm()
				if err != nil {
					cast.LogErr(err.Error(), nil)
					m.mainForm, m.state = rootForm()
				}
			case "auth-cmd":
				var err error
				m.mainForm, m.state, err = AuthForm()
				if err != nil {
					cast.LogErr(err.Error(), nil)
					m.mainForm, m.state = rootForm()
				}
			case "exit-cmd":
				cmds = append(cmds, tuicmd.StartQuitState())
				m.mainForm, m.state = rootForm()
			default:
				m.mainForm, m.state = rootForm()
			}
			// m.mainForm, m.state = rootForm()
		} else if f.State == huh.StateAborted {
			// maybe return to uiconsole or launch tea.quit?
			m.mainForm, m.state = rootForm()
		} else {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "esc":
				case "h":
					m.mainForm = f
					// TOOD: show help¿
				default:
					m.mainForm = f
				}
			default:
				m.mainForm = f
			}
		}
		// f.State = huh.StateNormal
	} else {
		// non-form, here is the browser model or the filepicker model
	}

	switch msg.(type) {
	case uibrowser.QuitBrowserMsg:
		m.mainForm, m.state = rootForm()
	case uifilepicker.QuitFilePickerMsg:
		m.mainForm, m.state = rootForm()
	case uiauth.QuitAuthMsg:
		m.mainForm, m.state = rootForm()
	}

	return m, tea.Batch(cmds...)
}

func (m *Menu) View() string {
	return m.mainForm.View()
}

func startServerModeCmd() tea.Cmd {
	return func() tea.Msg {
		// TODO starting server mode cannot go back to menu
		// so we dont need uimenu anymore, exit from uimenu :)
		// err := executive.InitDirector(true, true) // Server mode
		// if err != nil {
		// 	cast.LogErr(err.Error(), nil)
		// }
		if executive.MDirector == nil {
			err := executive.InitDirector(true, true) // Server mode
			if err != nil {
				cast.LogErr(err.Error(), nil)
				panic(err.Error())
			}
		}
		errc := executive.InitServerMode()
		err := <-errc
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				cast.LogErr(err.Error(), nil)
			}
		}
		executive.MDirector.Wait()
		return nil // TODO: return servermode stopped msg?
	}
}

func QuitMenuCmd() tea.Cmd {
	return func() tea.Msg {
		return QuitMenuMsg{}
	}
}

func (m *Menu) Init() tea.Cmd {
	return m.mainForm.Init()
}

func New() *Menu {
	m := newModel()
	return &m
}
