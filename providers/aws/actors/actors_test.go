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
