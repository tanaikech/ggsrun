// Package utl (getdriveinformation.go) :
// These method is for retrieving Drive Information.
package utl

import (
	"fmt"
	"net/url"
	"os"
	"path"
)

// getDriveInf : Get drive information using Drive API.
func (p *FileInf) getDriveInf() error {
	u, err := url.Parse(driveapiv3)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "about")
	q := u.Query()
	q.Set("fields", p.SearchFields)
	u.RawQuery = q.Encode()
	r := &RequestParams{
		Method:      "GET",
		APIURL:      u.String(),
		Data:        nil,
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	p.reqAndGetRawResponse(r)
	return nil
}

// GetDriveInformation : Get Drive Information.
func (p *FileInf) GetDriveInformation() *FileInf {
	if err := p.getDriveInf(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return p
}
