package app

import (
	"bytes"
	"context"
	"fmt"
	"ggsrun/internal/utl"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	json "github.com/goccy/go-json"
	"github.com/urfave/cli"
)

// RunTUIFunc is a callback initialized by the internal/tui package.
var RunTUIFunc func(c *cli.Context) error

// TUIProgressCallback is a function that can be set by the TUI to receive status updates.
var TUIProgressCallback func(string)


// GetAuthenticatedAuthContainer initializes and returns an authenticated AuthContainer.
func GetAuthenticatedAuthContainer(c *cli.Context) *AuthContainer {
	return defAuthContainer(c).ggsrunIni(c).goauth()
}

// DefDownloadContainerExported returns a FileInf initialized for downloads using defDownloadContainer.
func (a *AuthContainer) DefDownloadContainerExported(c *cli.Context) *utl.FileInf {
	return a.defDownloadContainer(c)
}

// DefUploadContainerExported returns a FileInf initialized for uploads using defUploadContainer.
func (a *AuthContainer) DefUploadContainerExported(c *cli.Context) *utl.FileInf {
	return a.defUploadContainer(c)
}

// TuiUpload performs a file or directory upload using concurrentUpload.
func TuiUpload(c *cli.Context, a *AuthContainer) (interface{}, error) {
	return concurrentUpload(context.Background(), c, a)
}

// TuiDownload performs a file or directory download using concurrentDownload.
func TuiDownload(c *cli.Context, a *AuthContainer) (interface{}, error) {
	return concurrentDownload(context.Background(), c, a)
}

// TuiExecuteGas invokes a function in the Google Apps Script project.
func TuiExecuteGas(scriptID string, functionName string, args []interface{}, a *AuthContainer) (string, error) {
	epara := struct {
		Function   string        `json:"function"`
		Parameters []interface{} `json:"parameters,omitempty"`
		DevMode    bool          `json:"devMode"`
	}{
		Function:   functionName,
		Parameters: args,
		DevMode:    true,
	}
	re, err := json.Marshal(epara)
	if err != nil {
		return "", err
	}
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      "https://script.googleapis.com/v1/scripts/" + scriptID + ":run",
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       370,
	}
	body, err := r.FetchAPI()
	if err != nil {
		return string(body), err
	}
	return string(body), nil
}

// TuiUpdateDriveMetadata updates name, description, and/or modifiedTime of a Drive file.
func TuiUpdateDriveMetadata(fileID string, name string, description string, modTime *time.Time, a *AuthContainer) error {
	metadata := map[string]interface{}{}
	if name != "" {
		metadata["name"] = name
	}
	if description != "" {
		metadata["description"] = description
	}
	if modTime != nil {
		metadata["modifiedTime"] = modTime.Format(time.RFC3339)
	}

	re, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	r := &utl.RequestParams{
		Method:      "PATCH",
		APIURL:      "https://www.googleapis.com/drive/v3/files/" + fileID,
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	_, err = r.FetchAPI()
	return err
}

// TuiCopyDriveFile copies a Drive file to a new parent folder, optionally renaming it.
func TuiCopyDriveFile(fileID string, newName string, parentID string, a *AuthContainer) (string, error) {
	metadata := map[string]interface{}{}
	if newName != "" {
		metadata["name"] = newName
	}
	if parentID != "" {
		metadata["parents"] = []string{parentID}
	}

	re, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      "https://www.googleapis.com/drive/v3/files/" + fileID + "/copy",
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		return "", err
	}
	var res struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &res)
	return res.ID, nil
}

// TuiCreateDriveFolder creates a new folder on Google Drive.
func TuiCreateDriveFolder(name string, parentID string, a *AuthContainer) (string, error) {
	metadata := map[string]interface{}{
		"name":     name,
		"mimeType": "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		metadata["parents"] = []string{parentID}
	}

	re, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      "https://www.googleapis.com/drive/v3/files",
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	body, err := r.FetchAPI()
	if err != nil {
		return "", err
	}
	var res struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &res)
	return res.ID, nil
}

// TuiRunExe1 runs the exe1 command logic and returns the response payload, start time, and last process ID.
func TuiRunExe1(c *cli.Context, a *AuthContainer) (resp string, startTime time.Time, lastProcessID string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal process exited: %v", r)
		}
	}()

	e := a.defExecutionContainer()

	// Handle script ID override if provided
	if c.String("scriptid") != "" {
		e.GgsrunCfg.Scriptid = c.String("scriptid")
	}

	// Re-initialize/override execution function from dynamic TUI context
	fSlice := c.StringSlice("function")
	if len(fSlice) > 0 && fSlice[0] != "" {
		e.Param.Function = fSlice[0]
	} else if c.String("function") != "" {
		e.Param.Function = c.String("function")
	}

	// Set up robust cleanup with exit and signal hooks
	var rollbackOnce sync.Once
	performRollback := func() {
		rollbackOnce.Do(func() {
			needsUpdate := false
			// Always restore the original project state when executing via TUI/FD
			if len(e.InitVal.originalFiles) > 0 {
				e.UpdateStatus("Restoring original script project state...")
				e.Project.Files = e.InitVal.originalFiles
				needsUpdate = true
			} else {
				if e.InitVal.tempFileNameToCleanup != "" || len(e.InitVal.uploadedFilesToCleanup) > 0 {
					if e.InitVal.tempFileNameToCleanup != "" {
						var newFiles []File
						for _, f := range e.Project.Files {
							if f.Name != e.InitVal.tempFileNameToCleanup {
								newFiles = append(newFiles, f)
							}
						}
						e.Project.Files = newFiles
						needsUpdate = true
					}
					if len(e.InitVal.uploadedFilesToCleanup) > 0 {
						cleanupMap := make(map[string]bool)
						for _, name := range e.InitVal.uploadedFilesToCleanup {
							cleanupMap[name] = true
						}
						var newFiles []File
						for _, f := range e.Project.Files {
							if !cleanupMap[f.Name] {
								newFiles = append(newFiles, f)
							}
						}
						e.Project.Files = newFiles
						needsUpdate = true
					}
				}
			}
			if needsUpdate {
				e.projectUpdate2()
			}
		})
	}

	utl.CleanUpHandler = performRollback
	defer func() {
		utl.CleanUpHandler = nil
		performRollback()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		performRollback()
		os.Exit(1)
	}()
	defer func() {
		signal.Stop(sigChan)
	}()

	if err := e.autoValidateAndDeployManifest(c, "e1"); err != nil {
		return "", time.Time{}, "", err
	}

	e.exe1Function(c).
		executionAPIwithoutServer(c).
		esenderForExe1(c)

	if len(e.Msg) > 0 {
		e.FeedBackData.Response.Result.Message = e.Msg
	}

	if len(e.FeedBackData.Error.Message) > 0 {
		var errMsg string
		if len(e.FeedBackData.Error.Detailes) > 0 && e.FeedBackData.Error.Detailes[0].ErrorMessage != "" {
			errMsg = e.FeedBackData.Error.Detailes[0].ErrorMessage
		} else {
			errMsg = e.FeedBackData.Error.Message
		}
		b, _ := json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
		return string(b), e.InitVal.pstart, e.LastProcessID, fmt.Errorf("Script Error on GAS side: %s (code: %d)", errMsg, e.FeedBackData.Error.Code)
	}

	b, err := json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
	if err != nil {
		return "", time.Time{}, "", err
	}
	return string(b), e.InitVal.pstart, e.LastProcessID, nil
}

// TuiFetchLogsOnly fetches logs from Cloud Logging for a completed TUI execution session.
func TuiFetchLogsOnly(a *AuthContainer, startTime time.Time, functionName string, lastProcessID string) ([]GASLog, error) {
	e := a.defExecutionContainer()
	// Create a dummy cli.Context
	c := &cli.Context{}
	logs := e.fetchGASLogs(c, startTime, functionName, lastProcessID)
	return logs, nil
}

// TuiRunExe2 runs the exe2 command logic and returns the response payload as a JSON string.
func TuiRunExe2(c *cli.Context, a *AuthContainer) (resp string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal process exited: %v", r)
		}
	}()

	e := a.defExecutionContainer()

	// Handle script ID override if provided
	if c.String("scriptid") != "" {
		e.GgsrunCfg.Scriptid = c.String("scriptid")
	}

	if err := e.autoValidateAndDeployManifest(c, "e2"); err != nil {
		return "", err
	}

	e.exe2Function(c)

	if len(e.FeedBackData.Error.Message) > 0 {
		var errMsg string
		if len(e.FeedBackData.Error.Detailes) > 0 && e.FeedBackData.Error.Detailes[0].ErrorMessage != "" {
			errMsg = e.FeedBackData.Error.Detailes[0].ErrorMessage
		} else {
			errMsg = e.FeedBackData.Error.Message
		}
		b, _ := json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
		return string(b), fmt.Errorf("Script Error on GAS side: %s (code: %d)", errMsg, e.FeedBackData.Error.Code)
	}

	b, err := json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TuiRunWebApps runs the webapps command logic and returns the response payload as a JSON string.
func TuiRunWebApps(c *cli.Context, a *AuthContainer) (resp string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal process exited: %v", r)
		}
	}()

	e := a.defExecutionContainer()

	// Handle script ID override if provided
	if c.String("scriptid") != "" {
		e.GgsrunCfg.Scriptid = c.String("scriptid")
	}

	var rawScript string
	var isTemp bool
	var tempFileName string

	if c.String("stringscript") != "" {
		rawScript = c.String("stringscript")
		isTemp = true
	} else {
		scriptFile := c.String("scriptfile")
		if scriptFile == "" {
			return "", fmt.Errorf("no script. Please set GAS script using '-s' or '--stringscript'")
		}
		rawScript = utl.ConvGasToPut(c)
		isTemp = true
	}

	if isTemp && e.GgsrunCfg.Scriptid != "" && e.GgsrunCfg.Accesstoken != "" && c.String("scriptfile") != "" {
		timestamp := time.Now().Format("20060102150405")
		tempFileName = "ggsrun_web_temp_" + timestamp

		e.projectBackup(c)

		defer func() {
			var newFiles []File
			for _, f := range e.Project.Files {
				if f.Name != tempFileName {
					newFiles = append(newFiles, f)
				}
			}
			e.Project.Files = newFiles
			e.projectUpdate2()
			e.autoValidateAndDeployManifest(c, "w")
		}()

		filedata := File{
			Name:   tempFileName,
			Type:   "SERVER_JS",
			Source: rawScript,
		}
		e.Project.Files = append(e.Project.Files, filedata)
		e.projectUpdate2()

		if err := e.autoValidateAndDeployManifest(c, "w"); err != nil {
			return "", err
		}
	} else {
		if err := e.autoValidateAndDeployManifest(c, "w"); err != nil {
			return "", err
		}
	}

	val := c.String("value")
	var argStr string
	if val != "" {
		if numRe.MatchString(val) || arrayRe.MatchString(val) || objRe.MatchString(val) || val == "true" || val == "false" || val == "null" {
			argStr = val
		} else {
			argStr = fmt.Sprintf("%q", val)
		}
	}

	wrappedScript := fmt.Sprintf(`(function() {
%s
if (typeof main !== 'undefined') {
	return main(%s);
}
return "Execution completed, but main() was not defined in the local script.";
})()`, rawScript, argStr)

	quotedBytes, _ := json.Marshal(wrappedScript)
	quotedScript := string(quotedBytes)

	e.webAppswithServerForExe3(quotedScript, c)

	if len(e.FeedBackData.Error.Message) > 0 {
		var errMsg string
		if len(e.FeedBackData.Error.Detailes) > 0 && e.FeedBackData.Error.Detailes[0].ErrorMessage != "" {
			errMsg = e.FeedBackData.Error.Detailes[0].ErrorMessage
		} else {
			errMsg = e.FeedBackData.Error.Message
		}
		b, _ := json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
		return string(b), fmt.Errorf("Script Error on GAS side: %s (code: %d)", errMsg, e.FeedBackData.Error.Code)
	}

	b, err := json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TuiGetFileContent downloads the raw content of a Google Drive file.
func TuiGetFileContent(fileID string, a *AuthContainer) ([]byte, error) {
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      "https://www.googleapis.com/drive/v3/files/" + fileID + "?alt=media",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       60,
	}
	body, err := r.FetchAPI()
	if err != nil {
		return nil, err
	}
	return body, nil
}

type progressReader struct {
	io.ReadCloser
	onProgress func(n int)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.ReadCloser.Read(p)
	if n > 0 {
		pr.onProgress(n)
	}
	return n, err
}


