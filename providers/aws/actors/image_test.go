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
