// Package utl (transfer.go) :
// These methods are for downloading, uploading and retrieving file list from or to Google Drive.
package utl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

const (
	sdownloadurl = "https://script.google.com/feeds/download/export?id="
	lurl         = "https://www.googleapis.com/drive/v3/files?"
	driveapiurl  = "https://www.googleapis.com/drive/v3/files/"
	uploadurl    = "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&"
)

// FileInf : File information for downloading and uploading
type FileInf struct {
	Accesstoken  string         `json:"-"`
	DlMime       string         `json:"-"`
	MimeType     string         `json:"mimeType,omitempty"`
	Workdir      string         `json:"-"`
	PstartTime   time.Time      `json:"-"`
	WantExt      string         `json:"-"`
	WantName     string         `json:"-"`
	WebLink      string         `json:"webContentLink,omitempty"`
	WebView      string         `json:"webViewLink,omitempty"`
	SearchByName string         `json:"-"`
	SearchByID   string         `json:"-"`
	FileID       string         `json:"id,omitempty"`
	FileName     string         `json:"name,omitempty"`
	SaveName     string         `json:"saved_file_name,omitempty"`
	Parents      []string       `json:"parents,omitempty"`
	UpFilename   []string       `json:"upload_file_name,omitempty"`
	UpFileID     []string       `json:"uid,omitempty"`
	UppedFiles   []uploadedFile `json:"uploaded_files,omitempty"`
	TotalEt      float64        `json:"TotalElapsedTime,omitempty"`
	Msgar        []string       `json:"message,omitempty"`
}

// dlError : Error messages.
type dlError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// project : Project structure
type project struct {
	Files []struct {
		ID     string `json:"id,omitempty"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Source string `json:"source"`
	} `json:"files"`
}

// filea : Individual file in a project
type filea struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

// fileListSt : File list.
type fileListSt struct {
	NextPageToken string `json:"nextPageToken,omitempty"`
	Files         []struct {
		ID                string   `json:"id,omitempty"`
		Name              string   `json:"name,omitempty"`
		MimeType          string   `json:"mimeType,omitempty"`
		Parents           []string `json:"parents,omitempty"`
		CreatedTime       string   `json:"createdTime,omitempty"`
		ModifiedTime      string   `json:"modifiedTime,omitempty"`
		FullFileExtension string   `json:"fullFileExtension,omitempty"`
		Size              string   `json:"size,omitempty"`
		WebLink           string   `json:"webContentLink,omitempty"`
		WebView           string   `json:"webViewLink,omitempty"`
	}
}

// fileUploaderMeta : For uploading scripts.
type fileUploaderMeta struct {
	Name     string   `json:"name"`
	Parents  []string `json:"parents,omitempty"`
	MimeType string   `json:"mimeType"`
}

// uploadedFile : For uploading files.
type uploadedFile struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	MimeType string   `json:"mimeType"`
	Parents  []string `json:"parents,omitempty"`
}

//dispDup : For duplicating values.
type dispDup struct {
	Name         string
	FileID       string
	MimeType     string
	ModifiedTime string
}

// saveScript : Back up a project.
func (p *FileInf) saveScript(data []byte, c *cli.Context) *FileInf {
	var f project
	json.Unmarshal(data, &f)
	if c.Bool("rawdata") {
		filename := filepath.Join(p.Workdir, p.FileName+".json")
		p.SaveName = p.FileName + ".json"
		p.Msgar = append(p.Msgar, fmt.Sprintf("Saved project as a JSON file '%s.json'.", p.FileName))
		btok, _ := json.MarshalIndent(f, "", "\t")
		ioutil.WriteFile(filename, btok, 0777)
	} else {
		p.SaveName = ""
		if len(f.Files) == 1 {
			p.Msgar = append(p.Msgar, fmt.Sprintf("%s has %d script.", p.FileName, len(f.Files)))
		} else {
			p.Msgar = append(p.Msgar, fmt.Sprintf("%s has %d scripts.", p.FileName, len(f.Files)))
		}
		for _, e := range f.Files {
			saveName := p.FileName + "_" + e.Name + "." + func(ex string) string {
				var eext string
				if len(ex) > 0 {
					eext = ex
				} else {
					eext = "gs"
				}
				return eext
			}(p.WantExt)
			src := fmt.Sprintf("// Script ID in Project = %s \n%s", e.ID, e.Source)
			ioutil.WriteFile(filepath.Join(p.Workdir, saveName), []byte(src), 0777)
			p.Msgar = append(p.Msgar, fmt.Sprintf("Script was downloaded as '%s'.", saveName))
		}
	}
	return p
}

// Downloader : Download files.
func (p *FileInf) Downloader(c *cli.Context) *FileInf {
	ext := strings.ToLower(p.WantExt)
	if len(ext) > 0 {
		p.DlMime = extToMime(ext)
	} else {
		p.DlMime, ext = defFormat(p.MimeType)
	}
	if len(p.FileID) > 0 {
		var body []byte
		var gm map[string]interface{}
		json.Unmarshal([]byte(googlemimetypes), &gm)
		if gm["exportFormats"].(map[string]interface{})[p.MimeType] != nil {
			for _, e := range gm["exportFormats"].(map[string]interface{})[p.MimeType].([]interface{}) {
				if e == p.DlMime {
					p.SaveName = p.FileName + "." + ext
				}
			}
			if len(p.SaveName) == 0 {
				dispRes, _ := json.MarshalIndent(gm["exportFormats"], "", "  ")
				fmt.Fprintf(os.Stderr, "Error: Bad extension or No extension. It supports as follows.\n%s ", string(dispRes))
				os.Exit(1)
			}
			if p.MimeType == "application/vnd.google-apps.script" {
				p, body = p.writeFile(sdownloadurl + p.FileID + "&format=json")
				p.saveScript(body, c)
			} else {
				p, _ = p.writeFile(driveapiurl + p.FileID + "/export?mimeType=" + p.DlMime)
			}
		} else {
			p.SaveName = p.FileName
			p, _ = p.writeFile(driveapiurl + p.FileID + "?alt=media")
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: Please input File Name or File ID. ")
		os.Exit(1)
	}
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}

// writeFile : Create files on local.
func (p *FileInf) writeFile(durl string) (*FileInf, []byte) {
	r := &RequestParams{
		Method:      "GET",
		APIURL:      durl,
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	var er dlError
	json.Unmarshal(body, &er)
	if err != nil || er.Error.Code-300 >= 0 {
		fmt.Print(fmt.Sprintf("Error: %s Status code is %d. FileID: %s  ", er.Error.Message, er.Error.Code, p.FileID))
		os.Exit(1)
	}
	if p.MimeType != "application/vnd.google-apps.script" {
		ioutil.WriteFile(filepath.Join(p.Workdir, p.SaveName), body, 0777)
		p.Msgar = append(p.Msgar, fmt.Sprintf("File was downloaded as '%s'.", p.SaveName))
	}
	return p, body
}

// nameToID :
func (p *FileInf) nameToID(name string) ([]byte, error) {
	number := 1000
	tokenparams := url.Values{}
	tokenparams.Set("orderBy", "name")
	tokenparams.Set("pageSize", strconv.Itoa(number))
	tokenparams.Set("q", "name='"+name+"' and trashed=false")
	tokenparams.Set("fields", "files(createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size,webContentLink,webViewLink)")
	r := &RequestParams{
		Method:      "GET",
		APIURL:      lurl + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	return r.FetchAPI()
}

// idToName : Convert file ID to file name.
func (p *FileInf) idToName(id string) ([]byte, error) {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size,webContentLink,webViewLink")
	r := &RequestParams{
		Method:      "GET",
		APIURL:      driveapiurl + id + "?" + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	return r.FetchAPI()
}

// GetFileinf : Retrieve file infomation using Drive API.
func (p *FileInf) GetFileinf() *FileInf {
	if len(p.FileID) > 0 {
		body, err := p.idToName(p.FileID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: File ID '%s' Not found. %v .", p.FileID, err)
			os.Exit(1)
		}
		var er dlError
		json.Unmarshal(body, &er)
		if err != nil || er.Error.Code-300 >= 0 {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Error: %s Status code is %d. ", er.Error.Message, er.Error.Code))
			os.Exit(1)
		}
		json.Unmarshal(body, &p)
	} else if len(p.WantName) > 0 {
		finf, err := p.nameToID(p.WantName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v. ", err)
			os.Exit(1)
		}
		var fl fileListSt
		json.Unmarshal(finf, &fl)
		if len(fl.Files) == 1 {
			p.FileID = fl.Files[0].ID
			p.FileName = fl.Files[0].Name
			p.MimeType = fl.Files[0].MimeType
			p.WebLink = fl.Files[0].WebLink
			p.WebView = fl.Files[0].WebView
		} else if len(fl.Files) > 1 {
			fmt.Printf("# %d files were found. Please download them using File ID.\n", len(fl.Files))
			for i := range fl.Files {
				dd := &dispDup{
					Name:         fl.Files[i].Name,
					FileID:       fl.Files[i].ID,
					MimeType:     fl.Files[i].MimeType,
					ModifiedTime: fl.Files[i].ModifiedTime,
				}
				rd, _ := json.MarshalIndent(dd, "", "  ")
				fmt.Printf("%s\n", rd)
			}
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Error: File name '%s' is not found. ", p.WantName)
			os.Exit(1)
		}
	}
	if p.MimeType == "application/vnd.google-apps.folder" {
		fmt.Fprintf(os.Stderr, "Error: '%s' is a Folder. Cannot download Folder yet. ", p.FileID)
		os.Exit(1)
	}
	return p
}

// extToMime : Convert from extension to mimeType of the file on Local.
func extToMime(ext string) string {
	var fm map[string]interface{}
	json.Unmarshal([]byte(extVsmime), &fm)
	st, _ := fm[strings.Replace(strings.ToLower(ext), ".", "", 1)].(string)
	return st
}

// defFormat : Default download format
func defFormat(mime string) (string, string) {
	var df map[string]interface{}
	json.Unmarshal([]byte(defaultformat), &df)
	dmime, _ := df[mime].(string)
	var ext string
	var fm map[string]interface{}
	json.Unmarshal([]byte(extVsmime), &fm)
	for i, v := range fm {
		if v == dmime {
			ext = i
		}
	}
	return dmime, ext
}

// scriptUploader : For uploading scripts.
func (p *FileInf) scriptUploader(metadata map[string]interface{}, pr []byte) *FileInf {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "id,mimeType,name,parents")
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part := make(textproto.MIMEHeader)
	part.Set("Content-Type", "application/json")
	data, err := w.CreatePart(part)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	re, _ := json.Marshal(metadata)
	if _, err = io.Copy(data, bytes.NewReader(re)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	data, err = w.CreatePart(part)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	if _, err = io.Copy(data, bytes.NewReader(pr)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	w.Close()
	r := &RequestParams{
		Method:      "POST",
		APIURL:      uploadurl + tokenparams.Encode(),
		Data:        &b,
		Contenttype: w.FormDataContentType(),
		Accesstoken: p.Accesstoken,
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	var uf uploadedFile
	json.Unmarshal(body, &uf)
	p.UppedFiles = append(p.UppedFiles, uf)
	return p
}

// fileUploader : For uploading files.
func (p *FileInf) fileUploader(metadata map[string]interface{}, file string) *FileInf {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "id,mimeType,name,parents")
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part := make(textproto.MIMEHeader)
	part.Set("Content-Type", "application/json")
	data, err := w.CreatePart(part)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	re, _ := json.Marshal(metadata)
	if _, err = io.Copy(data, bytes.NewReader(re)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	fs, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	defer fs.Close()
	data, err = w.CreateFormFile("file", file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	if _, err = io.Copy(data, fs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	w.Close()
	r := &RequestParams{
		Method:      "POST",
		APIURL:      uploadurl + tokenparams.Encode(),
		Data:        &b,
		Contenttype: w.FormDataContentType(),
		Accesstoken: p.Accesstoken,
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	var uf uploadedFile
	json.Unmarshal(body, &uf)
	p.UppedFiles = append(p.UppedFiles, uf)
	return p
}

// Uploader : Main method for uploading
// "$ ggsrun u -f t1.gs,t2.gs" or "$ ggsrun u -f "t1.gs, t2.gs""
func (p *FileInf) Uploader(c *cli.Context) *FileInf {
	if len(c.String("projectname")) == 0 {
		for _, elm := range p.UpFilename {
			metadata := &fileUploaderMeta{
				Name:    filepath.Base(elm),
				Parents: []string{c.String("parentfolderid")},
				MimeType: func(flag bool) string {
					var r string
					if !flag {
						r = extToGMime(filepath.Ext(elm))
					}
					return r
				}(c.Bool("noconvert")),
			}
			if metadata.MimeType == "application/vnd.google-apps.script" {
				var pr project
				filedata := &filea{
					Name:   metadata.Name,
					Type:   "server_js",
					Source: ConvGasToUpload(elm),
				}
				pr.Files = append(pr.Files, *filedata)
				pre, _ := json.Marshal(pr)
				upmeta, _ := json.Marshal(metadata)
				var u map[string]interface{}
				json.Unmarshal(upmeta, &u)
				if len(c.String("parentfolderid")) == 0 {
					delete(u, "parents")
				}
				_ = p.scriptUploader(u, pre)
				p.Msgar = append(p.Msgar, fmt.Sprintf("Uploaded %s as %s. ", filepath.Base(elm), metadata.Name))
			} else {
				upmeta, _ := json.Marshal(metadata)
				var u map[string]interface{}
				json.Unmarshal(upmeta, &u)
				if len(c.String("parentfolderid")) == 0 {
					delete(u, "parents")
				}
				p.fileUploader(u, elm)
				p.Msgar = append(p.Msgar, fmt.Sprintf("Uploaded %s as %s.", filepath.Base(elm), metadata.Name))
			}
		}
	} else {
		metadata := &fileUploaderMeta{
			Name:     c.String("projectname"),
			Parents:  []string{c.String("parentfolderid")},
			MimeType: "application/vnd.google-apps.script",
		}
		upmeta, _ := json.Marshal(metadata)
		var u map[string]interface{}
		json.Unmarshal(upmeta, &u)
		if len(c.String("parentfolderid")) == 0 {
			delete(u, "parents")
		}
		var pr project
		for _, elm := range p.UpFilename {
			if filepath.Ext(elm) == ".gs" ||
				filepath.Ext(elm) == ".gas" ||
				filepath.Ext(elm) == ".js" ||
				filepath.Ext(elm) == ".htm" ||
				filepath.Ext(elm) == ".html" {
				filedata := &filea{
					Name: strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1),
					Type: func(ex string) string {
						var scripttype string
						switch ex {
						case ".gs", ".gas", ".js":
							scripttype = "server_js"
						case ".htm", ".html":
							scripttype = "html"
						}
						return scripttype
					}(filepath.Ext(elm)),
					Source: ConvGasToUpload(elm),
				}
				pr.Files = append(pr.Files, *filedata)
			}
		}
		p.Msgar = append(p.Msgar, fmt.Sprintf("Uploaded %d scripts as a project with a name of '%s'.", len(p.UpFilename), metadata.Name))
		pre, _ := json.Marshal(pr)
		_ = p.scriptUploader(u, pre)
	}
	return p
}

// extToGMime : Convert from extension to mimeType of the files on Google.
func extToGMime(ext string) string {
	var fm map[string]interface{}
	json.Unmarshal([]byte(extVsmime), &fm)
	st, _ := fm[strings.Replace(strings.ToLower(ext), ".", "", 1)].(string)
	if len(st) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Extension of '%s' cannot be uploaded. ", ext)
		os.Exit(1)
	}
	var gm map[string]interface{}
	json.Unmarshal([]byte(googlemimetypes), &gm)
	return gm["importFormats"].(map[string]interface{})[st].([]interface{})[0].(string)
}

// GetFileList : Retrieving file list on Google Drive.
func (p *FileInf) GetFileList(c *cli.Context) *FileInf {
	if len(c.String("searchbyname")) > 0 {
		p.SearchByName = c.String("searchbyname")
		body, err := p.nameToID(p.SearchByName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v. ", err)
			os.Exit(1)
		}
		var fl fileListSt
		json.Unmarshal(body, &fl)
		if len(fl.Files) == 1 {
			p.FileID = fl.Files[0].ID
			p.FileName = fl.Files[0].Name
			p.MimeType = fl.Files[0].MimeType
			p.Parents = fl.Files[0].Parents
			p.WebView = fl.Files[0].WebView
		} else if len(fl.Files) > 1 {
			for i := range fl.Files {
				fmt.Printf("{\n  Name: \"%s\",\n  ID: \"%s\",\n  ModifiedTime: \"%s\",\n  URL: \"%s\"\n}\n", fl.Files[i].Name, fl.Files[i].ID, fl.Files[i].ModifiedTime, fl.Files[i].WebView)
			}
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Error: File name '%s' is not found. ", p.SearchByName)
			os.Exit(1)
		}
		p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
		return p
	}
	if len(c.String("searchbyid")) > 0 {
		p.SearchByID = c.String("searchbyid")
		body, err := p.idToName(p.SearchByID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: File ID '%s' is not found. ", p.SearchByID)
			os.Exit(1)
		}
		json.Unmarshal(body, &p)
		p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
		return p
	}
	var fm fileListSt
	var fl fileListSt
	var dmy fileListSt
	fm.NextPageToken = ""
	for i := 0; ; {
		_ = i
		body, err := p.getList(fm.NextPageToken)
		json.Unmarshal(body, &fl)
		fm.NextPageToken = fl.NextPageToken
		fm.Files = append(fm.Files, fl.Files...)
		fl.NextPageToken = ""
		fl.Files = dmy.Files
		if len(fm.NextPageToken) == 0 || err != nil {
			break
		}
	}
	var fol, fil []string
	for i := range fm.Files {
		if strings.Contains(fm.Files[i].MimeType, "folder") {
			fol = append(fol, fm.Files[i].Name)
		} else {
			fil = append(fil, fm.Files[i].Name)
		}
	}
	p.Msgar = append(p.Msgar, fmt.Sprintf("Total: %d, File: %d, Folder: %d", len(fm.Files), len(fil), len(fol)))
	p.Msgar = append(p.Msgar, fmt.Sprintf("If you want a file list, please use option '-s' or '-f'. The file name is automatically given."))
	if c.Bool("stdout") {
		fmt.Printf("| File name | File ID | Modified time | Create time | Type |\n")
		var ftype string
		for i := range fm.Files {
			if strings.Contains(fm.Files[i].MimeType, "folder") {
				ftype = "Folder"
			} else {
				ftype = "File"
			}
			fmine := fmt.Sprintf("| %s | %s | %s | %s | %s |\n", fm.Files[i].Name, fm.Files[i].ID, fm.Files[i].ModifiedTime, fm.Files[i].CreatedTime, ftype)
			fmt.Print(fmine)
		}
	}
	if c.Bool("file") {
		filename := filepath.Join(p.Workdir, p.PstartTime.Format("Files_20060102_150405")+".json")
		p.Msgar = append(p.Msgar, fmt.Sprintf("Saved a JSON file as %s.", filename))
		btok, _ := json.MarshalIndent(fm, "", "\t")
		ioutil.WriteFile(filename, btok, 0777)
	}
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}

// getList : For retrieving file list.
func (p *FileInf) getList(ptoken string) ([]byte, error) {
	number := 1000
	tokenparams := url.Values{}
	tokenparams.Set("orderBy", "name")
	tokenparams.Set("pageSize", strconv.Itoa(number))
	tokenparams.Set("q", "trashed=false")
	tokenparams.Set("fields", "files(createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size),nextPageToken")
	if len(ptoken) > 0 {
		tokenparams.Set("pageToken", ptoken)
	}
	r := &RequestParams{
		Method:      "GET",
		APIURL:      lurl + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	return body, err
}
