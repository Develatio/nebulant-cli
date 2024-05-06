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

	"github.com/charmbracelet/huh"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/executive"
	"github.com/develatio/nebulant-cli/subsystem"
	"github.com/develatio/nebulant-cli/term"
)

func PathV2(nblc *subsystem.NBLcommand) error {
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

	var path string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Type the path of the file or nebulant://...").
				Value(&path).
				// Validating fields is easy. The form will mark erroneous fields
				// and display error messages accordingly.
				Validate(validate),
		)).Run()
	if err != nil {
		return err
	}

	if len(path) <= 0 {
		term.PrintInfo("No path provided.\n")
		return nil
	}
	term.PrintInfo("Processing blueprint...\n")
	irb, err := blueprint.NewIRBFromAny(path, &blueprint.IRBGenConfig{})
	if err != nil {
		return err
	}
	err = executive.InitDirector(false, false)
	if err != nil {
		return err
	}
	executive.MDirector.HandleIRB <- &executive.HandleIRBConfig{IRB: irb}
	executive.MDirector.Wait()
	executive.MDirector.Clean()
	return nil
}
