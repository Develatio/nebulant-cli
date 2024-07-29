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

package actors

// Considerations:
// - Only one instance of runActor per script or cmd. Keep in mind that for each
// execution there must be an output and it must be stored, so the functionality
// of executing multiple scripts with an instance of runActor should not be
// implemented.
//

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"golang.org/x/net/html/charset"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/util"
)

type BodyType string
type PartType string

const (
	BodyTypeNone               BodyType = "none"
	BodyTypeFormData           BodyType = "form-data"
	BodyTypeXWWWFormUrlencoded BodyType = "x-www-form-urlencoded"
	BodyTypeRaw                BodyType = "raw"
	BodyTypeBinary             BodyType = "binary"
)

const (
	PartTypeText PartType = "text"
	PartTypeFile PartType = "file"
)

type httpBodyMultiPart struct {
	Name        *string  `json:"name" validate:"required"`
	Value       *string  `json:"value" validate:"required"`
	PType       PartType `json:"type" validate:"required"`
	ContentType *string  `json:"content_type"`
}

type httpBodyUrlencoded struct {
	Name  *string `json:"name" validate:"required"`
	Value *string `json:"value" validate:"required"`
}

type httpHeader struct {
	Enabled bool    `json:"enabled"`
	Key     *string `json:"name" validate:"required"`
	Value   *string `json:"value" validate:"required"`
}

// issue #11
type httpRequestParameters struct {
	Method           *string       `json:"http_verb" validate:"required"`
	Url              *string       `json:"endpoint" validate:"required"`
	Headers          []*httpHeader `json:"headers"`
	BodyType         BodyType      `json:"body_type" validate:"required"`
	IgnoreInvalidSSL bool          `json:"ignore_invalid_certs"`
	CookieJarName    *string       `json:"cookie_jar"`
}

type httpRequestParametersMultiPartBody struct {
	MultipartBody []*httpBodyMultiPart `json:"body" validate:"required"`
}

type httpRequestParametersUrlEncodedBody struct {
	UrlEncodedBody []*httpBodyUrlencoded `json:"body" validate:"required"`
}

type httpRequestParametersRawBody struct {
	RawBody *string `json:"body" validate:"required"`
}

type httpRequestParametersBinaryBody struct {
	BinaryBody *string `json:"body" validate:"required"`
}

type httpRequestOutput struct {
	Status     string                 `json:"status"`
	StatusCode int                    `json:"status_code"`
	Headers    string                 `json:"headers"`
	Body       encoding.TextMarshaler `json:"body"`
	FilePath   string                 `json:"filepath"`
}

type httpRequestOutputMarshalerBody struct {
	filepath string
}

func (h *httpRequestOutputMarshalerBody) MarshalText() ([]byte, error) {
	return os.ReadFile(h.filepath)
}

// RunRemoteScript func
func HttpRequest(ctx *ActionContext) (*base.ActionOutput, error) {
	var err error
	var req *http.Request

	p := &httpRequestParameters{}
	if err = json.Unmarshal(ctx.Action.Parameters, p); err != nil {
		return nil, err
	}
	if p.Url == nil {
		return nil, fmt.Errorf("http endpoint parameter of HTTP request cannot be empty")
	}
	u, err := url.Parse(*p.Url)
	if err != nil {
		return nil, err
	}

	if ctx.Rehearsal {
		return nil, nil
	}

	// issue #11
	switch p.BodyType {
	case BodyTypeFormData:
		body := new(bytes.Buffer)
		w := multipart.NewWriter(body)
		// already called below. Defer for security.
		defer w.Close()
		// body part definitions
		// body is:
		// {
		// 	name: "campo1",
		// 	value: "valor1",
		// 	type: "", // "text" \ "file"
		//		// "type:text" -> content_type is the value of "Content-Type" part header
		//		// "type:file" -> content_type is ignored. The path of file is in the value attr
		// content_type: "",
		// }
		param := &httpRequestParametersMultiPartBody{}
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}

		// loop for body parts
		for _, part := range param.MultipartBody {
			switch part.PType {
			case PartTypeFile: // type file: read file and write content
				file, err := os.Open(*part.Value)
				if err != nil {
					return nil, err
				}
				defer file.Close()
				finf, err := file.Stat()
				if err != nil {
					return nil, err
				}
				ff, err := w.CreateFormFile(*part.Name, finf.Name())
				if err != nil {
					return nil, err
				}
				_, err = io.Copy(ff, file)
				if err != nil {
					return nil, err
				}
				ctx.Logger.LogDebug("Append file part: " + finf.Name())
			case PartTypeText: // type text, write as part and override content_type
				h := make(textproto.MIMEHeader)
				n := strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(*part.Name)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, n))
				if part.ContentType != nil {
					h.Set("Content-Type", *part.ContentType)
				} else {
					h.Set("Content-Type", "text/plain")
				}
				ff, err := w.CreatePart(h)
				if err != nil {
					return nil, err
				}
				_, err = ff.Write([]byte(*part.Value))
				if err != nil {
					return nil, err
				}
				ctx.Logger.LogDebug("Append text content: " + *part.Value)
			}
		}
		// extremly important. w.Close() will write the
		// last part of the body boundary
		err := w.Close()
		if err != nil {
			return nil, err
		}
		// request with formatted body parts
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
		// there is the boundary
		req.Header.Set("Content-Type", w.FormDataContentType())
	case BodyTypeNone:
		// no body with type none
		req, err = http.NewRequest(*p.Method, *p.Url, http.NoBody)
		if err != nil {
			return nil, err
		}
	case BodyTypeXWWWFormUrlencoded:
		// body part definitions
		// body is {name: "campo1", value: "valor1"}
		param := &httpRequestParametersUrlEncodedBody{}
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}
		// append key:value
		fdata := url.Values{}
		for _, kv := range param.UrlEncodedBody {
			fdata.Add(*kv.Name, *kv.Value)
			ctx.Logger.LogDebug("Append urlenc key: " + *kv.Name)
		}
		// url-encoded reader
		body := strings.NewReader(fdata.Encode())
		// request with url-encoded body
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case BodyTypeRaw:
		param := &httpRequestParametersRawBody{}
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}
		body := strings.NewReader(*param.RawBody)
		// the content type should be setted by the user
		// in httpRequestParameters.Headers
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
	case BodyTypeBinary:
		param := &httpRequestParametersBinaryBody{}
		// the body contains the path of a file
		if err := util.UnmarshalValidJSON(ctx.Action.Parameters, param); err != nil {
			return nil, err
		}
		file, err := os.Open(*param.BinaryBody)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		fcontent, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		// detect the file mime type
		ct := http.DetectContentType(fcontent)
		body := bytes.NewBuffer(fcontent)
		req, err = http.NewRequest(*p.Method, *p.Url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", ct)
	default:
		// default is like none type
		req, err = http.NewRequest(*p.Method, *p.Url, http.NoBody)
		if err != nil {
			return nil, err
		}
	}

	ctx.Logger.LogDebug("Requesting " + req.URL.String())
	req.Header.Set("Accept-Charset", "utf-8")
	// set headers
	for _, hh := range p.Headers {
		if !hh.Enabled {
			continue
		}
		if req.Header.Get(*hh.Key) != "" {
			// already seted header
			continue
		}
		err := ctx.Store.Interpolate(hh.Value)
		if err != nil {
			return nil, err
		}
		ctx.Logger.LogDebug("Use request header, " + *hh.Key + ": " + *hh.Value)
		req.Header.Set(*hh.Key, *hh.Value)
	}

	tr := &http.Transport{
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    false,
		ResponseHeaderTimeout: 30 * time.Second,
		// #nosec G402 -- Leave to user the choose to be insecure
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: p.IgnoreInvalidSSL,
		},
	}

	// creates a completly new cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// handles user-managed cookie jar
	if p.CookieJarName != nil && *p.CookieJarName != "" {
		jari := ctx.Store.GetPrivateVar("cookiejar_" + *p.CookieJarName)
		if jari == nil {
			// store the previous created jar into the store with
			// the provided name
			ctx.Store.SetPrivateVar("cookiejar_"+*p.CookieJarName, jar)
		} else {
			// recover stored cookie with the provided name
			if storedjar, ok := jari.(*cookiejar.Jar); ok {
				jar = storedjar
			} else {
				return nil, fmt.Errorf("cannot recover cookie %s", *p.CookieJarName)
			}
		}
	}

	client := &http.Client{Transport: tr, Jar: jar}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("http request error"), err)
	}
	defer resp.Body.Close()

	result := httpRequestOutput{}

	// debug status
	ctx.Logger.LogDebug("Request response status: " + resp.Status)
	result.Status = resp.Status
	result.StatusCode = resp.StatusCode

	// debug headers
	sw := new(strings.Builder)
	err = resp.Header.Write(sw)
	if err != nil {
		return nil, err
	}
	ctx.Logger.LogDebug("Response Headers: " + sw.String())
	result.Headers = sw.String()

	// debug body
	var rawbody io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip", "x-gzip": // LZ77 CRC 32bits, unix like gzip, x-gzip as http1.1 alias
		rawbody, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer rawbody.Close()
	case "deflate": // zlib rfc 1950 + deflate rfc 1951
		rawbody, err = zlib.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer rawbody.Close()
	case "br": // brotli
		rawbody = io.NopCloser(brotli.NewReader(resp.Body))
	default:
		rawbody = resp.Body
	}

	// Charset decode
	// contentType can be empty string
	contentType := resp.Header.Get("Content-Type")
	// NewReader returns an io.Reader that converts the content of rawbody to UTF-8.
	// https://pkg.go.dev/golang.org/x/net/html/charset#NewReader
	// charset.NewReader will scan body if contentType is bad/empty string
	// https://cs.opensource.google/go/x/net/+/refs/tags/v0.8.0:html/charset/charset.go;l=71
	// also track this TODO comment and the default behavior
	// https://cs.opensource.google/go/x/net/+/refs/tags/v0.8.0:html/charset/charset.go;l=102
	dcr, err := charset.NewReader(rawbody, contentType)
	if err != nil {
		return nil, err
	}

	f, err := os.CreateTemp("", "nbl*")
	if err != nil {
		log.Fatal(err)
	}
	result.FilePath = f.Name()
	defer f.Close()

	ctx.Logger.LogDebug("Storing http response into tmp file " + f.Name())

	bar := cast.NewProgress(resp.ContentLength, path.Base(u.Path), "", "", "", "")
	if err != nil {
		return nil, errors.Join(fmt.Errorf("internal err creating the progress bar"), err)
	}

	// swb := new(strings.Builder)
	written, err := io.Copy(io.MultiWriter(f, bar), dcr) // #nosec G110 -- The user is free to get decompression bomb
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error decoding content"), err)
	}

	swb := make([]byte, 5000)
	n, err := f.ReadAt(swb, 0)
	if err != nil && err != io.EOF {
		return nil, errors.Join(fmt.Errorf("error reading from request tmp file"), err)
	}
	err = nil
	ctx.Logger.LogDebug("Body: " + string(swb))
	if written > int64(n) {
		ctx.Logger.LogDebug("...[Truncated]")
	}

	result.Body = &httpRequestOutputMarshalerBody{
		filepath: f.Name(),
	}

	aout := base.NewActionOutput(ctx.Action, result, nil)
	return aout, err
}
