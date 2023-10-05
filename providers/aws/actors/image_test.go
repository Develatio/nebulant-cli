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

package actors_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	"github.com/develatio/nebulant-cli/providers/aws/actors"
	"github.com/develatio/nebulant-cli/storage"
)

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
func (l *fakeLogger) SetThreadID(ti string)   {}

type fakeEC2Client struct {
	ec2iface.EC2API
}

func mockapi() (*session.Session, error) {
	actors.NewActionContext = func(awsSess *session.Session, action *blueprint.Action, store base.IStore, logger base.ILogger) *actors.ActionContext {
		return &actors.ActionContext{
			AwsSess: awsSess,
			Action:  action,
			Store:   store,
			Logger:  logger,
			NewEC2Client: func() ec2iface.EC2API {
				return &fakeEC2Client{}
			},
		}
	}
	sess, serr := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if serr != nil {
		return nil, serr
	}
	return sess, nil
}

func (f *fakeEC2Client) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	if *input.Filters[0].Name == "none" {
		return &ec2.DescribeImagesOutput{
			Images: []*ec2.Image{},
		}, nil
	}
	if *input.Filters[0].Name == "onlyone" {
		return &ec2.DescribeImagesOutput{
			Images: []*ec2.Image{
				{
					Name: aws.String("Image1"),
				},
			},
		}, nil
	}
	return &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				Name: aws.String("Image1"),
			},
			{
				Name: aws.String("Image2"),
			},
		},
	}, nil
}

func TestFindImages(t *testing.T) {
	sess, err := mockapi()
	if err != nil {
		t.Errorf(err.Error())
	}

	params := ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("architecture"),
				Values: []*string{aws.String("x86_64")},
			},
		},
	}

	// normal call
	p, err := json.Marshal(params)
	if err != nil {
		t.Errorf(err.Error())
	}
	action := &blueprint.Action{
		Provider:   "aws",
		ActionName: "find_images",
		Parameters: p,
	}
	lg := &fakeLogger{}
	ctx := actors.NewActionContext(sess, action, storage.NewStore(), lg)
	aout, err := actors.FindImages(ctx)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(aout.Records) <= 0 {
		t.Error("Aout return zero records")
	}

	// malformed params should return err
	b := action.Parameters[0]
	action.Parameters[0] = []byte("-")[0]
	ctx = actors.NewActionContext(sess, action, storage.NewStore(), lg)
	_, err = actors.FindImages(ctx)
	if err == nil {
		t.Errorf("Malformed params should fail")
	}
	action.Parameters[0] = b

}

func TestFindOneImage(t *testing.T) {
	sess, err := mockapi()
	if err != nil {
		t.Errorf(err.Error())
	}

	params := ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("architecture"),
				Values: []*string{aws.String("x86_64")},
			},
		},
	}

	p, err := json.Marshal(params)
	if err != nil {
		t.Errorf(err.Error())
	}
	action := &blueprint.Action{
		Provider:   "aws",
		ActionName: "findone_image",
		Parameters: p,
	}
	lg := &fakeLogger{}

	// find one should fail on many results
	ctx := actors.NewActionContext(sess, action, storage.NewStore(), lg)
	_, err = actors.FindOneImage(ctx)
	if err == nil {
		t.Errorf("FindOneImage should fail on many results")
	}

	// err from parent (FindImages)
	action.Parameters[0] = []byte("-")[0]
	ctx = actors.NewActionContext(sess, action, storage.NewStore(), lg)
	_, err = actors.FindOneImage(ctx)
	if err == nil {
		t.Errorf("FindOneImage should fail if parent FindImages fails")
	}

	// one result should go ok
	params = ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("onlyone"),
				Values: []*string{aws.String("x86_64")},
			},
		},
	}
	p, err = json.Marshal(params)
	if err != nil {
		t.Errorf(err.Error())
	}
	action.Parameters = p
	ctx = actors.NewActionContext(sess, action, storage.NewStore(), lg)
	_, err = actors.FindOneImage(ctx)
	if err != nil {
		t.Errorf(err.Error())
	}

	// none result should fail
	params = ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("none"),
				Values: []*string{aws.String("x86_64")},
			},
		},
	}
	p, err = json.Marshal(params)
	if err != nil {
		t.Errorf(err.Error())
	}
	action.Parameters = p
	ctx = actors.NewActionContext(sess, action, storage.NewStore(), lg)
	_, err = actors.FindOneImage(ctx)
	if err == nil {
		t.Errorf("None result should fail in FindOneImage")
	}
}
