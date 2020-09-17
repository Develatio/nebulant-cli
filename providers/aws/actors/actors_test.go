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
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/storage"
)

func TestDeepInterpolation(t *testing.T) {

	store := storage.NewStore()
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
	}, "generic")

	text1 := "holi"
	text2 := "holi {{ SINGLE_VAR_NAME }}"
	awsinput := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("cucutras"),
				Values: []*string{
					&text1,
					&text2,
				},
			},
		},
	}

	err := store.DeepInterpolation(awsinput.Filters)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text1 != "holi" {
		t.Errorf("text interpolation failed")
	}
	if text2 != "holi varvalue" {
		t.Errorf("text interpolation failed")
	}
}
