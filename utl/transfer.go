// Package utl (transfer.go) :
// These methods are for downloading, uploading files and folders, and retrieving file list from or to Google Drive.
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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"
)

const (
	driveapiv3          = "https://www.googleapis.com/drive/v3"
	lurl                = "https://www.googleapis.com/drive/v3/files?"
	driveapiurl         = "https://www.googleapis.com/drive/v3/files/"
	driveapiurlv2       = "https://www.googleapis.com/drive/v2/files/"
	uploadurl           = "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&"
	lengthOfProjectId   = 57
	maxSizeForMultipart = 5242880
)

// FileInf : File information for downloading and uploading
type FileInf struct {
	ProjectID         string            `json:"project_id,omitempty"`
	ParentID          string            `json:"parentId,omitempty"`
	RevisionID        string            `json:"revisionid,omitempty"`
	FileID            string            `json:"id,omitempty"`
	FileName          string            `json:"name,omitempty"`
	SaveName          string            `json:"saved_file_name,omitempty"`
	MimeType          string            `json:"mimeType,omitempty"`
	UpFilename        []string          `json:"upload_file_name,omitempty"`
	UpFileID          []string          `json:"uid,omitempty"`
	Parents           []string          `json:"parents,omitempty"`
	FileSize          string            `json:"size,omitempty"`
	WebLink           string            `json:"webContentLink,omitempty"`
	WebView           string            `json:"webViewLink,omitempty"`
	CreatedTime       *time.Time        `json:"createdTime,omitempty"`
	ModifiedTime      *time.Time        `json:"modifiedTime,omitempty"`
	UppedFiles        []uploadedFile    `json:"uploaded_files,omitempty"`
	LastModifyingUser *lastmodifieduser `json:"lastModifyingUser,omitempty"`
	Owners            []owners          `json:"owners,omitempty"`
	Msgar             []string          `json:"message,omitempty"`
	TotalEt           float64           `json:"TotalElapsedTime,omitempty"`
	FolderTree        *fileListDl       `json:"folderTreeWithFiles,omitempty"`
	SearchQuery       string            `json:"query,omitempty"`
	SearchFields      string            `json:"fields,omitempty"`
	SearchRegex       string            `json:"regex,omitempty"`
	SearchedFiles     []fileS           `json:"searchedFiles,omitempty"`
	SearchedResult    string            `json:"searchedResult,omitempty"`
	InputtedMimeType  []string          `json:"inputtedMimeType,omitempty"`
	ReturnedResult    interface{}       `json:"returnedResult,omitempty"`

	Accesstoken       string        `json:"-"`
	BoundScriptName   string        `json:"-"`
	ChunkSize         int64         `json:"-"`
	ConvertTo         string        `json:"-"`
	DlMime            string        `json:"-"`
	GoogleDocName     string        `json:"-"`
	OverWrite         bool          `json:"-"`
	PermissionInfo    permissionInf `json:"-"`
	Progress          bool          `json:"-"`
	ProjectType       string        `json:"-"`
	PstartTime        time.Time     `json:"-"`
	RawProject        bool          `json:"-"`
	SearchByName      string        `json:"-"`
	SearchByID        string        `json:"-"`
	ShowFileInf       bool          `json:"-"`
	Size              int64         `json:"-"`
	Skip              bool          `json:"-"`
	UseServiceAccount string        `json:"-"`
	WantExt           string        `json:"-"`
	WantName          string        `json:"-"`
	Workdir           string        `json:"-"`
	Zip               bool          `json:"-"`
}

// owners : Owners of file
type owners struct {
	Name         string `json:"displayName,omitempty"`
	PermissionID string `json:"permissionId"`
	Email        string `json:"emailAddress,omitempty"`
}

// lastmodifieduser : Last modified user of file
type lastmodifieduser struct {
	Name  string `json:"displayName,omitempty"`
	Email string `json:"emailAddress,omitempty"`
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

// newProject : Create new project
type newProject struct {
	ParentId string `json:"parentId,omitempty"`
	Title    string `json:"title"`
}

// fileS : Structure of a file information.
type fileS struct {
	Kind              string     `json:"kind,omitempty"`
	ID                string     `json:"id,omitempty"`
	Name              string     `json:"name,omitempty"`
	MimeType          string     `json:"mimeType,omitempty"`
	Starred           bool       `json:"starred,omitempty"`
	Trashed           bool       `json:"trashed,omitempty"`
	ExplicitlyTrashed bool       `json:"explicitlyTrashed,omitempty"`
	Parents           []string   `json:"parents,omitempty"`
	CreatedTime       *time.Time `json:"createdTime,omitempty"`
	ModifiedTime      *time.Time `json:"modifiedTime,omitempty"`
	ModifiedByMeTime  *time.Time `json:"modifiedByMeTime,omitempty"`
	ModifiedByMe      bool       `json:"modifiedByMe,omitempty"`
	FullFileExtension string     `json:"fullFileExtension,omitempty"`
	Size              string     `json:"size,omitempty"`
	Version           string     `json:"version,omitempty"`
	HasThumbnail      bool       `json:"hasThumbnail,omitempty"`
	ThumbnailVersion  string     `json:"thumbnailVersion,omitempty"`
	ViewedByMe        bool       `json:"viewedByMe,omitempty"`
	WebLink           string     `json:"webContentLink,omitempty"`
	WebView           string     `json:"webViewLink,omitempty"`
	Owners            []owners   `json:"owners,omitempty"`
	Shared            bool       `json:"shared,omitempty"`
	Permissions       []struct {
		Kind         string `json:"kind,omitempty"`
		ID           string `json:"id,omitempty"`
		Type         string `json:"type,omitempty"`
		EmailAddress string `json:"emailAddress,omitempty"`
		Role         string `json:"role,omitempty"`
		DisplayName  string `json:"displayName,omitempty"`
		PhotoLink    string `json:"photoLink,omitempty"`
		Deleted      bool   `json:"deleted,omitempty"`
	} `json:"permissions,omitempty"`
	OwnedByMe                    bool              `json:"ownedByMe,omitempty"`
	ViewersCanCopyContent        bool              `json:"viewersCanCopyContent,omitempty"`
	CopyRequiresWriterPermission bool              `json:"copyRequiresWriterPermission,omitempty"`
	WritersCanShare              bool              `json:"writersCanShare,omitempty"`
	PermissionIds                []string          `json:"permissionIds,omitempty"`
	FolderColorRgb               string            `json:"folderColorRgb,omitempty"`
	QuotaBytesUsed               string            `json:"quotaBytesUsed,omitempty"`
	IsAppAuthorized              bool              `json:"isAppAuthorized,omitempty"`
	Path                         string            `json:"path,omitempty"`
	OutMimeType                  string            `json:"outMimeType,omitempty"`
	LastModifyingUser            *lastmodifieduser `json:"lastModifyingUser,omitempty"`
}

// fileListSt : File list.
type fileListSt struct {
	NextPageToken string  `json:"nextPageToken,omitempty"`
	Files         []fileS `json:"files,omitempty"`
}

// fileUploaderMeta : For uploading scripts.
type fileUploaderMeta struct {
	Name     string   `json:"name"`
	Parents  []string `json:"parents,omitempty"`
	MimeType string   `json:"mimeType"`
}

// uploadedFile : For uploading files.
type uploadedFile struct {
	fileS
}

//dispDup : For duplicating values.
type dispDup struct {
	Name         string
	FileID       string
	MimeType     string
	ModifiedTime string
}

// saveScript : Back up a project. Save a project as a raw, each file and zip.
func (p *FileInf) saveScript(data []byte) *FileInf {
	var f project
	json.Unmarshal(data, &f)
	if p.RawProject {
		filename := filepath.Join(p.Workdir, p.FileName+".json")
		if chkFile(filename) && !p.OverWrite {
			if !p.Skip {
				fmt.Fprintf(os.Stderr, "Error: '%s' is exsinting. If you want to overwrite the file, please use option '--overwrite'.", filename)
				os.Exit(1)
			} else {
				if p.Progress {
					fmt.Printf("File of '%s' is not saved and skipped, because the file is duplicated.\n", filename)
				}
				p.Msgar = append(p.Msgar, fmt.Sprintf("File of '%s' was not saved and skipped, because the file is duplicated.", filename))
			}
		} else {
			p.SaveName = p.FileName + ".json"
			p.Msgar = append(p.Msgar, fmt.Sprintf("Saved project as a JSON file '%s.json'.", p.FileName))
			btok, _ := json.MarshalIndent(f, "", "\t")
			ioutil.WriteFile(filename, btok, 0777)
		}
	} else {
		p.SaveName = ""
		if len(f.Files) == 1 {
			p.Msgar = append(p.Msgar, fmt.Sprintf("%s has %d script.", p.FileName, len(f.Files)))
		} else {
			p.Msgar = append(p.Msgar, fmt.Sprintf("%s has %d scripts.", p.FileName, len(f.Files)))
		}
		zh := &zipFileHeads{}
		for _, e := range f.Files {
			eType := strings.ToLower(e.Type)
			saveName := p.FileName + "_" + e.Name + "." + func(ex, ty string) string {
				var eext string
				if len(ex) > 0 {
					eext = ex
				} else {
					switch ty {
					case "server_js":
						eext = "gs"
					case "html":
						eext = "html"
					case "json":
						eext = "json"
					default:
						eext = "txt"
					}
				}
				return eext
			}(p.WantExt, eType)
			z := &zipFileHead{
				Name:     saveName,
				Modified: p.PstartTime,
				Method:   8,
				Body:     []byte(e.Source),
			}
			zh.Files = append(zh.Files, *z)
		}
		if p.Zip {
			buf := zh.doFilesZip(p.zipComment())
			zn := p.FileName + ".zip"
			zipFileName := filepath.Join(p.Workdir, zn)
			if chkFile(zipFileName) && !p.OverWrite {
				if !p.Skip {
					fmt.Fprintf(os.Stderr, "Error: '%s' is exsinting. If you want to overwrite the file, please use option '--overwrite'.", zipFileName)
					os.Exit(1)
				} else {
					if p.Progress {
						fmt.Printf("File of '%s' is not saved and skipped, because the file is duplicated.\n", p.FileName)
					}
					p.Msgar = append(p.Msgar, fmt.Sprintf("File of '%s' was not saved and skipped, because the file is duplicated.", p.FileName))
				}
			} else {
				if p.Progress {
					fmt.Printf("Project file '%s' is downloaded.\n", p.FileName)
				}
				ioutil.WriteFile(zipFileName, buf.Bytes(), 0777)
				p.Msgar = append(p.Msgar, fmt.Sprintf("%d scripts in the project were saved as '%s'.", len(zh.Files), zn))
			}
		} else {
			for _, e := range zh.Files {
				scriptFileName := filepath.Join(p.Workdir, e.Name)
				if chkFile(scriptFileName) && !p.OverWrite {
					if !p.Skip {
						fmt.Fprintf(os.Stderr, "Error: '%s' is exsinting. If you want to overwrite the file, please use option '--overwrite'.", scriptFileName)
						os.Exit(1)
					} else {
						if p.Progress {
							fmt.Printf("File of '%s' is not saved and skipped, because the file is duplicated.\n", scriptFileName)
						}
						p.Msgar = append(p.Msgar, fmt.Sprintf("File of '%s' was not saved and skipped, because the file is duplicated.", scriptFileName))
					}
				} else {
					if p.Progress {
						fmt.Printf("Script '%s' is downloaded.\n", e.Name)
					}
					ioutil.WriteFile(scriptFileName, e.Body, 0777)
					p.Msgar = append(p.Msgar, fmt.Sprintf("Script was downloaded as '%s'.", e.Name))
				}
			}
		}
	}
	return p
}

// chkFile : Check the existence of file and directory in local PC.
func chkFile(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// Downloader : Download files.
func (p *FileInf) Downloader(c *cli.Context) *FileInf {
	if p.MimeType == "application/vnd.google-apps.folder" {
		err := p.DlFolders()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			os.Exit(1)
		}
		p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
		return p
	}
	ext := strings.ToLower(p.WantExt)
	if len(ext) > 0 {
		p.DlMime = extToMime(ext)
	} else {
		p.DlMime, ext = defFormat(p.MimeType)
	}
	if len(p.FileID) > 0 && c.String("deletefile") == "" {
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
				u, _ := url.Parse(appsscriptapi)
				u.Path = path.Join(u.Path, p.FileID+"/content")
				p.writeFile(u.String())
			} else if p.MimeType != "" {
				p.writeFile(driveapiurl + p.FileID + "/export?mimeType=" + p.DlMime)
			}
		} else {
			p.SaveName = p.FileName
			p.writeFile(driveapiurl + p.FileID + "?alt=media")
		}
	} else if c.String("deletefile") != "" {
		p.deleteFile(c.String("deletefile"))
		p.Msgar = append(p.Msgar, fmt.Sprintf("File with fileId '%s' was deleted.", c.String("deletefile")))
	} else {
		p.Msgar = append(p.Msgar, "Error: Please input File Name or File ID. Please check HELP using 'ggsrun d --help'.")
	}
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}

// chunks : For io.Reader
type chunks struct {
	io.Reader
	cChunk   int64
	End      int64
	FileName string
}

// Read : For io.Reader
func (c *chunks) Read(dat []byte) (int, error) {
	n, err := c.Reader.Read(dat)
	c.cChunk += int64(n)
	if err == nil {
		if c.End == 0 {
			fmt.Printf("\rNow downloading '%s' (bytes)... %d", c.FileName, c.cChunk)
		} else {
			fmt.Printf("\rNow downloading '%s' (bytes)... %d / %d", c.FileName, c.cChunk, c.End)
		}
	}
	return n, err
}

// writeFile : Create files on local.
func (p *FileInf) writeFile(durl string) *FileInf {
	var err error
	timeOut, err := func(size int64, err error) (int64, error) {
		if err == nil || size == 0 {
			switch {
			case size < 100000000:
				return 3600, nil
			case size > 100000000:
				return 0, nil
			}
		}
		return 0, fmt.Errorf("%s", err)
	}(strconv.ParseInt(p.FileSize, 10, 64))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error 1: %v. ", err)
		os.Exit(1)
	}
	r := &RequestParams{
		Method:      "GET",
		APIURL:      durl,
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       timeOut,
	}
	res, err := r.FetchAPIRaw()
	if err != nil {
		errmsg, er := ioutil.ReadAll(res.Body)
		if er != nil {
			os.Exit(1)
		}
		defer res.Body.Close()
		fmt.Fprintf(os.Stderr, "Error 2: %v. %s", err, errmsg)
		os.Exit(1)
	}
	var body []byte
	var dFileName string
	if p.MimeType == "application/vnd.google-apps.script" {
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error 3: %v. ", err)
			os.Exit(1)
		}
		defer res.Body.Close()
		var er dlError
		json.Unmarshal(body, &er)
		if err != nil || er.Error.Code-300 >= 0 {
			fmt.Print(fmt.Sprintf("Error: %s. (Status code is %d)\nFileID: %s\n", er.Error.Message, er.Error.Code, p.FileID))
			if er.Error.Message == "Request had insufficient authentication scopes." {
				DispScopeError1()
			}
			os.Exit(1)
		}
		return p.saveScript(body)
	}
	dFileName = filepath.Join(p.Workdir, p.SaveName)
	if chkFile(dFileName) && !p.OverWrite {
		if !p.Skip {
			fmt.Fprintf(os.Stderr, "Error: '%s' is exsinting. If you want to overwrite the file, please use option '--overwrite'.", dFileName)
			os.Exit(1)
		} else {
			if p.Progress {
				fmt.Printf("File of '%s' is not saved and skipped, because the file is duplicated.\n", p.SaveName)
			}
			p.Msgar = append(p.Msgar, fmt.Sprintf("File of '%s' was not saved and skipped, because the file is duplicated.", p.SaveName))
		}
	} else {
		file, err := os.Create(dFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error 4: %v. ", err)
			os.Exit(1)
		}
		if p.Progress {
			chunks := &chunks{
				Reader: res.Body,
				End: func() int64 {
					size, _ := strconv.ParseInt(p.FileSize, 10, 64)
					if size > 0 {
						return size
					}
					return 0
				}(),
				FileName: p.FileName,
			}
			_, err = io.Copy(file, chunks)
		} else {
			_, err = io.Copy(file, res.Body)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error 5: %v. ", err)
			os.Exit(1)
		}
		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error 6: %v. ", err)
			os.Exit(1)
		}
		if p.Progress {
			fmt.Printf("\n")
		}
		defer func() {
			file.Close()
			res.Body.Close()
		}()
		p.Msgar = append(p.Msgar, fmt.Sprintf("File was downloaded as '%s'. Size was %d bytes.", p.SaveName, fileInfo.Size()))
		return p
	}
	return p
}

// deleteFile : Delete a file using a file ID on own Google Drive.
func (p *FileInf) deleteFile(id string) {
	r := &RequestParams{
		Method:      "DELETE",
		APIURL:      driveapiurl + id,
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n%s\n", err, body)
		os.Exit(1)
	}
	return
}

// nameToID : Convert filename to file ID
func (p *FileInf) nameToID(name string) ([]byte, error) {
	number := 1000
	tokenparams := url.Values{}
	tokenparams.Set("orderBy", "name")
	tokenparams.Set("pageSize", strconv.Itoa(number))
	tokenparams.Set("q", "name='"+name+"' and trashed=false")
	tokenparams.Set("fields", "files(createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size,webContentLink,webViewLink,lastModifyingUser(displayName,emailAddress),owners(displayName,emailAddress,permissionId))")
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

// idToName : Convert file ID to filename.
func (p *FileInf) idToName(id string) ([]byte, error) {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size,webContentLink,webViewLink,lastModifyingUser(displayName,emailAddress),owners(displayName,emailAddress,permissionId)")
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

// ChkBoundOrStandalone : Check whether the fileId is a bound script or a standalone script.
func (p *FileInf) ChkBoundOrStandalone(fileId string) ([]byte, error, bool) {
	body, err := p.idToName(fileId)
	if err != nil && len(fileId) == lengthOfProjectId {
		return body, err, false
	} else if err != nil && len(fileId) < lengthOfProjectId {
		fmt.Fprintf(os.Stderr, "Error: File ID '%s' Not found. %v .", fileId, err)
		os.Exit(1)
	}
	var er dlError
	json.Unmarshal(body, &er)
	if err != nil || er.Error.Code-300 >= 0 {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Error: %s Status code is %d. ", er.Error.Message, er.Error.Code))
		os.Exit(1)
	}
	return body, err, true
}

// GetFileinf : Retrieve file infomation using Drive API.
func (p *FileInf) GetFileinf() *FileInf {
	if len(p.FileID) > 0 {
		if body, _, chk := p.ChkBoundOrStandalone(p.FileID); chk {
			json.Unmarshal(body, &p)
		} else {
			p.getBoundScriptInf(p.FileID)
		}
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
			p.FileSize = fl.Files[0].Size
			p.MimeType = fl.Files[0].MimeType
			p.Parents = fl.Files[0].Parents
			p.WebLink = fl.Files[0].WebLink
			p.WebView = fl.Files[0].WebView
			p.Owners = fl.Files[0].Owners
			p.CreatedTime = fl.Files[0].CreatedTime
			p.ModifiedTime = fl.Files[0].ModifiedTime
			p.LastModifyingUser = fl.Files[0].LastModifyingUser
		} else if len(fl.Files) > 1 {
			fmt.Printf("# %d files were found. Please download them using File ID.\n", len(fl.Files))
			for i := range fl.Files {
				dd := &dispDup{
					Name:         fl.Files[i].Name,
					FileID:       fl.Files[i].ID,
					MimeType:     fl.Files[i].MimeType,
					ModifiedTime: fl.Files[i].ModifiedTime.In(time.Local).Format("20060102 15:04:05 MST"),
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
	return p
}

// extToMime : Convert from extension to mimeType of the file on Local.
func extToMime(ext string) string {
	var fm map[string]interface{}
	json.Unmarshal([]byte(extVsmime), &fm)
	st, _ := fm[strings.Replace(strings.ToLower(ext), ".", "", 1)].(string)
	return st
}

// mimeToExt : Convert from mimeType to extension of the file.
func mimeToExt(mime string) string {
	var fm map[string]interface{}
	json.Unmarshal([]byte(extVsmime), &fm)
	var ext string
	for e, i := range fm {
		if i == mime {
			ext = e
			break
		}
	}
	return ext
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
	tokenparams.Set("fields", "createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size,webContentLink,webViewLink,lastModifyingUser(displayName,emailAddress),owners(displayName,emailAddress,permissionId)")
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
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n%s\n", err, body)
		os.Exit(1)
	}
	var uf uploadedFile
	json.Unmarshal(body, &uf)
	p.UppedFiles = append(p.UppedFiles, uf)
	return p
}

// fileUploader : For uploading files.
func (p *FileInf) fileUploader(metadata map[string]interface{}, file string) *FileInf {
	var err error
	var fs *os.File
	if file != "" {
		fs, err = os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", err)
			os.Exit(1)
		}
		defer fs.Close()
		fstatus, err := fs.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v.\n", err)
			os.Exit(1)
		}
		if fstatus.Size() > maxSizeForMultipart {
			rbody := p.ResumableUpload(metadata, fs, fstatus)
			var uf uploadedFile
			json.Unmarshal(rbody, &uf)
			p.UppedFiles = append(p.UppedFiles, uf)
			p.Msgar = append(p.Msgar, fmt.Sprintf("'%s' (%d bytes) was uploaded by resumable upload.", metadata["name"], fstatus.Size()))
			return p
		}
	}
	tokenparams := url.Values{}
	tokenparams.Set("fields", "createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size,webContentLink,webViewLink,lastModifyingUser(displayName,emailAddress),owners(displayName,emailAddress,permissionId)")
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
	if file != "" {
		data, err = w.CreateFormFile("file", file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v. ", err)
			os.Exit(1)
		}
		if _, err = io.Copy(data, fs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v. ", err)
			os.Exit(1)
		}
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
		fmt.Fprintf(os.Stderr, "Error: %s\n%s\n", err, body)
		os.Exit(1)
	}
	var uf uploadedFile
	json.Unmarshal(body, &uf)
	p.UppedFiles = append(p.UppedFiles, uf)
	return p
}

// Uploader : Main method for uploading
// "$ ggsrun u -f t1.gs,t2.gs" or "$ ggsrun u -f "t1.gs, t2.gs""
// Upload type is automatically selected by the file size.
func (p *FileInf) Uploader(c *cli.Context) *FileInf {
	if c.String("projectname") == "" && len(p.UpFilename) == 0 {
		p.Msgar = append(p.Msgar, "Error: No options. Please check HELP using 'ggsrun u --help'.")
	} else if c.String("projectname") == "" && len(p.UpFilename) > 0 && p.ParentID == "" {
		for _, elm := range p.UpFilename {
			metadata := &fileUploaderMeta{
				Name:    filepath.Base(elm),
				Parents: []string{c.String("parentfolderid")},
				MimeType: func(flag bool) string {
					if !flag {
						convto := strings.ToLower(p.ConvertTo)
						switch {
						case convto == "":
							return extToGMime(filepath.Ext(elm))
						case convto == "document" || convto == "doc" || convto == "docs":
							return "application/vnd.google-apps.document"
						case convto == "spreadsheet" || convto == "sheet" || convto == "spread":
							return "application/vnd.google-apps.spreadsheet"
						case convto == "slides" || convto == "slide" || convto == "presentation":
							return "application/vnd.google-apps.presentation"
						default:
							return extToGMime(convto)
						}
					}
					return ""
				}(c.Bool("noconvert")),
			}
			if metadata.MimeType == "application/vnd.google-apps.script" {
				if p.UseServiceAccount != "" {
					return p.whenServiceAccountIsUsed()
				}
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
		if p.UseServiceAccount != "" {
			return p.whenServiceAccountIsUsed()
		}
		if p.ParentID == "" {
			if p.ProjectType == "standalone" {
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
				pre := p.createProject(c.String("timezone"))
				_ = p.scriptUploader(u, pre)
				p.createdprojectresult(len(p.UpFilename), metadata.Name)
			} else {
				parentId := p.createGoogleDocs(c)
				p.createProjectInGoogleDocs(c, parentId)
			}
		} else {
			if c.String("projectname") != "" {
				p.createProjectInGoogleDocs(c, p.ParentID)
			} else {
				fmt.Fprintf(os.Stderr, "Error: No project name. Please input project name using '--projectname' or '-pn' and try again.\n")
				os.Exit(1)
			}
		}
	}
	p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
	return p
}

// createProjectInGoogleDocs : Create new project as a bound script.
func (p *FileInf) createProjectInGoogleDocs(c *cli.Context, parentId string) {
	metadata := &newProject{
		ParentId: parentId,
		Title:    c.String("projectname"),
	}
	meta, _ := json.Marshal(metadata)
	asi := p.boundScriptCreator(meta)
	manifests := p.getBoundScript(asi.ScriptId).getManifests(c.String("timezone"))
	pre := p.createProjectForAppsScriptApi(asi.ScriptId).setManifests(manifests)
	_ = p.ProjectUpdateByAppsScriptApi(pre)
	p.createdprojectresult(len(p.UpFilename), metadata.Title)
}

// createGoogleDocs : Create new Google Docs (spreadsheet, document, slide and form)
func (p *FileInf) createGoogleDocs(c *cli.Context) string {
	metadata := &fileUploaderMeta{
		Name: func(c *cli.Context) string {
			if c.String("googledocname") != "" {
				return c.String("googledocname")
			}
			return c.String("projectname")
		}(c),
		Parents: func(folderId string) []string {
			if folderId != "" {
				return []string{folderId}
			}
			return []string{}
		}(c.String("parentfolderid")),
		MimeType: func(ptype string) string {
			var ret string
			switch strings.ToLower(ptype) {
			case "spreadsheet":
				ret = "application/vnd.google-apps.spreadsheet"
			case "document":
				ret = "application/vnd.google-apps.document"
			case "slide":
				ret = "application/vnd.google-apps.presentation"
			case "form":
				ret = "application/vnd.google-apps.form"
			}
			return ret
		}(p.ProjectType),
	}
	upmeta, _ := json.Marshal(metadata)
	var u map[string]interface{}
	json.Unmarshal(upmeta, &u)
	p.fileUploader(u, "")
	return p.UppedFiles[0].ID
}

// createdprojectresult : Result of created project
func (p *FileInf) createdprojectresult(num int, filename string) {
	if num > 0 {
		p.Msgar = append(p.Msgar, fmt.Sprintf("Uploaded %d scripts as new project with a name of '%s'.", num, filename))
	} else {
		p.Msgar = append(p.Msgar, fmt.Sprintf("New project was created as the filename of '%s'.", filename))
	}
}

// createProject : Create new project as json
func (p *FileInf) createProject(timeZone string) []byte {
	var pr project
	if len(p.UpFilename) > 0 {
		for _, elm := range p.UpFilename {
			if ChkExtention(filepath.Ext(elm)) {
				filedata := &filea{
					Name:   strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1),
					Type:   ExtToType(filepath.Ext(elm), false),
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
		filedata := &filea{
			Name:   "Code",
			Type:   "server_js",
			Source: "function myFunction() {\n  \n}\n",
		}
		pr.Files = append(pr.Files, *filedata)
	}
	if timeZone != "" {
		filedata := &filea{
			Name:   "appsscript",
			Type:   "json",
			Source: "{\n  \"timeZone\": \"" + timeZone + "\",\n  \"dependencies\": {\n  },\n  \"exceptionLogging\": \"STACKDRIVER\"\n}\n",
		}
		pr.Files = append(pr.Files, *filedata)
	}
	pre, _ := json.Marshal(pr)
	return pre
}

// ChkExtention : Check extension of inputted files.
func ChkExtention(ex string) bool {
	switch strings.ToLower(ex) {
	case ".gs", ".gas", ".js", ".htm", ".html", ".json":
		return true
	default:
		return false
	}
}

// ExtToType : Convert extension to scripttype for project.
func ExtToType(ex string, uppercase bool) string {
	var scripttype string
	switch strings.ToLower(ex) {
	case ".gs", ".gas", ".js":
		scripttype = "server_js"
	case ".htm", ".html":
		scripttype = "html"
	case ".json":
		scripttype = "json"
	}
	if uppercase {
		scripttype = strings.ToUpper(scripttype)
	}
	return scripttype
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
			p.WebLink = fl.Files[0].WebLink
			p.Owners = fl.Files[0].Owners
			p.CreatedTime = fl.Files[0].CreatedTime
			p.ModifiedTime = fl.Files[0].ModifiedTime
			p.LastModifyingUser = fl.Files[0].LastModifyingUser
		} else if len(fl.Files) > 1 {
			for i := range fl.Files {
				fmt.Printf("{\n  Name: \"%s\",\n  ID: \"%s\",\n  ModifiedTime: \"%s\",\n  URL: \"%s\"\n}\n",
					fl.Files[i].Name,
					fl.Files[i].ID,
					fl.Files[i].ModifiedTime.In(time.Local).Format("20060102 15:04:05 MST"),
					fl.Files[i].WebView,
				)
			}
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Error: File name '%s' is not found. How about trying this using file ID, again?", p.SearchByName)
			os.Exit(1)
		}
		p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
		return p
	}
	if len(c.String("searchbyid")) > 0 {
		p.SearchByID = c.String("searchbyid")
		if body, _, chk := p.ChkBoundOrStandalone(p.SearchByID); chk {
			json.Unmarshal(body, &p)
		} else {
			p.getBoundScriptInf(p.SearchByID)
		}
		p.TotalEt = math.Trunc(time.Now().Sub(p.PstartTime).Seconds()*1000) / 1000
		return p
	}
	q := "trashed=false"
	fields := "files(createdTime,fullFileExtension,id,mimeType,modifiedTime,name,parents,size),nextPageToken"
	fm := p.GetListLoop(q, fields)
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
		buffer := &bytes.Buffer{}
		w := new(tabwriter.Writer)
		w.Init(buffer, 0, 4, 1, ' ', 0)
		fmt.Fprintf(w, "\n%s\t%s\t%s\t%s\t%s\n", "# FileName", "# FileID", "# ModifiedTime", "# CreatedTime", "# Type")
		var ftype string
		for i := range fm.Files {
			if strings.Contains(fm.Files[i].MimeType, "folder") {
				ftype = "Folder"
			} else {
				ftype = "File"
			}
			fmt.Fprintf(
				w, "%s\t%s\t%s\t%s\t%s\n",
				fm.Files[i].Name,
				fm.Files[i].ID,
				fm.Files[i].ModifiedTime.In(time.Local).Format("20060102 15:04:05 MST"),
				fm.Files[i].CreatedTime.In(time.Local).Format("20060102 15:04:05 MST"),
				ftype,
			)
		}
		w.Flush()
		fmt.Printf("%s\n", buffer)
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

// GetListLoop : Loop for retrieving file list.
func (p *FileInf) GetListLoop(q, fields string) fileListSt {
	var err error
	var body []byte
	var fm fileListSt
	var fl fileListSt
	var dmy fileListSt
	fm.NextPageToken = ""
	for {
		body, err = p.getList(fm.NextPageToken, q, fields)
		json.Unmarshal(body, &fl)
		fm.NextPageToken = fl.NextPageToken
		fm.Files = append(fm.Files, fl.Files...)
		fl.NextPageToken = ""
		fl.Files = dmy.Files
		if len(fm.NextPageToken) == 0 || err != nil {
			break
		}
	}
	if err != nil {
		p.errHandlingFromFetch(body)
	}
	return fm
}

// getList : For retrieving file list.
func (p *FileInf) getList(ptoken, q, fields string) ([]byte, error) {
	number := 1000
	tokenparams := url.Values{}
	tokenparams.Set("orderBy", "name")
	tokenparams.Set("pageSize", strconv.Itoa(number))
	tokenparams.Set("q", q)
	tokenparams.Set("fields", fields)
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

// whenServiceAccountIsUsed : When ServiceAccount is used, there are some limitations.
func (p *FileInf) whenServiceAccountIsUsed() *FileInf {
	p.Msgar = append(p.Msgar, fmt.Sprintf("Warning: In the current stage, script files cannot be uploaded and downloaded by Service Account yet."))
	return p
}

// saveResponse : Save response from API.
func (p *FileInf) saveResponse(body []byte) {
	var rm interface{}
	json.Unmarshal(body, &rm)
	p.ReturnedResult = rm
}

// reqAndGetRawResponse : Request and retrieve raw resource.
func (p *FileInf) reqAndGetRawResponse(r *RequestParams) {
	res, err := r.FetchAPI()
	if err != nil {
		p.errHandlingFromFetch(res)
		return
	}
	p.saveResponse(res)
}
