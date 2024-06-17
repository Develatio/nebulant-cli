package downloader

import (
	"compress/bzip2"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/develatio/nebulant-cli/cast"
)

// 1500MB zip file size limit (a reasonable, but arbitrary value)
const maxZipFileSize = 1500 * 1024 * 1024

var httpclient *http.Client

func GetHttpClient() *http.Client {
	if httpclient != nil {
		return httpclient
	}
	tr := &http.Transport{
		MaxIdleConnsPerHost: 30,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		},
		// ReadIdleTimeout: 30 * time.Second,
		DisableCompression: true,
	}
	httpclient = &http.Client{Transport: tr}
	return httpclient
}

func DownloadFileWithProgressBar(url string, outfilepath string, msg string) error {
	// TODO
	startTime := time.Now()
	cast.LogDebug("Downloading "+url, nil)
	client := GetHttpClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "nebulant")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(http.StatusText(resp.StatusCode))
	}

	file, err := os.OpenFile(outfilepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304 -- Not a file inclusion, just a well know otput file
	if err != nil {
		return err
	}
	defer file.Close()
	bar := cast.NewProgress(resp.ContentLength, msg, "", "", "", "")
	ioreader := resp.Body
	if strings.ToLower(resp.Header.Get("Content-Type")) == "application/x-bzip2" {
		bz2dec := bzip2.NewReader(resp.Body)
		// limit bzip file size, LimitReader will launch
		// EOF on read > maxZipFileSize
		ioreader = io.NopCloser(io.LimitReader(bz2dec, maxZipFileSize))
	}

	var buf []byte
	_, err = io.CopyBuffer(io.MultiWriter(file, bar), ioreader, buf)
	if err != nil {
		return err
	}

	elapsedTime := time.Since(startTime).String()
	cast.LogDebug("downloaded in "+elapsedTime, nil)
	return nil
}
