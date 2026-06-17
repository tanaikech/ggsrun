# ggsrun

<a name="top"></a>

# Update History

## ggsrun

- v1.0.0 (April 24, 2017)
  Initial release.

- v1.1.0 (April 28, 2017)
  1. Added a command for updating existing project on Google Drive. The detail information is [here](help/README.md#updateproject).
  2. Added "TotalElapsedTime" for Show File List and Search Files.

- v1.2.1 (May 28, 2017)
  1. ggsrun.cfg got be able to be read using the environment variable `GGSRUN_CFG_PATH`.

- v1.3.0 (August 30, 2017)
  1. From this version, [container-bound scripts](https://developers.google.com/apps-script/guides/bound) can be downloaded.

- v1.3.2 (October 20, 2017)
  1. From this version, scripts in a project can be rearranged interactively on your terminal and/or a configuration file.

- v1.3.3 (October 30, 2017)
  1. Added function to modify "Manifests" (appsscript.json).

- v1.4.0 (January 25, 2018)
  1. Integrated the official Google Apps Script API. Both standalone scripts and container-bound scripts can be rearranged and executed seamlessly.

- v1.4.1 (February 9, 2018)
  1. The resumable-upload method was added. Files are automatically uploaded in chunks.

- v1.5.0 (October 27, 2018)
  1. ggsrun got to be able to download all files and folders in the specific folder in Google Drive while maintaining the folder structure.

- v1.6.0 (November 30, 2018)
  1. Files got to be able to be searched using the search query and regex.

- v1.7.0 (December 27, 2018)
  1. Manage permissions of files.
  2. Get Drive Information.
  3. **ggsrun got to be able to be used by not only OAuth2, but also Service Account.**

- v2.0.0 (February 25, 2022)
  1. Modified using the latest libraries.

- v2.0.3 (June 14, 2025)
  1. Rebuild with go1.24.4.

- **v3.0.0 (May 2026) - Massive Concurrency Update**
  1. The core engine was completely rewritten for Go 1.26+.
  2. Integrated `golang.org/x/sync/errgroup` and `mpb/v8` for **true massively parallel concurrent** uploads and downloads.
  3. Deprecated the legacy Out-Of-Band (OOB) OAuth flow. Introduced an automatic local loopback listener that securely opens the browser and intercepts the auth token automatically.

- **v3.1.0 (May 2026) - Recursive Structure Update**
  1. Re-engineered the Drive file transfer logic. Uploads and downloads now recursively map deeply nested local/Drive folder structures.
  2. Introduced an ASCII visual tree mapping printout before transfers begin.

- **v3.2.0 (May 2026) - The AI/MCP Architecture Update**
  1. Transformed `ggsrun` into a background daemon capability.
  2. Redefined Config and Credentials paths priority (`--credentials`, `--config`, and `GGSRUN_CFG_PATH`).

- **v3.2.2 (May 2026) - Pure MCP Node Evolution**
  1. Finalized the `mcp` command. `ggsrun mcp` acts as a pure I/O backend for LLM clients (like Gemini CLI or Cursor) over stdio JSON-RPC. It intercepts tools calls (`searchfiles`, `download`, `upload`, `exe1`), executes them, and returns JSON without polluting standard output with human-readable logs.

- **v5.0.0 (May 2026) - The Omnibus Architecture Rewrite**
  1. Engine fundamentally rewritten for Go 1.26.3+. Implemented channel-based concurrency (`errgroup`), freeze-proof TUI (`mpb/v8`), SIMD JSON parsing (`goccy/go-json`), native MCP server (`ggsrun mcp`), Shared Drives full-support, auto MIME-mapping, isolated fault tolerance, and OAuth2 loopback automation.

- **v5.0.1 (May 2026) - Execution Engine Hardening & Double-Eval Eradication**
  1. Eliminated the V8 engine double-eval 500 server crash during dynamic script execution by enforcing IIFE and JSON-literal payload encoding.
  2. Redefined `-f` flag mapping for proper API gateway resolution in `exe2`.
  3. Added precision deployment documentation for stateful and stateless execution modes.

- **v5.0.2 (May 2026) - Secure Web Apps Protocol Upgrade**
  1. Upgraded the `webapps` command to natively support "Only myself" execution deployments by bridging OAuth tokens (`drive` scope) across Google's HTTP 302 Auth Redirects.
  2. Ported the IIFE/JSON-literal double-eval protections from `exe2` to `webapps`.

- **v5.0.3 (May 2026) - CLI UX Overhaul & Dynamic TUI Integration**
  1. Introduced a highly visual, modern Terminal UI (TUI) powered by `pterm` for `exe1`, `exe2`, and `webapps` commands.
  2. Added interactive loading spinners with anti-ghosting fixed-width padding (`%-70s`) and beautifully structured execution reports.
  3. Maintained strict backward compatibility by preserving pure JSON output streams via the `-j` flag for CI/CD pipeline automation.

- **v5.1.0 (May 2026) - Advanced Conflict Resolution Engine**
  1. Introduced a robust pre-computation conflict resolution matrix for both `download` and `upload` commands via the new `--conflict-mode` (`-cm`) flag. Users can now choose from `skip`, `overwrite`, `rename` (appends timestamp `_YYYYMMDD_HHMMSS` to avoid collisions), or `update` (syncs only if the source file is newer than the target).
  2. Includes interactive fallback CLI prompts if no mode is specified.
  3. Deprecated the legacy `--overwrite` (`-o`) and `--skip` (`-s`) options in favor of `--conflict-mode`.
  4. To avoid Drive API rate limits during massive concurrent uploads, metadata query is pre-fetched in bulk.

- **v5.1.1 (May 2026) - Modular Handlers & Enhanced MCP Server Core**
  1. Refactored the codebase to modularize legacy single-file command handlers into dedicated, organized handler files (`handler_download.go`, `handler_upload.go`, `handler_transfer.go`, `handler_mcp.go`, `handler_execute.go`).
  2. Strengthened the MCP server core (`ggsrun mcp`) by capturing stdout and stderr execution logs for comprehensive error recovery.
  3. Embedded full support for `--conflict-mode` inside the MCP JSON-RPC schemas and standardized file transfer outputs into `TransferResult` to support interactive multi-turn collision resolution in LLM conversations.
  4. Fully updated pre-built binaries for all major architectures.

- **v5.2.0 (June 2026) - Go standard layout, WSL2 browser integration, Web Apps URL registration, and CLI hardening**
  1. Reorganized the codebase to follow the standard Go project structure (`main.go`, `/internal/app/`, `/internal/utl/`).
  2. Expanded `ggsrun auth` to request Web Apps URL registration and dynamically persist it in `ggsrun.cfg`, allowing `ggsrun w` to run without the `-u` option.
  3. Integrated WSL 2 environment detection to prompt the user to choose between the Windows host browser (via `wslview` or `cmd.exe`), WSL/Ubuntu native browser, or manual URL copy-pasting.
  4. Upgraded `ggsrun e1`, `ggsrun e2`, and `ggsrun w` commands to dynamically print full CLI flag helps alongside custom usage examples when executed without arguments.

- **v5.2.1 (June 2026) - Dynamic CLI Help Customization, Beacon Script Integration, and Namespace Binding**
  1. Integrated comprehensive execution command examples (including stateless beacon checks) dynamically within both the `--help` flag screens and optionless execution error overlays for `e1`, `e2`, and `w` modes.
  2. Resolved a namespace issue where evaluated scripts calling `ggsrunif.Beacon()` inside the library threw a `ggsrunif is not defined` ReferenceError, by binding `ggsrunif` to the library's global execution context.

- **v5.2.2 (June 2026) - MCP Help Display Expansion, Safety Review Prompt, Dual-Mode Conflict Engine, and File-Level Error Feedback**
  1. Expanded `ggsrun mcp -h` (and `--help`) to display all exposed MCP tool names and their detailed description outputs directly.
  2. Implemented strict programmatic safety review prompts inside the `exe1` MCP tool description, instructing LLMs to statically analyze Apps Script payloads for API mutations (write/update/delete) and obtain user Y/N confirmations before running, while allowing read-only scripts to run automatically.
  3. Re-designed the conflict resolution engine into a dual-mode system:
     - For MCP server sessions (`GGSRUN_MCP_MODE=true` environment variable), conflict resolution is fully automated and non-interactive. Naming collisions default to `OverwriteIfNewer` (overwriting only if source timestamp is newer), with optional parameters for `Ignore` (unconditional skip) and `Rename` (auto-renaming with sequential numbers/timestamps).
     - For raw CLI sessions, the legacy v5.2.1 behavior is strictly preserved, prompting the user interactively (or returning pending status in JSON parser mode) upon name collisions.
  4. Enhanced file-level error feedback inside concurrent download and upload loops. Non-fatal transfer failures (e.g. API HTTP errors 400, 403, 404, or 429) no longer crash the queue but are returned with explicit error details inside the JSON `files` metadata array.
  5. Strictly adhered to Go 1.26 best practices: implemented context propagation to folders/files APIs and applied structured error wrapping via `fmt.Errorf` and `%w`.

- **v5.2.3 (June 2026) - Directory Reuse Conflict Resolution, Output Control, and CLI/MCP Alignment**
  1. Upgraded the directory upload conflict resolution mechanism: the tool now silently and recursively reuses existing remote folders (without prompting) while maintaining strict interactive conflict resolution only for individual files.
  2. Aligned `--conflict-mode` behavior for `-j` / `--jsonparser` CLI runs to match the automated, non-interactive MCP mode (defaulting to `OverwriteIfNewer`, but overridable using `--cm` or `--conflict-mode`).
  3. Hardened the output control engine for upload and download operations: when executing with the `-j` (`--jsonparser`) option, all human-readable TUI outputs (e.g. pterm logs, directory structure visual trees, success alerts) and concurrent progress bars (`mpb`) are completely suppressed, returning only clean JSON.
  4. Enabled `--cm` as a valid shorthand alias for `--conflict-mode` inside download and upload routines to ensure CLI parameter compatibility.
  5. Strictly adhered to Go 1.26 best-practice context propagation and wrapped error reporting.

- **v5.2.4 (June 2026) - Latest MIME Type Formats, CLI Option Help Details, Concurrent Conversion Overhaul, and Destination Directory Support**
  1. Updated internal MIME type mapping definitions (`extVsmime`, `googlemimetypes`, `defaultformat`, `mimeVsEx` in `googlemimetypes.go`) to synchronize with the latest Google Drive API `importFormats` and `exportFormats`.
  2. Revamped the CLI options help display for `--extension` (`-e` in download/revision commands) and `--convertto` (`-c` in upload command) to list all supported file formats, resolving ambiguity.
  3. Overhauled the concurrent upload engine (`handler_upload.go`): enabled parallel upload streams to handle `--convertto` and `--noconvert` directly without falling back to the legacy single-threaded uploader, and added robust warning feedback that skips unsupported conversions gracefully.
  4. Hardened the concurrent download engine (`handler_download.go`): integrated export capability validation via `IsExportable` and `ExtToMime` to verify file format compatibility before requesting Drive API `/export` downloads.
  5. Added the `--destination` (`-d`) option to the `download` and `revision` commands to allow specifying the target local directory for saving downloaded files, defaulting to the current directory.

- **v5.3.0 (June 2026) - Responsive TUI Filer (FD Mode) Enhancements, Focus Persistence, and Platform Compatibility Fixes**
  1. Refactored TUI Filer (FD Mode) popup layouts to be responsive. Custom-implemented 70% width centering using `tview.Flex` for error messages, sorting lists, text inputs, MIME conversions, help menu, and file details, preventing text clipping.
  2. Implemented focus persistence across filer operations: focus remains locked on the pre-action panel and table after file transfers, deletions, and GAS executions.
  3. Added wrap-around navigation to local and remote file tables.
  4. Mapped the `y` key to yank (copy) selected file absolute paths (local) or File IDs (remote) to the clipboard.
  5. Resolved cross-compilation errors on 32-bit Linux platforms (e.g., `linux/arm`) by explicitly casting `syscall.Stat_t` `Ctim` fields to `int64` inside platform-specific build files (`file_info_linux.go`, `file_info_darwin.go`).
  6. Updated the test suite (`fd_test.go`) to accommodate the new `TextView`-based popup structures.

- **v5.3.2 (June 2026) - Script Upload Flag Registration and TUI Focus Fallbacks**
  1. Fixed a TUI crash (`panic: internal 1`) on converting and uploading `.js`/`.gs` files to standalone Apps Script projects, caused by unregistered `"projectname"` and `"googledocname"` flags in `createOpContext` which led to empty title creation calls.
  2. Implemented remote text file previewing on Enter inside `ggsrun fd` (TUI), automatically downloading and showing the contents for MIME types starting with `text/` or matching JSON, XML, or JavaScript.
  3. Overhauled focus restoration inside `showTextPreview` to fall back to the global `lastActiveTable` variable when restoring focus, preventing focus from being lost to closed loading overlays.
  4. Replaced hardcoded conversion switch cases in `getConvertOptions` with dynamic checks calling `utl.GetImportTargets` to align convertible options with the official specification, automatically bypassing conversion prompts for unsupported types.

- **v5.3.1 (June 2026) - Script Upload Routing Fixes, Non-Convertible Upload Fallbacks, and TUI Error Propagation**
  1. Fixed a bug in `concurrentUpload` where uploading `.js`/`.gs`/`.gas` files without `--noconvert` attempted a resumable binary upload (resulting in HTTP 400 Bad Request); redirected these script uploads to the legacy script uploader (`p.Uploader`) to correctly create/update Google Apps Script projects.
  2. Overrode script source MIME type resolution to `text/plain` when uploading raw `.js`/`.gs`/`.gas` files as-is (with `--noconvert`) to prevent API errors.
  3. Resolved a bug where uploading files without Workspace conversion mappings (e.g., `.zip`, `.mp3`) were skipped from uploads by default; updated the conversion detection logic to upload them as-is (with no conversion) when no explicit conversion format is requested.
  4. Programmatically caught silent transfer failures in the TUI filer (`ggsrun fd`) by asserting and inspecting returned `TransferResult` and `FileInf` objects, correctly raising error alerts to the user rather than failing without reaction.

## Server

- v1.0.0 (April 24, 2017)
  Initial release.

- **v1.2.1 (June 2026) - V8 Modernization, Log Sheet Lazy Loading, and Namespace Scope Resolution**
  1. Refactored the core library script into an optimized V8 ES6 class structure.
  2. Added lazy-loading of log spreadsheets to bypass spreadsheet lookup overhead on non-logging runs (such as Beacon checks).
  3. Replaced deprecated `arguments.callee` with named recursive functions in `foldertree` and transitioned to the modern `File.moveTo` method for folder reorganization.
  4. Implemented flexible password verification that securely bypasses password checks when none is configured on the server, permitting seamless token-based execution.
  5. Bound `ggsrunif` globally to the library context to permit evaluated script payloads to call namespace alias methods safely.

**You can read "How to install" at [here](https://github.com/tanaikech/ggsrun/blob/master/README.md#howtoinstall).**

[TOP](#top)
