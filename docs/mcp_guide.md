# ggsrun - Model Context Protocol (MCP) Server Guide

This guide describes how to configure, run, and test `ggsrun` as a native Model Context Protocol (MCP) Server, enabling AI agents to autonomously manage your Google Drive files and execute Google Apps Script code.

---

## Table of Contents
1. [Introduction to MCP](#1-introduction-to-mcp)
2. [Server Configuration](#2-server-configuration)
   - [Configuring Antigravity CLI](#configuring-antigravity-cli)
   - [Configuring Claude Desktop](#configuring-claude-desktop)
3. [Exposed Tools Specification](#3-exposed-tools-specification)
4. [Standardized Output Schema (TransferResult)](#4-standardized-output-schema-transferresult)
5. [Manual Testing via Stdin/Stdout](#5-manual-testing-via-stdinstdout)
6. [AI Prompt Engineering & Scenarios](#6-ai-prompt-engineering--scenarios)

---

## 1. Introduction to MCP

The **Model Context Protocol (MCP)** is an open standard that allows foundation models (LLMs) to securely interact with local tools, data, and systems. 

Running the command `$ ggsrun mcp` starts a persistent, stateful background daemon that communicates via standard input/output (`stdin`/`stdout`). This enables LLM clients to:
* Autonomously search, list, download, and upload files on standard and Shared Drives.
* Update remote Apps Script projects statefully and execute functions.
* Parse execution returns and handle multi-turn operations (like interactive conflict resolution) dynamically.

---

## 2. Server Configuration

The MCP server dynamically loads credentials and authorizations from your local `ggsrun.cfg` configuration file. Make sure you have completed the authentication process (`ggsrun setup` or `ggsrun auth`) before configuring the server.

### Configuring Antigravity CLI
To register the server inside the **Antigravity CLI** environment, append the configuration block to your global configuration file located at `~/.gemini/config/mcp_config.json`:

```json
{
  "mcpServers": {
    "ggsrun-drive-agent": {
      "command": "/absolute/path/to/ggsrun",
      "args": ["mcp"]
    }
  }
}
```

### Configuring Claude Desktop
To integrate the server with **Claude Desktop**, open your local configuration file (Mac: `~/Library/Application Support/Claude/claude_desktop_config.json`, Windows: `%APPDATA%\Claude\claude_desktop_config.json`) and add:

```json
{
  "mcpServers": {
    "ggsrun-drive-agent": {
      "command": "/absolute/path/to/ggsrun",
      "args": ["mcp"]
    }
  }
}
```

---

## 3. Exposed Tools Specification

The server exposes five core tools to connected LLM agents:

1. **`searchfiles`**:
   * *Arguments*: `query` (Google Drive v3 query string), `regex` (Optional filename regular expression filter).
   * *Purpose*: Locates files across standard and Shared Drives.
2. **`download`**:
   * *Arguments*: `fileid` (Comma-separated IDs), `workers` (Parallel threads), `extension` (Export format), `conflict_mode` (`skip`, `overwrite`, `rename`, `update`).
   * *Purpose*: Downloads files/folders concurrently.
3. **`upload`**:
   * *Arguments*: `filename` (Local path), `parentid` (Drive folder ID), `convertto` (`sheet`, `doc`, `slide`), `conflict_mode`, `projectname`, `gas` (Boolean).
   * *Purpose*: Uploads files and folders with optional auto-conversion. For folder uploads, setting `gas: true` uploads the folder as a single standalone GAS project (which triggers recursive file verification), while `gas: false` uploads it recursively as a normal folder on Google Drive. If `gas` is omitted when uploading a directory, the tool returns a text response asking for clarification.
4. **`exe1`**:
   * *Arguments*: `scriptfile` (Local path), `stringscript` (Inline script), `function` (Target function name/arguments), `sandbox` (Path to sandbox config).
   * *Purpose*: Synchronizes and statefully runs Apps Script functions under optional sandboxing.
5. **`filelist`**:
   * *Arguments*: `limit` (Max files), `searchbyid` (Specific metadata query).
   * *Purpose*: Fast directory listings and file information lookup.

---

## 4. Standardized Output Schema (`TransferResult`)

To enable LLM agents to programmatically parse results and orchestrate multi-turn interactions, `ggsrun` returns transfers in a standardized JSON schema called `TransferResult`.

### Example `TransferResult` JSON Payload
```json
{
  "message": ["Upload completed, but some conflicts require resolution."],
  "files": [
    {
      "name": "report_new.csv",
      "fileId": "1a2b3c4d5e6f7G8H9I0J",
      "mimeType": "text/csv",
      "url": "https://drive.google.com/file/d/1a2b3c4d5e6f7G8H9I0J/view",
      "size": 2048,
      "localPath": "/local/workspace/report_new.csv",
      "status": "uploaded"
    }
  ],
  "pendingConflicts": [
    {
      "name": "report_existing.csv",
      "mimeType": "text/csv",
      "size": 4096,
      "localPath": "/local/workspace/report_existing.csv",
      "status": "pending_conflict"
    }
  ],
  "actionRequired": "Conflicts detected. Please invoke upload again with a conflict-mode: 'skip', 'overwrite', 'rename', or 'update' for the pending files."
}
```

---

## 5. Manual Testing via Stdin/Stdout

You can manually inspect JSON-RPC request-response lifecycles by piping formatted payloads directly into the command line.

### Test 1: Server Initialization
Verifies the server initializes successfully and reports correct capability schemas.
```bash
$ echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}},"id":1}' | ggsrun mcp
```

### Test 2: List Exposed Tools
Retrieves the schemas, descriptions, and arguments of all tools.
```bash
$ echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}' | ggsrun mcp
```

### Test 3: Invoke File Search (`searchfiles`)
Test querying the Google Drive API v3 through the tool layer:
```bash
$ echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"searchfiles","arguments":{"query":"name = '\''test_script.js'\'' and trashed = false"}},"id":3}' | ggsrun mcp
```

### Test 4: Invoke Stateful Apps Script Execution (`exe1`)
Executes the `main` entry function on Apps Script using the local configuration fallback:
```bash
$ echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"exe1","arguments":{"scriptfile":"./test_sandbox.js","function":"main"}},"id":4}' | ggsrun mcp
```

---

## 6. AI Prompt Engineering & Scenarios

You can instruct your connected AI agent to perform complex operations using these scenario designs.

### Scenario A: Multi-turn Interactive Conflict Resolution
* **Prompt**: "Please upload the local files `data1.csv` and `data2.csv` to Google Drive. Do not set a conflict-mode flag initially. If there are name collisions, parse the `pendingConflicts` results, ask me how to resolve them, and re-run with my choice."
* **Expected Agent Workflow**:
  1. Agent invokes `upload` for both files without a `conflict_mode`.
  2. If `data1.csv` already exists, `ggsrun` uploads `data2.csv` successfully, but returns `data1.csv` inside `pendingConflicts` with `"status": "pending_conflict"`.
  3. The agent parses the payload and asks: *"I uploaded `data2.csv` successfully. However, `data1.csv` already exists. Would you like to skip, overwrite, rename, or update it?"*
  4. You reply: *"Please rename it."*
  5. The agent invokes `upload` specifically for `data1.csv` with `conflict_mode` set to `"rename"`.

### Scenario B: Seamless Script Execution
* **Prompt**: "Develop a local GAS script `cleanup_records.gs` that deletes empty rows in a Google Sheet. Upload the script to my Apps Script project (Script ID: `PROJECT_ID`) and execute the `cleanup` function."
* **Expected Agent Workflow**:
  1. The agent writes the Javascript logic locally to `cleanup_records.gs`.
  2. The agent calls the `exe1` tool, passing `cleanup_records.gs` as the `scriptfile` argument and `cleanup` as the `function` argument.
  3. The script executes remotely, and the agent parses and displays the returned execution results.

---

### Related Links:
- 🚀 **[Setup & Onboarding Guide](setup_guide.md)** - Enable Workspace APIs and acquire loopback tokens.
- 📖 **[Command Reference Manual](commands_reference.md)** - Detailed description of underlying CLI command tools.
- 🛡️ **[Security Sandbox Guide](sandbox_guide.md)** - Restrict and secure Google Workspace API scopes during agent actions.
- 🔄 **[Stateful Execution Lifecycle Guide](exe1_lifecycle.md)** - Step-by-step backup, sandboxing, execution, and rollback details for the `exe1` command.
- 🏡 **[Back to Home](../README.md)**
