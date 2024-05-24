// MIT License
//
// Copyright (C) 2022  Develatio Technologies S.L.

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

package assets

import (
	"net/url"
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
		Filters: []AssetDefinitionFilter{func(v interface{}, terms url.Values) bool {
			region := terms.Get("region")
			if region == "" {
				return true
			}
			if id, exists := v.(*CustomAwsImage).ImageIds[region]; exists {
				v.(*CustomAwsImage).ImageId = &id
				return true
			}
			return false
		}},
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
}
