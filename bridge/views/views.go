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
package views

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/develatio/nebulant-cli/bridge/assets"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	ws "github.com/develatio/nebulant-cli/netproto/websocket"
	"github.com/develatio/nebulant-cli/nhttpd"
	"github.com/develatio/nebulant-cli/nsterm"
	"github.com/gorilla/websocket"

	_ "embed"
)

var Bridge *Puente = &Puente{pools: make(map[string]*pool), xtermindex: assets.XTERM}

func Serve() error {
	if config.DEBUG {
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				cast.LogInfo(fmt.Sprintf("Alloc: %v MiB\tHeapInuse: %v MiB\tFrees: %v MiB\tSys: %v MiB\tNumGC: %v", m.Alloc/1024/1024, m.HeapInuse/1024/1024, m.Frees/1024/1024, m.Sys/1024/1024, m.NumGC), nil)
				pools := 0
				consumers := 0
				for _, p := range Bridge.pools {
					pools++
					consumers = consumers + len(p.consumerConn)
				}
				cast.LogInfo(fmt.Sprintf("Pools: %v\tConsumers: %v", pools, consumers), nil)
				time.Sleep(10 * time.Second)
			}
		}()
	}

	Bridge.secret = *config.BridgeSecretFlag
	srv := nhttpd.GetServer()
	addr := config.BridgeAddrFlag
	srv.SetAddr(*addr)
	srv.SetSecure(*config.BridgeCertPathFlag, *config.BridgeKeyPathFlag)
	// TODO: config origin from
	// env var, conf file or flag
	srv.AddOrigin(*config.BridgeOriginFlag)

	srv.AddView(`^/new$`, Bridge.newView)
	srv.AddView(`^/cli$`, Bridge.cliView)
	srv.AddView(`^/consumer/(.+)$`, Bridge.consumerView)

	if *config.BridgeXtermRootPath != "" {
		mc := &memCache{elem: map[string][]byte{}, ct: make(map[string]string)}
		foundindex := false
		err := filepath.WalkDir(*config.BridgeXtermRootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			relpath, err := filepath.Rel(*config.BridgeXtermRootPath, path)
			if err != nil {
				return err
			}
			cast.LogDebug(fmt.Sprintf("Serving file from mem: %s", relpath), nil)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if relpath == "index.html" {
				Bridge.xtermindex = string(data)
				foundindex = true
				return nil
			}

			mc.elem["/xterm/"+relpath] = data
			mc.ct["/xterm/"+relpath] = mime.TypeByExtension(filepath.Ext(relpath))
			srv.AddView(`^/xterm/`+relpath+`$`, mc.serveFileCacheView)
			return nil
		})
		if !foundindex {
			cast.LogWarn(fmt.Sprintf("index.html not found in %s", *config.BridgeXtermRootPath), nil)
		}
		if err != nil {
			return errors.Join(fmt.Errorf("impossible to walk directories"), err)
		}
	}

	srv.AddView(`^/xterm/(.+)$`, Bridge.xtermjsView)

	errc := srv.ServeIfNot()
	err := <-errc
	return err
}

type memCache struct {
	elem map[string][]byte
	ct   map[string]string
	// req  int
}

func (m *memCache) serveFileCacheView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	if len(matches) <= 0 {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("Nothing to see here ¯\\_(ツ)_/¯"))
		return
	}
	match := matches[0]
	if len(match) <= 0 {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("Nothing to see here (╯°□°）╯︵ ┻━┻"))
		return
	}
	if _, exists := m.elem[matches[0][0]]; !exists {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("Nothing to see here ʕ⌐■ᴥ■ʔ"))
		return
	}
	w.Header().Add("Content-Type", m.ct[matches[0][0]])
	w.WriteHeader(http.StatusOK)
	w.Write(m.elem[matches[0][0]])
}

type newInBody struct {
	Auth string `json:"auth"`
}

type newOutBody struct {
	Token string `json:"token"`
}

// type cliInBody struct {
// 	Token string `json:"token"`
// }

// type consumerInBody struct {
// 	Token string `json:"token"`
// }

type pool struct {
	mu           sync.Mutex
	token        string
	cliConn      *websocket.Conn
	vpty         *nsterm.VPTY2
	consumerConn map[*websocket.Conn]bool
}

func (p *pool) syncAddConsumer(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consumerConn[conn] = true
}

func (p *pool) syncDeleteConsumer(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.consumerConn, conn)
}

type Puente struct {
	secret     string
	mu         sync.Mutex
	pools      map[string]*pool
	xtermindex string
}

func (p *Puente) syncAddPool(pl *pool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pools[pl.token] = pl
}

// TODO: add timeout to unused pools
func (p *Puente) newView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !POST", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body := http.MaxBytesReader(w, r.Body, 65536)
	data, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}
	inBody := &newInBody{}
	err = json.Unmarshal(data, inBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if inBody.Auth == "" {
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !Auth", http.StatusForbidden)
		return
	}

	if inBody.Auth != p.secret {
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !!Auth", http.StatusForbidden)
		return
	}

	b := make([]byte, 25)
	_, err = rand.Read(b)
	if err != nil {
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !Token", http.StatusInternalServerError)
		return
	}
	token := base64.StdEncoding.EncodeToString([]byte(b))
	token, _ = strings.CutSuffix(token, "==")

	newpool := &pool{
		token:        token,
		vpty:         nsterm.NewVirtPTY(),
		consumerConn: make(map[*websocket.Conn]bool),
	}
	newpool.vpty.SetLDisc(nsterm.NewMultiUserLdisc())
	p.syncAddPool(newpool)

	w.Header().Set("Content-Type", "application/json")

	outBody := &newOutBody{
		Token: newpool.token,
	}

	err = json.NewEncoder(w).Encode(outBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusInternalServerError)
	}
}

func (p *Puente) cliView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		cast.LogDebug("HTTPERR NO GET METHOD", nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !GET", http.StatusMethodNotAllowed)
		return
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		cast.LogDebug("HTTP ERR NO AUTH HEADER", nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !Authorization", http.StatusForbidden)
		return
	}

	sauth := strings.Split(auth, " ")
	if len(sauth) != 2 {
		cast.LogDebug("HTTP ERR NO AUTH HEADER 2", nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !!Authorization", http.StatusForbidden)
		return
	}

	token := sauth[1]

	pool, exists := p.pools[token]
	if !exists {
		cast.LogDebug("NO ", nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !token", http.StatusForbidden)
		return
	}

	cast.LogDebug("Upgrading...", nil)
	var upgrader = websocket.Upgrader{
		// ReadBufferSize:  MAXREADSIZE,
		// WriteBufferSize: MAXWRITESIZE,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		cast.LogDebug("Upgrading err"+err.Error(), nil)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// TODO: test if already exists

	pool.cliConn = conn
	moutfd := pool.vpty.MustarFD()
	wsrw := ws.NewWebSocketReadWriteCloser(conn)
	go func() {
		_, _ = io.Copy(wsrw, moutfd)
		moutfd.Close()
	}()
	cast.LogInfo("CLI connected", nil)
	_, _ = io.Copy(moutfd, wsrw)
	pool.vpty.Close()
	delete(p.pools, token)
}

func (p *Puente) consumerView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !GET", http.StatusMethodNotAllowed)
		return
	}

	cast.LogDebug(fmt.Sprintf("connecting consumer to %s", matches[0][1]), nil)
	pool, exists := p.pools[matches[0][1]]
	if !exists {
		cast.LogDebug(fmt.Sprintf("no token %s", matches[0][1]), nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !token", http.StatusForbidden)
		return
	}

	var upgrader = websocket.Upgrader{
		// ReadBufferSize:  MAXREADSIZE,
		// WriteBufferSize: MAXWRITESIZE,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		cast.LogInfo(fmt.Sprintf("cannot upgrade %s", matches[0][1]), nil)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	pool.syncAddConsumer(conn)
	sp := pool.vpty.NewSluvaPort()
	pool.vpty.CursorSluva(sp) // forces ldisc to know this new port
	soutfd := sp.OutFD()
	wsrw := ws.NewWebSocketReadWriteCloser(conn)
	go func() {
		_, _ = io.Copy(soutfd, wsrw)
	}()
	_, _ = io.Copy(wsrw, soutfd)
	wsrw.Close()
	cast.LogDebug("Consumer out", nil)
	pool.syncDeleteConsumer(conn)
	pool.vpty.DestroyPort(sp)
}

func (p *Puente) xtermjsView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		cast.LogDebug("HTTPERR NO GET METHOD", nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !GET", http.StatusMethodNotAllowed)
		return
	}

	_, exists := p.pools[matches[0][1]]
	if !exists {
		cast.LogDebug(fmt.Sprintf("no token %s", matches[0][1]), nil)
		http.Error(w, "(╯°□°)╯︵ ɹoɹɹƎ !valid token", http.StatusForbidden)
		return
	}

	xterm := p.xtermindex

	xterm = strings.Replace(xterm, "{TOKEN}", matches[0][1], 1)
	xterm = strings.Replace(xterm, "{HOST}", r.Host, 1)
	w.Write([]byte(xterm))
}
