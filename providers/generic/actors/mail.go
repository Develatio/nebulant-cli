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
	"fmt"
	"net"
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
	SSL              bool                    `json:"ssl"`
	Subject          *string                 `json:"subject"`
	Body             *sendMailParametersBody `json:"body"`
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
			msg = append(msg, []byte("MIME-Version: 1.0\r\n")...)
			// alternative subtype for same data
			// rfc2046 5.1.4.
			// https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.4
			msg = append(msg, []byte("Content-Type: multipart/alternative; boundary=boundary42\r\n")...)

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
			msg = append(msg, []byte("--boundary42\r\n")...)
			msg = append(msg, []byte("Content-Type: text/plain; charset=utf-8\r\n")...)
			msg = append(msg, []byte("\r\n")...)
			msg = append(msg, []byte(*params.Body.Plain)...)
			msg = append(msg, []byte("\r\n")...)

			// html boundary
			msg = append(msg, []byte("--boundary42\r\n")...)
			msg = append(msg, []byte("Content-Type: text/html; charset=utf-8\r\n")...)
			msg = append(msg, []byte("\r\n")...)
			msg = append(msg, []byte(*params.Body.HTML)...)
			msg = append(msg, []byte("\r\n")...)

			// finish boundary
			msg = append(msg, []byte("--boundary42--\r\n")...)
			msg = append(msg, []byte("\r\n")...)

		} else if params.Body.HTML != nil && len(*params.Body.HTML) > 0 {
			// html body
			msg = append(msg, []byte("MIME-Version: 1.0\r\n")...)
			msg = append(msg, []byte("Content-Type: text/html; charset=utf-8\r\n")...)
			// line break to start body
			msg = append(msg, []byte("\r\n")...)
			// body
			msg = append(msg, []byte(*params.Body.HTML)...)
			msg = append(msg, []byte("\r\n")...)
		} else if params.Body.Plain != nil && len(*params.Body.Plain) > 0 {
			msg = append(msg, []byte("MIME-Version: 1.0\r\n")...)
			msg = append(msg, []byte("Content-Type: text/plain; charset=utf-8\r\n")...)
			// line break to start body
			msg = append(msg, []byte("\r\n")...)
			// body
			msg = append(msg, []byte(*params.Body.Plain)...)
			msg = append(msg, []byte("\r\n")...)
		} else {
			// empty body
			msg = append(msg, []byte("MIME-Version: 1.0\r\n")...)
			msg = append(msg, []byte("Content-Type: text/plain; charset=utf-8\r\n")...)
			// line break to start body
			msg = append(msg, []byte("\r\n")...)
			// empty body
			msg = append(msg, []byte("\r\n")...)
		}
	}

	host := net.JoinHostPort(*params.Server, strconv.Itoa(*params.Port))

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

	if params.SSL { // #nosec G402 -- Leave to user the choose to be insecure
		// Use SSL at beginning
		conn, err = tls.Dial("tcp", host, &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: params.IgnoreInvalidSSL,
			ServerName:         host,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Use regular tcp for startls
		conn, err = net.Dial("tcp", host)
		if err != nil {
			return nil, err
		}
	}

	err = smtp.SendMail(conn, auth, *params.From, to, msg)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
