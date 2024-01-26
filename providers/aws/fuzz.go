// Nebulant
// Copyright (C) 2021  Develatio Technologies S.L.

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

package aws

import (
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/runtime"
)

// Fuzz func
// TODO: #error Currently only Linux is supported. More info: issue #2
func Fuzz(data []byte) int {
	bp, err := blueprint.NewFromBytes(data)
	if err != nil {
		if bp != nil {
			panic("bp != nil on error")
		}
		return 0
	}

	irb, err := blueprint.GenerateIRB(bp, &blueprint.IRBGenConfig{})
	if err != nil {
		panic(err)
	}
	rt := runtime.NewRuntime(irb)
	actx := rt.NewAContext(nil, &bp.Actions[0])
	aws := new(Provider)
	aout, aerr := aws.HandleAction(actx)
	if aerr != nil {
		if aout != nil {
			panic("bp != nil on error")
		}
		return 0
	}
	return 1
}
