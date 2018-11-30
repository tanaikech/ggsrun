// Package utl (searchfiles.go) :
// These methods are for searching files in Google Drive using search query and regex.
package utl

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// useRegex : Retrieve files using regex.
func (p *FileInf) useRegex(fm fileListSt) {
	re := regexp.MustCompile(p.SearchRegex)
	for _, e := range fm.Files {
		if re.MatchString(e.Name) {
			p.SearchedFiles = append(p.SearchedFiles, e)
		}
	}
}

// getFields : Get fields.
func (p *FileInf) getFields() string {
	if p.SearchFields == "" {
		return "files(id,mimeType,name),nextPageToken"
	}
	if !strings.Contains(p.SearchFields, "nextPageToken") {
		p.SearchFields += ",nextPageToken"
	}
	if p.SearchRegex != "" {
		if !strings.Contains(p.SearchFields, "name") {
			p.SearchFields = strings.Replace(p.SearchFields, "files(", "files(name,", 1)
		}
	}
	return p.SearchFields
}

// getQuery : Get search query.
func (p *FileInf) getQuery() string {
	if p.SearchQuery != "" {
		return p.SearchQuery
	}
	return ""
}

// SearchFiles : Searching files on Google Drive using query and regex.
func (p *FileInf) SearchFiles() *FileInf {
	q := p.getQuery()
	fields := p.getFields()
	fm := p.GetListLoop(q, fields)
	if p.SearchRegex != "" {
		p.useRegex(fm)
	} else {
		p.SearchedFiles = fm.Files
	}
	p.SearchedResult = fmt.Sprintf("Number of file information is %d.", len(p.SearchedFiles))
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}
