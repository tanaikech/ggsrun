// Package main (projectupdater.go) :
// These methods are for updating project.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ggsrun/utl"

	"github.com/urfave/cli"
)

// projectUpdateControl : Main method for updating project.
func (e *ExecutionContainer) projectUpdateControl(c *cli.Context) *utl.FileInf {
	if len(c.String("projectid")) > 0 {
		e.GgsrunCfg.Scriptid = c.String("projectid")
		if len(c.String("filename")) > 0 {
			if !c.Bool("deletefiles") {
				return e.defUpdateProjectContainer(c).
					projectBackup(c).
					ProjectMaker().
					projectUpdate2().
					dispUpdateProjectContainer()
			}
			return e.defUpdateProjectContainer(c).
				projectBackup(c).
				filesInProjectRemover().
				projectUpdate2().
				dispUpdateProjectContainer()
		}
		if c.Bool("rearrange") {
			e.defUpdateProjectContainer(c).
				projectBackup(c).
				rearrangeByTerminal()
		}
		if len(c.String("rearrangewithfile")) > 0 {
			data := getRearrangeTemplate(c.String("rearrangewithfile"))
			e.defUpdateProjectContainer(c).
				projectBackup(c).
				rearrangeByFile(data)
		}
	} else {
		e.Msg = append(e.Msg, "Error: No options. Please check HELP using 'ggsrun ud --help'.")
	}
	return e.dispUpdateProjectContainer()
}

// projectUpdateForBoundScript : Update bound-script project
// func (e *ExecutionContainer) projectUpdateForBoundScript() *ExecutionContainer {
// 	p := e.convExecutionContainerToFileInf()
// 	var pr *utl.ProjectForAppsScriptApi
// 	var pp *utl.FilesForAppsScriptApi
// 	pr.ScriptId = e.Project.ScriptId
// 	for _, f := range e.Project.Files {
// 		pp.Name = f.Name
// 		pp.Type = f.Type
// 		pp.Source = f.Source
// 		pr.Files = append(pr.Files, *pp)
// 	}
// 	_ = p.ProjectUpdateByAppsScriptApi(pr)
// 	e.Msg = append(e.Msg, "Project was updated.")
// 	return e
// }

// ProjectMaker : Recreates the project using uploaded scripts.
func (e *ExecutionContainer) ProjectMaker() *ExecutionContainer {
	for _, elm := range e.UpFiles {
		if utl.ChkExtention(filepath.Ext(elm)) {
			filedata := &File{
				Name:   strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1),
				Type:   utl.ExtToType(filepath.Ext(elm), false),
				Source: utl.ConvGasToUpload(elm),
			}
			var overwrite bool
			for i, v := range e.Project.Files {
				if v.Name == filedata.Name {
					e.Project.Files[i].Source = filedata.Source
					e.Msg = append(e.Msg, fmt.Sprintf("'%s' (%s) in project was overwritten.", v.Name, v.Type))
					overwrite = true
				}
			}
			if !overwrite {
				e.Project.Files = append(e.Project.Files, *filedata)
			}
		} else {
			e.Msg = append(e.Msg, fmt.Sprintf("File of '%s' cannot be used for updating project.", elm))
		}
	}
	p := e.convExecutionContainerToFileInf()
	body, err, _ := p.ChkBoundOrStandalone(e.GgsrunCfg.Scriptid)
	if err == nil {
		json.Unmarshal(body, &p)
		e.Msg = append(e.Msg, fmt.Sprintf("Filename is '%s'.", p.FileName))
	}
	e.Msg = append(e.Msg, fmt.Sprintf("Project ID is '%s'.", e.Scriptid))
	return e
}

// filesInProjectRemover : Remove files in project.
func (e *ExecutionContainer) filesInProjectRemover() *ExecutionContainer {
	temp := e.Project
	var outr []string
	for _, elm := range e.UpFiles {
		res, removed := removeEle(temp, elm)
		if removed {
			outr = append(outr, elm)
		}
		temp = res
	}
	if len(temp.Files) == 1 {
		fmt.Fprintf(os.Stderr, "Error: You cannot remove all files except for 'appsscript.json' in the project.\n")
		os.Exit(1)
	}
	e.Project = temp
	p := e.convExecutionContainerToFileInf()
	body, err, _ := p.ChkBoundOrStandalone(e.GgsrunCfg.Scriptid)
	if err == nil {
		json.Unmarshal(body, &p)
		e.Msg = append(e.Msg, fmt.Sprintf("Filename is '%s'.", p.FileName))
	}
	e.Msg = append(e.Msg, fmt.Sprintf("Project ID is '%s'.", e.Scriptid))
	if len(outr) == 0 {
		fmt.Fprintf(os.Stderr, "[ %s ] were not found in the project. No files were removed from the project.\n", strings.Join(e.UpFiles, ", "))
		os.Exit(1)
	} else {
		e.Msg = append(e.Msg, fmt.Sprintf("Files of [ %s ] were removed from the project.", strings.Join(outr, ", ")))
	}
	return e
}

// removeEle : Remove an element from an array.
func removeEle(project *Project, elm string) (*Project, bool) {
	temp := &Project{}
	ff := strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1)
	if ff != "appsscript" {
		for _, v := range project.Files {
			if v.Name != ff {
				temp.Files = append(temp.Files, v)
			}
		}
	} else {
		return project, false
	}
	if len(project.Files) != len(temp.Files) {
		return temp, true
	}
	return temp, false
}
