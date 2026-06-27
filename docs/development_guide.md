# ggsrun - Local Development and Testing Guide

This guide describes how to configure, develop, and execute the automated test suites for `ggsrun`.

---

## Table of Contents
1. [Overview](#1-overview)
2. [Environment Configuration (.env)](#2-environment-configuration-env)
3. [Test Suites Structure](#3-test-suites-structure)
   - [CLI Mode Tests (cli_test.go)](#cli-mode-tests-cli_testgo)
   - [MCP Server Tests (mcp_test.go)](#mcp-server-tests-mcp_testgo)
   - [FD (TUI) Mode Tests (fd_test.go)](#fd-tui-mode-tests-fd_testgo)
4. [Running the Tests](#4-running-the-tests)
5. [Writing New Tests](#5-writing-new-tests)
   - [Adding a TUI Event Test](#adding-a-tui-event-test)
   - [Mocking Google Drive Interactions](#mocking-google-drive-interactions)

---

## 1. Overview

`ggsrun` contains an extensive, robust suite of unit and integration tests. These tests are divided into separate packages and modules, allowing developers to verify CLI logic, MCP JSON-RPC routing, and Interactive TUI event-loops synchronously or asynchronously.

---

## 2. Environment Configuration (`.env`)

To run the automated tests successfully, you should configure a local `.env` file in the root directory of the repository. This file is excluded from version control (`.gitignore`) to protect your private credentials and target IDs.

Create a file named `.env` in your repository root and configure the following variables (replacing placeholders with your own Google Drive test IDs):

```env
# Google Drive File IDs for testing (prepare your own files, do NOT use these IDs)
GGSRUN_TEST_GAS_PROJECT_FILEID_ON_GOOGLE_DRIVE="YOUR_TEST_GAS_PROJECT_FILE_ID"
GGSRUN_TEST_GOOGLE_DOCS_FILEID_ON_GOOGLE_DRIVE="YOUR_TEST_GOOGLE_DOCS_FILE_ID"
GGSRUN_TEST_PDF_FILEID_ON_GOOGLE_DRIVE="YOUR_TEST_PDF_FILE_ID"

# Local test data paths (pre-packaged inside the repository)
TESTDATA_GAS_SCRIPT1="internal/app/testdata/sampleGAS"
TESTDATA_IMAGE1="internal/app/testdata/sampleImage1.png"
TESTDATA_MARKDOWN1="internal/app/testdata/sample.md"
TESTDATA_PDF1="internal/app/testdata/sample_pdf_1.pdf"

# Local test data paths for TUI/FD Mode
TESTDATA2_GAS_SCRIPT2="internal/tui/testdata/sampleGAS"
TESTDATA_IMAGE2="internal/tui/testdata/sampleImage1.png"
TESTDATA_MARKDOWN2="internal/tui/testdata/sample.md"
TESTDATA_PDF2="internal/tui/testdata/sample_pdf_1.pdf"
```

> [!CAUTION]
> **Never commit your private `.env` file or Google Cloud credentials to Git.** The `.gitignore` in this project is pre-configured to block `.env` and `client_secret.json` from being tracked.

---

## 3. Test Suites Structure

The `ggsrun` test harness is modular and categorized as follows:

### CLI Mode Tests (`internal/app/cli_test.go`)
* **Role**: Validates CLI command-line parsing, argument validation rules, and structural logic.
* **Checks**:
  - Verification of download file conversion formats (`.gs` code to local `.js`/`.html` packages).
  - Validation of automatic zip-download file packing (`download -z`).
  - Correct routing of `.js`/`.gs`/`.gas` uploads to the Apps Script project builder instead of triggering standard Drive resumable uploads.
  - Verification of parallel download and upload streams.

### MCP Server Tests (`internal/app/mcp_test.go`)
* **Role**: Verifies that the JSON-RPC interface over stdin/stdout conforms strictly to the Model Context Protocol standard.
* **Checks**:
  - Validates `initialize` requests and JSON-RPC handshakes.
  - Verifies that all registered tools (`searchfiles`, `download`, `upload`, `exe1`, `filelist`) have complete, syntactically correct input schemas.
  - Validates tool invocation payload structures and error wrapping.

### FD (TUI) Mode Tests (`internal/tui/fd_test.go`)
* **Role**: Simulates standard user keyboard and cursor interactions inside the split-screen file manager.
* **Checks**:
  - Spawns a robust virtual screen using `tcell.SimulationScreen`.
  - Simulates keyboard keystrokes (`Tab`, `Up`/`Down`, `Space`, `Enter`).
  - Verifies list wrap-around logic and focus persistent locks.
  - Asserts correct centering and width of 70% responsive dialogs.
  - Validates file deletion and sort popup state transitions.

---

## 4. Running the Tests

To run the entire test suite, make sure you are in the project root directory and run the following command to run all packages uncached:

```bash
$ go test -count=1 ./...
```

If you only want to run the TUI or MCP specific test suites:

```bash
# Run only TUI/FD Mode simulation tests
$ go test -v ./internal/tui/

# Run only CLI and MCP tool tests
$ go test -v ./internal/app/
```

---

## 5. Writing New Tests

### Adding a TUI Event Test

When adding new interactive features or keybindings in FD Mode, you can easily write automated keyboard simulations using `tcell.SimulationScreen` in `internal/tui/fd_test.go`:

```go
func TestNewKeybinding(t *testing.T) {
    // Create mock screen
    screen := tcell.NewSimulationScreen("")
    app := tview.NewApplication().SetScreen(screen)

    // Initialize your TUI tables
    // ...

    // Inject keystroke event
    screen.InjectKey(tcell.KeyTab, ' ', tcell.ModNone)
    app.Draw()

    // Assert screen results or focused state
    // ...
}
```

### Mocking Google Drive Interactions

For integration tests that connect to Google Drive, `ggsrun` uses standard interfaces. You can supply mock Google Drive API client handlers under `internal/app/` to isolate local testing from remote cloud network latencies.

---

### Related Links:
- 🚀 **[Setup & Onboarding Guide](setup_guide.md)** - Learn how to build and configure `ggsrun`.
- 📖 **[Command Reference Manual](commands_reference.md)** - Reference CLI parameters.
- 🏡 **[Back to Home](../README.md)**
