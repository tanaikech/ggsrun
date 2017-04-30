// Package main (sender.go) :
// These methods are for sending GAS scripts to Google Drive.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tanaikech/ggsrun/utl"
	"github.com/urfave/cli"
)

// Exe1Function :
func (e *ExecutionContainer) exe1Function(c *cli.Context) *ExecutionContainer {
	if len(c.String("scriptfile")) > 0 || c.Bool("backup") {
		return e.projectBackup(c).projectUpdateIni(utl.ConvGasToPut(c)).projectUpdate()
	}
	return e
}

// Exe2Function :
func (e *ExecutionContainer) exe2Function(c *cli.Context) *ExecutionContainer {
	if c.Bool("foldertree") {
		btof := "function main(){return new ggsrun(null, null, null).foldertree()}"
		return e.executionAPIwithServer(utl.ConvStringToRun(c, btof)).esenderForExe2(c)
	}
	if c.Bool("convert") && len(c.String("value")) > 0 {
		btof := "function main(e){return new ggsrun(e, null, null).nodocsdownloader()}"
		return e.executionAPIwithServer(utl.ConvStringToRun(c, btof)).esenderForExe2(c).byteSliceConverter()
	} else if c.Bool("convert") && len(c.String("value")) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No File ID. Please set it using '-v [ File ID ]'. ")
		os.Exit(1)
	}
	if len(c.String("stringscript")) > 0 {
		return e.executionAPIwithServer(utl.ConvStringToRun(c, c.String("stringscript"))).esenderForExe2(c)
	}
	return e.executionAPIwithServer(utl.ConvGasToRun(c)).esenderForExe2(c)
}

// ExecutionAPIwithoutServer :
func (e *ExecutionContainer) executionAPIwithoutServer(c *cli.Context) *ExecutionContainer {
	if len(e.Param.Function) == 0 {
		e.Param.Function = deffuncwithout
		e.Msg = append(e.Msg, fmt.Sprintf("Executed default function '%s()'.", deffuncwithout))
	}
	e.Param.DevMode = true
	return e
}

// executionAPIwithServer :
func (e *ExecutionContainer) executionAPIwithServer(sendscript string) *ExecutionContainer {
	if len(sendscript) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No script. Please set GAS script using '-s'. ")
		os.Exit(1)
	}
	if len(e.Param.Function) == 0 {
		e.Param.Function = deffuncserv
	}
	scr := &Com{
		Com:     sendscript,
		Exefunc: e.Param.Function,
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
		json.Unmarshal(body, &e.FeedBackData)
		if e.FeedBackData.Error.Status == "UNAUTHENTICATED" {
			if len(e.chkAtoken().Error) > 0 {
				fmt.Printf("Invalid Access token. Please retrieve it again using command '%s auth'.\nCurrent access token is '%s'.\n", appname, e.GgsrunCfg.Accesstoken)
				os.Exit(1)
			}
			fmt.Printf("Authorization Error: Please check SCOPEs of your GAS script and server using GAS Script Editor.\nIf the SCOPEs have changed, modify them in '%s' and delete a line of 'refresh_token', then, execute '%s' again. You can retrieve new access token with modified SCOPEs.\n", cfgFile, appname)
			os.Exit(1)
		}
		if e.FeedBackData.Error.Message == "PERMISSION_DENIED" &&
			e.FeedBackData.Error.Code == 403 {
			fmt.Printf("Error: Please check Execution API at Developer console.\nIf Execution API is unable, please enable it. Or please check 'client_secret.json'. It might be that that is not for the project with Execution API.\n")
			os.Exit(1)
		}
		if e.FeedBackData.Error.Message == "Requested entity was not found." &&
			e.FeedBackData.Error.Code == 404 {
			fmt.Printf("Error: Please check the deployment of API executable and/or the ggsrun server.\n - If you use command 'e1', please deploy API executable again. If you use command 'e2', please check both again.\n - After deployed API executable, please save each scripts on the project again. This is very important point!\n - When you use the server as library, please confirm server.\n - Also you can use 'Logger.log(ggsrunif.Beacon())' at Google Apps Script Editor to confirm server condition.\n - Also, please check the script ID.")
			os.Exit(1)
		}
		if e.FeedBackData.Error.Detailes[0].ErrorMessage == "The script completed but the returned value is not a supported return type." &&
			e.FeedBackData.Error.Code == 500 {
			fmt.Printf("Error: %s\n", e.FeedBackData.Error.Detailes[0].ErrorMessage)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %s.\n%s", err, body)
		os.Exit(1)
	}
}

// MarshalJSON : For exe1
func (e *e1para) MarshalJSON() ([]byte, error) {
	var outd string
	if len(e.Parameters) > 0 {
		if regexp.MustCompile("^[+-]?[0-9]*[\\.]?[0-9]+$").Match([]byte(e.Parameters[0].(string))) ||
			regexp.MustCompile("^\\[|\\]$").Match([]byte(e.Parameters[0].(string))) ||
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
		fmt.Fprintf(os.Stderr, "Error: Format of values is wrong. Double and single quotates have to be escaped.\n - Inputted value was  %s\n", c.String("value"))
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
		if len(e.FeedBackData.Error.Detailes[0].ScriptStackTraceElements) > 0 {
			dat = fmt.Sprintf("{code: %d, message: %s, function: %s, linenumber: %d}", e.FeedBackData.Error.Code, e.FeedBackData.Error.Message, e.FeedBackData.Error.Detailes[0].ScriptStackTraceElements[0].Function, e.FeedBackData.Error.Detailes[0].ScriptStackTraceElements[0].LineNumber)
		} else {
			dat = fmt.Sprintf("{code: %d, message: %s}", e.FeedBackData.Error.Code, e.FeedBackData.Error.Message)
		}
		e.Msg = append(e.Msg, dat)
	} else {
		var rs map[string]interface{}
		json.Unmarshal(body, &rs)
		e.FeedBackData.Response.Result.Result = rs["response"].(map[string]interface{})["result"]
	}
	if len(e.FeedBackData.Error.Detailes) > 0 {
		dat = fmt.Sprintf("{detailmessage: %s}", e.FeedBackData.Error.Detailes[0].ErrorMessage)
		e.Msg = append(e.Msg, dat)
	}
	e.FeedBackData.Response.Result.TotalEt = math.Trunc(time.Now().Sub(e.InitVal.pstart).Seconds()*1000) / 1000
	e.FeedBackData.Response.Result.Uapi = eapir1
	e.Msg = append(e.Msg, fmt.Sprintf("Function '%s()' was run.", e.Param.Function))
	return e
}

// esenderForExe2 : Sends GAS to Google and retrieves results.
func (e *ExecutionContainer) esenderForExe2(c *cli.Context) *ExecutionContainer {
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
	e.FeedBackData.Response.Result.TotalEt = math.Trunc(time.Now().Sub(e.InitVal.pstart).Seconds()*1000) / 1000
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
	e.Msg = append(e.Msg, fmt.Sprintf("'%s()' in the script was run using ggsrun server. Server function is '%s()'.", deffuncwith, e.Param.Function))
	return e
}

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
			Type:   "server_js",
			Source: sendscript,
		}
		e.Project.Files = append(e.Project.Files, *filedata)
	}
	return e
}

// ProjectUpdate :
func (e *ExecutionContainer) projectUpdate() *ExecutionContainer {
	script, _ := json.Marshal(e.Project)
	metadata, _ := json.Marshal(&ProjectUpdaterMeta{MimeType: "application/vnd.google-apps.script"})
	tokenparams := url.Values{}
	tokenparams.Set("fields", "id,mimeType,name,parents")
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part := make(textproto.MIMEHeader)
	part.Set("Content-Type", "application/json")
	data, err := w.CreatePart(part)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	if _, err = io.Copy(data, bytes.NewReader(metadata)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	data, err = w.CreatePart(part)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	if _, err = io.Copy(data, bytes.NewReader(script)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	w.Close()
	r := &utl.RequestParams{
		Method:      "PATCH",
		APIURL:      uploadurl + e.GgsrunCfg.Scriptid + "?uploadType=multipart&" + tokenparams.Encode(),
		Data:        &b,
		Contenttype: w.FormDataContentType(),
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       10,
	}
	res, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Project cannot be updated. So the script cannot be executed. May your project be not a stand alone script? 'e1' command cannot be used to the stand alone script. Even if your script is a bound script, you can download a project using '-b' option.\n%v\n", err)
		os.Exit(1)
	}
	e.Msg = append(e.Msg, "Project was updated.")
	_ = res
	return e
}

// ProjectBackup :
func (e *ExecutionContainer) projectBackup(c *cli.Context) *ExecutionContainer {
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      sdownloadurl + e.GgsrunCfg.Scriptid + "&format=json",
		Data:        nil,
		Contenttype: "",
		Accesstoken: e.GgsrunCfg.Accesstoken,
		Dtime:       10,
	}
	res, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. Please check project ID. Inputted project ID is '%s'.\n", err, e.Scriptid)
		os.Exit(1)
	}
	json.Unmarshal(res, &e.Project)
	if c.Bool("backup") {
		btok, _ := json.MarshalIndent(e.Project, "", "\t")
		filename := e.InitVal.pstart.Format("20060102_150405") + ".gs"
		ioutil.WriteFile(filepath.Join(e.InitVal.workdir, filename), btok, 0777)
		dat := fmt.Sprintf("Project was saved as '%s'.", filename)
		e.Msg = append(e.Msg, dat)
	}
	return e
}

// WebAppswithServerForExe3 : Sends GAS to Google and retrieves results.
func (e *ExecutionContainer) webAppswithServerForExe3(script string, c *cli.Context) *ExecutionContainer {
	if len(c.String("url")) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No URL for Web Apps.")
		os.Exit(1)
	}
	tokenparams := url.Values{}
	tokenparams.Set("com", script)
	tokenparams.Set("pass", c.String("password"))
	tokenparams.Set("log", strconv.FormatBool(c.Bool("log")))
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      c.String("url"),
		Data:        strings.NewReader(tokenparams.Encode()),
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       370,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Please check Web Apps Service and/or URL of it. Web Apps Service might not be deployed. ")
		os.Exit(1)
	}
	json.Unmarshal(body, &e.FeedBackData.Response.Result)
	e.FeedBackData.Response.Result.TotalEt = math.Trunc(time.Now().Sub(e.InitVal.pstart).Seconds()*1000) / 1000
	e.FeedBackData.Response.Result.Uapi = wapps
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

// ByteSliceConverter :
func (e *ExecutionContainer) byteSliceConverter() *ExecutionContainer {
	if !strings.Contains(fmt.Sprintf("%s", e.FeedBackData.Response.Result.Result), "Error") {
		var f ByteSliceFile
		rr, _ := json.Marshal(e.FeedBackData.Response.Result.Result)
		json.Unmarshal(rr, &f)
		c := make([]uint8, len(f.FileData))
		for n := range f.FileData {
			c[n] = uint8(f.FileData[n])
		}
		ioutil.WriteFile(f.Name, c, 0777)
		e.FeedBackData.Response.Result.Result = "### Byte Slice of File ###"
		e.Msg = append(e.Msg, fmt.Sprintf("File was downloaded as '%s'. MimeType is '%s'.", f.Name, f.MimeType))
	} else {
		e.FeedBackData.Response.Result.Result = "Server isn't installed or Wrong File ID."
	}
	return e
}
