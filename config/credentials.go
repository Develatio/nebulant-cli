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
//
// The code of this file was bassed on WebSocket Chat example from
// gorilla websocket lib: https://github.com/gorilla/websocket/blob/master/examples/chat/client.go

package config

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	math_rand "math/rand"
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
	Version       string                `json:"version"`
	Credentials   map[string]Credential `json:"credentials"`
	ActiveProfile string                `json:"active_profile"`
}

// Credential struct
type Credential struct {
	// AuthToken.uuid
	Access *string `json:"uuid"`
	// pwd:ssh-rsa
	AuthToken *string `json:"auth_token"`
	//
	Denied bool `json:"denied"`
}

// ProfileOrganization struct
type ProfileOrganization struct {
	Name *string `json:"name"`
	Slug string  `json:"slug" validate:"required"`
}

// Profile struct
type Profile struct {
	Name         string              `json:"name"`
	Organization ProfileOrganization `json:"current_organization" validate:"required"`
}

func createEmptyCredentialsFile() (int, error) {
	credentialsPath := filepath.Join(AppHomePath(), "credentials")
	_, err := os.Stat(credentialsPath)
	if os.IsNotExist(err) {
		crs := &CredentialsStore{
			Credentials: map[string]Credential{},
		}
		return SaveCredentialsFile(crs)
	}
	return 0, nil
}

func ReadCredentialsFile() (*CredentialsStore, error) {
	credentialsPath := filepath.Join(AppHomePath(), "credentials")

	jsonFile, err := os.Open(credentialsPath) // #nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	var crs CredentialsStore
	if err := json.Unmarshal(byteValue, &crs); err != nil {
		var crsv1 credentialsStoreV1
		if err2 := json.Unmarshal(byteValue, &crsv1); err2 != nil {
			return nil, errors.Join(err, err2)
		}

		crs.Version = "2"
		crs.Credentials = crsv1.Credentials
		return &crs, nil
	}
	return &crs, nil
}

func SaveCredentialsFile(crs *CredentialsStore) (int, error) {
	credentialsPath := filepath.Join(AppHomePath(), "credentials")

	file, err := os.OpenFile(credentialsPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304-- file inclusion from var needed
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
	crs, err := ReadCredentialsFile()
	if err != nil {
		return nil, err
	}

	if credentialName == "" {
		credentialName = crs.ActiveProfile
		if credentialName == "" {
			credentialName = FALLBACK_PROFILE_NAME // default
		}
	}

	if credential, exists := crs.Credentials[credentialName]; exists {
		return &credential, nil
	}
	return nil, fmt.Errorf("Credential of profile [%s] not found", credentialName)
}

func GetJar() (*cookiejar.Jar, error) {
	if cachedjar == nil {
		_cj, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		cachedjar = _cj
	}
	return cachedjar, nil
}

func LoginWithCredentialName(ctx context.Context, name string) (*cookiejar.Jar, error) {
	crs, err := ReadCredentialsFile()
	if err != nil {
		return nil, err
	}
	if credential, exists := crs.Credentials[name]; exists {
		recovjar := cachedjar
		cachedjar = nil
		jar, err := Login(ctx, &credential)
		if err != nil {
			cachedjar = recovjar
			return nil, err
		}
		return jar, err
	} else {
		return nil, fmt.Errorf("unknown credential")
	}
}

func Login(ctx context.Context, credential *Credential) (*cookiejar.Jar, error) {
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
	jar, err := GetJar()
	if err != nil {
		return nil, err
	}
	c := http.Client{
		Transport: tr,
		Jar:       jar,
	}
	sso_login_url := url.URL{
		Scheme: BASE_SCHEME,
		Host:   BACKEND_API_HOST,
		Path:   BACKEND_SSO_LOGIN_PATH,
	}

	if credential.AuthToken == nil {
		return nil, fmt.Errorf("cannot login: empty auth token")
	}
	pwd := strings.Split(*credential.AuthToken, ":")[0]
	esecret, err := encrypt(credential, []byte(pwd))
	if err != nil {
		return nil, err
	}

	access := *credential.Access
	if len(*credential.Access) > 36 {
		baccess, err := hex.DecodeString(*credential.Access)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("bad access string"))
		}
		access = string(baccess)
		if len(*credential.Access) != 36 {
			return nil, fmt.Errorf("bad access string")
		}
	}

	body := []byte(`{
		"access": "` + access + `",
		"secret": "` + esecret + `"
	}`)

	req, err := http.NewRequest("POST", sso_login_url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	resp, err := c.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot login :(")
		}
		return nil, fmt.Errorf(string(b))
	}

	cachedjar = jar

	err = _fillProfile()
	if err != nil {
		return nil, err
	}

	return jar, nil
}

func _fillProfile() error {
	url := url.URL{
		Scheme: BASE_SCHEME,
		Host:   BACKEND_API_HOST,
		Path:   BACKEND_ME_PATH,
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}
	// assume that this func is called just after login and
	// cachedjar is always valid
	client := &http.Client{Jar: cachedjar}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rawbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 399 {
		fmt.Println()
		return fmt.Errorf("server error (%v): %s", resp.StatusCode, rawbody)
	}

	prf := &Profile{}
	if err := util.UnmarshalValidJSON(rawbody, prf); err != nil {
		return err
	}
	if prf.Organization.Slug == "" {
		return fmt.Errorf("bad company in your profile")
	}
	PROFILE = prf

	return nil
}

func SetTokenAsDefault(name string) error {
	crs, err := ReadCredentialsFile()
	if err != nil {
		return err
	}

	if _, exists := crs.Credentials[name]; !exists {
		return fmt.Errorf("unknown credential")
	}

	crs.ActiveProfile = name
	_, err = SaveCredentialsFile(crs)
	if err != nil {
		return err
	}

	cr := crs.Credentials[name]
	CREDENTIAL = &cr

	return nil
}

func RemoveToken(name string) error {
	crs, err := ReadCredentialsFile()
	if err != nil {
		return err
	}

	if _, exists := crs.Credentials[name]; !exists {
		return fmt.Errorf("unknown credential")
	}

	if crs.ActiveProfile == name {
		return fmt.Errorf("cannot delete default credential")
	}

	delete(crs.Credentials, name)

	_, err = SaveCredentialsFile(crs)
	if err != nil {
		return err
	}

	return nil
}

func RequestToken(ctx context.Context) error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}
	c := http.Client{Transport: tr}
	sso_url := url.URL{
		Scheme: BASE_SCHEME,
		Host:   BACKEND_API_HOST,
		Path:   BACKEND_REQUEST_NEW_SSO_TOKEN_PATH,
	}
	body := []byte(`{
		"description": "Nebulant CLI"
	}`)

	req, err := http.NewRequest("POST", sso_url.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	token := resp.Header.Get("url-token")
	if token == "" {
		return fmt.Errorf("empty token received")
	}

	panel_url := url.URL{
		Scheme: BASE_SCHEME,
		Host:   PANEL_HOST,
		Path:   fmt.Sprintf(PANEL_SSO_TOKEN_VALIDATION_PATH, resp.Header.Get("url-token")),
	}

	account_url := url.URL{
		Scheme:   BASE_SCHEME,
		Host:     BACKEND_ACCOUNT_HOST,
		Path:     BACKEND_ENTRY_POINT_PATH,
		RawQuery: "next=" + panel_url.String(),
	}
	err = browser.OpenURL(account_url.String())
	if err != nil {
		return err
	}

	rawbody, err := io.ReadAll(resp.Body)
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
	if credential.Denied {
		return fmt.Errorf("token request denied")
	}

	if credential.AuthToken == nil {
		return fmt.Errorf("empty token received")
	}

	if *credential.AuthToken == "" {
		return fmt.Errorf("empty token received")
	}

	crs, err := ReadCredentialsFile()
	if err != nil {
		return err
	}

	// new token, new profile
	profile := fmt.Sprintf("%d", math_rand.Int()) // #nosec G404 -- Weak random is OK here
	crs.ActiveProfile = profile
	crs.Credentials[profile] = credential
	_, err = SaveCredentialsFile(crs)
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
