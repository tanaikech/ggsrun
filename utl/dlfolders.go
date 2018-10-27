// Package utl (dlfolders.go) :
// These methods are for downloading all files and folders from a folder in Google Drive.
package utl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// folderTree : Struct for folder tree.
type folderTree struct {
	IDs     [][]string `json:"id"`
	Names   []string   `json:"names"`
	Folders []string   `json:"folders"`
}

// aboutGet : Get owners of Google Drive.
type aboutGet struct {
	User owners `json:"user"`
}

// forFT : For creating folder tree.
type forFT struct {
	Name   string   `json:"name"`
	ID     string   `json:"id"`
	Parent string   `json:"parent"`
	Tree   []string `json:"tree"`
}

// forFTTemp : For creating folder tree.
type forFTTemp struct {
	Temp []forFT
}

// folderTr : For creating folder tree.
type folderTr struct {
	Temp   [][]forFT
	Search string
}

// fileListDl : Struct for file list.
type fileListDl struct {
	SearchedFolder       fileS         `json:"searchedFolder"`
	FolderTree           *folderTree   `json:"folderTree"`
	FileList             []fileListEle `json:"fileList"`
	TotalNumberOfFiles   int64         `json:"totalNumberOfFiles,omitempty"`
	TotalNumberOfFolders int64         `json:"totalNumberOfFolders,omitempty"`
}

// fileListEle : Struct for file list.
type fileListEle struct {
	FolderTree []string `json:"folderTree"`
	Files      []fileS  `json:"files"`
}

// mime2ext : Convert mimeType to extension.
func mime2ext(mime string) string {
	var obj map[string]interface{}
	json.Unmarshal([]byte(mimeVsEx), &obj)
	res, _ := obj[mime].(string)
	return res
}

// downloadFilesFromFolder : Download file using API key.
func (p *FileInf) downloadFilesFromFolder(f fileS) error {
	temp := []string{p.Workdir, p.SaveName, p.FileName, p.FileSize, p.MimeType}
	p.Workdir = f.Path
	ext := filepath.Ext(f.Name)
	if ext == "" {
		f.Name += mime2ext(f.MimeType)
	}
	p.SaveName = f.Name
	p.FileName = f.Name
	p.FileSize = f.Size
	p.MimeType = f.MimeType
	p.Zip = true
	durl := func() string {
		if f.OutMimeType == "" {
			return driveapiurl + f.ID + "?alt=media"
		}
		if f.MimeType == "application/vnd.google-apps.script" {
			return appsscriptapi + "/" + f.ID + "/content"
		}
		return driveapiurl + f.ID + "/export?mimeType=" + f.OutMimeType
	}()
	p.writeFile(durl)
	p.Workdir, p.SaveName, p.FileName, p.FileSize, p.MimeType = func(e []string) (string, string, string, string, string) {
		return e[0], e[1], e[2], e[3], e[4]
	}(temp)
	return nil
}

// makeFileByCondition : Make file by condition.
func (p *FileInf) makeFileByCondition(file fileS) error {
	if er := chkFile(filepath.Join(file.Path, file.Name)); er {
		if !p.OverWrite && !p.Skip {
			return fmt.Errorf("'%s' is existing in '%s'. If you want to overwrite, please use an option '--overwrite'", file.Name, file.Path)
		}
		if p.OverWrite && !p.Skip {
			return p.downloadFilesFromFolder(file)
		}
	} else {
		return p.downloadFilesFromFolder(file)
	}
	return nil
}

// makeDir : Make a directory by checking duplication.
func (p *FileInf) makeDir(folder string) error {
	if er := chkFile(folder); !er {
		if err := os.Mkdir(folder, 0777); err != nil {
			return err
		}
	} else {
		if !p.OverWrite && !p.Skip {
			return fmt.Errorf("'%s' is existing. If you want to overwrite, please use an option '--overwrite'", folder)
		}
	}
	return nil
}

// makeDirByCondition : Make directory by condition.
func (p *FileInf) makeDirByCondition(dir string) error {
	var err error
	if er := chkFile(dir); er {
		if !p.OverWrite && !p.Skip {
			return fmt.Errorf("'%s' is existing. If you want to overwrite, please use option '--overwrite' or '--skip'", dir)
		}
		if p.OverWrite && !p.Skip {
			if err = p.makeDir(dir); err != nil {
				return err
			}
		}
	} else {
		if err = p.makeDir(dir); err != nil {
			return err
		}
	}
	return nil
}

// initDownload : Download files by Drive API using API key.
func (p *FileInf) initDownload(fileList *fileListDl) error {
	var err error
	if p.Progress {
		fmt.Printf("Download files from a folder '%s'.\n", fileList.SearchedFolder.Name)
		fmt.Printf("There are %d files and %d folders in the folder.\n", fileList.TotalNumberOfFiles, fileList.TotalNumberOfFolders-1)
		fmt.Println("Starting download.")
	}
	idToName := map[string]interface{}{}
	for i, e := range fileList.FolderTree.Folders {
		idToName[e] = fileList.FolderTree.Names[i]
	}
	for _, e := range fileList.FileList {
		path := p.Workdir
		for _, dir := range e.FolderTree {
			path = filepath.Join(path, idToName[dir].(string))
		}
		err = p.makeDirByCondition(path)
		if err != nil {
			return err
		}
		for _, file := range e.Files {
			file.Path = path
			size, _ := strconv.ParseInt(file.Size, 10, 64)
			p.Size = size
			err = p.makeFileByCondition(file)
			if err != nil {
				return err
			}
		}
	}
	p.Msgar = append(p.Msgar, fmt.Sprintf("There were %d files and %d folders in the folder.", fileList.TotalNumberOfFiles, fileList.TotalNumberOfFolders-1))
	return nil
}

// google2ms : Convert mimeType from google to ms.
func google2ms(mimeType string) string {
	var obj map[string]interface{}
	json.Unmarshal([]byte(defaultformat), &obj)
	res, _ := obj[mimeType].(string)
	return res
}

// dupChkFoldersFiles : Check duplication of folder names and filenames.
func (p *FileInf) dupChkFoldersFiles(fileList *fileListDl) {
	dupChk1 := map[string]bool{}
	cnt1 := 2
	for i, folderName := range fileList.FolderTree.Names {
		if !dupChk1[folderName] {
			dupChk1[folderName] = true
		} else {
			fileList.FolderTree.Names[i] = folderName + "_" + strconv.Itoa(cnt1)
		}
	}
	extt := strings.ToLower(p.WantExt)
	for i, list := range fileList.FileList {
		if len(list.Files) > 0 {
			dupChk2 := map[string]bool{}
			cnt2 := 2
			for j, file := range list.Files {
				if !dupChk2[file.Name] {
					dupChk2[file.Name] = true
				} else {
					ext := filepath.Ext(file.Name)
					if ext != "" {
						fileList.FileList[i].Files[j].Name = file.Name[0:len(file.Name)-len(ext)] + "_" + strconv.Itoa(cnt2) + ext
					} else {
						fileList.FileList[i].Files[j].Name = file.Name + "_" + strconv.Itoa(cnt2)
					}
					cnt2++
				}
				if extt != "" {
					if google2ms(file.MimeType) != "" {
						cmime := extToMime(extt)
						if cmime != "" {
							fileList.FileList[i].Files[j].OutMimeType = cmime
						} else {
							fileList.FileList[i].Files[j].OutMimeType, _ = defFormat(file.MimeType)
						}
					}
				} else {
					fileList.FileList[i].Files[j].OutMimeType, _ = defFormat(file.MimeType)
				}
			}
		}
	}
}

// getFilesFromFolder : Retrieve file list from folder list.
func (p *FileInf) getFilesFromFolder(folderTree *folderTree) *fileListDl {
	f := &fileListDl{}
	f.SearchedFolder.ID = p.FileID
	f.SearchedFolder.Name = p.FileName
	f.SearchedFolder.MimeType = p.MimeType
	f.SearchedFolder.Owners = p.Owners
	f.SearchedFolder.CreatedTime = p.CreatedTime
	f.SearchedFolder.ModifiedTime = p.ModifiedTime
	f.FolderTree = folderTree
	for i, id := range folderTree.Folders {
		q := "'" + id + "' in parents and mimeType != 'application/vnd.google-apps.folder' and trashed=false"
		fields := "files(createdTime,description,id,mimeType,modifiedTime,name,owners,parents,permissions,shared,size),nextPageToken"
		fm := p.GetListLoop(q, fields)
		var fe fileListEle
		fe.FolderTree = folderTree.IDs[i]
		fe.Files = append(fe.Files, fm.Files...)
		f.FileList = append(f.FileList, fe)
	}
	f.TotalNumberOfFolders = int64(len(f.FolderTree.Folders))
	f.TotalNumberOfFiles = func() int64 {
		var c int64
		for _, e := range f.FileList {
			c += int64(len(e.Files))
		}
		return c
	}()
	return f
}

// getDlFoldersS : Retrieve each folder from folder list using folder ID. This is for shared folders.
func (fr *folderTr) getDlFoldersS(searchFolderName string) *folderTree {
	fT := &folderTree{}
	fT.Folders = append(fT.Folders, fr.Search)
	fT.Names = append(fT.Names, searchFolderName)
	fT.IDs = append(fT.IDs, []string{fr.Search})
	for _, e := range fr.Temp {
		for _, f := range e {
			fT.Folders = append(fT.Folders, f.ID)
			var tmp []string
			tmp = append(tmp, f.Tree...)
			tmp = append(tmp, f.ID)
			fT.IDs = append(fT.IDs, tmp)
			fT.Names = append(fT.Names, f.Name)
		}
	}
	return fT
}

// getAllfoldersRecursively : Recursively get folder tree using Drive API.
func (p *FileInf) getAllfoldersRecursively(id string, parents []string, fls *folderTr) *folderTr {
	q := "'" + id + "' in parents and mimeType='application/vnd.google-apps.folder' and trashed=false"
	fields := "files(id,mimeType,name,parents,size),nextPageToken"
	fm := p.GetListLoop(q, fields)
	var temp forFTTemp
	for _, e := range fm.Files {
		ForFt := &forFT{
			ID:   e.ID,
			Name: e.Name,
			Parent: func() string {
				if len(e.Parents) > 0 {
					return e.Parents[0]
				}
				return ""
			}(),
			Tree: append(parents, id),
		}
		temp.Temp = append(temp.Temp, *ForFt)
	}
	if len(temp.Temp) > 0 {
		fls.Temp = append(fls.Temp, temp.Temp)
		for _, e := range temp.Temp {
			p.getAllfoldersRecursively(e.ID, e.Tree, fls)
		}
	}
	return fls
}

// createFolderTreeID : Create a folder tree.
func createFolderTreeID(fm fileListSt, id string, parents []string, fls *folderTr) *folderTr {
	var temp forFTTemp
	for _, e := range fm.Files {
		if len(e.Parents) > 0 && e.Parents[0] == id {
			ForFt := &forFT{
				ID:   e.ID,
				Name: e.Name,
				Parent: func() string {
					if len(e.Parents) > 0 {
						return e.Parents[0]
					}
					return ""
				}(),
				Tree: append(parents, id),
			}
			temp.Temp = append(temp.Temp, *ForFt)
		}
	}
	if len(temp.Temp) > 0 {
		fls.Temp = append(fls.Temp, temp.Temp)
		for _, e := range temp.Temp {
			createFolderTreeID(fm, e.ID, e.Tree, fls)
		}
	}
	return fls
}

// getOwner : Judge owner of the file and Google Drive. true is the same. false is the difference.
func (p *FileInf) getOwner() bool {
	r := &RequestParams{
		Method:      "GET",
		APIURL:      "https://www.googleapis.com/drive/v3/about?fields=user",
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	res, _ := r.FetchAPI()
	ag := &aboutGet{}
	json.Unmarshal(res, &ag)
	for _, e := range p.Owners {
		if reflect.DeepEqual(e, ag.User) {
			return true
		}
	}
	return false
}

// DlFolders : Main method for downloading folders
func (p *FileInf) DlFolders() error {
	if p.Progress {
		fmt.Printf("Files are downloaded from a folder '%s'.\n", p.FileName)
		fmt.Println("Getting values to download.")
	}
	p.Msgar = append(p.Msgar, fmt.Sprintf("Files were downloaded from folder '%s'.", p.FileName))
	folT := &folderTree{}
	if p.getOwner() {
		q := "mimeType='application/vnd.google-apps.folder' and trashed=false"
		fields := "files(id,mimeType,name,parents,size),nextPageToken"
		fm := p.GetListLoop(q, fields)
		fT1 := &folderTr{Search: p.FileID}
		folT = createFolderTreeID(fm, p.FileID, []string{}, fT1).getDlFoldersS(p.FileName)
	} else {
		fT2 := &folderTr{Search: p.FileID}
		folT = p.getAllfoldersRecursively(p.FileID, []string{}, fT2).getDlFoldersS(p.FileName)
	}
	fileList := p.getFilesFromFolder(folT)
	p.dupChkFoldersFiles(fileList)
	if p.ShowFileInf {
		p.FolderTree = fileList
		return nil
	}
	dlres := p.initDownload(fileList)
	if dlres != nil {
		return dlres
	}
	return nil
}
