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

package actors

// Considerations:
// - Only one instance of copyActor per script or cmd. Keep in mind that for each
// execution there must be an output and it must be stored, so the functionality
// of executing multiple scripts with an instance of copyActor should not be
// implemented.
//

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/develatio/nebulant-cli/base"
	nebulantssh "github.com/develatio/nebulant-cli/netproto/ssh"
	"github.com/develatio/nebulant-cli/util"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"
)

type scpCopyParametersPath struct {
	Dst       *string `json:"dest"`
	Overwrite bool    `json:"overwrite"`
	Recursive bool    `json:"recursive"`
	Src       *string `json:"src"`
}

type scpCopyParameters struct {
	Source               *string                 `json:"source"`
	Target               *string                 `json:"target"`
	Port                 uint16                  `json:"port"`
	Paths                []scpCopyParametersPath `json:"paths" validate:"required"`
	Username             *string                 `json:"username" validate:"required"`
	PrivateKeyPath       *string                 `json:"privkeyPath"`
	PrivateKeyPassphrase *string                 `json:"passphrase"`
	PrivateKey           *string                 `json:"privkey"`
	Password             *string                 `json:"password"`
}

func getSshConfig(params *scpCopyParameters) (*ssh.ClientConfig, error) {
	var err error
	sshConfig := &ssh.ClientConfig{
		User: *params.Username,
		//#nosec G106 -- Allow config this? Hacker comunity feedback needed.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if len(params.Paths) <= 0 {
		return nil, fmt.Errorf("please set source and destination paths")
	}

	if params.PrivateKeyPath != nil {
		var key []byte
		key, err = ioutil.ReadFile(*params.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		// Create the Signer for this private key.
		var signer ssh.Signer
		if params.PrivateKeyPassphrase != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(*params.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else if params.PrivateKey != nil {
		// Create the Signer for this private key.
		var signer ssh.Signer
		if params.PrivateKeyPassphrase != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(*params.PrivateKey), []byte(*params.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(*params.PrivateKey))
		}
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else if params.Password != nil {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*params.Password),
		}
	} else {
		// Use ssh agent for auth
		sshAgent, err := nebulantssh.GetSSHAgentClient()
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(sshAgent.Signers),
		}
	}
	return sshConfig, nil
}

func RemoteCopy(ctx *ActionContext) (*base.ActionOutput, error) {
	return ScpCopy(ctx)
}

func ScpCopy(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	params := new(scpCopyParameters)
	err = util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if err != nil {
		return nil, err
	}

	if (params.Source != nil && params.Target != nil) || (params.Source == nil && params.Target == nil) {
		return nil, fmt.Errorf("please set source OR target")
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	upload := true
	// Remote to local (upload)
	remoteAddress := params.Target
	// Remote to local (download)
	if params.Source != nil {
		upload = false
		remoteAddress = params.Source
	}

	err = ctx.Store.Interpolate(remoteAddress)
	if err != nil {
		return nil, err
	}

	if strings.Trim(*remoteAddress, " ") == "" {
		return nil, fmt.Errorf("the target addr is empty. Please provide one")
	}

	ctx.Logger.LogDebug("Connecting to " + *remoteAddress + " to upload files...")

	sshConfig, sshConfErr := getSshConfig(params)
	if sshConfErr != nil {
		return nil, sshConfErr
	}

	port := "22"
	if params.Port != 0 {
		port = fmt.Sprintf("%d", params.Port)
	}
	*remoteAddress = *remoteAddress + ":" + port

	scpClient, err := scp.NewClient(*remoteAddress, sshConfig, &scp.ClientOption{})
	if err != nil {
		return nil, err
	}
	defer scpClient.Close()

	var copyErr error
	for i := 0; i < len(params.Paths); i++ {
		if upload {
			ctx.Logger.LogInfo("Uploading " + *params.Paths[i].Src + " to " + *remoteAddress + ":" + *params.Paths[i].Dst + " ...")
			if params.Paths[i].Recursive {
				copyErr = scpClient.CopyDirToRemote(*params.Paths[i].Src, *params.Paths[i].Dst, &scp.DirTransferOption{})
			} else {
				copyErr = scpClient.CopyFileToRemote(*params.Paths[i].Src, *params.Paths[i].Dst, &scp.FileTransferOption{})
			}
		} else {
			ctx.Logger.LogInfo("Downloading " + *remoteAddress + ":" + *params.Paths[i].Src + " to " + *params.Paths[i].Dst + " ...")
			if params.Paths[i].Recursive {
				copyErr = scpClient.CopyDirFromRemote(*params.Paths[i].Src, *params.Paths[i].Dst, &scp.DirTransferOption{})
			} else {
				copyErr = scpClient.CopyFileFromRemote(*params.Paths[i].Src, *params.Paths[i].Dst, &scp.FileTransferOption{})
			}
		}
		if copyErr != nil {
			return nil, copyErr
		}
	}

	return nil, nil
}
