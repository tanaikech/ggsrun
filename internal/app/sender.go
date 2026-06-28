// Package main (sender.go) :
// These methods are for sending GAS scripts to Google Drive.
package app

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ggsrun/internal/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// Regex cache for parameter evaluation
var (
	numRe   = regexp.MustCompile(`^[+-]?[0-9]*[\.]?[0-9]+$`)
	arrayRe = regexp.MustCompile(`^\[.*\]$`)
	objRe   = regexp.MustCompile(`^{.*}$`)
)

// Exe1Function : Updates the project and executes the script.
func (e *ExecutionContainer) exe1Function(c *cli.Context) *ExecutionContainer {

	if c.String("stringscript") != "" {
		rawScript := c.String("stringscript")
		var err error
		rawScript, err = InjectSandbox(rawScript, c.String("sandbox"))
		if err != nil {
			e.FailStatus("Sandbox Injection Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
		timestamp := time.Now().Format("20060102150405")
		tempFileName := "ggsrun_exe1_temp_" + timestamp
		e.UpdateStatus("Preparing project update and backup (temporary file)...")
		e.projectBackup(c)

		filedata := File{
			Name:   tempFileName,
			Type:   "SERVER_JS",
			Source: rawScript,
		}
		e.Project.Files = append(e.Project.Files, filedata)
		e.projectUpdate2()

		e.InitVal.tempFileNameToCleanup = tempFileName
	} else if isStdinPiped() {
		rawScript, err := readAllStdin()
		if err != nil {
			e.FailStatus("Stdin Read Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
		rawScript, err = InjectSandbox(rawScript, c.String("sandbox"))
		if err != nil {
			e.FailStatus("Sandbox Injection Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
		timestamp := time.Now().Format("20060102150405")
		tempFileName := "ggsrun_exe1_temp_" + timestamp
		e.UpdateStatus("Preparing project update and backup (temporary file)...")
		e.projectBackup(c)

		filedata := File{
			Name:   tempFileName,
			Type:   "SERVER_JS",
			Source: rawScript,
		}
		e.Project.Files = append(e.Project.Files, filedata)
		e.projectUpdate2()

		e.InitVal.tempFileNameToCleanup = tempFileName
	} else if c.String("scriptfile") != "" {
		scriptfileStr := c.String("scriptfile")
		var upFiles []string
		rawFiles := regexp.MustCompile(`\s*,\s*`).Split(scriptfileStr, -1)
		for _, f := range rawFiles {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			fi, err := os.Stat(f)
			if err != nil {
				pterm.Error.Printf("File/Directory not found: %s\n", f)
				utl.Exit(1)
			}
			if fi.IsDir() {
				err = filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						upFiles = append(upFiles, path)
					}
					return nil
				})
				if err != nil {
					pterm.Error.Printf("Error walking directory %s: %v\n", f, err)
					utl.Exit(1)
				}
			} else {
				upFiles = append(upFiles, f)
			}
		}

		if len(upFiles) == 0 {
			pterm.Error.Println("No valid upload files found.")
			utl.Exit(1)
		}

		e.UpdateStatus("Preparing project update and backup...")
		e.projectBackup(c)

		var uploadedNames []string
		for _, elm := range upFiles {
			ext := filepath.Ext(elm)
			base := filepath.Base(elm)
			if base == "appsscript.json" {
				content, err := os.ReadFile(elm)
				if err != nil {
					pterm.Error.Printf("Error reading %s: %v\n", elm, err)
					utl.Exit(1)
				}
				// Robustness: ensure uploaded appsscript.json has executionApi and webapp required configs
				var manifest map[string]interface{}
				if err := json.Unmarshal(content, &manifest); err == nil {
					modified := false

					// Retrieve original manifest if available
					var origManifest map[string]interface{}
					for _, origFile := range e.InitVal.originalFiles {
						if origFile.Name == "appsscript" && strings.ToUpper(origFile.Type) == "JSON" {
							json.Unmarshal([]byte(origFile.Source), &origManifest)
							break
						}
					}

					if origManifest != nil {
						// Merge dependencies
						origDeps, hasOrigDeps := origManifest["dependencies"].(map[string]interface{})
						newDeps, hasNewDeps := manifest["dependencies"].(map[string]interface{})
						if hasOrigDeps {
							if !hasNewDeps {
								manifest["dependencies"] = origDeps
								modified = true
							} else {
								// Merge libraries
								origLibs, _ := origDeps["libraries"].([]interface{})
								newLibs, _ := newDeps["libraries"].([]interface{})
								
								// Local helper to merge libraries
								libMap := make(map[string]interface{})
								getLibKey := func(lib interface{}) string {
									if m, ok := lib.(map[string]interface{}); ok {
										if id, ok := m["libraryId"].(string); ok && id != "" {
											return id
										}
										if sym, ok := m["userSymbol"].(string); ok && sym != "" {
											return sym
										}
									}
									return ""
								}
								for _, lib := range origLibs {
									key := getLibKey(lib)
									if key != "" {
										libMap[key] = lib
									}
								}
								for _, lib := range newLibs {
									key := getLibKey(lib)
									if key != "" {
										libMap[key] = lib
									}
								}
								var mergedLibs []interface{}
								for _, lib := range libMap {
									mergedLibs = append(mergedLibs, lib)
								}
								newDeps["libraries"] = mergedLibs

								// Merge enabledAdvancedServices
								origServices, _ := origDeps["enabledAdvancedServices"].([]interface{})
								newServices, _ := newDeps["enabledAdvancedServices"].([]interface{})
								
								serviceMap := make(map[string]interface{})
								getServiceKey := func(srv interface{}) string {
									if m, ok := srv.(map[string]interface{}); ok {
										if id, ok := m["serviceId"].(string); ok && id != "" {
											return id
										}
										if sym, ok := m["userSymbol"].(string); ok && sym != "" {
											return sym
										}
									}
									return ""
								}
								for _, srv := range origServices {
									key := getServiceKey(srv)
									if key != "" {
										serviceMap[key] = srv
									}
								}
								for _, srv := range newServices {
									key := getServiceKey(srv)
									if key != "" {
										serviceMap[key] = srv
									}
								}
								var mergedServices []interface{}
								for _, srv := range serviceMap {
									mergedServices = append(mergedServices, srv)
								}
								newDeps["enabledAdvancedServices"] = mergedServices

								modified = true
							}
						}

						// Merge executionApi
						if _, ok := manifest["executionApi"]; !ok {
							if origExec, ok := origManifest["executionApi"]; ok {
								manifest["executionApi"] = origExec
								modified = true
							}
						}
						// Merge webapp
						if _, ok := manifest["webapp"]; !ok {
							if origWebapp, ok := origManifest["webapp"]; ok {
								manifest["webapp"] = origWebapp
								modified = true
							}
						}
						// Merge other missing root level keys from original
						for k, v := range origManifest {
							if k == "dependencies" || k == "executionApi" || k == "webapp" {
								continue
							}
							if _, ok := manifest[k]; !ok {
								manifest[k] = v
								modified = true
							}
						}
					}

					// Ensure executionApi exists
					if _, ok := manifest["executionApi"]; !ok {
						manifest["executionApi"] = map[string]interface{}{"access": "MYSELF"}
						modified = true
					}

					if modified {
						if newContent, err := json.MarshalIndent(manifest, "", "  "); err == nil {
							content = newContent
						}
					}
				}
				found := false
				for i, v := range e.Project.Files {
					if v.Name == "appsscript" && strings.ToUpper(v.Type) == "JSON" {
						e.Project.Files[i].Source = string(content)
						found = true
						break
					}
				}
				if !found {
					e.Project.Files = append(e.Project.Files, File{
						Name:   "appsscript",
						Type:   "JSON",
						Source: string(content),
					})
				}
			} else if utl.ChkExtention(ext) {
				name := "ggsrun/" + strings.TrimSuffix(base, ext) + ".gs"
				filetype := utl.ExtToType(ext, true)
				if filetype == "HTML" {
					name = "ggsrun/" + strings.TrimSuffix(base, ext) + ".html"
				} else if filetype == "JSON" {
					name = "ggsrun/" + strings.TrimSuffix(base, ext) + ".json"
				}
				source := utl.ConvGasToUpload(elm)
				if filetype == "SERVER_JS" {
					var err error
					source, err = InjectSandbox(source, c.String("sandbox"))
					if err != nil {
						e.FailStatus("Sandbox Injection Error")
						pterm.Error.Println(err)
						utl.Exit(1)
					}
				}

				var exists bool
				var existingIndex int
				for i, v := range e.Project.Files {
					if v.Name == name {
						exists = true
						existingIndex = i
						break
					}
				}

				choice := c.String("conflict")
				if exists && choice == "" && !c.Bool("jsonparser") && os.Getenv("GGSRUN_MCP_MODE") != "true" {
					var err error
					choice, err = pterm.DefaultInteractiveSelect.
						WithDefaultText(fmt.Sprintf("Script '%s' already exists in remote GAS project. Action?", name)).
						WithOptions([]string{"overwrite", "add"}).
						Show()
					if err != nil {
						choice = "overwrite" // fallback
					}
				}
				if choice == "" {
					choice = "overwrite" // default
				}

				finalName := name
				if exists && choice == "overwrite" {
					e.Project.Files[existingIndex].Source = source
					e.Project.Files[existingIndex].Type = filetype
					e.Msg = append(e.Msg, fmt.Sprintf("'%s' (%s) in project was overwritten.", name, filetype))
				} else if exists && choice == "add" {
					extPart := filepath.Ext(name)
					basePart := strings.TrimSuffix(name, extPart)
					
					suffix := 1
					for {
						newName := fmt.Sprintf("%s_%d%s", basePart, suffix, extPart)
						nameExists := false
						for _, v := range e.Project.Files {
							if v.Name == newName {
								nameExists = true
								break
							}
						}
						if !nameExists {
							finalName = newName
							break
						}
						suffix++
					}

					e.Project.Files = append(e.Project.Files, File{
						Name:   finalName,
						Type:   filetype,
						Source: source,
					})
					e.Msg = append(e.Msg, fmt.Sprintf("'%s' (%s) was added to project as '%s'.", name, filetype, finalName))
				} else {
					e.Project.Files = append(e.Project.Files, File{
						Name:   name,
						Type:   filetype,
						Source: source,
					})
				}
				uploadedNames = append(uploadedNames, finalName)
			} else {
				pterm.Warning.Printf("File '%s' ignored (unsupported extension).\n", elm)
			}
		}

		e.projectUpdate2()

		if c.Bool("deleteScript") {
			e.InitVal.uploadedFilesToCleanup = uploadedNames
		}
	} else if c.Bool("backup") {
		e.UpdateStatus("Preparing project backup...")
		e.projectBackup(c)
	}

	// Robustness: we no longer need unconditional compilation sleep because manifest preservation (executionApi and webapp)
	// avoids the 404 error completely. Transient latency can still be handled by adaptive 404-retries inside execution API calls.

	return e
}

// exe2Function : Sends GAS script to server library without updating the project.
func (e *ExecutionContainer) exe2Function(c *cli.Context) *ExecutionContainer {
	e.UpdateStatus("Resolving local script and evaluating IIFE bindings...")
	var rawScript string
	isConvert := c.Bool("convert")

	// 1. Resolve raw script contents
	if c.Bool("foldertree") {
		rawScript = "function main(){return new ggsrun(null, null, null).foldertree();}"
	} else if isConvert {
		if len(c.String("value")) == 0 {
			e.FailStatus("Validation Error")
			pterm.Error.Println("No File ID. Please set it using '-v [ File ID ]'.")
			utl.Exit(1)
		}
		rawScript = "function main(e){return new ggsrun(e, null, null).nodocsdownloader();}"
	} else if len(c.String("stringscript")) > 0 {
		rawScript = c.String("stringscript")
	} else if isStdinPiped() {
		var err error
		rawScript, err = readAllStdin()
		if err != nil {
			e.FailStatus("Stdin Read Error")
			pterm.Error.Println(err)
			utl.Exit(1)
		}
	} else {
		scriptFile := c.String("scriptfile")
		if scriptFile == "" {
			e.FailStatus("Validation Error")
			pterm.Error.Println("No script. Please set GAS script using '-s' or '--stringscript'.")
			utl.Exit(1)
		}
		rawScript = utl.ConvGasToPut(c)
	}

	// 2. Resolve target API server function (GAS Endpoint)
	serverFuncName := c.String("function")
	if serverFuncName == "" {
		serverFuncName = deffuncserv // default is "ExecutionApi"
	}

	// 3. Resolve and format arguments (value) for local main() function
	val := c.String("value")
	var argStr string
	if val != "" {
		if numRe.MatchString(val) || arrayRe.MatchString(val) || objRe.MatchString(val) || val == "true" || val == "false" || val == "null" {
			argStr = val
		} else {
			argStr = fmt.Sprintf("%q", val)
		}
	}

	// 4. Wrap with IIFE ensuring main() is the universal entry point
	wrappedScript := fmt.Sprintf(`(function() {
%s
if (typeof main !== 'undefined') {
	return main(%s);
}
return "Execution completed, but main() was not defined in the local script.";
})()`, rawScript, argStr)

	// 5. CRITICAL FIX for GAS eval(eval(rec.com)) architecture.
	// Encode the IIFE as a JSON string literal.
	quotedBytes, _ := json.Marshal(wrappedScript)
	quotedScript := string(quotedBytes)

	// 6. Execute API sequence
	e = e.executionAPIwithServer(quotedScript, serverFuncName).esenderForExe2(c)
	if isConvert {
		e = e.byteSliceConverter()
	}
	return e
}

// ExecutionAPIwithoutServer : For exe1
func (e *ExecutionContainer) executionAPIwithoutServer(c *cli.Context) *ExecutionContainer {
	if len(e.Param.Function) == 0 {
		e.Param.Function = deffuncwithout
		e.Msg = append(e.Msg, fmt.Sprintf("Executed default function '%s()'.", deffuncwithout))
	}
	e.Param.DevMode = true
	return e
}

// executionAPIwithServer : Binds parameters securely for exe2
func (e *ExecutionContainer) executionAPIwithServer(sendscript string, serverFuncName string) *ExecutionContainer {
	e.UpdateStatus("Formulating Execution API payload...")
	if len(sendscript) == 0 {
		e.FailStatus("Validation Error")
		pterm.Error.Println("No script payload provided.")
		utl.Exit(1)
	}

	// Set the API target endpoint to the specified GAS wrapper
	e.Param.Function = serverFuncName

	scr := &Com{
		Com:     sendscript,
		Exefunc: "main", // Informational logging on server
		Log:     e.InitVal.log,
	}

	scri, _ := json.Marshal(scr)
	e.Param.Parameters = []string{string(scri)}
	e.Param.DevMode = true
	return e
}

// executionError : Check error for execution API
func (e *ExecutionContainer) executionError(body []byte, err error) {
	if err != nil {
		e.FailStatus("API Execution Failed")
		json.Unmarshal(body, &e.FeedBackData)
		if e.FeedBackData.Error.Status == "UNAUTHENTICATED" {
			if len(e.chkAtoken().Error) > 0 {
				pterm.Error.Printf("Invalid Access token. Please retrieve it again using command '%s auth'.\nCurrent access token is '%s'.\n", appname, e.GgsrunCfg.Accesstoken)
				utl.Exit(1)
			}
			pterm.Error.Printf("Authorization Error: Please check SCOPEs of your GAS script and server using GAS Script Editor.\nIf the SCOPEs have changed, modify them in '%s' and delete a line of 'refresh_token', then, execute '%s' again. You can retrieve new access token with modified SCOPEs.\n", cfgFile, appname)
			utl.Exit(1)
		}
		if e.FeedBackData.Error.Message == "PERMISSION_DENIED" &&
			e.FeedBackData.Error.Code == 403 {
			pterm.Error.Println("Please check Execution API at Developer console.\nIf Execution API is unable, please enable it. Or please check 'client_secret.json'. It might be that that is not for the project with Execution API.")
			utl.Exit(1)
		}
		if e.FeedBackData.Error.Message == "Requested entity was not found." &&
			e.FeedBackData.Error.Code == 404 {
			pterm.Error.Println("Please check the deployment of API executable and/or the ggsrun server.\n - If you use command 'e1', please deploy API executable again. If you use command 'e2', please check both again.\n - After deployed API executable, please save each scripts on the project again. This is very important point!\n - When you use the server as library, please confirm server.\n - Also you can use 'Logger.log(ggsrunif.Beacon())' at Google Apps Script Editor to confirm server condition.\n - Also, please check the script ID.")
			utl.Exit(1)
		}
		if len(e.FeedBackData.Error.Detailes) > 0 && e.FeedBackData.Error.Detailes[0].ErrorMessage == "The script completed but the returned value is not a supported return type." &&
			e.FeedBackData.Error.Code == 500 {
			pterm.Error.Printf("%s\n", e.FeedBackData.Error.Detailes[0].ErrorMessage)
			utl.Exit(1)
		}
		pterm.Error.Printf("%s.\n%s\n", err, body)
		utl.Exit(1)
	}
}

// EsenderForExe1 : Sends GAS to Google and retrieves results.
func (e *ExecutionContainer) esenderForExe1(c *cli.Context) *ExecutionContainer {
	e.UpdateStatus("Executing GAS function via Execution API...")
	var paraint []interface{}
	fSlice := c.StringSlice("function")
	if len(fSlice) > 1 {
		for _, argStr := range fSlice[1:] {
			var parsedVal interface{}
			if err := json.Unmarshal([]byte(argStr), &parsedVal); err == nil {
				paraint = append(paraint, parsedVal)
			} else {
				paraint = append(paraint, argStr)
			}
		}
	} else {
		valStr := c.String("value")
		if len(valStr) > 0 {
			var parsedVal interface{}
			if err := json.Unmarshal([]byte(valStr), &parsedVal); err == nil {
				paraint = []interface{}{parsedVal}
			} else {
				paraint = []interface{}{valStr}
			}
		}
	}
	epara := &e1para{
		Function:   e.Param.Function,
		Parameters: paraint,
		DevMode:    e.Param.DevMode,
	}
	re, _ := json.Marshal(epara)
	if len(re) == 0 {
		e.FailStatus("Validation Error")
		pterm.Error.Printf("Format of values is wrong. Double and single quotates have to be escaped.\n - Inputted value was  %s\n", c.String("value"))
		utl.Exit(1)
	}
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      executionurl + e.GgsrunCfg.Scriptid + ":run",
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       370,
	}
	var body []byte
	var err error
	for attempt := 1; attempt <= 4; attempt++ {
		r.Data = bytes.NewBuffer(re)
		body, err = r.FetchAPI()
		if err == nil {
			break
		}

		var tempFeed FeedBackData
		_ = json.Unmarshal(body, &tempFeed)
		if tempFeed.Error.Code == 404 && tempFeed.Error.Message == "Requested entity was not found." {
			if attempt < 4 {
				if !c.Bool("jsonparser") {
					e.UpdateStatus(fmt.Sprintf("Execution API returned 404. Retrying (attempt %d/3) in 2s for GAS compilation...", attempt))
				}
				time.Sleep(2000 * time.Millisecond)
				continue
			}
		}
		break
	}
	e.executionError(body, err)
	json.Unmarshal(body, &e.FeedBackData)
	var dat string
	if len(e.FeedBackData.Error.Message) > 0 {
		if len(e.FeedBackData.Error.Detailes) > 0 && len(e.FeedBackData.Error.Detailes[0].ScriptStackTraceElements) > 0 {
			dat = fmt.Sprintf("{code: %d, message: %s, function: %s, linenumber: %d}", e.FeedBackData.Error.Code, e.FeedBackData.Error.Message, e.FeedBackData.Error.Detailes[0].ScriptStackTraceElements[0].Function, e.FeedBackData.Error.Detailes[0].ScriptStackTraceElements[0].LineNumber)
		} else {
			dat = fmt.Sprintf("{code: %d, message: %s}", e.FeedBackData.Error.Code, e.FeedBackData.Error.Message)
		}
		e.Msg = append(e.Msg, dat)
	} else {
		var rs map[string]interface{}
		json.Unmarshal(body, &rs)
		if resp, ok := rs["response"].(map[string]interface{}); ok {
			e.FeedBackData.Response.Result.Result = resp["result"]
		}
	}
	if len(e.FeedBackData.Error.Detailes) > 0 {
		dat = fmt.Sprintf("{detailmessage: %s}", e.FeedBackData.Error.Detailes[0].ErrorMessage)
		e.Msg = append(e.Msg, dat)
	}
	e.FeedBackData.Response.Result.TotalEt = math.Trunc(time.Since(e.InitVal.pstart).Seconds()*1000) / 1000
	e.FeedBackData.Response.Result.Uapi = eapir1
	e.Msg = append(e.Msg, fmt.Sprintf("Function '%s()' was run.", e.Param.Function))
	return e
}

// esenderForExe2 : Sends GAS to Google and retrieves results with self-healing recovery.
func (e *ExecutionContainer) esenderForExe2(c *cli.Context) *ExecutionContainer {
	e.UpdateStatus("Executing script dynamically via ggsrunif server...")
	re, _ := json.Marshal(e.Param)
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      executionurl + e.GgsrunCfg.Scriptid + ":run",
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       370,
	}
	var body []byte
	var err error
	for attempt := 1; attempt <= 4; attempt++ {
		r.Data = bytes.NewBuffer(re)
		body, err = r.FetchAPI()
		if err == nil {
			break
		}

		var tempFeed FeedBackData
		_ = json.Unmarshal(body, &tempFeed)
		if tempFeed.Error.Code == 404 && tempFeed.Error.Message == "Requested entity was not found." {
			if attempt < 4 {
				if !c.Bool("jsonparser") {
					e.UpdateStatus(fmt.Sprintf("Execution API returned 404. Retrying (attempt %d/3) in 2s for GAS compilation...", attempt))
				}
				time.Sleep(2000 * time.Millisecond)
				continue
			}
		}
		break
	}

	// Pre-inspect for function not found error to trigger self-healing
	var testFeedBack FeedBackData
	json.Unmarshal(body, &testFeedBack)

	isMissingExecutionApi := strings.Contains(testFeedBack.Error.Message, "Script function not found: ExecutionApi") ||
		(len(testFeedBack.Error.Detailes) > 0 && strings.Contains(testFeedBack.Error.Detailes[0].ErrorMessage, "Script function not found: ExecutionApi"))

	if (err != nil || len(testFeedBack.Error.Message) > 0) && isMissingExecutionApi {
		e.UpdateStatus("Self-healing: Deploying 'ExecutionApi' helper function...")
		if errRecover := e.recoverMissingExecutionApi(c); errRecover == nil {
			msg := "Auto-installed helper script 'ggsrun_api_helper.gs' successfully."
			if !c.Bool("jsonparser") {
				pterm.Success.Println("==================================================")
				pterm.Success.Println(msg)
				pterm.Success.Println("==================================================")
			}
			e.Msg = append(e.Msg, msg)

			// Wait for GAS to compile and sync the new helper script
			e.UpdateStatus("Waiting 3 seconds for GAS script compilation...")
			time.Sleep(3000 * time.Millisecond)

			// Retry request after helper deployment
			body, err = r.FetchAPI()
			json.Unmarshal(body, &e.FeedBackData)
			e.FeedBackData.Response.Result.AutoInstalledHelper = true
		} else {
			e.executionError(body, err)
		}
	} else {
		e.executionError(body, err)
		json.Unmarshal(body, &e.FeedBackData)
	}

	var dat string
	if len(e.FeedBackData.Error.Message) > 0 {
		dat = fmt.Sprintf("{code: %d, message: %s}", e.FeedBackData.Error.Code, e.FeedBackData.Error.Message)
		e.Msg = append(e.Msg, dat)
	}
	if len(e.FeedBackData.Error.Detailes) > 0 {
		if strings.Contains(e.FeedBackData.Error.Detailes[0].ErrorMessage, deffuncserv) {
			dat = fmt.Sprintf("{server_error: Server for ggsrun is NOT found. Please deploy the server which is a library for GAS as 'ggsrunif'. Sctipt ID of the library is '%s'.}", serverid)
		} else {
			dat = fmt.Sprintf("{detailmessage: %s}", e.FeedBackData.Error.Detailes[0].ErrorMessage)
		}
		e.Msg = append(e.Msg, dat)
		return e
	}
	e.FeedBackData.Response.Result.TotalEt = math.Trunc(time.Since(e.InitVal.pstart).Seconds()*1000) / 1000
	e.FeedBackData.Response.Result.Uapi = eapir2
	dlfileinf, _ := json.Marshal(e.FeedBackData.Response.Result.Result)
	var rs map[string]interface{}
	if err := json.Unmarshal(dlfileinf, &rs); err == nil {
		fid, ok := rs["fileid"].(string)
		if ok {
			e.DlFileByScript.Fileid = fid
		}
		exn, ok := rs["extension"].(string)
		if ok {
			e.DlFileByScript.Extension = exn
		}
		if len(fid) > 0 && len(exn) > 0 {
			delete(rs, "fileid")
			delete(rs, "extension")
			e.FeedBackData.Response.Result.Result = rs
			res := e.defDownloadByScriptContainer().
				GetFileinf().
				Downloader(c)
			e.Msg = append(e.Msg, res.Msgar...)
		}
	}
	e.Msg = append(e.Msg, fmt.Sprintf("Target API '%s()' was evaluated and executed successfully.", e.Param.Function))
	return e
}

// projectUpdateIni : Initialize for updating project
func (e *ExecutionContainer) projectUpdateIni(sendscript string) *ExecutionContainer {
	var overwrite bool
	for i := range e.Project.Files {
		if e.Project.Files[i].Name == defprojectname {
			e.Project.Files[i].Source = sendscript
			overwrite = true
		}
	}
	if !overwrite {
		filedata := &File{
			Name:   defprojectname,
			Type:   "SERVER_JS",
			Source: sendscript,
		}
		e.Project.Files = append(e.Project.Files, *filedata)
	}
	return e
}

// projectUpdate2 : In this method, the project is updated using Apps Script API.
func (e *ExecutionContainer) projectUpdate2() *ExecutionContainer {
	e.UpdateStatus("Uploading project files via Apps Script API...")
	script, _ := json.Marshal(e.Project)
	tokenparams := url.Values{}
	tokenparams.Set("fields", "files,scriptId")
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/content")
	r := &utl.RequestParams{
		Method:      "PUT",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        bytes.NewBuffer(script),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err := r.FetchAPI()
	if err != nil {
		e.FailStatus("Project Upload Failed")
		pterm.Error.Printf("%v. ", err)
		utl.DispScopeError2(res)
		utl.Exit(1)
	}
	e.Msg = append(e.Msg, "Project was updated.")
	return e
}

// projectBackup : Download and backup project (Apps Script API v1)
func (e *ExecutionContainer) projectBackup(c *cli.Context) *ExecutionContainer {
	e.UpdateStatus("Backing up existing project...")
	tokenparams := url.Values{}
	tokenparams.Set("fields", "files,scriptId")
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/content")
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err := r.FetchAPI()
	if err != nil {
		e.FailStatus("Project Backup Failed")
		pterm.Error.Printf("%v.\n%v\n\n", err, string(res))
		pterm.Warning.Println("One of reasons of error :\n Was the inputted project ID correct?")
		utl.DispScopeError2(res)
		utl.Exit(1)
	}
	json.Unmarshal(res, &e.Project)
	if e.Project != nil && len(e.Project.Files) > 0 {
		e.InitVal.originalFiles = make([]File, len(e.Project.Files))
		copy(e.InitVal.originalFiles, e.Project.Files)
	}
	if c.Bool("backup") {
		btok, _ := json.MarshalIndent(e.Project, "", "\t")
		filename := e.InitVal.pstart.Format("20060102_150405") + ".gs"
		os.WriteFile(filepath.Join(e.InitVal.workdir, filename), btok, 0777)
		dat := fmt.Sprintf("Project was saved as '%s'.", filename)
		e.Msg = append(e.Msg, dat)
	}
	return e
}

// webAppswithServerForExe3 : Sends GAS to Google securely, traversing OAuth redirects.
func (e *ExecutionContainer) webAppswithServerForExe3(script string, c *cli.Context) *ExecutionContainer {
	e.UpdateStatus("Transmitting payload to Web Apps endpoint...")
	targetURL := c.String("url")
	if targetURL == "" {
		targetURL = e.GgsrunCfg.WebappsUrl
	}
	if len(targetURL) == 0 {
		e.FailStatus("Validation Error")
		pterm.Error.Println("No URL for Web Apps. Please supply it via option '-u [Web Apps URL]' or configure it in ggsrun.cfg.")
		utl.Exit(1)
	}

	tokenparams := url.Values{}
	tokenparams.Set("com", script)
	tokenparams.Set("pass", c.String("password"))
	tokenparams.Set("log", strconv.FormatBool(c.Bool("log")))

	var accessToken string
	var authStatusMsg string
	var authUsed bool

	// Determine if valid OAuth context exists to bypass "Anyone" restrictions.
	if e.GgsrunCfg != nil && e.GgsrunCfg.Accesstoken != "" {
		hasScope := false
		for _, s := range e.GgsrunCfg.Scopes {
			if strings.Contains(s, "auth/drive") || strings.Contains(s, "auth/drive.readonly") {
				hasScope = true
				break
			}
		}
		if hasScope {
			accessToken = e.GgsrunCfg.Accesstoken
			authUsed = true
			authStatusMsg = "Access Token with Drive scope was utilized for secure Web Apps execution."
		} else {
			authStatusMsg = "Access Token found but lacks 'drive' or 'drive.readonly' scope. Proceeding anonymously."
		}
	} else {
		authStatusMsg = "No Access Token found (ggsrun auth not executed). Proceeding anonymously."
	}

	// Custom HTTP Client is mandatory here.
	// Standard Go clients strip the "Authorization" header upon encountering a 302 Redirect.
	// Google Web Apps explicitly rely on 302 Redirects to execute scripts securely.
	client := &http.Client{
		Timeout: 370 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			if accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			}
			return nil
		},
	}

	req, err := http.NewRequest("POST", targetURL, strings.NewReader(tokenparams.Encode()))
	if err != nil {
		e.FailStatus("Network Initialization Failed")
		pterm.Error.Printf("Failed to create request: %v\n", err)
		utl.Exit(1)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		e.FailStatus("Network Transport Failed")
		pterm.Error.Println("Please check Web Apps Service and/or URL of it. Web Apps Service might not be deployed.")
		utl.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.FailStatus("I/O Error")
		pterm.Error.Printf("Failed to read response: %v\n", err)
		utl.Exit(1)
	}

	// Catch the scenario where a user sets it to "Only myself" but has no valid token,
	// returning the Google HTML login page instead of JSON.
	if resp.StatusCode != http.StatusOK {
		e.FailStatus("Authentication Boundary Hit")
		pterm.Error.Printf("Web Apps returned Status Code %d.\nIf you set 'Who has access' to 'Only myself', ensure you have executed 'ggsrun auth' and have Drive scopes.\nResponse: %s\n", resp.StatusCode, string(body))
		utl.Exit(1)
	}

	err = json.Unmarshal(body, &e.FeedBackData.Response.Result)
	if err != nil {
		e.FailStatus("Format Parsing Error")
		pterm.Error.Printf("Failed to parse Web Apps response. (Are you hitting a login wall? Ensure your OAuth scopes are correct or set access to 'Anyone').\nError: %v\n", err)
		utl.Exit(1)
	}

	e.FeedBackData.Response.Result.TotalEt = math.Trunc(time.Since(e.InitVal.pstart).Seconds()*1000) / 1000
	e.FeedBackData.Response.Result.Uapi = wapps

	// Inject security audit info into the JSON payload
	e.FeedBackData.Response.Result.TokenAuthUsed = authUsed
	e.FeedBackData.Response.Result.TokenAuthMsg = authStatusMsg
	e.Msg = append(e.Msg, authStatusMsg)

	dlfileinf, _ := json.Marshal(e.FeedBackData.Response.Result.Result)
	var rs map[string]interface{}
	if err := json.Unmarshal(dlfileinf, &rs); err == nil {
		e.DlFileByScript.Fileid, _ = rs["fileid"].(string)
		e.DlFileByScript.Extension, _ = rs["extension"].(string)
		if len(e.DlFileByScript.Fileid) > 0 && len(e.DlFileByScript.Extension) > 0 {
			delete(rs, "fileid")
			delete(rs, "extension")
			e.FeedBackData.Response.Result.Result = rs
			e.Msg = append(e.Msg, "This mode cannot download files. Because this mode is not authorization.")
		}
	}
	return e
}

// byteSliceConverter : Process conversion payload
func (e *ExecutionContainer) byteSliceConverter() *ExecutionContainer {
	if !strings.Contains(fmt.Sprintf("%s", e.FeedBackData.Response.Result.Result), "Error") {
		var f ByteSliceFile
		rr, _ := json.Marshal(e.FeedBackData.Response.Result.Result)
		json.Unmarshal(rr, &f)
		c := make([]uint8, len(f.FileData))
		for n := range f.FileData {
			c[n] = uint8(f.FileData[n])
		}
		os.WriteFile(f.Name, c, 0777)
		e.FeedBackData.Response.Result.Result = "### Byte Slice of File ###"
		e.Msg = append(e.Msg, fmt.Sprintf("File was downloaded as '%s'. MimeType is '%s'.", f.Name, f.MimeType))
	} else {
		e.FeedBackData.Response.Result.Result = "Server isn't installed or Wrong File ID."
	}
	return e
}

// autoValidateAndDeployManifest checks the GAS project manifest (appsscript.json) and injects missing definitions, re-deploying if updated.
func (e *ExecutionContainer) autoValidateAndDeployManifest(c *cli.Context, mode string) error {
	scriptID := e.GgsrunCfg.Scriptid
	if scriptID == "" {
		return nil
	}
	if e.GgsrunCfg.Accesstoken == "" {
		return nil
	}

	e.UpdateStatus("Validating appsscript.json manifest...")

	tokenparams := url.Values{}
	tokenparams.Set("fields", "files,scriptId")
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, scriptID+"/content")
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err := r.FetchAPI()
	if err != nil {
		pterm.Warning.Printf("Could not load project manifest: %v\n", err)
		return nil
	}

	var proj Project
	if err := json.Unmarshal(res, &proj); err != nil {
		return fmt.Errorf("failed to parse project content: %w", err)
	}

	manifestIdx := -1
	for idx, f := range proj.Files {
		if f.Name == "appsscript" && strings.ToUpper(f.Type) == "JSON" {
			manifestIdx = idx
			break
		}
	}

	var manifest map[string]interface{}
	if manifestIdx != -1 {
		if err := json.Unmarshal([]byte(proj.Files[manifestIdx].Source), &manifest); err != nil {
			manifest = make(map[string]interface{})
		}
	} else {
		manifest = make(map[string]interface{})
	}

	modified := false

	// 8.1. e1 and e2 executionApi check
	if mode == "e1" || mode == "e2" {
		if _, ok := manifest["executionApi"]; !ok {
			manifest["executionApi"] = map[string]interface{}{"access": "MYSELF"}
			modified = true
		}
	}

	// 8.2. e2 and w dependencies check
	if mode == "e2" || mode == "w" {
		deps, ok := manifest["dependencies"].(map[string]interface{})
		if !ok {
			deps = make(map[string]interface{})
			manifest["dependencies"] = deps
		}
		libs, ok := deps["libraries"].([]interface{})
		if !ok {
			libs = []interface{}{}
		}
		foundLib := false
		for _, lib := range libs {
			if libMap, ok := lib.(map[string]interface{}); ok {
				if libMap["userSymbol"] == "ggsrunif" {
					foundLib = true
					break
				}
			}
		}
		if !foundLib {
			newLib := map[string]interface{}{
				"userSymbol":      "ggsrunif",
				"libraryId":       "115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov",
				"version":         "0",
				"developmentMode": true,
			}
			libs = append(libs, newLib)
			deps["libraries"] = libs
			modified = true
		}
	}

	// 8.3. w webapp check
	if mode == "w" {
		if _, ok := manifest["webapp"]; !ok {
			manifest["webapp"] = map[string]interface{}{
				"executeAs": "USER_DEPLOYING",
				"access":    "MYSELF",
			}
			modified = true
		}
	}

	if !modified {
		return nil
	}

	e.UpdateStatus("Updating appsscript.json with missing definitions...")

	updatedManifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	if manifestIdx != -1 {
		proj.Files[manifestIdx].Source = string(updatedManifestBytes)
	} else {
		newFile := File{
			Name:   "appsscript",
			Type:   "JSON",
			Source: string(updatedManifestBytes),
		}
		proj.Files = append(proj.Files, newFile)
	}

	updatedProjBytes, _ := json.Marshal(proj)
	u, _ = url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, scriptID+"/content")
	r = &utl.RequestParams{
		Method:      "PUT",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(updatedProjBytes),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err = r.FetchAPI()
	if err != nil {
		return fmt.Errorf("failed to update project content: %w", err)
	}

	e.UpdateStatus("Creating new project version...")

	type versionReq struct {
		Description string `json:"description"`
	}
	vReq := versionReq{Description: "ggsrun auto-update version"}
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
		return fmt.Errorf("failed to create project version: %w", err)
	}

	var vRes struct {
		VersionNumber int `json:"versionNumber"`
	}
	if err := json.Unmarshal(res, &vRes); err != nil {
		return fmt.Errorf("failed to parse version response: %w", err)
	}

	e.UpdateStatus(fmt.Sprintf("Deploying project (Version %d)...", vRes.VersionNumber))

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
		return fmt.Errorf("failed to list deployments: %w", err)
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
	if err := json.Unmarshal(res, &deplList); err != nil {
		return fmt.Errorf("failed to parse deployments list: %w", err)
	}

	var targetDeploymentID string
	matchType := "API_EXECUTABLE"
	if mode == "w" {
		matchType = "WEB_APP"
	}

	for _, d := range deplList.Deployments {
		for _, ep := range d.EntryPoints {
			if ep.EntryPointType == matchType {
				targetDeploymentID = d.DeploymentID
				break
			}
		}
		if targetDeploymentID != "" {
			break
		}
	}

	if targetDeploymentID != "" {
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
		reqBody.DeploymentConfig.Description = "ggsrun auto-update deployment"
		reqBodyBytes, _ := json.Marshal(reqBody)

		u, _ = url.Parse(appsscriptapi)
		u.Path = path.Join(u.Path, scriptID+"/deployments/"+targetDeploymentID)
		r = &utl.RequestParams{
			Method:      "PUT",
			APIURL:      u.String(),
			Data:        bytes.NewBuffer(reqBodyBytes),
			Accesstoken: e.GgsrunCfg.Accesstoken,
			Dtime:       30,
		}
		_, err = r.FetchAPI()
		if err != nil {
			return fmt.Errorf("failed to update deployment: %w", err)
		}
		e.Msg = append(e.Msg, fmt.Sprintf("Updated deployment %s to version %d", targetDeploymentID, vRes.VersionNumber))
	} else {
		type createDeplReq struct {
			VersionNumber    int    `json:"versionNumber"`
			ManifestFileName string `json:"manifestFileName"`
			Description      string `json:"description"`
		}
		var reqBody createDeplReq
		reqBody.VersionNumber = vRes.VersionNumber
		reqBody.ManifestFileName = "appsscript"
		reqBody.Description = "ggsrun auto-created deployment"
		reqBodyBytes, _ := json.Marshal(reqBody)

		u, _ = url.Parse(appsscriptapi)
		u.Path = path.Join(u.Path, scriptID+"/deployments")
		r = &utl.RequestParams{
			Method:      "POST",
			APIURL:      u.String(),
			Data:        bytes.NewBuffer(reqBodyBytes),
			Accesstoken: e.GgsrunCfg.Accesstoken,
			Dtime:       30,
		}
		_, err = r.FetchAPI()
		if err != nil {
			return fmt.Errorf("failed to create deployment: %w", err)
		}
		e.Msg = append(e.Msg, fmt.Sprintf("Created new deployment for version %d", vRes.VersionNumber))
	}

	return nil
}

// recoverMissingExecutionApi adds 'ggsrun_api_helper.gs' to restore ExecutionApi calls when missing.
func (e *ExecutionContainer) recoverMissingExecutionApi(c *cli.Context) error {
	tokenparams := url.Values{}
	tokenparams.Set("fields", "files,scriptId")
	u, _ := url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/content")
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      u.String() + "?" + tokenparams.Encode(),
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err := r.FetchAPI()
	if err != nil {
		return err
	}

	var proj Project
	if err := json.Unmarshal(res, &proj); err != nil {
		return err
	}

	helperSource := "const ExecutionApi = e => ggsrunif.ExecutionApi(e);"
	foundHelper := false
	for i, f := range proj.Files {
		if f.Name == "ggsrun_api_helper" && strings.ToUpper(f.Type) == "SERVER_JS" {
			proj.Files[i].Source = helperSource
			foundHelper = true
			break
		}
	}
	if !foundHelper {
		newFile := File{
			Name:   "ggsrun_api_helper",
			Type:   "SERVER_JS",
			Source: helperSource,
		}
		proj.Files = append(proj.Files, newFile)
	}

	updatedProjBytes, _ := json.Marshal(proj)
	u, _ = url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/content")
	r = &utl.RequestParams{
		Method:      "PUT",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(updatedProjBytes),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	_, err = r.FetchAPI()
	if err != nil {
		return err
	}

	type versionReq struct {
		Description string `json:"description"`
	}
	vReq := versionReq{Description: "ggsrun auto-recovery helper"}
	vReqBytes, _ := json.Marshal(vReq)
	u, _ = url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/versions")
	r = &utl.RequestParams{
		Method:      "POST",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(vReqBytes),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	res, err = r.FetchAPI()
	if err != nil {
		return err
	}

	var vRes struct {
		VersionNumber int `json:"versionNumber"`
	}
	if err := json.Unmarshal(res, &vRes); err != nil {
		return err
	}

	u, _ = url.Parse(appsscriptapi)
	u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/deployments")
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
		return err
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

	var targetDeploymentID string
	for _, d := range deplList.Deployments {
		for _, ep := range d.EntryPoints {
			if ep.EntryPointType == "API_EXECUTABLE" {
				targetDeploymentID = d.DeploymentID
				break
			}
		}
		if targetDeploymentID != "" {
			break
		}
	}

	if targetDeploymentID != "" {
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
		reqBody.DeploymentConfig.Description = "ggsrun auto-recovery deployment"
		reqBodyBytes, _ := json.Marshal(reqBody)

		u, _ = url.Parse(appsscriptapi)
		u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/deployments/"+targetDeploymentID)
		r = &utl.RequestParams{
			Method:      "PUT",
			APIURL:      u.String(),
			Data:        bytes.NewBuffer(reqBodyBytes),
			Accesstoken: e.GgsrunCfg.Accesstoken,
			Dtime:       30,
		}
		_, err = r.FetchAPI()
		return err
	} else {
		type createDeplReq struct {
			VersionNumber    int    `json:"versionNumber"`
			ManifestFileName string `json:"manifestFileName"`
			Description      string `json:"description"`
		}
		var reqBody createDeplReq
		reqBody.VersionNumber = vRes.VersionNumber
		reqBody.ManifestFileName = "appsscript"
		reqBody.Description = "ggsrun auto-created deployment"
		reqBodyBytes, _ := json.Marshal(reqBody)

		u, _ = url.Parse(appsscriptapi)
		u.Path = path.Join(u.Path, e.GgsrunCfg.Scriptid+"/deployments")
		r = &utl.RequestParams{
			Method:      "POST",
			APIURL:      u.String(),
			Data:        bytes.NewBuffer(reqBodyBytes),
			Accesstoken: e.GgsrunCfg.Accesstoken,
			Dtime:       30,
		}
		_, err = r.FetchAPI()
		return err
	}
}
