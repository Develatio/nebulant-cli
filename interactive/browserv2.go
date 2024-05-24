//go:build !js

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

package interactive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
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
	options = append(options, huh.NewOption("...Back", ""))
	for _, coll := range collections {
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
		err = promptBlueprintv2(nblc, rd.Results)
		if err != nil {
			return err
		}
	}
	return nil
}

func promptBlueprintv2(nblc *subsystem.NBLcommand, blueprints []*blueprintSerializerv2) error {
	var options []huh.Option[string]
	mopts := make(map[string]*blueprintSerializerv2)
	options = append(options, huh.NewOption("...Back", ""))
	for _, bp := range blueprints {
		mopts[bp.Slug] = bp
		opt := huh.NewOption(bp.Name, bp.Slug)
		options = append(options, opt)
	}

L:
	for {
		var bpSlug string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose bp").
					Options(
						options...,
					).
					Value(&bpSlug),
			)).Run()
		if err != nil {
			return err
		}

		if bpSlug == "" {
			break L
		}

		bp := mopts[bpSlug]
		term.PrintInfo("Looking for blueprints...\n")
		data, err := httpReqv2("GET", fmt.Sprintf(config.BACKEND_SNAPSHOTS_LIST_PATH, bp.CollectionSlug, bpSlug), nil)
		if err != nil {
			return err
		}
		rd := &resultsSnapshots{}
		if err := json.Unmarshal(data, rd); err != nil {
			return err
		}

		var vopts []huh.Option[string]
		vopts = append(vopts, huh.NewOption("...Back", ""))
		vopts = append(vopts, huh.NewOption("Current (no version)", "no version"))
		for _, bp := range rd.Results {
			opt := huh.NewOption(bp.Version, bp.Version)
			vopts = append(vopts, opt)
		}

		var bpVersion string
		for {
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose version").
						Options(
							vopts...,
						).
						Value(&bpVersion),
				)).Run()
			if err != nil {
				return err
			}

			if bpVersion == "" {
				continue
			}
			break
		}
		if bpVersion == "no version" {
			bpVersion = ""
		}

		cast.LogInfo("Obtaining blueprint "+bpSlug+"...", nil)
		nbPath := fmt.Sprintf("%s/%s/%s:%s", bp.OrganizationSlug, bp.CollectionSlug, bpSlug, bpVersion)
		bpUrl, err := blueprint.ParseURL(nbPath)
		if err != nil {
			return err
		}
		irb, err := blueprint.NewIRBFromAny(bpUrl, &blueprint.IRBGenConfig{})
		if err != nil {
			return err
		}
		err = executive.InitDirector(false, true)
		if err != nil {
			return err
		}
		executive.MDirector.HandleIRB <- &executive.HandleIRBConfig{IRB: irb}
		executive.MDirector.Wait()
		executive.RemoveDirector()
	}
	return nil
}
