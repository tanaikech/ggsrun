/*
Package main (doc.go) :
This is a modern, highly-concurrent CLI tool to execute Google Apps Script (GAS) on a terminal and manage Google Drive infrastructure.

# Architecture Overhaul (Go 1.26+)

The core engine of `ggsrun` has been completely rewritten to embrace Go's native concurrency capabilities.

1. Massively Parallel I/O: By leveraging `golang.org/x/sync/errgroup` and real-time proxy progress bars (`github.com/vbauerster/mpb/v8`), the tool now streams Drive file uploads and downloads concurrently. It entirely bypasses legacy in-memory processing, drastically reducing footprint and network latency.
2. Ultra-Fast JSON Parsing: Integrated `github.com/goccy/go-json` (SIMD-accelerated JSON parser) replacing standard `encoding/json` to maximize CPU throughput when handling heavy API payloads.
3. Modern TUI & UX: Implemented `github.com/pterm/pterm` for highly structured, color-coded, and professional terminal outputs (tables, interactive spinners, warnings).
4. Secure OAuth Lifecycle: The outdated OOB manual flow has been excised. The application now provisions a localized loopback listener, opening the browser via `os/exec` and securely intercepting tokens with an automated HTML UI acknowledgment.
5. Autonomous Readiness: New `mcp` command prepares the tool for MCP (Model Context Protocol). It acts as a pure I/O backend for LLM clients (like Gemini CLI or Cursor) over stdio JSON-RPC, executing tool calls directly on your infrastructure without requiring any LLM API keys locally.

# Features of "ggsrun" are as follows.

1. Develops GAS using your terminal and text editor.
2. Executes GAS by giving values to your script.
3. Downloads files concurrently from Google Drive with stunning progress visualizations.
4. Uploads files concurrently to Google Drive via native Resumable upload wrappers.
5. Downloads standalone script and bound script.
6. Downloads all files and folders in a specific folder.
7. Upload script files and create projects as standalone scripts and container-bound scripts.
8. Manage Permissions of files.
9. Seach files in Google Drive using search query and regex.
10. ggsrun supports both robust OAuth2 looping and Service Accounts natively.

# How to Execute Google Apps Script Using ggsrun

If you do not have an active OAuth session, run the automated auth flow:
$ ggsrun auth

Execute your script with Execution API:
$ ggsrun e1 -s sample.gs

Execute a specific function:
$ ggsrun e1 -f foo

Concurrently Upload Files:
$ ggsrun upload -f "a.txt, b.txt, c.txt"

Concurrently Download Files:
$ ggsrun download -i "[FILE_ID_1], [FILE_ID_2], [FILE_ID_3]"
*/
package main
