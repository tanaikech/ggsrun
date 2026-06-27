# ggsrun

![](help/images/fig1a.jpg)

<a name="top"></a>
[![Go Version](https://img.shields.io/badge/Go-1.26.4+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![MCP Ready](https://img.shields.io/badge/MCP-Ready-8A2BE2?style=for-the-badge)](https://modelcontextprotocol.io)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=for-the-badge)]()
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge)](LICENCE)

---

## 🚀 Overview

**ggsrun** is an enterprise-grade CLI application and Model Context Protocol (MCP) Server designed to relentlessly orchestrate Google Drive I/O operations and statefully execute Google Apps Script (GAS) natively from your local terminal or via autonomous Large Language Model (LLM) agents.

Built on Go 1.26.4+, `ggsrun` transitions legacy single-threaded processing into a state-of-the-art, channel-based concurrent streaming architecture. It delivers maximum network throughput, native Shared Drives (Omni-Drive) support, advanced MIME resolution, resilient self-healing retries, and a native **Security Sandbox** to restrict Google Workspace APIs during autonomous executions.

---

## 🗺️ Documentation Directory Index

To make onboarding, development, and advanced operations as seamless as possible, our documentation is fully modularized. Navigate directly to your area of interest:

| Guide / Manual | Description | Link |
| :--- | :--- | :--- |
| 🚀 **Setup & Onboarding Guide** | Step-by-step instructions to configure GCP, generate OAuth2 client secrets, link Google Apps Script, and deploy gateway endpoints. | **[Setup Guide](docs/setup_guide.md)** |
| 📖 **Command Reference Manual** | In-depth breakdown of all 15+ subcommands (e.g. `exe1`, `exe2`, `download`, `upload`), recipes, and Mermaid architectural diagrams. | **[Command Reference](docs/commands_reference.md)** |
| 🛡️ **Security Sandbox Guide** | Explains the memory-based wrapper injection, whitelist configuration schemas (`sandbox_config.json`), and security validation scenarios. | **[Sandbox Guide](docs/sandbox_guide.md)** |
| 🔄 **Stateful Execution Lifecycle** | Deep-dive explanation of the backup, sandboxing, upload, execution, and resilient rollback phases of the `exe1` command. | **[Execution Lifecycle](docs/exe1_lifecycle.md)** |
| 🤖 **MCP Server Manual** | Configuration to run `ggsrun` as an autonomous tool provider in AI environments (like Claude Desktop or Antigravity), schemas, and scenarios. | **[MCP Server Guide](docs/mcp_guide.md)** |
| 💻 **Interactive TUI Filer Guide** | All keyboard shortcuts, search highlighter rules, clipboard integration, and history for the split-screen Japanese PC-98 Filer Mode (`ggsrun fd`). | **[TUI Filer Guide](docs/tui_guide.md)** |
| 🧪 **Local Development & Testing** | Detailed guide for configuring environment variables (`.env`), understanding CLI/TUI mock tests, and running automated test suites. | **[Development Guide](docs/development_guide.md)** |
| 🔬 **Manual Integration Tests Suite** | Guided verification commands to manually test authorization, sandbox policies, tool schemas, and local file operations. | **[Manual Tests Suite](manual-tests/README.md)** |

---

## ✨ Features of ggsrun

1. **Terminal GAS Development**: Develop Google Apps Script using your favorite local text editors and terminals seamlessly.
2. **Dynamic Script Execution**: Execute GAS directly by injecting values, arguments, and payloads into your scripts dynamically.
3. **Massively Parallel Downloads**: Pull files and folders concurrently from Google Drive with stunning progress visualizations.
4. **Massively Parallel Uploads**: Push folders recursively with native Resumable upload wrappers and automated chunks handling.
5. **Flexible Project Formats**: Download standalone scripts or container-bound projects flawlessly.
6. **Recursive Folder Structuring**: Map local directories to Google Drive folders recursively, retaining absolute directory trees.
7. **Multi-Format Container Synced Uploads**: Upload script files and instantly provision standalone scripts OR container-bound scripts.
8. **Permissions Orchestration**: Inspect, list, and manage file and folder sharing permissions across your entire Drive.
9. **Advanced Metadata Search**: Query your Google Drive utilizing Google Drive API v3 query syntax and local filename Regular Expressions (Regex).
10. **Flexible Authentication**: Natively supports both robust browser loopback OAuth2 and secure Service Accounts.
11. **Security Sandboxing**: Officially integrates with the **Antigravity CLI** via an embedded in-memory security sandbox wrapper (`--sandbox`) to guard Workspace resources.

---

## 🏛️ The 5 Pillars of the v5 Architecture

### A. Massively Parallel I/O & UI
Legacy single-threaded processing has been completely replaced. `ggsrun` implements a channel-based worker pool built on `golang.org/x/sync/errgroup` to maximize network throughput on massive folder hierarchies.
* *Learn more in the **[Command Reference Manual](docs/commands_reference.md#4-file-and-directory-transfers)**.*

### B. Full Shared Drive (Omni-Drive) Support
The v5 engine forces `supportsAllDrives=true` and `includeItemsFromAllDrives=true` across all Google Drive API permutations, allowing enterprise Shared Drives to be mapped and managed seamlessly.
* *Learn more in the **[Command Reference Manual](docs/commands_reference.md#5-drive-querying-and-search)**.*

### C. Intelligent GAS & MIME Resolution
The extraction logic dynamically categorizes Google Workspace entities. Downloader rules bypass the standard Drive API for GAS files, automatically routing to the Apps Script API to download scripts natively as structured JSON or packaged ZIP archives.
* *Learn more in the **[Command Reference Manual](docs/commands_reference.md#4-file-and-directory-transfers)**.*

### D. Robust Fault Tolerance & Auto-Retry
The v5 execution and transfer phase is strictly non-blocking. Google API Rate Limits (HTTP 429) and Server Errors (5xx) trigger an exponential backoff sequence per-worker, ensuring resilient self-healing.
* *Learn more in the **[Command Reference Manual](docs/commands_reference.md#8-conflict-resolution-guide)**.*

### E. MCP (Model Context Protocol) Integration
Running `ggsrun mcp` transforms the application into an autonomous JSON-RPC background server via standard I/O, allowing Large Language Model (LLM) agents to search, transfer, and execute scripts safely.
* *Learn more in the **[MCP Server Manual](docs/mcp_guide.md)**.*

---

## 📦 Rapid Installation

### Method A: Build from Source (Using Go)
Requires Go 1.26.4 or higher installed. Compile and install the binary natively:
```bash
$ go install github.com/tanaikech/ggsrun@latest
```

### Method B: Download Pre-built Binaries
Alternatively, download the pre-compiled binary matching your CPU architecture and operating system from the [Official Releases Page](https://github.com/tanaikech/ggsrun/releases):
* **macOS (Darwin)**: `ggsrun_darwin_amd64` (Intel) or `ggsrun_darwin_arm64` (Apple Silicon M1/M2/M3)
* **Linux**: `ggsrun_linux_amd64` (or matching ARM, 32-bit, or MIPS variants)
* **Windows**: `ggsrun_windows_amd64.exe` (or 32-bit, ARM variants)

---

## 🧪 Post-Installation Verification

Once you have downloaded or compiled `ggsrun`, verify that the binary is functional inside your terminal:

### 1. Verify Command Help
Run the help display command to verify the binary executes and lists available options:
```bash
$ ggsrun --help
```

### 2. Run Diagnostics
Once you have gone through the quick setup process (**[detailed step-by-step in the Onboarding Guide](docs/setup_guide.md)**), verify API connectivity and loopback token health by running:
```bash
$ ggsrun status
```

---

## 💬 Q&A & Troubleshooting

For general Q&A, standard Google API errors, and runtime limits, please refer to the detailed **[Legacy Q&A Guide](help/README.md#qa)**.

For setup, Web App redirects, headless authentications, and 404s, please consult the **[Setup Guide Troubleshooting Section](docs/setup_guide.md#7-troubleshooting-diagnostics)**.

---

## 📄 License & Author

* **License**: [MIT License](LICENCE)
* **Author**: [Tanaike](https://tanaikech.github.io/about/) (Contact: tanaike@hotmail.com)  
  *For advanced enterprise integrations, custom architectural consultations, or security audits.*
