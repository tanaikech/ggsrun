// Package main (projectupdater.go) :
// These methods are for updating project.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanaikech/ggsrun/utl"
	"github.com/urfave/cli"
)

// projectUpdateControl : Main method for updating project.
func (e *ExecutionContainer) projectUpdateControl(c *cli.Context) *utl.FileInf {
	if len(c.String("projectid")) > 0 {
		e.GgsrunCfg.Scriptid = c.String("projectid")
		if len(c.String("filename")) > 0 {
			return e.defUpdateProjectContainer(c).
				projectBackup(c).
				ProjectMaker().
				projectUpdate().
				dispUpdateProjectContainer()
		}
		if c.Bool("rearrange") {
			e.defUpdateProjectContainer(c).
				projectBackup(c).
				rearrangeByTerminal()
		}
		if len(c.String("rearrangewithfile")) > 0 {
			var data []string
			f, err := os.Open(c.String("rearrangewithfile"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Script '%s' is not found.\n", c.String("rearrangewithfile"))
				os.Exit(1)
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if scanner.Text() == "end" {
					break
				}
				if scanner.Text() != "" {
					data = append(data, scanner.Text())
				}
			}
			if scanner.Err() != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", scanner.Err())
				os.Exit(1)
			}
			e.defUpdateProjectContainer(c).
				projectBackup(c).
				rearrangeByFile(data)
		}
	} else {
		e.Msg = append(e.Msg, "Error: No options. Please check HELP using 'ggsrun ud -help'.")
	}
	return e.dispUpdateProjectContainer()
}

// ProjectMaker : Recreates the project using uploaded scripts.
func (e *ExecutionContainer) ProjectMaker() *ExecutionContainer {
	for _, elm := range e.UpFiles {
		if filepath.Ext(elm) == ".gs" ||
			filepath.Ext(elm) == ".gas" ||
			filepath.Ext(elm) == ".js" ||
			filepath.Ext(elm) == ".htm" ||
			filepath.Ext(elm) == ".html" ||
			filepath.Ext(elm) == ".json" {
			filedata := &File{
				Name: strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1),
				Type: func(ex string) string {
					var scripttype string
					switch ex {
					case ".gs", ".gas", ".js":
						scripttype = "server_js"
					case ".htm", ".html":
						scripttype = "html"
					case ".json":
						scripttype = "json"
					}
					return scripttype
				}(filepath.Ext(elm)),
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
		}
	}
	e.Msg = append(e.Msg, fmt.Sprintf("Project ID '%s' was uploaded.", e.Scriptid))
	return e
}
