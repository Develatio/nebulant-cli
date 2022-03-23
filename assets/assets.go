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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/develatio/nebulant-cli/cast"
)

const INDEX_TOKEN_SIZE = 3
const SUB_INDEX_TOKEN_SIZE = 2
const USED_BY_SPA = "SPA"

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

type AssetDefinition struct {
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

// 	// _, err := MakeIndex(imgassetdef)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// _, err = MakeSubIndex(imgassetdef)
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

func writeIndexFile(fpath string, list *index) (int, error) {
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
	for tkn, positions := range list.Parts {
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
		if len(t) <= 0 {
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
	cast.LogInfo(fmt.Sprintf("%s: Alloc: %v MiB\tHeapInuse: %v MiB\tFrees: %v MiB\tSys: %v MiB\tNumGC: %v", prefix, m.Alloc/1024/1024, m.HeapInuse/1024/1024, m.Frees/1024/1024, m.Sys/1024/1024, m.NumGC), nil)
}

func DownloadAsset(url string, assetdef *AssetDefinition) error {
	err := os.MkdirAll(filepath.Dir(assetdef.FilePath), os.ModePerm)
	if err != nil {
		return err
	}

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

	file, err := os.OpenFile(assetdef.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func MakeIndex(assetdef *AssetDefinition) (int, error) {
	n, err := MakeMainIndex(assetdef)
	if err != nil {
		return n, err
	}
	nn, err := MakeSubIndex(assetdef)
	if err != nil {
		return n + nn, err
	}
	return n + nn, nil
}

// MakeMainIndex func TODO: explore ways to not store the entire file into ram
func MakeMainIndex(assetdef *AssetDefinition) (int, error) {
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
	// loop arr values
	for dec.More() {
		m := assetdef.FreshItem()

		// store file position of current item
		byteinit := dec.InputOffset()

		err := dec.Decode(m)
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

	n, err := writeIndexFile(assetdef.IndexPath+".tmp", idx)
	if err != nil {
		return 0, fmt.Errorf("MainIndex:" + err.Error())
	}
	if err := os.Rename(assetdef.IndexPath+".tmp", assetdef.IndexPath); err != nil {
		return 0, err
	}
	return n, nil
}

func MakeSubIndex(assetdef *AssetDefinition) (int, error) {
	idx := &index{}
	idx.Parts = make(map[string]map[int64]bool)

	// Open the file of the index
	input, err := os.Open(assetdef.IndexPath)
	if err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}
	defer input.Close()
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
	// while the array contains values
	for dec.More() {
		var m tinyIndexItem
		// store current item file position
		byteinit := dec.InputOffset()

		err := dec.Decode(&m)
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

	n, err := writeIndexFile(assetdef.SubIndexPath+".tmp", idx)
	if err != nil {
		return 0, fmt.Errorf("SubIndex:" + err.Error())
	}
	if err := os.Rename(assetdef.SubIndexPath+".tmp", assetdef.SubIndexPath); err != nil {
		return 0, err
	}
	return n, nil
}

func Search(sr *SearchRequest, assetdef *AssetDefinition) (*SearchResult, error) {
	searchres := &SearchResult{Count: 0}
	term := strings.ToLower(sr.SearchTerm)
	subidx := &indexList{}
	subIdxFile, _ := os.Open(assetdef.SubIndexPath)
	defer subIdxFile.Close()

	bv, err := ioutil.ReadAll(subIdxFile)
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
	for _, stk := range schtkns {
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
		// to all search token,
		// discard if not
		for _, stk := range schtkns {
			if stk == idxitem.Token {
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
		// WIP limit implementation
		// TODO: sort and pagination
		if sr.Limit > 0 && count >= sr.Limit {
			break
		}
	}

	searchres.Count = count
	return searchres, nil
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
