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
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

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
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
	Description string `json:"description"`
}

type resultsBlueprintv2 struct {
	Count   int                      `json:"count"`
	Results []*blueprintSerializerv2 `json:"results"`
}

// TODO: move to common, this code is duplicated in
// runtime/debugger.go
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

func Browserv2(nblc *subsystem.NBLcommand) error {
	if config.CREDENTIAL.AuthToken == nil {
		return fmt.Errorf("auth token not found. Please set NEBULANT_TOKEN_ID and NEBULANT_TOKEN_SECRET environment variables or use 'nebulant auth' command to authenticate and generate a CLI token")
	}

	term.PrintInfo("Looking for collections...\n")

	data, err := httpReqv2("GET", config.BACKEND_COLLECTION_LIST_PATH, nil)
	if err != nil {
		return err
	}
	rp := &resultsCollectionv2{}
	if err := json.Unmarshal(data, rp); err != nil {
		return err
	}
	err = promptCollectionv2(nblc, rp.Results)
	if err != nil {
		return err
	}
	return nil
}

func promptCollectionv2(nblc *subsystem.NBLcommand, collections []*collectionSerializerv2) error {
	var options []huh.Option[string]
	options = append(options, huh.NewOption("...", ""))
	for _, coll := range collections {
		fmt.Println(coll)
		opt := huh.NewOption(coll.Name, coll.Slug)
		options = append(options, opt)
	}

L:
	for {
		var collSlug string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose collection").
					Options(
						options...,
					).
					Value(&collSlug),
			)).Run()
		if err != nil {
			return err
		}

		if collSlug == "" {
			break L
		}

		term.PrintInfo("Looking for blueprints...\n")
		data, err := httpReqv2("GET", fmt.Sprintf(config.BACKEND_COLLECTION_BLUEPRINT_LIST_PATH, collSlug), nil)
		if err != nil {
			return err
		}
		rd := &resultsBlueprintv2{}
		if err := json.Unmarshal(data, rd); err != nil {
			return err
		}
		fmt.Println(rd.Results)
		// rd.Results = append(rd.Results, &blueprintSerializerv2{Name: "Back"})
		err = promptBlueprintv2(nblc, rd.Results)
		if err != nil {
			return err
		}
	}
	return nil
}

func promptBlueprintv2(nblc *subsystem.NBLcommand, collections []*blueprintSerializerv2) error {
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

	var options []huh.Option[string]
	options = append(options, huh.NewOption("...", ""))
	for _, coll := range collections {
		opt := huh.NewOption(coll.Name, coll.UUID)
		options = append(options, opt)
	}

L:
	for {
		var bpUUID string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose bp").
					Options(
						options...,
					).
					Value(&bpUUID),
			)).Run()
		if err != nil {
			return err
		}

		if bpUUID == "" {
			break L
		}

		cast.LogInfo("Obtaining blueprint "+bpUUID+"...", nil)
		irb, err := blueprint.NewIRBFromAny("nebulant://"+bpUUID, &blueprint.IRBGenConfig{})
		if err != nil {
			return err
		}
		err = executive.InitDirector(false, true)
		if err != nil {
			return err
		}
		executive.MDirector.HandleIRB <- &executive.HandleIRBConfig{IRB: irb}
		executive.MDirector.Wait()
		executive.MDirector.Clean()
	}
	return nil
}
