// Package utl (dlrevfile.go) :
// These methods are for retrieving revision ID and downloading revision files from Google Drive.
package utl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"
)

// revisionList : Struct for revision list
type revisionList struct {
	Revisions []struct {
		ID           string    `json:"id"`
		ModifiedTime time.Time `json:"modifiedTime"`
	}
}

// downloadRevisionFile : Download revision file using revision ID
func (p *FileInf) downloadRevisionFile() *FileInf {
	ext := strings.ToLower(p.WantExt)
	if len(ext) > 0 {
		p.DlMime = extToMime(ext)
	} else {
		p.DlMime, ext = defFormat(p.MimeType)
	}
	var dURLq string
	var gm map[string]interface{}
	json.Unmarshal([]byte(googlemimetypes), &gm)
	if gm["exportFormats"].(map[string]interface{})[p.MimeType] != nil {
		for _, e := range gm["exportFormats"].(map[string]interface{})[p.MimeType].([]interface{}) {
			if e == p.DlMime {
				p.SaveName = p.FileName + "." + ext
			}
		}
		if p.MimeType == "application/vnd.google-apps.script" {
			fmt.Println("Error: I think that the project file does NOT support revisions yet. If you know how to get revisions for projects using API, please tell me. I'm glad.")
			os.Exit(1)
		} else {
			params := url.Values{}
			params.Set("revisions", p.RevisionID)
			params.Set("mimeType", p.DlMime)
			dURLq = driveapiurl + p.FileID + "/export?" + params.Encode()
		}
	} else {
		p.SaveName = p.FileName
		dURLq = driveapiurl + p.FileID + "/revisions/" + p.RevisionID + "?alt=media"
	}
	p, _ = p.writeFile(dURLq)
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}

// GetRevisionList : Display revision IDs.
func (p *FileInf) GetRevisionList(c *cli.Context) *FileInf {
	if len(p.FileID) > 0 && len(c.String("download")) == 0 {
		params := url.Values{}
		params.Set("fields", "revisions(id,modifiedTime)")
		params.Set("pageSize", "1000")
		r := &RequestParams{
			Method:      "GET",
			APIURL:      driveapiurl + p.FileID + "/revisions?" + params.Encode(),
			Data:        nil,
			Contenttype: "application/x-www-form-urlencoded",
			Accesstoken: p.Accesstoken,
			Dtime:       10,
		}
		body, err := r.FetchAPI()
		var er dlError
		json.Unmarshal(body, &er)
		if err != nil || er.Error.Code-300 >= 0 {
			fmt.Print(fmt.Sprintf("Error: %s (Status code is %d)\n", er.Error.Message, er.Error.Code))
			os.Exit(1)
		}
		var rl revisionList
		json.Unmarshal(body, &rl)
		ar := rl.Revisions
		if len(ar) > 0 {
			buffer := &bytes.Buffer{}
			w := new(tabwriter.Writer)
			w.Init(buffer, 0, 4, 1, ' ', 0)
			fmt.Fprintf(w, "\n%s\t%s\n", "# Revision ID", "# ModifedTime")
			for _, e := range ar {
				fmt.Fprintf(w, "%s\t%s\n", e.ID, e.ModifiedTime.In(time.Local))
			}
			w.Flush()
			fmt.Printf("%s\n", buffer)
		}
		p.Msgar = append(p.Msgar, fmt.Sprintf("Revision ID list was retrieved."))
	}
	if len(p.FileID) > 0 && len(c.String("download")) > 0 {
		p.FileID = c.String("fileid")
		p.RevisionID = c.String("download")
		return p.GetFileinf().downloadRevisionFile()
	}
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}
