// Package main (handler.go) :
// Lightweight command routers for generic ggsrun functions.
package app

import (
	"fmt"
	"ggsrun/internal/utl"
	"os"
	"path/filepath"
	"time"

	json "github.com/goccy/go-json"
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

	pterm.DefaultSection.Println("ggsrun Configuration Status")
	pterm.Info.Printf("ggsrun Version: v%s\n\n", c.App.Version)

	pterm.Info.Println("Priority of ggsrun.cfg search paths:")
	pterm.Info.Println("  1. --config <dir> flag")
	pterm.Info.Println("  2. --credentials <file> flag (uses parent directory)")
	pterm.Info.Println("  3. Current Working Directory")
	pterm.Info.Println("  4. Environment variable $GGSRUN_CFG_PATH (fallback)")
	fmt.Println()

	checkFile := func(dir string) string {
		if dir == "" {
			return "Not Set"
		}
		p := filepath.Join(dir, "ggsrun.cfg")
		absP, _ := filepath.Abs(p)
		if _, err := os.Stat(absP); err == nil {
			return fmt.Sprintf("Found (%s)", absP)
		}
		return fmt.Sprintf("Not Found (%s)", absP)
	}

	pterm.Info.Println("Checking search paths:")
	if a.InitVal.customConfig != "" {
		pterm.Info.Printf("  1. --config:          %s\n", checkFile(a.InitVal.customConfig))
	} else {
		pterm.Info.Println("  1. --config:          Not Set")
	}

	if a.InitVal.customCred != "" {
		pterm.Info.Printf("  2. --credentials:     %s\n", checkFile(filepath.Dir(a.InitVal.customCred)))
	} else {
		pterm.Info.Println("  2. --credentials:     Not Set")
	}

	pterm.Info.Printf("  3. Current Directory: %s\n", checkFile(a.InitVal.workdir))
	pterm.Info.Printf("  4. GGSRUN_CFG_PATH:   %s\n", checkFile(a.InitVal.envConfig))
	fmt.Println()

	resolvedPath := a.resolveConfigFile()
	absResolvedPath, _ := filepath.Abs(resolvedPath)
	pterm.Success.Printf("Active configuration file in use:\n  %s\n\n", absResolvedPath)

	maskString := func(s string) string {
		if s == "" {
			return ""
		}
		if len(s) <= 8 {
			return "********"
		}
		return s[:4] + "..." + s[len(s)-4:]
	}

	type MaskedCfg struct {
		Scriptid            string   `json:"script_id"`
		Clientid            string   `json:"client_id"`
		Clientsecret        string   `json:"client_secret"`
		Refreshtoken        string   `json:"refresh_token"`
		Accesstoken         string   `json:"access_token"`
		Expiresin           int64    `json:"expires_in"`
		Scopes              []string `json:"scopes"`
		ExecutionApiChecked bool     `json:"execution_api_checked"`
		WebappsUrl          string   `json:"webapps_url"`
	}

	masked := MaskedCfg{
		Scriptid:            a.GgsrunCfg.Scriptid,
		Clientid:            maskString(a.GgsrunCfg.Clientid),
		Clientsecret:        maskString(a.GgsrunCfg.Clientsecret),
		Refreshtoken:        maskString(a.GgsrunCfg.Refreshtoken),
		Accesstoken:         maskString(a.GgsrunCfg.Accesstoken),
		Expiresin:           a.GgsrunCfg.Expiresin,
		Scopes:              a.GgsrunCfg.Scopes,
		ExecutionApiChecked: a.GgsrunCfg.ExecutionApiChecked,
		WebappsUrl:          a.GgsrunCfg.WebappsUrl,
	}

	if jsonBytes, err := json.MarshalIndent(masked, "", "  "); err == nil {
		pterm.Info.Println("Active Configuration Contents:")
		fmt.Println(string(jsonBytes))
		fmt.Println()
	}

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
