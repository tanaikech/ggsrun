// Package main (handler_mcp.go) :
// Model Context Protocol (MCP) server core transport logic.
package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"ggsrun/internal/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// sendMCPResponse securely serializes and transmits JSON-RPC results strictly over stdout.
func sendMCPResponse(id interface{}, result interface{}) {
	res := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	b, _ := json.Marshal(res)
	fmt.Println(string(b))
}

// runMCP : MCP Node over stdio
func runMCP(c *cli.Context) error {
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).Println("🤖 ggsrun MCP Server initialized")
	pterm.Info.Println("System: Go 1.26.4 concurrency engine engaged.")
	pterm.Info.Println("Status: Listening on stdin/stdout for MCP JSON-RPC messages...")
	pterm.Warning.Println("NOTE: This server acts as a pure I/O backend for LLM clients.\nNo LLM API keys are required or used by this process.")

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req map[string]interface{}
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		method, _ := req["method"].(string)
		id := req["id"]

		switch method {
		case "initialize":
			sendMCPResponse(id, map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "ggsrun-mcp-server",
					"version": "5.3.4",
				},
			})

		case "notifications/initialized":
			// Acknowledge

		case "tools/list":
			sendMCPResponse(id, map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "searchfiles",
						"description": "Search Google Drive files using standard Google Drive API v3 query syntax. Supports optional regex filtering on filenames.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "The Google Drive API v3 query string. Syntax examples: `name = 'MyScript.gs' and trashed = false`, `mimeType = 'application/vnd.google-apps.folder'`, `fullText contains 'target'`. Consult Google Drive API documentation for detailed syntax.",
								},
								"regex": map[string]interface{}{
									"type":        "string",
									"description": "Optional regular expression pattern evaluated against filenames in search results to filter files locally (e.g., `^test_.*\\.gs$`).",
								},
							},
							"required": []string{"query"},
						},
					},
					{
						"name":        "download",
						"description": "Download files or recursive folder trees from Google Drive to the local environment using File/Folder IDs. Supports on-the-fly format conversion (export) for Google Workspace entities (Docs, Sheets, Slides, Drawings, Video, Photos) by specifying the `extension` parameter. If a folder ID is supplied, it recursively retrieves the directory structure and applies format conversion to compatible files. Mismatched conversion requests (e.g., trying to export Slides as xlsx) will print a warning and gracefully skip the file while continuing the parallel queue. Example: download a Document as pdf or md.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"fileid": map[string]interface{}{
									"type":        "string",
									"description": "The Google Drive File ID or Folder ID to download. Specify multiple IDs separated by commas for concurrent downloads (e.g., `id1,id2`).",
								},
								"filename": map[string]interface{}{
									"type":        "string",
									"description": "Optional local filename or file path to save the downloaded file as. If omitted, uses the remote filename on Google Drive.",
								},
								"destination": map[string]interface{}{
									"type":        "string",
									"description": "Optional local directory path to save downloaded files into. Defaults to the current working directory.",
								},
								"extension": map[string]interface{}{
									"type":        "string",
									"description": "Optional file extension format to convert Google Workspace files. Supported mappings: \n- Google Docs -> docx, pdf, rtf, html, txt, md (or markdown), odt, epub, zip\n- Google Sheets -> xlsx, ods, csv, tsv, pdf, zip\n- Google Slides -> pptx, pdf, odp, txt\n- Google Drawings -> svg, png, pdf, jpeg\n- Google Video -> mp4\n- Google Photos/Pix -> png, jpeg\nUnsupported conversions are skipped with a warning. Example: set 'pdf' to convert a Doc/Sheet/Slide to PDF.",
								},
								"conflict-mode": map[string]interface{}{
									"type":        "string",
									"description": "Action to perform on conflict when a file already exists locally. Values: 'OverwriteIfNewer' (overwrite local file only if remote is newer), 'Ignore' (unconditionally skip), 'Rename' (auto-rename with timestamp/number). Default is 'OverwriteIfNewer'.",
									"enum":        []string{"OverwriteIfNewer", "Ignore", "Rename"},
								},
								"rawdata": map[string]interface{}{
									"type":        "boolean",
									"description": "If true, downloads and saves the raw JSON metadata/payload of a Google Apps Script project instead of extracting individual files. Default is false.",
								},
							},
							"required": []string{"fileid"},
						},
					},
					{
						"name":        "upload",
						"description": "Upload local files or entire recursive directories to Google Drive. Automatically converts local files to Google Workspace formats by default, or explicitly via the `convertto` parameter. Folders are recursively mapped and uploaded. If a conversion fails (e.g., format mismatch or unsupported extension), it will print a warning and skip that file while keeping the remaining parallel queue active. Example: upload local csv converting to a Google Sheet.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"filename": map[string]interface{}{
									"type":        "string",
									"description": "The local file path or directory path to upload. Multiple paths can be comma-separated.",
								},
								"parentfolderid": map[string]interface{}{
									"type":        "string",
									"description": "Optional destination Google Drive Folder ID. If omitted, uploads directly to the user's My Drive root folder.",
								},
								"conflict-mode": map[string]interface{}{
									"type":        "string",
									"description": "Action to perform on conflict when a file with the same name already exists in the destination folder. Values: 'OverwriteIfNewer' (overwrite remote file only if local is newer), 'Ignore' (unconditionally skip), 'Rename' (auto-rename with timestamp/number). Default is 'OverwriteIfNewer'.",
									"enum":        []string{"OverwriteIfNewer", "Ignore", "Rename"},
								},
								"convertto": map[string]interface{}{
									"type":        "string",
									"description": "Optional target Google Workspace format. Supported values: 'doc' (converts to Google Docs), 'sheet' (converts to Google Sheets), 'slide' (converts to Google Slides). If omitted, ggsrun automatically maps extensions (e.g., .docx/.rtf/.html/.txt/.md/.png/.jpeg -> Google Docs; .xlsx/.xls/.csv/.tsv -> Google Sheets; .pptx/.ppt/.odp -> Google Slides; .mp4/.ogg/.mov/.webm -> Google Video). Files that cannot be converted are skipped with a warning.",
								},
								"noconvert": map[string]interface{}{
									"type":        "boolean",
									"description": "If true, bypasses automatic Google Apps format conversion and uploads files in their raw binary format (e.g., uploading .xlsx as a raw binary Excel file instead of a Google Sheet). Default is false.",
								},
								"projectname": map[string]interface{}{
									"type":        "string",
									"description": "Optional Apps Script project name when uploading local script files to create a new remote project.",
								},
							},
							"required": []string{"filename"},
						},
					},
					{
						"name":        "exe1",
						"description": "Upload/synchronize a local Google Apps Script file, a local directory, or raw script string to a remote Google Apps Script project, and execute a specified entry function with optional arguments in a single step. Returns the response payload as JSON.\n\nCRITICAL TOOL SELECTION RULE FOR LLM AGENTS:\nUse this tool (and NOT `updateproject`) if the user wants to upload and execute a script or directory in a single operation. Do NOT call `updateproject` followed by another tool if the user wants execution; `exe1` handles both uploading and execution. `updateproject` is only for updating/overwriting project files without executing any functions.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"scriptid": map[string]interface{}{
									"type":        "string",
									"description": "The unique Script ID of the Apps Script project on Google Drive. Priority:\n1. If a script ID is provided in the prompt/input, use it.\n2. If not provided in the prompt/input, but a script ID is defined in the configuration file 'ggsrun.cfg', do NOT pass this parameter (it will default to the configured script ID).\n3. If it is in neither, you MUST ask the user to provide the script ID before running the tool.",
								},
								"scriptfile": map[string]interface{}{
									"type":        "string",
									"description": "Path to a local Google Apps Script source file (.gs, .js, .txt, etc.) or local directory containing multiple scripts to upload/synchronize before execution. Either `scriptfile` or `stringscript` is required.",
								},
								"stringscript": map[string]interface{}{
									"type":        "string",
									"description": "Raw Google Apps Script code provided directly as an inline string to upload/synchronize before execution. Either `scriptfile` or `stringscript` is required.",
								},
								"sandbox": map[string]interface{}{
									"type":        "string",
									"description": "Optional path to a configuration JSON file to control API and URL sandboxing.",
								},
								"function": map[string]interface{}{
									"oneOf": []map[string]interface{}{
										{
											"type":        "string",
											"description": "The name of the entry function to execute in the remote script project (e.g., `myFunction`).",
										},
										{
											"type": "array",
											"items": map[string]interface{}{
												"type": "string",
											},
											"description": "The entry function name followed by arguments (e.g., [\"myFunction\", \"arg1\", \"arg2\"]). First is function name, subsequent are arguments.",
										},
									},
								},
								"value": map[string]interface{}{
									"type":        "string",
									"description": "Optional argument/value to pass into the executed function as a parameter. (Fallback option if function array/slice is not used)",
								},
								"deleteScript": map[string]interface{}{
									"type":        "boolean",
									"description": "If set to true, files uploaded via this specific execution will be automatically deleted from the remote GAS project after execution completes. (Strictly for exe1 only)",
								},
								"conflict": map[string]interface{}{
									"type":        "string",
									"description": "Conflict resolution strategy when duplicate script name exists: 'overwrite' (default) or 'add' (adds as a new file with unique name suffix like _1).",
									"enum":        []string{"overwrite", "add"},
								},
								"confirm": map[string]interface{}{
									"type":        "boolean",
									"description": "Must be set to true to explicitly approve execution after reviewing the security analysis report.",
								},
							},
							"required": []string{"function"},
						},
					},
					{
						"name":        "filelist",
						"description": "List files or search by exact File Name or File ID on Google Drive. Outputs corresponding file details.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"searchbyname": map[string]interface{}{
									"type":        "string",
									"description": "Search for files by exact file name. Resolves to matching File IDs.",
								},
								"searchbyid": map[string]interface{}{
									"type":        "string",
									"description": "Search for a file by its unique File ID. Resolves to the file name.",
								},
							},
						},
					},
					{
						"name":        "updateproject",
						"description": "Synchronize and overwrite local source files or directories to an existing Google Apps Script (GAS) project on Google Drive. Specify the target `projectid` and local paths in `filename`.\n\nCRITICAL TOOL SELECTION RULE FOR LLM AGENTS:\nDo NOT use this tool if the user wants to execute a script function after uploading. Use `exe1` instead, which handles both uploading and execution in a single step.\n\nCRITICAL SECURITY & SAFETY RULES FOR LLM AGENTS:\n1. Since this tool unconditionally OVERWRITES files in the remote GAS project, you MUST present the list of local files (including recursively listing files if directories are specified) to the user and explicitly obtain their confirmation (Y/N) before executing this tool. Do NOT guess or automate this confirmation.\n2. Only call this tool after the user has explicitly reviewed the file list and approved the overwrite.\n3. When uploading selected local files to the GAS project, if a file with the same name already exists in the project and there are no specific instructions in the user's prompt on whether to overwrite or add it, you MUST explicitly ask the user first whether they want to overwrite the existing file or add it under a unique name, and specify that choice in the `conflict` parameter ('overwrite' or 'add').",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"projectid": map[string]interface{}{
									"type":        "string",
									"description": "The unique SCRIPT/PROJECT ID of the target Google Apps Script project on Google Drive to be updated/overwritten.",
								},
								"filename": map[string]interface{}{
									"type":        "string",
									"description": "Local source file names or directory paths, comma-separated. If a directory is specified, all files inside are recursively processed.",
								},
								"backup": map[string]interface{}{
									"type":        "boolean",
									"description": "Optional. Generate a local backup of the remote project prior to updating. Default is false.",
								},
								"deletefiles": map[string]interface{}{
									"type":        "boolean",
									"description": "Optional. Delete specified filenames from the remote project. Default is false.",
								},
								"conflict": map[string]interface{}{
									"type":        "string",
									"description": "Conflict resolution strategy when duplicate script name exists: 'overwrite' (default) or 'add' (adds as a new file with unique name suffix like _1).",
									"enum":        []string{"overwrite", "add"},
								},
							},
							"required": []string{"projectid", "filename"},
						},
					},
				},
			})

		case "tools/call":
			params, _ := req["params"].(map[string]interface{})
			name, _ := params["name"].(string)
			argsMap, _ := params["arguments"].(map[string]interface{})
			if argsMap == nil {
				argsMap = make(map[string]interface{})
			}

			if name == "exe1" {
				if _, ok := argsMap["deleteScript"]; !ok {
					argsMap["deleteScript"] = true
				}
				scriptfile, _ := argsMap["scriptfile"].(string)
				stringscript, _ := argsMap["stringscript"].(string)
				confirm, _ := argsMap["confirm"].(bool)

				// Retrieve combined script content
				code, err := getScriptContentForAnalysis(scriptfile, stringscript)
				if err != nil {
					sendMCPResponse(id, map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": fmt.Sprintf("Error retrieving script content for analysis: %v", err),
							},
						},
					})
					continue
				}

				// Perform static analysis
				report, hasWrite := analyzeGASScript(code)

				if !confirm {
					var responseText string
					if hasWrite {
						responseText = fmt.Sprintf("%s\n\n⚠️ SECURITY WARNING: Write, update, or delete operations were detected in the script. Execution has been blocked for safety.\nTo proceed, please explicitly confirm execution by calling this tool again with \"confirm\": true.", report)
					} else {
						responseText = fmt.Sprintf("%s\n\nℹ️ Security Review Complete. This script appears to be read-only.\nTo proceed with execution, please call this tool again with \"confirm\": true.", report)
					}
					sendMCPResponse(id, map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": responseText,
							},
						},
					})
					continue
				}
			}

			var cmdArgs []string
			cmdArgs = append(cmdArgs, name)
			for k, v := range argsMap {
				if v == nil || k == "confirm" {
					continue
				}
				if k == "function" {
					if slice, ok := v.([]interface{}); ok {
						for _, item := range slice {
							cmdArgs = append(cmdArgs, "-f", fmt.Sprintf("%v", item))
						}
					} else {
						cmdArgs = append(cmdArgs, "-f", fmt.Sprintf("%v", v))
					}
					continue
				}
				if boolVal, ok := v.(bool); ok {
					if boolVal {
						cmdArgs = append(cmdArgs, "--"+k)
					}
					continue
				}
				valStr := fmt.Sprintf("%v", v)
				if valStr == "" {
					continue
				}
				cmdArgs = append(cmdArgs, "--"+k, valStr)
			}
			cmdArgs = append(cmdArgs, "--jsonparser")

			exePath := os.Getenv("GGSRUN_TEST_EXE_PATH")
			if exePath == "" {
				var err error
				exePath, err = os.Executable()
				if err != nil {
					exePath = "ggsrun"
				}
			}

			cmd := exec.Command(exePath, cmdArgs...)
			cmd.Env = append(os.Environ(), "GGSRUN_MCP_MODE=true")
			var stdoutBuf, stderrBuf bytes.Buffer
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf
			err := cmd.Run()

			resultText := stdoutBuf.String()
			if err != nil {
				// Capture both error and stderr buffer to ensure the Agent reads the explicit prompt instructions
				resultText = fmt.Sprintf("Execution Error: %v\nStderr/Logs:\n%s\nStdout:\n%s", err, stderrBuf.String(), resultText)
			}

			sendMCPResponse(id, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": resultText,
					},
				},
			})
		}
	}

	if err := scanner.Err(); err != nil {
		pterm.Error.Printf("MCP Transport breakdown: %v\n", err)
	}

	return nil
}

// getScriptContentForAnalysis collects and combines script contents from file or directory, and inline strings.
func getScriptContentForAnalysis(scriptfile, stringscript string) (string, error) {
	var combined bytes.Buffer

	if stringscript != "" {
		combined.WriteString(stringscript)
		combined.WriteByte('\n')
	}

	if scriptfile != "" {
		rawFiles := regexp.MustCompile(`\s*,\s*`).Split(scriptfile, -1)
		for _, f := range rawFiles {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			fi, err := os.Stat(f)
			if err != nil {
				return "", fmt.Errorf("file/directory not found: %s", f)
			}
			if fi.IsDir() {
				err = filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						ext := filepath.Ext(path)
						if utl.ChkExtention(ext) || filepath.Base(path) == "appsscript.json" {
							content, err := os.ReadFile(path)
							if err == nil {
								combined.Write(content)
								combined.WriteByte('\n')
							}
						}
					}
					return nil
				})
				if err != nil {
					return "", err
				}
			} else {
				content, err := os.ReadFile(f)
				if err != nil {
					return "", err
				}
				combined.Write(content)
				combined.WriteByte('\n')
			}
		}
	}

	return combined.String(), nil
}

// analyzeGASScript parses the GAS code and identifies Google API resources and operations.
func analyzeGASScript(code string) (string, bool) {
	lines := strings.Split(code, "\n")

	type resourceMatch struct {
		reads  []string
		writes []string
	}
	matches := make(map[string]*resourceMatch)
	getResources := func(res string) *resourceMatch {
		if _, ok := matches[res]; !ok {
			matches[res] = &resourceMatch{}
		}
		return matches[res]
	}

	writeKeywords := regexp.MustCompile(`\.(sendEmail|createDraft|createLabel|addLabel|markMessage|star|unstar|archive|moveTo|trash|delete|remove|createFile|createFolder|addFile|addFolder|removeFile|removeFolder|setTrashed|setSharing|setDescription|setName|setContent|addEditor|addViewer|removeEditor|removeViewer|insertFile|update|setValue|setValues|appendRow|clear|clearContent|clearFormat|deleteActiveSheet|deleteRow|deleteRows|deleteColumn|deleteColumns|insertSheet|insertRow|insertRows|insertColumn|insertColumns|create|createEvent|createEventFromSeries|deleteEvent|deleteCalendar|createCalendar|setColor|setLocation|setTitle|setTime|append|insert|replace|setText|saveAndClose|fetch)\b`)

	hasWrite := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		var resource string
		if strings.Contains(trimmed, "GmailApp") || strings.Contains(trimmed, "MailApp") {
			resource = "📧 Gmail / Email Services"
		} else if strings.Contains(trimmed, "DriveApp") || strings.Contains(trimmed, "Drive") {
			resource = "📂 Google Drive"
		} else if strings.Contains(trimmed, "SpreadsheetApp") || strings.Contains(trimmed, "Spreadsheet") {
			resource = "📊 Google Spreadsheet"
		} else if strings.Contains(trimmed, "CalendarApp") {
			resource = "📅 Google Calendar"
		} else if strings.Contains(trimmed, "DocumentApp") {
			resource = "📝 Google Document (Docs)"
		} else if strings.Contains(trimmed, "SlidesApp") {
			resource = "🎨 Google Slides"
		} else if strings.Contains(trimmed, "FormApp") {
			resource = "📝 Google Forms"
		} else if strings.Contains(trimmed, "UrlFetchApp") {
			resource = "🌐 External Network Egress (UrlFetchApp)"
		} else {
			continue
		}

		resObj := getResources(resource)

		if loc := writeKeywords.FindStringIndex(trimmed); loc != nil {
			matchText := trimmed[loc[0]:loc[1]]
			exists := false
			for _, w := range resObj.writes {
				if w == matchText {
					exists = true
					break
				}
			}
			if !exists {
				resObj.writes = append(resObj.writes, matchText)
			}
			hasWrite = true
		} else {
			if strings.Contains(trimmed, ".") {
				methodRe := regexp.MustCompile(`\.([a-zA-Z0-9_]+)\(`)
				if m := methodRe.FindStringSubmatch(trimmed); len(m) > 1 {
					methodName := "." + m[1]
					exists := false
					for _, r := range resObj.reads {
						if r == methodName {
							exists = true
							break
						}
					}
					if !exists {
						resObj.reads = append(resObj.reads, methodName)
					}
				}
			}
		}
	}

	if len(matches) == 0 {
		return "• No specific Google Apps Script API resource or operation was statically identified in the source code.", false
	}

	var report bytes.Buffer
	report.WriteString("### 🛡️ GgSrun Security Guardrails - Script Safety Report\n")
	report.WriteString("We analyzed the Google Apps Script content and identified the following potential resource interactions:\n\n")

	for res, m := range matches {
		report.WriteString(fmt.Sprintf("• **%s**:\n", res))
		if len(m.reads) > 0 {
			report.WriteString("  - Read Operations:\n")
			for _, r := range m.reads {
				report.WriteString(fmt.Sprintf("    - `%s` (detected in code)\n", r))
			}
		}
		if len(m.writes) > 0 {
			report.WriteString("  - Write/Update/Delete Operations:\n")
			for _, w := range m.writes {
				report.WriteString(fmt.Sprintf("    - `%s` (detected in code)\n", w))
			}
		}
	}

	return report.String(), hasWrite
}
