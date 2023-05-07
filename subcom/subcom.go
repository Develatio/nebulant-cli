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
	"fmt"
	"strings"
)

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
