// Package utl (fetcher.go) :
// These methods are for retrieving data from URL.
package utl

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// RequestParams : Parameters for FetchAPI
type RequestParams struct {
	Method      string
	APIURL      string
	Data        io.Reader
	Contenttype string
	Accesstoken string
	Dtime       int64
}

// FetchAPI : For fetching data to URL.
func (r *RequestParams) FetchAPI() ([]byte, error) {
	req, err := http.NewRequest(
		r.Method,
		r.APIURL,
		r.Data,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", r.Contenttype)
	req.Header.Set("Authorization", "Bearer "+r.Accesstoken)
	client := &http.Client{
		Timeout: time.Duration(r.Dtime) * time.Second,
	}
	res, err := client.Do(req)
	if err != nil || res.StatusCode-300 >= 0 {
		var msg []byte
		var er string
		if res == nil {
			msg = []byte(err.Error())
			er = err.Error()
		} else {
			errmsg, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v. ", err)
				os.Exit(1)
			}
			msg = errmsg
			er = "Status Code: " + strconv.Itoa(res.StatusCode)
		}
		return msg, errors.New(er)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return body, err
}
