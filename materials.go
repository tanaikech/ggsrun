// Package main (materials.go) :
// Materials for ggsrun.
package main

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tanaikech/ggsrun/utl"
	"github.com/urfave/cli"
)

// const :
const (
	appname  = "ggsrun"
	serverid = "115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov"

	eapir1 = "Execution API without server"
	eapir2 = "Execution API with server"
	wapps  = "Web Apps with server"

	clientsecretFile = "client_secret.json"
	cfgFile          = "ggsrun.cfg"
	cfgpathenv       = "GGSRUN_CFG_PATH"

	deffuncserv    = "ggsrunif.ExecutionApi"
	deffuncwith    = "main"
	deffuncwithout = "main"
	defprojectname = appname
	defPort        = 8080

	oauthurl      = "https://accounts.google.com/o/oauth2/"
	sdownloadurl  = "https://script.google.com/feeds/download/export?id="
	executionurl  = "https://script.googleapis.com/v1/scripts/"
	driveapiurl   = "https://www.googleapis.com/drive/v3/files/"
	chkatutl      = "https://www.googleapis.com/oauth2/v3/"
	uploadurl     = "https://www.googleapis.com/upload/drive/v3/files/"
	appsscriptapi = "https://script.googleapis.com/v1/projects"
)

// InitVal : Initial values
type InitVal struct {
	pstart  time.Time
	workdir string
	cfgdir  string
	usedDir string // "work" for working directory or "env" for directory declared by the environment variable.
	update  bool
	log     bool
	Port    int
}

// ResMsg : Response message also included errors
type ResMsg struct {
	Msg []string
}

// GgsrunCfg : Configuration file for ggsrun
type GgsrunCfg struct {
	Scriptid     string   `json:"script_id"`
	Clientid     string   `json:"client_id"`
	Clientsecret string   `json:"client_secret"`
	Refreshtoken string   `json:"refresh_token"`
	Accesstoken  string   `json:"access_token,omitempty"`
	Expiresin    int64    `json:"expires_in,omitempty"`
	Scopes       []string `json:"scopes"`
}

// Cinstalled : File of client-secret.json
type Cinstalled struct {
	ClientID                string   `json:"client_id"`
	Projectid               string   `json:"project_id"`
	Authuri                 string   `json:"auth_uri"`
	Tokenuri                string   `json:"token_uri"`
	Authproviderx509certurl string   `json:"auth_provider_x509_cert_url"`
	Clientsecret            string   `json:"client_secret"`
	Redirecturis            []string `json:"redirect_uris"`
}

// Cs : Client_secret.json
type Cs struct {
	Cid Cinstalled `json:"installed,omitempty"`
	Ciw Cinstalled `json:"web,omitempty"`
}

// Atoken : Accesstoken given from Google
type Atoken struct {
	Accesstoken  string `json:"access_token"`
	Refreshtoken string `json:"refresh_token"`
	Expiresin    int64  `json:"expires_in"`
}

// ChkAt : Condition of accesstoken retrieved using Drive API
type ChkAt struct {
	Azu        string `json:"azu,omitempty"`
	Aud        string `json:"aud,omitempty"`
	Scope      string `json:"scope,omitempty"`
	Exp        string `json:"exp,omitempty"`
	Expiresin  string `json:"expires_in,omitempty"`
	Accesstype string `json:"access_type,omitempty"`
	Error      string `json:"error_description,omitempty"`
}

// Param : Payload for Execution API
type Param struct {
	Function   string   `json:"function"`
	Parameters []string `json:"parameters,omitempty"`
	DevMode    bool     `json:"devMode"`
}

// e1para : Parameter for exe1
type e1para struct {
	Function   string        `json:"function"`
	Parameters []interface{} `json:"parameters,omitempty"`
	DevMode    bool          `json:"devMode"`
}

// Com : Structure of data using Execution API
type Com struct {
	Com     string `json:"com,omitempty"`
	Exefunc string `json:"exefunc,omitempty"`
	Log     bool   `json:"log"`
}

// Project : Project for uploading using Drive API
type Project struct {
	ScriptId string `json:"scriptId,omitempty"`
	Files    []File `json:"files"`
}

// File : Individual file in a project
type File struct {
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Source         string          `json:"source"`
	CreateTime     string          `json:"createTime,omitempty"`
	UpdateTime     string          `json:"updateTime,omitempty"`
	Creator        *creator        `json:"creator,omitempty"`
	LastModifyUser *lastmodifyuser `json:"lastModifyUser,omitempty"`
}

// creator : Creator
type creator struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// lastmodifyuser : lastModifyUser
type lastmodifyuser struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// FeedBackData : Feedbacked data from function using Execution API (modified)
type FeedBackData struct {
	Name     string `json:"name"`
	Done     bool   `json:"done"`
	Response struct {
		Type   string   `json:"@type"`
		Result Resvalue `json:"result"`
	} `json:"response"`
	Error struct {
		Code     int        `json:"code,omitempty"`
		Message  string     `json:"message,omitempty"`
		Status   string     `json:"status,omitempty"`
		Detailes []ErrorMsg `json:"details,omitempty"`
	} `json:"error"`
}

// ErrorMsg :
type ErrorMsg struct {
	Type                     string `json:"@type,omitempty"`
	ErrorMessage             string `json:"errorMessage,omitempty"`
	ErrorType                string `json:"errorType,omitempty"`
	ScriptStackTraceElements []struct {
		Function   string `json:"function,omitempty"`
		LineNumber int    `json:"lineNumber,omitempty"`
	} `json:"scriptStackTraceElements,omitempty"`
}

// Resvalue : Results of ggsrun
type Resvalue struct {
	Result   interface{}   `json:"result"`
	Logger   []interface{} `json:"logger,omitempty"`
	GoogleEt float64       `json:"GoogleElapsedTime,omitempty"`
	TotalEt  float64       `json:"TotalElapsedTime,omitempty"`
	Date     string        `json:"ScriptDate,omitempty"`
	Uapi     string        `json:"API,omitempty"`
	Message  []string      `json:"message,omitempty"`
}

// DlFileByScript : Information of download file by script
type DlFileByScript struct {
	Fileid    string `json:"-"`
	Extension string `json:"-"`
}

// ProjectUpdaterMeta : Metadata for updating a project
type ProjectUpdaterMeta struct {
	MimeType string `json:"mimeType"`
}

type updateProjectFiles struct {
	UpFiles []string
}

// ByteSliceFile : File with byte slice
type ByteSliceFile struct {
	FileData []int  `json:"result"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
}

// serverInfToGetCode : For getting auth code
type serverInfToGetCode struct {
	Response chan authCode
	Start    chan bool
	End      chan bool
}

// authCode : For getting auth code
type authCode struct {
	Code string
	Err  error
}

// AuthContainer : Struct container for using OAuth2
type AuthContainer struct {
	*InitVal   // Initial values
	*ResMsg    // Response message
	*GgsrunCfg // Config for ggsrun
	*Param     // Payload for Execution API
	*Cs        // Client_secret.json
	*Atoken    // Accesstoken from Google
	*ChkAt     // Check accesstoken
}

// ExecutionContainer : Struct container for using Execution API.
// 1. Upload script using Execution API
// 2. Update project using Drive API and execute script using Execution API
type ExecutionContainer struct {
	*InitVal            // Initial values
	*ResMsg             // Response message
	*GgsrunCfg          // Config for ggsrun
	*Param              // Payload for Execution API
	*FeedBackData       // Feedbacked data from function using Execution API
	*Project            // Project for uploading using Drive API
	*DlFileByScript     // Information of download file by script
	*updateProjectFiles // Files for updating Project
}

// DefAuthContainer : Struct container for authorization
func defAuthContainer(c *cli.Context) *AuthContainer {
	var err error
	a := &AuthContainer{
		&InitVal{},
		&ResMsg{},
		&GgsrunCfg{},
		&Param{},
		&Cs{},
		&Atoken{},
		&ChkAt{},
	}
	a.InitVal.pstart = time.Now()
	a.InitVal.workdir, err = filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	if c.Command.Names()[0] == "auth" {
		a.InitVal.cfgdir = a.InitVal.workdir
	} else {
		a.InitVal.cfgdir = os.Getenv(cfgpathenv)
		if a.InitVal.cfgdir == "" {
			a.InitVal.cfgdir = a.InitVal.workdir
		}
	}
	a.Param.Function = c.String("function")
	a.InitVal.log = c.Bool("log")
	a.InitVal.Port = defPort
	if c.Command.Names()[0] == "auth" {
		if c.Int("port") != 0 {
			a.InitVal.Port = c.Int("port")
		}
	}

	// Default scopes for using Execution API and Drive API
	// If you want to use own scopes, please write them to configuration file.
	// They are used for retrieving access token.
	//
	// From v1.4.0, https://www.googleapis.com/auth/script.projects was added to scope.
	a.GgsrunCfg.Scopes = []string{
		"https://www.googleapis.com/auth/drive",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/drive.scripts",
		"https://www.googleapis.com/auth/script.external_request",
		"https://www.googleapis.com/auth/script.scriptapp",
		"https://www.googleapis.com/auth/spreadsheets",
		"https://www.googleapis.com/auth/script.projects",
	}
	return a
}

// DefExecutionContainer : Struct container for using Execution API
func (a *AuthContainer) defExecutionContainer() *ExecutionContainer {
	e := &ExecutionContainer{
		&InitVal{},
		&ResMsg{},
		&GgsrunCfg{},
		&Param{},
		&FeedBackData{},
		&Project{},
		&DlFileByScript{},
		&updateProjectFiles{},
	}
	e.GgsrunCfg = a.GgsrunCfg
	e.InitVal = a.InitVal
	e.Msg = a.Msg
	e.Param = a.Param
	return e
}

// DefExecutionContainerWebApps : Struct container for using WebApps
func defExecutionContainerWebApps() *ExecutionContainer {
	var err error
	e := &ExecutionContainer{
		&InitVal{},
		&ResMsg{},
		&GgsrunCfg{},
		&Param{},
		&FeedBackData{},
		&Project{},
		&DlFileByScript{},
		&updateProjectFiles{},
	}
	e.InitVal.pstart = time.Now()
	e.InitVal.workdir, err = filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	e.InitVal.cfgdir = os.Getenv(cfgpathenv)
	if e.InitVal.cfgdir == "" {
		e.InitVal.cfgdir = e.InitVal.workdir
	}
	return e
}

// DefDownloadContainer : Struct container for downloading files
func (a *AuthContainer) defDownloadContainer(c *cli.Context) *utl.FileInf {
	p := &utl.FileInf{
		Msgar:       a.Msg,
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Workdir:     a.InitVal.workdir,
		PstartTime:  a.InitVal.pstart,
		FileID:      c.String("fileid"),
		ProjectID: func(c *cli.Context) string {
			id := c.String("projectid")
			if c.String("fileid") != "" && c.String("projectid") != "" {
				id = ""
			}
			return id
		}(c),
		BoundScriptName: c.String("boundscriptname"),
		WantExt:         c.String("extension"),
		WantName:        c.String("filename"),
		Progress:        c.Bool("jsonparser"),
		OverWrite:       c.Bool("overwrite"),
		RawProject:      c.Bool("rawdata"),
		ShowFileInf:     c.Bool("showfilelist"),
		Skip:            c.Bool("skip"),
		Zip:             c.Bool("zip"),
	}
	return p
}

// DefUploadContainer : Struct container for uploading files
func (a *AuthContainer) defUploadContainer(c *cli.Context) *utl.FileInf {
	p := &utl.FileInf{
		Msgar:       a.Msg,
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Workdir:     a.InitVal.workdir,
		PstartTime:  a.InitVal.pstart,
		ChunkSize: func(chnk int64) int64 {
			if chnk < 1 {
				return 1048576
			}
			return chnk * 1048576
		}(c.Int64("chunksize")),
		UpFilename: func(filenames string) []string {
			if filenames != "" {
				return regexp.MustCompile(`\s*,\s*`).Split(filenames, -1)
			}
			return nil
		}(c.String("filename")),
		ParentID: c.String("parentid"),
		ProjectType: func(ptype string) string {
			var ret string
			switch strings.ToLower(ptype) {
			case "spreadsheet", "spreadsheets", "sheet", "sheets":
				ret = "spreadsheet"
			case "document", "documents", "doc":
				ret = "document"
			case "slide", "slides":
				ret = "slide"
			case "form":
				ret = "form"
			default:
				ret = ptype
			}
			return ret
		}(c.String("projecttype")),
		GoogleDocName: c.String("googledocname"),
		ConvertTo:     c.String("convertto"),
	}
	return p
}

// dispUpdateProjectContainer : Struct container for downloading files by GAS
func (e *ExecutionContainer) dispUpdateProjectContainer() *utl.FileInf {
	p := &utl.FileInf{
		Msgar:   e.Msg,
		TotalEt: math.Trunc(time.Now().Sub(e.InitVal.pstart).Seconds()*1000) / 1000,
	}
	return p
}

// defDownloadByScriptContainer : Struct container for downloading files by GAS
func (e *ExecutionContainer) defDownloadByScriptContainer() *utl.FileInf {
	p := &utl.FileInf{
		Msgar:       e.Msg,
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Workdir:     e.InitVal.workdir,
		PstartTime:  e.InitVal.pstart,
		FileID:      e.DlFileByScript.Fileid,
		WantExt:     e.DlFileByScript.Extension,
	}
	return p
}

// defUpdateProjectContainer : Struct container for downloading files by GAS
func (e *ExecutionContainer) defUpdateProjectContainer(c *cli.Context) *ExecutionContainer {
	e.UpFiles = regexp.MustCompile(`\s*,\s*`).Split(c.String("filename"), -1)
	return e
}

// convExecutionContainerToFileInf : Convert ExecutionContainer to FileInf
func (e *ExecutionContainer) convExecutionContainerToFileInf() *utl.FileInf {
	p := &utl.FileInf{
		Accesstoken: e.Accesstoken,
	}
	return p
}

// adaptProjectForAppsScriptApi : Adapt project for Apps Script Api
func (e *ExecutionContainer) adaptProjectForAppsScriptApi() *ExecutionContainer {
	// e.Project.ScriptId = ""
	for i, f := range e.Project.Files {
		e.Project.Files[i].Type = strings.ToLower(f.Type)
		e.Project.Files[i].CreateTime = ""
		e.Project.Files[i].UpdateTime = ""
		e.Project.Files[i].Creator = nil
		e.Project.Files[i].LastModifyUser = nil
	}
	return e
}
