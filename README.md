# ggsrun

![](help/images/fig1a.jpg)

[![Go Version](https://img.shields.io/badge/Go-1.26.4+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![MCP Ready](https://img.shields.io/badge/MCP-Ready-8A2BE2?style=for-the-badge)](https://modelcontextprotocol.io)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=for-the-badge)]()
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge)](LICENCE)

---

## 🚀 Overview

**ggsrun** is an enterprise-grade CLI application and Model Context Protocol (MCP) Server designed to manage Google Drive files and execute Google Apps Script (GAS) natively from a local terminal or via autonomous AI agents.

Built on Go 1.26.4+, `ggsrun` utilizes a modern streaming concurrent architecture using channel-based worker pools. This delivers maximum network throughput, native Shared Drives (Omni-Drive) support, advanced MIME type resolution, resilient rate-limit backoffs, and an intelligent **Security Sandbox** to restrict Google Workspace APIs during autonomous executions.

---

## 🗺️ Documentation Directory Index

To make onboarding as easy as possible, our documentation is fully modular. Choose from the dedicated guides below:

| Guide | Description |
| :--- | :--- |
| 🚀 **[Setup & Onboarding Guide](docs/setup_guide.md)** | Step-by-step instructions to configure Google Cloud, generate Desktop credentials, link your GAS project, bind execution libraries, and deploy API Executables/Web Apps. |
| 📖 **[Command Reference Manual](docs/commands_reference.md)** | Deep dive into all 15+ subcommands (such as `exe1`, `exe2`, `download`, `upload`), flag parameter definitions, conflict resolution strategies, and recipe examples. |
| 🛡️ **[Security Sandbox Guide](docs/sandbox_guide.md)** | Explains how the native script-level `--sandbox` runtime, memory-based proxy wrapper injection, and whitelist config schemas secure Google Workspace data. |
| 🤖 **[MCP Server Guide](docs/mcp_guide.md)** | Detailed setup to register `ggsrun mcp` as an autonomous tool provider inside LLM clients (such as Claude Desktop or Gemini Code Assist), JSON-RPC payloads, and prompt templates. |
| 🧪 **[Manual Integration Tests Suite](manual-tests/README.md)** | Guided integration tests to manually verify sandbox execution policies, MCP server requests, Drive querying, and the interactive split-screen TUI file manager. |

---

## 📦 Rapid Installation

### Method A: Build from Source (Using Go)
Ensure you have Go 1.26.4 or higher installed. Compile and install the binary natively:
```bash
$ go install github.com/tanaikech/ggsrun@latest
```

### Method B: Download Pre-built Binaries
Alternatively, download the pre-compiled binary matching your CPU architecture and operating system from the [Official Releases Page](https://github.com/tanaikech/ggsrun/releases):
* **macOS**: `ggsrun_darwin_amd64` (Intel) or `ggsrun_darwin_arm64` (Apple Silicon M1/M2/M3)
* **Linux**: `ggsrun_linux_amd64` (or ARM variant)
* **Windows**: `ggsrun_windows_amd64.exe`

---

## 🧪 Post-Installation Verification

Once you have downloaded or compiled `ggsrun`, you can instantly verify that the executable path is functional and resolve its environment status.

### 1. Verify Command Help
Verify the binary executes and lists available commands:
```bash
$ ggsrun --help
```

### 2. Verify Credentials Connection
Once you have gone through the quick configuration flow (**[detailed in the Setup Guide](docs/setup_guide.md)**), verify API connectivity and token health by running:
```bash
$ ggsrun status
```

---

## 💬 Q&A & Troubleshooting

For general Q&A, standard Google API errors, and runtime timeouts, please refer to the detailed **[Legacy Q&A Guide](help/README.md#qa)**.

---

## 📄 License & Author

* **License**: [MIT License](LICENCE)
* **Author**: [Tanaike](https://tanaikech.github.io/about/) (Contact: tanaike@hotmail.com)  
  *For advanced enterprise integrations, custom architectural consultations, or security audits.*
