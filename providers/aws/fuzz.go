// MIT License
//
// Copyright (C) 2021  Develatio Technologies S.L.

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

package aws

// import (
// 	"github.com/develatio/nebulant-cli/blueprint"
// 	"github.com/develatio/nebulant-cli/runtime"
// )

// // Fuzz func
// // TODO: #error Currently only Linux is supported. More info: issue #2
// func Fuzz(data []byte) int {
// 	bp, err := blueprint.NewFromBytes(data)
// 	if err != nil {
// 		if bp != nil {
// 			panic("bp != nil on error")
// 		}
// 		return 0
// 	}

// 	irb, err := blueprint.GenerateIRB(bp, &blueprint.IRBGenConfig{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	rt := runtime.NewRuntime(irb)
// 	actx := rt.NewAContext(nil, &bp.Actions[0])
// 	aws := new(Provider)
// 	aout, aerr := aws.HandleAction(actx)
// 	if aerr != nil {
// 		if aout != nil {
// 			panic("bp != nil on error")
// 		}
// 		return 0
// 	}
// 	return 1
// }
