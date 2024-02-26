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

var server = &Httpd{urls: make(map[*regexp.Regexp]ViewFunc), validOrigins: make(map[string]bool)}

func GetServer() *Httpd {
	return server
}

type ViewFunc func(w http.ResponseWriter, r *http.Request, matches [][]string)

// Httpd struct
type Httpd struct {
	validOrigins map[string]bool
	on           bool
	srv          *http.Server
	errors       []error
	consumers    []chan error
	urls         map[*regexp.Regexp]ViewFunc
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

	// allow http scheme so safari can connect <https> builder -> <http> localhost
	if url.Scheme == "http" {
		url.Scheme = "https"
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

	pip := net.ParseIP(config.SERVER_ADDR)
	if !pip.IsPrivate() && !pip.IsLoopback() {
		cast.LogWarn("You are using a public ip. Please note that this could result in a security hole!", nil)
	}

	addr := net.JoinHostPort(config.SERVER_ADDR, config.SERVER_PORT)

	// prevent slowloris DDoS attack (G114)
	h.srv = &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           serveMux,
	}
	cast.LogInfo(fmt.Sprintf("Listening on %s", addr), nil)
	h.on = true
	go func() {
		// WIP: explorar si podemos generar los certs al vuelo, que obviamente
		// no serán válidos a ojos del navegador, pero nos permite establecer
		// conexiones seguras en caso de conectarnos con el debugger en modo
		// remoto
		// https:
		// err := http.ListenAndServeTLS(*addr, "localhost.crt", "localhost.key", nil)
		err := h.srv.ListenAndServe()
		if err != nil {
			h.errors = append(h.errors, err)
		}
		h.Shutdown()
	}()
	return consumer
}

func (h *Httpd) GetAddr() string {
	if h.srv != nil {
		return h.srv.Addr
	}
	return ""
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
