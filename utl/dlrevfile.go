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
	"path"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"
)

// revisionListOLD : Struct for revision list
type revisionList struct {
	Revisions []struct {
		ID           string    `json:"id,omitempty"`
		ModifiedTime time.Time `json:"modifiedTime,omitempty"`
	}
}

// revisionListv2 : Struct for revision list
type revisionListv2 struct {
	Items []struct {
		ID           string    `json:"id,omitempty"`
		ModifiedDate time.Time `json:"modifiedDate,omitempty"`
		ExportLinks  struct {
			// for spreadsheet
			Elsx string `json:"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,omitempty"`
			CSV  string `json:"text/csv,omitempty"`
			// for document
			HTML string `json:"text/html,omitempty"`
			Docx string `json:"application/vnd.openxmlformats-officedocument.wordprocessingml.document,omitempty"`
			// for slide
			Pptx string `json:"application/vnd.openxmlformats-officedocument.presentationml.presentation,omitempty"`
			// for drawing
			Svg string `json:"image/svg+xml,omitempty"`
			PNG string `json:"image/png,omitempty"`
			Jpg string `json:"image/jpeg,omitempty"`

			Text string `json:"text/plain,omitempty"`
			PDF  string `json:"application/pdf,omitempty"`
			ZIP  string `json:"application/zip,omitempty"`
		} `json:"exportLinks,omitempty"`
	} `json:"items,omitempty"`
}

// GetRevisionList : Display revision IDs.
func (p *FileInf) GetRevisionList(c *cli.Context) *FileInf {
	if c.String("fileid") == "" && c.String("download") == "" && c.String("createversion") == "" {
		p.Msgar = append(p.Msgar, "Error: No options. Please check HELP using 'ggsrun r --help'.")
	} else {
		p.GetFileinf()
		if p.MimeType == "application/vnd.google-apps.spreadsheet" ||
			p.MimeType == "application/vnd.google-apps.document" ||
			p.MimeType == "application/vnd.google-apps.presentation" ||
			p.MimeType == "application/vnd.google-apps.drawing" {
			p.getRevFromGoogleDocs(c)
		} else if p.MimeType == "application/vnd.google-apps.script" || len(p.FileID) == lengthOfProjectId {
			p.versionForProject(c)
		} else {
			p.getRevFromExGoogleDocs(c)
		}
	}
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}

// versionForProject : Manage versions for project
func (p *FileInf) versionForProject(c *cli.Context) *FileInf {
	if c.String("createversion") == "" {
		pvl := p.getProjectVersionListInit()
		if c.String("download") == "" {
			p.dispProjectVersionList(pvl)
		} else {
			p.FileID = c.String("fileid")
			p.RevisionID = c.String("download")
			u, _ := url.Parse(appsscriptapi)
			u.Path = path.Join(u.Path, p.FileID+"/content")
			params := url.Values{}
			params.Set("versionNumber", p.RevisionID)
			verUrl := u.String() + "?" + params.Encode()
			p, body := p.writeFile(verUrl)
			p.saveScript(body, c)
		}
	} else {
		p.createProjectVersion(c.String("createversion"))
	}
	return p
}

// getVerFromProject : Retrieve version from project
func (p *FileInf) dispProjectVersionList(pvl *projectVersionList) {
	ar := pvl.Versions
	if len(ar) > 0 {
		buffer := &bytes.Buffer{}
		w := new(tabwriter.Writer)
		w.Init(buffer, 0, 4, 1, ' ', 0)
		fmt.Fprintf(w, "\n%s\t%s\t%s\n", "# versionNumber", "# description", "# createTime")
		for _, e := range ar {
			fmt.Fprintf(w, "%d\t%s\t%s\n", e.VersionNumber, e.Description, e.CreateTime.In(time.Local).Format("20060102_15:04:05"))
		}
		w.Flush()
		fmt.Printf("%s\n", buffer)
	}
	p.Msgar = append(p.Msgar, fmt.Sprintf("Version list of '%s' was retrieved.", p.FileName))
}

// getRevFromGoogleDocs : Display revision IDs from Google Docs.
func (p *FileInf) getRevFromGoogleDocs(c *cli.Context) {
	if len(p.FileID) > 0 {
		params := url.Values{}
		params.Set("fields", "items(exportLinks,id,modifiedDate)")
		params.Set("maxResults", "1000")
		r := &RequestParams{
			Method:      "GET",
			APIURL:      driveapiurlv2 + p.FileID + "/revisions?" + params.Encode(),
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
		var rl revisionListv2
		json.Unmarshal(body, &rl)
		if len(c.String("download")) == 0 {
			ar := rl.Items
			if len(ar) > 0 {
				buffer := &bytes.Buffer{}
				w := new(tabwriter.Writer)
				w.Init(buffer, 0, 4, 1, ' ', 0)
				fmt.Fprintf(w, "\n%s\t%s\n", "# Revision ID", "# ModifedTime")
				for _, e := range ar {
					fmt.Fprintf(w, "%s\t%s\n", e.ID, e.ModifiedDate.In(time.Local))
				}
				w.Flush()
				fmt.Printf("%s\n", buffer)
			}
			p.Msgar = append(p.Msgar, fmt.Sprintf("Revision ID list was retrieved."))
			el, _ := json.Marshal(rl.Items[0].ExportLinks)
			var obj map[string]interface{}
			json.Unmarshal(el, &obj)
			var extAr []string
			for e := range obj {
				ext := mimeToExt(e)
				extAr = append(extAr, ext)
			}
			p.Msgar = append(p.Msgar, fmt.Sprintf("Extensions which can be outputted are '%s'.", strings.Join(extAr, ", ")))
		} else {
			p.FileID = c.String("fileid")
			p.RevisionID = c.String("download")
			for i := range rl.Items {
				if rl.Items[i].ID == c.String("download") {
					ext := strings.ToLower(p.WantExt)
					if len(ext) > 0 {
						p.DlMime = extToMime(ext)
						if len(p.DlMime) == 0 {
							fmt.Fprintf(os.Stderr, "Error: '%s' is wrong extension.\n", ext)
							os.Exit(1)
						}
					} else {
						p.DlMime, ext = defFormat(p.MimeType)
					}
					dlf, _ := json.Marshal(rl.Items[i])
					var obj map[string]interface{}
					json.Unmarshal(dlf, &obj)
					p.SaveName = p.FileName + "." + ext
					dURLq, _ := obj["exportLinks"].(map[string]interface{})[p.DlMime].(string)
					p, _ = p.writeFile(dURLq)
					break
				}
			}
			if len(p.SaveName) == 0 {
				fmt.Fprintf(os.Stderr, "Error: '%s' is wrong revision number.\n", c.String("download"))
				os.Exit(1)
			}
		}
	}
}

// downloadRevisionFile : Download revision file using revision ID
func (p *FileInf) downloadRevisionFile() {
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
}

// getRevFromExGoogleDocs : Display revision IDs from files except for Google Docs.
func (p *FileInf) getRevFromExGoogleDocs(c *cli.Context) {
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
		p.GetFileinf().downloadRevisionFile()
	}
}
