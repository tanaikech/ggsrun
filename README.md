# ggsrun

![](help/images/fig1a.jpg)

<a name="top"></a>
[![Go Version](https://img.shields.io/badge/Go-1.26.3+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![MCP Ready](https://img.shields.io/badge/MCP-Ready-8A2BE2?style=for-the-badge)](https://modelcontextprotocol.io)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=for-the-badge)]()
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge)](LICENCE)

## Table of Contents

- [ggsrun](#ggsrun)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features of ggsrun](#features-of-ggsrun)
  - [The 5 Pillars of the v5 Architecture](#the-5-pillars-of-the-v5-architecture)
    - [A. Massively Parallel I/O \& UI](#a-massively-parallel-io--ui)
    - [B. Full Shared Drive (Omni-Drive) Support](#b-full-shared-drive-omni-drive-support)
    - [C. Intelligent GAS \& MIME Resolution](#c-intelligent-gas--mime-resolution)
    - [D. Robust Fault Tolerance \& Auto-Retry](#d-robust-fault-tolerance--auto-retry)
    - [E. MCP (Model Context Protocol) Integration](#e-mcp-model-context-protocol-integration)
  - [Installation \& Setup](#installation--setup)
    - [1. Install ggsrun](#1-install-ggsrun)
    - [2. Obtain Google Cloud Credentials](#2-obtain-google-cloud-credentials)
    - [3. Automated Authorization (OAuth2 Loopback)](#3-automated-authorization-oauth2-loopback)
    - [4. Set Up Execution Server (GAS Side)](#4-set-up-execution-server-gas-side)
      - [Step 4.1: Bind the Server Library](#step-41-bind-the-server-library)
      - [Step 4.2: Inject the Gateway Code](#step-42-inject-the-gateway-code)
      - [Step 4.3: Deploy as API Executable (For `exe1` \& `exe2`)](#step-43-deploy-as-api-executable-for-exe1--exe2)
      - [Step 4.4: Deploy as Web App (For `webapps`)](#step-44-deploy-as-web-app-for-webapps)
  - [Command Reference \& Usage](#command-reference--usage)
    - [Authentication \& MCP](#authentication--mcp)
    - [Massively Parallel Download](#massively-parallel-download)
    - [Massively Parallel Upload](#massively-parallel-upload)
  - [Model Context Protocol (MCP) Server \& LLM Integration](#model-context-protocol-mcp-server--llm-integration)
    - [MCP Server Configuration for Antigravity CLI](#mcp-server-configuration-for-antigravity-cli)
    - [1. Exposed Tools](#1-exposed-tools)
    - [2. Standardized JSON Output (TransferResult)](#2-standardized-json-output-transferresult)
    - [3. AI Agent Prompt Scenarios \& Expected Behaviors](#3-ai-agent-prompt-scenarios--expected-behaviors)
  - [Deep Dive: Executing Google Apps Script (exe1, exe2, webapps)](#deep-dive-executing-google-apps-script-exe1-exe2-webapps)
    - [Mode 1: `exe1` (Stateful Project Execution)](#mode-1-exe1-stateful-project-execution)
      - [Architecture Workflow](#architecture-workflow)
    - [Mode 2: `exe2` (Stateless Dynamic Execution)](#mode-2-exe2-stateless-dynamic-execution)
      - [Architecture Workflow](#architecture-workflow-1)
    - [Mode 3: `webapps` (Anonymous OR Secure Endpoint Execution)](#mode-3-webapps-anonymous-or-secure-endpoint-execution)
      - [Architecture Workflow](#architecture-workflow-2)
  - [Advanced Configurations](#advanced-configurations)
    - [Modifying OAuth Scopes](#modifying-oauth-scopes)
  - [Troubleshooting](#troubleshooting)
  - [Licence \& Author](#licence--author)
  - [Update History](#update-history)
    - [ggsrun](#ggsrun-1)
    - [Server](#server)

---

## Overview

**ggsrun** is an enterprise-grade CLI tool and MCP (Model Context Protocol) Server designed to relentlessly orchestrate Google Drive I/O operations and execute Google Apps Script (GAS) natively from a local terminal.

With the release of **v5.1.1**, `ggsrun` transcends its origins as a mere CLI tool. Built on Go 1.26.3+, the execution engine has been entirely rewritten from legacy serial processing into a channel-based, streaming concurrent architecture. It now serves as a high-performance, fault-tolerant I/O backend fully integrated with Omni-Drive (Shared Drives) support, advanced MIME resolution, secure redirect-following Auth logic, and a native **MCP Server Mode** allowing LLM agents to autonomously manage your cloud infrastructure.

---

## Features of ggsrun

1. Develops GAS using your terminal and text editor seamlessly.
2. Executes GAS directly by injecting values into your script dynamically.
3. Downloads files concurrently from Google Drive with stunning progress visualizations.
4. Uploads files concurrently to Google Drive via native Resumable upload wrappers.
5. Downloads standalone scripts and container-bound scripts flawlessly.
6. Recursively downloads all files and folders retaining absolute directory structures.
7. Uploads script files and creates projects as standalone scripts OR container-bound scripts.
8. Manages file and folder permissions across your entire Drive.
9. Searches files in Google Drive utilizing advanced search queries and Regex.
10. Supports both robust OAuth2 looping and Service Accounts natively.

---

## The 5 Pillars of the v5 Architecture

### A. Massively Parallel I/O & UI

Legacy pseudo-asynchronous processing has been eradicated. `ggsrun` now utilizes a channel-based worker pool built on `golang.org/x/sync/errgroup` to maximize network throughput on massive folder trees.

### B. Full Shared Drive (Omni-Drive) Support

The v5 engine forces `supportsAllDrives=true` and `includeItemsFromAllDrives=true` across all Google Drive API permutations. Enterprise users can now execute bulk batch operations targeting deeply nested structures within organizational Shared Drives.

### C. Intelligent GAS & MIME Resolution

The extraction logic dynamically categorizes Google workspace entities. Requests targeting GAS code bypass the standard Drive API and are automatically routed to the Apps Script API, securely landing as structured `.json` locally.

### D. Robust Fault Tolerance & Auto-Retry

The v5 execution phase is strictly non-blocking. HTTP 429 (Rate Limits) and 5xx (Server Errors) trigger a mathematical exponential backoff sequence per-worker, ensuring aggressive self-healing.

### E. MCP (Model Context Protocol) Integration

Running `ggsrun mcp` transforms the application into an autonomous JSON-RPC server via `stdio`. Large Language Model (LLM) agents can natively invoke internal capabilities without requiring any API keys locally.

---

## Installation & Setup

### 1. Install ggsrun

#### Using Go

Requires Go 1.26.3 or higher. Pull and compile the latest binary natively:

```bash
$ go install github.com/tanaikech/ggsrun@latest
```

#### Downloading Pre-built Binaries

Alternatively, you can download pre-built binaries directly from the [Releases page](https://github.com/tanaikech/ggsrun/releases).

The following compiled binaries are available:

* **macOS (Darwin)**
  * `ggsrun_darwin_amd64`
  * `ggsrun_darwin_arm64`
* **Linux**
  * `ggsrun_linux_386`
  * `ggsrun_linux_amd64`
  * `ggsrun_linux_arm64`
  * `ggsrun_linux_arm7`
  * `ggsrun_linux_mips`
  * `ggsrun_linux_mipsle`
* **FreeBSD**
  * `ggsrun_freebsd_amd64`
  * `ggsrun_freebsd_arm64`
* **Windows**
  * `ggsrun_windows_386.exe`
  * `ggsrun_windows_amd64.exe`
  * `ggsrun_windows_arm64.exe`

### 2. Obtain Google Cloud Credentials

1. Access the [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new Project.
3. Navigate to **APIs & Services > Library**. Enable both the **Google Drive API** and **Google Apps Script API**.
4. Configure the **OAuth consent screen** (External/Internal).
5. Navigate to **Credentials > Create Credentials > OAuth client ID**. Select **Desktop app**.
6. Download the resulting JSON file, move it to your working directory, and rigorously rename it to exactly `client_secret.json`.

### 3. Automated Authorization (OAuth2 Loopback)

With your `client_secret.json` in the current directory, execute:

```bash
$ ggsrun auth
```

`ggsrun` spins up a secure local loopback listener. Your default browser will launch, request authorization, and securely hand the token back to the CLI. A persistent `ggsrun.cfg` file is generated.

### 4. Set Up Execution Server (GAS Side)

To execute arbitrary GAS functions locally without permanent deployments (`exe2` and `webapps` modes), you must establish a gateway endpoint on Google Apps Script using the `ggsrunif` library.

#### Step 4.1: Bind the Server Library

1. Navigate to the [Google Apps Script Dashboard](https://script.google.com/) and create a **New Project**.
2. Click the `+` icon next to **Libraries**.
3. Input the Target Script ID: `115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov`.
4. Set the **Identifier** to `ggsrunif` and select the **latest version**.

#### Step 4.2: Inject the Gateway Code

Replace the default code in `Code.gs` with the following ultra-lightweight wrappers.

```javascript
const doPost = (e) => ggsrunif.WebApps(e, "pass1");
const ExecutionApi = (e) => ggsrunif.ExecutionApi(e);
```

_(Note: Change `"pass1"` to a secure custom password if you plan to execute webapps anonymously)._

#### Step 4.3: Deploy as API Executable (For `exe1` & `exe2`)

1. Click **Deploy** > **New Deployment**.
2. Choose **API Executable**.
3. Set **Who has access** strictly to **Only myself**.
4. Click **Deploy**. Copy the **Script ID** for the `-i` flag.

#### Step 4.4: Deploy as Web App (For `webapps`)

1. Click **Deploy** > **New Deployment**.
2. Choose **Web app**.
3. Set **Execute as** to **Me**.
4. Set **Who has access** to **Only myself**.
   _(Note: This highly secure setting requires the `ggsrun` CLI to be authenticated via `ggsrun auth` with Drive scopes enabled. If you need to trigger the webapp anonymously from a CI/CD pipeline without a token, set access to **Anyone** and rely on the `-p` password flag)._
5. Click **Deploy**. Copy the generated **Web app URL** for the `-u` flag.

---

## Command Reference & Usage

### Authentication & MCP

| Command           | Action                                                                                         |
| :---------------- | :--------------------------------------------------------------------------------------------- |
| `$ ggsrun auth`   | Initiates the secure OAuth2 loopback flow. Use `--port` to change the binding port.            |
| `$ ggsrun status` | Health diagnostic tool to verify the validity and expiration of your current Access Token.     |
| `$ ggsrun mcp`    | Starts the stdio-bound MCP Server. Listens for tools like `searchfiles`, `download`, `upload`. |

### Massively Parallel Download

Target IDs can belong to Standard Drives, Shared Drives, or Team Drives seamlessly. `ggsrun` natively handles the recursive mapping of folders and parallel byte-streaming.

| Command                                                 | Action                                                                                              |
| :------------------------------------------------------ | :-------------------------------------------------------------------------------------------------- |
| `$ ggsrun download -i "FILE_ID1, FILE_ID2" -w 5`        | Downloads specific files utilizing 5 parallel channel workers.                                      |
| `$ ggsrun download -i "FOLDER_ID" -w 10`                | Recursively maps and downloads an entire folder tree concurrently.                                  |
| `$ ggsrun download -i "SPREADSHEET_ID" -e xlsx`         | Directs the Drive API to transpile and export a native Google Sheet into an `.xlsx` binary.         |
| `$ ggsrun download -i "FOLDER_ID" -m "application/pdf"` | Recursively downloads a folder, but filters specifically to retrieve only PDF files.                |
| `$ ggsrun download -i "SCRIPT_ID" -z`                   | Downloads an entire GAS project, bundles all `.js`/`.html` files, and saves it as a `.zip` archive. |
| `$ ggsrun download -i "SCRIPT_ID" -r`                   | Downloads a GAS project natively as raw `.json` payload.                                            |
| `$ ggsrun download -i "FOLDER_ID" -cm update`            | Recursively downloads a folder, updating only files that are newer on Drive.                        |

### Massively Parallel Upload

Pushes local hierarchical structures to Google Drive asynchronously. Resumable chunked uploads are inherently supported for massive binaries (default chunk size: 100MB).

| Command                                                                      | Action                                                                                                              |
| :--------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------------ |
| `$ ggsrun upload -f "a.txt, b.txt" -p "FOLDER_ID"`                           | Uploads multiple individual files sequentially or concurrently.                                                     |
| `$ ggsrun upload -f "/path/to/folder" -p "FOLDER_ID" -w 5`                   | Uploads a local directory recursively, mimicking the exact file tree on Google Drive.                               |
| `$ ggsrun upload -f "script.js" --projectname "MyAPI"`                       | Uploads a local file and provisions it as a brand new Standalone GAS Project.                                       |
| `$ ggsrun upload -f "script.js" -pid "SHEET_ID" --projecttype "spreadsheet"` | Uploads a script and provisions it as a **Container-Bound Script** directly attached to the specified Google Sheet. |
| `$ ggsrun upload -f "data.csv" -c "sheet"`                                   | Uploads a CSV file and automatically commands Google Drive to convert it into a native Google Spreadsheet.          |
| `$ ggsrun upload -f "large_file.mp4" --chunksize 250`                        | Accelerates massive file transfers by increasing the Resumable Upload chunk size to 250MB.                          |
| `$ ggsrun upload -f "/path/to/folder" -p "FOLDER_ID" -cm rename`             | Uploads a directory, appending timestamps to any conflicting filenames on Drive.                                    |

### Conflict Resolution Mode

Both `download` and `upload` commands support the `--conflict-mode` (or `-cm`) flag to handle collisions when files already exist in the target destination.

If not specified, `ggsrun` will default to an **interactive CLI prompt** allowing you to dynamically select the resolution per collision.

| Conflict Mode | Behavior (Download) | Behavior (Upload) |
| :--- | :--- | :--- |
| `skip` | Skips downloading the file if it already exists locally. | Skips uploading the file if it already exists on Google Drive. |
| `overwrite` | Overwrites the local file. | Overwrites the remote file on Google Drive (triggers a `PATCH` request). |
| `rename` | Appends a timestamp (`_YYYYMMDD_HHMMSS`) to the filename locally. | Appends a timestamp (`_YYYYMMDD_HHMMSS`) to the filename on Google Drive. |
| `update` | Downloads only if the remote file is newer than the local file. | Uploads/updates only if the local file is newer than the remote file. |

> [!NOTE]
> The legacy `--overwrite` (`-o`) and `--skip` (`-s`) flags in `download` are deprecated. Please migrate to `--conflict-mode`.
> For massive concurrent uploads, metadata queries are pre-fetched in bulk to bypass Google Drive API rate limits.

---

## Model Context Protocol (MCP) Server & LLM Integration

Running `$ ggsrun mcp` transforms `ggsrun` into a native **Model Context Protocol (MCP) Server**, communicating with LLM clients (such as Claude Desktop, Cursor, or specialized AI agents) over standard I/O (`stdin`/`stdout`). 

With the release of **v5.1.1**, the MCP capabilities are enhanced to fully expose modern conflict resolution and deliver deeply structured JSON results.

### MCP Server Configuration for Antigravity CLI

To configure `ggsrun` as an MCP server inside your **Antigravity CLI** environment, specify the server details in your global config file at `~/.gemini/config/mcp_config.json`.

Add the following JSON configuration snippet, ensuring that the `command` value points to your exact local `ggsrun` executable path:

```json
{
  "mcpServers": {
    "ggsrun-drive-agent": {
      "command": "/path/to/ggsrun",
      "args": ["mcp"]
    }
  }
}
```

### 1. Exposed Tools
The MCP server exposes the following high-level tools to your AI agent:
- `searchfiles`: Search Google Drive files using queries (e.g., `name='target' and trashed=false`).
- `download`: Download files or folders by File ID. Includes a `--conflict-mode` option to handle name collisions.
- `upload`: Upload a local file or recursive folder to a Google Drive location. Includes a `--conflict-mode` option.
- `exe1`: Stateful execution of Google Apps Script projects.
- `filelist`: Exact name search for files, returning Google Drive File IDs.

### 2. Standardized JSON Output (`TransferResult`)
When executing transfer operations (uploads/downloads), `ggsrun` outputs a standardized JSON payload structure named `TransferResult`. This allows your AI agent to reliably parse the result, extract metadata, and identify multi-turn actions like conflict resolution.

**Example `TransferResult` JSON structure:**
```json
{
  "message": [
    "Upload processed successfully."
  ],
  "files": [
    {
      "name": "file_2.txt",
      "fileId": "1a2b3c4d5e6f7g8h9i0j",
      "mimeType": "text/plain",
      "url": "https://drive.google.com/file/d/1a2b3c4d5e6f7g8h9i0j/view",
      "size": 1024,
      "localPath": "/local/path/file_2.txt",
      "status": "uploaded"
    }
  ],
  "pendingConflicts": [
    {
      "name": "file_1.txt",
      "mimeType": "text/plain",
      "size": 2048,
      "localPath": "/local/path/file_1.txt",
      "status": "pending_conflict"
    }
  ],
  "actionRequired": "Conflicts detected. Please invoke upload again with a conflict-mode: 'skip', 'overwrite', 'rename', or 'update' for the pending files."
}
```

### 3. AI Agent Prompt Scenarios & Expected Behaviors

To help your AI agent interact effectively with the `ggsrun` MCP server, use the following standardized and optimized prompts.

#### Scenario A: Batch Upload with Interactive Conflict Resolution
Test how the AI coordinates partial execution and handles unexpected collisions when some files already exist in Google Drive while others do not.

* **Setup:** Ensure `file_1.txt` already exists on your Google Drive, while `file_2.txt` is a brand-new local file.
* **Agent Prompt:**
  > "Please upload `file_1.txt` and `file_2.txt` to Google Drive using the `upload` tool. Do not specify the conflict mode initially. If there are pending conflicts, ask me how to resolve them."
* **Expected Interaction Flow:**
  1. The AI invokes the `upload` tool for both files without passing the `--conflict-mode` argument.
  2. The `ggsrun` backend uploads `file_2.txt` successfully and populates it in the `files` array, but registers `file_1.txt` under `pendingConflicts` with `"status": "pending_conflict"`.
  3. The AI parses the `TransferResult` and successfully reports: *"I have uploaded `file_2.txt` (ID: ...). However, `file_1.txt` already exists. Would you like to skip, overwrite, rename, or update it?"*
  4. You reply: *"Please overwrite it."*
  5. The AI intelligently invokes the `upload` tool specifically for `file_1.txt` with `conflict-mode` set to `"overwrite"`.

#### Scenario B: Granular Metadata Extraction and Parsing
Test if the AI can retrieve full file metadata from `TransferResult` and report specific file properties precisely.

* **Agent Prompt:**
  > "Please download the file with ID `[YOUR_FILE_ID]` from Google Drive. Tell me exactly where it was saved (`localPath`) and its `size` from the result."
* **Expected Interaction Flow:**
  1. The AI invokes the `download` tool passing the target file ID.
  2. `ggsrun` performs the parallel download and returns a standardized JSON structure containing the file array.
  3. The AI successfully parses the `files` array in `TransferResult` and replies to you with clear, accurate metadata: *"The file has been saved to `[localPath]` and its size is `[size]` bytes."*

---

## Deep Dive: Executing Google Apps Script (exe1, exe2, webapps)

### Mode 1: `exe1` (Stateful Project Execution)

`exe1` relies on the Apps Script API to permanently upload (sync) your local `.js/.gs` file to the remote GAS project, and then invokes a specific function via the Execution API.

**When to use:** You want to permanently update the code on the cloud and run it. Requires an OAuth Token.

**Step-by-Step:**

1. Create a local script `sample.gs`:
   ```javascript
   function targetFunction(data) {
     return "Processed data: " + data;
   }
   ```
2. Execute the CLI:
   ```bash
   $ ggsrun exe1 -i [YOUR_SCRIPT_ID] -s sample.gs -f targetFunction -v "Hello World"
   ```

#### Architecture Workflow

```mermaid
sequenceDiagram
    autonumber
    participant CLI as ggsrun (Local PC)
    participant AAPI as Apps Script API
    participant EAPI as Execution API
    participant GAS as Remote GAS Project

    CLI->>AAPI: PUT /v1/projects/{id}/content<br>(Push local .js files)
    AAPI-->>CLI: 200 OK (Project Overwritten)
    CLI->>EAPI: POST /v1/scripts/{id}:run<br>Target: targetFunction
    EAPI->>GAS: trigger targetFunction()
    Note right of GAS: Executes utilizing the<br>permanently saved code
    GAS-->>EAPI: Return Value
    EAPI-->>CLI: Pure JSON Result
```

### Mode 2: `exe2` (Stateless Dynamic Execution)

`exe2` is the pinnacle of dynamic execution. It **does not modify or update** your remote GAS project files. Instead, it reads your local script, heavily sanitizes it into a secure JSON-encoded string, and transmits it as a payload to the `ExecutionApi` wrapper.

**When to use:** Rapid local prototyping and executing complex data-extraction algorithms on the cloud without polluting the production GAS project's codebase. Requires an OAuth Token.

**Step-by-Step:**

1. Create a local script `compute.js`. **The local entry point must be `main()`**:
   ```javascript
   function main(multiplier) {
     return multiplier * 10;
   }
   ```
2. Execute the CLI:
   ```bash
   $ ggsrun exe2 -i [YOUR_SCRIPT_ID] -f ExecutionApi -s compute.js -v 5 -j
   ```

#### Architecture Workflow

```mermaid
sequenceDiagram
    autonumber
    participant CLI as ggsrun (Local PC)
    participant API as Execution API
    participant GAS as GAS Project (ggsrunif)
    participant V8 as V8 Engine

    CLI->>CLI: Wrap local code in IIFE<br>Encode to JSON literal
    CLI->>API: POST /v1/scripts/{id}:run<br>Target: ExecutionApi
    API->>GAS: trigger ExecutionApi(payload)
    GAS->>V8: eval(script string)
    Note right of V8: Executes stateless logic<br>without saving files to Drive
    V8-->>GAS: Return Object/Value
    GAS-->>API: Response Wrapper
    API-->>CLI: Pure JSON Result
```

### Mode 3: `webapps` (Anonymous OR Secure Endpoint Execution)

`webapps` functions similarly to `exe2` (stateless dynamic evaluation) but bypasses the Google Execution API entirely. Instead, it routes the payload through a standard HTTP POST request to a deployed Google Web App URL.

**When to use:**

- **Secure Mode:** You want to execute arbitrary scripts natively on a highly-secured ("Only myself") endpoint utilizing the `drive` scope OAuth token.
- **Anonymous Mode:** You need to execute GAS scripts from a remote CI/CD pipeline **without deploying an OAuth token**. (Requires the Web App to be deployed as "Anyone" and utilizes the `-p` password flag).

**Step-by-Step:**

1. Create your local logic script `report.js` (entry point `main()`).
2. Execute the CLI:
   ```bash
   $ ggsrun webapps -u "https://script.google.com/macros/s/[WEB_APP_ID]/exec" -p pass1 -s report.js -j
   ```
   _(Note: If `ggsrun auth` has been executed locally, the CLI automatically detects the token, bypasses the `-p` requirement, and securely traverses Google's 302 redirects to execute the code. The `-j` JSON output will include `tokenAuthUsed: true`.)_

#### Architecture Workflow

```mermaid
sequenceDiagram
    autonumber
    participant CLI as ggsrun (Local PC)
    participant URL as Web App URL
    participant GAS as GAS Project (doPost)
    participant V8 as V8 Engine

    CLI->>CLI: URL-encode payload & Verify Token
    alt Has OAuth Token
        CLI->>URL: HTTP POST (Bearer Token attached)
        URL-->>CLI: 302 Redirect (Google Auth)
        CLI->>URL: Follow Redirect (Bearer Token re-attached)
    else Anonymous Mode
        CLI->>URL: HTTP POST (No Token, requires "Anyone" access)
    end

    URL->>GAS: trigger doPost(e)
    GAS->>V8: eval(script string)
    V8-->>GAS: Return Object/Value
    GAS-->>URL: ContentService (MimeType.JSON)
    URL-->>CLI: Pure JSON Result
```

---

## Advanced Configurations

### Modifying OAuth Scopes

By default, `ggsrun` requests all necessary scopes for Drive and GAS execution. If you need to inject custom scopes or trim existing ones:

1. Open the `ggsrun.cfg` file generated in your working directory.
2. Locate the `"scopes": [ ... ]` JSON array.
3. Add or remove your desired Google API scopes.
4. Save the file and simply run `$ ggsrun auth` again.
   The CLI will automatically re-read your modified configuration, launch the browser, and provision a new token with your exact custom scopes.

---

## Troubleshooting

**1. Web Apps Returns Status Code 200, but output is HTML**
If you set your Web App to "Only myself" but the CLI returns a parsing error with HTML, it means your `ggsrun` lacks the proper OAuth token. Run `ggsrun auth` to generate a token with the `drive` scope, which the CLI will automatically use to authenticate the Web App request across the Google 302 Redirects.

**2. "Requested entity was not found" or 404 Errors**
If utilizing GAS execution (`e1` / `e2`), verify the target project is currently deployed as an **API Executable** on the latest version. Un-deployed or draft states cannot be invoked externally.

**3. Headless Server Authentication**
If `ggsrun auth` detects a headless Linux environment (where it cannot spawn a local browser loopback), it elegantly degrades into manual mode. It prints the URL; copy it into an external browser, authorize, and paste the code block back into standard input.

---

## Licence & Author

**Licence:** [MIT](LICENCE)

**Author:** [Tanaike](https://tanaikech.github.io/about/)  
For architectural questions, advanced enterprise integrations, or bug disclosures, contact: tanaike@hotmail.com

---

## Update History

### ggsrun

- **v5.1.1 (May 2026) - Modular Handlers & Enhanced MCP Server Core**
  Refactored the codebase to modularize legacy single-file command handlers into dedicated, organized handler files (`handler_download.go`, `handler_upload.go`, `handler_transfer.go`, `handler_mcp.go`, `handler_execute.go`). Strengthened the MCP server core (`ggsrun mcp`) by capturing stdout and stderr execution logs for comprehensive error recovery. Embedded full support for `--conflict-mode` inside the MCP JSON-RPC schemas and standardized file transfer outputs into `TransferResult` to support interactive multi-turn collision resolution in LLM conversations. Fully updated pre-built binaries for all major architectures.
- **v5.1.0 (May 2026) - Advanced Conflict Resolution Engine**
  Introduced a robust pre-computation conflict resolution matrix for both `download` and `upload` commands via the new `--conflict-mode` (`-cm`) flag. Users can now choose from `skip`, `overwrite`, `rename` (appends timestamp `_YYYYMMDD_HHMMSS` to avoid collisions), or `update` (syncs only if the source file is newer than the target). Includes interactive fallback CLI prompts if no mode is specified. Deprecated the legacy `--overwrite` (`-o`) and `--skip` (`-s`) options in favor of `--conflict-mode`. To avoid Drive API rate limits during massive concurrent uploads, metadata query is pre-fetched in bulk.
- **v5.0.3 (May 2026) - CLI UX Overhaul & Dynamic TUI Integration**
  Introduced a highly visual, modern Terminal UI (TUI) powered by `pterm` for `exe1`, `exe2`, and `webapps` commands. Added interactive loading spinners with anti-ghosting fixed-width padding (`%-70s`) and beautifully structured execution reports. Maintained strict backward compatibility by preserving pure JSON output streams via the `-j` flag for CI/CD pipeline automation.
- **v5.0.2 (May 2026) - Secure Web Apps Protocol Upgrade**
  Upgraded the `webapps` command to natively support "Only myself" execution deployments by bridging OAuth tokens (`drive` scope) across Google's HTTP 302 Auth Redirects. Ported the IIFE/JSON-literal double-eval protections from `exe2` to `webapps`.
- **v5.0.1 (May 2026) - Execution Engine Hardening & Double-Eval Eradication**
  Eliminated the V8 engine double-eval 500 server crash during dynamic script execution by enforcing IIFE and JSON-literal payload encoding. Redefined `-f` flag mapping for proper API gateway resolution in `exe2`. Added precision deployment documentation for stateful and stateless execution modes.
- **v5.0.0 (May 2026) - The Omnibus Architecture Rewrite**
  Engine fundamentally rewritten for Go 1.26.3+. Implemented channel-based concurrency (`errgroup`), freeze-proof TUI (`mpb/v8`), SIMD JSON parsing (`goccy/go-json`), native MCP server (`ggsrun mcp`), Shared Drives full-support, auto MIME-mapping, isolated fault tolerance, and OAuth2 loopback automation.
- **v3.2.2 (May 2026) - Pure MCP Node Evolution**
  Finalized the `mcp` command backend logic.
- **v3.2.0 (May 2026) - The AI/MCP Architecture Update**
  Transformed `ggsrun` into a background daemon capability. Redefined Config and Credentials path priority.
- **v3.1.0 (May 2026) - Recursive Structure Update**
  Re-engineered Drive file transfer logic mapping deeply nested structures.
- **v3.0.0 (May 2026) - Massive Concurrency Update**
  Core engine rewritten for Go 1.26+. Deprecated OOB OAuth.
- **v2.0.3 (June 2025)** Rebuild with go1.24.4.
- **v2.0.0 (February 2022)** Modified using the latest libraries.
- **v1.7.0 (December 2018)** Manage permissions; Service Account integration.
- **v1.6.0 (November 2018)** Files searchable via query and regex.
- **v1.5.0 (October 2018)** Recursive folder downloads while maintaining structure.
- **v1.4.1 (February 2018)** Resumable upload chunking added.
- **v1.4.0 (January 2018)** Official Google Apps Script API integration.
- **v1.3.3 (October 2017)** Manifest modification support (`appsscript.json`).
- **v1.3.2 (October 2017)** Interactive script rearrangement feature.
- **v1.3.0 (August 2017)** Container-bound script support.
- **v1.2.1 (May 2017)** Added `GGSRUN_CFG_PATH` environment variable support.
- **v1.1.0 (April 2017)** Update project command and `TotalElapsedTime` additions.
- **v1.0.0 (April 2017)** Initial release.

### Server

- **v1.0.0 (April 2017)** Initial release.

[Back to Top](#top)
