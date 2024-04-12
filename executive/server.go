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
	"io"
	"log"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/develatio/nebulant-cli/assets"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/nhttpd"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
)

var ServerMode bool = false

func InitServerMode() chan error {
	ServerMode = true
	defer func() {
		ServerMode = false
	}()
	go func() {
		cast.LogInfo("Updating assets in bg...", nil)
		aerr := assets.UpgradeAssets(false, false)
		if aerr != nil {
			cast.LogErr("Error loading assets", nil)
			cast.LogErr(aerr.Error(), nil)
		}
	}()

	srv := nhttpd.GetServer()
	srv.AddOrigin(config.FrontOrigin)
	srv.AddOrigin(config.BridgeOrigin)
	srv.AddOrigin(config.FrontOriginPre)
	srv.AddOrigin(config.FrontOrigin)

	srv.AddView(`/ws/.+$`, wsView)
	srv.AddView(`/handshake`, handshakeView)
	srv.AddView(`/stop/.+$`, stopBlueprintView)
	srv.AddView(`/pause/.+$`, pauseBlueprintView)
	srv.AddView(`/resume/.+$`, resumeBlueprintView)
	srv.AddView(`/blueprint/.+$`, blueprintView)
	srv.AddView(`/autocomplete/$`, autocompleteView)
	srv.AddView(`/assets/(.+)$`, assetsView)
	srv.AddView(`/proxy.html$`, proxyView)

	cast.LogInfo("The server mode is designed to be used with the Builder: "+config.FrontUrl, nil)
	return srv.ServeIfNot()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 400,
}

// HandshakeResponse struct
type HandshakeResponse struct {
	Version string `json:"version"`
}

type ValidationError struct {
	ValActionID string `json:"validation_actionid"`
	ValKey      string `json:"validation_key"`
	ValTag      string `json:"validation_tag"`
	ValError    string `json:"validation_error"`
}

// HandshakeRequest struct
type HandshakeRequest struct {
	Version string `json:"version"`
}

// AutocompleteResponse struct
type AutocompleteResponse struct {
	Code             string                           `json:"code"`
	Fail             bool                             `json:"Fail"`
	Errors           []string                         `json:"Errors"`
	ValidationErrors []*ValidationError               `json:"validation_errors"`
	Result           map[string][]*base.StorageRecord `json:"result"`
}

// GenericResponse struct
type GenericResponse struct {
	Code             string             `json:"code"`
	Fail             bool               `json:"Fail"`
	Errors           []string           `json:"Errors"`
	ValidationErrors []*ValidationError `json:"validation_errors"`
}

func parseErrors(err error) ([]string, []*ValidationError) {
	var errs []string
	var valerrs []*ValidationError
	if valErrs, isValErr := err.(validator.ValidationErrors); isValErr {
		// Field() returns the fields name with the tag name taking
		// precedence over the field's actual name.
		// eq. JSON name "fname"
		for _, fieldErr := range valErrs {
			errs = append(errs, fieldErr.Error())
			valerrs = append(valerrs, &ValidationError{ValKey: fieldErr.Field(), ValError: fieldErr.Error()})
		}
	} else if irbErrs, isIrbErrs := err.(blueprint.IRBErrors); isIrbErrs {
		// irb errors
		for _, irbErr := range irbErrs {
			// validation err, could be many
			if valErrs, isValErr := irbErr.WErr().(validator.ValidationErrors); isValErr {
				for _, fieldErr := range valErrs {
					valerrs = append(valerrs, &ValidationError{
						ValActionID: irbErr.ActionID(),
						ValKey:      fieldErr.Field(),
						ValTag:      fieldErr.Tag(),
						ValError:    fieldErr.Error(),
					})
					errs = append(errs, fieldErr.Error())
				}
			} else {
				errs = append(errs, irbErr.Error())
			}
		}
	} else {
		errs = append(errs, err.Error())
	}
	return errs, valerrs
}

func wsView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	// TODO: add wsocket features to server and prevent
	// getting the server from view
	srv := nhttpd.GetServer()
	upgrader.CheckOrigin = srv.ValidateOrigin
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
			cast.PushEvent(cast.EventWaitingForState, &reu)
			MDirector.ExecInstruction <- &ExecCtrlInstruction{
				Instruction:   ExecState,
				ExecutionUUID: &reu,
			}
		}
	}
}

func handshakeView(w http.ResponseWriter, r *http.Request, matches [][]string) {
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
	body := http.MaxBytesReader(w, r.Body, 65536)
	data, readBodyErr := io.ReadAll(body)
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

func autocompleteView(w http.ResponseWriter, r *http.Request, matches [][]string) {
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
	data, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "E05 "+err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	bp, err := blueprint.NewFromBuilder(data)
	if err != nil {
		http.Error(w, "E06 "+err.Error(), http.StatusBadRequest)
		return
	}

	// rand.Seed(time.Now().UnixNano())
	randIntString := fmt.Sprintf("%d", rand.Int()) // #nosec G404 -- Weak random is OK here
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

	// fLink := &cast.BusConsumerLink{
	// 	LogChan:        make(chan *cast.BusData, 100),
	// 	CommonChan:     make(chan *cast.BusData, 100),
	// 	AllowEventData: true,
	// }
	// cast.SBusConnect(fLink)
	// MDirector.RegisterManager <- manager
	manager := NewManager(true)
	manager.PrepareIRB(irb)
	MDirector.HandleIRB <- &HandleIRBConfig{Manager: manager}

	w.Header().Set("Content-Type", "application/json")

	// for fback := range fLink.CommonChan {
	// 	// Ignore feedback without exec id
	// 	if fback.ExecutionUUID == nil {
	// 		continue
	// 	}
	// 	// Ignore messages not from this bp
	// 	if *fback.ExecutionUUID != *bp.ExecutionUUID {
	// 		continue
	// 	}
	// 	// exit on EOF
	// 	if fback.TypeID == cast.BusDataTypeEOF {
	// 		return
	// 	}
	// 	// Ignore non-event messages
	// 	if fback.TypeID != cast.BusDataTypeEvent {
	// 		continue
	// 	}
	// 	// save manager on registered manager event
	// 	if *fback.EventID == cast.EventRegisteredManager {
	// 		manager = fback.Extra["manager"].(*Manager)
	// 		continue
	// 	}
	// 	// skip if manager isn't saved
	// 	if manager == nil {
	// 		continue
	// 	}
	// 	// skip if event isn't EventRuntimeOut
	// 	if *fback.EventID != cast.EventRuntimeOut {
	// 		continue
	// 	}
	// 	// here the manager has ended, get last action result and
	// 	// bring back throught http

	// eventlistener is ready because we called RegisterIRB before
	eventlistener := manager.Runtime.NewEventListener()
	eventlistener.WaitUntil([]base.EventCode{base.RuntimeEndEvent})

	status := http.StatusAccepted

	savedOutputs := manager.Runtime.SavedActionOutputs()

	// Extract saved outputs and store into response struct
	resp := &AutocompleteResponse{}
	resp.Result = make(map[string][]*base.StorageRecord)
	outputLen := len(savedOutputs)
	for i := 0; i < outputLen; i++ {
		actionID := savedOutputs[i].Action.ActionID
		resp.Result[actionID] = savedOutputs[i].Records
		for e := 0; e < len(savedOutputs[i].Records); e++ {
			record := savedOutputs[i].Records[e]
			if record.Fail {
				if status == http.StatusAccepted {
					status = http.StatusBadRequest
				}
				_, ok := err.(*base.ProviderAuthError)
				if ok {
					status = http.StatusUnauthorized
				}
				resp.Fail = true
				resp.Errors = append(resp.Errors, record.Error.Error())
			}
		}
	}

	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "E08 "+err.Error(), http.StatusBadRequest)
	}
	// break
	//}
	// cast.SBusDisconnect(fLink)
}

func stopBlueprintView(w http.ResponseWriter, r *http.Request, matches [][]string) {
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

	select {
	case MDirector.ExecInstruction <- &ExecCtrlInstruction{
		Instruction:   ExecStop,
		ExecutionUUID: &remoteExecutionUUID,
	}:
		cast.SBus.SetExecutionStatus(remoteExecutionUUID, false)
	default:
		http.Error(w, http.StatusText(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}
	// echo
	cast.PushEvent(cast.EventRuntimeStopping, &remoteExecutionUUID)
	w.WriteHeader(http.StatusAccepted)
}

func pauseBlueprintView(w http.ResponseWriter, r *http.Request, matches [][]string) {
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
	select {
	case MDirector.ExecInstruction <- &ExecCtrlInstruction{
		Instruction:   ExecPause,
		ExecutionUUID: &remoteExecutionUUID,
	}:
		cast.SBus.SetExecutionStatus(remoteExecutionUUID, false)
	default:
		http.Error(w, http.StatusText(http.StatusExpectationFailed), http.StatusExpectationFailed)
		return
	}

	// echo
	cast.PushEvent(cast.EventRuntimePausing, &remoteExecutionUUID)
	w.WriteHeader(http.StatusAccepted)
}

func resumeBlueprintView(w http.ResponseWriter, r *http.Request, matches [][]string) {
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
	cast.PushEvent(cast.EventRuntimeResuming, &remoteExecutionUUID)
	w.WriteHeader(http.StatusAccepted)
}

func blueprintView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		resp := &GenericResponse{
			Code:             "E01",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusMethodNotAllowed)},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E01 "+err.Error(), http.StatusMethodNotAllowed)
		}
		return
	}

	defer r.Body.Close()
	body := http.MaxBytesReader(w, r.Body, 4000000)
	data, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		errs, valerrs := parseErrors(err)
		resp := &GenericResponse{
			Code:             "E02",
			Fail:             true,
			Errors:           append(errs, http.StatusText(http.StatusRequestEntityTooLarge)),
			ValidationErrors: valerrs,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E02 "+err.Error(), http.StatusRequestEntityTooLarge)
		}
		return
	}
	bp, err := blueprint.NewFromBuilder(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errs, valerrs := parseErrors(err)
		resp := &GenericResponse{
			Code:             "E03",
			Fail:             true,
			Errors:           append(errs, http.StatusText(http.StatusBadRequest)),
			ValidationErrors: valerrs,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E03 "+err.Error(), http.StatusBadRequest)
		}
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
		cast.PushFilteredBusData(clientUUID, extra)
	}

	cast.PushEvent(cast.EventManagerPrepareBPStart, bp.ExecutionUUID)
	irb, err := blueprint.GenerateIRB(bp, &blueprint.IRBGenConfig{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errs, valerrs := parseErrors(err)
		resp := &GenericResponse{
			Code:             "E04",
			Fail:             true,
			Errors:           errs,
			ValidationErrors: valerrs,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E04 "+err.Error(), http.StatusBadRequest)
		}
		cast.PushEvent(cast.EventManagerPrepareBPEndWithErr, bp.ExecutionUUID)
		return
	}
	cast.PushEvent(cast.EventManagerPrepareBPEnd, bp.ExecutionUUID)

	MDirector.HandleIRB <- &HandleIRBConfig{IRB: irb}
	w.WriteHeader(http.StatusAccepted)
}

func assetsView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		resp := &GenericResponse{
			Code:             "E01",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusMethodNotAllowed)},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E01 "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	// https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
	if assets.State.IsUpgradeInProgress() {
		w.WriteHeader(512) // 512 is not part of ianna assignments but means "not updated"
		resp := &GenericResponse{
			Code:             "E00",
			Fail:             true,
			Errors:           []string{http.StatusText(512), "Upgrade assets in progress"},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E00 "+err.Error(), 512)
		}
		return
	}

	if assets.State.NeedUpgrade() {
		cast.LogWarn("Assets need to be updated", nil)
	}

	u := r.URL
	asset_id := matches[0][1]
	asset_id = strings.TrimSuffix(asset_id, "/")

	assetdef, ok := assets.AssetsDefinition[asset_id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		resp := &GenericResponse{
			Code:             "E02",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusNotFound), "Unknown asset " + asset_id},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E02 "+err.Error(), http.StatusNotFound)
		}
		return
	}

	q := u.Query()
	searchq := q.Get("search")
	if len(searchq) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp := &GenericResponse{
			Code:             "E03",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusBadRequest), "Search query not found"},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E03 "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	// TODO: implement sort and pagination
	rlimit := q.Get("limit")
	roffset := q.Get("offset")
	rsort := q.Get("sort")
	var limit int64 = 0
	var offset int64 = 0
	if len(rlimit) > 0 {
		var err error
		limit, err = strconv.ParseInt(rlimit, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := &GenericResponse{
				Code:             "E04",
				Fail:             true,
				Errors:           []string{http.StatusText(http.StatusBadRequest), err.Error()},
				ValidationErrors: nil,
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "E04 "+err.Error(), http.StatusBadRequest)
			}
			return
		}
	}
	if len(roffset) > 0 {
		var err error
		offset, err = strconv.ParseInt(roffset, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := &GenericResponse{
				Code:             "E04b",
				Fail:             true,
				Errors:           []string{http.StatusText(http.StatusBadRequest), err.Error()},
				ValidationErrors: nil,
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "E04b "+err.Error(), http.StatusBadRequest)
			}
			return
		}
	}

	searchres, err := assets.Search(&assets.SearchRequest{
		SearchTerm:  searchq,
		FilterTerms: q,
		Limit:       int(limit),
		Offset:      int(offset),
		Sort:        rsort,
	}, assetdef)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := &GenericResponse{
			Code:             "E05",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusBadRequest), err.Error()},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E05 "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(searchres)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resp := &GenericResponse{
			Code:             "E06",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusUnprocessableEntity), err.Error()},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E06 "+err.Error(), http.StatusUnprocessableEntity)
		}
		return
	}
}

func proxyView(w http.ResponseWriter, r *http.Request, matches [][]string) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		resp := &GenericResponse{
			Code:             "E01",
			Fail:             true,
			Errors:           []string{http.StatusText(http.StatusMethodNotAllowed)},
			ValidationErrors: nil,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "E01 "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cache-Control", "no-cache")

	w.Write([]byte(assets.PROXYHTTP))
}
