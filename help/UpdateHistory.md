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

- **v5.2.0 (June 2026) - Go standard layout, WSL2 browser integration, Web Apps URL registration, and CLI UX hardening**
  1. Reorganized the codebase to follow the standard Go project structure (`main.go`, `/internal/app/`, `/internal/utl/`).
  2. Expanded `ggsrun auth` to request Web Apps URL registration and dynamically persist it in `ggsrun.cfg`, allowing `ggsrun w` to run without the `-u` option.
  3. Integrated WSL 2 environment detection to prompt the user to choose between the Windows host browser (via `wslview` or `cmd.exe`), WSL/Ubuntu native browser, or manual URL copy-pasting.
  4. Upgraded `ggsrun e1`, `ggsrun e2`, and `ggsrun w` commands to dynamically print full CLI flag helps alongside custom usage examples when executed without arguments.

## Server

- v1.0.0 (April 24, 2017)

  Initial release.

**You can read "How to install" at [here](https://github.com/tanaikech/ggsrun/blob/master/README.md#howtoinstall).**

[TOP](#top)
