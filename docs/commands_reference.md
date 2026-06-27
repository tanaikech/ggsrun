# ggsrun - Command Reference and Usage Guide

This document provides a highly detailed specification, flag listing, and practical execution examples for all commands supported by `ggsrun`.

---

## Table of Contents
1. [General Command Layout](#1-general-command-layout)
2. [Command Categorization](#2-command-categorization)
3. [Executing Google Apps Script](#3-executing-google-apps-script)
   - [exe1 (Stateful Script Execution)](#exe1-stateful-script-execution)
   - [exe2 (Stateless API Execution)](#exe2-stateless-api-execution)
   - [webapps (Web App Execution)](#webapps-web-app-execution)
4. [File and Directory Transfers](#4-file-and-directory-transfers)
   - [download (Parallel Pull)](#download-parallel-pull)
   - [upload (Parallel Push)](#upload-parallel-push)
   - [updateproject (Project Code Sync)](#updateproject-project-code-sync)
5. [Drive Querying and Search](#5-drive-querying-and-search)
   - [filelist (List Files)](#filelist-list-files)
   - [searchfiles (Advanced Search)](#searchfiles-advanced-search)
6. [Interactive Terminal File Manager](#6-interactive-terminal-file-manager)
   - [fd (PC-98 Filer Mode)](#fd-pc-98-filer-mode)
7. [System and Authentication Utilities](#7-system-and-authentication-utilities)
   - [setup (Quick Onboarding)](#setup-quick-onboarding)
   - [auth (OAuth Loopback)](#auth-oauth-loopback)
   - [status (Health Diagnostic)](#status-health-diagnostic)
8. [Conflict Resolution Guide](#8-conflict-resolution-guide)

---

## 1. General Command Layout

All `ggsrun` commands support the following global options to override configuration directories or point to explicit API credentials:

```bash
$ ggsrun <command> [options]
```

| Global Flag | Alias | Description |
| :--- | :--- | :--- |
| `--credentials <path>` | `--cred` | Absolute path to a custom Google Cloud credentials JSON file. |
| `--config <dir>` | `--conf` | Custom folder containing `ggsrun.cfg` (overrides default search priorities). |

---

## 2. Command Categorization

`ggsrun` categorizes its operations into distinct high-performance modules:

* **Execution Engine**: `exe1`, `exe2`, `webapps`
* **Concurreny Transfers**: `download`, `upload`, `updateproject`
* **Metadata & Quota**: `filelist`, `searchfiles`, `driveinformation`
* **Interactive TUI**: `fd`
* **Onboarding & Auth**: `setup`, `auth`, `status`

---

## 3. Executing Google Apps Script

### `exe1` (Stateful Script Execution)
* **Aliases**: `e1`
* **Purpose**: Synchronizes local scripts or directories to a remote GAS project on Google Drive and triggers a designated entry function.
* **Automation Safety**: If `--deleteScript` (`-d`) is supplied, all script files uploaded during execution are automatically and cleanly removed from the remote project once execution finishes.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--scriptid` | `-i` | String | Target Google Apps Script Project ID. |
| `--scriptfile` | `-s` | String | Path to local script file (`.gs`, `.js`) or directory. |
| `--stringscript`| `-ss` | String | Inline GAS script snippet provided as a raw string. |
| `--function` | `-f` | StringSlice | Repeats to pass function name and arguments sequentially. |
| `--deleteScript`| `-d` | Boolean | Safely auto-deletes uploaded files remotely post-execution. |
| `--conflict` | | String | Remote file conflict strategy: `overwrite` (default) or `add`. |
| `--jsonparser` | `-j` | Boolean | Mutes terminal UI spinners and returns pure JSON streams. |
| `--sandbox` | | String | Path to `sandbox_config.json` to sandbox APIs/URLs. Set to `bypass` to disable. |

#### Execution Recipes
* **Execute local script with sequential arguments**:
  ```bash
  $ ggsrun exe1 -i "SCRIPT_ID" -s "my_logic.js" -f "processData" -f "arg_val1" -f "arg_val2"
  ```
* **Recursively upload a local directory, run a function, and auto-cleanup**:
  ```bash
  $ ggsrun exe1 -i "SCRIPT_ID" -s "./src" -f "main" --deleteScript
  ```
* **Run inline script with default fallback script ID**:
  ```bash
  $ ggsrun exe1 -ss "function main() { return 'Hello!'; }" -f "main" -j
  ```

---

### `exe2` (Stateless API Execution)
* **Aliases**: `e2`
* **Purpose**: Transmits a local script payload inside a JSON-encoded string directly to the remote Google Apps Script `ExecutionApi` wrapper. Bypasses file updates completely.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--scriptid` | `-i` | String | Remote GAS Project ID containing the gateway server. |
| `--scriptfile` | `-s` | String | Path to local script file (Entry point MUST be named `main()`). |
| `--value` | `-v` | String | A raw string or JSON string argument passed into `main()`. |

#### Execution Recipes
* **Stateless execution passing a JSON payload**:
  ```bash
  $ ggsrun exe2 -i "SCRIPT_ID" -f ExecutionApi -s "extract.js" -v '{"limit":10}' -j
  ```

---

### `webapps` (Web App Execution)
* **Aliases**: `w`
* **Purpose**: Routes script payloads via HTTP POST directly to a deployed Google Web App gateway URL. Supports secure redirect authorization.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--webappurl` | `-u` | String | Target Web App URL deployment endpoint. |
| `--password` | `-p` | String | Security password parameter verifying execution on anonymous endpoints. |
| `--scriptfile` | `-s` | String | Local script file containing the entry `main()` function. |

#### Execution Recipes
* **Trigger secure Web App using OAuth loopback token**:
  ```bash
  $ ggsrun webapps -u "https://script.google.com/macros/s/XXX/exec" -s "job.js" -j
  ```

---

## 4. File and Directory Transfers

### `download` (Parallel Pull)
* **Aliases**: `d`
* **Purpose**: Recursively pulls files or entire directory trees from Google Drive utilizing concurrent channel workers.
* **Auto-Conversion**: Automatically exports native Google Workspace entities (Docs, Sheets, Slides) into user-defined formats (e.g., Markdown, XLSX, PDF) on-the-fly.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--scriptid` | `-i` | String | Comma-separated File or Folder IDs to download concurrently. |
| `--workers` | `-w` | Integer | Parallel network workers to spawn (Default: 5). |
| `--extension` | `-e` | String | Export conversion targets: `xlsx`, `docx`, `pdf`, `md`, etc. |
| `--conflict-mode`| `-cm` | String | Conflict resolution: `skip`, `overwrite`, `rename`, `update`. |
| `--destination`| `-d` | String | Destination folder directory path (Default: current directory). |
| `--zip` | `-z` | Boolean | Packages download files together into a clean local `.zip` file. |

#### Transfer Recipes
* **Download folder structure concurrently with conflict updates**:
  ```bash
  $ ggsrun download -i "FOLDER_ID" -w 10 -cm update -d "./local_sync"
  ```
* **Download and transpile Google Doc into Markdown**:
  ```bash
  $ ggsrun download -i "DOC_FILE_ID" -e md
  ```
* **Download GAS project as a packaged ZIP**:
  ```bash
  $ ggsrun download -i "SCRIPT_ID" -z
  ```

---

### `upload` (Parallel Push)
* **Aliases**: `u`
* **Purpose**: Recursively uploads local files or entire directories to Google Drive. Fully supports Resumable chunked uploading for massive binaries.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--filename` | `-f` | String | Comma-separated local file or directory paths to upload. |
| `--parentid` | `-p` | String | Target Google Drive Parent Folder ID. |
| `--convertto` | `-c` | String | Convert to Drive format: `sheet`, `doc`, `slide`. |
| `--conflict-mode`| `-cm` | String | Conflict resolution: `skip`, `overwrite`, `rename`, `update`. |
| `--projectname`| | String | Sets the remote script title when uploading script files standalone. |

#### Transfer Recipes
* **Upload local folder structure with 5 parallel workers**:
  ```bash
  $ ggsrun upload -f "./documents" -p "DRIVE_FOLDER_ID" -w 5 -cm rename
  ```
* **Upload CSV and automatically convert to Google Spreadsheet**:
  ```bash
  $ ggsrun upload -f "metrics.csv" -c sheet -p "DRIVE_FOLDER_ID"
  ```

---

### `updateproject` (Project Code Sync)
* **Aliases**: `ud`
* **Purpose**: Synchronizes local directory structures or file groups to an existing remote GAS project. 
* **User Safety**: Before updating, `ggsrun` lists all target modifications in a clean bullet list and requests hard terminal confirmation (Y/N) to protect remote files from accidental loss.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--scriptid` | `-p` | String | Target GAS Project Script ID. |
| `--filename` | `-f` | String | Local file or folder paths to synchronize. |
| `--deletefiles`| | Boolean | Deletes files from remote project if they are missing locally. |
| `--backup` | `-b` | Boolean | Downloads a local backup (`.zip` or `.json`) before applying updates. |

#### Sync Recipes
* **Sync local src folder recursively**:
  ```bash
  $ ggsrun updateproject -p "SCRIPT_ID" -f "./src"
  ```
* **Sync folder and delete omitted remote files with local backup**:
  ```bash
  $ ggsrun updateproject -p "SCRIPT_ID" -f "./src" --deletefiles -b
  ```

---

## 5. Drive Querying and Search

### `filelist` (List Files)
* **Aliases**: `ls`
* **Purpose**: Lists all files and folders available on Google Drive.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--limit` | | Integer | Max file count limit to output (Default: 100). |
| `--searchbyid` | | String | Performs direct metadata query lookup on a specific File ID. |

```bash
$ ggsrun filelist --limit 10
```

---

### `searchfiles` (Advanced Search)
* **Aliases**: `sf`
* **Purpose**: Searches your Google Drive using precise Google Drive API v3 query syntax and local filename Regex filtering. Supports Shared Drives seamlessly.

#### Command-specific Flags
| Flag | Shorthand | Type | Description |
| :--- | :--- | :--- | :--- |
| `--query` | `-q` | String | Google Drive API v3 standard query syntax. |
| `--regex` | `-r` | String | Regular expression to filter file name outcomes locally. |

#### Search Recipes
* **Search for non-trashed spreadsheets**:
  ```bash
  $ ggsrun searchfiles --query "mimeType = 'application/vnd.google-apps.spreadsheet' and trashed = false"
  ```
* **Search folders using Regex filter**:
  ```bash
  $ ggsrun searchfiles --query "mimeType = 'application/vnd.google-apps.folder'" --regex "^Test_.*"
  ```

---

## 6. Interactive Terminal File Manager

### `fd` (PC-98 Filer Mode)
* **Purpose**: Launches a dual-pane, split-screen terminal user interface (TUI) file manager inspired by the PC-98 Japanese classic. It bridges your local computer with Google Drive.
* **Key capabilities**: Copy (`F1`), Move (`F2`), and Delete (`F3`) files between panels. Search files (`F8`) recursively on local system or Drive-wide with yellow highlight overlays. Execute GAS scripts (`e` key) on focused files. Open files in a host web browser.

```bash
$ ggsrun fd
```

*For complete keybindings and navigation parameters, check [manual-tests/README.md](../manual-tests/README.md#keybindings-reference).*

---

## 7. System and Authentication Utilities

### `setup` (Quick Onboarding)
* **Purpose**: Automatically enables all 6 required Workspace APIs via Google Cloud Quick Flow redirect and initializes your default configuration (`ggsrun.cfg`) in seconds.

```bash
$ ggsrun setup
```

---

### `auth` (OAuth Loopback)
* **Purpose**: Performs local OAuth2 authorization loopback.

```bash
$ ggsrun auth --port 8080
```

---

### `status` (Health Diagnostic)
* **Aliases**: `st`
* **Purpose**: Quick diagnostic tool verifying the validity, expiry, and permissions of your local OAuth2 credentials and configuration paths.

```bash
$ ggsrun status
```

---

## 8. Conflict Resolution Guide

Commands such as `download` and `upload` handle file name collisions based on the `--conflict-mode` (`-cm`) flag:

| Mode | Download Action | Upload Action |
| :--- | :--- | :--- |
| **`skip`** | If local file exists, the download is skipped. | If remote file name exists, the upload is skipped. |
| **`overwrite`** | Overwrites the local file on disk. | Replaces the file content on Google Drive (`PATCH` request). |
| **`rename`** | Appends a timestamp (`_YYYYMMDD_HHMMSS`) locally. | Appends a timestamp (`_YYYYMMDD_HHMMSS`) to remote name. |
| **`update`** | Overwrites only if the remote file is newer. | Overwrites only if the local file is newer. |

If `-cm` is omitted, `ggsrun` presents an **interactive terminal selection prompt** allowing you to resolve collisions individually. When running in pure JSON mode (`-j`), the prompt is bypassed, defaulting to `update` (OverwriteIfNewer) to protect your pipeline.
