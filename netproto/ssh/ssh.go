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
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func GetSSHAgentClient() (agent.ExtendedAgent, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	agentServer, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	agentClient := agent.NewClient(agentServer)
	return agentClient, nil
}

// singleWriter struct, copied from ssh.session
type singleWriter struct {
	b  bytes.Buffer
	mu sync.Mutex
}

// Write func, copied from ssh.session
func (w *singleWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *singleWriter) String() string {
	return w.b.String()
}

// NewSSHClient var
// params grow
var NewSSHClient = func(stdout io.Writer, stderr io.Writer) *sshClient {
	sshc := &sshClient{
		Stdout: stdout,
		Stderr: stderr,
	}
	sshc.Env = make(map[string]string)
	return sshc
}

// sshClient obj
type sshClient struct {
	client *ssh.Client
	Stdout io.Writer
	Stderr io.Writer
	Env    map[string]string
}

// Connect func
func (s *sshClient) Connect(addr string, config *ssh.ClientConfig) error {
	// cast.LogInfo( "Connecting ssh to "+addr+" ...")

	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	s.client = sshClient
	// cast.LogInfo( "Connected?")
	return nil
}

// Disconnect func
func (s *sshClient) Disconnect() error {
	return s.client.Close()
}

// RunCmd func
func (s *sshClient) RunCmd(cmd string) error { // stdout, stderr, error
	session, sesserr := s.client.NewSession()
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
	cmd = inlineEnv + cmd
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
	file, err := os.Open(localPath) //#nosec G304 -- This is indeed an inclusion, but the user is fully responsible for this.
	if err != nil {
		return err
	}
	defer file.Close()

	session, sesserr := s.client.NewSession()
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

	session, sesserr := s.client.NewSession()
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
	scriptTxt = inlineEnv + scriptTxt
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
