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
	"github.com/develatio/nebulant-cli/blueprint"
)

// ProviderInitFunc type
type ProviderInitFunc func(store IStore) (IProvider, error)

// IProvider interface
type IProvider interface {
	HandleAction(action *blueprint.Action) (*ActionOutput, error)
	DumpPrivateVars(freshStore IStore)
}

type ProviderAuthError struct {
	Err error
}

func (r *ProviderAuthError) Error() string {
	return r.Err.Error()
}
