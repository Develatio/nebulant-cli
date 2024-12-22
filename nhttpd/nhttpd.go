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
package nhttpd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
)

var server = &Httpd{
	urls: make(map[*regexp.Regexp]ViewFunc),
	validOrigins: map[string]bool{
		// The localhost server should always be allowed since this is how Safari
		// will connect (using the TransparentProxy)
		"http://" + net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT): true,
	},
}

func GetServer() *Httpd {
	addr := net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT)
	server.addr = addr
	return server
}

type ViewFunc func(w http.ResponseWriter, r *http.Request, matches [][]string)

// Httpd struct
type Httpd struct {
	validOrigins map[string]bool
	on           bool
	srv          *http.Server
	scheme       string
	errors       []error
	consumers    []chan error
	urls         map[*regexp.Regexp]ViewFunc
	addr         string
	certPath     *string
	keyPath      *string
}

func (h *Httpd) SetSecure(cert string, key string) {
	h.certPath = &cert
	h.keyPath = &key
}

func (h *Httpd) SetAddr(addr string) {
	h.addr = addr
}

func (h *Httpd) AddView(path string, view ViewFunc) {
	h.urls[regexp.MustCompile(path)] = view
}

func (h *Httpd) AddOrigin(origin string) {
	h.validOrigins[origin] = true
}

func (h *Httpd) ValidateOrigin(r *http.Request) bool {
	surl := r.Header.Get("Origin")
	// browser send no Origin header on directly access
	if surl == "" {
		return true
	}

	// Fail on missing Origin header or bad url
	url, err := url.Parse(surl)
	if err != nil {
		return false
	}

	burl, err := url.MarshalBinary()
	if err != nil {
		return false
	}
	surl = string(burl)
	cast.LogDebug("Validating origin: "+surl, nil)

	if _, exists := h.validOrigins["*"]; exists {
		return true
	}

	if _, exists := h.validOrigins[surl]; exists {
		return true
	}
	return false
}

func (h *Httpd) httpMiddleware(w http.ResponseWriter, r *http.Request) error {
	if !h.ValidateOrigin(r) {
		w.WriteHeader(http.StatusBadRequest)
		// The Nebulant Bridge keeps retring the
		// connection. This is a problem when a
		// debug server is up without server
		// mode on because this err is displaying
		// every
		return fmt.Errorf("bad Origin")
	}
	surl := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", surl)
	if r.Method == "OPTIONS" {
		pnacors := r.Header.Get("Access-Control-Request-Private-Network")
		if pnacors == strings.Trim(strings.ToLower("true"), " ") {
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
		}
	}

	return nil
}

func (h *Httpd) route(w http.ResponseWriter, r *http.Request) {
	err := h.httpMiddleware(w, r)
	if err != nil {
		cast.LogErr(err.Error(), nil)
		return
	}

	var vfn ViewFunc
	var vrgx *regexp.Regexp
	for rgx, fn := range h.urls {
		if rgx.MatchString(r.URL.Path) {
			vfn = fn
			vrgx = rgx
			break
		}
	}
	if vfn != nil {
		matches := vrgx.FindAllStringSubmatch(r.URL.Path, -1)
		vfn(w, r, matches)
	} else {
		http.Error(w, "404 Not found", http.StatusNotFound)
	}
}

func (h *Httpd) ServeIfNot() chan error {
	consumer := make(chan error)
	h.consumers = append(h.consumers, consumer)
	if h.on {
		return consumer
	}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", h.route)

	// prevent slowloris DDoS attack (G114)
	h.srv = &http.Server{
		Addr:              h.addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           serveMux,
	}
	cast.LogInfo(fmt.Sprintf("Listening on %s", h.addr), nil)
	h.on = true
	go func() {
		var err error
		defer h.Shutdown()

		// TLS Server
		if h.certPath != nil && *h.certPath != "" {
			if h.keyPath == nil || *h.keyPath == "" {
				h.errors = append(h.errors, fmt.Errorf("TLS server err: empty key path"))
				return
			}
			h.scheme = "https"
			if err = h.srv.ListenAndServeTLS(*h.certPath, *h.keyPath); err != nil {
				err = errors.Join(fmt.Errorf("TLS server err. cert path: %s", *h.certPath), err)
				h.errors = append(h.errors, err)
			}
			return
		}

		// Insecure Server
		h.scheme = "http"
		err = h.srv.ListenAndServe()
		if err = errors.Join(fmt.Errorf("server err"), err); err != nil {
			h.errors = append(h.errors, err)
		}
	}()
	return consumer
}

func (h *Httpd) GetAddr() string {
	if h.srv != nil {
		return h.srv.Addr
	}
	return h.addr
}

func (h *Httpd) GetScheme() string {
	if h.srv != nil {
		return h.srv.Addr
	}
	return h.addr
}

func (h *Httpd) Shutdown() error {
	serr := h.srv.Shutdown(context.Background())
	h.errors = append(h.errors, serr)
	err := errors.Join(h.errors...)
	for _, consumer := range h.consumers {
		consumer <- err
	}
	h.on = false
	return err
}
