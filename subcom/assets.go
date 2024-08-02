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
	"net/url"
	"strconv"

	"github.com/develatio/nebulant-cli/assets"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/subsystem"
)

func parseAssetsFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("assets", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant assets [command] [options]\n")
		fmt.Fprintf(fs.Output(), "\nCommands:\n")
		fmt.Fprintf(fs.Output(), "  upgrade\t\tLookup for new assets update\n")
		fmt.Fprintf(fs.Output(), "  build\t\t\tLocaly build asset index\n")
		fmt.Fprintf(fs.Output(), "  search\t\tSearch for data into assets\n")
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func parseAssetsUpgradeFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	config.ForceUpgradeAssetsFlag = fs.Bool("u", false, "Force upgrade assets. Download prebuild-index.")
	config.ForceUpgradeAssetsNoDownloadFlag = fs.Bool("uu", false, "Force upgrade assets. Skip download prebuild-index and build locally.")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant assets upgrade [options]\n")
		fmt.Fprintf(fs.Output(), "\nLookup for new assets upgrade\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		subsystem.PrintDefaults(fs)
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[2:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func parseAssetsBuildFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	fs.String("f", "", "Input file. Ej. -f ./file.json")
	fs.String("a", "", "Asset ID. Ej. -a aws_images")
	fs.String("d", "", "Output dir to save generated files")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant assets build [options]\n")
		fmt.Fprintf(fs.Output(), "\nLocally build asset index\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		subsystem.PrintDefaults(fs)
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[2:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func parseAssetsSearchFs(cmdline *flag.FlagSet) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(cmdline.Output())
	fs.String("a", "", "\tSearch into the `asset` ID. Ej. aws_images")
	fs.String("t", "", "\tSearch term")
	fs.String("f", "", "\tFilter terms. Ej. -f region=us-east-1")
	fs.Int("l", 0, "\tLimit")
	fs.Int("o", 0, "\tOffset")
	fs.String("s", "", "\tSort. Ej. $.Name")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "\nUsage: nebulant assets search [options]\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		subsystem.PrintDefaults(fs)
		fmt.Fprintf(fs.Output(), "\n\n")
	}
	err := fs.Parse(cmdline.Args()[2:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func AssetsCmd(nblc *subsystem.NBLcommand) (int, error) {
	cmdline := nblc.CommandLine()
	fs, err := parseAssetsFs(cmdline)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, nil
		}
		return 1, err
	}

	// subsubcmd := fs.Arg(0)
	subsubcmd := cmdline.Arg(1)
	switch subsubcmd {
	case "upgrade":
		_, err = parseAssetsUpgradeFs(cmdline)
		if err != nil {
			return 1, err
		}
		err := assets.UpgradeAssets(*config.ForceUpgradeAssetsFlag, *config.ForceUpgradeAssetsNoDownloadFlag)
		if err != nil {
			return 1, err
		}
		return 0, nil
	case "build":
		fs, err := parseAssetsBuildFs(cmdline)
		if err != nil {
			return 1, err
		}
		ff := fs.Lookup("f")
		if ff == nil || ff.Value.String() == "" {
			fs.PrintDefaults()
			return 1, fmt.Errorf("please set input file")
		}
		aa := fs.Lookup("a")
		if aa == nil || aa.Value.String() == "" {
			fs.PrintDefaults()
			return 1, fmt.Errorf("please set asset_id")
		}
		dd := fs.Lookup("d")
		if dd == nil || dd.Value.String() == "" {
			fs.PrintDefaults()
			return 1, fmt.Errorf("please set outputdir/")
		}
		// TODO: rm old concatenated syntax and pass as individual arguments
		err = assets.GenerateIndexFromFile(ff.Value.String() + ":" + aa.Value.String() + ":" + dd.Value.String())
		if err != nil {
			return 1, err
		}
		return 0, nil
	case "search":
		fs, err := parseAssetsSearchFs(cmdline)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0, nil
			}
			return 1, err
		}
		assetid := fs.Lookup("a").Value.String()
		assetdef, ok := assets.AssetsDefinition[assetid]
		if !ok {
			fs.PrintDefaults()
			return 1, fmt.Errorf("unknown asset id")
		}
		srq := &assets.SearchRequest{
			SearchTerm: fs.Lookup("t").Value.String(),
			Sort:       fs.Lookup("s").Value.String(),
		}
		srq.Limit, err = strconv.Atoi(fs.Lookup("l").Value.String())
		if err != nil {
			return 1, fmt.Errorf("invalid search pagination limit")
		}
		srq.Offset, err = strconv.Atoi(fs.Lookup("o").Value.String())
		if err != nil {
			return 1, fmt.Errorf("invalid search pagination offset")
		}

		ff := fs.Lookup("f").Value.String()
		if ff != "" {
			ft, err := url.ParseQuery(ff)
			if err != nil {
				return 1, errors.Join(err, fmt.Errorf("bad filter terms"))
			}
			srq.FilterTerms = ft
		}

		searchres, err := assets.Search(srq, assetdef)
		if err != nil {
			return 1, err
		}
		cast.LogInfo("Found "+fmt.Sprintf("%v", searchres.Count)+" items", nil)

		for e, item := range searchres.Results {
			b := assetdef.MarshallIndentItem(item)
			cast.LogInfo(fmt.Sprintf("Result %v / %v -> %v", e, searchres.Count, b), nil)
			if e >= 10 {
				cast.LogInfo("[...]", nil)
				break
			}
		}
	default:
		fs.Usage()
		return 1, fmt.Errorf("please provide some subcommand to assets")
	}
	return 0, nil
}
