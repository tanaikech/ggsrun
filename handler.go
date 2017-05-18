// Package main (handler.go) :
// Handler for ggsrun
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tanaikech/ggsrun/utl"
	"github.com/urfave/cli"
)

// exeAPIWithout : exe1
// Update project and Execution API withour server script.
func exeAPIWithout(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		exe1Function(c).
		executionAPIwithoutServer(c).
		esenderForExe1(c).
		dispResult(c)
	return nil
}

// exeAPIWith : exe2
// No update project. Only execute GAS using Execution API with server script.
func exeAPIWith(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		exe2Function(c).
		dispResult(c)
	return nil
}

// webAppsWith : exe3
// No update project. Only execute GAS using Web Apps with server script.
func webAppsWith(c *cli.Context) error {
	defExecutionContainerWebApps().
		webAppswithServerForExe3(utl.ConvGasToRun(c), c).
		dispResult(c)
	return nil
}

// downloadFiles : Download files from Google Drive.
func downloadFiles(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetFileinf().
		Downloader(c)
	dispTransferResult(c, res)
	return nil
}

// uploadFiles : Uploads files
func uploadFiles(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defUploadContainer(c).
		Uploader(c)
	dispTransferResult(c, res)
	return nil
}

// updateProject : Updates projects and scripts
func updateProject(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		projectUpdateControl(c)
	dispTransferResult(c, res)
	return nil
}

// revisionFiles : Retrieves revision IDs and downloads revision files.
func revisionFiles(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetRevisionList(c)
	dispTransferResult(c, res)
	return nil
}

// showFileList : Shows file list on Google Drive
func showFileList(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetFileList(c)
	dispTransferResult(c, res)
	return nil
}

// reAuth : Retrieve tokens again.
func reAuth(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		reAuth()
	fmt.Print("Done.")
	return nil
}

// dispResult : Display result
func (e *ExecutionContainer) dispResult(c *cli.Context) {
	var dispRes []byte
	if len(e.Msg) > 0 {
		e.FeedBackData.Response.Result.Message = e.Msg
	}
	if c.Bool("jsonparser") {
		dispRes, _ = json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
	} else {
		dispRes, _ = json.Marshal(e.FeedBackData.Response.Result)
	}
	if c.Bool("onlyresult") {
		if c.Bool("jsonparser") {
			onlyres, _ := json.MarshalIndent(e.FeedBackData.Response.Result.Result, "", "  ")
			fmt.Printf("%s\n", string(onlyres))
		} else {
			onlyres, _ := json.Marshal(e.FeedBackData.Response.Result.Result)
			fmt.Printf("%s\n", string(onlyres))
		}
	} else {
		fmt.Printf("%v\n", string(dispRes))
	}
}

// dispTransferResult : Display result
func dispTransferResult(c *cli.Context, f *utl.FileInf) {
	var dispRes []byte
	if c.Bool("jsonparser") {
		dispRes, _ = json.MarshalIndent(f, "", "  ")
	} else {
		dispRes, _ = json.Marshal(f)
	}
	fmt.Printf("%s\n", string(dispRes))
}

// commandNotFound :
func commandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "'%s' is not a %s command. Check '%s --help' or '%s -h'.", command, c.App.Name, c.App.Name, c.App.Name)
	os.Exit(2)
}
