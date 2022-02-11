//go:build !js

// Nebulant
// Copyright (C) 2022  Develatio Technologies S.L.

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

package interactive

import (
	"fmt"
	"os"

	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/term"
	"github.com/manifoldco/promptui"
)

func Path() error {
	validate := func(input string) error {
		if len(input) <= 0 {
			return nil
		}
		if len(input) >= 12 && input[:11] == "nebulant://" {
			return nil
		}
		fi, err := os.Stat(input)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return fmt.Errorf("directory not allowed")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Path",
		Validate: validate,
	}
	path, err := prompt.Run()
	if err != nil {
		return err
	}
	if len(path) <= 0 {
		term.PrintInfo("No path provided.\n")
		return nil
	}
	term.PrintInfo("Processing blueprint...\n")
	irb, err := blueprint.NewIRBFromAny(path)
	if err != nil {
		return err
	}
	err = executive.InitDirector(false, false)
	if err != nil {
		return err
	}
	executive.MDirector.HandleIRB <- irb
	executive.MDirector.Wait()
	executive.MDirector.Clean()
	return nil
}
