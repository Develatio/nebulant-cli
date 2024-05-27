// MIT License
//
// Copyright (C) 2023  Develatio Technologies S.L.

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
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

type r2UploadParametersPath struct {
	Bucket *string `json:"bucket" validate:"required"`
	Dst    *string `json:"dest" validate:"required"`
	Src    *string `json:"src" validate:"required"`
}

type r2UploadParameters struct {
	Paths []r2UploadParametersPath `json:"paths" validate:"required"`
}

type asyncWalk struct {
	paths chan string
	errs  []error
}

func (w *asyncWalk) walk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		w.errs = append(w.errs, err)
		w.close()
		return err
	}
	if !info.IsDir() {
		w.paths <- path
	}
	return nil
}

func (w *asyncWalk) close() {
	close(w.paths)
}

type r2uploadonefile struct {
	Uploader *manager.Uploader
	Logger   base.ILogger
	Basepath string
	Wpath    string
	Dst      string
	Bucket   string
}

func (r *r2uploadonefile) upload() (*manager.UploadOutput, error) {
	// relative path
	rel, err := filepath.Rel(r.Basepath, r.Wpath)
	if err != nil {
		return nil, err
	}

	upfile, err := os.Open(r.Wpath)
	if err != nil {
		return nil, err
	}
	defer upfile.Close()
	r.Logger.LogDebug(fmt.Sprintf("uploading file to bucket %s and key %v", r.Bucket, filepath.Join(r.Dst, rel)))
	result, err := r.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: &r.Bucket,
		Key:    aws.String(filepath.Join(r.Dst, rel)),
		Body:   upfile,
	})
	if err != nil {
		var mu manager.MultiUploadFailure
		if errors.As(err, &mu) {
			return nil, errors.Join(err, fmt.Errorf("upload ID: %s", mu.UploadID()))
		}
		return nil, err
	}

	// current file uploaded
	// r.Logger.LogInfo(fmt.Sprintf("uploaded file %s -> %s", r.Wpath, result.Location))
	return result, nil
}

// R2Upload func
func R2Upload(ctx *ActionContext) (*base.ActionOutput, error) {
	params := &r2UploadParameters{}
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	for i := 0; i < len(params.Paths); i++ {
		upp := params.Paths[i]
		if upp.Dst != nil {
			if ndst, ok := strings.CutPrefix(*upp.Dst, "/"); ok {
				ctx.Logger.LogWarn(fmt.Sprintf("starting slash of path %s will be removed: %s", *upp.Dst, ndst))
				upp.Dst = &ndst
			}
		}
	}

	err := ctx.Store.DeepInterpolation(params)
	if err != nil {
		return nil, err
	}

	client := ctx.NewS3Client()
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10 MiB
		u.Concurrency = 3
	})
	r2up := &r2uploadonefile{
		Uploader: uploader,
	}

	ctx.Logger.LogDebug("Uploading...")
	for i := 0; i < len(params.Paths); i++ {
		upp := params.Paths[i]
		file, err := os.Open(*upp.Src)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// upload dir or file
		fi, err := file.Stat()
		switch {
		case err != nil:
			return nil, err
		case fi.IsDir():
			// upload dir
			err := file.Close()
			if err != nil {
				return nil, err
			}

			ctx.Logger.LogDebug("Walking files...")
			// walk
			ww := &asyncWalk{}
			ww.paths = make(chan string)
			go func() {
				err := filepath.Walk(*upp.Src, ww.walk)
				if err != nil {
					ww.errs = append(ww.errs, err)
				}
				ww.close()
			}()

			// range over channel
			// range over ww.paths while walk is still
			// walking. ww.close() will EOF the range
			for wpath := range ww.paths {
				ctx.Logger.LogInfo(fmt.Sprintf("uploading file %s", wpath))
				r2up.Logger = ctx.Logger
				r2up.Basepath = *upp.Src
				r2up.Wpath = wpath
				r2up.Dst = *upp.Dst
				r2up.Bucket = *upp.Bucket
				out, err := r2up.upload()
				if err != nil {
					return nil, err
				}
				ctx.Logger.LogInfo(fmt.Sprintf("uploaded file %s -> %s", wpath, out.Location))

				// check for errs
				if len(ww.errs) > 0 {
					return nil, errors.Join(ww.errs...)
				}
			}
		default:
			// upload file
			r2up.Logger = ctx.Logger
			r2up.Basepath = filepath.Dir(*upp.Src)
			r2up.Wpath = *upp.Src
			r2up.Dst = *upp.Dst
			r2up.Bucket = *upp.Bucket
			out, err := r2up.upload()
			if err != nil {
				return nil, err
			}
			ctx.Logger.LogInfo(fmt.Sprintf("uploaded file %s -> %s", *upp.Src, out.Location))
		}
	}

	aout := base.NewActionOutput(ctx.Action, nil, nil)
	return aout, nil
}
