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

package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/ipc"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var injecfuncs = `nebulant_inline_helper () {
	READVARSTRICT=0
	NULL=$(echo -e "\x10")

	if [ "$1" = "readvar" ]; then
		VARNAME=$2

		if [ "$VARNAME" = "-strict" ]; then
			READVARSTRICT=1
			VARNAME=$3
		fi
		if [ "$VARNAME" = "--strict" ]; then
			READVARSTRICT=1
			VARNAME=$3
		fi

		if [ "$NEBULANT_IPCSID" = "" ]; then
			echo "cannot find IPC server ID" >&2
			return 1
		fi

		if [ "$NEBULANT_IPCCID" = "" ]; then
			echo "cannot find IPC consumer ID" >&2
			return 1
		fi

		socat -V &>/dev/null
		if [ $? -gt 0 ]; then
			echo "socat is required, please install it" >&2
			return 1
		fi

		RES=$(echo -e "$NEBULANT_IPCSID $NEBULANT_IPCCID readvar $VARNAME" | socat -,ignoreeof unix-connect:/tmp/ipc_$NEBULANT_IPCSID.sock)
		if [ $? -gt 0 ]; then
			echo "there was a problem communicating with the IPC server" >&2
			return 1
		fi
		if [ "$RES" = "$NULL" ]; then
			if [ $READVARSTRICT -eq 1 ]; then
				echo "undefined var" >&2
			fi
			return 1
		fi
		echo -e "$RES"
		return 0
	fi

	echo "nebulant-cli inline helper"
	echo "Unknow command"
	echo ""
	echo "Available commands:"
	echo "Usage: nebulant readvar [flags] [variable name]"
	echo -e "\t-strict\t\t\tForce err msg instead empty string"
	return 1
} && export -f nebulant_inline_helper`

type ClientConfigParameters struct {
	Target *string `json:"target" validate:"required"`
	Port   uint16  `json:"port"`
	//
	Username             *string `json:"username" validate:"required"`
	PrivateKey           *string `json:"privkey"`
	PrivateKeyPath       *string `json:"privkeyPath"`
	PrivateKeyPassphrase *string `json:"passphrase"`
	Password             *string `json:"password"`
	//
	Proxies []*ClientConfigParameters `json:"proxies"`
}

func GetSSHClientConfig(cc *ClientConfigParameters) (*ssh.ClientConfig, error) {
	var err error

	if cc.Username == nil {
		if cc.Target == nil {
			return nil, fmt.Errorf("please, provide username")
		}
		return nil, fmt.Errorf("please, provide username for %v", *cc.Target)
	}
	sshConfig := &ssh.ClientConfig{
		User: *cc.Username,
		// #nosec G106 -- Allow config this? Hacker comunity feedback needed.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         20 * time.Second,
	}

	if cc.PrivateKeyPath != nil {
		key, err := os.ReadFile(*cc.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		// Create the Signer for this private key.
		var signer ssh.Signer
		if cc.PrivateKeyPassphrase != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(*cc.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else if cc.PrivateKey != nil {
		// Create the Signer for this private key.
		var signer ssh.Signer
		if cc.PrivateKeyPassphrase != nil {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(*cc.PrivateKey), []byte(*cc.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(*cc.PrivateKey))
		}
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else if cc.Password != nil {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*cc.Password),
		}
	} else {
		// Use ssh agent for auth
		sshAgent, err := GetSSHAgentClient()
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(sshAgent.Signers),
		}
	}
	return sshConfig, nil
}

func GetSSHAgentClient() (agent.ExtendedAgent, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	agentServer, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	agentClient := agent.NewClient(agentServer)
	return agentClient, nil
}

type SSHClientEventType int

const (
	SSHClientEventDialing = iota
	SSHClientEventConnected
	SSHClientEventMasterClosed
	SSHClientEventClosed
	SSHClientEventError
)

type SSHClientEvent struct {
	Type      SSHClientEventType
	Error     error
	SSHClient *sshClient
}

var NewSSHClient = func() *sshClient {
	c := &sshClient{}
	c.Events = make(chan *SSHClientEvent, 10)
	c.Env = make(map[string]string)
	return c
}

// sshClient obj
type sshClient struct {
	DialAddr string
	// conn
	clientConn *ssh.Client
	//
	Stdout io.Writer
	Stderr io.Writer
	// store of those clients jumping
	// from self client
	subClients []*sshClient
	// will be shared between clients
	Events chan *SSHClientEvent
	Env    map[string]string
	//
	ipcs *ipc.IPC
}

// Connect func
func (s *sshClient) Connect(addr string, config *ssh.ClientConfig) error {
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	s.clientConn = sshClient
	return nil
}

func (s *sshClient) GetConn() *ssh.Client {
	return s.clientConn
}

func (s *sshClient) Dial(addr string, config *ssh.ClientConfig) (*sshClient, error) {
	s.DialAddr = addr
	s.Events <- &SSHClientEvent{Type: SSHClientEventDialing, SSHClient: s}
	if s.clientConn == nil {
		// first connection, call Connect and return itself
		err := s.Connect(addr, config)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("cannot connect to %v", addr), err)
		}
		return s, nil
	}

	// start tcp connection from previos s.clientConn
	conn, err := s.clientConn.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot connect to %v", addr), err)
	}

	// start handshake and ssh proto stuff
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	cc := ssh.NewClient(c, chans, reqs)
	sshclient := &sshClient{
		clientConn: cc,
		Env:        s.Env,
		Stdout:     s.Stdout,
		Stderr:     s.Stderr,
		Events:     s.Events,
	}
	s.subClients = append(s.subClients, sshclient)
	return sshclient, nil
}

func (s *sshClient) Listen(network string, address string) (net.Listener, error) {
	return s.clientConn.Listen(network, address)
}

func (s *sshClient) StartIPC() (*ipc.IPCConsumer, error) {
	lid := fmt.Sprintf("%d", rand.Int()) // #nosec G404 -- Weak random is OK here
	fullpath := filepath.Join("/tmp", "ipc_"+lid+".sock")
	l, err := s.Listen("unix", fullpath)
	if err != nil {
		err0 := fmt.Errorf("cannot listen in remote file %v", fullpath)
		return nil, errors.Join(err0, err)
	}
	ipcs, err := ipc.NewListenerIPCServer(l, lid)
	if err != nil {
		return nil, err
	}

	ipcc := &ipc.IPCConsumer{
		ID:     fmt.Sprintf("%v-c", lid),
		Stream: make(chan *ipc.PipeData),
	}

	ipcs.AppendConsumer(ipcc)
	if s.Env == nil {
		s.Env = make(map[string]string)
	}

	s.Env["NEBULANT_IPCSID"] = ipcs.GetUUID()
	s.Env["NEBULANT_IPCCID"] = ipcc.ID
	s.Env["NEBULANT_CLI_PATH"] = "nebulant_inline_helper"

	go func() {
		err := ipcs.Accept()
		if err != nil {
			ipcs.Errors <- err
		}
	}()

	go func() {
	L:
		for {
			select {
			case err := <-ipcs.Errors:
				s.Events <- &SSHClientEvent{Type: SSHClientEventError, SSHClient: s, Error: err}
			default:
				if len(ipcs.Errors) > 0 {
					continue
				}
				if ipcs.IsClosed() {
					break L
				}
				time.Sleep(200000 * time.Microsecond)
			}
		}
	}()

	s.ipcs = ipcs
	return ipcc, nil
}

func (s *sshClient) DialWithProxies(ccp *ClientConfigParameters) (*sshClient, error) {
	var connections []*ClientConfigParameters
	if len(ccp.Proxies) > 0 {
		connections = append(connections, ccp.Proxies...)
	}

	// Last server to connect. If there is no proxies, this
	// is the last and the unique server to connect.
	connections = append(connections, &ClientConfigParameters{
		Target:               ccp.Target,
		Port:                 ccp.Port,
		Username:             ccp.Username,
		PrivateKey:           ccp.PrivateKey,
		PrivateKeyPath:       ccp.PrivateKeyPath,
		PrivateKeyPassphrase: ccp.PrivateKeyPassphrase,
		Password:             ccp.Password,
	})

	sshClient := s
	for _, ccp := range connections {
		// ctx.Logger.LogDebug("Connecting to addr " + addr + " ...")
		sshClientConfig, err := GetSSHClientConfig(ccp)
		if err != nil {
			return nil, err
		}
		addr := fmt.Sprintf("%v:22", *ccp.Target)
		if ccp.Port != 0 {
			addr = fmt.Sprintf("%v:%d", *ccp.Target, ccp.Port)
		}
		sshClient, err = sshClient.Dial(addr, sshClientConfig)
		if err != nil {
			return nil, err
		}
		// defer sshClient.Disconnect()
	}
	return sshClient, nil
}

// Disconnect func
func (s *sshClient) Disconnect() error {
	var errs []error
	// close IPCS listener
	if s.ipcs != nil {
		// this should rm sock file
		// https://cs.opensource.google/go/x/crypto/+/refs/tags/v0.12.0:ssh/streamlocal.go;l=99
		err := s.ipcs.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	// close sublcients
	for _, sc := range s.subClients {
		err := sc.Disconnect()
		if err != nil {
			errs = append(errs, err)
		}
	}
	s.Events <- &SSHClientEvent{Type: SSHClientEventClosed, SSHClient: s}
	// close itself
	if s.clientConn != nil {
		err := s.clientConn.Close()
		errs = append(errs, err)
	}
	// from doc: "Join returns nil if every value in errs is nil."
	return errors.Join(errs...)
}

// RunCmd func
func (s *sshClient) RunCmd(cmd string) error { // stdout, stderr, error
	session, sesserr := s.clientConn.NewSession()
	if sesserr != nil {
		return sesserr
	}
	defer session.Close()

	session.Stdout = s.Stdout
	session.Stderr = s.Stderr
	inlineEnv := ""
	for n, v := range s.Env {
		err := session.Setenv(n, v)
		if err != nil {
			// AcceptEnv should be enabled to accept envs from session.Setenv
			// https://manpages.debian.org/unstable/openssh-server/sshd_config.5.en.html#AcceptEnv
			// but this is commonly not allowed because:
			// https://serverfault.com/questions/427522/why-is-acceptenv-considered-insecure
			// so here the setenv err gets gracefully ignored and env var gets inserted inline
			inlineEnv = inlineEnv + "export " + n + "=$( cat <<EOF\n" + v + "\nEOF\n) && "
		}
	}

	// The returned error is nil if the command runs, has no problems copying
	// stdin, stdout, and stderr, and exits with a zero exit status.
	// If the remote server does not send an exit status, an error of
	// type *ExitMissingError is returned. If the command completes
	// unsuccessfully or is interrupted by a signal, the error is of type
	// *ExitError. Other error types may be returned for I/O problems.
	// cast.LogInfo( "ssh> "+cmd)
	cmd = inlineEnv + injecfuncs + " && " + cmd
	runerr := session.Run(cmd)
	// This freezes the execution of the command as we want. Do not use:
	// if err := session.Wait(); err != nil {
	// 	return err
	// }
	return runerr
}

// RunScriptFromLocalPath func. Run local script file sending it to the remote machine
func (s *sshClient) RunScriptFromLocalPath(localPath string) error { // stdout, stderr, error
	var stdin bytes.Buffer
	file, err := os.Open(localPath) // #nosec G304 -- This is indeed an inclusion, but the user is fully responsible for this.
	if err != nil {
		return err
	}
	defer file.Close()

	session, sesserr := s.clientConn.NewSession()
	if sesserr != nil {
		return sesserr
	}
	defer session.Close()
	session.Stdin = &stdin
	session.Stdout = s.Stdout
	session.Stderr = s.Stderr
	inlineEnv := ""
	for n, v := range s.Env {
		err := session.Setenv(n, v)
		if err != nil {
			// AcceptEnv should be enabled to accept envs from session.Setenv
			// https://manpages.debian.org/unstable/openssh-server/sshd_config.5.en.html#AcceptEnv
			// but this is commonly not allowed because:
			// https://serverfault.com/questions/427522/why-is-acceptenv-considered-insecure
			// so here the setenv err gets gracefully ignored and env var gets inserted inline
			inlineEnv = inlineEnv + n + "=$( cat <<EOF\n" + v + "\nEOF\n)\n"
		}
	}
	inlineEnv = inlineEnv + "\n" + injecfuncs + "\n"
	envr := strings.NewReader(inlineEnv)
	r := io.MultiReader(envr, file)

	_, err = io.Copy(&stdin, r)
	if err != nil {
		return err
	}

	// The returned error is nil if the command runs, has no problems copying
	// stdin, stdout, and stderr, and exits with a zero exit status.
	// If the remote server does not send an exit status, an error of
	// type *ExitMissingError is returned. If the command completes
	// unsuccessfully or is interrupted by a signal, the error is of type
	// *ExitError. Other error types may be returned for I/O problems.
	// cast.LogInfo( "ssh> " + cmd)
	// runerr := session.Run(cmd)
	if err := session.Shell(); err != nil {
		return err
	}
	if err := session.Wait(); err != nil {
		return err
	}
	return nil
}

// RunScriptFromText func
func (s *sshClient) RunScriptFromText(txt *string) error {
	var stdin bytes.Buffer
	scriptTxt := *txt

	session, sesserr := s.clientConn.NewSession()
	if sesserr != nil {
		return sesserr
	}
	defer session.Close()
	session.Stdin = &stdin
	session.Stdout = s.Stdout
	session.Stderr = s.Stderr
	inlineEnv := ""
	for n, v := range s.Env {
		err := session.Setenv(n, v)
		if err != nil {
			// AcceptEnv should be enabled to accept envs from session.Setenv
			// https://manpages.debian.org/unstable/openssh-server/sshd_config.5.en.html#AcceptEnv
			// but this is commonly not allowed because:
			// https://serverfault.com/questions/427522/why-is-acceptenv-considered-insecure
			// so here the setenv err gets gracefully ignored and env var gets inserted inline
			inlineEnv = inlineEnv + n + "=$( cat <<EOF\n" + v + "\nEOF\n)\n"
		}
	}
	scriptTxt = inlineEnv + "\n" + injecfuncs + "\n" + scriptTxt
	r := strings.NewReader(scriptTxt)
	_, err := io.Copy(&stdin, r)
	if err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}
	if err := session.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *sshClient) NewSCPClientFromExistingSSH() (*scp.Client, error) {
	if s.clientConn == nil {
		return nil, fmt.Errorf("cannot get scp client: ssh not connected")
	}
	return scp.NewClientFromExistingSSH(s.clientConn, &scp.ClientOption{})
}
