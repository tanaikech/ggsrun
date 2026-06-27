# ggsrun - Manual Verification and Testing Guide

This directory contains utility files, test scripts, and structured guidelines to manually verify the features, commands, and security sandboxing of `ggsrun`.

---

## Prerequisites

Before starting, ensure that:
1. You have built the latest `ggsrun` binary in the repository root:
   ```bash
   go build -o ggsrun main.go
   ```
2. You have a valid configuration file (`ggsrun.cfg`) configured on your system (e.g., in `~/.config/` or `~/myTools/`). You can specify its path using the `GGSRUN_CFG_PATH` environment variable if it is placed in a custom directory.

---

## 1. Authentication & System Health Status

Verify that `ggsrun` can successfully resolve configuration paths, check credential validity, and connect to Google APIs.

**Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun status
```

**Expected Result:**
* A successful response displaying the resolved config path, valid token status, and Google Drive connection confirmation.

---

## 2. Model Context Protocol (MCP) Server

`ggsrun` includes a native Model Context Protocol (MCP) server that listens over stdin/stdout, allowing LLM clients (like Claude, Gemini, etc.) to securely list and invoke Google Drive/GAS tools.

### Test Case A: Server Initialization
Sends an MCP standard `initialize` request to verify the server initializes and returns correct capability schemas.

**Run Command:**
```bash
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}},"id":1}' | GGSRUN_CFG_PATH=~/myTools ./ggsrun mcp
```

**Expected Result:**
* The server outputs initialization/diagnostic messages on `stderr` (e.g. `🤖 ggsrun MCP Server initialized`).
* It outputs a single, clean JSON-RPC response on `stdout` matching the standard structure:
  ```json
  {"id":1,"jsonrpc":"2.0","result":{"capabilities":{"tools":{}},"protocolVersion":"2024-11-05","serverInfo":{"name":"ggsrun-mcp-server","version":"5.3.4"}}}
  ```

### Test Case B: List Available Tools
Retrieves the definitions and input schemas of all tools exposed by the MCP server.

**Run Command:**
```bash
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}' | GGSRUN_CFG_PATH=~/myTools ./ggsrun mcp
```

**Expected Result:**
* A JSON-RPC response on `stdout` containing the array of exposed tools (`searchfiles`, `download`, `upload`, `exe1`, `filelist`) along with their descriptions and schema constraints.

---

## 3. Security Sandbox & Whitelisting (`exe1`)

Verify that local/remote Google Apps Script execution can be sandboxed using the native `--sandbox` option.

### Test Case A: Sandbox Bypass / Direct Execution
In this mode, sandboxing is completely disabled. `ggsrun` will upload and execute the script directly on GAS without any wrapper injection.

**Run Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun exe1 -s manual-tests/test_sandbox.js -f main --conflict overwrite -d --sandbox bypass
```
*(Note: You can also use `--sandbox none`)*

**Expected Result:**
* The execution succeeds, but all unauthorized/dummy API calls (like `DriveApp.getFileById()` or `MailApp.sendEmail()`) will trigger **native Google Apps Script API errors** (e.g., "Item not found" or "Authorization required" native errors) instead of being intercepted by the sandbox wrapper.

---

### Test Case B: Default Ultra-Strict Sandbox (Omitted Config)
When `--sandbox` is left empty or omitted, `ggsrun` defaults to applying a strict sandbox where all whitelists are empty.

**Run Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun exe1 -s manual-tests/test_sandbox.js -f main --conflict overwrite -d --sandbox ""
```

**Expected Result:**
* Every external API call (Drive, Mail, URL Fetch) will be intercepted and blocked by the wrapper, throwing a `Sandbox Runtime Blocked` error.
* Output logs will report `[FAIL]` for allowed tests and `[PASS]` for blocked tests because the strict default blocks **everything** unconditionally:
  ```json
  [
    "--- UrlFetchApp Tests ---",
    "[FAIL] URL connection failed (Unexpected): https://httpbin.org/anything/allowed -> Sandbox Runtime Blocked: URL 'https://httpbin.org/anything/allowed' is not whitelisted. Default policy is BLOCK ALL.",
    "[PASS] URL connection blocked (Expected): https://httpbin.org/anything/blocked -> Sandbox Runtime Blocked: URL 'https://httpbin.org/anything/blocked' is not whitelisted. Default policy is BLOCK ALL.",
    "[PASS] URL connection blocked (Expected): https://example.com/unregistered -> Sandbox Runtime Blocked: URL 'https://example.com/unregistered' is not whitelisted. Default policy is BLOCK ALL.",
    ...
  ]
  ```

---

### Test Case C: Whitelist-Configured Sandboxing
In this mode, you pass a user-defined configuration JSON file listing the allowed resources.

**Run Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun exe1 -s manual-tests/test_sandbox.js -f main --conflict overwrite -d --sandbox manual-tests/sandbox_config.json
```

**Expected Result:**
* Whitelisted items will bypass the sandbox wrapper (and report `[PASS]`).
* Blacklisted or unregistered items will be caught by the wrapper and blocked (reporting `[PASS]`).
* You should see **100% PASS** results in the returned array:
  ```json
  [
    "--- UrlFetchApp Tests ---",
    "[PASS] URL connection succeeded (Expected): https://httpbin.org/anything/allowed",
    "[PASS] URL connection blocked (Expected): https://httpbin.org/anything/blocked -> Sandbox Runtime Blocked: URL 'https://httpbin.org/anything/blocked' is explicitly blacklisted.",
    "[PASS] URL connection blocked (Expected): https://example.com/unregistered -> Sandbox Runtime Blocked: URL 'https://example.com/unregistered' is not whitelisted. Default policy is BLOCK ALL.",
    "--- DriveApp Tests ---",
    "[PASS] Drive file retrieval bypassed wrapper successfully (Expected API error since dummy ID is used): 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P -> ...",
    "[PASS] Drive file retrieval blocked (Expected): 9X8Y7Z6W5V4U3T2S1R0Q9P8O7N6M5L4K -> Sandbox Runtime Blocked: File ID '9X8Y7Z6W5V4U3T2S1R0Q9P8O7N6M5L4K' is not in the whitelist.",
    "--- MailApp Tests ---",
    "[PASS] Mail send bypassed wrapper successfully (Expected API scope authorization error): allowed-tester@example.com -> ...",
    "[PASS] Mail send blocked (Expected): blocked-tester@example.com -> Sandbox Runtime Blocked: Recipient address 'blocked-tester@example.com' is not whitelisted."
  ]
  ```

---

## 4. Drive & File Management Operations

Verify high-speed file operations and Drive queries.

### Test Case A: List Files
**Run Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun filelist --limit 5
```

### Test Case B: Search Files (Query & Regex)
**Run Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun searchfiles --query "name contains 'sandbox' and trashed = false"
```

---

## 5. Interactive PC-98 Style TUI File Manager

Verify the interactive split-screen TUI.

**Run Command:**
```bash
GGSRUN_CFG_PATH=~/myTools ./ggsrun fd
```

**Expected Result:**
* A terminal user interface starts, showing local files on one side and remote Google Drive folders on the other. Use arrow keys to navigate and `Tab` to switch panels. Press `q` to exit.
