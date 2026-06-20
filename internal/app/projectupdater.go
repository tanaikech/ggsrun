// Package main (projectupdater.go) :
// These methods are for updating project.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ggsrun/internal/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// projectUpdateControl : Main method for updating project.
func (e *ExecutionContainer) projectUpdateControl(c *cli.Context) *utl.FileInf {
	if len(c.String("projectid")) > 0 {
		e.GgsrunCfg.Scriptid = c.String("projectid")
		if len(c.String("filename")) > 0 {
			e.defUpdateProjectContainer(c)
			isMCP := os.Getenv("GGSRUN_MCP_MODE") == "true" || c.Bool("jsonparser")
			if !isMCP {
				pterm.Info.Println("Target files to overwrite the remote GAS project:")
				list := pterm.BulletListPrinter{}
				for _, f := range e.UpFiles {
					list.Items = append(list.Items, pterm.BulletListItem{Level: 0, Text: f})
				}
				list.Render()
				confirm, err := pterm.DefaultInteractiveConfirm.
					WithDefaultText("Are you sure you want to overwrite the remote GAS project with these files?").
					Show()
				if err != nil || !confirm {
					pterm.Warning.Println("Operation cancelled by user.")
					utl.Exit(1)
				}
			}

			if !c.Bool("deletefiles") {
				return e.
					projectBackup(c).
					ProjectMaker(c).
					projectUpdate2().
					dispUpdateProjectContainer()
			}
			return e.
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

// ProjectMaker : Recreates the project using uploaded scripts.
func (e *ExecutionContainer) ProjectMaker(c *cli.Context) *ExecutionContainer {
	for _, elm := range e.UpFiles {
		if utl.ChkExtention(filepath.Ext(elm)) {
			filedata := &File{
				Name:   strings.Replace(filepath.Base(elm), filepath.Ext(elm), "", -1),
				Type:   utl.ExtToType(filepath.Ext(elm), false),
				Source: utl.ConvGasToUpload(elm),
			}

			// Check if file with same name already exists
			var exists bool
			var existingIndex int
			for i, v := range e.Project.Files {
				if v.Name == filedata.Name {
					exists = true
					existingIndex = i
					break
				}
			}

			choice := c.String("conflict")
			if exists && choice == "" && !c.Bool("jsonparser") && os.Getenv("GGSRUN_MCP_MODE") != "true" {
				var err error
				choice, err = pterm.DefaultInteractiveSelect.
					WithDefaultText(fmt.Sprintf("Script '%s' already exists in remote GAS project. Action?", filedata.Name)).
					WithOptions([]string{"overwrite", "add"}).
					Show()
				if err != nil {
					choice = "overwrite" // fallback
				}
			}
			if choice == "" {
				choice = "overwrite" // default
			}

			if exists && choice == "overwrite" {
				e.Project.Files[existingIndex].Source = filedata.Source
				e.Msg = append(e.Msg, fmt.Sprintf("'%s' (%s) in project was overwritten.", filedata.Name, e.Project.Files[existingIndex].Type))
			} else if exists && choice == "add" {
				// Find a unique name
				baseName := filedata.Name
				suffix := 1
				for {
					newName := fmt.Sprintf("%s_%d", baseName, suffix)
					nameExists := false
					for _, v := range e.Project.Files {
						if v.Name == newName {
							nameExists = true
							break
						}
					}
					if !nameExists {
						filedata.Name = newName
						break
					}
					suffix++
				}
				e.Project.Files = append(e.Project.Files, *filedata)
				e.Msg = append(e.Msg, fmt.Sprintf("'%s' (%s) was added to project as '%s'.", baseName, filedata.Type, filedata.Name))
			} else {
				// No duplicate exists, just add normally
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
		pterm.Error.Println("You cannot remove all files except for 'appsscript.json' in the project.")
		utl.Exit(1)
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
		pterm.Warning.Printf("[ %s ] were not found in the project. No files were removed from the project.\n", strings.Join(e.UpFiles, ", "))
		utl.Exit(1)
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
