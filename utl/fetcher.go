// Package utl (fetcher.go) :
// These methods are for retrieving data from URL with optimized concurrency limits.
package utl

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	json "github.com/goccy/go-json"
)

// RequestParams : Parameters for FetchAPI
type RequestParams struct {
	Method        string
	APIURL        string
	Data          io.Reader
	Contenttype   string
	ContentLength string
	ContentRange  string
	Accesstoken   string
	Dtime         int64
}

// Global HTTP Client optimized for multiplexed and concurrent I/O
var globalHTTPClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

// errHandlingFromFetch : Add error messages to Msgar.
func (p *FileInf) errHandlingFromFetch(body []byte) {
	var em map[string]interface{}
	json.Unmarshal(body, &em)
	if errorObj, ok := em["error"].(map[string]interface{}); ok {
		if errorsList, ok := errorObj["errors"].([]interface{}); ok && len(errorsList) > 0 {
			if erMsgBase2, ok := errorsList[0].(map[string]interface{}); ok {
				erCode := errorObj["code"].(float64)
				erLoc, _ := erMsgBase2["location"].(string)
				erMsg, _ := erMsgBase2["message"].(string)
				p.Msgar = append(p.Msgar, fmt.Sprintf("Status code is %d, location is '%s', Error message is '%s'.", int(erCode), erLoc, erMsg))
			}
		}
	}
}

// FetchAPI : For fetching data to URL.
func (r *RequestParams) FetchAPI() ([]byte, error) {
	req, err := http.NewRequest(r.Method, r.APIURL, r.Data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", r.Contenttype)
	req.Header.Set("Authorization", "Bearer "+r.Accesstoken)

	client := globalHTTPClient
	client.Timeout = time.Duration(r.Dtime) * time.Second

	res, err := client.Do(req)
	if err != nil || res.StatusCode-300 >= 0 {
		var msg []byte
		var er string
		if res == nil {
			msg = []byte(err.Error())
			er = err.Error()
		} else {
			errmsg, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			msg = errmsg
			er = "Status Code: " + strconv.Itoa(res.StatusCode)
		}
		return msg, errors.New(er)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return body, err
}

// FetchAPIRaw : For fetching data to URL. Raw data (http.Response) from API is returned.
func (r *RequestParams) FetchAPIRaw() (*http.Response, error) {
	req, err := http.NewRequest(r.Method, r.APIURL, r.Data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", r.Contenttype)
	req.Header.Set("Authorization", "Bearer "+r.Accesstoken)

	client := globalHTTPClient
	client.Timeout = time.Duration(r.Dtime) * time.Second

	res, err := client.Do(req)
	if err != nil || res.StatusCode-300 >= 0 {
		var er string
		if res == nil {
			er = err.Error()
		} else {
			er = "Status Code: " + strconv.Itoa(res.StatusCode)
		}
		return res, errors.New(er)
	}
	return res, err
}

// FetchAPIres : For fetching data to URL.
func (r *RequestParams) FetchAPIres() (*http.Response, error) {
	req, err := http.NewRequest(r.Method, r.APIURL, r.Data)
	if err != nil {
		return nil, err
	}
	if len(r.ContentLength) > 0 {
		req.Header.Set("Content-Length", r.ContentLength)
	}
	if len(r.ContentRange) > 0 {
		req.Header.Set("Content-Range", r.ContentRange)
	}
	if len(r.Contenttype) > 0 {
		req.Header.Set("Content-Type", r.Contenttype)
	}
	if len(r.Accesstoken) > 0 {
		req.Header.Set("Authorization", "Bearer "+r.Accesstoken)
	}

	client := globalHTTPClient
	client.Timeout = time.Duration(r.Dtime) * time.Second

	return client.Do(req)
}
