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
  - [The 5 Pillars of v5.0.0](#the-5-pillars-of-v500)
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
  - [Command Reference \& Usage](#command-reference--usage)
    - [Authentication](#authentication)
    - [Massively Parallel Download](#massively-parallel-download)
    - [Massively Parallel Upload](#massively-parallel-upload)
    - [Execute GAS (Google Apps Script)](#execute-gas-google-apps-script)
    - [MCP Server Mode (LLM Integration)](#mcp-server-mode-llm-integration)
  - [Under the Hood: v5.0.0 Architecture](#under-the-hood-v500-architecture)
  - [Advanced Configurations](#advanced-configurations)
  - [Troubleshooting](#troubleshooting)
  - [Licence \& Author](#licence--author)
  - [Update History](#update-history)
    - [ggsrun](#ggsrun-1)
    - [Server](#server)

---

## Overview

**ggsrun** is an enterprise-grade CLI tool and MCP (Model Context Protocol) Server designed to relentlessly orchestrate Google Drive I/O operations and execute Google Apps Script (GAS) natively from a local terminal.

With the release of **v5.0.0**, `ggsrun` transcends its origins as a mere CLI tool. Built on Go 1.26.3+, the execution engine has been entirely rewritten from legacy serial processing into a channel-based, streaming concurrent architecture. It now serves as a high-performance, fault-tolerant I/O backend fully integrated with Omni-Drive (Shared Drives) support, advanced MIME resolution, and a native **MCP Server Mode** allowing LLM agents to autonomously manage your cloud infrastructure.

Existing commands maintain 100% backward compatibility, but their underlying execution speeds and stability have been exponentially magnified.

---

## The 5 Pillars of v5.0.0

The v5.0.0 architecture is anchored on five core technical capabilities:

### A. Massively Parallel I/O & UI

Legacy pseudo-asynchronous processing has been eradicated. `ggsrun` now utilizes a channel-based worker pool built on `golang.org/x/sync/errgroup` to maximize network throughput on massive folder trees.

- Utilize the `--workers` (or `-w`) flag (default: `5`) to dictate parallel execution limits.
- Powered by `github.com/vbauerster/mpb/v8` and `github.com/pterm/pterm`, the completely freeze-proof Terminal UI dynamically rendering concurrent job allocations, structural trees, and real-time MB/s metrics. Edge-case UI hangs from zero-byte files, unmeasurable Google Docs exports, or volatile network environments have been completely engineered out of the system.

### B. Full Shared Drive (Omni-Drive) Support

The v5.0.0 engine forces `supportsAllDrives=true` and `includeItemsFromAllDrives=true` across all Google Drive API permutations (recursive tree searches, metadata extraction, uploads/downloads). Enterprise users can now execute bulk batch operations targeting deeply nested structures within organizational Shared Drives without arbitrary 404/403 permission failures.

### C. Intelligent GAS & MIME Resolution

The extraction logic dynamically categorizes Google workspace entities:

- **Smart API Routing:** Requests targeting GAS code (`application/vnd.google-apps.script`) bypass the standard Drive API and are automatically routed to the Apps Script API (`/v1/projects/{scriptId}/content`), securely landing as structured `.json` locally.
- **Unexportable Entity Skipping:** Hardcoded exclusions for inherently unexportable workspace elements (Google Sites, Maps, Forms, Shortcuts) prevent catastrophic 400 Bad Request errors, elegantly skipping them before API dispatch.
- **Dynamic Office Mapping:** Declaring extensions via `-e xlsx` or `-e docx` triggers the internal MIME translation engine, instructing Google Drive to transpile Workspace documents to native Microsoft Office binaries on the fly.

### D. Robust Fault Tolerance & Auto-Retry

The v5.0.0 execution phase is strictly non-blocking.

- **Batch Isolation:** A 400 (Bad Request) or 403 (Forbidden) on a singular file will never crash an overarching batch job; failures are isolated, tagged as warnings, and execution continues to 100% completion.
- **Exponential Backoff Engine:** HTTP 429 (Rate Limits) and 5xx (Server Errors) trigger a mathematical backoff sequence per-worker (1s -> 2s -> 4s, maximum 3 retries) ensuring aggressive self-healing without manual intervention.

### E. MCP (Model Context Protocol) Integration

The ultimate game-changer for AI ecosystems. Running `ggsrun mcp` transforms the application into an autonomous JSON-RPC server via `stdio`. Large Language Model (LLM) agents (Claude Desktop, Cursor, Gemini) can natively invoke internal capabilities: `searchfiles`, `download`, `upload`, and `exe1`.

- Human-readable logs and progress bars are strictly sequestered to `Stderr`.
- Pure JSON-RPC payloads are exclusively routed through `Stdout`.
- The LLM orchestrator requires zero API keys or OAuth states; it inherits `ggsrun`'s secured loopback context.

---

## Installation & Setup

### 1. Install ggsrun

Requires Go 1.26.3 or higher. Pull and compile the latest binary natively:

```bash
$ go install github.com/tanaikech/ggsrun@latest
```

### 2. Obtain Google Cloud Credentials

To authenticate against Google APIs, `ggsrun` requires an OAuth 2.0 Client ID.

1. Access the [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new Project.
3. Navigate to **APIs & Services > Library**. Enable both the **Google Drive API** and **Google Apps Script API**.
4. Configure the **OAuth consent screen** (External/Internal, add your email to Test Users).
5. Navigate to **Credentials > Create Credentials > OAuth client ID**. Select **Desktop app**.
6. Download the resulting JSON file, move it to your working directory, and rigorously rename it to exactly `client_secret.json`.

### 3. Automated Authorization (OAuth2 Loopback)

The deprecated Out-Of-Band (OOB) manual copy-paste flow is completely dead. With your `client_secret.json` in the current directory, execute:

```bash
$ ggsrun auth
```

`ggsrun` spins up a secure local loopback listener (default `localhost:8080`). Your default browser will launch, request authorization, and securely hand the token back to the CLI. A persistent `ggsrun.cfg` file is generated.

_(Note: Port allocation can be modified via `--port` if `8080` is saturated)._

### 4. Set Up Execution Server (GAS Side)

To execute arbitrary GAS functions locally, you must establish an endpoint server on Google Apps Script.

1. Navigate to [script.google.com](https://script.google.com/) > New Project.
2. Add Library (`+` icon). Target Script ID: `115-19njNHlbT-NI0hMPDnVO1sdrw2tJKCAJgOTIAPbi_jq3tOo4lVRov`.
3. Set Identifier to `ggsrunif`, select the latest version.
4. Deploy the project via **Deploy > New Deployment > API Executable**.

---

## Command Reference & Usage

`ggsrun` provides absolute command-line parity with its backend capabilities. Below are structural examples of v5.0.0 functionality.

### Authentication

| Command       | Description                                | Flags                    |
| :------------ | :----------------------------------------- | :----------------------- |
| `ggsrun auth` | Initiates the secure OAuth2 loopback flow. | `--port` (default: 8080) |

### Massively Parallel Download

Target IDs can belong to Standard Drives, Shared Drives, or Team Drives seamlessly.

| Command                                         | Action                                                                                      |
| :---------------------------------------------- | :------------------------------------------------------------------------------------------ |
| `$ ggsrun download -i "FOLDER_ID" -w 5`         | Recursively maps and downloads a folder tree utilizing 5 parallel channel workers.          |
| `$ ggsrun download -i "SPREADSHEET_ID" -e xlsx` | Directs the Drive API to transpile and export a native Google Sheet into an `.xlsx` binary. |

_Note: If the `-f` flag (local filename) is omitted during multiple file downloads, v5.0.0 dynamically falls back to the native Google Drive file name._

### Massively Parallel Upload

Pushes local hierarchical structures to Google Drive asynchronously. Resumable chunked uploads are inherently supported for massive binaries.

| Command                                                                | Action                                                                                               |
| :--------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------- |
| `$ ggsrun upload -f "/path/to/local/folder" -p "DEST_FOLDER_ID" -w 10` | Uploads the directory recursively to the target Google Drive Folder utilizing 10 concurrent threads. |

### Execute GAS (Google Apps Script)

| Command                                                         | Action                                                                                                                  |
| :-------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------- |
| `$ ggsrun e1 -i "SCRIPT_ID" -f "myFunction" -v "argumentValue"` | Passes payload `argumentValue` directly to `myFunction` inside the target GAS project and returns the evaluated result. |

### MCP Server Mode (LLM Integration)

Transforms `ggsrun` into a persistent JSON-RPC node.

| Command        | Action                                                                                                     |
| :------------- | :--------------------------------------------------------------------------------------------------------- |
| `$ ggsrun mcp` | Starts the stdio-bound MCP Server. Listens for tools like `searchfiles`, `download`, `upload`, and `exe1`. |

**MCP Configuration Example (Cursor / Claude Desktop):**
Because LLMs run background daemons lacking standard `$PWD` scopes, anchor the context path via the `GGSRUN_CFG_PATH` environment variable.

```json
{
  "mcpServers": {
    "ggsrun-drive-agent": {
      "command": "ggsrun",
      "args": ["mcp"],
      "env": {
        "GGSRUN_CFG_PATH": "/absolute/path/to/your/credentials/dir"
      }
    }
  }
}
```

---

## Under the Hood: v5.0.0 Architecture

- **SIMD JSON Processing:** `ggsrun` v5.0.0 leverages `github.com/goccy/go-json`. By utilizing SIMD (Single Instruction, Multiple Data) CPU instructions, the deserialization of massive Google Drive JSON payloads is executed magnitudes faster than standard library parsers.
- **Worker Choking & Memory Preservation:** By bounding goroutines within strict `errgroup` pools, `ggsrun` maintains aggressive throughput without exhausting system RAM or violating host OS file descriptor limits.
- **Deterministic Type Masking:** When handling recursive operations, unexportable formats (like Google Sites) are identified at the metadata parsing layer and structurally purged from the channel queue before arbitrary bytes are requested, cutting latency and eliminating API thrashing.

---

## Advanced Configurations

For robust integration into CI/CD pipelines or background daemon modes, authentication routing relies on strict hierarchical priority resolution:

1. **Explicit Flags:** Highest priority (`--config` and `--credentials`).
2. **Environment Variable:** `GGSRUN_CFG_PATH`.
3. **Working Directory:** Standard `$PWD` execution.

Defining `GGSRUN_CFG_PATH` guarantees state stability regardless of where the binary is invoked on the host OS.

---

## Troubleshooting

**1. "Requested entity was not found" or 404 Errors**
If utilizing GAS execution (`e1` / `e2`), verify the target project is currently deployed as an **API Executable** on the latest version. Un-deployed or draft states cannot be invoked externally.

**2. "Script Error on GAS side: Insufficient Permission"**
The remote GAS project requires higher authorization scopes than currently granted. Open the web-based Google Apps Script editor, run the script manually once, and clear the Google security prompt.

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
