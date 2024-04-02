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

package subsystem

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/providers/aws"
	"github.com/develatio/nebulant-cli/providers/azure"
	"github.com/develatio/nebulant-cli/providers/cloudflare"
	"github.com/develatio/nebulant-cli/providers/generic"
	"github.com/develatio/nebulant-cli/providers/hetzner"
	"github.com/develatio/nebulant-cli/term"
)

type SCType int

const (
	SecMain SCType = iota
	SecRuntime
	SecHidden
)

type NBLcommand struct {
	UpgradeTerm   bool
	WelcomeMsg    bool
	InitProviders bool
	Help          string
	Sec           SCType
	//
	Call    func(*NBLcommand) (int, error)
	cmdline *flag.FlagSet
	Stdin   io.ReadCloser
	Stdout  io.WriteCloser
}

func (n *NBLcommand) Run(cmdline *flag.FlagSet) (int, error) {
	n.cmdline = cmdline
	if n.Stdin == nil {
		n.Stdin = term.Stdin
	}
	if n.Stdout == nil {
		n.Stdout = term.Stdout
	}
	return n.Call(n)
}

func (n *NBLcommand) CommandLine() *flag.FlagSet {
	return n.cmdline
}

var NBLCommands map[string]*NBLcommand

func PrintDefaults(f *flag.FlagSet) {
	f.VisitAll(func(ff *flag.Flag) {
		var b strings.Builder
		fmt.Fprintf(&b, "  -%s ", ff.Name)
		name, usage := flag.UnquoteUsage(ff)
		if len(name) > 0 {
			b.WriteString(name)
		}
		l := 25 - (len(b.String()) + len(name))
		for i := 0; i < l; i++ {
			b.WriteString(" ")
		}
		b.WriteString(usage)
		if ff.DefValue != "" && ff.DefValue != "false" {
			fmt.Fprintf(&b, " (default %v)", ff.DefValue)
		}
		fmt.Fprint(f.Output(), b.String(), "\n")
	})
}

func PrepareCmd(cmd *NBLcommand) error {
	var err error
	if cmd.UpgradeTerm {
		// Init Term
		err = term.UpgradeTerm()
		if err != nil {
			// cast.SBus.Close().Wait()
			// fmt.Println("cannot init term :(")
			// fmt.Println(err.Error())
			// os.Exit(1)
			return errors.Join(fmt.Errorf("cannot init term :("), err)
		}
		term.ConfigColors()
	}

	if config.DEBUG {
		cast.LogDebug("Debug mode activated. Testing message levels...", nil)
		cast.LogInfo("Info message", nil)
		cast.LogWarn("Warning message", nil)
		cast.LogErr("Error message", nil)
		cast.LogCritical("Critical message", nil)
	}

	if cmd.WelcomeMsg {
		_, err = term.Println(term.Magenta+"Nebulant CLI"+term.Reset, "- A cloud builder by", term.Blue+"develat.io"+term.Reset)
		if err != nil {
			fmt.Println("Nebulant CLI - A cloud builder by develat.io")
		}
		_, err = term.Println(term.Gray+" Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler, term.Reset)
		if err != nil {
			fmt.Println("Version: v"+config.Version, "-", config.VersionDate, runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		}
		term.PrintInfo(" Welcome :)\n")
	}

	// Init Providers
	if cmd.InitProviders {
		cast.SBus.RegisterProviderInitFunc("aws", aws.New)
		cast.SBus.RegisterProviderInitFunc("azure", azure.New)
		cast.SBus.RegisterProviderInitFunc("generic", generic.New)
		cast.SBus.RegisterProviderInitFunc("executionControl", generic.New)
		cast.SBus.RegisterProviderInitFunc("execution-control", generic.New)
		cast.SBus.RegisterProviderInitFunc("hetznerCloud", hetzner.New)
		cast.SBus.RegisterProviderInitFunc("cloudflare", cloudflare.New)
		blueprint.ActionValidators["providerValidator"] = func(action *blueprint.Action) error {
			if _, err := cast.SBus.GetProviderInitFunc(action.Provider); err != nil {
				return err
			}
			return nil
		}
		blueprint.ActionValidators["awsValidator"] = aws.ActionValidator
		blueprint.ActionValidators["azureValidator"] = azure.ActionValidator
		blueprint.ActionValidators["genericsValidator"] = generic.ActionValidator
		blueprint.ActionValidators["hetznerValidator"] = hetzner.ActionValidator
		blueprint.ActionValidators["cloudflareValidator"] = cloudflare.ActionValidator
	}

	return nil
}

func ConfArgs(fflag *flag.FlagSet, arguments []string) error {
	// var err error
	// compat flags
	sflag := fflag.Bool("s", false, "Ignored for compatibility")
	//
	config.VersionFlag = fflag.Bool("v", false, "Show version and exit.")
	config.DebugFlag = fflag.Bool("x", false, "Enable debug.")
	config.ParanoicDebugFlag = fflag.Bool("xx", false, "Enable paranoic debug.")
	config.Ipv6Flag = fflag.Bool("6", false, "Force ipv6")
	config.DisableColorFlag = fflag.Bool("c", false, "Disable colors.")
	config.ForceTerm = fflag.Bool("ft", false, "Force terminal. Bypass no-term detection.")
	config.BridgeAddrFlag = fflag.String("b", "", "self-hosted bridge addr:port (ipv4) or [::1]:port (ipv6).")
	config.BridgeSecretFlag = fflag.String("bs", config.BRIDGE_SECRET, "self-hosted bridge auth secret string (overrides env NEBULANT_BRIDGE_SECRET).")

	fflag.Usage = func() {
		var runtimecmds []string
		var orderedcmdtxt []string
		for cmdtxt, cmd := range NBLCommands {
			if cmd.Sec == SecHidden {
				continue
			}
			orderedcmdtxt = append(orderedcmdtxt, cmdtxt)
		}
		sort.Strings(orderedcmdtxt)
		fmt.Fprint(fflag.Output(), "\nUsage: nebulant [flags] [command]\n")
		fmt.Fprint(fflag.Output(), "\nFlags:\n")
		PrintDefaults(fflag)
		fmt.Fprint(fflag.Output(), "\nCommands:\n")
		for _, cmdtxt := range orderedcmdtxt {
			cmd := NBLCommands[cmdtxt]
			if cmd.Sec == SecRuntime {
				runtimecmds = append(runtimecmds, cmd.Help)
				continue
			}
			fmt.Fprint(fflag.Output(), cmd.Help)
		}
		fmt.Fprint(fflag.Output(), "\n\nRuntime commands:\n")
		for _, hh := range runtimecmds {
			fmt.Fprint(fflag.Output(), hh)
		}
		// fmt.Fprint(fflag.Output(), "  readvar\t\t"+term.EmojiSet["Key"]+" Read blueprint variable value during runtime\n")
		fmt.Fprint(fflag.Output(), "\n\nrun nebulant [command] --help to show help for a command\n")
	}

	if err := fflag.Parse(arguments); err != nil {
		fflag.PrintDefaults()
		return err
	}
	if *sflag {
		fmt.Fprint(fflag.Output(), "\n\ndeprecated flag. Use 'serve' command instead: ./nebulant serve\n")
		return fmt.Errorf("deprecated flag err")
	}
	return nil
}

func Run(sc string) (int, error) {
	if cmd, exists := NBLCommands[sc]; exists {
		// prepare cmd
		err := PrepareCmd(cmd)
		if err != nil {
			return 1, err
		}
		// finally run command
		return cmd.Run(flag.CommandLine)
	} else {
		flag.Usage()
		return 1, fmt.Errorf("unknown command")
	}
}
