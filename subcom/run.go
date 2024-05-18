// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

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

package subcom

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
)

func parseRunFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	config.ForceFile = fs.Bool("f", false, "Run local file")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant [file://, nebulant://][org/coll/bp][filepath] [--varname=varvalue --varname=varvalue]\n\n")
		fmt.Fprintf(fs.Output(), "Examples:\n")
		fmt.Fprintf(fs.Output(), "\tnebulant develatio/utils/debug\n")
		fmt.Fprintf(fs.Output(), "\tnebulant nebulant://develatio/utils/debug\n")
		fmt.Fprintf(fs.Output(), "\tnebulant -f ./local/file/project.nbp\n")
		fmt.Fprintf(fs.Output(), "\tnebulant file://local/file/project.nbp\n")
		fmt.Fprintf(fs.Output(), "\tnebulant run develatio/utils/debug\n")
		fmt.Fprintf(fs.Output(), "\tnebulant run nebulant://develatio/utils/debug\n")
		fmt.Fprintf(fs.Output(), "\tnebulant run -f ./local/file/project.nbp\n")
		fmt.Fprintf(fs.Output(), "\tnebulant run file://local/file/project.nbp\n")
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func RunCmd(nblc *subsystem.NBLcommand) (int, error) {
	fs, err := parseRunFs(nblc.CommandLine())
	if err != nil {
		return 1, err
	}
	bluePrintFilePath := fs.Arg(0)
	if bluePrintFilePath == "" {
		fs.Usage()
		return 1, fmt.Errorf("please provide addr to the blueprint you want to execute")
	}

	cast.LogInfo("Processing blueprint...", nil)
	irbConf := &blueprint.IRBGenConfig{}
	args := fs.Args()
	if len(args) > 1 {
		irbConf.Args = args[1:]
	}
	var bpUrl *blueprint.BlueprintURL
	if config.ForceFile != nil && *config.ForceFile {
		bpUrl, err = blueprint.ParsePath(bluePrintFilePath)
	} else {
		bpUrl, err = blueprint.ParseURL(bluePrintFilePath)
	}
	if err != nil {
		return 1, err
	}

	irb, err := blueprint.NewIRBFromAny(bpUrl, irbConf)
	if err != nil {
		if bpUrl.Scheme != "file" && bpUrl.UrlPath != "" {
			if fi, err2 := os.Stat(bpUrl.FilePath); err2 == nil && !fi.IsDir() {
				return 1, errors.Join(err, fmt.Errorf("did you want to run file %s?, try adding the -f attribute: `nebulant -f %s`. You can also use file:// scheme: `nebulant file://%s", bpUrl.FilePath, bpUrl.FilePath, bpUrl.FilePath))
			}
			fmt.Println("b")
		}
		return 1, err
	}
	// Director in one run mode
	err = executive.InitDirector(false, false)
	if err != nil {
		return 1, err
	}
	executive.MDirector.HandleIRB <- &executive.HandleIRBConfig{IRB: irb}
	executive.MDirector.Wait()
	return 0, nil
}
