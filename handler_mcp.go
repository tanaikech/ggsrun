// Package main (handler_mcp.go) :
// Model Context Protocol (MCP) server core transport logic.
package main

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
						"description": "Search Google Drive files using query parameters (e.g., name='target' and trashed=false).",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{"type": "string"},
							},
							"required": []string{"query"},
						},
					},
					{
						"name":        "download",
						"description": "Download file(s) or folders from Drive by File ID. Use for retrieving content structure.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"fileid":        map[string]interface{}{"type": "string"},
								"conflict-mode": map[string]interface{}{"type": "string", "description": "Action on conflict: skip, overwrite, rename, update. DO NOT guess this value. If empty, conflicts will be returned as pending. Ask the user before using this argument.", "enum": []string{"skip", "overwrite", "rename", "update"}},
							},
							"required": []string{"fileid"},
						},
					},
					{
						"name":        "upload",
						"description": "Upload a local file or recursive folder to Google Drive.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"filename":       map[string]interface{}{"type": "string"},
								"parentfolderid": map[string]interface{}{"type": "string"},
								"conflict-mode":  map[string]interface{}{"type": "string", "description": "Action on conflict: skip, overwrite, rename, update. DO NOT guess this value. If empty, conflicts will be returned as pending. Ask the user before using this argument.", "enum": []string{"skip", "overwrite", "rename", "update"}},
							},
							"required": []string{"filename"},
						},
					},
					{
						"name":        "exe1",
						"description": "Execute a specific GAS function on a Google Apps Script project. Returns JSON payload.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"scriptid": map[string]interface{}{"type": "string"},
								"function": map[string]interface{}{"type": "string"},
								"value":    map[string]interface{}{"type": "string"},
							},
							"required": []string{"scriptid", "function"},
						},
					},
					{
						"name":        "filelist",
						"description": "List files or search by name exactly. Outputs file IDs.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"searchbyname": map[string]interface{}{"type": "string"},
							},
							"required": []string{"searchbyname"},
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
				cmdArgs = append(cmdArgs, "--"+k, fmt.Sprintf("%v", v))
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
