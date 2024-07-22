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

package uibrowser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/tuicmd"
)

type lvlState uint

const (
	defaultTime              = time.Minute
	collectionState lvlState = iota
	blueprintState
	snapshotState
	emptyState
)

type QuitBrowserMsg struct{}

type collectionSerializerv2 struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	BlueprintsCount int    `json:"n_blueprints"`
	Slug            string `json:"slug"`
}

type resultsCollectionv2 struct {
	Count   int                       `json:"count"`
	Results []*collectionSerializerv2 `json:"results"`
}

type blueprintSerializerv2 struct {
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Description      string `json:"description"`
	CollectionSlug   string `json:"collection_slug"`
	OrganizationSlug string `json:"organization_slug"`
}

type resultsBlueprintv2 struct {
	Count   int                      `json:"count"`
	Results []*blueprintSerializerv2 `json:"results"`
}

type snapshotSerializer struct {
	Changelog      string `json:"changelog"`
	Version        string `json:"version"`
	Public         bool   `json:"public"`
	IsLatestStable bool   `json:"is_latest_stable"`
	IsLatestBeta   bool   `json:"is_latest_beta"`
	// CreatedDate    `json:"created_date"`
}

type resultsSnapshots struct {
	Count   int                   `json:"count"`
	Results []*snapshotSerializer `json:"results"`
}

func httpReqv2(method string, path string, body interface{}) ([]byte, error) {
	url := url.URL{
		Scheme: config.BASE_SCHEME,
		Host:   config.BACKEND_API_HOST,
		Path:   path,
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	reqBody := bytes.NewBuffer(rawBody)
	req, err := http.NewRequest(method, url.String(), reqBody)
	if err != nil {
		return nil, err
	}
	jar, err := config.Login(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Jar: jar}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rawbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 399 {
		fmt.Println()
		return nil, fmt.Errorf(strconv.Itoa(resp.StatusCode) + " server error: " + string(rawbody))
	}
	return rawbody, nil
}

type BrowserForm struct {
	state            lvlState
	mainForm         *huh.Form
	organizationSlug string
	collectionSlug   string
	blueprintSlug    string
}

func emptyForm() (*huh.Form, lvlState) {
	return huh.NewForm(
		huh.NewGroup(huh.NewNote())), emptyState
}

func collectionsForm() (*huh.Form, lvlState, error) {
	data, err := httpReqv2("GET", config.BACKEND_COLLECTION_LIST_PATH, nil)
	if err != nil {
		return nil, emptyState, err
	}
	rp := &resultsCollectionv2{}
	if err := json.Unmarshal(data, rp); err != nil {
		return nil, emptyState, err
	}

	var options []huh.Option[string]
	options = append(options, huh.NewOption("...Back", ".."))
	for _, coll := range rp.Results {
		opt := huh.NewOption(coll.Name, coll.Slug)
		options = append(options, opt)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title("Choose collection").
				Options(options...),
		)), collectionState, nil
}

func blueprintsForm(collSlug string) (*huh.Form, lvlState, error) {
	data, err := httpReqv2("GET", fmt.Sprintf(config.BACKEND_COLLECTION_BLUEPRINT_LIST_PATH, collSlug), nil)
	if err != nil {
		return nil, emptyState, err
	}
	rd := &resultsBlueprintv2{}
	if err := json.Unmarshal(data, rd); err != nil {
		return nil, emptyState, err
	}

	var options []huh.Option[string]
	mopts := make(map[string]*blueprintSerializerv2)
	options = append(options, huh.NewOption("...Back", ".."))
	for _, bp := range rd.Results {
		mopts[bp.Slug] = bp
		opt := huh.NewOption(bp.Name, fmt.Sprintf("%s/%s/%s", bp.OrganizationSlug, bp.CollectionSlug, bp.Slug))
		options = append(options, opt)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title("Choose bp").
				Options(options...),
		)), blueprintState, nil
}

func snapshotsForm(collectionSlug string, bpSlug string) (*huh.Form, lvlState, error) {
	data, err := httpReqv2("GET", fmt.Sprintf(config.BACKEND_SNAPSHOTS_LIST_PATH, collectionSlug, bpSlug), nil)
	if err != nil {
		return nil, emptyState, err
	}
	rd := &resultsSnapshots{}
	if err := json.Unmarshal(data, rd); err != nil {
		return nil, emptyState, err
	}

	var options []huh.Option[string]
	options = append(options, huh.NewOption("...Back", ".."))
	options = append(options, huh.NewOption("Current (no version)", "no version"))
	for _, bp := range rd.Results {
		opt := huh.NewOption(bp.Version, bp.Version)
		options = append(options, opt)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("value").
				Title("Choose version").
				Options(options...),
		)), snapshotState, nil
}

func newModel() (*BrowserForm, error) {
	f, s, err := collectionsForm()
	if err != nil {
		return nil, err
	}
	return &BrowserForm{
		state:    s,
		mainForm: f,
	}, nil
}

// type YesQuitMsg struct{}

// func (m *BrowserForm) Init() tea.Cmd {
// 	// start the timer and spinner on program start
// 	return tea.Batch(m.spinner.Tick, readCastBusCmd(m.lk))
// }

func (m *BrowserForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := m.mainForm.Update(msg)
	cmds = append(cmds, cmd)
	if f, ok := form.(*huh.Form); ok {
		if f.State == huh.StateCompleted {
			kl := f.GetString("value")
			switch kl {
			case "":
				m.mainForm, m.state = emptyForm()
				cmds = append(cmds, QuitBrowserCmd())
			case "..":
				switch m.state {
				case snapshotState:
					var err error
					m.mainForm, m.state, err = blueprintsForm(m.collectionSlug)
					if err != nil {
						m.mainForm, m.state = emptyForm()
						cast.LogErr(err.Error(), nil)
					}
				case blueprintState:
					var err error
					m.mainForm, m.state, err = collectionsForm()
					if err != nil {
						m.mainForm, m.state = emptyForm()
						cast.LogErr(err.Error(), nil)
					}
				case collectionState:
					// back on collectin list closes the browser
					m.mainForm, m.state = emptyForm()
					cmds = append(cmds, QuitBrowserCmd())
				default:
					// not implemented, force quit
					m.mainForm, m.state = emptyForm()
					cmds = append(cmds, QuitBrowserCmd())
				}
			default:
				// collection, bp or snapshot has been selected:
				switch m.state {
				case collectionState:
					var err error
					m.collectionSlug = kl
					m.mainForm, m.state, err = blueprintsForm(m.collectionSlug)
					if err != nil {
						m.mainForm, m.state = emptyForm()
						cast.LogErr(err.Error(), nil)
					}
				case blueprintState:
					var err error
					cc := strings.Split(kl, "/")
					m.organizationSlug = cc[0]
					m.collectionSlug = cc[1]
					m.blueprintSlug = cc[2]
					m.mainForm, m.state, err = snapshotsForm(m.collectionSlug, m.blueprintSlug)
					if err != nil {
						m.mainForm, m.state = emptyForm()
						cast.LogErr(err.Error(), nil)
					}
				case snapshotState:
					vv := kl
					if vv == "no version" {
						vv = ""
					}
					m.mainForm, m.state = emptyForm()
					cmds = append(cmds, msgRunRemoteStartCmd(), tuicmd.RunRemoteBPCmd(m.organizationSlug, m.collectionSlug, m.blueprintSlug, vv))
				default:
					// not implemented, force quit
					m.mainForm, m.state = emptyForm()
					cmds = append(cmds, QuitBrowserCmd())
				}

				// TODO: browse to selected elem
			}
			// m.mainForm, m.state = rootForm()
		} else if f.State == huh.StateAborted {
			m.mainForm, m.state = emptyForm()
			cmds = append(cmds, QuitBrowserCmd())
		} else {
			m.mainForm = f
		}
	}

	switch msg := msg.(type) {
	case tuicmd.RunRemoteBPCmdENDMsg:
		var err error
		if msg.Err != nil {
			cast.LogErr(msg.Err.Error(), nil)
		}
		m.mainForm, m.state, err = snapshotsForm(m.collectionSlug, m.blueprintSlug)
		if err != nil {
			m.mainForm, m.state = emptyForm()
			// TODO: feedback err
		}
	}

	return m, tea.Batch(cmds...)
}

func msgRunRemoteStartCmd() tea.Cmd {
	return func() tea.Msg {
		return tuicmd.RunRemoteBPCmdStartMsg{}
	}
}

func (m *BrowserForm) View() string {
	return m.mainForm.View()
}

func QuitBrowserCmd() tea.Cmd {
	return func() tea.Msg {
		return QuitBrowserMsg{}
	}
}

func (m *BrowserForm) Init() tea.Cmd {
	return m.mainForm.Init()
}

func New() (*BrowserForm, error) {
	m, err := newModel()
	if err != nil {
		return nil, err
	}
	return m, nil
}
