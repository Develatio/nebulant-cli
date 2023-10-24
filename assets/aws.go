// Nebulant
// Copyright (C) 2022  Develatio Technologies S.L.

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

package assets

import (
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/develatio/nebulant-cli/config"
)

type CustomAwsImage struct {
	Architecture       *string           `json:"Architecture"`
	CreationDate       *string           `json:"CreationDate"`
	Description        *string           `json:"Description"`
	EnaSupport         *bool             `json:"EnaSupport"`
	Hypervisor         *string           `json:"Hypervisor"`
	ImageType          *string           `json:"ImageType"`
	Name               *string           `json:"Name"`
	OwnerId            *string           `json:"OwnerId"`
	PlatformDetails    *string           `json:"PlatformDetails"`
	VirtualizationType *string           `json:"VirtualizationType"`
	ImageIds           map[string]string `json:"ImageIds"`
	ImageId            *string           `json:"ImageId"`
}

func init() {
	// base aws images
	AssetsDefinition["aws/images"] = &AssetDefinition{
		Name:         "AWS Images",
		IndexPath:    filepath.Join(config.AppHomePath(), "assets", "aws_images.idx"),
		SubIndexPath: filepath.Join(config.AppHomePath(), "assets", "aws_images.subidx"),
		FilePath:     filepath.Join(config.AppHomePath(), "assets", "aws_images.asset"),
		FreshItem:    func() interface{} { return &CustomAwsImage{} },
		MarshallIndentItem: func(v interface{}) string {
			return awsutil.Prettify(v)
		},
		Filter: func(v interface{}) bool {
			// filter sample:
			// return *v.(*CustomAwsImage).Architecture == "x86_64"
			return true
		},
		LookPath: []string{
			"$.Architecture",
			"$.Name",
			"$.Description",
			"$.BlockDeviceMappings[].Ebs.VolumeType",
			"$.BlockDeviceMappings[].Ebs.SnapshotId",
			"$.ImageId",
			"$.ImageLocation",
			"$.OwnerId",
		},
		Alias: [][]string{
			{"x64", "x86_64"},
		},
	}

	// copy base aws images
	_a := *AssetsDefinition["aws/images"]
	AssetsDefinition["aws/us-east-1/images"] = &_a
	// filter by region
	AssetsDefinition["aws/us-east-1/images"].Filter = func(v interface{}) bool {
		if id, exists := v.(*CustomAwsImage).ImageIds["us-east-1"]; exists {
			v.(*CustomAwsImage).ImageId = &id
			return true
		}
		return false
	}
}
