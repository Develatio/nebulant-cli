// Nebulant
// Copyright (C) 2024  Develatio Technologies S.L.

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

package blueprint

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
)

type BlueprintURL struct {
	Scheme           string
	OrganizationSlug string
	CollectionSlug   string
	BlueprintSlug    string
	Version          string
	Path             string
}

func ParseURL(path string) (*BlueprintURL, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	out := &BlueprintURL{}
	switch u.Scheme {
	case "nebulant":
		out.Scheme = u.Scheme
	case "file":
		out.Path = u.Path
		fi, err := os.Stat(out.Path)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			return nil, fmt.Errorf("directory not allowed")
		}
		return out, nil
	case "":
		out.Scheme = "nebulant"
	default:
		return nil, fmt.Errorf("unknown path scheme for %s", path)
	}

	r := regexp.MustCompile(`(?:([-a-zA-Z0-9_]+)\/+)?([-a-zA-Z0-9_]+)\/+([-a-zA-Z0-9_]+)(?::([-a-zA-Z0-9_.]+))?`)
	matches := r.FindAllStringSubmatch(path, -1)
	if matches == nil {
		return nil, fmt.Errorf("bad bp path: %s", path)
	}

	out.OrganizationSlug = matches[0][1]
	out.CollectionSlug = matches[0][2]
	out.BlueprintSlug = matches[0][3]
	out.Version = matches[0][4]

	if out.CollectionSlug == "" {
		return nil, fmt.Errorf("bad bp path: %s", path)
	}

	return out, nil
}
