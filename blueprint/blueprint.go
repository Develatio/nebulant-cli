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

package blueprint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/util"
	"golang.org/x/mod/semver"
)

// JoinThreadsActionName const
const JoinThreadsActionName = "join_threads"

type wrappedBlueprint struct {
	ExecutionUUID *string         `json:"execution_uuid"`
	Detail        string          `json:"detail"`
	Blueprint     json.RawMessage `json:"blueprint"`
}

// Blueprint struct
type Blueprint struct {
	ExecutionUUID *string
	Actions       []Action `json:"actions"`
	MinCLIVersion *string  `json:"min_cli_version"`
	Raw           *[]byte
}

// Action struct
type Action struct {
	// Filled internally.
	Parents          []*Action
	JoinThreadsPoint bool
	KnowParentIDs    map[string]bool
	SafeID           *string
	// GENERICS //
	Provider    string     `json:"provider" validate:"required"`
	ActionID    string     `json:"action_id" validate:"required"`
	ActionName  string     `json:"action" validate:"required"`
	FirstAction bool       `json:"first_action"`
	NextAction  NextAction `json:"next_action"`

	// Delegated parse.
	Input      json.RawMessage `json:"input" validate:"required"`
	Parameters json.RawMessage `json:"parameters" validate:"required"`

	// Actor out vars names.
	Output         *string `json:"output"`
	SaveRawResults bool    `json:"save_raw_results"`
	DebugNetwork   bool    `json:"debug_network"`
}

type ConditionalNextActions struct {
	True  []string `json:"true"`
	False []string `json:"false"`
}

// NextAction struct
type NextAction struct {
	// Filled internally
	NextOk []*Action
	// Used internally. Is this a conditional next?
	ConditionalNext bool
	//
	NextOkTrue  []*Action
	NextOkFalse []*Action
	//
	NextKo []*Action
	// Parsed in precompiling.
	Ok json.RawMessage `json:"ok"`
	Ko json.RawMessage `json:"ko"`
}

// InternalParameters struct
// Config params defined to be used internally.
type InternalParameters struct {
	Waiters []string `json:"_waiters"`
}

func TestMinCliVersion(bp *Blueprint) error {
	if bp.MinCLIVersion == nil {
		return nil
	}
	if !semver.IsValid("v" + *bp.MinCLIVersion) {
		return fmt.Errorf("invalid min_cli_version value: " + *bp.MinCLIVersion)
	}
	if !semver.IsValid("v" + config.Version) {
		return fmt.Errorf("invalid nebulant version: " + config.Version)
	}
	if c := semver.Compare("v"+*bp.MinCLIVersion, "v"+config.Version); c == 1 {
		return fmt.Errorf("min CLI version not satisfied for this blueprint. Needed: " + *bp.MinCLIVersion + ". Got: " + config.Version)
	}
	return nil
}

// NewFromFile func
func NewFromFile(path string) (*Blueprint, error) {
	jsonFile, err := os.Open(path) //#nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	wrap := &wrappedBlueprint{}
	if err := json.Unmarshal(byteValue, wrap); err != nil {
		return nil, err
	}
	var bp Blueprint
	if err := util.UnmarshalValidJSON(wrap.Blueprint, &bp); err != nil {
		return nil, err
	}
	if err := TestMinCliVersion(&bp); err != nil {
		return nil, err
	}
	if bp.ExecutionUUID == nil || *bp.ExecutionUUID == "" {
		rand.Seed(time.Now().UnixNano())
		randIntString := fmt.Sprintf("%d", rand.Int()) //#nosec G404 -- Weak random is OK here
		bp.ExecutionUUID = &randIntString
	}
	bp.Raw = &byteValue
	return &bp, nil
}

// NewFromBytes func
func NewFromBytes(data []byte) (*Blueprint, error) {
	var bp Blueprint
	if err := util.UnmarshalValidJSON(data, &bp); err != nil {
		return nil, err
	}
	bp.Raw = &data
	if err := TestMinCliVersion(&bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

// NewFromBuilder func
func NewFromBuilder(data []byte) (*Blueprint, error) {
	var bp Blueprint
	wrap := &wrappedBlueprint{}
	if err := json.Unmarshal(data, wrap); err != nil {
		return nil, err
	}
	if err := util.UnmarshalValidJSON(wrap.Blueprint, &bp); err != nil {
		return nil, err
	}
	if err := TestMinCliVersion(&bp); err != nil {
		return nil, err
	}
	bp.ExecutionUUID = wrap.ExecutionUUID
	if bp.ExecutionUUID == nil || *bp.ExecutionUUID == "" {
		rand.Seed(time.Now().UnixNano())
		randIntString := fmt.Sprintf("%d", rand.Int()) //#nosec G404 -- Weak random is OK here
		bp.ExecutionUUID = &randIntString
	}

	return &bp, nil
}

func NewIRBFromAny(any string) (*IRBlueprint, error) {
	var bp *Blueprint
	var err error
	if len(any) > 11 && any[:11] == "nebulant://" {
		bp, err = NewFromBackend(any[11:])
		if err != nil {
			return nil, err
		}
	} else {
		bp, err = NewFromFile(any)
		if err != nil {
			return nil, err
		}
	}
	irb, err := GenerateIRB(bp, &IRBGenConfig{})
	if err != nil {
		return nil, err
	}
	return irb, nil
}

// NewFromBackend func
func NewFromBackend(path string) (*Blueprint, error) {
	if config.CREDENTIALS.AuthToken == nil {
		return nil, fmt.Errorf("auth token not found. Please set NEBULANT_TOKEN_ID and NEBULANT_TOKEN_SECRET env vars")
	}

	url := url.URL{Scheme: config.BackendProto, Host: config.BackendURLDomain, Path: "/apiv1/cli/" + path}
	rawBody, _ := json.Marshal(map[string]string{
		"version": config.Version,
	})
	reqBody := bytes.NewBuffer(rawBody)
	req, err := http.NewRequest("GET", url.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", *config.CREDENTIALS.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rawbody, _ := ioutil.ReadAll(resp.Body)
	body := &wrappedBlueprint{}
	if resp.StatusCode != http.StatusOK {
		if body.Detail != "" {
			return nil, fmt.Errorf(strconv.Itoa(resp.StatusCode) + " cannot obtain blueprint from server: " + body.Detail + " " + string(rawbody))
		}
		fmt.Println()
		return nil, fmt.Errorf(strconv.Itoa(resp.StatusCode) + " cannot obtain blueprint from server " + string(rawbody))
	}
	if err := json.Unmarshal(rawbody, body); err != nil {
		return nil, err
	}
	bp, err := NewFromBytes(body.Blueprint)
	if err != nil {
		return nil, err
	}

	bp.ExecutionUUID = body.ExecutionUUID

	return bp, nil
}

// IRBlueprint struct. Intermediate Representation Blueprint. Precompiler.
type IRBlueprint struct {
	BP            *Blueprint
	ExecutionUUID *string
	// [thread-action-id][thread-path]*Action
	JoinThreadPoints map[string]*Action
	Actions          map[string]*Action
	StartAction      *Action
}

// IRBGenConfig struct
type IRBGenConfig struct {
	AllowResultReturn bool
	// Not implemented
	// PreventLoop       bool
}

type IRBError interface {
	// String representation, needed for error interface
	Error() string
	// ActionID if exists, empty string if not
	ActionID() string
	// The raw wrapped error
	WErr() error
}

type iRBError struct {
	actionID string
	wErr     error
}

func (ie *iRBError) Error() string {
	return ie.wErr.Error()
}

func (ie *iRBError) ActionID() string {
	return ie.actionID
}

func (ie *iRBError) WErr() error {
	return ie.wErr
}

type IRBErrors []IRBError

func (ies IRBErrors) Error() string {
	var res string
	var ie *iRBError

	for i := 0; i < len(ies); i++ {
		ie = ies[i].(*iRBError)
		res = res + "\n" + ie.Error()
	}
	return res
}

// GenerateIRB func
func GenerateIRB(bp *Blueprint, irbConf *IRBGenConfig) (*IRBlueprint, error) {
	var errors IRBErrors
	irb := &IRBlueprint{
		BP:               bp,
		Actions:          make(map[string]*Action),
		JoinThreadPoints: make(map[string]*Action),
	}

	err := PreValidate(bp)
	if err != nil {
		errors = append(errors, &iRBError{wErr: err})
	}

	irb.ExecutionUUID = bp.ExecutionUUID

	// iterate over bp, check provider access
	for i := 0; i < len(bp.Actions); i++ {
		irb.Actions[bp.Actions[i].ActionID] = &bp.Actions[i]
		if !irbConf.AllowResultReturn {
			irb.Actions[bp.Actions[i].ActionID].SaveRawResults = false
		}

		// Detecting first action
		if bp.Actions[i].FirstAction {
			irb.StartAction = &bp.Actions[i]
		}

		// Set safe ID. Can be used internally when no user feedback is needed
		// or when no box track is needed.
		safeID := strconv.Itoa(i)
		irb.Actions[bp.Actions[i].ActionID].SafeID = &safeID
		if irb.Actions[bp.Actions[i].ActionID].ActionName == JoinThreadsActionName {
			irb.Actions[bp.Actions[i].ActionID].JoinThreadsPoint = true
			irb.JoinThreadPoints[bp.Actions[i].ActionID] = irb.Actions[bp.Actions[i].ActionID]
		}
	}

	if irb.StartAction == nil {
		errors = append(errors, &iRBError{wErr: fmt.Errorf("no first action found in blueprint")})
		// return nil, fmt.Errorf("no first action found in blueprint")
	}

	for _, action := range irb.Actions {
		// empty parameters
		if len(action.Parameters) <= 0 {
			action.Parameters = []byte("{}")
		}

		if action.ActionName == "condition" && action.Provider == "generic" {
			action.NextAction.ConditionalNext = true
		}

		if action.Output != nil && strings.ToLower(*action.Output) == "env" {
			errors = append(errors, &iRBError{
				actionID: action.ActionID,
				wErr:     fmt.Errorf("no first action found in blueprint"),
			})
			// return nil, fmt.Errorf("invalid output var name ENV. ENV is a reserved word")
		}

		// parse and fill next and parents
		nextOkActions, nextTrueActions, nextFalseActions, err := parseNextActions(action.NextAction.Ok, irb.Actions)
		if err != nil {
			errors = append(errors, &iRBError{wErr: err})
			// return nil, err
		}
		if nextOkActions != nil {
			for _, nextact := range nextOkActions {
				nextact.Parents = append(nextact.Parents, action)
			}
			// Add all actions (true and false) to NextOk.
			action.NextAction.NextOk = nextOkActions
			action.NextAction.NextOkTrue = nextTrueActions
			action.NextAction.NextOkFalse = nextFalseActions
		}

		nextKoActions, _, _, err := parseNextActions(action.NextAction.Ko, irb.Actions)
		if err != nil {
			errors = append(errors, &iRBError{wErr: err})
			// return nil, err
		}
		if nextKoActions != nil {
			for _, nextact := range nextKoActions {
				nextact.Parents = append(nextact.Parents, action)
			}
			action.NextAction.NextKo = nextKoActions
		}

		for _, vl := range ActionValidators {
			if err := vl(action); err != nil {
				errors = append(errors, &iRBError{actionID: action.ActionID, wErr: err})
				// return nil, fmt.Errorf("Error found in action " + action.ActionID + ":\n" + err.Error())
			}
		}
	}

	// prepare for join threads
	for _, action := range irb.JoinThreadPoints {
		knowParents := buildParentPaths(action)
		action.KnowParentIDs = knowParents
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return irb, nil
}

func buildParentPaths(action *Action) map[string]bool {
	knowParents := make(map[string]bool)
	queueParents := action.Parents
	var processingParents []*Action

	for len(queueParents) > 0 {
		processingParents = queueParents
		queueParents = nil
		for _, p := range processingParents {
			knowParents[p.ActionID] = true
			for _, pp := range p.Parents {
				if _, exists := knowParents[pp.ActionID]; exists {
					continue
				}
				if action == pp { // skip self
					continue
				}
				queueParents = append(queueParents, pp)
			}
		}
	}

	return knowParents
}

func parseNextActions(okko json.RawMessage, actions map[string]*Action) ([]*Action, []*Action, []*Action, error) {
	// missing field
	if len(okko) <= 0 {

		return nil, nil, nil, nil
	}

	// allowed null field
	if string(okko) == "null" {
		return nil, nil, nil, nil
	}

	// allowed null field
	if okko == nil {
		return nil, nil, nil, nil
	}

	var err error
	var nextActions []*Action
	// Next actions from a conditional box with True at output.
	var nextTrueActions []*Action
	// Next actions from a conditional box with False output.
	var nextFalseActions []*Action
	var normalActions []string
	var conditionals *ConditionalNextActions = new(ConditionalNextActions)

	// for ["id", "id2", ...] format
	err = json.Unmarshal(okko, &normalActions)
	if err == nil {
		for _, oneID := range normalActions {
			action := actions[oneID]
			if action == nil {
				return nil, nil, nil, fmt.Errorf("reference to unknown action")
			}
			nextActions = append(nextActions, action)
		}
		return nextActions, nextTrueActions, nextFalseActions, nil
	}

	// for {"true": ["id", "id2", ... ], "false": ["id", "id2", ... ]} format
	err = util.UnmarshalValidJSON(okko, conditionals)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse Next syntax")
	}

	if len(conditionals.True) > 0 {
		for _, oneID := range conditionals.True {
			action := actions[oneID]
			if action == nil {
				return nil, nil, nil, fmt.Errorf("reference to unknown action")
			}
			nextActions = append(nextActions, action)
			nextTrueActions = append(nextTrueActions, action)
		}
	}

	if len(conditionals.False) > 0 {
		for _, oneID := range conditionals.False {
			action := actions[oneID]
			if action == nil {
				return nil, nil, nil, fmt.Errorf("reference to unknown action")
			}
			nextActions = append(nextActions, action)
			nextFalseActions = append(nextFalseActions, action)
		}
	}

	return nextActions, nextTrueActions, nextFalseActions, nil
}
