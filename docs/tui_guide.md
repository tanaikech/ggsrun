# ggsrun - Interactive TUI Filer (FD Mode) Guide

This guide provides a comprehensive specification, behavioral insights, development history, and usage instructions for the Interactive Terminal User Interface (TUI) File Manager (`ggsrun fd`), affectionately known as **FD Mode**.

---

## Table of Contents
1. [Overview & Visual Aesthetics](#1-overview--visual-aesthetics)
2. [Motivation & Inspiration](#2-motivation--inspiration)
3. [The Unified Developer Prompt](#3-the-unified-developer-prompt)
4. [Development & Release History (v5.3.3)](#4-development--release-history-v533)
   - [📊 Consumed Resources](#-consumed-resources)
   - [💡 Efficiency & Success Review](#-efficiency--success-review)
   - [🛠️ Key Improvements & Hardening](#-key-improvements--hardening)
5. [How to Launch](#5-how-to-launch)
6. [Interactive Controls & Keybindings Reference](#6-interactive-controls--keybindings-reference)
7. [Premium TUI Features (v5.3.6+)](#7-premium-tui-features-v536)
   - [Function Key Actions](#function-key-actions)
   - [Recursive & Drive-Wide Search (`F8`)](#recursive--drive-wide-search-f8)
   - [Yellow Search Highlighting](#ui-highlighting)
   - [Web View Link Integration (`i` details)](#web-view-link-details)
   - [Directory Tree Preview](#directory-tree-preview)
   - [Real-Time Individual Progress Bars](#real-time-individual-progress-bars)

---

## 1. Overview & Visual Aesthetics

The FD Mode of `ggsrun` launches an immersive, dual-pane Terminal User Interface (TUI) that bridges your local computer filesystem (upper panel) and Google Drive (lower panel) side-by-side. 

With this interface, you can perform high-speed, concurrent file transfers, organize folder hierarchies, rename assets, preview text files, convert MIME formats, and even statefully execute Google Apps Script code directly within your console.

---

## 2. Motivation & Inspiration

The development of the FD Mode (TUI Filer) in `ggsrun` was driven by two central ideas:

1. **Nostalgia & High-Velocity Utility**: Drawing deep inspiration from the classic **"FD" filer** software widely used on Japanese PC platforms (such as the NEC PC-9801 series) during the late 1980s and 1990s. Dual-pane, keyboard-driven file managers are famously efficient. Bringing this lightweight, responsive local-to-cloud filer to `ggsrun` dramatically improves developer productivity compared to clicking through the standard web Google Drive UI.
2. **AI Capability Verification**: Serving as a rigorous, real-world benchmark to evaluate and showcase the agentic engineering capabilities of the **Antigravity CLI (AI Coding Assistant)**. Engineering an interactive, cross-platform TUI with complex focus persistence, mock event-simulation testing (`tcell.SimulationScreen`), and multi-architecture build safety represents a premier challenge, proving that AI agents can build production-grade, highly reliable system software autonomously.

---

## 3. The Unified Developer Prompt

The entire TUI filer feature set, layout refactoring, clipboard synchronization, and platform-specific compilation parameters were implemented autonomously by the AI based on the following unified developer prompt:

```markdown
# Goal

Implement a dual-pane Terminal User Interface (TUI) File Manager (referred to as "FD mode") for the `ggsrun` Go application. This mode must allow users to manage local files (upper panel) and Google Drive files (lower panel) side-by-side, perform file transfers/operations, and execute Google Apps Script (GAS) directly from the interface. The codebase must compile across all target platforms (Linux 64/32bit/ARM, macOS, Windows) and include comprehensive unit tests.

---

## 1. UI Architecture & Responsive Layout

- **Dual-Pane Layout**: Split the main screen vertically or horizontally using `tview` (upper panel for the local filesystem, lower panel for Google Drive).
- **Status Bar**: Display keybindings and status info at the bottom.
- **Visual Distinction**:
  - Differentiate directories and files by color (e.g., color folders green/yellow).
  - Render file details columns clearly, including Name, Size, Modified Date, and Permissions (human-readable system permissions for local, owner/sharing state for Google Drive).
  - Ensure column values (like directory name or file ID) do not disappear or glitch during scroll navigation.
- **70% Responsive Dialogs & Popups**:
  - All popup windows (errors, execution prompts, sorting selection, conversion prompts, operation help, file details, and execution results) must occupy exactly 70% of the terminal width (using a `tview.Flex` layout with 15% margins on both sides).
  - For long-text dialogs like "File Details" (`showFileDetails`) and "Error Messages" (`showError`), use a scrollable `tview.TextView` inside the 70% container instead of standard fixed-width modals, ensuring no content clips.
  - In `showExecutionResult`, center the execution logs inside a 70% width and 70% height responsive popup window.
  - For text input prompts (`promptTextInput`), set the input field width to fill the dialog width (`SetFieldWidth(0)`).

---

## 2. Navigation & File Operations

- **Keyboard Shortcuts**:
  - `Tab`: Switch focus between the Local and Remote panels.
  - `Up` / `Down`: Navigate the file list. Add **Wrap-around** logic (cursor wraps to the top when going past the last item, and to the bottom when going past the first).
  - `Space`: Toggle multi-selection.
  - `Enter`: Enter directories, preview text files (local), open non-script files in a browser, or open script explorers (remote).
  - `F5` / `F6` / `F8`: Copy, Move, or Delete selected items between panels.
  - `c` / `m`: Copy or Move items within the same panel.
  - `n`: Rename file/folder.
  - `t`: Edit last modified timestamp.
  - `d`: Edit description (remote files only).
  - `x`: Convert and save file formats in place (remote only).
  - `y` (Yank): Copy the selected file's absolute path (local) or File ID (remote) to the system clipboard.
  - `i`: Open the 70% responsive detailed file metadata inspector.
  - `r`: Refresh file lists.
  - `q`: Safely exit TUI mode.
- **Focus Persistence**:
  - Crucial UX Requirement: The active panel and cursor focus must remain unchanged before and after any operations (such as file deletions, transfers, or script executions). Do not automatically shift focus to the local panel after remote operations.

---

## 3. GAS Script Execution Engine (`exe1` / `exe2` / `webapps`)

- Map the `e` key to trigger GAS script execution.
- Provide a choice between three execution modes:
  1. `exe1`: Update remote project and execute a function.
  2. `exe2`: Execute local script directly via Google API (executes `main` function only).
  3. `webapps`: Execute local script via Web Apps URL (executes `main` function only).
- **Execution Workflow**:
  - Before running, prompt the user to input the target `Script ID` (for exe1/exe2) or `Web Apps URL` (for webapps). If these parameters are already configured in `ggsrun.cfg`, display them as the default placeholder value.
  - For `exe1`, the execution is run with sandboxing bypassed (since the user can verify the script manually in FD mode) and automatically rolls back all changes to the remote GAS project by default, ensuring a clean remote project state without prompting.
  - Show the script ID or Web Apps URL being utilized during the execution.
  - Clearly state in the UI that `exe2` and `webapps` only execute the `main` function.
  - If execution fails or is cancelled, safely return the focus to the previous panel/table.

---

## 4. Cross-Platform Compilation & Fallbacks

- To support multi-architecture building (especially for 32-bit Linux/ARM targets), separate file system creation time metrics into build-tagged files:
  - `file_info_linux.go` (target `linux`, utilizing `stat.Ctim`)
  - `file_info_darwin.go` (target `darwin`, utilizing `stat.Ctimespec`)
  - `file_info_windows.go` (target `windows`, utilizing `syscall.Win32FileAttributeData`)
  - `file_info_fallback.go` (default fallback returning standard modification time)
- **Safe Type Casting**:
  - Ensure all system-specific `Sec` and `Nsec` fields are explicitly cast to `int64` (e.g., `int64(stat.Ctim.Sec)`) before passing them to `time.Unix()` to prevent compilation failures on architectures where they are represented as `int32`.

---

## 5. Testing Requirements

- Provide a robust mock test suite in `fd_test.go` using `tcell.SimulationScreen`.
- Ensure all key behaviors (navigation, deletions, sorting, mime conversions, details modal rendering) are fully testable.
- Adjust tests to locate the newly refactored `TextView` details/error containers within `tview.Flex` instead of asserting the presence of `*tview.Modal`.
```

---

## 4. Development & Release History (v5.3.3)

### 📊 Consumed Resources
* **Conversations**: 12 sessions (long-term multi-turn development spanning automated context compactions).
* **Development Time**: Approximately 3 hours of continuous parallel analysis, system testing, and type hardening.
* **Quota Footprint**: High. Complex CSS-like terminal layouts, event loops, and multi-platform compilation debugging pushed the context size to several hundred thousand tokens.

### 💡 Efficiency & Success Review
* **Mock Simulation Screen**: The event-driven simulation test environment implemented in `fd_test.go` using `tcell.SimulationScreen` proved extremely robust. This allowed for 100% automated layout and keystroke assertions, completely bypassing the need for manual, slow visual terminal debugging.
* **Platform Segregation via Build Tags**: Isolating file system metrics inside `file_info_linux.go`, `file_info_darwin.go`, and others bypassed standard architectural compilation bugs, allowing clean automated cross-building for targets like `linux/386`, `linux/arm`, and Windows systems.

### 🛠️ Key Improvements & Hardening
* **Recursive GAS Project Sync**: Integrated recursive walk algorithms when triggering project modifications via the TUI, supporting nested structures.
* **Visual Overwrite Indicators**: Local script overwrites are rendered in clean lists using `pterm.BulletListPrinter` before files are dispatched to Google APIs.
* **TUI Overwrite Guardrails**: Added a hard modal intercepting project updates to prompt a confirmation (Y/N) before modifying remote production script manifests.
* **ZIP Archive Integration**: Standardized downloading Apps Script projects directly into packaged local `.zip` files under the filer.
* **Popup Refactoring**: Replaced all legacy `tview.Modal` widgets with a custom, highly responsive `tview.Flex` centered popup system (15% left margin, 70% dialog width, 15% right margin), completely eliminating text clipping on low-resolution consoles.
* **Focus Locking**: Active panel focus is locked and persistent. Operations do not shift focus, allowing seamless workflows.
* **Wrap-around & Clipboard Yanking**: Navigating past lists wraps around automatically. Pressing `y` extracts the file path (local) or Google File ID (remote) directly to your clipboard.

---

## 5. How to Launch

To start the Interactive Terminal File Manager, run the following command in your terminal:

```bash
$ ggsrun fd
```

*Note: The filer loads your credentials dynamically from `ggsrun.cfg`. Ensure you have run `ggsrun setup` first (see the [Setup & Onboarding Guide](setup_guide.md)).*

---

## 6. Interactive Controls & Keybindings Reference

Use the following keyboard shortcuts to control the dual-pane manager:

| Keybinding | Action | Focus Target |
| :--- | :--- | :--- |
| **`Tab`** | Switches cursor focus between the Local (Upper) and Remote (Lower) panels. | Global |
| **`Up / Down`** | Navigates the active panel file list. Supports **Wrap-around** navigation. | Active Panel |
| **`Space`** | Toggles multi-selection state on the highlighted item. Selected items turn green. | Active Panel |
| **`Enter`** | Open/enter directory, preview local text files, open Google Drive files in browser (WSL2 optimized), or browse Google Apps Script source files. | Highlighted Item |
| **`F1`** | Copies selected item(s) to the opposite panel (Local ➔ Remote = Upload; Remote ➔ Local = Download). | Multi-Selection |
| **`F2`** | Moves selected item(s) to the opposite panel (Copies first, then deletes source). | Multi-Selection |
| **`F3`** | Deletes highlighted or multi-selected item(s) with an interactive confirmation modal. | Multi-Selection |
| **`F5`** | Creates a new Local Directory or Remote Google Drive folder in the currently open folder. | Active Panel |
| **`F8`** | Launches recursive folder searching (local) or Drive-wide search queries (remote). | Active Panel |
| **`c` / `m`** | Copy or Move items within the same panel. | Multi-Selection |
| **`n`** | Renames the highlighted file or directory. | Highlighted Item |
| **`t`** | Edits the file's Last Modified timestamp (Local or Remote). | Highlighted Item |
| **`d`** | Edits the description metadata string (Google Drive files only). | Highlighted Item (Remote) |
| **`x`** | Converts Google Workspace formats and saves them in place on Google Drive. | Highlighted Item (Remote) |
| **`e`** | Statefully executes Google Apps Script code (selects `exe1`, `exe2`, or `webapps` dynamically). | Selected File |
| **`i`** | Displays detailed metadata inside a 70% responsive centered popup (includes **Web View Link**). | Highlighted Item |
| **`s`** | Sorts files (opens popup to choose sort keys like Name, Size, Date, or Order). | Active Panel |
| **`y` (Yank)** | Copies the highlighted file's absolute path (local) or File ID (remote) to the clipboard. | Highlighted Item |
| **`r`** | Refreshes file and directory lists in both panels (also clears F8 search highlighting). | Global |
| **`q`** | Safely exits FD Mode (prompts confirmation dialog first). | Global |

---

## 7. Premium TUI Features (v5.3.6+)

Starting with **v5.3.6**, FD Mode has been enhanced with enterprise-grade capabilities to deliver a beautiful, premium visual experience:

### Function Key Actions
All core file system operations (Copy, Move, Delete, Create Folder, and Search) are mapped directly to standard function keys (`F1` to `F8`). This conforms with standard double-pane file managers (like Norton Commander or Midnight Commander), maximizing muscle-memory efficiency.

### Recursive & Drive-Wide Search (`F8`)
Pressing `F8` dynamically prompts you for a query search:
* **Local Table**: Performs an ultra-fast, recursive walk search across all directories and subdirectories starting from your currently open local directory.
* **Remote Table**: Standardizes search queries across your entire Google Drive infrastructure, including enterprise Shared Drives and Team Drives.

### Yellow Search Highlighting
When search results are returned, the active panel's borders, titles, and item counts turn **yellow**, giving you a high-visibility, instant cue that you are looking at a filtered search subset. A helper banner `(Press 'r' to return to normal view)` is printed. Pressing `r` clears the search parameters, restores standard borders, and re-lists default files.

### Web View Link Integration (`i` details)
When highlighting a file on Google Drive and pressing `i`, the metadata viewer includes the `webViewLink`. You can double-click or copy this URL to open the asset directly inside a web browser, or preview spreadsheet columns on your host system.

### Directory Tree Preview
Before folder transfers begin, `ggsrun` prints an elegant, ASCII-rendered directory tree structure of the source directory, letting you preview what you are about to copy or upload:

```text
📁 local_src/
├── 📄 Code.js
├── 📁 assets/
│   ├── 📄 logo.png
│   └── 📄 styles.css
└── 📄 index.html
```

### Real-Time Individual Progress Bars
During file transfers, the CLI overlay renders individual, real-time streaming progress bars for every file. If you are uploading 5 files in parallel, you will see 5 active bars loading concurrently, giving you premium visibility.

---

> [!TIP]
> **WSL2 and Headless browser support:**
> If you are running `ggsrun fd` inside WSL2 (Windows Subsystem for Linux), pressing `Enter` on a Google Drive file will automatically detect your host Windows browser and open the file seamlessly in Windows! If running on a headless Linux server, it gracefully logs the URL to standard error.

---

### Related Links:
- 🚀 **[Setup & Onboarding Guide](setup_guide.md)** - Get authenticated and configured before starting the TUI.
- 📖 **[Command Reference Manual](commands_reference.md)** - Learn about the underlying CLI flags and transfer commands.
- 🧪 **[Manual Integration Tests Suite](../manual-tests/README.md)** - Verify your TUI configurations.
- 🏡 **[Back to Home](../README.md)**
