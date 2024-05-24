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

package actors

// Considerations:
// - Only one instance of copyActor per script or cmd. Keep in mind that for each
// execution there must be an output and it must be stored, so the functionality
// of executing multiple scripts with an instance of copyActor should not be
// implemented.
//

import (
	"fmt"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/base"
	nebulantssh "github.com/develatio/nebulant-cli/netproto/ssh"
	"github.com/develatio/nebulant-cli/util"
	"github.com/povsister/scp"
)

type scpCopyParametersPath struct {
	Dst       *string `json:"dest"`
	Overwrite bool    `json:"overwrite"`
	Recursive bool    `json:"recursive"`
	Src       *string `json:"src"`
}

type scpCopyParameters struct {
	nebulantssh.ClientConfigParameters
	Source *string                 `json:"source"`
	Paths  []scpCopyParametersPath `json:"paths" validate:"required"`
}

func RemoteCopy(ctx *ActionContext) (*base.ActionOutput, error) {
	return ScpCopy(ctx)
}

func ScpCopy(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	var i int
	params := new(scpCopyParameters)
	err = util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if err != nil {
		return nil, err
	}

	if (params.Source != nil && params.Target != nil) || (params.Source == nil && params.Target == nil) {
		return nil, fmt.Errorf("please set source OR target machine")
	}

	for i = 0; i < len(params.Paths); i++ {
		if params.Paths[i].Src == nil || len(*params.Paths[i].Src) <= 0 {
			return nil, fmt.Errorf("cannot use empty paths for remote copy")
		}
		if params.Paths[i].Dst == nil || len(*params.Paths[i].Dst) <= 0 {
			return nil, fmt.Errorf("cannot use empty paths for remote copy")
		}
	}

	if i <= 0 {
		return nil, fmt.Errorf("please set at least one path for remote copy")
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	err = ctx.Store.DeepInterpolation(params)
	if err != nil {
		return nil, err
	}

	upload := true
	// Remote to local (upload)
	remoteAddress := params.Target
	// Remote to local (download)
	if params.Source != nil {
		upload = false
		remoteAddress = params.Source
	}

	if strings.Trim(*remoteAddress, " ") == "" {
		return nil, fmt.Errorf("the target addr is empty. Please provide one")
	}

	ctx.Logger.LogDebug("Connecting to " + *remoteAddress + " to upload files...")

	var proxies []*nebulantssh.ClientConfigParameters
	for _, prx := range params.Proxies {
		raddr := prx.Target
		err = ctx.Store.Interpolate(raddr)
		if err != nil {
			return nil, err
		}
		if strings.Trim(*raddr, " ") == "" {
			return nil, fmt.Errorf("the proxy target addr is empty. Please provide one")
		}
		proxies = append(proxies, &nebulantssh.ClientConfigParameters{
			Target:               raddr,
			Port:                 prx.Port,
			Username:             prx.Username,
			PrivateKey:           prx.PrivateKey,
			PrivateKeyPath:       prx.PrivateKeyPath,
			PrivateKeyPassphrase: prx.PrivateKeyPassphrase,
			Password:             prx.Password,
			Proxies:              prx.Proxies,
		})
	}

	sshClient := nebulantssh.NewSSHClient()
	mainclient := sshClient
	out := make(chan bool)
	defer func() {
		err := mainclient.Disconnect()
		if err != nil {
			ctx.Logger.LogWarn(err.Error())
		}
		out <- true
	}()
	mainClientEvents := sshClient.Events
	go func() {
	L1:
		for {
			select {
			case evt := <-mainClientEvents:
				addr := evt.SSHClient.DialAddr
				if evt.Type == nebulantssh.SSHClientEventMasterClosed {
					ctx.Logger.LogDebug(fmt.Sprintf("SCP Closing %v...", addr))
					break L1
				}
				if evt.Type == nebulantssh.SSHClientEventDialing {
					ctx.Logger.LogDebug(fmt.Sprintf("SCP Dialing %v...", addr))
				}
				if evt.Type == nebulantssh.SSHClientEventClosed {
					ctx.Logger.LogDebug(fmt.Sprintf("SCP Closing %v...", addr))
				}
			case <-out:
				ctx.Logger.LogDebug("Should out from go routine")
				break L1
			default:
				ctx.Logger.LogDebug("Waiting scp event...")
				time.Sleep(200000 * time.Microsecond)
			}
		}
	}()
	sshClient, err = sshClient.DialWithProxies(&nebulantssh.ClientConfigParameters{
		Target:               remoteAddress,
		Port:                 params.Port,
		Username:             params.Username,
		PrivateKey:           params.PrivateKey,
		PrivateKeyPath:       params.PrivateKeyPath,
		PrivateKeyPassphrase: params.PrivateKeyPassphrase,
		Password:             params.Password,
		Proxies:              proxies,
	})
	if err != nil {
		return nil, err
	}

	scpClient, err := sshClient.NewSCPClientFromExistingSSH()
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
				ctx.Logger.LogDebug("Done upload")
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

	ctx.Logger.LogDebug("Done SCP")
	return nil, nil
}
