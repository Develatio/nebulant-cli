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

package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unicode"

	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/develatio/nebulant-cli/blueprint"
)

// compressStruct func
func compressStruct(path string, v reflect.Value, cs map[string]*AttrTreeValue, il map[interface{}]bool) {
	// While v is a pointer or interface, keep calling v.Elem() to finally get
	// value that pointer point to or value inside interface
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr {
			if _, exists := il[v.Interface()]; exists {
				// Infinite loop throught pointer detected. Skip.
				return
			}
			// Prevent store nil pointers
			if !v.IsNil() {
				il[v.Interface()] = true
			}
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Invalid:
		return
	case reflect.Slice, reflect.Array:
		cs[path] = &AttrTreeValue{
			IsString:    false,
			Description: "[" + v.Kind().String() + "]",
			Value:       v.Interface(),
		}
		// Iterate over slice/array
		for i := 0; i < v.Len(); i++ {
			// Recursive call, append array index to path
			compressStruct(path+"["+strconv.Itoa(i)+"]", v.Index(i), cs, il)
		}
	case reflect.Struct:
		// Prevent store root object
		if path != "__plain" {
			cs[path] = &AttrTreeValue{
				IsString:    false,
				Value:       v.Interface(),
				Description: "[" + v.Kind().String() + "]",
			}
		}

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

			// Look for tag with "json/xml/notGoStruct" name
			// this is the human-readable field name
			var mt string
			marshalTagName := ""

			// AWS
			mt = t.Field(i).Tag.Get("locationName") // aws
			if mt != "" {
				marshalTagName = "locationName"
			}

			// Azure or normal JSON struct
			if mt == "" {
				mt = t.Field(i).Tag.Get("json") // azure?
				if mt != "" {
					marshalTagName = "json"
				}
			}

			// Azure xml mode
			if mt == "" {
				mt = t.Field(i).Tag.Get("xml") // azure
				if mt != "" {
					marshalTagName = "xml"
				}
			}

			// fallback, get the raw field name
			if mt == "" {
				mt = t.Field(i).Name
			}

			_, detectedTagName := cs["__marshalTagName"]
			if !detectedTagName && marshalTagName != "" {
				cs["__marshalTagName"] = &AttrTreeValue{
					IsString:    true,
					Value:       marshalTagName,
					Description: "the detected marshall tag name",
				}
			}

			// Recursive call, append field name to path
			compressStruct(path+"."+mt, v.Field(i), cs, il)
		}
	default:
		// Dump final value to path
		cs[path] = &AttrTreeValue{
			IsString:    true,
			Value:       fmt.Sprintf("%v", v.Interface()),
			Description: "String value",
		}
	}
}

// AttrTreeValue struct
type AttrTreeValue struct {
	IsString    bool
	Value       interface{}
	Description string
}

// String func
func (a *AttrTreeValue) String() string {
	if a.IsString {
		return fmt.Sprintf("%v", a.Value)
	}
	return a.Description
}

type StorageRecordStack struct {
	Items []interface{}
}

// StorageRecord struct
type StorageRecord struct {
	ValueID    string                    `json:"valueID"`
	RefName    string                    `json:"refName"`
	Aout       *ActionOutput             `json:"-"`
	Action     *blueprint.Action         `json:"-"`
	RawSource  interface{}               `json:"-"`
	Value      interface{}               `json:"value"`
	PlainValue map[string]*AttrTreeValue `json:"-"`
	JSONValue  []byte                    `json:"rawjson"`
	IsString   bool                      `json:"-"`
	Fail       bool                      `json:"fail"`
	Error      error                     `json:"-"`
	ErrorStr   string                    `json:"error"`
}

func (sr *StorageRecord) BuildInternals() error {
	// discard empty ref name
	if len(sr.RefName) <= 0 {
		return nil
	}
	if sr.Error != nil {
		sr.ErrorStr = sr.Error.Error()
	}

	vof := reflect.ValueOf(sr.Value)
	if vof.Kind() == reflect.Ptr {
		vof = vof.Elem()
	}

	switch sr.Value.(type) {
	case nil:
		cs := make(map[string]*AttrTreeValue)
		sr.JSONValue = []byte(nil)
		sr.IsString = false
		sr.PlainValue = cs
		return nil
	}

	if vof.Kind() == reflect.String {
		cs := make(map[string]*AttrTreeValue)
		sr.JSONValue = []byte(sr.Value.(string))
		sr.IsString = true
		sr.PlainValue = cs
	} else if vof.Kind() == reflect.Struct {
		var err error
		var enc []byte
		var marshallTagName string = ""
		cs := make(map[string]*AttrTreeValue)

		compressStruct(".__plain", vof, cs, make(map[interface{}]bool))

		if _, exists := cs["__marshalTagName"]; exists {
			marshallTagName = cs["__marshalTagName"].Value.(string)
		}

		if marshallTagName == "locationName" {
			// aws built in marshal
			enc, err = jsonutil.BuildJSON(sr.Value)
			if err == nil {
				var prettyJSON bytes.Buffer
				err = json.Indent(&prettyJSON, enc, "", "    ")
				enc = prettyJSON.Bytes()
			}
		} else if marshallTagName == "json" || marshallTagName == "" {
			// generic json marshall
			enc, err = json.MarshalIndent(sr.Value, "", "    ")
		} else {
			// unsuported tag
			return fmt.Errorf("Unsuperted marshal tag " + marshallTagName)
		}

		if err == nil {
			sr.JSONValue = enc
		} else {
			return err
		}
		sr.PlainValue = cs

	} else {
		return fmt.Errorf("Invalid " + sr.RefName + " output [" + vof.Kind().String() + "]")
	}

	return nil
}

// IStore interface
type IStore interface {
	StoreProvider(name string, provider IProvider)
	GetProvider(providerName string) (IProvider, error)
	Duplicate() IStore
	ExistsProvider(providerName string) bool
	GetLogger() ILogger
	SetLogger(logger ILogger)
	GetPrivateVar(varname string) interface{}
	SetPrivateVar(varname string, value interface{})
	Merge(IStore)
	GetActionOutputByActionID(actionID *string) (*ActionOutput, error)
	Insert(record *StorageRecord, providerPrefix string) error
	Interpolate(sourcetext *string) error
	GetPlain() (map[string]string, error)
	GetRawJSONValues() (map[string]json.RawMessage, error)
	DumpValuesToShellFile() (*os.File, error)
	DumpValuesToJSONFile() (*os.File, error)
	GetByRefName(refname string) (*StorageRecord, error)
	DeepInterpolation(v interface{}) error
	ExistsRefName(refname string) bool
}
