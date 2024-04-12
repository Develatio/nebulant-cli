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
	"path/filepath"
	"runtime"
	"strconv"
)

// Version var
var Version = "DEV build"

// VersionDate var
var VersionDate = ""

// VersionCommit var
var VersionCommit = ""

// VersionGo var
var VersionGo = ""

// WSScheme var
var WSScheme string = "wss"

// BackendProto var
var BackendProto string = "https"

// BackendURLDomain var
var BackendURLDomain string = "api.nebulant.io"

// AccountURLDomain var
var AccountURLDomain string = "account.nebulant.io"

// PanelURLDomain var
var PanelURLDomain string = "panel.nebulant.io"

// FrontOrigin var
var FrontOrigin string = "https://builder.nebulant.io"

// BridgeOrigin var
var BridgeOrigin string = "https://bridge.nebulant.io"

// FrontUrl var
var FrontUrl string = "https://builder.nebulant.io"

// FrontOriginPre var
var FrontOriginPre string = "https://builder.nebulant.dev"

// DEBUG conf
var DEBUG bool = false

// PARANOICDEBUG conf
var PARANOICDEBUG bool = false

// PROFILING conf
var PROFILING bool = false

// BACKEND_AUTH_TOKEN conf
var BACKEND_AUTH_TOKEN = ""

// ACTIVE_CONF_PROFILE
var ACTIVE_CONF_PROFILE = "default"

// CREDENTIAL
var CREDENTIAL *Credential = &Credential{}

// Server addr
var SERVER_ADDR = "localhost"

// Server port
var SERVER_PORT = "15678"

// Server cert file path
var SERVER_CERT = ""

// Server key file path
var SERVER_KEY = ""

// Bridge addr
var BRIDGE_ADDR = ""

// Bridge port
var BRIDGE_PORT = "16789"

// Bridge secret
var BRIDGE_SECRET = os.Getenv("NEBULANT_BRIDGE_SECRET")

// AssetDescriptorURL conf
var AssetDescriptorURL = "https://builder-assets.nebulant.io/assets.json"

// UpdateDescriptorURL conf
var UpdateDescriptorURL string = "https://releases.nebulant.io/version.json"

// arg argv conf

var ServerModeFlag *bool
var AddrFlag *string
var BridgeAddrFlag *string
var VersionFlag *bool
var DebugFlag *bool
var ParanoicDebugFlag *bool
var Ipv6Flag *bool
var DisableColorFlag *bool
var UpgradeAssetsFlag *bool
var ForceUpgradeAssetsFlag *bool
var LookupAssetFlag *string
var ForceTerm *bool
var BuildAssetIndexFlag *string
var ForceUpgradeAssetsNoDownloadFlag *bool
var BridgeSecretFlag *string
var BridgeOriginFlag *string

var BridgeCertPathFlag *string
var BridgeKeyPathFlag *string
var BridgeXtermRootPath *string

var ForceNoTerm = false

func AppHomePath() string {
	var userHomePath string
	if runtime.GOOS == "windows" {
		userHomePath = os.Getenv("USERPROFILE")
	} else {
		userHomePath = os.Getenv("HOME")
	}
	return filepath.Join(userHomePath, ".nebulant")
}

// Order of credentials:
// * Environment Variables
// * Shared Credentials file
func init() {
	// ensure config dir
	assetsdir := filepath.Join(AppHomePath(), "assets")
	err := os.MkdirAll(assetsdir, os.ModePerm)
	if err != nil {
		log.Panic(err.Error())
	}

	// ensure credentials file
	_, err = createEmptyCredentialsFile()
	if err != nil {
		log.Panic(err.Error())
	}

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
		CREDENTIAL = credential
	}

	// Use credentials from env vars if exists
	if os.Getenv("NEBULANT_TOKEN_ID") != "" && os.Getenv("NEBULANT_TOKEN_SECRET") != "" {
		var data string = os.Getenv("NEBULANT_TOKEN_ID") + ":" + os.Getenv("NEBULANT_TOKEN_SECRET")
		CREDENTIAL.AuthToken = &data
	}
}
