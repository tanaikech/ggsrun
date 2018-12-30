// Package utl (zip.go) :
// These methods are for zipping files.
package utl

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"log"
	"time"
)

// zipFileHeads : For doZip()
type zipFileHeads struct {
	Files []zipFileHead
}

// zipFileHead : For doZip()
type zipFileHead struct {
	Name     string
	Modified time.Time
	Method   uint16
	Body     []byte
}

// doFilesZip : Zipping *zipFileHeads.
func (f *zipFileHeads) doFilesZip(comment string) *bytes.Buffer {
	b := new(bytes.Buffer)
	z := zip.NewWriter(b)
	z.SetComment(comment)
	for _, file := range f.Files {
		fh := &zip.FileHeader{
			Name:     file.Name,
			Modified: file.Modified,
			Method:   file.Method,
		}
		zf, err := z.CreateHeader(fh)
		if err != nil {
			log.Fatal(err)
		}
		if _, err = zf.Write(file.Body); err != nil {
			log.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		log.Fatal(err)
	}
	return b
}

// zipComment : Add information of project to zip file as a comment.
func (p *FileInf) zipComment() string {
	commentStruct := struct {
		Name                   string `json:"projectName,omitempty"`
		ID                     string `json:"fileId,omitempty"`
		CreatedTime            string `json:"cretedTime,omitempty"`
		ModifiedTime           string `json:"modifiedTime,omitempty"`
		OwnerName              string `json:"ownerName,omitempty"`
		OwnerEmail             string `json:"ownerEmail,omitempty"`
		LastModifyingUserName  string `json:"lastModifyingUserName,omitempty"`
		LastModifyingUserEmail string `json:"lastModifyingUserEmail,omitempty"`
	}{
		p.FileName,
		p.FileID,
		func() string {
			if p.CreatedTime != nil {
				return p.CreatedTime.In(time.Local).Format("20060102 15:04:05 MST")
			}
			return ""
		}(),
		func() string {
			if p.CreatedTime != nil {
				return p.ModifiedTime.In(time.Local).Format("20060102 15:04:05 MST")
			}
			return ""
		}(),
		p.Owners[0].Name,
		p.Owners[0].Email,
		p.LastModifyingUser.Name,
		p.LastModifyingUser.Email,
	}
	comment, _ := json.Marshal(commentStruct)
	return string(comment)
}
