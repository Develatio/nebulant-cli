// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

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
//
// The code of this file was bassed on WebSocket Chat example from
// gorilla websocket lib: https://github.com/gorilla/websocket/blob/master/examples/chat/client.go

package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/util"
	"github.com/pkg/browser"
	"golang.org/x/crypto/ssh"
)

var cachedjar *cookiejar.Jar

type credentialsStoreV1 struct {
	Version     int                   `json:"version"`
	Credentials map[string]Credential `json:"credentials"`
}

// Ej:
//
//	{
//		"default": {
//			"auth_token": "TOKENHASH"
//		}
//	}
//
// Credentials struct
type CredentialsStore struct {
	Version     string                `json:"version"`
	Credentials map[string]Credential `json:"credentials"`
}

// Credential struct
type Credential struct {
	// AuthToken.uuid
	Access *string `json:"uuid"`
	// pwd:ssh-rsa
	AuthToken *string `json:"auth_token"`
}

func readCredentialsFile() (*CredentialsStore, error) {
	credentialsPath := filepath.Join(AppHomePath(), "credentials")

	jsonFile, err := os.Open(credentialsPath) //#nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var crs CredentialsStore
	if err := json.Unmarshal(byteValue, &crs); err != nil {
		var crsv1 credentialsStoreV1
		if err2 := json.Unmarshal(byteValue, &crsv1); err2 != nil {
			// use erros.Join() when go 1.20
			return nil, err2
		}

		crs.Version = "2"
		crs.Credentials = crsv1.Credentials
		return &crs, nil
	}
	return &crs, nil
}

func saveCredentialsFile(crs *CredentialsStore) (int, error) {
	credentialsPath := filepath.Join(AppHomePath(), "credentials")

	file, err := os.OpenFile(credentialsPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec G304-- file inclusion from var needed
	if err != nil {
		return 0, err
	}
	defer file.Close()
	data, err := json.Marshal(crs)
	if err != nil {
		return 0, err
	}
	return file.Write(data)
}

// ReadCredential func
func ReadCredential(credentialName string) (*Credential, error) {
	crs, err := readCredentialsFile()
	if err != nil {
		return nil, err
	}
	if credential, exists := crs.Credentials[credentialName]; exists {
		return &credential, nil
	}
	return nil, fmt.Errorf("Credential not found")
}

func Login(credential *Credential) (*cookiejar.Jar, error) {
	// TODO: test expiration and re-login
	if cachedjar != nil {
		return cachedjar, nil
	}
	if credential == nil {
		credential = CREDENTIAL
	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := http.Client{
		Transport: tr,
		Jar:       jar,
	}
	sso_login_url := url.URL{
		Scheme: BackendProto,
		Host:   BackendURLDomain,
		Path:   "/apiv1/authx/sso/login/",
	}

	pwd := strings.Split(*credential.AuthToken, ":")[0]
	esecret, err := encrypt(credential, []byte(pwd))
	if err != nil {
		return nil, err
	}
	body := []byte(`{
		"access": "` + *credential.Access + `",
		"secret": "` + esecret + `"
	}`)
	resp, err := c.Post(sso_login_url.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot login :(")
		}
		return nil, fmt.Errorf(string(b))
	}

	cachedjar = jar

	return jar, nil
}

func RequestToken() error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	c := http.Client{Transport: tr}
	sso_url := url.URL{
		Scheme: BackendProto,
		Host:   BackendURLDomain,
		Path:   "/apiv1/authx/sso/",
	}
	body := []byte(`{
		"description": "Nebulant CLI"
	}`)
	resp, err := c.Post(sso_url.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// r := bufio.NewReader(resp.Body)

	panel_url := url.URL{
		Scheme: BackendProto,
		Host:   PanelURLDomain,
		Path:   "organization/tokens/sso/" + resp.Header.Get("url-token"),
	}

	account_url := url.URL{
		Scheme:   BackendProto,
		Host:     AccountURLDomain,
		Path:     "/to/",
		RawQuery: "path=" + panel_url.String(),
	}
	err = browser.OpenURL(account_url.String())
	if err != nil {
		return err
	}

	// WIP
	rawbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	offset := 0
	for ; offset < len(rawbody); offset++ {
		if rawbody[offset] == 123 {
			break
		}
	}

	var credential Credential
	if err := util.UnmarshalValidJSON(rawbody[offset:], &credential); err != nil {
		return err
	}

	crs, err := readCredentialsFile()
	if err != nil {
		return err
	}

	crs.Credentials[ACTIVE_CONF_PROFILE] = credential
	_, err = saveCredentialsFile(crs)
	if err != nil {
		return err
	}

	return nil
}

func encrypt(credential *Credential, secret []byte) (string, error) {
	if credential == nil {
		credential = CREDENTIAL
	}
	if credential.AuthToken == nil {
		return "", fmt.Errorf("no auth token found")
	}
	atk := strings.Split(*credential.AuthToken, ":")

	_pbk, _, _, _, err := ssh.ParseAuthorizedKey([]byte("ssh-rsa " + atk[1]))
	if err != nil {
		return "", err
	}
	_cpbk := _pbk.(ssh.CryptoPublicKey).CryptoPublicKey()
	pubkey := _cpbk.(*rsa.PublicKey)

	random := rand.Reader
	ciphtxt, err := rsa.EncryptOAEP(sha256.New(), random, pubkey, secret, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(ciphtxt)
	return encoded, nil
}
