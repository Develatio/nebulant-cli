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
	FilePath         string
	UrlPath          string
}

func ParsePath(path string) (*BlueprintURL, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "":
		return ParseURL(fmt.Sprintf("file://%s", path))
	case "file":
		return ParseURL(path)
	default:
		return nil, fmt.Errorf("bad scheme for file: %s. Use file://... scheme or rm scheme from path", u.Scheme)
	}
}

func ParseURL(path string) (*BlueprintURL, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	out := &BlueprintURL{
		UrlPath:  u.Path,
		FilePath: fmt.Sprintf("%s%s", u.Host, u.Path),
	}
	switch u.Scheme {
	case "nebulant":
		out.Scheme = u.Scheme
	case "file":
		fi, err := os.Stat(out.FilePath)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			return nil, fmt.Errorf("directory not allowed")
		}
		out.Scheme = u.Scheme
		return out, nil
	case "":
		out.Scheme = "nebulant"
	default:
		return nil, fmt.Errorf("unknown path scheme for %s", path)
	}

	r := regexp.MustCompile(`(?:([-a-zA-Z0-9_]+)\/+)?([-a-zA-Z0-9_]+)\/+([-a-zA-Z0-9_]+)(?::([-a-zA-Z0-9_.]+))?`)
	matches := r.FindAllStringSubmatch(path, -1)
	if matches == nil {
		return nil, fmt.Errorf("bad remote bp path: %s", path)
	}

	out.OrganizationSlug = matches[0][1]
	out.CollectionSlug = matches[0][2]
	out.BlueprintSlug = matches[0][3]
	out.Version = matches[0][4]

	if out.CollectionSlug == "" {
		return nil, fmt.Errorf("bad remote collection bp path: %s", path)
	}

	return out, nil
}
