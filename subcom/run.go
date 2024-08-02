// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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
	config.ForceFileFlag = fs.Bool("f", false, "Run local file")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant run [org/coll/bp] [-f filepath] [--varname=varvalue --varname=varvalue]\n\n")
		fmt.Fprintf(fs.Output(), "Examples:\n")
		fmt.Fprintf(fs.Output(), "\tnebulant run develatio/utils/debug\n")
		fmt.Fprintf(fs.Output(), "\tnebulant run -f ./local/file/project.nbp\n")
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
		if errors.Is(err, flag.ErrHelp) {
			return 0, nil
		}
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
	if config.ForceFileFlag != nil && *config.ForceFileFlag {
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
