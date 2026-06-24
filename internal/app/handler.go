// Package main (handler.go) :
// Lightweight command routers for generic ggsrun functions.
package app

import (
	"ggsrun/internal/utl"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// updateProject : Updates projects and scripts
func updateProject(c *cli.Context) error {
	a := defAuthContainer(c)
	res := a.
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		projectUpdateControl(c)
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}

// revisionFiles : Retrieves revision IDs and downloads revision files.
func revisionFiles(c *cli.Context) error {
	a := defAuthContainer(c)
	res := a.
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetRevisionList(c)
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}

// showFileList : Shows file list on Google Drive
func showFileList(c *cli.Context) error {
	a := defAuthContainer(c)
	res := a.
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetFileList(c)
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}

// searchFilesByQueryAndRegex : Search files on Google Drive using search query and regex.
func searchFilesByQueryAndRegex(c *cli.Context) error {
	a := defAuthContainer(c)
	res := a.
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		SearchFiles()
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}

// managePermissions : Manage permissions.
func managePermissions(c *cli.Context) error {
	a := defAuthContainer(c)
	res := a.
		ggsrunIni(c).
		goauth().
		defPermissionsContainer(c).
		ManagePermissions()
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}

// getDriveInformation : Get drive information.
func getDriveInformation(c *cli.Context) error {
	a := defAuthContainer(c)
	res := a.
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetDriveInformation()
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}

// reAuth : Retrieve tokens again.
func reAuth(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		reAuth()
	pterm.Success.Println("Done.")
	return nil
}

// quickSetup : Simplified onboarding flow.
func quickSetup(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIniForSetup(c).
		quickSetup()
	pterm.Success.Println("Done.")
	return nil
}

// checkStatus : Health check
func checkStatus(c *cli.Context) error {
	a := defAuthContainer(c).ggsrunIni(c).goauth()
	pterm.Success.Println("Status: Authentication successful!")
	pterm.Info.Printf("Access Token valid. Length: %d characters.\n", len(a.GgsrunCfg.Accesstoken))
	pterm.Info.Printf("Expiration time: %v\n", time.Unix(a.GgsrunCfg.Expiresin, 0).Format(time.RFC3339))
	return nil
}

// commandNotFound :
func commandNotFound(c *cli.Context, command string) {
	pterm.Error.Printf("'%s' is not a %s command. Check '%s --help' or '%s -h'.\n", command, c.App.Name, c.App.Name, c.App.Name)
	utl.Exit(2)
}
