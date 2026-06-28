package app

import (
	"bytes"
	"fmt"
	"net/url"
	"path"

	"ggsrun/internal/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// recoverProject : recover command
func recoverProject(c *cli.Context) error {
	a := defAuthContainer(c)
	if !c.Bool("jsonparser") {
		a.Spinner, _ = pterm.DefaultSpinner.WithRemoveWhenDone(false).Start("Initializing recovery process...")
	}
	e := a.ggsrunIni(c).
		goauth().
		defExecutionContainer()

	scriptID := e.GgsrunCfg.Scriptid
	if scriptID == "" {
		e.FailStatus("Configuration Error")
		pterm.Error.Println("Script ID is not configured. Please supply it via '-i [Script ID]' or configure it in ggsrun.cfg.")
		utl.Exit(1)
	}

	e.UpdateStatus("Rebuilding recovery project payload...")

	recoveryScript := `const doPost = (e) => ggsrunif.WebApps(e, "pass1");
const ExecutionApi = (e) => ggsrunif.ExecutionApi(e);`

	recoveryManifest := `{
  "dependencies": {
    "enabledAdvancedServices": [],
    "libraries": [
      {
        "userSymbol": "ggsrunif",
        "version": "0",
        "libraryId": "115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov",
        "developmentMode": true
      }
    ]
  },
  "exceptionLogging": "STACKDRIVER",
  "executionApi": {
    "access": "MYSELF"
  },
  "runtimeVersion": "V8",
  "timeZone": "Asia/Tokyo",
  "webapp": {
    "executeAs": "USER_DEPLOYING",
    "access": "MYSELF"
  }
}`

	// Build a project containing ONLY the recovery files
	e.Project = &Project{
		Files: []File{
			{
				Name:   "appsscript",
				Type:   "JSON",
				Source: recoveryManifest,
			},
			{
				Name:   "ggsrun",
				Type:   "SERVER_JS",
				Source: recoveryScript,
			},
		},
	}

	e.UpdateStatus("Uploading recovery project to Google Cloud...")
	
	updatedProjBytes, err := json.Marshal(e.Project)
	if err != nil {
		e.FailStatus("Serialization Error")
		pterm.Error.Printf("Failed to serialize recovery project: %v\n", err)
		utl.Exit(1)
	}

	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, scriptID+"/content")
	r := &utl.RequestParams{
		Method:      "PUT",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(updatedProjBytes),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err := r.FetchAPI()
	if err != nil {
		e.FailStatus("Upload Failed")
		pterm.Error.Printf("Failed to upload recovery project: %v. Response: %s\n", err, string(res))
		utl.Exit(1)
	}

	// Deploy as a new version to ensure Execution API / Web App runs correctly
	e.UpdateStatus("Creating new recovery version...")
	type versionReq struct {
		Description string `json:"description"`
	}
	vReq := versionReq{Description: "ggsrun recovery version"}
	vReqBytes, _ := json.Marshal(vReq)
	u, _ = url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, scriptID+"/versions")
	r = &utl.RequestParams{
		Method:      "POST",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(vReqBytes),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err = r.FetchAPI()
	if err != nil {
		e.FailStatus("Version Creation Failed")
		pterm.Error.Printf("Failed to create version: %v\n", err)
		utl.Exit(1)
	}

	var vRes struct {
		VersionNumber int `json:"versionNumber"`
	}
	if err := json.Unmarshal(res, &vRes); err != nil {
		e.FailStatus("Parsing Error")
		pterm.Error.Printf("Failed to parse version response: %v\n", err)
		utl.Exit(1)
	}

	e.UpdateStatus(fmt.Sprintf("Deploying recovery version %d...", vRes.VersionNumber))
	
	// List deployments to find target ID
	u, _ = url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, scriptID+"/deployments")
	r = &utl.RequestParams{
		Method:      "GET",
		APIURL:      u.String(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err = r.FetchAPI()
	if err != nil {
		e.FailStatus("Deployments Query Failed")
		pterm.Error.Printf("Failed to list deployments: %v\n", err)
		utl.Exit(1)
	}

	var deplList struct {
		Deployments []struct {
			DeploymentID     string `json:"deploymentId"`
			DeploymentConfig struct {
				VersionNumber    int    `json:"versionNumber"`
				ManifestFileName string `json:"manifestFileName"`
				Description      string `json:"description"`
			} `json:"deploymentConfig"`
			EntryPoints []struct {
				EntryPointType string `json:"entryPointType"`
			} `json:"entryPoints"`
		} `json:"deployments"`
	}
	json.Unmarshal(res, &deplList)

	// Update existing API_EXECUTABLE or WEB_APP deployments
	for _, d := range deplList.Deployments {
		isTarget := false
		for _, ep := range d.EntryPoints {
			if ep.EntryPointType == "API_EXECUTABLE" || ep.EntryPointType == "WEB_APP" {
				isTarget = true
				break
			}
		}
		if isTarget {
			type updateDeplReq struct {
				DeploymentConfig struct {
					VersionNumber    int    `json:"versionNumber"`
					ManifestFileName string `json:"manifestFileName"`
					Description      string `json:"description"`
				} `json:"deploymentConfig"`
			}
			var reqBody updateDeplReq
			reqBody.DeploymentConfig.VersionNumber = vRes.VersionNumber
			reqBody.DeploymentConfig.ManifestFileName = "appsscript"
			reqBody.DeploymentConfig.Description = "ggsrun recovery deployment"
			reqBodyBytes, _ := json.Marshal(reqBody)

			u, _ = url.Parse(appsscriptapi)
			u.Path = path.Join(u.Path, scriptID+"/deployments/"+d.DeploymentID)
			r = &utl.RequestParams{
				Method:      "PUT",
				APIURL:      u.String(),
				Data:        bytes.NewBuffer(reqBodyBytes),
				Accesstoken: e.GgsrunCfg.Accesstoken,
				Dtime:       30,
			}
			_, _ = r.FetchAPI()
		}
	}

	if !c.Bool("jsonparser") {
		a.Spinner.Success("Recovery completed! GAS project has been restored to the pristine ggsrun state.")
	} else {
		fmt.Println(`{"status":"success","message":"Recovery completed"}`)
	}

	return nil
}
