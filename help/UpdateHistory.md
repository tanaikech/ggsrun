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

## Server

- v1.0.0 (April 24, 2017)

  Initial release.

**You can read "How to install" at [here](https://github.com/tanaikech/ggsrun/blob/master/README.md#howtoinstall).**

[TOP](#top)
