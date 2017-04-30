// Package main (projectupdater.go) :
// These methods are for updating project.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanaikech/ggsrun/utl"
	"github.com/urfave/cli"
)

// projectUpdateControl : Main method for updating project.
func (e *ExecutionContainer) projectUpdateControl(c *cli.Context) *utl.FileInf {
	if len(c.String("filename")) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No Files. Please set them using '-f [ File name ]'. ")
		os.Exit(1)
	}
	if len(c.String("projectid")) > 0 {
		e.GgsrunCfg.Scriptid = c.String("projectid")
	}
	return e.defUpdateProjectContainer(c).
		projectBackup(c).
		ProjectMaker().
		projectUpdate().
		dispUpdateProjectContainer()
}

// ProjectMaker : Recreates the project using uploaded scripts.
func (e *ExecutionContainer) ProjectMaker() *ExecutionContainer {
	for _, elm := range e.UpFiles {
		if filepath.Ext(elm) == ".gs" ||
			filepath.Ext(elm) == ".gas" ||
			filepath.Ext(elm) == ".js" ||
			filepath.Ext(elm) == ".htm" ||
			filepath.Ext(elm) == ".html" {
			filedata := &File{
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
				Source: utl.ConvGasToUpload(elm),
			}
			var overwrite bool
			for i, v := range e.Project.Files {
				if v.Name == filedata.Name {
					e.Project.Files[i].Source = filedata.Source
					e.Msg = append(e.Msg, fmt.Sprintf("Script '%s' in project was overwritten.", v.Name))
					overwrite = true
				}
			}
			if !overwrite {
				e.Project.Files = append(e.Project.Files, *filedata)
			}
		}
	}
	e.Msg = append(e.Msg, fmt.Sprintf("Project ID '%s' was uploaded.", e.Scriptid))
	return e
}
