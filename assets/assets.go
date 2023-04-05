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

package assets

import (
	"compress/bzip2"
	"crypto/md5" //#nosec G501-- weak, but ok
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bhmj/jsonslice"
	"github.com/develatio/nebulant-cli/cast"
	"github.com/develatio/nebulant-cli/config"
	"github.com/develatio/nebulant-cli/term"
	"github.com/develatio/nebulant-cli/util"
)

type UpgradeStateType int

// 1500MB zip file size limit (a reasonable, but arbitrary value)
const maxZipFileSize = 1500 * 1024 * 1024
const INDEX_TOKEN_SIZE = 3
const SUB_INDEX_TOKEN_SIZE = 2
const USED_BY_SPA = "SPA"
const (
	UpgradeStateNone UpgradeStateType = iota
	UpgradeStateInProgress
	UpgradeStateInProgressWithErr
	UpgradeStateEndWithErr
	UpgradeStateEndOK
)

type index struct {
	Parts map[string]map[int64]bool
}

type tinyIndexItem struct {
	Token string `json:"token"`
}

type indexItem struct {
	Token string `json:"token"`
	Text  string `json:"text"`
	// Term     string `json:"term"`
	ByteInits []int64 `json:"byteinit"`
}

type indexList struct {
	Parts []*indexItem `json:"parts"`
}

type matchInfo struct {
	token string
	count int
}

type assetsState struct {
	// use this for generic state
	LastUpgradeState    UpgradeStateType `json:"last_upgrade_status"`
	CurrentUpgradeState UpgradeStateType `json:"-"`
	LastUpgradeDate     time.Time        `json:"last_upgrade_date"`
	// Leter to Marty McFly: for by-asset status, define array or map here
}

func (a *assetsState) setUpgradeState(us UpgradeStateType) {
	a.CurrentUpgradeState = us
}

func (a *assetsState) saveState() error {
	a.LastUpgradeState = a.CurrentUpgradeState
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(config.AppHomePath(), "assets", "state"), data, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (a *assetsState) IsUpgradeInProgress() bool {
	return a.CurrentUpgradeState == UpgradeStateInProgress || a.CurrentUpgradeState == UpgradeStateInProgressWithErr
}

func (a *assetsState) NeedUpgrade() bool {
	if a.IsUpgradeInProgress() {
		return false
	}
	if a.LastUpgradeState == UpgradeStateNone {
		return true
	}
	if a.LastUpgradeState == UpgradeStateEndWithErr {
		return true
	}
	return false
	// TODO: test last date upgrade
	//return false
}

func (a *assetsState) loadState() error {
	jsonFile, err := os.Open(filepath.Join(config.AppHomePath(), "assets", "state")) //#nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		if os.IsNotExist(err) {
			a.CurrentUpgradeState = UpgradeStateNone
			a.LastUpgradeState = UpgradeStateNone
		} else {
			return err
		}
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, a); err != nil {
		return err
	}
	return nil
}

type AssetDefinition struct {
	Name         string
	FreshItem    func() interface{}
	LookPath     []string
	IndexPath    string
	SubIndexPath string
	FilePath     string
}

type SearchRequest struct {
	SearchTerm string
	Limit      int
	Offset     int
	Sort       string
}

type SearchResult struct {
	Count   int           `json:"count"`
	Results []interface{} `json:"results"`
}

type AssetRemoteDescription struct {
	ID      string `json:"id"`
	UsedBy  string `json:"used_by"`
	Hash    string `json:"hash"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

var AssetsDefinition map[string]*AssetDefinition = make(map[string]*AssetDefinition)

// var CurrentUpgradeState UpgradeStateType
// var LastUpgradeState UpgradeStateType
var State *assetsState

func (s *SearchRequest) Validate() (bool, error) {
	if !strings.HasPrefix(s.Sort, "-$") && !strings.HasPrefix(s.Sort, "$") {
		return false, fmt.Errorf("please, use $ or -$ at the beginning of the sort attr " + (s.Sort))
	}
	return true, nil
}

// func Poc() error {
// 	imgassetdef := &AssetDefinition{
// 		IndexPath:    "/tmp/test.idx",
// 		SubIndexPath: "/tmp/test.subidx",
// 		FilePath:     "./images.json",
// 		FreshItem:    func() interface{} { return &ec2.Image{} },
// 		LookPath: []string{
// 			"$.Architecture",
// 			"$.Name",
// 			"$.Description",
// 			"$.BlockDeviceMappings[].Ebs.VolumeType",
// 			"$.BlockDeviceMappings[].Ebs.SnapshotId",
// 			"$.ImageId",
// 			"$.ImageLocation",
// 		},
// 	}

// 	// _, err := makeIndex(imgassetdef)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// _, err = makeSubIndex(imgassetdef)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	srchres, err := Search(&SearchRequest{SearchTerm: "2.6.31-300-ec2-v-2.6.31-300.6-kernel.img.manifest.xml"}, imgassetdef)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("srchres", srchres)
// 	return nil
// }

func writeIndexFile(fpath string, list *index, name string) (int, error) {
	nn := 0
	file, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec G304-- file inclusion from var needed
	if err != nil {
		return 0, err
	}
	defer file.Close()

	n, err := file.Write([]byte(`{"parts":[`))
	if err != nil {
		return nn, err
	}
	nn = n + nn
	partcount := 0
	partlen := len(list.Parts)
	sb := term.OpenStatusBar()
	bar := sb.GetProgressBar(int64(partlen), " Writing "+name+" index file", false)
	for tkn, positions := range list.Parts {
		err := bar.Add(1)
		if err != nil {
			cast.LogWarn("progress bar err "+err.Error(), nil)
		}
		partcount++
		partsep := ""
		if partcount < partlen {
			partsep = ","
		}
		n, err := file.Write([]byte(`{"token":"` + tkn + `","byteinit":[`))
		if err != nil {
			return nn, err
		}
		nn = n + nn

		poscount := 0
		poslen := len(positions)
		for pos := range positions {
			poscount++
			possep := ""
			if poscount < poslen {
				possep = ","
			}
			n, err = file.Write([]byte(strconv.FormatInt(pos, 10) + possep))
			if err != nil {
				return nn, err
			}
			nn = n + nn
		}

		n, err = file.Write([]byte(`]}` + partsep))
		if err != nil {
			return nn, err
		}
		nn = n + nn
	}

	n, err = file.Write([]byte(`]}`))
	if err != nil {
		return nn, err
	}
	nn = n + nn
	return nn, nil
}

func recoverFilePosition(filepath string, position int64, container interface{}) error {
	f, err := os.Open(filepath) //#nosec G304-- file inclusion from var needed
	if err != nil {
		return err
	}
	defer f.Close() //#nosec G104 -- Unhandle is OK here
	if _, err = f.Seek(position, 0); err != nil {
		return err
	}
	var i int64
	var rtmp []byte
	for i = 0; i < 100; i++ {
		if _, err := f.Seek(position+i, 0); err != nil {
			return err
		}
		rtmp = make([]byte, 100)
		if _, err := f.Read(rtmp); err != nil {
			return err
		}
		if _, err := f.Seek(position+i, 0); err != nil {
			return err
		}
		if rtmp[0] == []byte("{")[0] {
			break
		}
	}
	dec := json.NewDecoder(f)
	err = dec.Decode(container)
	if err != nil {
		return err
	}
	return nil
}

func buildtokens(aa string, size int) ([]string, error) {
	var r []string
	reg, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		return r, err
	}
	bb := reg.ReplaceAllString(aa, "")

	for _, cc := range strings.Split(bb, " ") {
		t := strings.ToLower(strings.Trim(cc, " "))
		if len(t) <= 0 {
			continue
		}
		if len(t) <= size {
			r = append(r, t)
			continue
		}
		for i := 0; i+size <= len(t); i++ {
			r = append(r, t[i:i+size])
		}
	}
	return r, nil
}

func splitstring(aa string, size int) ([]string, error) {
	var r []string
	var rr map[string]bool = make(map[string]bool)

	reg, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		return r, err
	}
	bb := reg.ReplaceAllString(aa, "")

	for _, cc := range strings.Split(bb, " ") {
		t := strings.ToLower(strings.Trim(cc, " "))
		if len(t) <= 1 {
			continue
		}
		if len(t) <= size {
			if _, exists := rr[t]; !exists {
				rr[t] = true
				r = append(r, t)
			}
			continue
		}
		for i := 0; i+size <= len(t); i = i + size {
			if _, exists := rr[t[i:i+size]]; !exists {
				rr[t[i:i+size]] = true
				r = append(r, t[i:i+size])
			}
		}
		if r[len(r)-1] != cc[len(cc)-3:] {
			if _, exists := rr[cc[len(cc)-3:]]; !exists {
				rr[cc[len(cc)-3:]] = true
				r = append(r, cc[len(cc)-3:])
			}
		}
	}
	return r, nil
}

func logStats(prefix string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	cast.LogDebug(fmt.Sprintf("%s: Alloc: %v MiB\tHeapInuse: %v MiB\tFrees: %v MiB\tSys: %v MiB\tNumGC: %v", prefix, m.Alloc/1024/1024, m.HeapInuse/1024/1024, m.Frees/1024/1024, m.Sys/1024/1024, m.NumGC), nil)
}

func b2unzipWithProgressBar(file string, msg string) error {
	infile, err := os.OpenFile(file, os.O_RDONLY, 0600) // #nosec G304 -- Not a file inclusion, just a well know zipped file
	if err != nil {
		return err
	}
	defer infile.Close()

	outfile, err := os.OpenFile(file+"---.uztmp", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304 -- Not a file inclusion, just a well know otput file
	if err != nil {
		return err
	}
	defer outfile.Close()

	instats, err := infile.Stat()
	if err != nil {
		return err
	}

	if instats.Size() > maxZipFileSize {
		return fmt.Errorf("the downloaded zip file is too big")
	}

	bz2dec := bzip2.NewReader(infile)
	// limit bzip file size, LimitReader will launch
	// EOF on read > maxZipFileSize
	noZipBombReader := io.LimitReader(bz2dec, maxZipFileSize)

	sb := term.OpenStatusBar()
	bar := sb.GetProgressBar(instats.Size(), msg, false)

	_, err = io.Copy(io.MultiWriter(outfile, bar), noZipBombReader)
	if err != nil {
		return err
	}

	err = outfile.Close()
	if err != nil {
		return err
	}

	err = infile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(file+"---.uztmp", file)
	if err != nil {
		return err
	}

	return nil
}

func downloadFileWithProgressBar(url string, outputfile string, msg string) error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(http.StatusText(resp.StatusCode))
	}

	file, err := os.OpenFile(outputfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304 -- Not a file inclusion, just a well know otput file
	if err != nil {
		return err
	}
	defer file.Close()

	sb := term.OpenStatusBar()
	bar := sb.GetProgressBar(resp.ContentLength, msg, true)

	mimedetector := &util.MimeDetectorWriter{}

	_, err = io.Copy(io.MultiWriter(file, bar, mimedetector), resp.Body)
	if err != nil {
		return err
	}

	if mimedetector.MimeType != nil && *mimedetector.MimeType == "application/x-bzip2" {
		return b2unzipWithProgressBar(outputfile, "b2unzip")
	}
	return nil
}

func downloadAsset(remotedef *AssetRemoteDescription, localdef *AssetDefinition) error {
	err := os.MkdirAll(filepath.Dir(localdef.FilePath), os.ModePerm)
	if err != nil {
		return err
	}

	err = downloadFileWithProgressBar(remotedef.URL+".bz2", localdef.FilePath, " Downloading asset descriptor...")
	if err != nil {
		cast.LogDebug("Err on bz2 asset descriptor download: "+err.Error(), nil)
		err = downloadFileWithProgressBar(remotedef.URL, localdef.FilePath, " Downloading asset descriptor...")
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadIndex(remotedef *AssetRemoteDescription, localdef *AssetDefinition) error {
	err := os.MkdirAll(filepath.Dir(localdef.FilePath), os.ModePerm)
	if err != nil {
		return err
	}

	// download index
	err = downloadFileWithProgressBar(remotedef.URL+".idx.bz2", localdef.IndexPath, " Downloading index...")
	if err != nil {
		cast.LogDebug("Err on bz2 index download: "+err.Error(), nil)
		err = downloadFileWithProgressBar(remotedef.URL+".idx", localdef.IndexPath, " Downloading index...")
		if err != nil {
			return err
		}
	}

	// download subindex
	err = downloadFileWithProgressBar(remotedef.URL+".subidx.bz2", localdef.SubIndexPath, " Downloading subindex...")
	if err != nil {
		cast.LogDebug("Err on bz2 subindex download: "+err.Error(), nil)
		err = downloadFileWithProgressBar(remotedef.URL+".subidx", localdef.SubIndexPath, " Downloading subindex...")
		if err != nil {
			return err
		}
	}

	return nil
}

func makeIndex(assetdef *AssetDefinition) (int, error) {
	cast.LogInfo("Building Asset index "+assetdef.Name, nil)
	n, err := makeMainIndex(assetdef)
	if err != nil {
		return n, err
	}
	nn, err := makeSubIndex(assetdef)
	if err != nil {
		return n + nn, err
	}
	return n + nn, nil
}

// makeMainIndex func TODO: explore ways to not store the entire file into ram
func makeMainIndex(assetdef *AssetDefinition) (int, error) {
	var count int64 = 0
	idx := &index{}
	idx.Parts = make(map[string]map[int64]bool)

	// open data file
	// expected json structure:
	//      {
	//        "asdf": [
	//          {dataitem},
	//          {dataitem},
	//          ...
	//        ]
	//      }
	input, err := os.Open(assetdef.FilePath)
	if err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}
	defer input.Close()

	fi, err := input.Stat()
	if err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}

	sb := term.OpenStatusBar()
	bar := sb.GetProgressBar(fi.Size(), " Reading asset items", false)

	dec := json.NewDecoder(input)

	// read {
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}
	// read attr name
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}
	// read [
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}

	start := time.Now()
	readed := dec.InputOffset()
	err = bar.Add64(readed)
	if err != nil {
		cast.LogWarn("progress bar err "+err.Error(), nil)
	}
	// loop arr values
	for dec.More() {
		m := assetdef.FreshItem()

		// store file position of current item
		byteinit := dec.InputOffset()
		delta := byteinit - readed
		readed = byteinit
		err := bar.Add64(delta)
		if err != nil {
			cast.LogWarn("progress bar err "+err.Error(), nil)
		}

		err = dec.Decode(m)
		if err != nil {
			return 0, fmt.Errorf("MainIndex:" + err.Error())
		}

		if time.Since(start) > 10*time.Second {
			logStats("Building main index")
			start = time.Now()
		}

		vof := reflect.ValueOf(m)
		var terms []string
		getStrings("$", vof, &terms, assetdef.LookPath, make(map[interface{}]bool))
		text := strings.Join(terms, " ")
		ss, err := buildtokens(text, INDEX_TOKEN_SIZE)
		if err != nil {
			return 0, fmt.Errorf("MainIndex:" + err.Error())
		}
		for _, t := range ss {
			if _, exists := idx.Parts[t]; !exists {
				idx.Parts[t] = make(map[int64]bool)
			}
			idx.Parts[t][byteinit] = true
		}
		count++
	}

	// read }
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}

	n, err := writeIndexFile(assetdef.IndexPath+".tmp", idx, assetdef.Name)
	if err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}
	if err := os.Rename(assetdef.IndexPath+".tmp", assetdef.IndexPath); err != nil {
		return 0, err
	}
	return n, nil
}

func makeSubIndex(assetdef *AssetDefinition) (int, error) {
	idx := &index{}
	idx.Parts = make(map[string]map[int64]bool)

	// Open the file of the index
	input, err := os.Open(assetdef.IndexPath)
	if err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}
	defer input.Close()
	fi, err := input.Stat()
	if err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}

	sb := term.OpenStatusBar()
	bar := sb.GetProgressBar(fi.Size(), " Optimizing "+assetdef.Name+" index", false)
	dec := json.NewDecoder(input)

	// read {
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}
	// read attr name
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}
	// read [
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}

	start := time.Now()
	readed := dec.InputOffset()
	err = bar.Add64(readed)
	if err != nil {
		cast.LogWarn("progress bar err "+err.Error(), nil)
	}
	// while the array contains values
	for dec.More() {
		var m tinyIndexItem
		// store current item file position
		byteinit := dec.InputOffset()
		delta := byteinit - readed
		readed = byteinit
		err := bar.Add64(delta)
		if err != nil {
			cast.LogWarn("progress bar err "+err.Error(), nil)
		}

		err = dec.Decode(&m)
		if err != nil {
			return 0, fmt.Errorf("SubIndex:" + err.Error())
		}

		if time.Since(start) > 10*time.Second {
			logStats("Building subindex")
			start = time.Now()
		}

		// skip too small tokens
		if len(m.Token) < SUB_INDEX_TOKEN_SIZE {
			continue
		}

		tkn := string(m.Token[0:SUB_INDEX_TOKEN_SIZE])
		if _, exists := idx.Parts[tkn]; !exists {
			idx.Parts[tkn] = make(map[int64]bool)
		}
		idx.Parts[tkn][byteinit] = true
	}

	// read }
	if _, err := dec.Token(); err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}

	n, err := writeIndexFile(assetdef.SubIndexPath+".tmp", idx, "(sub)"+assetdef.Name)
	if err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}
	if err := os.Rename(assetdef.SubIndexPath+".tmp", assetdef.SubIndexPath); err != nil {
		return 0, err
	}
	return n, nil
}

func Search(sr *SearchRequest, assetdef *AssetDefinition) (*SearchResult, error) {
	if valid, err := sr.Validate(); !valid {
		return nil, err
	}
	searchres := &SearchResult{Count: 0}
	term := strings.ToLower(sr.SearchTerm)
	if len(term) <= 1 {
		return nil, fmt.Errorf("cannot lookup by requested term. Min char needed: 2")
	}
	subidx := &indexList{}
	subIdxFile, _ := os.Open(assetdef.SubIndexPath)
	defer subIdxFile.Close()

	bv, err := io.ReadAll(subIdxFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bv, subidx); err != nil {
		return nil, err
	}

	// Spit search term in parts of size 3
	// searchterm -> sea rch ter erm
	schtkns, err := splitstring(term, 3)
	if err != nil {
		return nil, err
	}
	// with 5 search term the system can
	// found the items eficiently. Adding
	// more terms is innecesary and add
	// ram/cpu consumption.
	if len(schtkns) > 5 {
		schtkns = schtkns[0:5]
	}

	// From sub-index file, extract relevant byte num of
	// relevant positions of the index file.
	var idxpositions map[int64]bool = make(map[int64]bool)
	e := 0
	for _, stk := range schtkns {
		if len(stk) < 2 {
			continue
		}
		e++
		start := string(stk[0:2])
		for _, subidxitem := range subidx.Parts {
			if subidxitem.Token != start {
				continue
			}
			// Positions inside the index file that
			// match with the search term first token
			for _, ip := range subidxitem.ByteInits {
				idxpositions[ip] = true
			}
		}
	}

	if e <= 0 {
		return nil, fmt.Errorf("not enough alphanumeric characters. Min char needed: 2")
	}

	if len(idxpositions) <= 0 {
		// None results found
		return searchres, nil

	}

	// From index file, extract those
	// relevant realfile positions that
	// match with the search term
	idxf, err := os.Open(assetdef.IndexPath)
	if err != nil {
		return nil, err
	}
	defer idxf.Close()

	fpositions := make(map[int64]*matchInfo)
	// iterate over those index positions found
	// into the subindex file
	for idxpos := range idxpositions {
		if _, err := idxf.Seek(idxpos, 0); err != nil {
			return nil, err
		}
		// Adjust cursor to match "{" of the file
		// and prevent middle tokens like ","
		var i int64
		var r []byte
		for i = 0; i < 100; i++ {
			if _, err := idxf.Seek(idxpos+i, 0); err != nil {
				return nil, err
			}
			r = make([]byte, 100)
			if _, err := idxf.Read(r); err != nil {
				return nil, err
			}
			if _, err := idxf.Seek(idxpos+i, 0); err != nil {
				return nil, err
			}
			if r[0] == []byte("{")[0] {
				break
			}
		}

		// decode index data at idxpos position
		idxitem := &indexItem{}
		dec := json.NewDecoder(idxf)
		if err := dec.Decode(&idxitem); err != nil {
			return nil, err
		}

		// test if index data is relevant
		// to all search tokens,
		// discard if not
		for _, stk := range schtkns {
			stklen := len(stk)
			if stklen > len(idxitem.Token) {
				continue
			}
			if stk == idxitem.Token[:stklen] {
				// valid index data found due to match
				// between index token term and one of the
				// tokens inside schtkns.
				// Extract the positions of relevant
				// data into the original file
				for _, ps := range idxitem.ByteInits {
					if _, exists := fpositions[ps]; !exists {
						fpositions[ps] = &matchInfo{
							count: 1,
							token: idxitem.Token,
						}
						continue
					}
					// Store how many times the search token
					// match against found positions. A
					// valid position should match against
					// all of the search tokens
					fpositions[ps].count++
				}
				break
			}
		}
	}

	if len(fpositions) <= 0 {
		return searchres, nil
	}

	cast.LogDebug("Search Tokens: "+fmt.Sprintf("%v", schtkns), nil)

	schtknscount := len(schtkns)
	count := 0
	for position, minfo := range fpositions {
		// position should match all
		// search tokens discard if not.
		if minfo.count != schtknscount {
			continue
		}

		// Obtain a fresh item to store recovered data
		img := assetdef.FreshItem()
		if err := recoverFilePosition(assetdef.FilePath, position, img); err != nil {
			return nil, err
		}

		// obtain texts
		vof := reflect.ValueOf(img)
		var t []string
		getStrings("$", vof, &t, assetdef.LookPath, make(map[interface{}]bool))
		text := strings.ToLower(strings.Join(t, " "))

		// The last validation. Search the terms inside
		// the recovered data.
		valid := true
		for _, bb := range strings.Split(term, " ") {
			cc := strings.ToLower(strings.Trim(bb, " "))
			if !strings.Contains(text, cc) {
				// All terms should be found inside the
				// recovered data.
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		count++
		searchres.Results = append(searchres.Results, vof.Interface())
		if sr.Limit > 0 && len(sr.Sort) <= 0 && sr.Offset <= 0 && count == sr.Limit {
			cast.LogDebug("Breaking search by limit", nil)
			break
		}
	}
	cast.LogDebug("Found items "+fmt.Sprintf("%v", len(searchres.Results)), nil)

	if len(sr.Sort) > 1 {
		if sr.Sort[0] == []byte("-")[0] {
			cast.LogDebug("Sorting desc "+sr.Sort, nil)
			sortResults(searchres.Results, sr.Sort[1:], true)
		} else {
			cast.LogDebug("Sorting asc "+sr.Sort, nil)
			sortResults(searchres.Results, sr.Sort, false)
		}
	}

	if len(searchres.Results) > sr.Offset {
		if sr.Limit > 0 {
			cast.LogDebug(fmt.Sprintf("Select by offset %v + limit %v", sr.Offset, sr.Limit), nil)
			if sr.Offset+sr.Limit > len(searchres.Results) {
				searchres.Results = searchres.Results[sr.Offset:]
			} else {
				searchres.Results = searchres.Results[sr.Offset : sr.Offset+sr.Limit]
			}
		} else {
			cast.LogDebug("select by offset", nil)
			searchres.Results = searchres.Results[sr.Offset:]
		}
	} else {
		searchres.Results = nil
	}

	cast.LogDebug("Return items "+fmt.Sprintf("%v", len(searchres.Results)), nil)

	searchres.Count = count
	return searchres, nil
}

func sortResults(results []interface{}, sortterm string, inverse bool) {
	sort.SliceStable(results, func(i, j int) bool {
		enci, err := json.Marshal(results[i])
		if err != nil {
			panic(err)
		}
		ence, err := json.Marshal(results[j])
		if err != nil {
			panic(err)
		}
		vali, err := jsonslice.Get(enci, strings.TrimSpace(sortterm))
		if err != nil {
			panic(err)
		}
		val := string(vali)
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			val = strings.TrimSuffix(val, "\"")
			val = strings.TrimPrefix(val, "\"")
			vali = []byte(val)
		}

		vale, err := jsonslice.Get(ence, strings.TrimSpace(sortterm))
		if err != nil {
			panic(err)
		}
		val = string(vale)
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			val = strings.TrimSuffix(val, "\"")
			val = strings.TrimPrefix(val, "\"")
			vale = []byte(val)
		}

		if inverse {
			// TODO: handle ints?
			return string(vali) > string(vale)
		}
		// TODO: handle ints?
		return string(vali) < string(vale)
	})
}

// getStrings func
func getStrings(path string, v reflect.Value, st *[]string, lookPaths []string, il map[interface{}]bool) {
	// While v is a pointer or interface, keep calling v.Elem() to finally get
	// value that pointer point to or value inside interface
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr {
			if _, exists := il[v.Interface()]; exists {
				// Infinite loop throught pointer detected. Skip.
				return
			}
			// Prevent store nil pointers
			if !v.IsNil() {
				il[v.Interface()] = true
			}
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Invalid:
		return
	case reflect.Slice, reflect.Array:
		// Iterate over slice/array
		for i := 0; i < v.Len(); i++ {
			// Recursive call, append array index to path
			getStrings(path+"[]", v.Index(i), st, lookPaths, il)
		}
	case reflect.Struct:
		t := v.Type()
		// Iterate over struct fields
		for i := 0; i < t.NumField(); i++ {

			// skip unexported fields (first letter lowercase)
			if unicode.IsLower([]rune(t.Field(i).Name)[0]) {
				continue
			}

			// skipt attrs starting by _
			if t.Field(i).Name == "_" {
				continue
			}

			// Recursive call, append field name to path
			getStrings(path+"."+t.Field(i).Name, v.Field(i), st, lookPaths, il)
		}
	default:
		if len(lookPaths) > 0 {
			for _, lpth := range lookPaths {
				if lpth == path {
					*st = append(*st, fmt.Sprintf("%v", v.Interface()))
					break
				}
			}
		} else {
			*st = append(*st, fmt.Sprintf("%v", v.Interface()))
		}
	}
}

func updateDescriptor(descpath string) error {
	err := os.MkdirAll(filepath.Dir(descpath), os.ModePerm)
	if err != nil {
		return err
	}

	cast.LogInfo("Updating asset descriptor "+config.AssetDescriptorURL+"...", nil)
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(config.AssetDescriptorURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(http.StatusText(resp.StatusCode))
	}

	file, err := os.OpenFile(descpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) //#nosec G304-- file inclusion from var needed
	if err != nil {
		return err
	}
	defer file.Close()

	sb := term.OpenStatusBar()
	bar := sb.GetProgressBar(resp.ContentLength, " Downloading asset descriptor", true)

	_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)

	if err != nil {
		return err
	}
	return nil
}

func GenerateIndexFromFile(term string) error {
	terms := strings.Split(term, ":")
	// cast.LogInfo("Building index of asset "+terms[1]+" from file "+terms[0]+" into dir "+terms[2], nil)
	def, assetExists := AssetsDefinition[terms[1]]
	if !assetExists {
		return fmt.Errorf("unknown asset id")
	}

	def.FilePath = terms[0]
	def.IndexPath = filepath.Join(terms[2], filepath.Base(terms[0])+".idx")
	def.SubIndexPath = filepath.Join(terms[2], filepath.Base(terms[0])+".subidx")
	_, err := makeIndex(def)
	if err != nil {
		return err
	}
	cast.LogInfo("Index gen DONE", nil)
	return nil
}

func UpgradeAssets(force bool, skipdownload bool) error {
	State.setUpgradeState(UpgradeStateInProgress)
	err := State.saveState()
	if err != nil {
		// use erros.Join() when go 1.20
		return err
	}
	var descriptor []*AssetRemoteDescription
	descpath := filepath.Join(config.AppHomePath(), "assets", "descriptor.json")

	if err := updateDescriptor(descpath); err != nil {
		State.setUpgradeState(UpgradeStateEndWithErr)
		err2 := State.saveState()
		if err2 != nil {
			// use erros.Join() when go 1.20
			return err2
		}
		return err
	}

	descfile, err := os.Open(descpath) //#nosec G304 -- Not a file inclusion, just a json read
	if err != nil {
		State.setUpgradeState(UpgradeStateEndWithErr)
		return err
	}

	defer descfile.Close()
	defer term.CloseStatusBar()

	byteValue, _ := io.ReadAll(descfile)
	if err := json.Unmarshal(byteValue, &descriptor); err != nil {
		State.setUpgradeState(UpgradeStateEndWithErr)
		err2 := State.saveState()
		if err2 != nil {
			// use erros.Join() when go 1.20
			return err2
		}
		return err
	}

	for _, desc := range descriptor {
		if desc.UsedBy == USED_BY_SPA {
			continue
		}
		if def, exists := AssetsDefinition[desc.ID]; exists {
			if _, err := os.Stat(def.FilePath); err == nil {
				f, err := os.Open(def.FilePath)
				if err != nil {
					err2 := State.saveState()
					if err2 != nil {
						// use erros.Join() when go 1.20
						return err2
					}
					State.setUpgradeState(UpgradeStateEndWithErr)
					return err
				}
				defer f.Close()

				h := md5.New() //#nosec G401-- weak, but ok
				if _, err := io.Copy(h, f); err != nil {
					cast.LogErr("Cannot determine asset"+desc.ID+" integrity due to "+err.Error(), nil)
					State.setUpgradeState(UpgradeStateInProgressWithErr)
					continue
				}

				filemd5 := fmt.Sprintf("%x", h.Sum(nil))
				if force || filemd5 != desc.Hash {
					err = downloadAsset(desc, def)
					if err != nil {
						State.setUpgradeState(UpgradeStateInProgressWithErr)
						cast.LogErr("Cannot download asset "+desc.ID+" due to "+err.Error(), nil)
						continue
					}
					err := os.Remove(def.IndexPath)
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						State.setUpgradeState(UpgradeStateInProgressWithErr)
						cast.LogErr("Cannot purge old index file "+err.Error(), nil)
						continue
					}
					err = os.Remove(def.SubIndexPath)
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						State.setUpgradeState(UpgradeStateInProgressWithErr)
						cast.LogErr("Cannot purge old index file "+err.Error(), nil)
						continue
					}
				} else {
					cast.LogInfo("Asset file "+desc.ID+" up to date. No download needed", nil)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				err = downloadAsset(desc, def)
				if err != nil {
					State.setUpgradeState(UpgradeStateInProgressWithErr)
					cast.LogErr("Cannot download asset "+desc.ID+" due to "+err.Error(), nil)
					continue
				}
			} else {
				State.setUpgradeState(UpgradeStateEndWithErr)
				err2 := State.saveState()
				if err2 != nil {
					// use erros.Join() when go 1.20
					return err2
				}
				return err
			}

			if _, err := os.Stat(def.IndexPath); err == nil {
				cast.LogInfo("Index of "+desc.ID+" up to date", nil)
			} else if errors.Is(err, os.ErrNotExist) {
				cast.LogInfo("Downloading "+desc.ID+" index in bg... (from "+desc.URL+")", nil)
				err = nil
				if !skipdownload {
					err = downloadIndex(desc, def)
				}
				if err != nil {
					cast.LogWarn("Cannot download "+desc.ID+" index: ("+err.Error()+")", nil)
					cast.LogWarn(desc.ID+": Since the index could not be downloaded, it will be built locally. This process is long and expensive.", nil)
				}
				if err != nil || skipdownload {
					cast.LogInfo("Building "+desc.ID+" index in bg... (from "+desc.URL+")", nil)
					_, err := makeIndex(def)
					if err != nil {
						State.setUpgradeState(UpgradeStateInProgressWithErr)
						cast.LogErr("Cannot build index of "+desc.ID+" due to "+err.Error(), nil)
						continue
					}
					cast.LogInfo("Building index of "+desc.ID+"...DONE", nil)
				}
			} else {
				State.setUpgradeState(UpgradeStateInProgressWithErr)
				cast.LogErr("Cannot determine index status "+err.Error(), nil)
			}

		} else {
			State.setUpgradeState(UpgradeStateInProgressWithErr)
			cast.LogWarn("Unknown asset descriptor "+desc.ID, nil)
		}
	}

	if State.CurrentUpgradeState == UpgradeStateInProgressWithErr {
		State.setUpgradeState(UpgradeStateEndWithErr)
		err2 := State.saveState()
		if err2 != nil {
			// use erros.Join() when go 1.20
			return err2
		}
	} else if State.CurrentUpgradeState == UpgradeStateInProgress {
		State.setUpgradeState(UpgradeStateEndOK)
		err2 := State.saveState()
		if err2 != nil {
			// use erros.Join() when go 1.20
			return err2
		}
	}

	cast.LogInfo("All assets up to date", nil)
	return nil
}

func init() {
	State = &assetsState{}
	err := State.loadState()
	if err != nil {
		// should panic?
		fmt.Println(err.Error())
	}
}
