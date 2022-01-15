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

package storage_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/storage"
)

// Provider struct
type testProvider struct {
}

// DumpStore func
func (p *testProvider) DumpStore(freshStore base.IStore) {
}

// HandleAction func
func (p *testProvider) HandleAction(action *blueprint.Action) (*base.ActionOutput, error) {
	return nil, nil
}

type fakeLogger struct{}

func (l *fakeLogger) LogCritical(s string)    {}
func (l *fakeLogger) LogErr(s string)         {}
func (l *fakeLogger) ByteLogErr(b []byte)     {}
func (l *fakeLogger) LogWarn(s string)        {}
func (l *fakeLogger) LogInfo(s string)        {}
func (l *fakeLogger) ByteLogInfo(b []byte)    {}
func (l *fakeLogger) LogDebug(s string)       {}
func (l *fakeLogger) Duplicate() base.ILogger { return l }
func (l *fakeLogger) SetActionID(ai string)   {}

type tsie struct {
	input    string
	expected string
}

func TestInterpolate(t *testing.T) {
	var err error
	lg := &fakeLogger{}
	store := storage.NewStore()
	store.SetLogger(lg)
	ref := "OUTPUT_VAR_NAME"
	action := &blueprint.Action{
		Provider: "generic",
		Output:   &ref,
	}
	demoresult := &ec2.Image{
		ImageId: aws.String("test"),
		Name:    aws.String("test1234"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("tagkey0"),
				Value: aws.String("tagvalue0"),
			},
			{
				Key:   aws.String("tagkey1"),
				Value: aws.String("tagvalue1"),
			},
		},
	}
	aout := base.NewActionOutput(action, demoresult, demoresult.ImageId)
	err = store.Insert(aout.Records[0], action.Provider)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	text := "a test {{ OUTPUT_VAR_NAME.tagSet[1].value }} {{ OUTPUT_VAR_NAME.tagSet[0].value }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "a test tagvalue1 tagvalue0" {
		t.Errorf("text interpolation failed")
	}

	text = "a test {{ OUTPUT_VAR_NAME.tagSet[1].value }} {{ OUTPUT_VAR_NAME.__json }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}

	text = "1234"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "1234" {
		t.Errorf("Interpolate without references fail")
	}

	err = store.Interpolate(nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	text = "{{ ENV. }}"
	err = store.Interpolate(&text)
	if err == nil {
		t.Errorf("malformed env reference should fail")
	}

	text = "{{ ENV.           }}"
	err = store.Interpolate(&text)
	if err == nil {
		t.Errorf("malformed env reference should fail")
	}

	text = "{{ UNDEFINED.UNEFINED }}"
	err = store.Interpolate(&text)
	if err == nil {
		t.Errorf("undefined reference should fail")
	}

	text = "{{ ENV.UNDEFINED }}"
	err = store.Interpolate(&text)
	if err == nil {
		t.Errorf("undefined env should not be allowed")
	}

	text = "{{ OUTPUT_VAR_NAME.__plain.tagSet[1].value }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "tagvalue1" {
		t.Errorf("interpolation fail, got %v, expected tagvalue1", text)
	}

	text = "{{ OUTPUT_VAR_NAME.__plain.undefined }}"
	err = store.Interpolate(&text)
	if err == nil {
		t.Errorf("plain undefined should fail")
	}

	text = "{{ . }}"
	err = store.Interpolate(&text)
	if err == nil {
		t.Errorf("malformed references should fail")
	}

	p, err := store.GetPlain()
	if err != nil {
		t.Errorf(err.Error())
	}
	if p["OUTPUT_VAR_NAME.tagSet[1].value"] != "tagvalue1" {
		t.Errorf("plain should register plain xpaths")
	}

	_, err = store.GetRawJSONValues()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestSingleReferences(t *testing.T) {
	var err error
	store := storage.NewStore()
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
	}, "generic")
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME2",
		Aout:    nil,
		Value:   "varvalue2",
	}, "generic")

	tests := []tsie{
		{
			input:    "a test {{ SINGLE_VAR_NAME }} {{ SINGLE_VAR_NAME2 }}",
			expected: "a test varvalue varvalue2",
		},
		{
			input:    "a test {{ SINGLE_VAR_NAME }}",
			expected: "a test varvalue",
		},
	}
	for _, ie := range tests {
		text := ie.input
		err = store.Interpolate(&text)
		if err != nil {
			t.Errorf(err.Error())
		}
		if text != ie.expected {
			t.Errorf("text interpolation failed got %v expected %v", text, ie.expected)
		}
	}
}

func TestDeepInterpolation(t *testing.T) {
	store := storage.NewStore()
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
	}, "generic")
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME2",
		Aout:    nil,
		Value:   "varvalue2",
	}, "generic")

	text1 := "holi"
	text2 := "holi {{ SINGLE_VAR_NAME }}"
	text3 := "holi2"
	text4 := "holi2 {{ SINGLE_VAR_NAME2 }}"
	awsinput1 := &ec2.DescribeImagesInput{
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
	awsinput2 := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("cucutras"),
				Values: []*string{
					&text3,
					&text4,
				},
			},
		},
	}

	inp := struct {
		A *ec2.DescribeImagesInput
		B *ec2.DescribeImagesInput
	}{
		A: awsinput1,
		B: awsinput2,
	}

	err := store.DeepInterpolation(inp)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text1 != "holi" {
		t.Errorf("text interpolation failed, expected %v, got %v", "holi", text1)
	}
	if text2 != "holi varvalue" {
		t.Errorf("text interpolation failed, expected %v, got %v", "holi varvalue", text2)
	}
	if text3 != "holi2" {
		t.Errorf("text interpolation failed, expected %v, got %v", "holi2", text3)
	}
	if text4 != "holi2 varvalue2" {
		t.Errorf("text interpolation failed, expected %v, got %v", "holi2 varvalue2", text4)
	}
}

func TestMixedReferences(t *testing.T) {
	var err error
	store := storage.NewStore()
	ref := "OUTPUT_VAR_NAME"
	action := &blueprint.Action{
		Provider: "generic",
		Output:   &ref,
	}
	demoresult := &ec2.Image{
		ImageId: aws.String("test"),
		Name:    aws.String("test1234"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("tagkey0"),
				Value: aws.String("tagvalue0"),
			},
			{
				Key:   aws.String("tagkey1"),
				Value: aws.String("tagvalue1"),
			},
		},
	}
	aout := base.NewActionOutput(action, demoresult, demoresult.ImageId)
	store.Insert(aout.Records[0], action.Provider)

	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
	}, "generic")

	text := "a test {{ OUTPUT_VAR_NAME }} {{ SINGLE_VAR_NAME }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	expected :=
		`a test {
    "imageId": "test",
    "name": "test1234",
    "tagSet": [
        {
            "key": "tagkey0",
            "value": "tagvalue0"
        },
        {
            "key": "tagkey1",
            "value": "tagvalue1"
        }
    ]
} varvalue`
	if text != expected {
		t.Errorf("text interpolation failed in mixed reference")
	}
	text = "a test {{ OUTPUT_VAR_NAME.imageId }} {{ SINGLE_VAR_NAME }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	expected = "a test test varvalue"
	if text != expected {
		t.Errorf("text interpolation failed in mixed reference")
	}
}

func TestMagicReferences(t *testing.T) {
	var err error
	store := storage.NewStore()
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
		Fail:    true,
		Error:   fmt.Errorf("err"),
	}, "generic")
	text := "{{ SINGLE_VAR_NAME.__haserror }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "true" {
		t.Errorf("text interpolation failed for .__haserror true")
	}

	text = "{{ SINGLE_VAR_NAME.__error }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "err" {
		t.Errorf("text interpolation failed for .__error")
	}

	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME2",
		Aout:    nil,
		Value:   "varvalue",
	}, "generic")
	text = "{{ SINGLE_VAR_NAME2.__haserror }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "false" {
		t.Errorf("text interpolation failed for .__haserror false")
	}

	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME3",
		Aout:    nil,
		Value:   "varvalue",
		ValueID: "id",
	}, "generic")
	text = "{{ SINGLE_VAR_NAME3.id }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "id" {
		t.Errorf("text interpolation failed for .id")
	}

	/////

	ref := "SINGLE_VAR_NAME4"
	action := &blueprint.Action{
		Provider: "generic",
		Output:   &ref,
	}
	demoresult := &ec2.Image{
		ImageId: aws.String("test"),
		Name:    aws.String("test1234"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("tagkey0"),
				Value: aws.String("tagvalue0"),
			},
			{
				Key:   aws.String("tagkey1"),
				Value: aws.String("tagvalue1"),
			},
		},
	}
	aout := base.NewActionOutput(action, demoresult, demoresult.ImageId)
	store.Insert(aout.Records[0], action.Provider)

	text = "{{ SINGLE_VAR_NAME4.__json }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	expected :=
		`{
    "imageId": "test",
    "name": "test1234",
    "tagSet": [
        {
            "key": "tagkey0",
            "value": "tagvalue0"
        },
        {
            "key": "tagkey1",
            "value": "tagvalue1"
        }
    ]
}`
	if text != expected {
		t.Errorf("text interpolation failed for .__json. \nExpected:\n %v \nGot:\n %v", expected, text)
	}
}

func TestEnv(t *testing.T) {
	os.Setenv("VARNAME", "VARVALUE")
	text := "{{ ENV.VARNAME }}"
	store := storage.NewStore()
	store.Interpolate(&text)
	if text != "VARVALUE" {
		t.Errorf("text interpolation failed for ENV.")
	}
}

func TestDuplicate(t *testing.T) {
	store := storage.NewStore()
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
		ValueID: "value1",
	}, "generic")
	store.SetPrivateVar("VARNAME", "varvalue")
	tp := &testProvider{}
	store.StoreProvider("generic", tp)
	ref := "SINGLE_VAR_NAME4"
	action := &blueprint.Action{
		ActionID: "actionid",
		Provider: "generic",
		Output:   &ref,
	}
	demoresult := &ec2.Image{
		ImageId: aws.String("test"),
		Name:    aws.String("test1234"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("tagkey0"),
				Value: aws.String("tagvalue0"),
			},
			{
				Key:   aws.String("tagkey1"),
				Value: aws.String("tagvalue1"),
			},
		},
	}
	aout := base.NewActionOutput(action, demoresult, demoresult.ImageId)
	store.Insert(aout.Records[0], action.Provider)
	store2 := store.Duplicate()

	a, _ := store.GetByRefName("SINGLE_VAR_NAME")
	b, _ := store2.GetByRefName("SINGLE_VAR_NAME")
	if a.Value.(string) != b.Value.(string) {
		t.Errorf("Duplicate fail: not equal")
	}
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue2",
	}, "generic")
	a, _ = store.GetByRefName("SINGLE_VAR_NAME")
	if a.Value.(string) == b.Value.(string) {
		t.Errorf("Duplicate fail: not operating dupe")
	}
	_, err := store.GetByRefName("UNDEFINEDVAR")
	if err == nil {
		t.Errorf("Undefined var should be nil")
	}
	a, _ = store.GetByValueID("value1", "generic")
	if a.Value.(string) != b.Value.(string) {
		t.Errorf("Duplicate fail: not operating dupe got %v, expected %v", a.Value.(string), b.Value.(string))
	}
	_, err = store.GetByValueID("badid", "generic")
	if err == nil {
		t.Errorf("Duplicate fail: not operating dupe")
	}
	pr := store2.GetPrivateVar("VARNAME")
	if pr.(string) != "varvalue" {
		t.Errorf("Duplicate fail: no private var. Expected 'varvalue', got %v", pr)
	}
	pr = store2.GetPrivateVar("UNDEFINEDVARNAME")
	if pr != nil {
		t.Errorf("Duplicate fail: var should be nil")
	}
	ac, _ := store2.GetActionOutputByActionID(&action.ActionID)
	if ac.Records[0].Action.ActionID != action.ActionID {
		t.Errorf("Duplicate fail by actionoutput")
	}
	fakeid := "fakeid"
	_, err = store2.GetActionOutputByActionID(&fakeid)
	if err == nil {
		t.Errorf("Duplicate fail: fake id should not return value")
	}
	store.Merge(store)
	store.Merge(store2)
	a, _ = store.GetByRefName("SINGLE_VAR_NAME")
	if a.Value.(string) != "varvalue" {
		t.Errorf("Merge fail")
	}
	_, err = store2.GetByRefName("UNDEFINEDVAR")
	if err == nil {
		t.Errorf("Undefined var should be nil")
	}
}

func TestMerge(t *testing.T) {
	store := storage.NewStore()
	store.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue",
	}, "generic")
	store2 := store.Duplicate()
	store3 := storage.NewStore()
	store3.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue2",
	}, "generic")
	store3.Insert(&base.StorageRecord{
		RefName: "SINGLE_VAR_NAME",
		Aout:    nil,
		Value:   "varvalue3",
	}, "generic")
	tp := &testProvider{}
	store3.StoreProvider("testprovider", tp)
	store.Merge(store2)
	store.Merge(store3)
	a, _ := store.GetByRefName("SINGLE_VAR_NAME")
	if a.Value.(string) != "varvalue3" {
		t.Errorf("Merge fail")
	}
	_, err := store.GetByRefName("UNDEFINEDVAR")
	if err == nil {
		t.Errorf("Undefined var should be nil")
	}
	tp3, err := store.GetProvider("testprovider")
	if err != nil {
		t.Errorf(err.Error())
	}
	if tp != tp3 {
		t.Errorf("provider should be the same after merge, got %v, expected %v", tp3, tp)
	}
}

func TestPrivatevar(t *testing.T) {
	store := storage.NewStore()
	store.SetPrivateVar("VARNAME", "varvalue")
	pr := store.GetPrivateVar("VARNAME")
	if pr.(string) != "varvalue" {
		t.Errorf("private var fail")
	}
	pr = store.GetPrivateVar("UNDEFINEDVARNAME")
	if pr != nil {
		t.Errorf("Duplicate fail: var should be nil")
	}
}

func TestGetByActionID(t *testing.T) {
	var err error
	store := storage.NewStore()
	ref := "OUTPUT_VAR_NAME"
	action := &blueprint.Action{
		ActionID: "actionTestID",
		Provider: "generic",
		Output:   &ref,
	}
	demoresult := &ec2.Image{
		ImageId: aws.String("test"),
		Name:    aws.String("test1234"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("tagkey0"),
				Value: aws.String("tagvalue0"),
			},
			{
				Key:   aws.String("tagkey1"),
				Value: aws.String("tagvalue1"),
			},
		},
	}
	aout := base.NewActionOutput(action, demoresult, demoresult.ImageId)
	store.Insert(aout.Records[0], action.Provider)
	expected := "actionTestID"
	aout2, err := store.GetActionOutputByActionID(&expected)
	if err != nil {
		t.Errorf(err.Error())
	}
	if aout2 != aout {
		t.Errorf("Err extracting aout by action id")
	}
}

func TestProviders(t *testing.T) {
	tp := &testProvider{}
	store := storage.NewStore()
	store.StoreProvider("testprovider", tp)
	tp2, err := store.GetProvider("testprovider")
	if err != nil {
		t.Errorf(err.Error())
	}
	if tp != tp2 {
		t.Errorf("Err storing provider")
	}
	_, err = store.GetProvider("undefinedprovider")
	if err == nil {
		t.Errorf("Undefined provider should not exist")
	}
	if !store.ExistsProvider("testprovider") {
		t.Errorf("provider should exists")
	}
	if store.ExistsProvider("undefinedprovider") {
		t.Errorf("provider should not exists")
	}
}

func TestLogger(t *testing.T) {
	lg := &fakeLogger{}
	store := storage.NewStore()
	store.SetLogger(lg)
	lg2 := store.GetLogger()
	if lg != lg2 {
		t.Errorf("logger should be the same")
	}
	lg2.LogCritical("test")
	lg2.LogErr("test")
	lg2.ByteLogErr([]byte("test"))
	lg2.LogWarn("test")
	lg2.LogInfo("test")
	lg2.ByteLogInfo([]byte("test"))
	lg2.LogDebug("test")
}

func TestJSONPath(t *testing.T) {
	var err error
	store := storage.NewStore()
	ref := "OUTPUT_VAR_NAME"
	action := &blueprint.Action{
		Provider: "generic",
		Output:   &ref,
	}
	demoresult := &ec2.Image{
		ImageId: aws.String("test"),
		Name:    aws.String("test1234"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("tagkey0"),
				Value: aws.String("tagvalue0"),
			},
			{
				Key:   aws.String("tagkey1"),
				Value: aws.String("tagvalue1"),
			},
		},
	}
	aout := base.NewActionOutput(action, demoresult, demoresult.ImageId)
	err = store.Insert(aout.Records[0], action.Provider)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	text := "{{ OUTPUT_VAR_NAME.tagSet | $[1].value }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "tagvalue1" {
		t.Errorf("text interpolation failed")
	}

	text = "{{ OUTPUT_VAR_NAME.__json | $.tagSet[1].value }}"
	err = store.Interpolate(&text)
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "tagvalue1" {
		t.Errorf("text interpolation failed")
	}

}
