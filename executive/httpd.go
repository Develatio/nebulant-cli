// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

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

package executive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandshakeResponse struct
type HandshakeResponse struct {
	Version string `json:"version"`
}

// HandshakeRequest struct
type HandshakeRequest struct {
	Version string `json:"version"`
}

// AutocompleteResponse struct
type AutocompleteResponse struct {
	Fail   bool                             `json:"Fail"`
	Errors []string                         `json:"Errors"`
	Result map[string][]*base.StorageRecord `json:"result"`
}

// GenericResponse struct
type GenericResponse struct {
	Fail   bool     `json:"Fail"`
	Errors []string `json:"Errors"`
}

type viewFunc func(w http.ResponseWriter, r *http.Request)

// Httpd struct
type Httpd struct {
	urls map[*regexp.Regexp]viewFunc
}

func (h *Httpd) validateOrigin(r *http.Request) bool {
	surl := r.Header.Get("Origin")
	// Fail on missing Origin header or bad url
	_, err := url.Parse(surl)
	if err != nil || surl == "" {
		return false
	}
	cast.LogDebug("Origin: "+surl, nil)
	// Allow on FrontOrigin match or wildcard in conf
	if surl == config.FrontOrigin || surl == config.FrontOriginPre || config.FrontOrigin == "*" {
		return true
	}
	return false
}

func (h *Httpd) httpMiddleware(w http.ResponseWriter, r *http.Request) error {
	if !h.validateOrigin(r) {
		w.WriteHeader(http.StatusBadRequest)
		return fmt.Errorf("bad Origin")
	}
	surl := r.Header.Get("Origin")
	cast.LogDebug("Access-Control-Allow-Origin: "+surl, nil)
	w.Header().Set("Access-Control-Allow-Origin", surl)
	return nil
}

func (h *Httpd) route(w http.ResponseWriter, r *http.Request) {
	var vfn viewFunc
	for rgx, fn := range h.urls {
		if rgx.MatchString(r.URL.Path) {
			vfn = fn
			break
		}
	}
	if vfn != nil {
		vfn(w, r)
	} else {
		http.Error(w, "404 Not found", http.StatusNotFound)
	}
}

// Serve func
func (h *Httpd) Serve(addr *string) error {
	h.urls = make(map[*regexp.Regexp]viewFunc)

	h.urls[regexp.MustCompile(`/ws/.+$`)] = func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = h.validateOrigin
		clientUUID := path.Base(r.URL.Path)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			cast.LogErr(err.Error(), nil)
			return
		}

		cast.NewWebSocketLogger(conn, clientUUID)
		if bc, err := cast.ER.GetByClient(clientUUID); err == nil {
			for reu, active := range bc {
				if !active {
					continue
				}
				// Wait for state luser! With love :*
				cast.PublishEvent(cast.EventWaitingForState, &reu)
				MDirector.ExecInstruction <- &ExecCtrlInstruction{
					Instruction:   ExecState,
					ExecutionUUID: &reu,
				}
			}
		}
	}

	h.urls[regexp.MustCompile(`/handshake`)] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
		err := h.httpMiddleware(w, r)
		if err != nil {
			cast.LogErr("WS Test fail: "+err.Error(), nil)
			return
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		body := http.MaxBytesReader(w, r.Body, 65536)
		data, readBodyErr := ioutil.ReadAll(body)
		if readBodyErr != nil {
			http.Error(w, "Handshake Failed", http.StatusPreconditionFailed)
		}
		hReq := &HandshakeRequest{}
		jsonErr := json.Unmarshal(data, hReq)
		if jsonErr != nil {
			http.Error(w, "Handshake Failed", http.StatusPreconditionFailed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		hRes := &HandshakeResponse{
			Version: config.Version,
		}

		if err := json.NewEncoder(w).Encode(hRes); err != nil {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusInternalServerError)
			return
		}

		cast.LogDebug("WS Test OK. Version: "+hReq.Version, nil)
	}

	h.urls[regexp.MustCompile(`/stop/.+$`)] = func(w http.ResponseWriter, r *http.Request) {
		err := h.httpMiddleware(w, r)
		if err != nil {
			cast.LogErr("Stop Request fail: "+err.Error(), nil)
			return
		}
		h.stopBlueprintView(w, r)
	}

	h.urls[regexp.MustCompile(`/pause/.+$`)] = func(w http.ResponseWriter, r *http.Request) {
		err := h.httpMiddleware(w, r)
		if err != nil {
			cast.LogErr("Pause Request fail: "+err.Error(), nil)
			return
		}
		h.pauseBlueprintView(w, r)
	}

	h.urls[regexp.MustCompile(`/resume/.+$`)] = func(w http.ResponseWriter, r *http.Request) {
		err := h.httpMiddleware(w, r)
		if err != nil {
			cast.LogErr("Resume Request fail: "+err.Error(), nil)
			return
		}
		h.resumeBlueprintView(w, r)
	}

	h.urls[regexp.MustCompile(`/blueprint/.+$`)] = func(w http.ResponseWriter, r *http.Request) {
		err := h.httpMiddleware(w, r)
		if err != nil {
			cast.LogErr("Blueprint Request fail: "+err.Error(), nil)
			return
		}
		h.blueprintView(w, r)
	}

	h.urls[regexp.MustCompile(`/autocomplete/$`)] = func(w http.ResponseWriter, r *http.Request) {
		err := h.httpMiddleware(w, r)
		if err != nil {
			cast.LogErr("Blueprint Request fail: "+err.Error(), nil)
			return
		}
		h.autocompleteView(w, r)
	}

	// start server
	cast.LogInfo("Listening on "+*addr, nil)
	http.HandleFunc("/", h.route)
	err := http.ListenAndServe(*addr, nil)
	// https:
	// err := http.ListenAndServeTLS(*addr, "localhost.crt", "localhost.key", nil)
	if err != nil {
		return err
	}
	return nil
}

func (h *Httpd) autocompleteView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body := http.MaxBytesReader(w, r.Body, 4000000)
	data, err := ioutil.ReadAll(body)
	if err != nil {
		http.Error(w, "E05 "+err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	bp, err := blueprint.NewFromBuilder(data)
	if err != nil {
		http.Error(w, "E06 "+err.Error(), http.StatusBadRequest)
		return
	}

	rand.Seed(time.Now().UnixNano())
	randIntString := fmt.Sprintf("%d", rand.Int()) //#nosec G404 -- Weak random is OK here
	bp.ExecutionUUID = &randIntString
	irbConf := &blueprint.IRBGenConfig{
		AllowResultReturn: true,
		// PreventLoop:       true,
	}
	irb, err := blueprint.GenerateIRB(bp, irbConf)
	if err != nil {
		http.Error(w, "E07 "+err.Error(), http.StatusBadRequest)
		return
	}

	fLink := &cast.FeedBackLink{
		FeedBackBus: make(chan *cast.FeedBack, 100),
	}
	cast.SBusConnect(fLink)
	// MDirector.RegisterManager <- manager
	MDirector.HandleIRB <- irb

	w.Header().Set("Content-Type", "application/json")
	var manager *Manager

	for fback := range fLink.FeedBackBus {
		// Ignore feedback without exec id
		if fback.ExecutionUUID == nil {
			continue
		}
		// Ignore messages not from this bp
		if *fback.ExecutionUUID != *bp.ExecutionUUID {
			continue
		}
		// exit on EOF
		if fback.TypeID == cast.FeedBackEOF {
			return
		}
		// Ignore non-event messages
		if fback.TypeID != cast.FeedBackEvent {
			continue
		}
		// save manager on registered manager event
		if *fback.EventID == cast.EventRegisteredManager {
			manager = fback.Extra["manager"].(*Manager)
			continue
		}
		// skip if manager isn't saved
		if manager == nil {
			continue
		}
		// skip if event isn't EventManagerOut
		if *fback.EventID != cast.EventManagerOut {
			continue
		}
		// here the manager has ended, get last action result and
		// bring back throught http

		status := http.StatusAccepted

		// Extract saved outputs and store into response struct
		resp := &AutocompleteResponse{}
		resp.Result = make(map[string][]*base.StorageRecord)
		outputLen := len(manager.ExternalRegistry.SavedOutputs)
		for i := 0; i < outputLen; i++ {
			actionID := manager.ExternalRegistry.SavedOutputs[i].Action.ActionID
			resp.Result[actionID] = manager.ExternalRegistry.SavedOutputs[i].Records
			for e := 0; e < len(manager.ExternalRegistry.SavedOutputs[i].Records); e++ {
				record := manager.ExternalRegistry.SavedOutputs[i].Records[e]
				if record.Fail {
					status = http.StatusBadRequest
					resp.Fail = true
					resp.Errors = append(resp.Errors, record.Error.Error())
				}
			}
		}

		w.WriteHeader(status)
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E08 "+err.Error(), http.StatusBadRequest)
		}
		break
	}
	cast.SBusDisconnect(fLink)
}

func (h *Httpd) stopBlueprintView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	remoteExecutionUUID := path.Base(r.URL.Path)
	MDirector.ExecInstruction <- &ExecCtrlInstruction{
		Instruction:   ExecStop,
		ExecutionUUID: &remoteExecutionUUID,
	}
	// echo
	cast.PublishEvent(cast.EventManagerStopping, &remoteExecutionUUID)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Httpd) pauseBlueprintView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	remoteExecutionUUID := path.Base(r.URL.Path)
	MDirector.ExecInstruction <- &ExecCtrlInstruction{
		Instruction:   ExecPause,
		ExecutionUUID: &remoteExecutionUUID,
	}
	// echo
	cast.PublishEvent(cast.EventManagerPausing, &remoteExecutionUUID)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Httpd) resumeBlueprintView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	remoteExecutionUUID := path.Base(r.URL.Path)
	MDirector.ExecInstruction <- &ExecCtrlInstruction{
		Instruction:   ExecResume,
		ExecutionUUID: &remoteExecutionUUID,
	}
	// echo
	cast.PublishEvent(cast.EventManagerResuming, &remoteExecutionUUID)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Httpd) blueprintView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body := http.MaxBytesReader(w, r.Body, 4000000)
	data, err := ioutil.ReadAll(body)
	if err != nil {
		http.Error(w, "E02 "+err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	bp, err := blueprint.NewFromBuilder(data)
	if err != nil {
		http.Error(w, "E03 "+err.Error(), http.StatusBadRequest)
		return
	}

	if bp.ExecutionUUID != nil {
		// Obtain clientid from url
		clientUUID := path.Base(r.URL.Path)
		// Create JOIN command for WebSocket logger and send it
		// to force ClientUUID to join in ExecutionUUID channel
		// where the program will write the log for this bp execution
		extra := make(map[string]interface{})
		extra["join"] = *bp.ExecutionUUID
		cast.PublishFiltered(clientUUID, extra)
	}

	cast.PublishEvent(cast.EventManagerPrepareBPStart, bp.ExecutionUUID)
	irb, err := blueprint.GenerateIRB(bp, &blueprint.IRBGenConfig{})
	if err != nil {
		cast.PublishEvent(cast.EventManagerPrepareBPEndWithErr, bp.ExecutionUUID)
		http.Error(w, "E04 "+err.Error(), http.StatusBadRequest)
		return
	}
	cast.PublishEvent(cast.EventManagerPrepareBPEnd, bp.ExecutionUUID)

	MDirector.HandleIRB <- irb
	w.WriteHeader(http.StatusAccepted)
}
