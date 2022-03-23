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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/config"
)

type CustomInstanceTypeInfo struct {
	InstanceType *string
	Summary      *string
}

func init() {
	AssetsDefinition["aws_instance_types"] = &AssetDefinition{
		IndexPath:    filepath.Join(config.AppHomePath(), "assets", "aws_instance_types.idx"),
		SubIndexPath: filepath.Join(config.AppHomePath(), "assets", "aws_instance_types.subidx"),
		FilePath:     filepath.Join(config.AppHomePath(), "assets", "aws_instance_types.asset"),
		FreshItem:    func() interface{} { return &CustomInstanceTypeInfo{} },
		LookPath: []string{
			"$.InstanceType",
			"$.Summary",
		},
	}
	AssetsDefinition["aws_images"] = &AssetDefinition{
		IndexPath:    filepath.Join(config.AppHomePath(), "assets", "aws_images.idx"),
		SubIndexPath: filepath.Join(config.AppHomePath(), "assets", "aws_images.subidx"),
		FilePath:     filepath.Join(config.AppHomePath(), "assets", "aws_images.asset"),
		FreshItem:    func() interface{} { return &ec2.Image{} },
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
	}
}
