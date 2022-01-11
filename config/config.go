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
	"log"
	"os"
	"strconv"
)

// Version var
var Version = "DEV build"

// VersionDate var
var VersionDate = ""

// WSScheme var
var WSScheme string = "wss"

// BackendProto var
var BackendProto string = "https"

// BackendURLDomain var
var BackendURLDomain string = "api.nebulant.io"

// FrontOrigin var
var FrontOrigin string = "https://builder.nebulant.io"

// FrontUrl var
var FrontUrl string = "https://builder.nebulant.io"

// FrontOriginPre var
var FrontOriginPre string = "https://builder.nebulant.dev"

// DEBUG conf
var DEBUG bool = false

// PROFILING conf
var PROFILING bool = false

// BACKEND_AUTH_TOKEN conf
var BACKEND_AUTH_TOKEN = ""

// ACTIVE_CONF_PROFILE
var ACTIVE_CONF_PROFILE = "default"

// CREDENTIALS
var CREDENTIALS *Credential = &Credential{}

// Order of credentials:
// * Environment Variables
// * Shared Credentials file
func init() {
	if os.Getenv("NEBULANT_DEBUG") != "" {
		var err error
		DEBUG, err = strconv.ParseBool(os.Getenv("NEBULANT_DEBUG"))
		if err != nil {
			log.Fatal(err)
		}
	}
	if os.Getenv("NEBULANT_PROFILING") != "" {
		var err error
		PROFILING, err = strconv.ParseBool(os.Getenv("NEBULANT_PROFILING"))
		if err != nil {
			log.Fatal(err)
		}
	}
	if os.Getenv("NEBULANT_CONF_PROFILE") != "" {
		ACTIVE_CONF_PROFILE = os.Getenv("NEBULANT_CONF_PROFILE")
	}

	// Load credentials from file
	credential, err := ReadCredential(ACTIVE_CONF_PROFILE)
	if err != nil && ACTIVE_CONF_PROFILE != "default" {
		log.Panic("Cannot read credentials from specified profile " + ACTIVE_CONF_PROFILE)
	}
	if credential != nil {
		CREDENTIALS = credential
	}

	// Use credentials from env vars if exists
	if os.Getenv("NEBULANT_TOKEN_ID") != "" && os.Getenv("NEBULANT_TOKEN_SECRET") != "" {
		var data string = os.Getenv("NEBULANT_TOKEN_ID") + ":" + os.Getenv("NEBULANT_TOKEN_SECRET")
		CREDENTIALS.AuthToken = &data
	}
}
