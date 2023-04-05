package util

import (
	"bytes"
	"net/http"
)

type MimeDetectorWriter struct {
	www      []byte
	MimeType *string
}

func (m *MimeDetectorWriter) Write(p []byte) (int, error) {
	// 512 is the snifflen from:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.20.3:src/net/http/sniff.go;l=13
	if m.MimeType != nil {
		return len(p), nil
	}
	if len(m.www) > 512 {
		m.MimeType = new(string)
		*m.MimeType = http.DetectContentType(m.www)
		// http.DetectContentType cant detect bzip2
		// https://github.com/golang/go/issues/32508
		if *m.MimeType == "application/octet-stream" {
			// try to dettect bzip
			// BZ (bzip magic) + h (huffman entropy)
			if bytes.HasPrefix(m.www, []byte("\x42\x5A\x68")) {
				if m.www[3] >= 0x31 && m.www[3] <= 0x39 {
					*m.MimeType = "application/x-bzip2"
				}
			}
		}
	}
	m.www = append(m.www, p...)
	return len(p), nil
}

func (m *MimeDetectorWriter) Close() error {
	return nil
}
