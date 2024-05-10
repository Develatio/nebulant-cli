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
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/util"
	"golang.org/x/mod/semver"
)

// JoinThreadsActionName const
const JoinThreadsActionName = "join_threads"
const DebugActionName = "debug"

type wrappedBlueprint struct {
	ExecutionUUID *string         `json:"execution_uuid"`
	Detail        string          `json:"detail"`
	Blueprint     json.RawMessage `json:"blueprint"`
}

// Blueprint struct
type Blueprint struct {
	ExecutionUUID   *string
	Actions         []Action `json:"actions"`
	MinCLIVersion   *string  `json:"min_cli_version"`
	Raw             *[]byte
	BuilderErrors   int `json:"n_errors"`
	BuilderWarnings int `json:"n_warnings"`
}

// Action struct
type Action struct {
	// Filled internally.
	Parents          []*Action
	JoinThreadsPoint bool
	DebugPoint       bool
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
	// Not documented
	MaxRetries *int `json:"max_retries"`
	RetryCount int
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
	// filled internally
	// detect loops through ok and ko
	NextOkLoop bool
	NextKoLoop bool
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
func NewFromFile(bpUrl *BlueprintURL) (*Blueprint, error) {
	jsonFile, err := os.Open(bpUrl.Path) // #nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
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
		// rand.Seed is deprecated: As of Go 1.20 there is no reason to call Seed with a random value
		// rand.Seed(time.Now().UnixNano())
		randIntString := fmt.Sprintf("%d", rand.Int()) // #nosec G404 -- Weak random is OK here
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
		randIntString := fmt.Sprintf("%d", rand.Int()) // #nosec G404 -- Weak random is OK here
		bp.ExecutionUUID = &randIntString
	}

	return &bp, nil
}

func NewIRBFromAny(bpurl *BlueprintURL, irbConf *IRBGenConfig) (*IRBlueprint, error) {
	var bp *Blueprint
	var err error

	switch bpurl.Scheme {
	case "nebulant":
		bp, err = NewFromBackend(bpurl)
		if err != nil {
			return nil, err
		}
	case "file":
		bp, err = NewFromFile(bpurl)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown bp url")
	}

	irb, err := GenerateIRB(bp, irbConf)
	if err != nil {
		return nil, err
	}
	return irb, nil
}

func NewFromMarket(bpUrl *BlueprintURL) (*Blueprint, error) {
	orgslug := bpUrl.OrganizationSlug
	if orgslug == "" {
		orgslug = bpUrl.CollectionSlug
	}
	path := ""
	if bpUrl.Version == "" {
		path = fmt.Sprintf(
			config.MARKETPLACE_GET_BLUEPRINT_PATH,
			orgslug,
			bpUrl.CollectionSlug,
			bpUrl.BlueprintSlug)
	} else {
		path = fmt.Sprintf(
			config.MARKETPLACE_GET_BLUEPRINT_VERSION_PATH,
			orgslug,
			bpUrl.CollectionSlug,
			bpUrl.BlueprintSlug,
			bpUrl.Version)
	}
	url := &url.URL{
		Scheme: config.BASE_SCHEME,
		Host:   config.MARKET_API_HOST,
		Path:   path,
	}
	return getRemoteBP(url)
}

// NewFromBackend func
func NewFromBackend(bpUrl *BlueprintURL) (*Blueprint, error) {
	if config.CREDENTIAL.AuthToken == nil {
		return NewFromMarket(bpUrl)
		// return nil, fmt.Errorf("auth token not found. Please set NEBULANT_TOKEN_ID and NEBULANT_TOKEN_SECRET environment variables or use 'nebulant auth' command to authenticate and generate a CLI token")
	}

	_, err := config.Login(nil)
	if err != nil {
		return NewFromMarket(bpUrl)
	}

	if config.PROFILE == nil {
		return NewFromMarket(bpUrl)
	}

	if config.PROFILE.Organization.Slug != bpUrl.OrganizationSlug {
		return NewFromMarket(bpUrl)
	}

	path := ""
	if bpUrl.Version == "" {
		path = fmt.Sprintf(
			config.BACKEND_GET_BLUEPRINT_PATH,
			bpUrl.CollectionSlug,
			bpUrl.BlueprintSlug)
	} else {
		path = fmt.Sprintf(
			config.BACKEND_GET_BLUEPRINT_VERSION_PATH,
			bpUrl.CollectionSlug,
			bpUrl.BlueprintSlug,
			bpUrl.Version)
	}

	url := &url.URL{
		Scheme: config.BASE_SCHEME,
		Host:   config.BACKEND_API_HOST,
		Path:   path,
	}

	return getRemoteBP(url)
}

func getRemoteBP(url *url.URL) (*Blueprint, error) {
	rawBody, _ := json.Marshal(map[string]string{
		"version": config.Version,
	})
	reqBody := bytes.NewBuffer(rawBody)
	req, err := http.NewRequest("GET", url.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Jar: config.GetJar()}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rawbody, _ := io.ReadAll(resp.Body)
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

func ParseBPArgs(args []string) ([]*IRBArg, error) {
	var parsed []*IRBArg
	for _, arg := range args {
		_arg, found := strings.CutPrefix(arg, "--")
		if !found {
			_arg, found = strings.CutPrefix(arg, "-")
			if !found {
				return nil, fmt.Errorf("malformed var name %v. Please prefix var names with one (-) or two (--) dashes. Example: --varname=value", arg)
			}
		}
		sepidx := strings.Index(_arg, "=")
		if sepidx == -1 || sepidx+1 == len(_arg) {
			parsed = append(parsed, &IRBArg{Name: strings.TrimSuffix(_arg, "="), Value: ""})
			continue
		}
		parsed = append(parsed, &IRBArg{Name: _arg[:sepidx], Value: _arg[sepidx+1:]})
	}

	return parsed, nil
}

// IRBArg struct. Represent parsed blueprint cli args
type IRBArg struct {
	Name  string
	Value string
}

// IRBlueprint struct. Intermediate Representation Blueprint. Precompiler.
type IRBlueprint struct {
	BP            *Blueprint
	ExecutionUUID *string
	// [thread-action-id][thread-path]*Action
	JoinThreadPoints map[string]*Action
	Actions          map[string]*Action
	StartAction      *Action
	Args             []*IRBArg
}

// IRBGenConfig struct
type IRBGenConfig struct {
	AllowResultReturn bool
	Args              []string
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
	if bp.BuilderErrors > 0 {
		return nil, fmt.Errorf("Refusing to run this blueprint as it contains " + fmt.Sprintf("%v", bp.BuilderErrors) + " errors")
	}
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

	// parse cli args
	pargs, err := ParseBPArgs(irbConf.Args)
	if err != nil {
		return nil, err
	}
	irb.Args = pargs

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
		if irb.Actions[bp.Actions[i].ActionID].ActionName == DebugActionName {
			irb.Actions[bp.Actions[i].ActionID].DebugPoint = true
		}
	}

	if irb.StartAction == nil {
		errors = append(errors, &iRBError{wErr: fmt.Errorf("no first action found in blueprint")})
		// return nil, fmt.Errorf("no first action found in blueprint")
	}

	// parse next and parents
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

	// WIP: ASK FOR INPUT HOWTO
	// go func() {
	// 	var first string
	// 	lin := term.AppendLine()
	// 	lin.Scanln("Enter value for var: ", &first)
	// 	// fmt.Println("Var setted to: ", first)
	// 	lin.Write([]byte("var setted to " + first))
	// 	err = lin.Close()
	// 	if err != nil {
	// 		panic("there is a problem with the terminal")
	// 	}
	// }()

	// parse end cajitas
	for _, action := range irb.Actions {
		action.NextAction.NextOk = replaceEndActions(action.NextAction.NextOk, true)
		action.NextAction.NextOkTrue = replaceEndActions(action.NextAction.NextOkTrue, true)
		action.NextAction.NextOkFalse = replaceEndActions(action.NextAction.NextOkFalse, true)
		action.NextAction.NextKo = replaceEndActions(action.NextAction.NextKo, false)
	}

	// prepare for join threads
	for _, action := range irb.JoinThreadPoints {
		knowParents := buildDirectAscendants(action)
		action.KnowParentIDs = knowParents
	}

	// detect loops
	for _, action := range irb.Actions {
	L1ok:
		for _, ca := range action.NextAction.NextOk {
			knowParents := buildDirectAscendants(ca)
			if _, exists := knowParents[action.ActionID]; exists {
				action.NextAction.NextOkLoop = true
				break L1ok
			}
		}
	L2ko:
		for _, ca := range action.NextAction.NextKo {
			knowParents := buildDirectAscendants(ca)
			if _, exists := knowParents[action.ActionID]; exists {
				action.NextAction.NextKoLoop = true
				break L2ko
			}
		}
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return irb, nil
}

// buildDirectAscendants func
// extract all degrees of direct ascendant
func buildDirectAscendants(action *Action) map[string]bool {
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

func replaceEndActions(actions []*Action, OK bool) []*Action {
	var nextActions []*Action
	for _, action := range actions {
		if action.ActionName == "end" && action.Provider == "generic" {
			if OK {
				nextActions = append(nextActions, action.NextAction.NextOk...)
				continue
			}
			nextActions = append(nextActions, action.NextAction.NextKo...)
			continue
		}
		nextActions = append(nextActions, action)
	}
	return nextActions
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

	// Parse non-conditional next actions

	// for ["id", "id2", ...] format
	err = json.Unmarshal(okko, &normalActions)
	if err == nil {
		for _, nextID := range normalActions {
			action := actions[nextID]
			if action == nil {
				return nil, nil, nil, fmt.Errorf("reference to unknown action")
			}
			nextActions = append(nextActions, action)
		}
		return nextActions, nextTrueActions, nextFalseActions, nil
	}

	// Parse conditional next actions

	// for {"true": ["id", "id2", ... ], "false": ["id", "id2", ... ]} format
	err = util.UnmarshalValidJSON(okko, conditionals)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse Next syntax")
	}

	if len(conditionals.True) > 0 {
		for _, nextTrueID := range conditionals.True {
			action := actions[nextTrueID]
			if action == nil {
				return nil, nil, nil, fmt.Errorf("reference to unknown action")
			}
			nextActions = append(nextActions, action)
			nextTrueActions = append(nextTrueActions, action)
		}
	}

	if len(conditionals.False) > 0 {
		for _, nextFalseID := range conditionals.False {
			action := actions[nextFalseID]
			if action == nil {
				return nil, nil, nil, fmt.Errorf("reference to unknown action")
			}
			nextActions = append(nextActions, action)
			nextFalseActions = append(nextFalseActions, action)
		}
	}

	return nextActions, nextTrueActions, nextFalseActions, nil
}
