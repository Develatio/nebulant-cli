//go:build !js

// MIT License
//
// Copyright (C) 2024  Develatio Technologies S.L.

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
				Title("Type the path. Optionally you can use file://...").
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
	bpUrl, err := blueprint.ParseURL(path)
	if err != nil {
		return err
	}
	irb, err := blueprint.NewIRBFromAny(bpUrl, &blueprint.IRBGenConfig{})
	if err != nil {
		return err
	}
	err = executive.InitDirector(false, false)
	if err != nil {
		return err
	}
	executive.MDirector.HandleIRB <- &executive.HandleIRBConfig{IRB: irb}
	executive.MDirector.Wait()
	executive.RemoveDirector()
	return nil
}
