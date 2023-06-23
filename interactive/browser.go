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
	"strconv"

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
	Name            string `json:"name"`
	UUID            string `json:"uuid"`
	Description     string `json:"description"`
	BlueprintsCount int    `json:"n_blueprints"`
}

type resultsProject struct {
	Count   int                  `json:"count"`
	Results []*projectSerializer `json:"results"`
}

type blueprintSerializer struct {
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
	Description string `json:"description"`
}

type resultsBlueprint struct {
	Count   int                    `json:"count"`
	Results []*blueprintSerializer `json:"results"`
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
	jar, err := config.Login(nil)
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
	rawbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 399 {
		fmt.Println()
		return nil, fmt.Errorf(strconv.Itoa(resp.StatusCode) + " server error: " + string(rawbody))
	}
	return rawbody, nil
}

func Browser() error {
	if config.CREDENTIAL.AuthToken == nil {
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
		Active:   term.EmojiSet["BackhandIndexPointingRight"] + " {{ .Name | magenta }} {{if .UUID}} ({{ .BlueprintsCount | red }}) {{end}}",
		Inactive: "   {{ .Name | cyan }} {{if .UUID}} ({{ .BlueprintsCount | red }}) {{end}}",
		Selected: "{{if .UUID}} " + term.EmojiSet["ThumbsUpSign"] + " {{ .Name | magenta }} {{end}}",
		Details: `{{if .UUID}}
-------------------- Project --------------------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ "Blueprint:" | faint }}	{{ .BlueprintsCount }}{{end}}`,
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

		term.PrintInfo("Looking for blueprints...\n")
		data, err := httpReq("GET", "project/"+item.UUID+"/blueprint/", nil)
		if err != nil {
			return err
		}
		rd := &resultsBlueprint{}
		if err := json.Unmarshal(data, rd); err != nil {
			return err
		}
		rd.Results = append(rd.Results, &blueprintSerializer{Name: "Back"})
		err = promptBlueprint(rd.Results)
		if err != nil {
			return err
		}
	}
	return nil
}

func promptBlueprint(projects []*blueprintSerializer) error {
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
		Active:   term.EmojiSet["BackhandIndexPointingRight"] + " {{ .Name | magenta }}",
		Inactive: "   {{ .Name | cyan }}",
		Selected: "{{if .UUID}} " + term.EmojiSet["Rocket"] + " {{ .Name | red }} {{end}}",
		Details: `{{if .UUID}}
-------------------- Blueprint --------------------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}{{end}}`,
	}
L:
	for {
		prompt := promptui.Select{
			Label:     "Select Blueprint",
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
