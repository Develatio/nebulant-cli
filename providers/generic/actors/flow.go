// MIT License
//
// Copyright (C) 2020  Develatio Technologies S.L.

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

package actors

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/util"
	"github.com/joho/godotenv"
)

type VarType string

const (
	VarTypeString             VarType = "string"
	VarTypeBool               VarType = "boolean"
	VarTypeInt                VarType = "int"
	VarTypeSelectableStatic   VarType = "selectable-static-values"
	VarTypeSelectableVariable VarType = "selectable-variables"
)

type defineVarsParametersVarOptions struct {
	Label string `json:"label" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type defineVarsParametersVar struct {
	AskAtRuntime *bool                            `json:"ask_at_runtime"`
	Key          string                           `json:"key" validate:"required"`
	Value        interface{}                      `json:"value"`
	Type         *VarType                         `json:"type"`
	Options      []defineVarsParametersVarOptions `json:"options"`
	Required     bool                             `json:"required"`
	Stack        *bool                            `json:"stack"`
}

func (d *defineVarsParametersVar) askForValue() error {
	// reflect.ValueOf(d.Value).IsNil()
	v := reflect.ValueOf(d.Value)
	isEmpty := v.Kind() == reflect.Ptr && v.IsNil()
	isNotValid := !v.IsValid()
	if d.AskAtRuntime != nil && *d.AskAtRuntime {
		switch *d.Type {
		case VarTypeString:
			var def string
			if !isEmpty && !isNotValid {
				def = d.Value.(string)
			}
			vcn, err := cast.PromptInput("Please, enter value for "+d.Key, d.Required, def)
			if err != nil {
				return err
			}
			val := <-vcn
			d.Value = val
		case VarTypeInt:
			var vv int
			if !isEmpty && !isNotValid {
				vv = d.Value.(int)
			}
			vcn, err := cast.PromptInt("Please, enter value for "+d.Key, d.Required, string(vv))
			if err != nil {
				return err
			}
			val := <-vcn
			d.Value = val
			v, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			d.Value = v
		case VarTypeSelectableStatic:
			// TODO: test if there is value
			// selected and mark option acordingly
			options := make(map[string]string)
			for _, obj := range d.Options {
				options[obj.Value] = obj.Label
			}
			vcn, err := cast.PromptSelect("Please, enter value for "+d.Key, d.Required, options)
			if err != nil {
				return err
			}
			val := <-vcn
			d.Value = val
		case VarTypeBool:
			def := "false"
			if !isEmpty && !isNotValid {
				vv := d.Value.(bool)
				if vv {
					def = "true"
				}
			}
			vcn, err := cast.PromptBool("Please, enter value for "+d.Key, d.Required, def)
			if err != nil {
				return err
			}
			val := <-vcn
			d.Value = val
		case VarTypeSelectableVariable:
			return fmt.Errorf("var type not supported yet")
		default:
			return fmt.Errorf("unknown var type")
		}
	}
	return nil
}

type defineVarsParameters struct {
	Vars  []*defineVarsParametersVar `json:"vars"`
	Files []string                   `json:"files"`
}

type SleepParameters struct {
	Seconds int64 `json:"seconds" validate:"required"`
}

type okkoParameters struct {
	Ok    bool    `json:"ok"`
	KoMsg *string `json:"komsg"`
	OkMsg *string `json:"okmsg"`
}

// Condition struct
type Condition struct {
	ctx        *ActionContext
	ID         string       `json:"id"`
	Field      string       `json:"field"`
	Operator   string       `json:"operator"`
	Value      string       `json:"value"`
	Rules      []*Condition `json:"rules"`
	Combinator string       `json:"combinator"`
	//
	Not bool `json:"not"`
}

func (c *Condition) evaluate() (bool, error) {
	c.ctx.Logger.LogDebug("Evaluating " + c.ID)
	if len(c.Rules) > 0 {
		if c.Combinator == "and" {
			return c.operate(true)
		}
		if c.Combinator == "or" {
			return c.operate(false)
		}
		return false, fmt.Errorf("unknown combinator")
	}

	var rawA = c.Field
	var rawB = c.Value

	if err := c.ctx.Store.Interpolate(&rawA); err != nil {
		return false, err
	}
	if err := c.ctx.Store.Interpolate(&rawB); err != nil {
		return false, err
	}

	// Int
	// Valid (int)a and (int)b
	// a and b should be int
	if intA, err := strconv.ParseInt(rawA, 10, 64); err == nil {
		if intB, err := strconv.ParseInt(rawB, 10, 64); err == nil {
			c.ctx.Logger.LogDebug("evaluate as int")
			switch c.Operator {
			case "=":
				return intA == intB, nil
			case "!=":
				return intA != intB, nil
			case ">":
				return intA > intB, nil
			case "<":
				return intA < intB, nil
			case "<=":
				return intA <= intB, nil
			case ">=":
				return intA >= intB, nil
			default:
				return false, fmt.Errorf("unknown operator for int types")
			}
		}
	}

	// Float
	// Invalid (int)a or (int)b, Valid (float)a and (float)b
	// one of both could be int
	if floatA, err := strconv.ParseFloat(rawA, 64); err == nil {
		if floatB, err := strconv.ParseFloat(rawB, 64); err == nil {
			c.ctx.Logger.LogDebug("evaluate as float")
			switch c.Operator {
			case "=":
				return floatA == floatB, nil
			case "!=":
				return floatA != floatB, nil
			case ">":
				return floatA > floatB, nil
			case "<":
				return floatA < floatB, nil
			case "<=":
				return floatA <= floatB, nil
			case ">=":
				return floatA >= floatB, nil
			default:
				return false, fmt.Errorf("unknown operator for float types")
			}
		}
	}

	// Bool
	// Accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False
	if boolA, err := strconv.ParseBool(rawA); err == nil {
		if boolB, err := strconv.ParseBool(rawB); err == nil {
			c.ctx.Logger.LogDebug("evaluate as bool")
			switch c.Operator {
			case "=":
				return boolA == boolB, nil
			case "!=":
			default:
				return false, fmt.Errorf("unknown operator for bool types")
			}
		}
	}

	// String
	// Invalid int, float and bool
	c.ctx.Logger.LogDebug("evaluate as string")
	if len(rawA) > 1 {
		if strings.HasPrefix(rawA, "\"") && strings.HasSuffix(rawA, "\"") {
			rawA = strings.TrimSuffix(rawA, "\"")
			rawA = strings.TrimPrefix(rawA, "\"")
		}
		if strings.HasPrefix(rawA, "'") && strings.HasSuffix(rawA, "'") {
			rawA = strings.TrimSuffix(rawA, "'")
			rawA = strings.TrimPrefix(rawA, "'")
		}
	}

	if len(rawB) > 1 {
		if strings.HasPrefix(rawB, "\"") && strings.HasSuffix(rawB, "\"") {
			rawB = strings.TrimSuffix(rawB, "\"")
			rawB = strings.TrimPrefix(rawB, "\"")
		}
		if strings.HasPrefix(rawB, "'") && strings.HasSuffix(rawB, "'") {
			rawB = strings.TrimSuffix(rawB, "'")
			rawB = strings.TrimPrefix(rawB, "'")
		}
	}

	switch c.Operator {
	case "=":
		c.ctx.Logger.LogDebug("Evaluating as " + rawA + " == " + rawB)
		return rawA == rawB, nil
	case "!=":
		c.ctx.Logger.LogDebug("Evaluating as " + rawA + " != " + rawB)
		return rawA != rawB, nil
	case ">":
		return utf8.RuneCountInString(rawA) > utf8.RuneCountInString(rawB), nil
	case "<":
		return utf8.RuneCountInString(rawA) < utf8.RuneCountInString(rawB), nil
	case "<=":
		return utf8.RuneCountInString(rawA) <= utf8.RuneCountInString(rawB), nil
	case ">=":
		return utf8.RuneCountInString(rawA) >= utf8.RuneCountInString(rawB), nil
	case "contains":
		c.ctx.Logger.LogDebug("test if '" + rawB + "' is into '" + rawA + "'")
		return strings.Contains(rawA, rawB), nil
	case "beginsWith":
		return strings.HasPrefix(rawA, rawB), nil
	case "endsWith":
		return strings.HasSuffix(rawA, rawB), nil
	case "doesNotContain":
		return !strings.Contains(rawA, rawB), nil
	case "doesNotBeginWith":
		return !strings.HasPrefix(rawA, rawB), nil
	case "doesNotEndWith":
		return !strings.HasSuffix(rawA, rawB), nil
	}

	return false, fmt.Errorf("unknown operator for string types")
}

func (c *Condition) operate(operator bool) (bool, error) {
	for _, condition := range c.Rules {
		condition.ctx = c.ctx
		r, err := condition.evaluate()
		if err != nil {
			c.ctx.Logger.LogDebug("Evaluation error")
			return false, err
		}
		c.ctx.Logger.LogDebug("Condition evaluated as " + strconv.FormatBool(r) + " within operator " + strconv.FormatBool(operator))
		// AND -> true
		// OR -> false
		//
		// Si r es false y el operador es true (AND):
		// if false != true -> entra en el if y sale con return false ya que
		// uno de los operandos no es true
		//
		// Si r es true y el operador es false (OR):
		// if true != false -> entra en el if y sale con return true ya que
		// uno solo de los operandos a true es suficiente
		if r != operator {
			c.ctx.Logger.LogDebug("Group evaluated r!operator " + strconv.FormatBool(r) + "!" + strconv.FormatBool(operator))
			return !operator, nil
		}
	}

	// None rules are met, return operator
	c.ctx.Logger.LogDebug("Group evaluated as " + strconv.FormatBool(operator))
	return operator, nil
}

type conditionParameters struct {
	Conditions *Condition `json:"conditions" validate:"required"`
}

type logParameters struct {
	Content *string `json:"content" validate:"required"`
}

type panicParameters struct {
	Content *string `json:"content"`
}

func NOOP(ctx *ActionContext) (*base.ActionOutput, error) {
	return nil, nil
}

// func Start(ctx *ActionContext) (*base.ActionOutput, error) {
// 	return nil, nil
// }

func Stop(ctx *ActionContext) (*base.ActionOutput, error) {
	return nil, nil
}

// ConditionParse func
func ConditionParse(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	params := new(conditionParameters)
	if err = util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	params.Conditions.ctx = ctx
	r, err := params.Conditions.evaluate()
	if err != nil {
		return nil, err
	}

	aout := base.NewActionOutput(ctx.Action, r, nil)
	return aout, nil
}

// Sleep func
func Sleep(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(SleepParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}
	ctx.Logger.LogInfo("Sleeping for " + strconv.FormatInt(params.Seconds, 10) + " seconds")
	// Duration == type int64
	time.Sleep(time.Duration(params.Seconds) * time.Second)
	return nil, nil
}

// OKKO func
func OKKO(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(okkoParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	if params.Ok {
		if params.OkMsg != nil && *params.OkMsg != "" {
			err := ctx.Store.Interpolate(params.OkMsg)
			if err != nil {
				return nil, err
			}
			ctx.Logger.LogInfo(*params.OkMsg)
		}
		return nil, nil
	}

	if params.KoMsg != nil {
		err := ctx.Store.Interpolate(params.KoMsg)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(*params.KoMsg)
	}
	return nil, fmt.Errorf("")
}

// Log func
func Log(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	params := new(logParameters)
	if err = util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	if err = ctx.Store.Interpolate(params.Content); err != nil {
		return nil, err
	}

	ctx.Logger.LogInfo(*params.Content)
	return nil, nil
}

// Panic func
func Panic(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	params := new(panicParameters)
	err = util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if err != nil {
		if ctx.Rehearsal {
			return nil, err
		}
		panic(&util.PanicData{
			PanicValue: "Halt for unknown reason. Bad halt params. Halting anyway.",
			PanicTrace: []byte("--"),
		})
	}
	if ctx.Rehearsal {
		return nil, nil
	}
	if params.Content == nil {
		panic(&util.PanicData{
			PanicValue: "Halt for unknown reason. Bad halt params. Halting anyway.",
			PanicTrace: []byte("--"),
		})
	}

	if *params.Content == "" {
		panic(&util.PanicData{
			PanicValue: "Halt for unknown reason. Bad halt params. Halting anyway.",
			PanicTrace: []byte("--"),
		})
	}
	ctx.Logger.LogErr(*params.Content)
	panic(&util.PanicData{
		PanicValue: *params.Content,
		PanicTrace: []byte("--"),
	})
}

// DefineVars func
func DefineVars(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(defineVarsParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	// validate var type/value
	for _, v := range params.Vars {
		// dont care here preset value,
		// it will be setted by user at runtime
		if v.AskAtRuntime != nil && *v.AskAtRuntime {
			continue
		}
		if v.Type == nil {
			v.Type = new(VarType)
			*v.Type = VarTypeString
		}
		// test type
		switch *v.Type {
		case VarTypeSelectableStatic, VarTypeSelectableVariable, VarTypeString:
			if _, isString := v.Value.(string); !isString {
				return nil, fmt.Errorf("Var type string mismatch for " + v.Key)
			}
		case VarTypeBool:
			if _, isBool := v.Value.(bool); !isBool {
				return nil, fmt.Errorf("Var type bool mismatch for " + v.Key)
			}
		case VarTypeInt:
			if _, isInt := v.Value.(int); !isInt {
				return nil, fmt.Errorf("Var type int mismatch for " + v.Key)
			}
		default:
			return nil, fmt.Errorf("Unknown vartype for key " + v.Key)
		}
	}

	if ctx.Rehearsal {
		jj, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		ctx.Action.Parameters = jj
		return nil, nil
	}

	for _, v := range params.Vars {
		// ask for value as needed
		for {
			err := v.askForValue()
			if err != nil && err == io.EOF {
				if v.Required {
					switch v.Value.(type) {
					case nil:
						ctx.Logger.LogWarn("var " + v.Key + " is required and cannot be null")
						continue
					case string:
						if v.Value == "" {
							ctx.Logger.LogWarn("var " + v.Key + " is required and cannot be empty")
							continue
						}
					}
				} else {
					switch v.Value.(type) {
					case nil:
						ctx.Logger.LogWarn("var " + v.Key + " is null and NOT required. Be catious")
					case string:
						if v.Value == "" {
							ctx.Logger.LogWarn("var " + v.Key + " is empty and NOT required. Be catious")
						}
					}
					break
				}
			} else if err != nil {
				return nil, err
			}
			break
		}

		var recordvalue interface{}
		varname := v.Key
		ctx.Logger.LogInfo("Setting var " + varname)

		switch v.Value.(type) {
		case string:
			switch *v.Type {
			case VarTypeString:
				varvalue := v.Value.(string)
				err := ctx.Store.Interpolate(&varvalue)
				if err != nil {
					return nil, err
				}
				varvalue = strings.Replace(varvalue, "\\{", "{", -1)
				varvalue = strings.Replace(varvalue, "\\}", "}", -1)
				recordvalue = varvalue
			default:
				recordvalue = v.Value
			}
		case nil:
			recordvalue = nil
			ctx.Logger.LogDebug("var " + v.Key + " is null and cannot be interpolated")
		}

		if v.Stack != nil && *v.Stack {
			// this is an stack var wich can acumulate values
			err := ctx.Store.Push(&base.StorageRecord{
				RefName: varname,
				Aout:    nil,
				Value:   recordvalue,
				// note that literal is not allowed
				Action: ctx.Action,
			}, ctx.Action.Provider)
			if err != nil {
				return nil, err
			}
		} else {
			err := ctx.Store.Insert(&base.StorageRecord{
				RefName: varname,
				Aout:    nil,
				Value:   recordvalue,
				Literal: true,
				Action:  ctx.Action,
			}, ctx.Action.Provider)
			if err != nil {
				return nil, err
			}
		}

	}

	// if params has .Files, we should read those files
	// and store var & values
	for _, file := range params.Files {
		envs, err := godotenv.Read(file)
		if err != nil {
			return nil, err
		}
		for varname := range envs {
			varvalue := envs[varname]
			ctx.Logger.LogInfo("Setting env var " + varname)
			err := ctx.Store.Interpolate(&varvalue)
			if err != nil {
				return nil, err
			}
			varvalue = strings.Replace(varvalue, "\\{", "{", -1)
			varvalue = strings.Replace(varvalue, "\\}", "}", -1)
			err = ctx.Store.Insert(&base.StorageRecord{
				RefName: varname,
				Aout:    nil,
				Literal: true,
				Value:   varvalue,
				Action:  ctx.Action,
			}, ctx.Action.Provider)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
