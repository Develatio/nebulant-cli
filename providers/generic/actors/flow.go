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

package actors

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
	"github.com/joho/godotenv"
)

type defineVarsParameters struct {
	Vars  map[string]string `json:"vars"`
	Files []string          `json:"files"`
}

type sleepParameters struct {
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
	var err error
	var rawA = c.Field
	var rawB = c.Value
	err = c.ctx.Store.Interpolate(&rawA)
	if err != nil {
		return false, err
	}
	err = c.ctx.Store.Interpolate(&rawB)
	if err != nil {
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
	if strings.HasPrefix(rawA, "\"") && strings.HasSuffix(rawA, "\"") {
		rawA = strings.TrimSuffix(rawA, "\"")
		rawA = strings.TrimPrefix(rawA, "\"")
	}
	if strings.HasPrefix(rawA, "'") && strings.HasSuffix(rawA, "'") {
		rawA = strings.TrimSuffix(rawA, "'")
		rawA = strings.TrimPrefix(rawA, "'")
	}
	if strings.HasPrefix(rawB, "\"") && strings.HasSuffix(rawB, "\"") {
		rawB = strings.TrimSuffix(rawB, "\"")
		rawB = strings.TrimPrefix(rawB, "\"")
	}
	if strings.HasPrefix(rawB, "'") && strings.HasSuffix(rawB, "'") {
		rawB = strings.TrimSuffix(rawB, "'")
		rawB = strings.TrimPrefix(rawB, "'")
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
	Content *string `json:"content" validate:"required"`
}

func NOOP(ctx *ActionContext) (*base.ActionOutput, error) {
	return nil, nil
}

func Start(ctx *ActionContext) (*base.ActionOutput, error) {
	return nil, nil
}

func Stop(ctx *ActionContext) (*base.ActionOutput, error) {
	return nil, nil
}

// ConditionParse func
func ConditionParse(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	params := new(conditionParameters)
	err = util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if err != nil {
		return nil, err
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
	params := new(sleepParameters)
	jsonErr := util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if jsonErr != nil {
		return nil, jsonErr
	}
	ctx.Logger.LogInfo("Sleeping for " + strconv.FormatInt(params.Seconds, 10) + " seconds")
	// Duration == type int64
	time.Sleep(time.Duration(params.Seconds) * time.Second)
	return nil, nil
}

// OKKO func
func OKKO(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(okkoParameters)
	jsonErr := util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if jsonErr != nil {
		return nil, jsonErr
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
	err = util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if err != nil {
		return nil, err
	}

	err = ctx.Store.Interpolate(params.Content)
	if err != nil {
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
		panic(&util.PanicData{
			PanicValue: "Halt for unknown reason. Bad halt params. Halting anyway.",
			PanicTrace: []byte("--"),
		})
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
	jsonErr := util.UnmarshalValidJSON(ctx.Action.Parameters, params)
	if jsonErr != nil {
		return nil, jsonErr
	}

	for varname := range params.Vars {
		varvalue := params.Vars[varname]
		ctx.Logger.LogInfo("Setting var " + varname)
		err := ctx.Store.Interpolate(&varvalue)
		if err != nil {
			return nil, err
		}
		varvalue = strings.Replace(varvalue, "\\{", "{", -1)
		varvalue = strings.Replace(varvalue, "\\}", "}", -1)
		err = ctx.Store.Insert(&base.StorageRecord{
			RefName: varname,
			Aout:    nil,
			Value:   varvalue,
			Action:  ctx.Action,
		}, ctx.Action.Provider)
		if err != nil {
			return nil, err
		}
	}

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
