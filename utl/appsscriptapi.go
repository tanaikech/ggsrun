// Package utl (appsscriptapi.go) :
// These methods are for using Apps Script API.
package utl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	appsscriptapi = "https://script.googleapis.com/v1/projects"
)

// AppsScriptApiInf : Information retrieved by Apps Script API
type AppsScriptApiInf struct {
	ScriptId       string         `json:"scriptId"`
	ParentId       string         `json:"parentId"`
	Title          string         `json:"title"`
	CreateTime     time.Time      `json:"createTime"`
	UpdateTime     time.Time      `json:"updateTime"`
	Creator        creator        `json:"creator"`
	LastModifyUser lastmodifyuser `json:"lastModifyUser"`
}

// creator : Creator
type creator struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// lastmodifyuser : lastModifyUser
type lastmodifyuser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ProjectForAppsScriptApi : Project structure for Apps Script API
type ProjectForAppsScriptApi struct {
	ScriptId string                  `json:"scriptId,omitempty"`
	Files    []FilesForAppsScriptApi `json:"files"`
}

// FilesForAppsScriptApi : A file structure for Apps Script API
type FilesForAppsScriptApi struct {
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Source         string          `json:"source"`
	CreateTime     time.Time       `json:"createTime,omitempty"`
	UpdateTime     time.Time       `json:"updateTime,omitempty"`
	Creator        *creator        `json:"creator,omitempty"`
	LastModifyUser *lastmodifyuser `json:"lastModifyUser,omitempty"`
}

// manifestsStruct : Struct of Manifests
type manifestsStruct struct {
	TimeZone          string        `json:"timeZone,omitempty"`
	OauthScopes       []interface{} `json:"oauthScopes,omitempty"`
	Dependencies      interface{}   `json:"dependencies,omitempty"`
	ExceptionLogging  interface{}   `json:"exceptionLogging,omitempty"`
	Webapp            interface{}   `json:"webapp,omitempty"`
	ExecutionApi      interface{}   `json:"executionApi,omitempty"`
	UrlFetchWhitelist []interface{} `json:"urlFetchWhitelist,omitempty"`
	Gmail             interface{}   `json:"gmail,omitempty"`
}

// projectVersionList : Struct for version list of project
type projectVersionList struct {
	Versions []struct {
		ScriptId      string    `json:"scriptId"`
		VersionNumber int       `json:"versionNumber"`
		Description   string    `json:"description"`
		CreateTime    time.Time `json:"createTime"`
	} `json:"versions"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// createVersionSt : Struct for creating version
type createVersionSt struct {
	VersionNumber string `json:"versionNumber,omitempty"`
	Description   string `json:"description"`
	CreateTime    string `json:"createTime,omitempty"`
}

// getBoundScriptInf : Retrieve information of boundscript.
func (p *FileInf) getBoundScriptInf(id string) {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "createTime,creator,lastModifyUser,parentId,scriptId,title,updateTime")
	r := &RequestParams{
		Method:      "GET",
		APIURL:      appsscriptapi + "/" + id + "?" + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: File ID '%s' is not found. ", id)
		DispScopeError2(body)
		os.Exit(1)
	}
	var i *AppsScriptApiInf
	json.Unmarshal(body, &i)
	p.FileName = i.Title
	p.MimeType = "application/vnd.google-apps.script"
	p.ParentID = i.ParentId
	p.FileID = i.ScriptId
	o := &owners{}
	o.Email = i.Creator.Email
	o.Name = i.Creator.Name
	p.Owners = append(p.Owners, *o)
	if i.LastModifyUser.Email != "" || i.LastModifyUser.Name != "" {
		lmu := &lastmodifieduser{
			i.LastModifyUser.Email,
			i.LastModifyUser.Name,
		}
		p.LastModifyingUser = lmu
	}
}

// getBoundScript : Retrieve boundscript.
func (p *FileInf) getBoundScript(id string) *ProjectForAppsScriptApi {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "files,scriptId")
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, id+"/content")
	r := &RequestParams{
		Method:      "GET",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: File ID '%s' is not found. ", p.SearchByID)
		DispScopeError2(body)
		os.Exit(1)
	}
	pf := &ProjectForAppsScriptApi{}
	json.Unmarshal(body, &pf)
	return pf
}

// boundScriptCreator : Create container bound-scripts in Google Docs.
func (p *FileInf) boundScriptCreator(metadata []byte) *AppsScriptApiInf {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "createTime,creator,lastModifyUser,parentId,scriptId,title,updateTime")
	r := &RequestParams{
		Method:      "POST",
		APIURL:      appsscriptapi + "?" + tokenparams.Encode(),
		Data:        bytes.NewBuffer(metadata),
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v.\n%v\n\n", err, string(body))
		fmt.Fprintf(os.Stderr, "One of reasons of error :\n Was the inputted parent ID correct?.\n")
		var u map[string]interface{}
		json.Unmarshal(body, &u)
		em := u["error"].(map[string]interface{})["message"]
		if em == "Request had insufficient authentication scopes." {
			DispScopeError1()
		} else if em == "Request contains an invalid argument." {
			fmt.Fprintf(os.Stderr, " - If this error occurs when you try to create project in Google Slides, this may be a bug. https://issuetracker.google.com/issues/72238499\n")
		}
		os.Exit(1)
	}
	var a *AppsScriptApiInf
	json.Unmarshal(body, &a)
	var uf uploadedFile
	uf.ID = a.ScriptId
	uf.Name = a.Title
	uf.MimeType = "application/vnd.google-apps.script"
	p.UppedFiles = append(p.UppedFiles, uf)
	return a
}

// ProjectUpdateByAppsScriptApi : For uploading project using Apps Script API.
func (p *FileInf) ProjectUpdateByAppsScriptApi(pr *ProjectForAppsScriptApi) *AppsScriptApiInf {
	pre, _ := json.Marshal(pr)
	tokenparams := url.Values{}
	tokenparams.Set("fields", "files,scriptId")
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, pr.ScriptId+"/content")
	r := &RequestParams{
		Method:      "PUT",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        bytes.NewBuffer(pre),
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		DispScopeError2(body)
		os.Exit(1)
	}
	var asi *AppsScriptApiInf
	json.Unmarshal(body, &asi)
	return asi
}

// createProjectForAppsScriptApi : Create json of project for Apps Script API.
func (p *FileInf) createProjectForAppsScriptApi(scriptId string) *ProjectForAppsScriptApi {
	pr := &ProjectForAppsScriptApi{}
	pr.ScriptId = scriptId
	if len(p.UpFilename) > 0 {
		for _, elm := range p.UpFilename {
			if ChkExtention(filepath.Ext(elm)) {
				filedata := &FilesForAppsScriptApi{
					Name:   strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1),
					Type:   ExtToType(filepath.Ext(elm), true),
					Source: ConvGasToUpload(elm),
				}
				pr.Files = append(pr.Files, *filedata)
			}
		}
		if len(pr.Files) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Inputted files cannot be used for GAS project.\n")
			os.Exit(1)
		}
	} else {
		filedata := &FilesForAppsScriptApi{
			Name:   "Code",
			Type:   "SERVER_JS",
			Source: "function myFunction() {\n  \n}\n",
		}
		pr.Files = append(pr.Files, *filedata)
	}
	return pr
}

// getManifests : Retrieve Manifests from data
func (pf *ProjectForAppsScriptApi) getManifests(timeZone string) *FilesForAppsScriptApi {
	manifests := &FilesForAppsScriptApi{}
	for _, e := range pf.Files {
		if e.Name == "appsscript" && e.Type == "JSON" {
			manifests.Name = e.Name
			manifests.Type = e.Type
			manifests.Source = e.Source
			break
		}
	}
	if timeZone != "" {
		var mf manifestsStruct
		json.Unmarshal([]byte(manifests.Source), &mf)
		mf.TimeZone = timeZone
		umf, err := json.MarshalIndent(mf, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			os.Exit(1)
		}
		manifests.Source = string(umf)
	}
	return manifests
}

// setManifests : Import Manifests to data
func (pf *ProjectForAppsScriptApi) setManifests(manifests *FilesForAppsScriptApi) *ProjectForAppsScriptApi {
	chkManifests := func(files []FilesForAppsScriptApi) bool {
		for _, e := range files {
			if e.Name == "appsscript" && e.Type == "JSON" {
				return true
			}
		}
		return false
	}(pf.Files)
	if !chkManifests {
		filedata := &FilesForAppsScriptApi{
			Name:   manifests.Name,
			Type:   manifests.Type,
			Source: manifests.Source,
		}
		pf.Files = append(pf.Files, *filedata)
	}
	return pf
}

// getProjectVersionListInit : Initial method for retrieving version list of project.
func (p *FileInf) getProjectVersionListInit() *projectVersionList {
	fm := &projectVersionList{}
	var fl projectVersionList
	var dmy projectVersionList
	fm.NextPageToken = ""
	for i := 0; ; {
		_ = i
		body, err := p.getProjectVersionList(fm.NextPageToken)
		json.Unmarshal(body, &fl)
		fm.NextPageToken = fl.NextPageToken
		fm.Versions = append(fm.Versions, fl.Versions...)
		fl.NextPageToken = ""
		fl.Versions = dmy.Versions
		if len(fm.NextPageToken) == 0 || err != nil {
			break
		}
	}
	return fm
}

// getProjectVersionList : Retrieve version list of project.
func (p *FileInf) getProjectVersionList(ptoken string) ([]byte, error) {
	number := 100
	tokenparams := url.Values{}
	tokenparams.Set("fields", "nextPageToken,versions")
	tokenparams.Set("pageSize", strconv.Itoa(number))
	if len(ptoken) > 0 {
		tokenparams.Set("pageToken", ptoken)
	}
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, p.FileID+"/versions")
	r := &RequestParams{
		Method:      "GET",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        nil,
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n%s\n", err, string(body))
		DispScopeError2(body)
		os.Exit(1)
	}
	return body, err
}

// createProjectVersion : Create new version of GAS project.
func (p *FileInf) createProjectVersion(description string) {
	var payload createVersionSt
	payload.Description = description
	payl, _ := json.Marshal(payload)
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, p.FileID+"/versions")
	r := &RequestParams{
		Method:      "POST",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(payl),
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n%s\n", err, string(body))
		DispScopeError2(body)
		os.Exit(1)
	}
	var rs map[string]interface{}
	json.Unmarshal(body, &rs)
	p.Msgar = append(p.Msgar, fmt.Sprintf("New version was created to '%s' as '%d'.", p.FileName, int(rs["versionNumber"].(float64))))
	p.Msgar = append(p.Msgar, fmt.Sprintf("Description is '%s'.", description))
}

// DispScopeError1 : Display about new scope of 'https://www.googleapis.com/auth/script.projects'.
func DispScopeError1() {
	fmt.Printf("\n\n##########\n")
	fmt.Fprintf(os.Stderr, "One of reasons of error :\n - Did you add new scope of 'https://www.googleapis.com/auth/script.projects' to 'ggsrun.cfg'? If this scope is not added yet, please add it, and run below.\n\n $ ggsrun auth\n\n By this, the scope is reflected.\n - And please enable Google Apps Script API at 'https://console.cloud.google.com/apis/library/script.googleapis.com/?project=### project ID ###'\n You can see '### project ID ###' in 'client_secret.json'.\n")
	fmt.Printf("##########\n")
}

// DispScopeError2 : Display about new scope of 'https://www.googleapis.com/auth/script.projects'.
func DispScopeError2(body []byte) {
	var u map[string]interface{}
	json.Unmarshal(body, &u)
	em := u["error"].(map[string]interface{})["message"]
	if em == "Request had insufficient authentication scopes." {
		DispScopeError1()
	}
}
