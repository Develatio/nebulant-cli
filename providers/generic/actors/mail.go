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

package actors

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/develatio/nebulant-cli/netproto/smtp"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/util"
)

type sendMailParametersBody struct {
	HTML  *string `json:"html"`
	Plain *string `json:"plain"`
}

type sendMailParameters struct {
	Username         *string                 `json:"username" validate:"required"`
	Password         *string                 `json:"password" validate:"required"`
	Server           *string                 `json:"server" validate:"required"`
	Port             *int                    `json:"port"`
	IgnoreInvalidSSL bool                    `json:"ignore_invalid_ssl"`
	ForceSSL         bool                    `json:"force_ssl"`
	Subject          *string                 `json:"subject"`
	Body             *sendMailParametersBody `json:"body"`
	Attachments      []string                `json:"attachments"`
	From             *string                 `json:"from" validate:"required"`
	To               []string                `json:"to" validate:"required"`
	CC               []string                `json:"cc"`
	BCC              []string                `json:"bcc"`
	ReplyTo          *string                 `json:"reply_to"`
}

func SendMail(ctx *ActionContext) (*base.ActionOutput, error) {
	params := new(sendMailParameters)
	if err := util.UnmarshalValidJSON(ctx.Action.Parameters, params); err != nil {
		return nil, err
	}

	// It  is  recommended
	// that,  if  present,  headers be sent in the order "Return-
	// Path", "Received", "Date",  "From",  "Subject",  "Sender",
	// "To", "cc", etc.
	// rfc822 4.1.
	// https://www.rfc-editor.org/rfc/rfc822.html#section-4.1

	// Once a field has been unfolded, it may be viewed as being com-
	//     posed of a field-name followed by a colon (":"), followed by a
	//     field-body, and  terminated  by  a  carriage-return/line-feed.
	// rfc822 3.1.2
	// https://www.rfc-editor.org/rfc/rfc822.html#section-3.1.2
	var msg []byte

	// Part of minimum rquired, rfc822 A.3.1
	// https://www.rfc-editor.org/rfc/rfc822.html#appendix-A.3.1
	msg = append(msg, []byte("From: "+*params.From+"\r\n")...)

	// Subject is optional, rfc822 4.1
	// https://www.rfc-editor.org/rfc/rfc822.html#section-4.1
	if params.Subject != nil {
		msg = append(msg, []byte("Subject: "+*params.Subject+"\r\n")...)
	}

	// "TO" is required to have at least one address, rfc822 A.3.1.
	// https://www.rfc-editor.org/rfc/rfc822.html#appendix-A.3.1
	if len(params.To) <= 0 {
		return nil, fmt.Errorf("'To' header is required and should has at least one address")
	}
	msg = append(msg, []byte("To: "+strings.Join(params.To, ", ")+"\r\n")...)

	// "CC" are required to contain at least one address, rfc822 C.3.4.
	// https://www.rfc-editor.org/rfc/rfc822.html#appendix-C.3.4
	if len(params.CC) > 0 {
		msg = append(msg, []byte("cc: "+strings.Join(params.CC, ", ")+"\r\n")...)
	}

	// BCC are excluded from headers

	// always mime 1.0
	msg = append(msg, []byte("MIME-Version: 1.0\r\n")...)
	mixed := false
	related := false

	if len(params.Attachments) > 0 {
		mixed = true
		msg = append(msg, []byte("Content-Type: multipart/mixed; boundary=mixedBoundary\r\n")...)
		msg = append(msg, []byte("\r\n")...)
	}

	if params.Body != nil {
		// Multipart body? if yes, switch to rfc2046
		// https://datatracker.ietf.org/doc/html/rfc2046
		// In the case of multipart entities, in which one or more different
		// sets of data are combined in a single body, a "multipart" media type
		// field must appear in the entity's header.  The body must then contain
		// one or more body parts, each preceded by a boundary delimiter line,
		// and the last one followed by a closing boundary delimiter line.
		// After its boundary delimiter line, each body part then consists of a
		// header area, a blank line, and a body area.  Thus a body part is
		// similar to an RFC 822 message in syntax, but different in meaning.
		// rfc2046
		// https://datatracker.ietf.org/doc/html/rfc2046#section-5.1
		if params.Body.HTML != nil && len(*params.Body.HTML) > 0 && params.Body.Plain != nil && len(*params.Body.Plain) > 0 {
			// if mixed, there is attachment file, so we add related boundary
			// into mixed boundary and alternative boundary into related boundary
			//
			// something like:
			//
			// mixed(
			// 	related(
			//  	alternative(
			// 			plain,
			// 			html
			// 		),
			// 		logo.jpg,
			// 		img2.png,
			// 		...
			// 	),
			// 	attach1,
			// 	attach2,
			// 	...
			// )
			//
			// something like:
			//
			// --mixed
			// 	--related
			// 		--alternative
			// 		--alternative--
			// 	--related--
			// --mixed
			// 	...
			// --mixed
			// 	...
			// --mixed--
			//
			if mixed {
				related = true
				msg = append(msg, []byte("--mixedBoundary\r\n")...)
				msg = append(msg, []byte("Content-Type: multipart/related; boundary=\"relatedBoundary\"\r\n")...)
				msg = append(msg, []byte("\r\n")...)
				msg = append(msg, []byte("--relatedBoundary\r\n")...)
			}

			// alternative subtype for same data
			// rfc2046 5.1.4.
			// https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.4
			msg = append(msg, []byte("Content-Type: multipart/alternative; boundary=alternativeBoundary\r\n")...)

			// The  body  is simply a sequence of lines containing ASCII charac-
			// ters.  It is separated from the headers by a null line  (i.e.,  a
			// line with nothing preceding the CRLF).
			// rfc 822 3.1
			// https://www.rfc-editor.org/rfc/rfc822.html#section-3.1
			msg = append(msg, []byte("\r\n")...)

			// Boundary text plain parts from rfc2046
			// https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.1
			// https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.4

			// text plain boundary
			msg = append(msg, []byte("--alternativeBoundary\r\n")...)
			msg = append(msg, []byte("Content-Type: text/plain; charset=utf-8\r\n")...)
			msg = append(msg, []byte("\r\n")...)
			msg = append(msg, []byte(*params.Body.Plain)...)
			// end of data
			msg = append(msg, []byte("\r\n")...)
			// line break
			msg = append(msg, []byte("\r\n")...)

			// html boundary
			msg = append(msg, []byte("--alternativeBoundary\r\n")...)
			msg = append(msg, []byte("Content-Type: text/html; charset=utf-8\r\n")...)
			msg = append(msg, []byte("\r\n")...)
			msg = append(msg, []byte(*params.Body.HTML)...)
			// end of data
			msg = append(msg, []byte("\r\n")...)
			// line break
			msg = append(msg, []byte("\r\n")...)

			// finish boundary
			msg = append(msg, []byte("--alternativeBoundary--\r\n")...)
			msg = append(msg, []byte("\r\n")...)

		} else if params.Body.HTML != nil && len(*params.Body.HTML) > 0 {
			// html body
			msg = append(msg, []byte("Content-Type: text/html; charset=utf-8\r\n")...)
			// line break to start body
			msg = append(msg, []byte("\r\n")...)
			// body
			msg = append(msg, []byte(*params.Body.HTML)...)
			// end of data
			msg = append(msg, []byte("\r\n")...)
			// line break
			msg = append(msg, []byte("\r\n")...)
		} else if params.Body.Plain != nil && len(*params.Body.Plain) > 0 {
			if mixed {
				msg = append(msg, []byte("--mixedBoundary\r\n")...)
			}
			msg = append(msg, []byte("Content-Type: text/plain; charset=utf-8\r\n")...)
			// line break to start body
			msg = append(msg, []byte("\r\n")...)
			// body
			msg = append(msg, []byte(*params.Body.Plain)...)
			// end of data
			msg = append(msg, []byte("\r\n")...)
			// line break
			msg = append(msg, []byte("\r\n")...)
		} else {
			if mixed {
				msg = append(msg, []byte("--mixedBoundary\r\n")...)
			}
			// empty body
			msg = append(msg, []byte("Content-Type: text/plain; charset=utf-8\r\n")...)
			// line break to start body
			msg = append(msg, []byte("\r\n")...)
			// empty body
			msg = append(msg, []byte("\r\n")...)
		}
	}

	if related {
		// finish related(alternative) boundary
		msg = append(msg, []byte("--relatedBoundary--\r\n")...)
		msg = append(msg, []byte("\r\n")...)
	}

	if mixed {
		for _, filepath := range params.Attachments {
			// start header
			msg = append(msg, []byte("--mixedBoundary\r\n")...)

			// read file
			ff, err := os.Open(filepath) //#nosec G304 -- File inclusion is necesary
			if err != nil {
				return nil, errors.Join(fmt.Errorf("cannot open file for mail attachment"), err)
			}
			defer ff.Close()
			data, err := io.ReadAll(ff)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("cannot read file for mail attachment"), err)
			}
			mimeType := http.DetectContentType(data)
			name := path.Base(filepath)

			// continue header
			msg = append(msg, []byte("Content-Disposition: attachment; filename="+name+"\r\n")...)
			msg = append(msg, []byte("Content-Type: "+mimeType+"; x-unix-mode=0644; name=\""+name+"\"\r\n")...)
			msg = append(msg, []byte("Content-Transfer-Encoding: base64\r\n")...)
			msg = append(msg, []byte("\r\n")...)

			// encode file
			dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
			base64.StdEncoding.Encode(dst, data)

			// write enconding to mail msg
			msg = append(msg, dst...)
			msg = append(msg, []byte("\r\n")...)
			// line break
			msg = append(msg, []byte("\r\n")...)

			// close ff even if we have defer ff.Close()
			// because we are in a loop
			err = ff.Close()
			if err != nil {
				return nil, errors.Join(fmt.Errorf("error after fiel reading for mail attachment"), err)
			}
		}

		// finish related(alternative) boundary
		msg = append(msg, []byte("--mixedBoundary--\r\n")...)
		msg = append(msg, []byte("\r\n")...)
	}

	hostport := net.JoinHostPort(*params.Server, strconv.Itoa(*params.Port))
	ctx.Logger.LogDebug("smtp: " + hostport)

	// Sending "Bcc" messages is accomplished by including an email address in
	// the to parameter but not including it in the msg headers.
	to := append(params.To, params.BCC...)
	to = append(to, params.CC...)

	if ctx.Rehearsal {
		return nil, nil
	}

	auth := smtp.PlainAuth("", *params.Username, *params.Password, *params.Server)
	ctx.Logger.LogDebug("Sending mail...")

	var conn net.Conn
	var err error

	// #nosec G402 -- Leave to user the choose to be insecure
	tlsconfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// either InsecureSkipVerify or ServerName should be setted
		InsecureSkipVerify: params.IgnoreInvalidSSL,
		ServerName:         *params.Server,
	}

	if params.ForceSSL { // #nosec G402 -- Leave to user the choose to be insecure
		// Use SSL at beginning
		conn, err = tls.Dial("tcp", hostport, tlsconfig)
		if err != nil {
			return nil, err
		}
	} else {
		// Use regular tcp for startls
		conn, err = net.Dial("tcp", hostport)
		if err != nil {
			return nil, err
		}
	}

	smctx := &smtp.SendMailCTX{
		Host:      *params.Server,
		Port:      *params.Port,
		Conn:      conn,
		Auth:      auth,
		TLSConfig: tlsconfig,
	}
	err = smtp.SendMail(smctx, *params.From, to, msg)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot send email"), err)
	}
	return nil, nil
}
