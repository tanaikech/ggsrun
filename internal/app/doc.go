/*
Package main (doc.go) :
This is a modern, highly-concurrent CLI tool to execute Google Apps Script (GAS) on a terminal and manage Google Drive infrastructure.

# Architecture Overhaul (v5.2.3 - Go 1.26+)

The core engine of `ggsrun` has been upgraded to include advanced self-healing features, expanded execution mechanics, and refined security/TUI flows:

1. Self-Healing & Auto-Deployments:
   Before executing scripts, `ggsrun` now inspects the remote project's `appsscript.json` manifest.
   - For `e1`/`e2` modes, it ensures the "executionApi" field is present (defaulting to "MYSELF").
   - For `e2`/`w` modes, it verifies the dependency configuration for the `ggsrunif` library.
   - For `w` (Web Apps) mode, it ensures the "webapp" configurations are defined.
   If any definitions are missing, they are dynamically injected, a new version is compiled, and the project is automatically re-deployed.

2. Auto-Recovery for Entry Functions (Self-Healing e2):
   If `e2` executes and encounters a missing target function error (e.g., "Script function not found: ExecutionApi"), the tool automatically uploads and deploys a helper script (`ggsrun_api_helper.gs`) defining the wrapper (`const ExecutionApi = e => ggsrunif.ExecutionApi(e);`) and retries the execution.

3. Interactive Setup Overhaul (ggsrun auth):
   The authorization command features a step-by-step setup:
   - Explicitly displays absolute paths of `ggsrun.cfg` and `client_secret.json`.
   - Offers customization of the configuration save path, with environmental mismatch warnings.
   - Prompts for pre-registration of Script IDs.
   - Displays all client info and requested OAuth scopes before launching the web server.

4. Expanded File Support:
   Local script parsing now supports `.txt` files in addition to `.gs` and `.js`.

5. Inline Scripts & Stdin Pipes:
   The `-ss` / `--stringscript` option and piped standard input (`cat script.js | ggsrun e1`) are now supported across `e1`, `e2`, and `w` subcommands. Temporal file uploads in `e1` and `w` are automatically cleaned up on completion.

6. Enhanced Config Path Visibility:
   Loads of `ggsrun.cfg` are reported on standard output. If `-j` (JSON mode) is specified, the absolute config path is injected into the output JSON top-level under the "config_path" key.

7. MCP Server Schema, Safety Reviews, & Mapping Enhancements:
   The `ggsrun mcp` command now features improved JSON schemas with comprehensive description fields and Drive API syntax examples.
   - The `exe1` tool schema adds `scriptfile` and `stringscript` properties to allow dynamic source execution, while `scriptid` has been made optional by automatically resolving the 'script_id' from the configuration file `ggsrun.cfg`.
   - Critical Safety static-analysis and user Y/N confirmation rules have been integrated into the `exe1` tool's description to guide LLM agents safely before script execution on write/update/delete operations.
   - Help command `ggsrun mcp -h` has been expanded to display all exposed MCP tool names and descriptions directly.

8. Redesigned Download & Upload Conflict-Mode Rules (v5.2.3):
   The file transfer conflict resolution engine has been upgraded to support recursive folder/directory conflict resolution along with individual files:
   - For MCP server sessions (triggered via `GGSRUN_MCP_MODE=true` environment variable), conflict resolution is fully automated and non-interactive:
     - Default behavior: Automatically checks file timestamps and overwrites only if the source file/folder is newer (`OverwriteIfNewer`).
     - Options: `OverwriteIfNewer` (overwrite only if newer), `Ignore` (skip completely on name collision), and `Rename` (auto-rename with sequential numbers/timestamps).
   - For raw CLI sessions, the legacy v5.2.1 behavior is strictly preserved for both files and directories, prompting the user interactively (or returning pending status in JSON parser mode) upon name collisions.

# Features of "ggsrun" are as follows.

1. Develops GAS using your terminal and text editor (supporting .gs, .js, and .txt).
2. Executes GAS by giving values to your script.
3. Downloads files concurrently from Google Drive with stunning progress visualizations.
4. Uploads files concurrently to Google Drive via native Resumable upload wrappers.
5. Downloads standalone script and bound script.
6. Downloads all files and folders in a specific folder.
7. Upload script files and create projects as standalone scripts and container-bound scripts.
8. Manage Permissions of files.
9. Search files in Google Drive using search query and regex.
10. ggsrun supports both robust OAuth2 looping and Service Accounts natively.

# How to Execute Google Apps Script Using ggsrun

If you do not have an active OAuth session, run the automated auth flow:
$ ggsrun auth

Execute your script with Execution API (e1):
$ ggsrun e1 -s sample.gs

Execute a script directly without updating project (e2):
$ ggsrun e2 -s sample.gs

Execute via Web Apps (w):
$ ggsrun w -u [WEB_APP_URL] -s sample.gs
*/
package app
