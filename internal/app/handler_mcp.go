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
					"version": "4.0.1",
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
						"description": "Download files or folder structures from Google Drive to the local environment using File/Folder IDs.",
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
								"conflict-mode": map[string]interface{}{
									"type":        "string",
									"description": "Action to perform on conflict when a file already exists locally. Values: 'skip' (skip downloading), 'overwrite' (overwrite local file), 'rename' (append number to local name), 'update' (overwrite only if remote is newer). Ask the user before using this argument.",
									"enum":        []string{"skip", "overwrite", "rename", "update"},
								},
							},
							"required": []string{"fileid"},
						},
					},
					{
						"name":        "upload",
						"description": "Upload local files or entire recursive directories to Google Drive.",
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
									"description": "Action to perform on conflict when a file with the same name already exists in the destination folder. Values: 'skip', 'overwrite', 'rename', 'update'. Ask the user before using this argument.",
									"enum":        []string{"skip", "overwrite", "rename", "update"},
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
						"description": "Upload/synchronize a local Google Apps Script file or raw script string to a remote Google Apps Script project, and execute a specified entry function. Returns the function execution response payload as JSON.",
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
				valStr := fmt.Sprintf("%v", v)
				if valStr == "" {
					continue
				}
				cmdArgs = append(cmdArgs, "--"+k, valStr)
			}
			cmdArgs = append(cmdArgs, "--jsonparser")

			exePath, err := os.Executable()
			if err != nil {
				exePath = "ggsrun"
			}

			cmd := exec.Command(exePath, cmdArgs...)
			var stdoutBuf, stderrBuf bytes.Buffer
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf
			err = cmd.Run()

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
