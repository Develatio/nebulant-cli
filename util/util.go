// MIT License
//
// Copyright (C) 2021  Develatio Technologies S.L.

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

package util

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"

	"github.com/go-playground/validator/v10"
)

// UIStore interface
type UIStore interface {
	GetValueByReference(reference *string) (interface{}, error)
}

// UnmarshalValidJSON func
func UnmarshalValidJSON(data []byte, v interface{}) error {
	jsonErr := json.Unmarshal(data, v)
	if jsonErr != nil {
		return jsonErr
	}

	// json tag validator
	validate := validator.New()
	vErr := validate.Struct(v)
	if vErr != nil {
		return vErr
	}

	// Validate()
	vvv := reflect.TypeOf(v)
	_, validable := vvv.MethodByName("Validate")
	if validable {
		ret := reflect.ValueOf(v).MethodByName("Validate").Call([]reflect.Value{})
		switch ret[0].Interface().(type) {
		case error:
			return ret[0].Interface().(error)
		}
	}

	return nil
}

// DeepCopy func
func DeepCopy(src interface{}, dst interface{}) error {
	enc, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return UnmarshalValidJSON(enc, dst)
}

func OpenUrl(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Run()
}

func Sha1SumOfFile(filepath string) ([]byte, error) {
	// check for sum
	f, err := os.Open(filepath) // #nosec G304 -- Not a file inclusion, just file read
	if err != nil {
		return nil, err
	}
	// ensure close
	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%x", h.Sum(nil))), nil
}

func ReadChecksumFile(filepath string) ([]byte, error) {
	_hhc, err := os.ReadFile(filepath) // #nosec G304 -- Not a file inclusion, just file read
	if err != nil {
		return nil, err
	}
	sum, _, _ := bytes.Cut([]byte(_hhc), []byte(" "))
	return sum, nil
}

type PanicData struct {
	PanicValue interface{}
	PanicTrace []byte
}

// https://go.dev/src/syscall/zerrors_linux_amd64.go
// https://go.dev/src/syscall/zerrors_****.go
func IsNetError(err error) bool {
	if _, yes := err.(*net.OpError); yes {
		return true
	}
	if _, yes := err.(*net.DNSError); yes {
		return true
	}
	if _, yes := err.(net.Error); yes {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	if errors.Is(err, syscall.ECONNABORTED) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	if errors.Is(err, syscall.ENOTCONN) {
		return true
	}
	if errors.Is(err, syscall.ENETDOWN) {
		return true
	}
	if errors.Is(err, syscall.ENETRESET) {
		return true
	}
	if errors.Is(err, syscall.ENETUNREACH) {
		return true
	}
	if errors.Is(err, syscall.EHOSTDOWN) {
		return true
	}
	if errors.Is(err, syscall.EHOSTUNREACH) {
		return true
	}
	if errors.Is(err, syscall.EREMOTE) {
		return true
	}
	return false
}

// ExpandDir supoprts ~/
// by now, the $HOME/dir1/dir2 is not suported
// check https://pkg.go.dev/os#ExpandEnv to add support to $HOME/dir/ expand
func ExpandDir(dir string) (string, error) {
	var err error
	if strings.HasPrefix(dir, "~/") || strings.HasPrefix(dir, "~\\") {
		ud, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = ud + dir[1:]
	}
	if dir == "~" {
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", nil
		}
	}
	return dir, nil
}
