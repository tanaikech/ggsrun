// Package main (sender.go) :
// These methods are for sending GAS scripts to Google Drive.
package main

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

	"ggsrun/utl"

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
	if len(c.String("scriptfile")) > 0 || c.Bool("backup") {
		e.UpdateStatus("Preparing project update and backup...")
		return e.projectBackup(c).projectUpdateIni(utl.ConvGasToPut(c)).projectUpdate2()
	}
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
			os.Exit(1)
		}
		rawScript = "function main(e){return new ggsrun(e, null, null).nodocsdownloader();}"
	} else if len(c.String("stringscript")) > 0 {
		rawScript = c.String("stringscript")
	} else {
		scriptFile := c.String("scriptfile")
		if scriptFile == "" {
			e.FailStatus("Validation Error")
			pterm.Error.Println("No script. Please set GAS script using '-s' or '--stringscript'.")
			os.Exit(1)
		}
		b, err := os.ReadFile(scriptFile)
		if err != nil {
			e.FailStatus("I/O Error")
			pterm.Error.Printf("Failed to read script file: %v\n", err)
			os.Exit(1)
		}
		rawScript = string(b)
	}

	// 2. Resolve target API server function (GAS Endpoint)
	serverFuncName := c.String("function")
	if serverFuncName == "" {
		serverFuncName = deffuncserv // default is "ggsrunif.ExecutionApi"
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
	// Encode the IIFE as a JSON string literal. This forces the first eval() on GAS
	// to simply strip the quotes and return the script code as a string.
	// The second eval() then safely executes the script.
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
		os.Exit(1)
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
				os.Exit(1)
			}
			pterm.Error.Printf("Authorization Error: Please check SCOPEs of your GAS script and server using GAS Script Editor.\nIf the SCOPEs have changed, modify them in '%s' and delete a line of 'refresh_token', then, execute '%s' again. You can retrieve new access token with modified SCOPEs.\n", cfgFile, appname)
			os.Exit(1)
		}
		if e.FeedBackData.Error.Message == "PERMISSION_DENIED" &&
			e.FeedBackData.Error.Code == 403 {
			pterm.Error.Println("Please check Execution API at Developer console.\nIf Execution API is unable, please enable it. Or please check 'client_secret.json'. It might be that that is not for the project with Execution API.")
			os.Exit(1)
		}
		if e.FeedBackData.Error.Message == "Requested entity was not found." &&
			e.FeedBackData.Error.Code == 404 {
			pterm.Error.Println("Please check the deployment of API executable and/or the ggsrun server.\n - If you use command 'e1', please deploy API executable again. If you use command 'e2', please check both again.\n - After deployed API executable, please save each scripts on the project again. This is very important point!\n - When you use the server as library, please confirm server.\n - Also you can use 'Logger.log(ggsrunif.Beacon())' at Google Apps Script Editor to confirm server condition.\n - Also, please check the script ID.")
			os.Exit(1)
		}
		if len(e.FeedBackData.Error.Detailes) > 0 && e.FeedBackData.Error.Detailes[0].ErrorMessage == "The script completed but the returned value is not a supported return type." &&
			e.FeedBackData.Error.Code == 500 {
			pterm.Error.Printf("%s\n", e.FeedBackData.Error.Detailes[0].ErrorMessage)
			os.Exit(1)
		}
		pterm.Error.Printf("%s.\n%s\n", err, body)
		os.Exit(1)
	}
}

// MarshalJSON : For exe1
func (e *e1para) MarshalJSON() ([]byte, error) {
	var outd string
	if len(e.Parameters) > 0 {
		if regexp.MustCompile(`^[+-]?[0-9]*[\.]?[0-9]+$`).Match([]byte(e.Parameters[0].(string))) ||
			regexp.MustCompile(`^\[|\]$`).Match([]byte(e.Parameters[0].(string))) ||
			regexp.MustCompile("^{|}$").Match([]byte(e.Parameters[0].(string))) {
			outd = fmt.Sprintf("{\"devMode\":%t, \"parameters\":%v, \"function\":%q}", e.DevMode, e.Parameters, e.Function)
		} else if regexp.MustCompile("([a-zA-Z]|[0-9].*[a-zA-Z]|[a-zA-Z].*[0-9])").Match([]byte(e.Parameters[0].(string))) {
			outd = fmt.Sprintf("{\"devMode\":%t, \"parameters\":%q, \"function\":%q}", e.DevMode, e.Parameters, e.Function)
		}
	} else {
		outd = fmt.Sprintf("{\"devMode\":%t, \"function\":%q}", e.DevMode, e.Function)
	}
	return []byte(outd), nil
}

// EsenderForExe1 : Sends GAS to Google and retrieves results.
func (e *ExecutionContainer) esenderForExe1(c *cli.Context) *ExecutionContainer {
	e.UpdateStatus("Executing GAS function via Execution API...")
	var paraint []interface{}
	if len(c.String("value")) > 0 {
		paraint = []interface{}{c.String("value")}
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
		os.Exit(1)
	}
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      executionurl + e.GgsrunCfg.Scriptid + ":run",
		Data:        bytes.NewBuffer(re),
		Contenttype: "application/json;charset=UTF-8",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       370,
	}
	body, err := r.FetchAPI()
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

// esenderForExe2 : Sends GAS to Google and retrieves results.
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
	body, err := r.FetchAPI()
	e.executionError(body, err)
	json.Unmarshal(body, &e.FeedBackData)
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
		os.Exit(1)
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
		os.Exit(1)
	}
	json.Unmarshal(res, &e.Project)
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
	if len(c.String("url")) == 0 {
		e.FailStatus("Validation Error")
		pterm.Error.Println("No URL for Web Apps.")
		os.Exit(1)
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

	req, err := http.NewRequest("POST", c.String("url"), strings.NewReader(tokenparams.Encode()))
	if err != nil {
		e.FailStatus("Network Initialization Failed")
		pterm.Error.Printf("Failed to create request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		e.FailStatus("Network Transport Failed")
		pterm.Error.Println("Please check Web Apps Service and/or URL of it. Web Apps Service might not be deployed.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.FailStatus("I/O Error")
		pterm.Error.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Catch the scenario where a user sets it to "Only myself" but has no valid token,
	// returning the Google HTML login page instead of JSON.
	if resp.StatusCode != http.StatusOK {
		e.FailStatus("Authentication Boundary Hit")
		pterm.Error.Printf("Web Apps returned Status Code %d.\nIf you set 'Who has access' to 'Only myself', ensure you have executed 'ggsrun auth' and have Drive scopes.\nResponse: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	err = json.Unmarshal(body, &e.FeedBackData.Response.Result)
	if err != nil {
		e.FailStatus("Format Parsing Error")
		pterm.Error.Printf("Failed to parse Web Apps response. (Are you hitting a login wall? Ensure your OAuth scopes are correct or set access to 'Anyone').\nError: %v\n", err)
		os.Exit(1)
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
