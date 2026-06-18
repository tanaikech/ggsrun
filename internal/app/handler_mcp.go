// Package main (handler_mcp.go) :
// Model Context Protocol (MCP) server core transport logic.
package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"

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
	pterm.Info.Println("System: Go 1.26.3 concurrency engine engaged.")
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
					"version": "5.2.2",
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
						"description": "Upload/synchronize a local Google Apps Script file or raw script string to a remote Google Apps Script project, and execute a specified entry function. Returns the function execution response payload as JSON.\n\nSAFETY & SECURITY RULES FOR LLM AGENTS:\nBefore executing any script using this tool, you MUST inspect and statically analyze the script content (either from `scriptfile` or `stringscript`) and perform the following safety evaluation:\n1. Identify if the script performs any write, update, or delete actions on Google APIs/resources (e.g., Google Drive, Gmail, Google Calendar, Google Sheets, etc.). Specifically, look for write methods like `.create*`, `.set*`, `.update*`, `.delete*`, `MailApp.sendEmail`, `GmailApp.sendEmail`, etc.\n2. If it contains ANY write, update, or delete actions, you MUST display a brief summary of the potential security risks and resource changes, list the specific write actions detected, and ask the user for confirmation (Y/N) before calling this tool.\n3. If the script is strictly READ-ONLY (only retrieving data, e.g., `.get*`, and no modification or email sending), you may proceed to call this tool directly without asking for confirmation.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"scriptid": map[string]interface{}{
									"type":        "string",
									"description": "The unique Script ID of the Apps Script project on Google Drive. If omitted, the tool automatically attempts to read the 'script_id' from the configuration file 'ggsrun.cfg' (checked under the path in the environment variable 'GGSRUN_CFG_PATH' if defined, or the current directory). Note: If 'script_id' is not defined in the configuration, this parameter is strictly required.",
								},
								"scriptfile": map[string]interface{}{
									"type":        "string",
									"description": "Path to a local Google Apps Script source file (.gs, .js, .txt, etc.) containing the script code to upload/synchronize before execution. Either `scriptfile` or `stringscript` is required.",
								},
								"stringscript": map[string]interface{}{
									"type":        "string",
									"description": "Raw Google Apps Script code provided directly as an inline string to upload/synchronize before execution. Either `scriptfile` or `stringscript` is required.",
								},
								"function": map[string]interface{}{
									"type":        "string",
									"description": "The name of the entry function to execute in the remote script project (e.g., `myFunction`).",
								},
								"value": map[string]interface{}{
									"type":        "string",
									"description": "Optional argument/value to pass into the executed function as a parameter.",
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
						"description": "Synchronize and overwrite local source files or directories to an existing Google Apps Script (GAS) project on Google Drive. Specify the target `projectid` and local paths in `filename`. \n\nCRITICAL SECURITY & SAFETY RULES FOR LLM AGENTS:\n1. Since this tool unconditionally OVERWRITES files in the remote GAS project, you MUST present the list of local files (including recursively listing files if directories are specified) to the user and explicitly obtain their confirmation (Y/N) before executing this tool. Do NOT guess or automate this confirmation.\n2. Only call this tool after the user has explicitly reviewed the file list and approved the overwrite.",
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

			var cmdArgs []string
			cmdArgs = append(cmdArgs, name)
			for k, v := range argsMap {
				if v == nil {
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
