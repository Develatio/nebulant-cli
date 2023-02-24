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

package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/bhmj/jsonslice"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/config"
)

// Store struct
type Store struct {
	// Records into this store. All records are stored in three ways
	// by reference name, action id and provider id:
	//
	// by reference like {{ VARNAME }}
	recordsByRefName map[string]*base.StorageRecord
	// by action id like Action: { ActionID: "d8s8a9...." }
	recordsByActionID map[string]*base.StorageRecord
	// by provider id like "providerprefix -" + ec2.Image.ImageId
	recordsByValueID map[string]*base.StorageRecord
	//
	//
	// attrTree the string key match with string key of records
	// and it's a representation of "key.path":"value" of struct stored into
	// base.ActionOutput.Raw|Underlying
	// attrTree map[string]map[string]*base.AttrTreeValue
	// records same as records but using ActionID as key
	aoutByActionID map[string]*base.ActionOutput
	// initialized providers
	providers map[string]base.IProvider
	// logger instance
	logger base.ILogger
	// private store to be used privately by providers and not exposed to user
	private map[string]interface{}
}

// NewStore func
func NewStore() *Store {
	store := Store{}
	store.recordsByRefName = make(map[string]*base.StorageRecord)
	store.recordsByActionID = make(map[string]*base.StorageRecord)
	store.recordsByValueID = make(map[string]*base.StorageRecord)

	store.aoutByActionID = make(map[string]*base.ActionOutput)
	store.providers = make(map[string]base.IProvider)
	store.private = make(map[string]interface{})
	return &store
}

// Merge func
func (s *Store) Merge(source base.IStore) {
	ss := source.(*Store)
	for k, v := range ss.recordsByRefName {
		s.recordsByRefName[k] = v
	}
	for k, v := range ss.recordsByActionID {
		s.recordsByActionID[k] = v
	}
	for k, v := range ss.aoutByActionID {
		s.aoutByActionID[k] = v
	}
	for k, v := range ss.providers {
		s.providers[k] = v
	}
	for k, v := range ss.private {
		s.private[k] = v
	}
}

// Duplicate func.
// Make a copy of current store to be used in newly created children threads.
func (s *Store) Duplicate() base.IStore {
	var recordsByRefName = make(map[string]*base.StorageRecord)
	var recordsByActionID = make(map[string]*base.StorageRecord)
	var recordsByValueID = make(map[string]*base.StorageRecord)
	var aoutByActionID = make(map[string]*base.ActionOutput)
	var private = make(map[string]interface{})

	for k, v := range s.recordsByRefName {
		vv := *v
		recordsByRefName[k] = &vv
	}
	for k, v := range s.recordsByActionID {
		vv := v
		recordsByActionID[k] = vv
	}
	for k, v := range s.recordsByValueID {
		vv := v
		recordsByValueID[k] = vv
	}
	for k, v := range s.aoutByActionID {
		vv := *v
		aoutByActionID[k] = &vv
	}
	for pvn, pv := range s.private {
		private[pvn] = pv
	}

	store := NewStore()
	store.recordsByRefName = recordsByRefName
	store.recordsByActionID = recordsByActionID
	store.recordsByValueID = recordsByValueID
	store.aoutByActionID = aoutByActionID
	store.private = private
	//
	vr := reflect.ValueOf(s.logger)
	if vr.Kind() == reflect.Ptr {
		store.logger = s.logger.Duplicate()
	}
	// NOTE: On Store duplication, no current providers instance are copied.
	// This is the desired behavior. Every stage/thread should start his own
	// provider instance with his own store. This is performed
	// by stage.GetProvider()

	// Dump src store complex values (private) into newly created store.
	// Commonly this includes a session copy
	for _, v := range s.providers {
		v.DumpPrivateVars(store)
	}
	return store
}

// StoreProvider func
func (s *Store) StoreProvider(name string, provider base.IProvider) {
	s.providers[name] = provider
}

// GetProvider func
func (s *Store) GetProvider(providerName string) (base.IProvider, error) {
	if provider, exists := s.providers[providerName]; exists {
		return provider, nil
	}
	return nil, fmt.Errorf("unkown provider")
}

// ExistsProvider func
func (s *Store) ExistsProvider(providerName string) bool {
	_, exists := s.providers[providerName]
	return exists
}

// GetLogger func
func (s *Store) GetLogger() base.ILogger {
	return s.logger
}

// SetLogger func
func (s *Store) SetLogger(logger base.ILogger) {
	s.logger = logger
}

// GetPrivateVar func
func (s *Store) GetPrivateVar(varname string) interface{} {
	if value, exists := s.private[varname]; exists {
		return value
	}
	return nil
}

// SetPrivateVar func
func (s *Store) SetPrivateVar(varname string, value interface{}) {
	s.private[varname] = value
}

// GetProvider func
func (s *Store) GetByRefName(refname string) (*base.StorageRecord, error) {
	if record, exists := s.recordsByRefName[refname]; exists {
		return record, nil
	}
	return nil, fmt.Errorf("unkown reference")
}

// GetProvider func
func (s *Store) GetByValueID(valueID string, providerPrefix string) (*base.StorageRecord, error) {
	if record, exists := s.recordsByValueID[providerPrefix+valueID]; exists {
		return record, nil
	}
	return nil, fmt.Errorf("unkown reference")
}

// GetActionOutputByActionID func
func (s *Store) GetActionOutputByActionID(actionID *string) (*base.ActionOutput, error) {
	if aout, exists := s.aoutByActionID[*actionID]; exists {
		return aout, nil
	}
	return nil, fmt.Errorf("output action by id does not exists")
}

// Insert func
func (s *Store) Insert(record *base.StorageRecord, providerPrefix string) error {
	if record.Action != nil {
		s.recordsByActionID[record.Action.ActionID] = record
		if record.Aout != nil {
			s.aoutByActionID[record.Action.ActionID] = record.Aout
		}
	}
	if len(record.ValueID) > 0 {
		s.recordsByValueID[providerPrefix+record.ValueID] = record
	}
	if len(record.RefName) > 0 {
		s.recordsByRefName[record.RefName] = record
	}
	return record.BuildInternals()
}

// can be called ReferenceInterpolation? maybe InterpolateReferences?
func (s *Store) Interpolate(sourcetext *string) error {
	if sourcetext == nil {
		return nil
	}
	// var finaltext string = sourcetext
	r := regexp.MustCompile(`{{([^{}]*)}}`)
	matches := r.FindAllStringSubmatch(*sourcetext, -1)
	// A string like "12345" instead "{{REFERENCE.id}}" can be used with this
	// function so an string without refereces is still valid
	if len(matches) <= 0 {
		return nil
	}

	for _, match := range matches {
		// match[0] == "{{ a.b.c }}""
		// match[1] == " a.b.c "
		var refpath string = strings.TrimSpace(match[1])
		var refname = ""

		// Catch AWS_EC2 from AWS_EC2.foo.bar or AWS_EC2[0]
		_m := regexp.MustCompile(`(?:\\.|[^.[|\\]+)+`).FindAllStringSubmatch(refpath, -1)
		if len(_m) <= 0 {
			return fmt.Errorf("cannot determine reference")
		}

		// obtain record in db referenced by refname
		// _m ->[[AWS_EC2] [networkInterfaceSet] [0]] ...]
		refname = _m[0][0]
		if strings.ToLower(refname) == "env" {
			refpath = strings.TrimPrefix(refpath, refname)
			refpath = strings.TrimPrefix(refpath, ".")
			if len(refpath) <= 0 {
				return fmt.Errorf("environment var access with empty var name")
			}
			varval, exists := os.LookupEnv(strings.TrimSpace(refpath))
			if !exists {
				return fmt.Errorf("environment var access with empty var name")
			}
			if varval == "" {
				s.logger.LogWarn("Interpolation results in an empty string replacement for " + match[0])
			}
			*sourcetext = strings.Replace(*sourcetext, match[0], varval, 1)
			continue
		}

		if strings.ToLower(refname) == "runtime" {
			refpath = strings.TrimPrefix(refpath, refname)
			refpath = strings.TrimPrefix(refpath, ".")
			if len(refpath) <= 0 {
				return fmt.Errorf("runtime var access with empty var name")
			}
			switch strings.ToLower(refpath) {
			case "os":
				*sourcetext = strings.Replace(*sourcetext, match[0], runtime.GOOS, 1)
			case "arch":
				*sourcetext = strings.Replace(*sourcetext, match[0], runtime.GOARCH, 1)
			case "numcpu":
				*sourcetext = strings.Replace(*sourcetext, match[0], strconv.Itoa(runtime.NumCPU()), 1)
			case "version":
				*sourcetext = strings.Replace(*sourcetext, match[0], config.Version, 1)
			case "versiondate":
				*sourcetext = strings.Replace(*sourcetext, match[0], config.VersionDate, 1)
			default:
				return fmt.Errorf("Unknown runtime var name " + refpath)
			}
			continue
		}

		record, exists := s.recordsByRefName[refname]
		if !exists {
			return fmt.Errorf("var reference " + refname + " does not exists (ES1)")
		}

		// refpath -> .foo.bar or [foo.bar] or empty if no path provided
		refpath = strings.TrimPrefix(refpath, refname)
		if len(refpath) <= 0 {
			if reflect.ValueOf(record.Value).Kind() == reflect.String {
				if record.Value.(string) == "" {
					s.logger.LogWarn("Interpolation results in an empty string replacement for " + match[0])
				}
				*sourcetext = strings.Replace(*sourcetext, match[0], record.Value.(string), 1)
			} else {
				// return json by default
				if string(record.JSONValue) == "" {
					s.logger.LogWarn("Interpolation results in an empty string replacement for " + match[0])
				}
				*sourcetext = strings.Replace(*sourcetext, match[0], string(record.JSONValue), 1)
			}
			continue
		}
		// add root char ($) to initial refpath,
		// this replaces AWS_EC2.foo.bar by $.foo.bar
		refpath = "$" + refpath

		r := regexp.MustCompile(`(?:\\.|"(.*?)"|[^|\\]+)+`)
		jpaths := r.FindAllStringSubmatch(refpath, -1)
		if len(jpaths) <= 0 {
			continue
		}

		var jpathTargetValue []byte = record.JSONValue
		for _, jpathm := range jpaths {
			jpath := jpathm[0]
			if strings.ToLower(jpath) == "$.__haserror" {
				if record.Fail {
					jpathTargetValue = []byte("true")
				} else {
					jpathTargetValue = []byte("false")
				}
			} else if strings.ToLower(jpath) == "$.__error" {
				jpathTargetValue = []byte(record.ErrorStr)
			} else if strings.ToLower(jpath) == "$.__internal" {
				jpathTargetValue = []byte(fmt.Sprintf("%v", record.Value))
			} else if strings.ToLower(jpath) == "$.__plain" {
				jpathTargetValue = []byte(fmt.Sprintf("%v", record.PlainValue))
			} else if strings.ToLower(jpath) == "$.__json" {
				jpathTargetValue = record.JSONValue
			} else if strings.ToLower(jpath) == "$.id" {
				if len(record.ValueID) <= 0 {
					return fmt.Errorf("var reference " + refname + " has no ID (ES2)")
				}
				jpathTargetValue = []byte(record.ValueID)
			} else if strings.HasPrefix(jpath, "$.__plain.") {
				attr, exists := record.PlainValue[jpath[1:]]
				if !exists {
					availPaths := fmt.Sprintf("%v", record.PlainValue)
					return fmt.Errorf("path " + jpath[1:] + " does not exists (ES3). Available paths: " + availPaths)
				}
				if attr.IsString {
					jpathTargetValue = []byte(attr.Value.(string))
				} else {
					jpathTargetValue = []byte(fmt.Sprintf("%v", attr.Value))
				}
			} else {
				enc, err := jsonslice.Get(jpathTargetValue, strings.TrimSpace(jpath))
				if err != nil {
					return fmt.Errorf("Invalid path " + jpath + " " + err.Error())
				}
				val := string(enc)
				if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
					var str string
					err = json.Unmarshal(enc, &str)
					if err != nil {
						return fmt.Errorf(err.Error() + ": `" + string(enc) + "`")
					}
					jpathTargetValue = []byte(str)
				} else if len(enc) <= 0 {
					// err?
				} else {
					var prettyJSON bytes.Buffer
					err = json.Indent(&prettyJSON, enc, "", "    ")
					if err != nil {
						return fmt.Errorf(err.Error() + ": `" + string(enc) + "`")
					}
					jpathTargetValue = prettyJSON.Bytes()
				}
			}
		}
		if string(jpathTargetValue) == "" {
			s.logger.LogWarn("Interpolation results in an empty string replacement for " + match[0])
		}
		*sourcetext = strings.Replace(*sourcetext, match[0], string(jpathTargetValue), 1)
	}
	return nil
}

func (s *Store) DeepInterpolation(v interface{}) error {
	return s.recursiveInterpolation(reflect.ValueOf(v), make(map[interface{}]bool))
}

func (s *Store) recursiveInterpolation(v reflect.Value, il map[interface{}]bool) error {
	// While v is a pointer or interface, keep calling v.Elem() to finally get
	// value that pointer point to or value inside interface
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr {
			if _, exists := il[v.Interface()]; exists {
				// Infinite loop throught pointer detected. Skip.
				return nil
			}
			// Prevent nil pointers
			if !v.IsNil() {
				il[v.Interface()] = true
			}
		}
		p := v
		v = v.Elem()
		if v.Kind() == reflect.String {
			ifce := p.Interface()
			txtp, isStringPointer := ifce.(*string)
			if !isStringPointer {
				return fmt.Errorf("internal error looking for string pointer")
			}
			err := s.Interpolate(txtp)
			if err != nil {
				return err
			}
			return nil
		}
	}

	switch v.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.Slice, reflect.Array:
		// Iterate over slice/array
		for i := 0; i < v.Len(); i++ {
			// Recursive call for every array item
			err := s.recursiveInterpolation(v.Index(i), il)
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		t := v.Type()
		// Iterate over struct fields
		for i := 0; i < t.NumField(); i++ {
			// skip unexported fields (first letter lowercase)
			if unicode.IsLower([]rune(t.Field(i).Name)[0]) {
				continue
			}
			// skipt attrs starting by _
			if t.Field(i).Name == "_" {
				continue
			}
			// Recursive call for every struct element
			err := s.recursiveInterpolation(v.Field(i), il)
			if err != nil {
				return err
			}
		}
	case reflect.String:
		return fmt.Errorf("unhandled deep interpolation string")

	}
	return nil
}

// GetPlain func
func (s *Store) GetPlain() (map[string]string, error) {
	result := make(map[string]string)
	for refname, sr := range s.recordsByRefName {
		if reflect.ValueOf(sr.Value).Kind() == reflect.String {
			result[refname+".__json"] = sr.Value.(string)
			result[refname] = sr.Value.(string)
		} else {
			result[refname+".__json"] = string(sr.JSONValue)
			result[refname] = string(sr.JSONValue)
		}
		if sr.Fail {
			result[refname+".__haserror"] = "true"
		} else {
			result[refname+".__haserror"] = "false"
		}
		for path, attr := range sr.PlainValue {
			spath := strings.TrimPrefix(path, ".__plain")
			if spath == refname || len(spath) <= 0 {
				continue
			}
			if enc, err := jsonslice.Get(sr.JSONValue, "$"+spath); err == nil {
				if len(enc) <= 0 {
					result[refname+spath] = attr.String()
					continue
				}
				val := string(enc)
				if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
					val = strings.TrimSuffix(val, "\"")
					val = strings.TrimPrefix(val, "\"")
					enc = []byte(val)
				}
				result[refname+spath] = string(enc)
			} else {
				result[refname+spath] = attr.String()
			}
		}
	}
	return result, nil
}

func (s *Store) GetRawJSONValues() (map[string]json.RawMessage, error) {
	result := make(map[string]json.RawMessage)
	for refname, sr := range s.recordsByRefName {
		if sr.IsString {
			enc, err := json.Marshal(string(sr.JSONValue))
			if err != nil {
				return nil, err
			}
			result[refname] = enc
			continue
		}
		result[refname] = sr.JSONValue
	}
	return result, nil
}

func (s *Store) DumpValuesToShellFile() (*os.File, error) {
	f, err := os.CreateTemp("", "nebulantshellvars.*")
	if err != nil {
		return nil, err
	}
	vars, err := s.GetPlain()
	if err != nil {
		return nil, err
	}
	if _, err := f.Write([]byte("# nebulant shell vars\ndeclare -A NEBULANT\n")); err != nil {
		if err2 := f.Close(); err2 != nil {
			return nil, fmt.Errorf(err.Error() + " " + err2.Error())
		}
		return nil, err
	}
	for path, value := range vars {
		if _, err := f.Write([]byte("NEBULANT[" + path + "]=$( cat <<EOF\n" + value + "\nEOF\n)\n")); err != nil {
			if err2 := f.Close(); err2 != nil {
				return nil, fmt.Errorf(err.Error() + " " + err2.Error())
			}
			return nil, err
		}
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	if err := os.Chmod(f.Name(), 0600); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Store) DumpValuesToJSONFile() (*os.File, error) {
	f, err := os.CreateTemp("", "nebulantjsonvars.*")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vars, err := s.GetRawJSONValues()
	if err != nil {
		return nil, err
	}
	enc, err := json.MarshalIndent(vars, "", "    ")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(enc); err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return f, nil
}
