package term

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/develatio/nebulant-cli/util"
)

type Pipe struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func NewPipe() (*Pipe, error) {
	return &Pipe{
		Stdout: Stdout,
		Stderr: Stderr,
		Stdin:  Stdin,
	}, nil
}

func DetermineOsShell() (string, error) {
	var shell string
	switch runtime.GOOS {
	case "windows":
		shell = os.Getenv("COMSPEC")
		if len(shell) <= 0 {
			shell = "cmd.exe"
		}
	case "darwin":
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		argv, err := util.CommandLineToArgv("dscl /Search -read \"/Users/" + user.Username + "\" UserShell")
		if err != nil {
			return "", err
		}
		out, err := exec.Command(argv[0], argv[1:]...).Output() // #nosec G204 -- allowed here
		if err != nil {
			return "", err
		}
		shell = string(out)
		shell = strings.Replace(shell, "UserShell: ", "", 1)
		shell = strings.Trim(shell, "\n")
		if len(shell) <= 0 {
			for _, sh := range []string{"zsh", "bash", "ksh"} {
				path, err := exec.LookPath(sh)
				if err != nil {
					continue
				}
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					shell = path
					break
				}
			}
		}
	case "linux", "openbsd", "freebsd":
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		out, err := exec.Command("getent", "passwd", user.Uid).Output() // #nosec G204 -- allowed here
		if err != nil {
			return "", err
		}
		parts := strings.SplitN(string(out), ":", 7)
		if len(parts) < 7 || parts[0] == "" || parts[0][0] == '+' || parts[0][0] == '-' {
			return "", fmt.Errorf("cannot determine OS shell")
		}
		shell = parts[6]
		shell = strings.Trim(shell, "\n")
		if len(shell) <= 0 {
			shell = "/bin/bash"
		}
	}

	if len(shell) <= 0 {
		return "", fmt.Errorf("cannot determine OS shell")
	}

	return shell, nil
}

func OpenLocalShell() error {
	shell, err := DetermineOsShell()
	if err != nil {
		return err
	}
	cmd := exec.Command(shell)
	cmd.Stdin = Stdin
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}
