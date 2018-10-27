// Package utl (resumableupload.go) :
// These methods are for Resumable Upload to Google Drive.
package utl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	pb "gopkg.in/cheggaaa/pb.v1"
)

const (
	resumableUrl = "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable"
)

// chunkPot: Structure for chunks.
type chunkPot struct {
	Total  int64
	Chunks []chunk
}

// chunk: Structure for a chunk.
type chunk struct {
	StartByte int64
	EndByte   int64
	NumByte   int64
}

// initResumableUpload: Initializing Resumable upload. Retrieving "Location".
func (p *FileInf) initResumableUpload(metadata map[string]interface{}) string {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "id,mimeType,name,parents")
	meta, _ := json.Marshal(metadata)
	r := &RequestParams{
		Method:      "POST",
		APIURL:      resumableUrl + "&" + tokenparams.Encode(),
		Data:        bytes.NewBuffer(meta),
		Contenttype: "application/json; charset=UTF-8",
		Accesstoken: p.Accesstoken,
		Dtime:       10,
	}
	res, err := r.FetchAPIres()
	if res.StatusCode != 200 || err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n%v\n", err, res)
		os.Exit(1)
	}
	return res.Header["Location"][0]
}

// pullDataFromFile : Pulling data from file by address and size.
func pullDataFromFile(fileBytes []byte, st, en int64) []byte {
	var startE, numE int64
	startE = st // Start is 0
	numE = en   // If this value is 0, it means end of the slice.
	endE := func(s, n int64, fileBytes []byte) int64 {
		if n > 0 {
			return s + n
		}
		return int64(len(fileBytes))
	}(startE, numE, fileBytes)
	return fileBytes[startE:endE]
}

// int64ToStr : Convert from int64 to string.
func int64ToStr(i64 int64) string {
	return strconv.FormatInt(i64, 10)
}

// getChunks : Return chunk data as an object.
// When the filesize is more than p.ChunkSize, the chunk size is p.ChunkSize.
// When the filesize is less than p.ChunkSize, the chunk size is the filesize. This means the single request.
func (p *FileInf) getChunks(size int64) *chunkPot {
	cP := &chunkPot{
		Total: size,
	}
	var numE int64
	if size > p.ChunkSize {
		// multiple request
		var repeat int64
		numE = p.ChunkSize
		endS := func(f, n int64) int64 {
			c := f % n
			if c == 0 {
				return 0
			}
			return c
		}(size, numE)
		repeat = size / numE
		var startAddress int64
		for i := 0; i <= int(repeat); i++ {
			startAddress = int64(i) * numE
			c := &chunk{}
			c.StartByte = startAddress
			if i < int(repeat) {
				c.EndByte = startAddress + numE - 1
				c.NumByte = numE
				cP.Chunks = append(cP.Chunks, *c)
			} else if i == int(repeat) && endS > 0 {
				c.EndByte = startAddress + endS - 1
				c.NumByte = endS
				cP.Chunks = append(cP.Chunks, *c)
			}
		}
	} else {
		// single request
		numE = size
		ck := &chunk{
			StartByte: 0,
			EndByte:   numE - 1,
			NumByte:   numE,
		}
		cP.Chunks = append(cP.Chunks, *ck)
	}
	return cP
}

// resToBody : Retrieve body from http response.
func resToBody(res *http.Response) []byte {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	return body
}

// resumableSingleRequest : Run sinble request.
func (cP *chunkPot) resumableSingleRequest(location string, fileBytes []byte, metadata map[string]interface{}) ([]byte, error) {
	r := &RequestParams{
		Method:        "PUT",
		APIURL:        location,
		Data:          bytes.NewBuffer(fileBytes),
		Contenttype:   metadata["mimeType"].(string),
		ContentLength: int64ToStr(cP.Total),
		Dtime:         300,
	}
	res, err := r.FetchAPIres()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	body := resToBody(res)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%v\n%v", err, string(body))
	}
	defer res.Body.Close()
	return body, nil
}

// resumableMultipleRequest : Run multiple request.
func (cP *chunkPot) resumableMultipleRequest(location string, fileBytes []byte, metadata map[string]interface{}) ([]byte, error) {
	bar := pb.StartNew(len(cP.Chunks))
	for _, e := range cP.Chunks {
		bar.Increment()
		r := &RequestParams{
			Method:        "PUT",
			APIURL:        location,
			Data:          bytes.NewBuffer(pullDataFromFile(fileBytes, e.StartByte, e.NumByte)),
			Contenttype:   metadata["mimeType"].(string),
			ContentLength: int64ToStr(e.NumByte),
			ContentRange:  "bytes " + int64ToStr(e.StartByte) + "-" + int64ToStr(e.EndByte) + "/" + int64ToStr(cP.Total),
			Dtime:         60,
		}
		res, err := r.FetchAPIres()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		switch res.StatusCode {
		case 308:
			continue
		case 200:
			bar.FinishPrint("Done.")
			return resToBody(res), nil
		default:
			return nil, fmt.Errorf("%v\n%v", err, string(resToBody(res)))
		}
	}
	return nil, nil
}

// ResumableUpload : Main method of Resumable upload.
func (p *FileInf) ResumableUpload(metadata map[string]interface{}, fs *os.File, fstatus os.FileInfo) []byte {
	var err error
	resUp := p.getChunks(fstatus.Size())
	location := p.initResumableUpload(metadata)
	fileBytes, err := ioutil.ReadAll(fs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer fs.Close()
	var r []byte
	if len(resUp.Chunks) == 1 {
		r, err = resUp.resumableSingleRequest(location, fileBytes, metadata)
	} else {
		r, err = resUp.resumableMultipleRequest(location, fileBytes, metadata)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n%v\n", err, string(r))
		os.Exit(1)
	}
	return r
}
