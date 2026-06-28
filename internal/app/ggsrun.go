// Package main (ggsrun.go) :
// This file is included all commands and options.
package app

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

func init() {
	// Strictly bind TUI and human-readable output to Stderr.
	// This preserves Stdout for pure JSON and MCP JSON-RPC streams.
	pterm.SetDefaultOutput(os.Stderr)
}

func getCommonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "credentials, cred",
			Usage: "Path to a custom credentials file (e.g., /path/to/client_secret.json). Sets the default ggsrun.cfg location to match.",
		},
		&cli.StringFlag{
			Name:  "config, conf",
			Usage: "Path to the directory containing ggsrun.cfg. Overrides all other path priorities.",
		},
	}
}

// main : main function
func Run() {
	app := cli.NewApp()
	app.Name = appname
	app.Authors = []cli.Author{
		{Name: "Tanaike [ https://github.com/tanaikech/ggsrun ] ", Email: "tanaike@hotmail.com"},
	}
	app.UsageText = "This is a CLI application for managing Google Drive and Google Apps Script (GAS). Powered by modern Go concurrency."
	app.Version = "5.3.9"
	app.Commands = []cli.Command{
		{
			Name:        "exe1",
			Aliases:     []string{"e1"},
			Usage:       "Updates the GAS project with a local script or directory and executes a specified function.",
			Description: "Requires an access token. Synchronizes local scripts or directories to the Drive project and triggers the execution API.\n\nAPI & URL Sandboxing (Default: STRICT BLOCK ALL):\n  To restrict Google APIs and external URL requests during execution, use the '--sandbox' option.\n  - Default behavior (option omitted or '--sandbox \"\"'): Applies an ultra-strict sandboxing with empty whitelists, blocking all external requests (Drive, Mail, URL fetch, etc.).\n  - To specify a custom whitelist: Pass the path to a JSON configuration file (e.g. '--sandbox sandbox_config.json').\n  - To completely disable sandboxing: Pass '--sandbox bypass' or '--sandbox none'.\n\n  The sandbox configuration JSON structure must follow this format:\n  {\n    \"allowedFileIds\": [\"id1\"],        // Whitelisted Drive file IDs\n    \"allowedFolderIds\": [\"id2\"],      // Whitelisted Drive folder IDs\n    \"allowedCalendarIds\": [\"primary\"],// Whitelisted Calendar IDs\n    \"allowedEventIds\": [],            // Whitelisted Calendar Event IDs\n    \"allowedEmails\": [\"a@b.com\"],     // Whitelisted email recipients for Mail/Gmail\n    \"allowedUrls\": [\"https://*\"],     // Whitelisted URL patterns (supports '*' wildcard)\n    \"blockedUrls\": [\"https://xxx\"]    // Blacklisted URL patterns (highest priority)\n  }\n\nUsage Examples:\n  1. Execute a local script file with sequential arguments:\n     ggsrun e1 -i [SCRIPT_ID] -f myFunction -f \"arg1\" -f \"arg2\"\n\n  2. Execute a local directory with automated cleanup:\n     ggsrun e1 -i [SCRIPT_ID] -s path/to/dir -f myFunction -d\n\n  3. Execute an inline script:\n     ggsrun e1 -i [SCRIPT_ID] -ss \"function main() { return 'hello'; }\"\n\n  4. Execute via standard input (pipe):\n     cat script.js | ggsrun e1 -i [SCRIPT_ID]\n\n  5. Run and backup the project before updating:\n     ggsrun e1 -i [SCRIPT_ID] -s script.gs -b\n\n  6. Execute a stateless beacon request:\n     ggsrun e1 -ss \"const main = (_) => ggsrunif.Beacon();\" -j",
			Action:      exeAPIWithout,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "scriptid, i",
					Usage: "Script ID of project on Google Drive",
				},
				&cli.StringFlag{
					Name:  "scriptfile, s",
					Usage: "GAS file (.gs, .gas, .js, .txt, .coffee) or directory path on local PC",
				},
				&cli.StringFlag{
					Name:  "stringscript, ss",
					Usage: "GAS script provided directly as strings.",
				},
				&cli.StringSliceFlag{
					Name:  "function, f",
					Usage: "Function name which is executed. Subsequent '-f' flags represent arguments sequentially. If script ID (-i) is omitted but -f is provided, the script_id from ggsrun.cfg will be used. Default is '" + deffuncwithout + "'.",
				},
				&cli.StringFlag{
					Name:  "value, v",
					Usage: "Give a value to the function which is executed. (Fallback option if subsequent '-f' is not used)",
				},
				&cli.BoolFlag{
					Name:  "deleteScript, d",
					Usage: "Automatically and safely delete uploaded files from the remote GAS project after execution completes. (Strictly for exe1 only)",
				},
				&cli.StringFlag{
					Name:  "conflict",
					Usage: "Conflict resolution strategy when duplicate script name exists: 'overwrite' or 'add'.",
				},
				&cli.BoolFlag{
					Name:  "backup, b",
					Usage: "Backup project with script ID you set as a local file.",
				},
				&cli.BoolFlag{
					Name:  "onlyresult, r",
					Usage: "Display only 'result' field in JSON outputs.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "sandbox",
					Usage: "Path to a configuration JSON file to control API and URL sandboxing. Default (empty) applies a strict sandbox (BLOCK ALL). Set to 'bypass' or 'none' to disable sandboxing.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "exe2",
			Aliases:     []string{"e2"},
			Usage:       "Executes a GAS script directly via the Execution API without updating the project files.",
			Description: "Requires an access token. The script string is uploaded as a payload to the server-side GAS library for execution.\n\nUsage Examples:\n  1. Execute a local script file:\n     ggsrun e2 -i [SCRIPT_ID] -s path/to/script.gs -v \"hello\"\n\n  2. Execute an inline script:\n     ggsrun e2 -i [SCRIPT_ID] -ss \"function main() { return 'hello'; }\"\n\n  3. Execute via standard input (pipe):\n     cat script.js | ggsrun e2 -i [SCRIPT_ID]\n\n  4. Get a folder tree of Google Drive:\n     ggsrun e2 -i [SCRIPT_ID] -t\n\n  5. Download a file using byte slice conversion:\n     ggsrun e2 -i [SCRIPT_ID] -conv -v [FILE_ID]\n\n  6. Execute a stateless beacon request:\n     ggsrun e2 -ss \"const main = (_) => ggsrunif.Beacon();\" -j",
			Action:      exeAPIWith,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "scriptid, i",
					Usage: "Script ID of project on Google Drive",
				},
				&cli.StringFlag{
					Name:  "scriptfile, s",
					Usage: "GAS file (.gs, .gas, .js, .txt, .coffee) on local PC",
				},
				&cli.StringFlag{
					Name:  "function, f",
					Usage: "Function name of server for executing GAS. Default is '" + deffuncserv + "'.",
				},
				&cli.StringFlag{
					Name:  "value, v",
					Usage: "Give a value to the function of GAS script which is executed.",
				},
				&cli.StringFlag{
					Name:  "stringscript, ss",
					Usage: "GAS script provided directly as strings.",
				},
				&cli.BoolFlag{
					Name:  "foldertree, t",
					Usage: "Display a folder tree on Google Drive as a structured array.",
				},
				&cli.BoolFlag{
					Name:  "convert, conv",
					Usage: "[Experiment] Download file using byte slice data. Use with '-v [File ID]'.",
				},
				&cli.BoolFlag{
					Name:  "log, l",
					Usage: "Record access log.",
				},
				&cli.BoolFlag{
					Name:  "onlyresult, r",
					Usage: "Display only 'result' in JSON outputs.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "webapps",
			Aliases:     []string{"w"},
			Usage:       "Executes a GAS script anonymously or securely via Web Apps URL.",
			Description: "No access token is required. Operates directly via POST requests to a deployed Web App.\n\nUsage Examples:\n  1. Execute a local script file:\n     ggsrun w -u [WEB_APPS_URL] -s path/to/script.gs -p password -v \"hello\"\n\n  2. Execute an inline script:\n     ggsrun w -u [WEB_APPS_URL] -ss \"function main() { return 'hello'; }\"\n\n  3. Execute via standard input (pipe):\n     cat script.js | ggsrun w -u [WEB_APPS_URL]\n\n  4. Execute a stateless beacon request:\n     ggsrun w -ss \"const main = (_) => ggsrunif.Beacon();\" -j",
			Action:      webAppsWith,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "url, u",
					Usage: "Deployed Web Apps URL.",
				},
				&cli.StringFlag{
					Name:  "scriptfile, s",
					Usage: "GAS file (.gs, .gas, .js, .txt, .coffee) on local PC",
				},
				&cli.StringFlag{
					Name:  "scriptid, i",
					Usage: "Script ID of project on Google Drive",
				},
				&cli.StringFlag{
					Name:  "stringscript, ss",
					Usage: "GAS script provided directly as strings.",
				},
				&cli.StringFlag{
					Name:  "value, v",
					Usage: "Give a value to the function of GAS script which is executed.",
				},
				&cli.StringFlag{
					Name:  "password, p",
					Usage: "Password to authenticate to the Web Apps (if configured).",
				},
				&cli.BoolFlag{
					Name:  "log, l",
					Usage: "Do not record access log. (Omitting this flag means 'Record log').",
				},
				&cli.BoolFlag{
					Name:  "onlyresult, r",
					Usage: "Display only 'result' in JSON outputs.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "download",
			Aliases:     []string{"d"},
			Usage:       "Concurrently downloads files from Google Drive with high-speed parallelism.",
			Description: "Requires an access token. Supports massive concurrent downloads and progressive bars.",
			Action:      downloadFiles,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "fileid, i",
					Usage: "File ID on Google Drive. Comma separated for parallel concurrent downloads.",
				},
				&cli.StringFlag{
					Name:  "filename, f",
					Usage: "File Name to save as locally.",
				},
				&cli.IntFlag{
					Name:  "workers, w",
					Usage: "Number of concurrent workers for parallel downloads.",
					Value: 5,
				},
				&cli.StringFlag{
					Name:  "extension, e",
					Usage: "Extension format for exported files. Supported: docx, pdf, rtf, html, txt, md, odt, epub, zip (Google Docs); xlsx, ods, csv, tsv, pdf, zip (Google Sheets); pptx, pdf, odp, txt (Google Slides); svg, png, pdf, jpeg (Google Drawings); mp4 (Google Video); png, jpeg (Google Photos/Pix).",
				},
				&cli.StringFlag{
					Name:  "destination, d",
					Usage: "Target local destination directory path for downloaded files (defaults to current directory).",
				},
				&cli.StringFlag{
					Name:  "mimetype, m",
					Usage: "mimeType filter for downloading from folders. ex: '-m \"mimeType1,mimeType2\"'",
				},
				&cli.StringFlag{
					Name:  "conflict-mode, cm",
					Usage: "Action on conflict: OverwriteIfNewer, Ignore, Rename. (MCP Mode Default: OverwriteIfNewer, CLI Mode Default: Interactive).",
				},
				&cli.BoolFlag{
					Name:  "rawdata, r",
					Usage: "Save a GAS project as raw JSON data.",
				},
				&cli.BoolFlag{
					Name:  "zip, z",
					Usage: "Create a zip archive comprising all scripts of a GAS project.",
				},
				&cli.StringFlag{
					Name:  "deletefile",
					Usage: "Delete a file permanently using a specified file ID.",
				},
				&cli.BoolFlag{
					Name:  "overwrite, o",
					Usage: "Overwrite local files if they already exist (Deprecated: use --conflict-mode=overwrite).",
				},
				&cli.BoolFlag{
					Name:  "skip, s",
					Usage: "Skip downloading files if they already exist locally (Deprecated: use --conflict-mode=skip).",
				},
				&cli.BoolFlag{
					Name:  "showfilelist, l",
					Usage: "Retrieve list of files instead of downloading them.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "upload",
			Aliases:     []string{"u"},
			Usage:       "Concurrently uploads files to Google Drive using resumable chunking.",
			Description: "Requires an access token. Accelerates bulk transfers using Go concurrency.",
			Action:      uploadFiles,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "filename, f",
					Usage: "Local file paths. Comma separated for concurrent parallel uploads.",
				},
				&cli.StringFlag{
					Name:  "parentfolderid, p",
					Usage: "Destination folder ID on Google Drive.",
				},
				&cli.IntFlag{
					Name:  "workers, w",
					Usage: "Number of concurrent workers for parallel uploads.",
					Value: 5,
				},
				&cli.StringFlag{
					Name:  "conflict-mode, cm",
					Usage: "Action on conflict: OverwriteIfNewer, Ignore, Rename. (MCP Mode Default: OverwriteIfNewer, CLI Mode Default: Interactive).",
				},
				&cli.StringFlag{
					Name:  "parentid, pid",
					Usage: "File ID of a Google Doc/Sheet for creating container-bound scripts.",
				},
				&cli.StringFlag{
					Name:  "timezone, tz",
					Usage: "Timezone configuration for newly created projects.",
				},
				&cli.StringFlag{
					Name:  "projectname, pn",
					Usage: "Target project name for uploaded GAS scripts.",
				},
				&cli.StringFlag{
					Name:  "googledocname, gn",
					Usage: "File name for newly created Google Docs.",
				},
				&cli.StringFlag{
					Name:  "projecttype, pt",
					Usage: "Type of bound-script. ('spreadsheet', 'document', 'slide', 'form'). Default is 'standalone'.",
					Value: "standalone",
				},
				&cli.Int64Flag{
					Name:  "chunksize, chunk",
					Usage: "Resumable upload chunk size (MB).",
					Value: 100,
				},
				&cli.StringFlag{
					Name:  "convertto, c",
					Usage: "Convert files automatically to Google Workspace formats. Supported: 'doc' (or 'document', 'docs' for Google Docs), 'sheet' (or 'spreadsheet', 'spread' for Google Sheets), 'slide' (or 'slides', 'presentation' for Google Slides). If not specified, file extensions are automatically mapped (e.g., docx/rtf/html/txt/md/pdf/png/jpeg/bmp/gif -> Google Docs; xlsx/xls/csv/tsv -> Google Sheets; pptx/ppt/odp -> Google Slides; mp4/ogg/mov/webm -> Google Video). Use --noconvert to skip conversion.",
				},
				&cli.BoolFlag{
					Name:  "noconvert, nc",
					Usage: "Bypass Google Apps format conversion.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "updateproject",
			Aliases:     []string{"ud"},
			Usage:       "Synchronizes local files to an existing GAS project.",
			Description: "Updates, adds, or removes files in an active GAS project workspace.",
			Action:      updateProject,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "filename, f",
					Usage: "Local source file names. Overwrites identical remote scripts.",
				},
				&cli.BoolFlag{
					Name:  "deletefiles",
					Usage: "Delete specified filenames from the remote project.",
				},
				&cli.StringFlag{
					Name:  "projectid, p",
					Usage: "Target destination Project ID.",
				},
				&cli.BoolFlag{
					Name:  "backup, b",
					Usage: "Generate a local backup of the project prior to updating.",
				},
				&cli.BoolFlag{
					Name:  "rearrange, r",
					Usage: "Interactively rearrange project scripts via terminal.",
				},
				&cli.StringFlag{
					Name:  "rearrangewithfile, rf",
					Usage: "Apply layout configurations from a specified template file.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "conflict",
					Usage: "Conflict resolution strategy when duplicate script name exists: 'overwrite' or 'add'.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "revisionfiles",
			Aliases:     []string{"r"},
			Usage:       "Retrieves revision history and downloads specific file revisions.",
			Description: "Requires an access token. Accesses historical version data of Google Drive documents.",
			Action:      revisionFiles,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "fileid, i",
					Usage: "File ID to inspect revisions.",
				},
				&cli.StringFlag{
					Name:  "download, d",
					Usage: "Specific Revision ID to download.",
				},
				&cli.StringFlag{
					Name:  "createversion, cv",
					Usage: "Generate a new version tag with a custom description.",
				},
				&cli.StringFlag{
					Name:  "extension, e",
					Usage: "Extension format for exported files. Supported: docx, pdf, rtf, html, txt, md, odt, epub, zip (Google Docs); xlsx, ods, csv, tsv, pdf, zip (Google Sheets); pptx, pdf, odp, txt (Google Slides); svg, png, pdf, jpeg (Google Drawings); mp4 (Google Video); png, jpeg (Google Photos/Pix).",
				},
				&cli.StringFlag{
					Name:  "destination, d",
					Usage: "Target local destination directory path for downloaded files (defaults to current directory).",
				},
				&cli.BoolFlag{
					Name:  "rawdata, r",
					Usage: "Save a project as raw JSON data.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "filelist",
			Aliases:     []string{"ls"},
			Usage:       "Queries and outputs a detailed list of files from Google Drive.",
			Description: "Performs exhaustive listings mapped by ID or name search.",
			Action:      showFileList,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "searchbyname, sn",
					Usage: "Search by File Name. Returns matching File IDs.",
				},
				&cli.StringFlag{
					Name:  "searchbyid, si",
					Usage: "Search by File ID. Returns the File Name.",
				},
				&cli.BoolFlag{
					Name:  "stdout, s",
					Usage: "Stream file list directly to standard output.",
				},
				&cli.BoolFlag{
					Name:  "file, f",
					Usage: "Save file list to a local JSON payload file.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "searchfiles",
			Aliases:     []string{"sf"},
			Usage:       "Searches Google Drive utilizing precise queries and regex matching.",
			Description: "Advanced Drive metadata retrieval powered by robust regex validation.",
			Action:      searchFilesByQueryAndRegex,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "query, q",
					Usage: "Advanced Search Query parameter.",
				},
				&cli.StringFlag{
					Name:  "fields, f",
					Usage: "Filter return fields.",
				},
				&cli.StringFlag{
					Name:  "regex, r",
					Usage: "Regex pattern for evaluating filenames natively.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "permissions",
			Aliases:     []string{"p"},
			Usage:       "Manages file permissions, ownership transfers, and access roles.",
			Description: "Modify, audit, or transfer security access to Drive resources.",
			Action:      managePermissions,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "fileid, fi",
					Usage: "Required. Target File ID.",
				},
				&cli.StringFlag{
					Name:  "permissionid, pi",
					Usage: "Specific permission ID to manipulate.",
				},
				&cli.BoolFlag{
					Name:  "create, c",
					Usage: "Provision new permission access.",
				},
				&cli.BoolFlag{
					Name:  "delete, d",
					Usage: "Revoke an existing permission token.",
				},
				&cli.StringFlag{
					Name:  "role",
					Usage: "Access role: owner, organizer, fileOrganizer, writer, commenter, reader.",
				},
				&cli.StringFlag{
					Name:  "type",
					Usage: "Grantee type: user, group, domain, anyone.",
				},
				&cli.StringFlag{
					Name:  "emailaddress, email",
					Usage: "Associated email address for permission binding.",
				},
				&cli.BoolFlag{
					Name:  "transferownership, transfer",
					Usage: "Forcefully transfer ownership. The current owner is degraded to writer.",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "driveinformation",
			Aliases:     []string{"di"},
			Usage:       "Retrieves quota and user information for the authenticated Drive.",
			Description: "Provides raw Drive telemetry and usage metadata.",
			Action:      getDriveInformation,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "fields, f",
					Usage: "Fields to interrogate.",
					Value: "storageQuota,user",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
				&cli.StringFlag{
					Name:  "serviceaccount, sa",
					Usage: "Path to a service account credentials.json file.",
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "auth",
			Usage:       "Initiates the OAuth2 authorization flow.",
			Description: "Automatically launches secure browser auth flow and stores tokens based on hierarchy configuration.",
			Action:      reAuth,
			Flags: append([]cli.Flag{
				&cli.IntFlag{
					Name:  "port, p",
					Usage: "Port binding for temporary OAuth loopback web server.",
					Value: 8080,
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "setup",
			Usage:       "Initiates the simplified quick-setup onboarding flow.",
			Description: "Guides you through quick setup by automatically opening a customized Google API Quick Flow link, assisting credentials registration, and launching the OAuth2 authorization process.",
			Action:      quickSetup,
			Flags: append([]cli.Flag{
				&cli.IntFlag{
					Name:  "port, p",
					Usage: "Port binding for temporary OAuth loopback web server.",
					Value: 8080,
				},
			}, getCommonFlags()...),
		},
		{
			Name:        "status",
			Aliases:     []string{"st"},
			Usage:       "Checks authentication status and API connectivity.",
			Description: "Quick health diagnostic tool for tokens and environment resolution.",
			Action:      checkStatus,
			Flags:       getCommonFlags(),
		},
		{
			Name:        "fd",
			Usage:       "Launches the interactive split-screen TUI file manager.",
			Description: "Launches the PC-98 style split-screen TUI file manager bridging local system and Google Drive.",
			Action: func(c *cli.Context) error {
				if RunTUIFunc != nil {
					return RunTUIFunc(c)
				}
				pterm.Error.Println("TUI module not initialized.")
				return nil
			},
			Flags: getCommonFlags(),
		},
		{
			Name:        "mcp",
			Aliases:     []string{"m"},
			Usage:       "Starts ggsrun as an MCP (Model Context Protocol) server for LLM clients.",
			Description: "Runs a stdio listener providing Drive/GAS tools to an MCP client. Does not require LLM API keys natively.\n\nExposed MCP Tools:\n  - searchfiles: Search Google Drive files using standard Google Drive API v3 query syntax. Supports optional regex filtering on filenames.\n  - download: Download files or folder structures from Google Drive to the local environment using File/Folder IDs.\n  - upload: Upload local files or entire recursive directories to Google Drive.\n  - exe1: Upload/synchronize a local Google Apps Script file or raw script string to a remote Google Apps Script project, and execute a specified entry function. Returns the function execution response payload as JSON.\n  - filelist: List files or search by exact File Name or File ID on Google Drive. Outputs corresponding file details.",
			Action:      runMCP,
			Flags:       getCommonFlags(),
		},
		{
			Name:        "recover",
			Aliases:     []string{"rc"},
			Usage:       "Restores the GAS project to the pristine recovery state.",
			Description: "Cleans up the remote GAS project by restoring it to the default recovery files (ggsrun.gs and appsscript.json).",
			Action:      recoverProject,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:  "scriptid, i",
					Usage: "Script ID of project on Google Drive",
				},
				&cli.BoolFlag{
					Name:  "jsonparser, j",
					Usage: "Bypass TUI and display outputs strictly as pure JSON.",
				},
			}, getCommonFlags()...),
		},
	}
	app.CommandNotFound = commandNotFound
	app.Run(os.Args)
}
