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
	"flag"

	"github.com/develatio/nebulant-cli/update"
)

func parseUpdate() (*flag.FlagSet, error) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	err := fs.Parse(flag.Args()[1:])
	if err != nil {
		return fs, err
	}
	return fs, nil
}

func UpdateCmd() (int, error) {
	_, err := parseUpdate()
	if err != nil {
		return 1, err
	}

	err = update.UpdateCLI("1")
	if err != nil {
		return 1, err
	}

	return 0, nil
}
