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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

//
// Ej:
// {
// 	"default": {
// 		"auth_token": "TOKENHASH"
// 	}
// }
// Credentials struct
type Credentials struct {
	Version     string                `json:"version"`
	Credentials map[string]Credential `json:"credentials"`
}

// Credential struct
type Credential struct {
	AuthToken *string `json:"auth_token"`
}

// ReadCredential func
func ReadCredential(credentialName string) (*Credential, error) {
	var userHomePath string
	if runtime.GOOS == "windows" {
		userHomePath = os.Getenv("USERPROFILE")
	} else {
		userHomePath = os.Getenv("HOME")
	}
	credentialsPath := filepath.Join(userHomePath, ".nebulant", "credentials")

	jsonFile, err := os.Open(credentialsPath) //#nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var crs Credentials
	if err := json.Unmarshal(byteValue, &crs); err != nil {
		return nil, err
	}

	if credential, exists := crs.Credentials[credentialName]; exists {
		return &credential, nil
	}
	return nil, fmt.Errorf("Credential not found")
}
