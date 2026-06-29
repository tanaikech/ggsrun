---
name: gas-execution
description: Guidelines for writing and executing Google Apps Script code using ggsrun and finding documentation via workspace-developer. Use when developing or running GAS applications.
---

# Google Apps Script Execution and Development Skill

Follow these guidelines when writing, reviewing, and executing Google Apps Script (GAS) applications:

## Script Execution via ggsrun

* **Execution Tool (`exe1`)**: Always prefer the `exe1` tool of the `ggsrun` MCP server to synchronize local code/directories and run the entry function in a single transaction.
* **In-Memory Sandbox (`sandbox`)**: For secure execution, pass the path of your whitelist configuration JSON (e.g. `sandbox: "sandbox_config.json"`) to `exe1`. This instructs `ggsrun` to inject wrappers (like `_wrappedSpreadsheetApp` and `_wrappedUrlFetchApp`) that restrict access only to whitelisted File IDs, domains, and email addresses.
* **Auto-Cleanup & Resilient Rollback**: `ggsrun` performs an in-memory backup of the remote project state (including all script files and the `appsscript.json` manifest). On execution completion, runtime sandboxing halts, or process interrupts (`Ctrl+C`), `ggsrun` automatically rolls back all file changes on the remote script project, leaving your codebase 100% clean and pristine. This cleanup behavior is enabled by default. To bypass this cleanup and leave the uploaded files in the remote project, use the `undeleteScript: true` option (or `--undeleteScript` / `--ud` flag in CLI mode).
* **Intelligent Manifest Merging**: When executing scripts that require Advanced Services or custom dependencies, you can upload a local `appsscript.json` manifest. `ggsrun` will automatically merge your dependencies (e.g., `enabledAdvancedServices` or custom `libraries`) with the remote project's existing manifest without breaking critical configurations like `executionApi` or `webapp` needed for execution.
* **Self-Healing Recovery**: If a script execution is interrupted and the remote project is left in an inconsistent state, you can run the `ggsrun recover` command to immediately restore the project to the pristine ggsrun state.
* **Designing Return Values**: The execution response displays only the value returned by the `return` statement of the executed function. Design your script to return a meaningful value representing the execution result.
* **Detailed Logs**: You can return detailed logs or execution summaries as the final return string to provide context and results to the user.
* **JSON Serialization**: If returning structured data (such as objects, arrays, or status maps), use `JSON.stringify()` to serialize it before returning.

## Other ggsrun MCP Server Tools

Beyond the stateful script execution (`exe1`), you can leverage the following tools from the `ggsrun-mcp` server for managing files, projects, and directories on Google Drive:

### 1. `searchfiles`
* **Purpose**: Search for files or folders in Google Drive using standard Drive API v3 queries, with optional local regex filtering on filenames.
* **Usage Guide**: Specify the `query` (e.g., `name = 'MyScript' and trashed = false` or `mimeType = 'application/vnd.google-apps.folder'`). Optionally provide a `regex` pattern to filter result names.
* **Example Case**: Finding the folder ID named "Backup" or discovering spreadsheet files with a matching pattern:
  * Arguments: `{ "query": "mimeType = 'application/vnd.google-apps.spreadsheet' and trashed = false", "regex": "^Production_.*" }`

### 2. `filelist`
* **Purpose**: Perform exact matches by File Name or resolve filenames by File ID on Google Drive.
* **Usage Guide**: Use `searchbyname` to find IDs matching an exact name, or `searchbyid` to get the metadata of a specific ID.
* **Example Case**: Quick lookup of a target Spreadsheet ID before execution:
  * Arguments: `{ "searchbyname": "TargetSpreadsheet" }`

### 3. `download`
* **Purpose**: Download files or entire recursive folder structures to your local system, supporting on-the-fly Google Workspace format export.
* **Usage Guide**: Provide the target `fileid`. Use the `extension` parameter to convert Google Docs, Sheets, Slides, or Drawings to local formats (e.g., Doc -> PDF/Markdown, Sheet -> Excel/CSV).
* **Example Case**: Fetching a Google Spreadsheet as an `.xlsx` file into a local `./downloads` folder:
  * Arguments: `{ "fileid": "SPREADSHEET_ID", "extension": "xlsx", "destination": "./downloads", "conflict-mode": "OverwriteIfNewer" }`

### 4. `upload`
* **Purpose**: Upload local files or recursive directory structures to Google Drive.
* **Usage Guide**: Specify the local `filename` to upload. By default, it automatically maps extensions to convert them into Google Workspace formats. You can set `convertto` (e.g., 'sheet', 'doc', 'slide') or use `noconvert: true` to bypass.
* **Example Case**: Uploading local CSV reports into Google Drive and converting them into Google Sheets under a specific folder:
  * Arguments: `{ "filename": "./data/report.csv", "convertto": "sheet", "parentfolderid": "FOLDER_ID" }`

### 5. `updateproject`
* **Purpose**: Synchronize local scripts to an existing remote GAS project.
* **CRITICAL Agent Rules**:
  1. Do NOT use this tool if you need to run the script afterward; use `exe1` instead.
  2. **User Confirmation Required**: Since it overwrites files remotely, you must display the list of local files to the user and obtain their permission (Y/N) before executing.
  3. If conflict resolution is ambiguous, ask the user whether to `overwrite` or `add` duplicates.
* **Example Case**: Deploying a folder of source files to an existing Apps Script project ID:
  * Arguments: `{ "projectid": "SCRIPT_ID", "filename": "./src/", "conflict": "overwrite", "backup": true }`

## Verifying GAS API Usage
* **Documentation Search**: When you are unsure about the methods, behaviors, or parameters of built-in GAS classes (e.g., `DriveApp`, `GmailApp`, `SpreadsheetApp`), query the `workspace-developer` MCP server to fetch detailed class references and usage examples.

## Code Review & Security Checklist

When reviewing or before executing generated GAS code, you must strictly perform a security code review.

### Security Checklist

1. **Google Drive & Document Access**: Does the script read, write, or delete files/folders (`DriveApp`, `SpreadsheetApp`, `DocumentApp`, `SlidesApp`)? Verify that only authorized files/folders are modified or deleted.
2. **Gmail & Mailing**: Does the script read emails, drafts, or send messages (`GmailApp`, `MailApp`)? Check if recipient addresses and message contents are safe and authorized.
3. **Calendar Events**: Does the script read, write, or delete calendar events (`CalendarApp`)? Confirm that only designated Calendars and Events are manipulated.
4. **Outbound Network Connections**: Does the script fetch external resources (`UrlFetchApp`)? Validate the target URL to ensure no credentials or sensitive data are being exfiltrated to untrusted endpoints.
5. **Hardcoded Secrets**: Are there any hardcoded API keys, OAuth tokens, or passwords? Ensure no credentials are exposed in the source code.
6. **Destructive Actions**: Does the code perform any bulk or irreversible deletion/overwriting of data?

### How to Proceed with Execution & Safety Gate

Before executing the script using `ggsrun`'s `exe1` tool, follow this secure orchestration workflow:

1. **Summarize Accessed Services**: Clearly show the user which Google APIs and external URLs the script will access.
2. **Configure Whitelists**: Ask the user to verify if they have configured `sandbox_config.json` (or your specific config file) to include the required File IDs, folder IDs, emails, and URLs.
3. **Handle MCP Safety Gate**:
   * For safety, the MCP server automatically intercepts `exe1` calls and performs a static analysis check (`analyzeGASScript`).
   * If any potential write/egress APIs are detected, and `"confirm": false` (or omitted), execution is **blocked** and returns a warning report.
   * To proceed with execution on authorized projects, you **MUST explicitly call the tool again setting `"confirm": true`**. Do not guess or skip this parameter when you have user consent.
