//go:build !js

// Nebulant
// Copyright (C) 2022  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package interactive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
	"github.com/manifoldco/promptui"
)

var organizationUUID = ""

type organizationSerializer struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type meSerializer struct {
	Organization *organizationSerializer `json:"organization"`
}

type projectSerializer struct {
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	Description   string `json:"description"`
	DiagramsCount int    `json:"n_diagrams"`
}

type resultsProject struct {
	Count   int                  `json:"count"`
	Results []*projectSerializer `json:"results"`
}

type diagramSerializer struct {
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
	Description string `json:"description"`
}

type resultsDiagram struct {
	Count   int                  `json:"count"`
	Results []*diagramSerializer `json:"results"`
}

func httpReq(method string, path string, body interface{}) ([]byte, error) {
	url := url.URL{
		Scheme: config.BackendProto,
		Host:   config.BackendURLDomain,
		Path:   "/apiv1/" + path,
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

	req.Header.Set("Authorization", *config.CREDENTIALS.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rawbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rawbody, nil
}

func Browser() error {
	if config.CREDENTIALS.AuthToken == nil {
		return fmt.Errorf("auth token not found. Please set NEBULANT_TOKEN_ID and NEBULANT_TOKEN_SECRET env vars")
	}

	term.PrintInfo("Looking for projects...\n")
	data, err := httpReq("GET", "authx/me/", nil)
	if err != nil {
		return err
	}
	me := &meSerializer{}
	if err := json.Unmarshal(data, me); err != nil {
		return err
	}

	// store org uuid
	organizationUUID = me.Organization.UUID

	data, err = httpReq("GET", "organization/"+organizationUUID+"/project/", nil)
	if err != nil {
		return err
	}
	rp := &resultsProject{}
	if err := json.Unmarshal(data, rp); err != nil {
		return err
	}
	rp.Results = append(rp.Results, &projectSerializer{Name: "Back"})
	fmt.Printf("\r")
	err = promptProject(rp.Results)
	if err != nil {
		return err
	}
	return nil
}

func promptProject(projects []*projectSerializer) error {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U0001F449 {{ .Name | magenta }} {{if .UUID}} ({{ .DiagramsCount | red }}) {{end}}",
		Inactive: "   {{ .Name | cyan }} {{if .UUID}} ({{ .DiagramsCount | red }}) {{end}}",
		Selected: "{{if .UUID}} \U0001F44D {{ .Name | magenta }} {{end}}",
		Details: `{{if .UUID}}
-------------------- Project --------------------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ "Diagrams:" | faint }}	{{ .DiagramsCount }}{{end}}`,
	}
L:
	for {
		prompt := promptui.Select{
			Label:     "Select Project",
			Items:     projects,
			Templates: templates,
			Stdout:    term.NoBellStdout,
			HideHelp:  true,
		}
		i, _, err := prompt.Run()
		item := projects[i]

		if err != nil {
			return err
		}
		if item.UUID == "" {
			break L
		}

		term.PrintInfo("Looking for diagrams...\n")
		data, err := httpReq("GET", "project/"+item.UUID+"/diagram/", nil)
		if err != nil {
			return err
		}
		rd := &resultsDiagram{}
		if err := json.Unmarshal(data, rd); err != nil {
			return err
		}
		rd.Results = append(rd.Results, &diagramSerializer{Name: "Back"})
		err = promptDiagram(rd.Results)
		if err != nil {
			return err
		}
	}
	return nil
}

func promptDiagram(projects []*diagramSerializer) error {
	defer func() {
		if r := recover(); r != nil {
			cast.LogErr("Unrecoverable error found. Feel free to send us feedback", nil)
			switch r := r.(type) {
			case *util.PanicData:
				v := fmt.Sprintf("%v", r.PanicValue)
				cast.LogErr(v, nil)
				cast.LogErr(string(r.PanicTrace), nil)
			default:
				cast.LogErr("Panic", nil)
				cast.LogErr("If you think this is a bug,", nil)
				cast.LogErr("please consider posting stack trace as a GitHub", nil)
				cast.LogErr("issue (https://github.com/develatio/nebulant-cli/issues)", nil)
				cast.LogErr("Stack Trace:", nil)
				cast.LogErr("Panic", nil)
				v := fmt.Sprintf("%v", r)
				cast.LogErr(v, nil)
				cast.LogErr(string(debug.Stack()), nil)
			}
		}
	}()

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U0001F449 {{ .Name | magenta }}",
		Inactive: "   {{ .Name | cyan }}",
		Selected: "{{if .UUID}} \U0001F680 {{ .Name | red }} {{end}}",
		Details: `{{if .UUID}}
-------------------- Diagram --------------------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}{{end}}`,
	}
L:
	for {
		prompt := promptui.Select{
			Label:     "Select Diagram",
			Items:     projects,
			Templates: templates,
			Stdout:    term.NoBellStdout,
			HideHelp:  true,
		}
		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		item := projects[i]

		if item.UUID == "" {
			break L
		}

		cast.LogInfo("Obtaining blueprint "+item.UUID+"...", nil)
		irb, err := blueprint.NewIRBFromAny("nebulant://" + item.UUID)
		if err != nil {
			return err
		}
		err = executive.InitDirector(false, true)
		if err != nil {
			return err
		}
		executive.MDirector.HandleIRB <- irb
		executive.MDirector.Wait()
		executive.MDirector.Clean()
	}
	return nil
}
