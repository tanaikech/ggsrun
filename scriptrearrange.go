// Package main (scriptrearrange.go) :
// These methods are for rearranging scripts in a project.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	rearrange "github.com/tanaikech/go-rearrange"
)

// rearrangeByTerminal : Rearranging scripts in a project using go-rearrange.
func (e *ExecutionContainer) rearrangeByTerminal() *ExecutionContainer {
	var baseProject Project
	baseProject = *e.Project
	var scripts []string
	for _, f := range e.Project.Files {
		scripts = append(scripts, f.Name)
	}
	changedIndx, _, err := rearrange.Do(scripts, 3, false, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	var input string
	fmt.Printf("## Please be careful.\n")
	fmt.Printf("## When the script is rearranged, the revision of script is reset once.\n")
	fmt.Printf("Reflect the rearranged result? [y or n] ... ")
	if _, err := fmt.Scan(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if input == "y" {
		s := spinner.New([]string{"/", "|", "\\", "|"}, 100*time.Millisecond)
		s.UpdateSpeed(200 * time.Millisecond)
		fmt.Printf("Please wait a moment...")
		s.Start()
		e.rearrange(baseProject, changedIndx)
		s.Stop()
		fmt.Printf("\n")
		return e
	}
	e.Msg = append(e.Msg, "Scripts of project were NOT rearranged.")
	return e
}

// rearrange : Rearranging scripts in a project using a configuration file.
func (e *ExecutionContainer) rearrangeByFile(data []string) *ExecutionContainer {
	var baseProject Project
	baseProject = *e.Project
	var temp []string
	dupChk := map[string]bool{}
	for _, e := range data {
		if !dupChk[e] {
			dupChk[e] = true
			temp = append(temp, e)
		}
	}
	if len(temp) == len(data) {
		if len(e.Project.Files) == len(data) {
			cn := 0
			for i, e := range e.Project.Files {
				if e.Name == data[i] {
					cn += 1
				}
			}
			if cn != len(e.Project.Files) {
				cn = 0
				var changedIndx []string
				for _, f := range data {
					for i, g := range e.Project.Files {
						if g.Name == f {
							cn += 1
							changedIndx = append(changedIndx, strconv.Itoa(i))
						}
					}
				}
				if cn == len(e.Project.Files) {
					e.rearrange(baseProject, changedIndx)
					return e
				}
				e.Msg = append(e.Msg, "Error: Script names of inputted file are different for script names in project.")
				return e
			}
			e.Msg = append(e.Msg, "Error: Order of inputted file are the same to the order in project.")
			return e
		}
		e.Msg = append(e.Msg, "Error: Number of script names of inputted file are different for number of scripts in project.")
		return e
	}
	e.Msg = append(e.Msg, "Error: There are duplicated names in script names of inputted file.")
	return e
}

// rearrange : Main method for rearranging scripts.
func (e *ExecutionContainer) rearrange(baseProject Project, changedIndx []string) {
	var temp1 Project
	const layout = "20060102_150405_"
	t := time.Now()
	temp1.Files = append(temp1.Files, File{
		Name:   "Dummy_" + t.Format(layout) + t.AddDate(0, 0, 2).Weekday().String(),
		Source: "// This is a dummy.",
		Type:   "SERVER_JS",
	})
	temp1.Files = append(temp1.Files, File{
		Name:   "appsscript",
		Source: "{}",
		Type:   "JSON",
	})
	e.Project = &temp1
	e.projectUpdate2()
	var temp2 Project
	for _, e := range changedIndx {
		idx, _ := strconv.Atoi(e)
		temp2.Files = append(temp2.Files, baseProject.Files[idx])
	}
	e.Project = &temp2
	e.projectUpdate2()
	var from, to []string
	for i, f := range e.Project.Files {
		from = append(from, baseProject.Files[i].Name)
		to = append(to, f.Name)
	}
	msg := fmt.Sprintf("Scripts in project were rearranged from [%s] to [%s].", strings.Join(from, ", "), strings.Join(to, ", "))
	e.Msg = []string{msg}
}

// getRearrangeTemplate : Retrieve data from template file for rearranging.
func getRearrangeTemplate(templateFile string) []string {
	var data []string
	f, err := os.Open(templateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Script '%s' is not found.\n", templateFile)
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
	return data
}
