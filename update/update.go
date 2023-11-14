// Nebulant
// Copyright (C) 2023 Develatio Technologies S.L.

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

package update

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/downloader"
	"github.com/develatio/nebulant-cli/util"
)

type Version struct {
	Version  string  `json:"version" validate:"required"`
	CodeName string  `json:"codename" validate:"required"`
	Date     string  `json:"date" validate:"required"`
	URL      string  `json:"url" validate:"required"`
	CheckSum string  `json:"checksum" validate:"required"`
	MSG      *string `json:"msg"`
}

type UpdateDescriptor struct {
	Versions map[string]*Version `json:"versions" validate:"required"`
}

type AlreadyUpToDateError struct {
	msg string
}

func (a *AlreadyUpToDateError) Error() string {
	return a.msg
}

type UpdateOutput struct {
	NewVersion *Version
}

// UpdateCLI func updates nebulant binary to latest version.
// version arg is the 1 of 1.2.4
func UpdateCLI(version string, force bool) (*UpdateOutput, error) {
	dir, err := os.MkdirTemp("", "nebulantupdate")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// download update descriptor
	upfilepath := filepath.Join(dir, "version.json")
	err = downloader.DownloadFileWithProgressBar(config.UpdateDescriptorURL, upfilepath, "Checking for update...")
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("cannot check for %s", config.UpdateDescriptorURL))
	}

	// unmarshall update descriptor file
	upfile, err := os.Open(upfilepath) // #nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		return nil, err
	}
	updateDescriptor := &UpdateDescriptor{}
	byteValue, _ := io.ReadAll(upfile)
	err = util.UnmarshalValidJSON(byteValue, updateDescriptor)
	if err != nil {
		return nil, err
	}

	if updateDescriptor.Versions == nil {
		return nil, fmt.Errorf("version %v not found", version)
	}

	if _, exists := updateDescriptor.Versions[version]; !exists {
		return nil, fmt.Errorf("version %v not found", version)
	}

	// get version info and decode urls
	newVersion := updateDescriptor.Versions[version]
	downurl := newVersion.URL
	downurl = strings.Replace(downurl, "{OS}", runtime.GOOS, -1)
	downurl = strings.Replace(downurl, "{ARCH}", runtime.GOARCH, -1)
	downurl = strings.Replace(downurl, "{EXE}", os.Getenv("GOEXE"), -1)
	downurlhash := strings.Replace(updateDescriptor.Versions[version].CheckSum, "{URL}", downurl, -1)

	// down checksum file
	newBinDwldHashPath := filepath.Join(dir, "nebulant.checksum")
	err = downloader.DownloadFileWithProgressBar(downurlhash, newBinDwldHashPath, fmt.Sprintf("Downloading %v...", newVersion.Version))
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("cannot download %s", downurlhash))
	}

	// get checksum file content
	suma, err := util.ReadChecksumFile(newBinDwldHashPath)
	if err != nil {
		return nil, err
	}

	// get current bin
	dstFilePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// sha1sum of current bin
	sumold, err := util.Sha1SumOfFile(dstFilePath)
	if err != nil {
		return nil, err
	}

	// compare server checksum with current local checksum
	if bytes.Equal(suma, sumold) {
		// server bin and local bin are the same
		if !force {
			return nil, &AlreadyUpToDateError{msg: "already up to date"}
		}
	}

	// down new bin
	newBinDwldPath := filepath.Join(dir, "nebulant")
	err = downloader.DownloadFileWithProgressBar(downurl, newBinDwldPath, fmt.Sprintf("Downloading %v...", newVersion.Version))
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("cannot download %s", downurl))
	}

	// sha1sum of new downloaded bin
	sumb, err := util.Sha1SumOfFile(newBinDwldPath)
	if err != nil {
		return nil, err
	}

	// test downloaded bin checksum
	if !bytes.Equal(suma, sumb) {
		return nil, fmt.Errorf("cannot verify downloaded binary")
	}

	// eval posible symlink
	dstFilePath, err = filepath.EvalSymlinks(dstFilePath)
	if err != nil {
		return nil, err
	}

	// move current binary to .old file
	oldFilePath := fmt.Sprintf("%v.old", dstFilePath)
	err = os.Rename(dstFilePath, oldFilePath)
	if err != nil {
		return nil, err
	}

	err = os.Rename(newBinDwldPath, dstFilePath)
	if err != nil {
		err2 := os.Rename(oldFilePath, dstFilePath)
		if err2 != nil {
			return nil, errors.Join(err, err2)
		}
		return nil, err
	}

	err = os.Chmod(dstFilePath, 0755) // #nosec G302 -- Here +x is needed
	if err != nil {
		err2 := os.Rename(oldFilePath, dstFilePath)
		if err2 != nil {
			return nil, errors.Join(err, err2)
		}
		return nil, err
	}

	err = os.Remove(oldFilePath)
	if err != nil {
		return nil, err
	}

	return &UpdateOutput{NewVersion: newVersion}, nil
}
