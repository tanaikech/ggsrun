package tui

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ggsrun/internal/app"
	"ggsrun/internal/utl"

	"github.com/gdamore/tcell/v2"
	"github.com/pterm/pterm"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

var (
	tuiApp                *tview.Application
	pages                 *tview.Pages
	localTable            *tview.Table
	remoteTable           *tview.Table
	statusBar             *tview.TextView
	currentLocalDir       string
	currentRemoteFolderID string
	remoteFolderStack     []FolderInfo
	localFiles            []FileEntry
	remoteFiles           []FileEntry
	authContainer         *app.AuthContainer
	mainCtx               *cli.Context
	testScreen            tcell.Screen

	selectedLocalPaths = make(map[string]bool)
	selectedRemoteIDs  = make(map[string]bool)

	lastActiveTable         tview.Primitive
	activeTableBeforeAction tview.Primitive
	localSortKey       = "none"
	localSortOrder     = "asc"
	remoteSortKey      = "none"
	remoteSortOrder    = "asc"
	currentLoadingView *tview.TextView
	loadingViewMu      sync.Mutex
	isAuthorized       = true

	transferProgressMu sync.Mutex
	transferProgress   map[string]jobProgress
	transferHeading    string
	transferLogLines   []string

	inSearchModeLocal  bool
	inSearchModeRemote bool

	// Mockable function variables for unit testing
	listLocalFilesFn   = listLocalFiles
	listRemoteFilesFn  = listRemoteFiles
	openBrowserFn      = openBrowser
	tuiUploadFn        = app.TuiUpload
	tuiDownloadFn      = app.TuiDownload
	getAuthContainerFn = app.GetAuthenticatedAuthContainer
	tuiExecuteGasFn    = app.TuiExecuteGas
	tuiUpdateDriveMetadataFn = app.TuiUpdateDriveMetadata
	tuiCopyDriveFileFn = app.TuiCopyDriveFile
	getBoundScriptExportedFn = func(auth *app.AuthContainer, c *cli.Context, scriptID string) *utl.ProjectForAppsScriptApi {
		p := auth.DefDownloadContainerExported(c)
		return p.GetBoundScriptExported(scriptID)
	}
	requestParamsFetchFn = func(r *utl.RequestParams) ([]byte, error) {
		return r.FetchAPI()
	}
	requestParamsFetchRawFn = func(r *utl.RequestParams) (*http.Response, error) {
		return r.FetchAPIRaw()
	}
	deleteRemoteFileFn = func(auth *app.AuthContainer, c *cli.Context, fileID string) error {
		delCtx := createOpContext(c, map[string]string{
			"deletefile": fileID,
			"jsonparser": "true",
		})
		p := auth.DefDownloadContainerExported(delCtx)
		p.Downloader(delCtx)
		return nil
	}
	tuiRunExe1Fn           = app.TuiRunExe1
	tuiRunExe2Fn           = app.TuiRunExe2
	tuiRunWebAppsFn        = app.TuiRunWebApps
	osReadFileFn           = os.ReadFile
	tuiCreateDriveFolderFn = app.TuiCreateDriveFolder
	searchRemoteDriveAllFn = searchRemoteDriveAll
	generateRemoteTreeFn   func(*app.AuthContainer, *cli.Context, string, string) ([]string, error)
)

type FolderInfo struct {
	ID   string
	Name string
}

type jobProgress struct {
	Transferred int64
	Total       int64
	Status      string
}

type FileEntry struct {
	Name        string
	Path        string
	MimeType    string
	ModTime     string
	CreatedTime string
	Size        int64
	IsDir       bool
	WebViewLink string
	Permissions string
	Description string
}

type TransferJob struct {
	SourcePath string
	SourceName string
	TargetMime string
	IsDir      bool
	MimeType   string
}

func init() {
	app.RunTUIFunc = RunTUI
	generateRemoteTreeFn = generateRemoteTree
}

func RunTUI(c *cli.Context) error {
	utl.TUIExitHandler = func(code int) {
		panic(fmt.Errorf("internal process exited with code %d", code))
	}
	defer func() {
		utl.TUIExitHandler = nil
	}()

	pterm.DisableOutput()
	mainCtx = c
	var hasConfig bool = true
	func() {
		defer func() {
			if r := recover(); r != nil {
				hasConfig = false
			}
		}()
		authContainer = getAuthContainerFn(c)
	}()
	isAuthorized = hasConfig
	if flag.Lookup("test.v") != nil {
		isAuthorized = true
	}

	var err error
	currentLocalDir, err = filepath.Abs(".")
	if err != nil {
		currentLocalDir = "."
	}

	currentRemoteFolderID = "root"
	remoteFolderStack = []FolderInfo{{ID: "root", Name: "root"}}

	selectedLocalPaths = make(map[string]bool)
	selectedRemoteIDs = make(map[string]bool)

	tuiApp = tview.NewApplication()
	if testScreen != nil {
		tuiApp.SetScreen(testScreen)
	}
	pages = tview.NewPages()

	localTable = tview.NewTable().SetBorders(false)
	localTable.SetFocusFunc(func() {
		lastActiveTable = localTable
	})
	localTable.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

	remoteTable = tview.NewTable().SetBorders(false)
	remoteTable.SetFocusFunc(func() {
		lastActiveTable = remoteTable
	})
	remoteTable.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

	statusBar = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(localTable, 0, 1, true)
	flex.AddItem(remoteTable, 0, 1, false)
	if !isAuthorized {
		warningBar := tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignLeft)
		warningText := "[red]WARNING: Credentials 'ggsrun.cfg' not found. Google Drive features are disabled.\n" +
			" -> Please download your credentials JSON file from Google Cloud Console,\n" +
			"    and run 'ggsrun auth' to authenticate and generate 'ggsrun.cfg'.\n" +
			" -> To place 'ggsrun.cfg' in a specific directory, set the 'GGSRUN_CFG_PATH' env var.[white]"
		warningBar.SetText(warningText)
		flex.AddItem(warningBar, 4, 0, false)
	}
	flex.AddItem(statusBar, 1, 0, false)

	pages.AddPage("main", flex, true, true)

	localTable.SetSelectionChangedFunc(func(row, column int) {
		updateStatus()
	})
	remoteTable.SetSelectionChangedFunc(func(row, column int) {
		updateStatus()
	})

	localTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			row, _ := localTable.GetSelection()
			if row == 1 {
				localTable.Select(localTable.GetRowCount()-1, 0)
				return nil
			}
		case tcell.KeyDown:
			row, _ := localTable.GetSelection()
			if row == localTable.GetRowCount()-1 {
				localTable.Select(1, 0)
				return nil
			}
		case tcell.KeyEnter:
			onEnterLocal()
			return nil
		case tcell.KeyF1:
			onCopy()
			return nil
		case tcell.KeyF2:
			onMove()
			return nil
		case tcell.KeyF3:
			onDelete()
			return nil
		case tcell.KeyF5:
			onCreateDirOrFolder()
			return nil
		case tcell.KeyF8:
			onSearch()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				toggleSelection()
				return nil
			case 'c':
				onDomainCopy()
				return nil
			case 'm':
				onDomainMove()
				return nil
			case 'n':
				onRename()
				return nil
			case 't':
				onChangeTimestamp()
				return nil
			case 'i':
				showFileDetails()
				return nil
			case 'r':
				inSearchModeLocal = false
				inSearchModeRemote = false
				refreshPanels()
				return nil
			case 'e':
				onExecuteLocalScript()
				return nil
			case 'h':
				showHelp()
				return nil
			case 's':
				showSortChoice()
				return nil
			case 'y':
				selectedRow := getSelectedRowIndex(localTable)
				if selectedRow >= 0 && selectedRow < len(localFiles) {
					absPath, err := filepath.Abs(localFiles[selectedRow].Path)
					if err == nil {
						if errCopy := copyToClipboard(absPath); errCopy == nil {
							statusBar.SetText(fmt.Sprintf(" [green]Copied path to clipboard:[white] %s", absPath))
						} else {
							showError("Failed to copy to clipboard: " + errCopy.Error())
						}
					}
				}
				return nil
			case 'q', 'Q':
				promptExit()
				return nil
			}
		}
		return event
	})

	remoteTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			row, _ := remoteTable.GetSelection()
			if row == 1 {
				remoteTable.Select(remoteTable.GetRowCount()-1, 0)
				return nil
			}
		case tcell.KeyDown:
			row, _ := remoteTable.GetSelection()
			if row == remoteTable.GetRowCount()-1 {
				remoteTable.Select(1, 0)
				return nil
			}
		case tcell.KeyEnter:
			onEnterRemote()
			return nil
		case tcell.KeyF1:
			onCopy()
			return nil
		case tcell.KeyF2:
			onMove()
			return nil
		case tcell.KeyF3:
			onDelete()
			return nil
		case tcell.KeyF5:
			onCreateDirOrFolder()
			return nil
		case tcell.KeyF8:
			onSearch()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				toggleSelection()
				return nil
			case 'c':
				onDomainCopy()
				return nil
			case 'm':
				onDomainMove()
				return nil
			case 'n':
				onRename()
				return nil
			case 't':
				onChangeTimestamp()
				return nil
			case 'd':
				onChangeDescription()
				return nil
			case 'x':
				onConvertMimeType()
				return nil
			case 'e':
				onExecuteRemoteScript()
				return nil
			case 'i':
				showFileDetails()
				return nil
			case 'r':
				inSearchModeLocal = false
				inSearchModeRemote = false
				refreshPanels()
				return nil
			case 'h':
				showHelp()
				return nil
			case 's':
				showSortChoice()
				return nil
			case 'y':
				selectedRow := getSelectedRowIndex(remoteTable)
				if selectedRow >= 0 && selectedRow < len(remoteFiles) {
					fileID := remoteFiles[selectedRow].Path
					if fileID != "" {
						if errCopy := copyToClipboard(fileID); errCopy == nil {
							statusBar.SetText(fmt.Sprintf(" [green]Copied file ID to clipboard:[white] %s", fileID))
						} else {
							showError("Failed to copy to clipboard: " + errCopy.Error())
						}
					}
				}
				return nil
			case 'q', 'Q':
				promptExit()
				return nil
			}
		}
		return event
	})

	tuiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			promptExit()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			if localTable.HasFocus() {
				tuiApp.SetFocus(remoteTable)
			} else {
				tuiApp.SetFocus(localTable)
			}
			updateStatus()
			return nil
		}
		return event
	})

	refreshPanels()

	if err := tuiApp.SetRoot(pages, true).Run(); err != nil {
		return err
	}

	return nil
}

func getSelectedRowIndex(table *tview.Table) int {
	row, _ := table.GetSelection()
	return row - 1
}

func toggleSelection() {
	if localTable.HasFocus() {
		selectedRow := getSelectedRowIndex(localTable)
		if selectedRow >= 0 && selectedRow < len(localFiles) {
			item := localFiles[selectedRow]
			if item.Name == ".." {
				return
			}
			selectedLocalPaths[item.Path] = !selectedLocalPaths[item.Path]
			populateTable(localTable, localFiles, true)
		}
	} else {
		selectedRow := getSelectedRowIndex(remoteTable)
		if selectedRow >= 0 && selectedRow < len(remoteFiles) {
			item := remoteFiles[selectedRow]
			if item.Name == ".." {
				return
			}
			selectedRemoteIDs[item.Path] = !selectedRemoteIDs[item.Path]
			populateTable(remoteTable, remoteFiles, false)
		}
	}
}

func getSelectedJobs(isLocal bool) []TransferJob {
	var jobs []TransferJob
	if isLocal {
		hasSelections := false
		for _, selected := range selectedLocalPaths {
			if selected {
				hasSelections = true
				break
			}
		}

		if hasSelections {
			for _, item := range localFiles {
				if selectedLocalPaths[item.Path] {
					jobs = append(jobs, TransferJob{
						SourcePath: item.Path,
						SourceName: item.Name,
						IsDir:      item.IsDir,
						MimeType:   item.MimeType,
					})
				}
			}
		} else {
			selectedRow := getSelectedRowIndex(localTable)
			if selectedRow >= 0 && selectedRow < len(localFiles) {
				item := localFiles[selectedRow]
				if item.Name != ".." {
					jobs = append(jobs, TransferJob{
						SourcePath: item.Path,
						SourceName: item.Name,
						IsDir:      item.IsDir,
						MimeType:   item.MimeType,
					})
				}
			}
		}
	} else {
		hasSelections := false
		for _, selected := range selectedRemoteIDs {
			if selected {
				hasSelections = true
				break
			}
		}

		if hasSelections {
			for _, item := range remoteFiles {
				if selectedRemoteIDs[item.Path] {
					jobs = append(jobs, TransferJob{
						SourcePath: item.Path,
						SourceName: item.Name,
						IsDir:      item.IsDir,
						MimeType:   item.MimeType,
					})
				}
			}
		} else {
			selectedRow := getSelectedRowIndex(remoteTable)
			if selectedRow >= 0 && selectedRow < len(remoteFiles) {
				item := remoteFiles[selectedRow]
				if item.Name != ".." {
					jobs = append(jobs, TransferJob{
						SourcePath: item.Path,
						SourceName: item.Name,
						IsDir:      item.IsDir,
						MimeType:   item.MimeType,
					})
				}
			}
		}
	}
	return jobs
}

func updateStatus() {
	var selected FileEntry
	var panelName string
	if localTable.HasFocus() {
		panelName = "Local"
		if selectedRow := getSelectedRowIndex(localTable); selectedRow >= 0 && selectedRow < len(localFiles) {
			selected = localFiles[selectedRow]
		}
	} else {
		panelName = "Drive"
		if selectedRow := getSelectedRowIndex(remoteTable); selectedRow >= 0 && selectedRow < len(remoteFiles) {
			selected = remoteFiles[selectedRow]
		}
	}

	summary := "No item selected"
	if selected.Name != "" {
		if selected.IsDir {
			summary = fmt.Sprintf("[%s DIR] %s", panelName, selected.Name)
		} else {
			summary = fmt.Sprintf("[%s FILE] %s | %s | %s", panelName, selected.Name, selected.MimeType, formatSize(selected.Size, false, selected.MimeType))
		}
	}

	actionsHelp := "F1:Copy  F2:Move  F3:Del  F5:Mkdir  F8:Srch  y:YankPath  i:Details  h:Help"
	if !localTable.HasFocus() {
		actionsHelp = "F1:Copy  F2:Move  F3:Del  F5:Mkdir  F8:Srch  y:YankID  i:Details  h:Help"
	}
	statusBar.SetText(fmt.Sprintf(" %-50s | %s", summary, actionsHelp))
}

func formatSize(size int64, isDir bool, mimeType string) string {
	if isDir {
		return "---"
	}
	if strings.Contains(mimeType, "application/vnd.google-apps") {
		return "[GoogleDoc]"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func listLocalFiles(dir string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs []FileEntry
	var files []FileEntry

	absDir, err := filepath.Abs(dir)
	if err == nil {
		parent := filepath.Dir(absDir)
		if parent != absDir {
			dirs = append(dirs, FileEntry{
				Name:  "..",
				Path:  parent,
				IsDir: true,
			})
		}
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		path := filepath.Join(dir, name)
		isDir := entry.IsDir()

		mimeType := ""
		if isDir {
			mimeType = "directory"
		} else {
			mimeType = utl.ExtToMime(filepath.Ext(name))
		}

		modTime := info.ModTime().Format("2006-01-02 15:04:05")
		size := info.Size()

		perm := info.Mode().String()
		createdTime := getLocalCreatedTime(info).Format("2006-01-02 15:04:05")

		item := FileEntry{
			Name:        name,
			Path:        path,
			MimeType:    mimeType,
			ModTime:     modTime,
			CreatedTime: createdTime,
			Size:        size,
			IsDir:       isDir,
			Permissions: perm,
		}

		if isDir {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}

	return append(dirs, files...), nil
}

func listRemoteFiles(auth *app.AuthContainer, c *cli.Context, folderID string) ([]FileEntry, error) {
	p := auth.DefDownloadContainerExported(c)
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	fields := "files(id,name,mimeType,modifiedTime,createdTime,size,webViewLink,owners(displayName),shared,description),nextPageToken"

	fl := p.GetListLoop(q, fields)

	var dirs []FileEntry
	var files []FileEntry

	if folderID != "root" {
		dirs = append(dirs, FileEntry{
			Name:     "..",
			Path:     "",
			IsDir:    true,
			MimeType: "application/vnd.google-apps.folder",
		})
	}

	for _, f := range fl.Files {
		isDir := f.MimeType == "application/vnd.google-apps.folder"

		var size int64
		if f.Size != "" {
			size, _ = strconv.ParseInt(f.Size, 10, 64)
		}

		modTime := ""
		if f.ModifiedTime != nil {
			modTime = f.ModifiedTime.In(time.Local).Format("2006-01-02 15:04:05")
		}

		createdTime := ""
		if f.CreatedTime != nil {
			createdTime = f.CreatedTime.In(time.Local).Format("2006-01-02 15:04:05")
		}

		perm := ""
		if len(f.Owners) > 0 {
			perm = "Owner: " + f.Owners[0].Name
			if f.Shared {
				perm += " (Shared)"
			} else {
				perm += " (Private)"
			}
		}

		item := FileEntry{
			Name:        f.Name,
			Path:        f.ID,
			MimeType:    f.MimeType,
			ModTime:     modTime,
			CreatedTime: createdTime,
			Size:        size,
			IsDir:       isDir,
			WebViewLink: f.WebView,
			Permissions: perm,
			Description: f.Description,
		}

		if isDir {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}

	return append(dirs, files...), nil
}

func simplifyMimeType(mime string, isDir bool) string {
	if isDir {
		return "Folder"
	}
	switch mime {
	case "application/vnd.google-apps.folder":
		return "Folder"
	case "application/vnd.google-apps.document":
		return "Google Docs"
	case "application/vnd.google-apps.spreadsheet":
		return "Google Sheets"
	case "application/vnd.google-apps.presentation":
		return "Google Slides"
	case "application/vnd.google-apps.form":
		return "Google Forms"
	case "application/vnd.google-apps.script":
		return "Google Apps Script"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/msword":
		return "DOCX"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/vnd.ms-excel":
		return "XLSX"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation", "application/vnd.ms-powerpoint":
		return "PPTX"
	case "text/csv":
		return "CSV"
	case "application/pdf":
		return "PDF"
	case "text/plain":
		return "TXT"
	case "text/html":
		return "HTML"
	case "text/markdown", "text/x-markdown":
		return "MD"
	default:
		if strings.HasPrefix(mime, "image/") {
			return "Image"
		}
		if strings.HasPrefix(mime, "audio/") {
			return "Audio"
		}
		if strings.HasPrefix(mime, "video/") {
			return "Video"
		}
		return "File"
	}
}

func populateTable(table *tview.Table, files []FileEntry, isLocal bool) {
	table.Clear()

	headers := []string{"Name", "Parent/Path", "MimeType", "Perm", "Created Date", "Update Date", "Size"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetExpansion(1)
		if i == 0 {
			cell.SetExpansion(2)
		}
		table.SetCell(0, i, cell)
	}

	for row, item := range files {
		displayName := item.Name
		if item.IsDir && item.Name != ".." {
			displayName = "[DIR] " + item.Name
		}

		isSelected := false
		if isLocal {
			isSelected = selectedLocalPaths[item.Path]
		} else {
			isSelected = selectedRemoteIDs[item.Path]
		}

		nameColor := tcell.ColorWhite
		if item.IsDir {
			nameColor = tcell.ColorTeal
		}
		if isSelected {
			displayName = "* " + displayName
			nameColor = tcell.ColorYellow
		}

		relPath := item.Path
		if item.Name == ".." {
			relPath = "../"
		} else {
			if isLocal {
				parentOfCurrentLocalDir := filepath.Dir(currentLocalDir)
				if r, err := filepath.Rel(parentOfCurrentLocalDir, item.Path); err == nil {
					relPath = r
				}
			} else {
				if len(remoteFolderStack) > 0 {
					currentFolder := remoteFolderStack[len(remoteFolderStack)-1]
					relPath = currentFolder.Name + "/" + item.Name
				} else {
					relPath = "root/" + item.Name
				}
			}
		}

		sizeStr := formatSize(item.Size, item.IsDir, item.MimeType)

		selStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
		table.SetCell(row+1, 0, tview.NewTableCell(displayName).SetTextColor(nameColor).SetSelectedStyle(selStyle))
		table.SetCell(row+1, 1, tview.NewTableCell(relPath).SetTextColor(tcell.GetColor("cyan")).SetSelectedStyle(selStyle))
		table.SetCell(row+1, 2, tview.NewTableCell(simplifyMimeType(item.MimeType, item.IsDir)).SetTextColor(tcell.ColorGreen).SetSelectedStyle(selStyle))
		table.SetCell(row+1, 3, tview.NewTableCell(item.Permissions).SetTextColor(tcell.ColorPink).SetSelectedStyle(selStyle))
		table.SetCell(row+1, 4, tview.NewTableCell(item.CreatedTime).SetTextColor(tcell.ColorLightGray).SetSelectedStyle(selStyle))
		table.SetCell(row+1, 5, tview.NewTableCell(item.ModTime).SetTextColor(tcell.ColorLightGray).SetSelectedStyle(selStyle))
		table.SetCell(row+1, 6, tview.NewTableCell(sizeStr).SetTextColor(tcell.ColorYellow).SetSelectedStyle(selStyle))
	}

	table.SetSelectable(true, false)
	table.SetFixed(1, 0)
}

func refreshPanels() {
	var err error
	if !inSearchModeLocal {
		localFiles, err = listLocalFilesFn(currentLocalDir)
		if err != nil {
			showError("Failed to list local files: " + err.Error())
		} else {
			sortFileEntries(localFiles, localSortKey, localSortOrder)
			populateTable(localTable, localFiles, true)
			localTable.SetTitle(" Local File System: " + currentLocalDir + " ")
			localTable.SetBorder(true)
			localTable.SetTitleColor(tview.Styles.TitleColor)
			localTable.SetBorderColor(tview.Styles.BorderColor)
		}
	}

	if isAuthorized {
		if !inSearchModeRemote {
			remoteFiles, err = listRemoteFilesFn(authContainer, mainCtx, currentRemoteFolderID)
			if err != nil {
				showError("Failed to list remote files: " + err.Error())
			} else {
				sortFileEntries(remoteFiles, remoteSortKey, remoteSortOrder)
				populateTable(remoteTable, remoteFiles, false)
				remotePath := "/"
				if len(remoteFolderStack) > 1 {
					var names []string
					for _, f := range remoteFolderStack {
						if f.Name != "root" {
							names = append(names, f.Name)
						}
					}
					remotePath = "/" + strings.Join(names, "/")
				}
				remoteTable.SetTitle(" Google Drive: " + remotePath + " ")
				remoteTable.SetBorder(true)
				remoteTable.SetTitleColor(tview.Styles.TitleColor)
				remoteTable.SetBorderColor(tview.Styles.BorderColor)
			}
		}
	} else {
		remoteFiles = nil
		remoteTable.Clear()
		remoteTable.SetCell(0, 0, tview.NewTableCell("  [red]Google Drive features are disabled (No ggsrun.cfg)[white]").SetSelectable(false))
		remoteTable.SetTitle(" Google Drive: (Unauthenticated) ")
		remoteTable.SetBorder(true)
		remoteTable.SetTitleColor(tview.Styles.TitleColor)
		remoteTable.SetBorderColor(tview.Styles.BorderColor)
	}

	updateStatus()
}

func createOpContext(mainCtx *cli.Context, extraFlags map[string]string) *cli.Context {
	appObj := cli.NewApp()
	set := flag.NewFlagSet("tui-op", flag.ContinueOnError)

	set.String("fileid", "", "")
	set.String("filename", "", "")
	set.String("projecttype", "", "")
	set.String("destination", "", "")
	set.String("parentfolderid", "", "")
	set.String("projectname", "", "")
	set.String("googledocname", "", "")
	set.String("config", "", "")
	set.String("credentials", "", "")
	set.String("extension", "", "")
	set.String("convertto", "", "")
	set.Bool("jsonparser", false, "")
	set.Bool("overwrite", false, "")
	set.Bool("noconvert", false, "")
	set.Bool("zip", false, "")
	set.Bool("rawdata", false, "")
	set.String("scriptid", "", "")
	set.String("scriptfile", "", "")
	set.String("stringscript", "", "")
	fSlice := &cli.StringSlice{}
	set.Var(fSlice, "function", "")
	set.String("value", "", "")
	set.Bool("backup", false, "")
	set.Bool("deleteScript", false, "")
	set.Bool("undeleteScript", false, "")
	set.Bool("onlyresult", false, "")
	set.Bool("foldertree", false, "")
	set.Bool("convert", false, "")
	set.Bool("log", false, "")
	set.String("url", "", "")
	set.String("password", "", "")
	set.String("deletefile", "", "")
	set.String("conflict", "", "")

	args := []string{}
	if mainCtx != nil {
		if val := mainCtx.String("config"); val != "" {
			args = append(args, "--config", val)
		}
		if val := mainCtx.String("credentials"); val != "" {
			args = append(args, "--credentials", val)
		}
	}
	for k, v := range extraFlags {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}

	_ = set.Parse(args)
	return cli.NewContext(appObj, set, nil)
}

func runTask(loadingMsg string, task func() error) {
	prevFocus := tuiApp.GetFocus()
	activeTableBeforeAction = lastActiveTable
	
	loadingView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	loadingView.SetBorder(true).SetTitle(" Working ")

	fmt.Fprintf(loadingView, "[yellow]%s[white]\n\n", loadingMsg)

	loadingFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	loadingFlex.AddItem(tview.NewBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(loadingView, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	loadingFlex.AddItem(inner, 15, 0, true)
	loadingFlex.AddItem(tview.NewBox(), 0, 1, false)

	var mu sync.Mutex
	cancelled := false

	loadingFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			mu.Lock()
			cancelled = true
			mu.Unlock()

			pages.RemovePage("loading")
			if activeTableBeforeAction != nil {
				tuiApp.SetFocus(activeTableBeforeAction)
			} else if lastActiveTable != nil {
				tuiApp.SetFocus(lastActiveTable)
			} else {
				tuiApp.SetFocus(prevFocus)
			}
			showError("Operation cancelled/interrupted by user.")
			return nil
		}
		return event
	})

	pages.AddPage("loading", loadingFlex, true, true)
	tuiApp.SetFocus(loadingFlex)

	loadingViewMu.Lock()
	currentLoadingView = loadingView
	loadingViewMu.Unlock()

	app.TUIProgressCallback = func(msg string) {
		tuiApp.QueueUpdateDraw(func() {
			loadingViewMu.Lock()
			defer loadingViewMu.Unlock()
			if currentLoadingView != nil {
				transferProgressMu.Lock()
				isTransferActive := transferProgress != nil
				transferProgressMu.Unlock()

				if isTransferActive {
					updateAndRenderProgress(msg)
				} else {
					fmt.Fprintln(currentLoadingView, " * "+msg)
					currentLoadingView.ScrollToEnd()
				}
			}
		})
	}

	go func() {
		defer func() {
			loadingViewMu.Lock()
			currentLoadingView = nil
			loadingViewMu.Unlock()
			app.TUIProgressCallback = nil

			if r := recover(); r != nil {
				tuiApp.QueueUpdateDraw(func() {
					pages.RemovePage("loading")
					refreshPanels()
					if activeTableBeforeAction != nil {
						tuiApp.SetFocus(activeTableBeforeAction)
					} else if lastActiveTable != nil {
						tuiApp.SetFocus(lastActiveTable)
					} else {
						tuiApp.SetFocus(prevFocus)
					}
					showError(fmt.Sprintf("Internal Error: %v", r))
				})
			}
		}()

		err := task()

		mu.Lock()
		wasCancelled := cancelled
		mu.Unlock()

		if wasCancelled {
			return
		}

		tuiApp.QueueUpdateDraw(func() {
			pages.RemovePage("loading")
			frontPage, _ := pages.GetFrontPage()
			if frontPage == "main" {
				refreshPanels()
				if activeTableBeforeAction != nil {
					tuiApp.SetFocus(activeTableBeforeAction)
				} else if lastActiveTable != nil {
					tuiApp.SetFocus(lastActiveTable)
				} else {
					tuiApp.SetFocus(prevFocus)
				}
			}
			if err != nil {
				if frontPage != "main" {
					if activeTableBeforeAction != nil {
						tuiApp.SetFocus(activeTableBeforeAction)
					} else if lastActiveTable != nil {
						tuiApp.SetFocus(lastActiveTable)
					} else {
						tuiApp.SetFocus(prevFocus)
					}
				}
				showError(err.Error())
			}
		})
	}()
}

func showError(msg string) {
	prevFocus := tuiApp.GetFocus()

	textView := tview.NewTextView().
		SetText("Error:\n" + msg).
		SetDynamicColors(false)
	textView.SetBorder(true).
		SetTitle(" Error ").
		SetTitleColor(tcell.ColorRed)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 15, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(textView, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	flex.AddItem(inner, 0, 70, true)
	flex.AddItem(tview.NewBox(), 0, 15, false)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
			pages.RemovePage("error")
			if activeTableBeforeAction != nil {
				tuiApp.SetFocus(activeTableBeforeAction)
			} else if lastActiveTable != nil {
				tuiApp.SetFocus(lastActiveTable)
			} else {
				tuiApp.SetFocus(prevFocus)
			}
			return nil
		}
		return event
	})

	pages.AddPage("error", flex, true, true)
	tuiApp.SetFocus(textView)
}

func readTextPreview(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 1024*32) // Read first 32KB
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	content := string(buf[:n])
	if strings.Contains(content, "\x00") {
		return "", fmt.Errorf("file appears to be binary")
	}

	return content, nil
}

func showTextPreview(title string, text string) {
	prevFocus := tuiApp.GetFocus()
	if lastActiveTable != nil {
		prevFocus = lastActiveTable
	}

	textView := tview.NewTextView().
		SetText(text).
		SetDynamicColors(false)

	textView.SetBorder(true).
		SetTitle(" Preview: " + title + " ").
		SetTitleColor(tcell.ColorGreen)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
			pages.RemovePage("preview")
			tuiApp.SetFocus(prevFocus)
			return nil
		}
		return event
	})

	pages.AddPage("preview", textView, true, true)
	tuiApp.SetFocus(textView)
}

func showGasProject(title, scriptID string) {
	prevFocus := tuiApp.GetFocus()

	runTask("Loading GAS Project contents...", func() error {
		projectContent := getBoundScriptExportedFn(authContainer, mainCtx, scriptID)

		tuiApp.QueueUpdateDraw(func() {
			if projectContent == nil || len(projectContent.Files) == 0 {
				showError("Project has no files or failed to load.")
				return
			}

			fileList := tview.NewList()
			codeView := tview.NewTextView().SetDynamicColors(false)
			codeView.SetBorder(true).SetTitle(" Source Code ")

			for i, f := range projectContent.Files {
				fileList.AddItem(f.Name+" ("+f.Type+")", "", rune(48+i%10), nil)
			}

			fileList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
				if index >= 0 && index < len(projectContent.Files) {
					codeView.SetText(projectContent.Files[index].Source)
					codeView.ScrollToBeginning()
				}
			})

			fileList.SetCurrentItem(0)
			codeView.SetText(projectContent.Files[0].Source)

			fileList.SetBorder(true).SetTitle(" Files ")

			flex := tview.NewFlex().SetDirection(tview.FlexColumn)
			flex.AddItem(fileList, 30, 0, true)
			flex.AddItem(codeView, 0, 1, false)

			status := tview.NewTextView().
				SetText(" Esc:Close  Tab:SwitchPane  e:ExecuteFunction ").
				SetTextColor(tcell.ColorYellow)

			mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)
			mainFlex.AddItem(flex, 0, 1, true)
			mainFlex.AddItem(status, 1, 0, false)

			mainFlex.SetBorder(true).SetTitle(" GAS Project: " + title + " ")

			gasInputCapture := func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEsc:
					pages.RemovePage("gas_project")
					tuiApp.SetFocus(prevFocus)
					return nil
				case tcell.KeyTab:
					if fileList.HasFocus() {
						tuiApp.SetFocus(codeView)
					} else {
						tuiApp.SetFocus(fileList)
					}
					return nil
				case tcell.KeyRune:
					if event.Rune() == 'e' {
						promptExecuteGas(scriptID)
						return nil
					}
				}
				return event
			}
			fileList.SetInputCapture(gasInputCapture)
			codeView.SetInputCapture(gasInputCapture)

			pages.AddPage("gas_project", mainFlex, true, true)
			tuiApp.SetFocus(fileList)
		})

		return nil
	})
}

func promptExecuteGas(scriptID string) {
	promptGasExecution(false, "", scriptID, false)
}

func promptGasExecution(isLocal bool, scriptFile string, remoteScriptID string, isMultiOrDir bool) {
	prevFocus := tuiApp.GetFocus()

	list := tview.NewList().
		AddItem("exe1 (Update remote project & execute function)", "Uploads local script and executes", '1', nil).
		AddItem("exe2 (Execute script directly via Execution API)", "Runs local script via server without uploading", '2', nil).
		AddItem("webapps (Execute script via Web Apps URL)", "Runs local script via Web Apps URL", '3', nil)

	list.SetBorder(true).
		SetTitle(" Select GAS Execution Mode ").
		SetTitleColor(tcell.ColorYellow)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(list, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	flex.AddItem(inner, 10, 0, true)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("exec_mode_choice")
			tuiApp.SetFocus(prevFocus)
			return nil
		}
		return event
	})

	list.SetSelectedFunc(func(itemIndex int, mainText string, secondaryText string, shortcut rune) {
		pages.RemovePage("exec_mode_choice")
		tuiApp.SetFocus(prevFocus)

		switch itemIndex {
		case 0:
			collectExeParams(true, isLocal, scriptFile, remoteScriptID, isMultiOrDir)
		case 1:
			collectExeParams(false, isLocal, scriptFile, remoteScriptID, isMultiOrDir)
		case 2:
			collectWebAppsParams(isLocal, scriptFile, remoteScriptID, isMultiOrDir)
		}
	})

	pages.AddPage("exec_mode_choice", flex, true, true)
	tuiApp.SetFocus(list)
}

func collectExeParams(isExe1 bool, isLocal bool, scriptFile string, remoteScriptID string, isMultiOrDir bool) {
	if isMultiOrDir && !isExe1 {
		showError("Directory or multiple script execution is strictly limited to 'exe1'. It is not supported for 'exe2'.")
		return
	}

	var defaultScriptID string
	if authContainer != nil && authContainer.GgsrunCfg != nil {
		defaultScriptID = authContainer.GgsrunCfg.Scriptid
	}

	modeName := "exe2"
	if isExe1 {
		modeName = "exe1"
	}

	promptTextInput(" ["+modeName+"] Script ID ", "Script ID: ", defaultScriptID, func(scriptID string) {
		if scriptID == "" {
			showError("Script ID is required.")
			return
		}

		onCollectArg := func(funcName string) {
			promptTitle := " [" + modeName + "] Argument Value (Optional) "
			if !isExe1 {
				promptTitle = " [" + modeName + " (main function only)] Argument Value (Optional) "
			}
			promptTextInput(promptTitle, "Argument Value: ", "", func(argVal string) {
				doExecute := func(undeleteScript bool, conflict string) {
					runTask("Executing script via "+modeName+" (Script ID: "+scriptID+")...", func() error {
						flags := map[string]string{
							"scriptid":   scriptID,
							"value":      argVal,
							"jsonparser": "true",
						}
						if conflict != "" {
							flags["conflict"] = conflict
						}
						if isExe1 {
							flags["function"] = funcName
							if undeleteScript {
								flags["undeleteScript"] = "true"
							}
						}

						if isLocal {
							if isMultiOrDir {
								flags["scriptfile"] = scriptFile
								if isExe1 && undeleteScript {
									flags["undeleteScript"] = "true"
								}
							} else {
								contentBytes, err := osReadFileFn(scriptFile)
								if err != nil {
									return fmt.Errorf("failed to read local script: %v", err)
								}
								flags["stringscript"] = string(contentBytes)
								if isExe1 && undeleteScript {
									flags["undeleteScript"] = "true"
								}
							}
						} else {
							opCtxForFetch := createOpContext(mainCtx, map[string]string{
								"scriptid": remoteScriptID,
							})
							projectContent := getBoundScriptExportedFn(authContainer, opCtxForFetch, remoteScriptID)
							if projectContent == nil || len(projectContent.Files) == 0 {
								return fmt.Errorf("failed to fetch remote GAS project content or project is empty")
							}
							var sb strings.Builder
							for _, f := range projectContent.Files {
								if f.Type == "SERVER_JS" {
									sb.WriteString(f.Source)
									sb.WriteString("\n")
								}
							}
							rawScript := sb.String()
							if rawScript == "" {
								return fmt.Errorf("no server script files found in the remote GAS project")
							}
							flags["stringscript"] = rawScript
							if isExe1 && undeleteScript {
								flags["undeleteScript"] = "true"
							}
						}

						opCtx := createOpContext(mainCtx, flags)
						var resp string
						var err error
						if isExe1 {
							resp, err = tuiRunExe1Fn(opCtx, authContainer)
						} else {
							resp, err = tuiRunExe2Fn(opCtx, authContainer)
						}

						tuiApp.QueueUpdateDraw(func() {
							showExecutionResult(funcName, resp, err)
						})
						return nil
					})
				}

				if isExe1 {
					prevFocus := tuiApp.GetFocus()

					showConflictModal := func(undeleteScript bool) {
						conflictModal := tview.NewModal().
							SetText("Duplicate filename conflict resolution strategy?\n\nChoose 'Overwrite' (Default) to replace existing script files in the remote GAS project, or 'Add' to upload as new files with unique names.").
							AddButtons([]string{"Overwrite", "Add"}).
							SetDoneFunc(func(buttonIndex int, buttonLabel string) {
								pages.RemovePage("conflict_choice")
								tuiApp.SetFocus(prevFocus)

								conflictChoice := "overwrite"
								if buttonLabel == "Add" {
									conflictChoice = "add"
								}
								doExecute(undeleteScript, conflictChoice)
							})
						pages.AddPage("conflict_choice", conflictModal, true, true)
						tuiApp.SetFocus(conflictModal)
					}

					cleanupModal := tview.NewModal().
						SetText("Clean up uploaded scripts on remote GAS project after execution? (-d)\n\nSelecting 'Yes' (Default) ensures your remote Google Apps Script project remains clean by removing any uploaded files after execution completes.").
						AddButtons([]string{"Yes", "No"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							pages.RemovePage("delete_script_confirm")
							tuiApp.SetFocus(prevFocus)

							undeleteScript := (buttonLabel == "No")
							showConflictModal(undeleteScript)
						})
					pages.AddPage("delete_script_confirm", cleanupModal, true, true)
					tuiApp.SetFocus(cleanupModal)
				} else {
					doExecute(false, "overwrite")
				}
			})
		}

		if isExe1 {
			promptTextInput(" ["+modeName+"] Function Name ", "Function: ", "main", func(funcName string) {
				if funcName == "" {
					funcName = "main"
				}
				onCollectArg(funcName)
			})
		} else {
			onCollectArg("main")
		}
	})
}

func collectWebAppsParams(isLocal bool, scriptFile string, remoteScriptID string, isMultiOrDir bool) {
	if isMultiOrDir {
		showError("Directory or multiple script execution is strictly limited to 'exe1'. It is not supported for 'webapps'.")
		return
	}

	defaultURL := ""
	if authContainer != nil && authContainer.GgsrunCfg != nil {
		defaultURL = authContainer.GgsrunCfg.WebappsUrl
	}

	promptTextInput(" [webapps] Web Apps URL ", "URL: ", defaultURL, func(url string) {
		if url == "" {
			showError("Web Apps URL is required.")
			return
		}

		promptTextInput(" [webapps] Password (Optional) ", "Password: ", "", func(password string) {
			promptTextInput(" [webapps (main function only)] Argument Value (Optional) ", "Argument Value: ", "", func(argVal string) {
				runTask("Executing script via webapps (URL: "+url+")...", func() error {
					flags := map[string]string{
						"url":        url,
						"password":   password,
						"value":      argVal,
						"jsonparser": "true",
					}

					if isLocal {
						contentBytes, err := osReadFileFn(scriptFile)
						if err != nil {
							return fmt.Errorf("failed to read local script: %v", err)
						}
						flags["stringscript"] = string(contentBytes)
					} else {
						targetScriptID := remoteScriptID
						if targetScriptID == "" && authContainer != nil && authContainer.GgsrunCfg != nil {
							targetScriptID = authContainer.GgsrunCfg.Scriptid
						}
						if targetScriptID != "" {
							opCtxForFetch := createOpContext(mainCtx, map[string]string{
								"scriptid": targetScriptID,
							})
							projectContent := getBoundScriptExportedFn(authContainer, opCtxForFetch, targetScriptID)
							if projectContent != nil && len(projectContent.Files) > 0 {
								var sb strings.Builder
								for _, f := range projectContent.Files {
									if f.Type == "SERVER_JS" {
										sb.WriteString(f.Source)
										sb.WriteString("\n")
									}
								}
								flags["stringscript"] = sb.String()
							}
						}
					}

					if remoteScriptID != "" {
						flags["scriptid"] = remoteScriptID
					} else if authContainer != nil && authContainer.GgsrunCfg != nil {
						flags["scriptid"] = authContainer.GgsrunCfg.Scriptid
					}

					opCtx := createOpContext(mainCtx, flags)
					resp, err := tuiRunWebAppsFn(opCtx, authContainer)

					tuiApp.QueueUpdateDraw(func() {
						showExecutionResult("webapps", resp, err)
					})
					return nil
				})
			})
		})
	})
}

func onExecuteLocalScript() {
	jobs := getSelectedJobs(true)
	if len(jobs) == 0 {
		showError("Please select a script file or directory to execute.")
		return
	}

	if len(jobs) == 1 {
		job := jobs[0]
		if job.IsDir {
			promptGasExecution(true, job.SourcePath, "", true)
			return
		}
		ext := strings.ToLower(filepath.Ext(job.SourceName))
		if ext != ".gs" && ext != ".js" && ext != ".txt" && ext != ".gas" {
			showError("Only Google Apps Script files (.gs, .js, .txt, .gas) can be executed.")
			return
		}
		promptGasExecution(true, job.SourcePath, "", false)
		return
	}

	// Multiple selected scripts
	var paths []string
	for _, job := range jobs {
		paths = append(paths, job.SourcePath)
	}
	joinedPaths := strings.Join(paths, ",")
	promptGasExecution(true, joinedPaths, "", true)
}

func onExecuteRemoteScript() {
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	selectedRow := getSelectedRowIndex(remoteTable)
	if selectedRow >= 0 && selectedRow < len(remoteFiles) {
		selected := remoteFiles[selectedRow]
		if selected.IsDir || selected.Name == ".." {
			showError("Please select a script file to execute.")
			return
		}
		if selected.MimeType != "application/vnd.google-apps.script" {
			showError("Only standalone Google Apps Script can be executed.")
			return
		}
		promptGasExecution(false, "", selected.Path, false)
	}
}

func showExecutionResult(funcName string, resp string, err error) {
	prevFocus := tuiApp.GetFocus()

	content := ""
	if err != nil {
		content = fmt.Sprintf("Error executing function:\n%v\n\nResponse details:\n%s", err, resp)
	} else {
		content = fmt.Sprintf("Function '%s' executed successfully!\n\nResponse JSON:\n%s", funcName, resp)
	}

	textView := tview.NewTextView().
		SetText(content).
		SetDynamicColors(false)

	textView.SetBorder(true).
		SetTitle(" Execution Result: " + funcName + " ").
		SetTitleColor(tcell.ColorGreen)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 15, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(textView, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	flex.AddItem(inner, 0, 70, true)
	flex.AddItem(tview.NewBox(), 0, 15, false)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
			pages.RemovePage("execution_result")
			if activeTableBeforeAction != nil {
				tuiApp.SetFocus(activeTableBeforeAction)
			} else if lastActiveTable != nil {
				tuiApp.SetFocus(lastActiveTable)
			} else {
				tuiApp.SetFocus(prevFocus)
			}
			return nil
		}
		return event
	})

	pages.AddPage("execution_result", flex, true, true)
	tuiApp.SetFocus(textView)
}

func showFileDetails() {
	var selected FileEntry
	var isLocal bool

	if localTable.HasFocus() {
		isLocal = true
		selectedRow := getSelectedRowIndex(localTable)
		if selectedRow >= 0 && selectedRow < len(localFiles) {
			selected = localFiles[selectedRow]
		}
	} else {
		isLocal = false
		selectedRow := getSelectedRowIndex(remoteTable)
		if selectedRow >= 0 && selectedRow < len(remoteFiles) {
			selected = remoteFiles[selectedRow]
		}
	}

	if selected.Name == "" || selected.Name == ".." {
		return
	}

	prevFocus := tuiApp.GetFocus()

	if isLocal {
		extra := getPlatformDetails(selected.Path)
		details := fmt.Sprintf(`  Name          : %s
  Local Path    : %s
  MimeType      : %s
  Permissions   : %s
  Created Date  : %s
  Last Modified : %s
  Size          : %s%s
`, selected.Name, selected.Path, selected.MimeType, selected.Permissions, selected.CreatedTime, selected.ModTime, formatSize(selected.Size, selected.IsDir, selected.MimeType), extra)

		textView := tview.NewTextView().
			SetText("Local File Details\n\n" + details).
			SetDynamicColors(false)
		textView.SetBorder(true).
			SetTitle(" Details ").
			SetTitleColor(tcell.ColorGreen)

		flex := tview.NewFlex().SetDirection(tview.FlexRow)
		flex.AddItem(tview.NewBox(), 0, 15, false)

		inner := tview.NewFlex().SetDirection(tview.FlexColumn)
		inner.AddItem(tview.NewBox(), 0, 15, false)
		inner.AddItem(textView, 0, 70, true)
		inner.AddItem(tview.NewBox(), 0, 15, false)

		flex.AddItem(inner, 0, 70, true)
		flex.AddItem(tview.NewBox(), 0, 15, false)

		textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
				pages.RemovePage("details")
				tuiApp.SetFocus(prevFocus)
				return nil
			}
			return event
		})

		pages.AddPage("details", flex, true, true)
		tuiApp.SetFocus(textView)
	} else {
		runTask("Fetching file details...", func() error {
			r := &utl.RequestParams{
				Method:      "GET",
				APIURL:      "https://www.googleapis.com/drive/v3/files/" + selected.Path + "?fields=id,name,mimeType,size,modifiedTime,createdTime,description,owners(displayName,emailAddress),shared,webViewLink",
				Accesstoken: authContainer.GgsrunCfg.Accesstoken,
				Dtime:       30,
			}
			body, err := requestParamsFetchFn(r)
			if err != nil {
				return err
			}

			var meta struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				MimeType     string `json:"mimeType"`
				Size         string `json:"size"`
				ModifiedTime string `json:"modifiedTime"`
				CreatedTime  string `json:"createdTime"`
				Description  string `json:"description"`
				Shared       bool   `json:"shared"`
				Owners       []struct {
					DisplayName  string `json:"displayName"`
					EmailAddress string `json:"emailAddress"`
				} `json:"owners"`
				WebViewLink  string `json:"webViewLink"`
			}
			_ = json.Unmarshal(body, &meta)

			sizeVal, _ := strconv.ParseInt(meta.Size, 10, 64)
			isDir := meta.MimeType == "application/vnd.google-apps.folder"

			ownerInfo := ""
			if len(meta.Owners) > 0 {
				ownerInfo = meta.Owners[0].DisplayName + " (" + meta.Owners[0].EmailAddress + ")"
			}

			createdFormatted := ""
			if t, err := time.Parse(time.RFC3339, meta.CreatedTime); err == nil {
				createdFormatted = t.In(time.Local).Format("2006-01-02 15:04:05")
			}
			modifiedFormatted := ""
			if t, err := time.Parse(time.RFC3339, meta.ModifiedTime); err == nil {
				modifiedFormatted = t.In(time.Local).Format("2006-01-02 15:04:05")
			}

			sharedStr := "Private"
			if meta.Shared {
				sharedStr = "Shared"
			}

			details := fmt.Sprintf(`  Name          : %s
  File ID       : %s
  Simple Type   : %s
  Full MimeType : %s
  Created Date  : %s
  Last Modified : %s
  Size          : %s
  Owner         : %s
  Sharing State : %s
  Description   : %s
  Web View Link : %s
`, meta.Name, meta.ID, simplifyMimeType(meta.MimeType, isDir), meta.MimeType, createdFormatted, modifiedFormatted, formatSize(sizeVal, isDir, meta.MimeType), ownerInfo, sharedStr, meta.Description, meta.WebViewLink)

			tuiApp.QueueUpdateDraw(func() {
				textView := tview.NewTextView().
					SetText("Google Drive File Details\n\n" + details).
					SetDynamicColors(false)
				textView.SetBorder(true).
					SetTitle(" Details ").
					SetTitleColor(tcell.ColorGreen)

				flex := tview.NewFlex().SetDirection(tview.FlexRow)
				flex.AddItem(tview.NewBox(), 0, 15, false)

				inner := tview.NewFlex().SetDirection(tview.FlexColumn)
				inner.AddItem(tview.NewBox(), 0, 15, false)
				inner.AddItem(textView, 0, 70, true)
				inner.AddItem(tview.NewBox(), 0, 15, false)

				flex.AddItem(inner, 0, 70, true)
				flex.AddItem(tview.NewBox(), 0, 15, false)

				textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
						pages.RemovePage("details")
						tuiApp.SetFocus(prevFocus)
						return nil
					}
					return event
				})

				pages.AddPage("details", flex, true, true)
				tuiApp.SetFocus(textView)
			})
			return nil
		})
	}
}

func getConvertOptions(srcMime string) map[string]string {
	options := make(map[string]string)
	switch srcMime {
	case "application/vnd.google-apps.document":
		options["Microsoft Word (.docx)"] = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		options["PDF Document (.pdf)"] = "application/pdf"
		options["Rich Text (.rtf)"] = "application/rtf"
		options["Plain Text (.txt)"] = "text/plain"
		options["ZIP Archive (.zip)"] = "application/zip"
	case "application/vnd.google-apps.spreadsheet":
		options["Microsoft Excel (.xlsx)"] = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		options["PDF Document (.pdf)"] = "application/pdf"
		options["CSV (.csv)"] = "text/csv"
		options["TSV (.tsv)"] = "text/tab-separated-values"
		options["ZIP Archive (.zip)"] = "application/zip"
	case "application/vnd.google-apps.presentation":
		options["Microsoft PowerPoint (.pptx)"] = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		options["PDF Document (.pdf)"] = "application/pdf"
		options["ZIP Archive (.zip)"] = "application/zip"
	default:
		targets := utl.GetImportTargets(srcMime)
		if len(targets) > 0 {
			for _, target := range targets {
				switch target {
				case "application/vnd.google-apps.document":
					options["Google Docs (.gdoc)"] = target
				case "application/vnd.google-apps.spreadsheet":
					options["Google Sheets (.gsheet)"] = target
				case "application/vnd.google-apps.presentation":
					options["Google Slides (.gslides)"] = target
				case "application/vnd.google-apps.drawing":
					options["Google Drawings (.gdraw)"] = target
				case "application/vnd.google-apps.vid":
					options["Google Video (.gvid)"] = target
				case "application/vnd.google-apps.script":
					options["Google Apps Script (.gas)"] = target
				default:
					options[target] = target
				}
			}
		}
	}
	return options
}

func startTransferSequence(jobs []TransferJob, isUpload bool, isMove bool) {
	var finalJobs []TransferJob
	var collectNext func(index int)

	collectNext = func(index int) {
		if index >= len(jobs) {
			runBatchTransfer(finalJobs, isUpload, isMove)
			return
		}

		job := jobs[index]
		if job.IsDir {
			finalJobs = append(finalJobs, job)
			collectNext(index + 1)
			return
		}

		options := make(map[string]string)
		var promptTitle string
		if isUpload {
			options = getConvertOptions(job.MimeType)
			promptTitle = " Convert MimeType for Upload: " + job.SourceName + " "
		} else {
			if strings.Contains(job.MimeType, "application/vnd.google-apps") {
				options = getConvertOptions(job.MimeType)
				promptTitle = " Select Export Format: " + job.SourceName + " "
			}
		}

		if len(options) > 0 {
			prevFocus := tuiApp.GetFocus()
			list := tview.NewList()
			if isUpload {
				list.AddItem("Upload as is (No conversion)", "", '0', nil)
			}

			keys := []string{}
			for k, v := range options {
				shortcut := rune(49 + len(keys))
				list.AddItem(k, v, shortcut, nil)
				keys = append(keys, v)
			}

			list.SetBorder(true).SetTitle(promptTitle).SetTitleColor(tcell.ColorYellow)

			flex := tview.NewFlex().SetDirection(tview.FlexRow)
			flex.AddItem(tview.NewBox(), 0, 1, false)

			inner := tview.NewFlex().SetDirection(tview.FlexColumn)
			inner.AddItem(tview.NewBox(), 0, 15, false)
			inner.AddItem(list, 0, 70, true)
			inner.AddItem(tview.NewBox(), 0, 15, false)

			flex.AddItem(inner, 10, 0, true)
			flex.AddItem(tview.NewBox(), 0, 1, false)

			list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					pages.RemovePage("convert_prompt")
					if lastActiveTable != nil {
						tuiApp.SetFocus(lastActiveTable)
					} else {
						tuiApp.SetFocus(prevFocus)
					}
					return nil
				}
				return event
			})

			list.SetSelectedFunc(func(itemIndex int, mainText string, secondaryText string, shortcut rune) {
				pages.RemovePage("convert_prompt")
				if lastActiveTable != nil {
					tuiApp.SetFocus(lastActiveTable)
				} else {
					tuiApp.SetFocus(prevFocus)
				}

				targetMime := secondaryText
				if isUpload && itemIndex == 0 {
					targetMime = ""
				}

				job.TargetMime = targetMime
				finalJobs = append(finalJobs, job)
				collectNext(index + 1)
			})

			pages.AddPage("convert_prompt", flex, true, true)
			tuiApp.SetFocus(list)
		} else {
			if !isUpload && strings.Contains(job.MimeType, "application/vnd.google-apps") {
				job.TargetMime = "application/pdf" // Google独自形式のデフォルトエクスポート
			}
			finalJobs = append(finalJobs, job)
			collectNext(index + 1)
		}
	}

	collectNext(0)
}

func runBatchTransfer(jobs []TransferJob, isUpload bool, isMove bool) {
	if len(jobs) == 0 {
		return
	}

	msg := ""
	if isUpload {
		msg = fmt.Sprintf("Uploading %d item(s) in parallel...", len(jobs))
	} else {
		msg = fmt.Sprintf("Downloading %d item(s) in parallel...", len(jobs))
	}

	runTask(msg, func() error {
		// Initialize progress map
		transferProgressMu.Lock()
		transferProgress = make(map[string]jobProgress)
		for _, j := range jobs {
			transferProgress[j.SourceName] = jobProgress{
				Status: "Pending",
			}
		}
		transferLogLines = nil
		transferHeading = msg
		transferProgressMu.Unlock()

		defer func() {
			transferProgressMu.Lock()
			transferProgress = nil
			transferLogLines = nil
			transferProgressMu.Unlock()
		}()

		// Generate directory tree preview for any directory jobs
		var treeLines []string
		for _, job := range jobs {
			if job.IsDir {
				treeLines = append(treeLines, fmt.Sprintf("Directory Tree for '%s':", job.SourceName))
				if isUpload {
					lines, err := generateLocalTree(job.SourcePath, "  ")
					if err == nil {
						treeLines = append(treeLines, lines...)
					} else {
						treeLines = append(treeLines, "  Error generating tree: "+err.Error())
					}
				} else {
					lines, err := generateRemoteTreeFn(authContainer, mainCtx, job.SourcePath, "  ")
					if err == nil {
						treeLines = append(treeLines, lines...)
					} else {
						treeLines = append(treeLines, "  Error generating tree: "+err.Error())
					}
				}
				treeLines = append(treeLines, "")
			}
		}

		if len(treeLines) > 0 {
			tuiApp.QueueUpdateDraw(func() {
				loadingViewMu.Lock()
				defer loadingViewMu.Unlock()
				if currentLoadingView != nil {
					for _, line := range treeLines {
						fmt.Fprintln(currentLoadingView, line)
					}
					currentLoadingView.ScrollToEnd()
				}
			})
			time.Sleep(2 * time.Second)
		}

		errChan := make(chan error, len(jobs))

		for _, job := range jobs {
			go func(j TransferJob) {
				var err error
				var res interface{}
				if isUpload {
					flags := map[string]string{
						"filename":       j.SourcePath,
						"parentfolderid": currentRemoteFolderID,
						"projecttype":    "standalone",
						"jsonparser":     "true",
					}
					if j.TargetMime != "" {
						flags["convertto"] = j.TargetMime
					} else {
						flags["noconvert"] = "true"
					}
					opCtx := createOpContext(mainCtx, flags)
					res, err = tuiUploadFn(opCtx, authContainer)
					if err == nil {
						if tr, ok := res.(app.TransferResult); ok {
							for _, f := range tr.Files {
								if strings.HasPrefix(f.Status, "failed") {
									err = fmt.Errorf("upload failed for '%s': %s", f.Name, f.Status)
									break
								}
							}
						} else if fi, ok := res.(*utl.FileInf); ok {
							for _, msg := range fi.Msgar {
								if strings.HasPrefix(strings.ToLower(msg), "error") {
									err = fmt.Errorf("%s", msg)
									break
								}
							}
						}
					}
				} else {
					flags := map[string]string{
						"fileid":      j.SourcePath,
						"destination": currentLocalDir,
						"jsonparser":  "true",
					}
					if j.TargetMime != "" {
						ext := utl.MimeToExt(j.TargetMime)
						if ext != "" {
							flags["extension"] = ext
						}
					}
					opCtx := createOpContext(mainCtx, flags)
					res, err = tuiDownloadFn(opCtx, authContainer)
					if err == nil {
						if tr, ok := res.(app.TransferResult); ok {
							for _, f := range tr.Files {
								if strings.HasPrefix(f.Status, "failed") {
									err = fmt.Errorf("download failed for '%s': %s", f.Name, f.Status)
									break
								}
							}
						} else if fi, ok := res.(*utl.FileInf); ok {
							for _, msg := range fi.Msgar {
								if strings.HasPrefix(strings.ToLower(msg), "error") {
									err = fmt.Errorf("%s", msg)
									break
								}
							}
						}
					}
				}
				errChan <- err
			}(job)
		}

		var firstErr error
		for i := 0; i < len(jobs); i++ {
			if err := <-errChan; err != nil && firstErr == nil {
				firstErr = err
			}
		}

		// Only perform deletions if all transfers succeeded
		if firstErr == nil && isMove {
			for _, j := range jobs {
				if isUpload {
					err := os.RemoveAll(j.SourcePath)
					if err != nil && firstErr == nil {
						firstErr = err
					}
				} else {
					err := deleteRemoteFileFn(authContainer, mainCtx, j.SourcePath)
					if err != nil && firstErr == nil {
						firstErr = err
					}
				}
			}
		}

		tuiApp.QueueUpdate(func() {
			selectedLocalPaths = make(map[string]bool)
			selectedRemoteIDs = make(map[string]bool)
		})

		return firstErr
	})
}

func copyLocal(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyLocalDir(src, dst)
	}
	return copyLocalFile(src, dst)
}

func copyLocalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyLocalDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		err = copyLocal(s, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func moveDriveFile(fileID string, newParentID string, a *app.AuthContainer) error {
	rMeta := &utl.RequestParams{
		Method:      "GET",
		APIURL:      "https://www.googleapis.com/drive/v3/files/" + fileID + "?fields=parents",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	metaBody, err := requestParamsFetchFn(rMeta)
	if err != nil {
		return err
	}
	var meta struct {
		Parents []string `json:"parents"`
	}
	_ = json.Unmarshal(metaBody, &meta)

	removeParents := strings.Join(meta.Parents, ",")

	urlStr := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?addParents=%s", fileID, newParentID)
	if removeParents != "" {
		urlStr += "&removeParents=" + removeParents
	}

	r := &utl.RequestParams{
		Method:      "PATCH",
		APIURL:      urlStr,
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	_, err = requestParamsFetchFn(r)
	return err
}

func exportRemoteFile(fileID string, fileName string, targetMime string, localDest string, a *app.AuthContainer) error {
	var urlStr string
	var err error
	isGoogleDoc := false

	rMeta := &utl.RequestParams{
		Method:      "GET",
		APIURL:      "https://www.googleapis.com/drive/v3/files/" + fileID + "?fields=mimeType,name",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	metaBody, err := requestParamsFetchFn(rMeta)
	if err == nil {
		var meta struct {
			MimeType string `json:"mimeType"`
			Name     string `json:"name"`
		}
		_ = json.Unmarshal(metaBody, &meta)
		isGoogleDoc = strings.Contains(meta.MimeType, "application/vnd.google-apps") && meta.MimeType != "application/vnd.google-apps.folder"
	}

	if isGoogleDoc {
		urlStr = "https://www.googleapis.com/drive/v3/files/" + fileID + "/export?mimeType=" + targetMime
	} else {
		urlStr = "https://www.googleapis.com/drive/v3/files/" + fileID + "?alt=media"
	}

	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      urlStr,
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       60,
	}
	res, err := requestParamsFetchRawFn(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := os.Create(localDest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}

func convertDriveFileInPlace(fileID string, targetMime string, a *app.AuthContainer) error {
	rMeta := &utl.RequestParams{
		Method:      "GET",
		APIURL:      "https://www.googleapis.com/drive/v3/files/" + fileID + "?fields=name,parents,mimeType",
		Accesstoken: a.GgsrunCfg.Accesstoken,
		Dtime:       30,
	}
	metaBody, err := requestParamsFetchFn(rMeta)
	if err != nil {
		return err
	}
	var meta struct {
		Name     string   `json:"name"`
		Parents  []string `json:"parents"`
		MimeType string   `json:"mimeType"`
	}
	_ = json.Unmarshal(metaBody, &meta)

	parentID := "root"
	if len(meta.Parents) > 0 {
		parentID = meta.Parents[0]
	}

	ext := utl.MimeToExt(targetMime)
	if ext == "" {
		ext = "tmp"
	}
	tempName := meta.Name + "_converted." + ext
	tempPath := filepath.Join(os.TempDir(), tempName)

	err = exportRemoteFile(fileID, meta.Name, targetMime, tempPath, a)
	if err != nil {
		return err
	}
	defer os.Remove(tempPath)

	opCtx := createOpContext(mainCtx, map[string]string{
		"filename":       tempPath,
		"parentfolderid": parentID,
		"jsonparser":     "true",
	})

	_, err = tuiUploadFn(opCtx, a)
	return err
}

func promptTextInput(title, label, defaultText string, onDone func(text string)) {
	prevFocus := tuiApp.GetFocus()

	inputField := tview.NewInputField().
		SetLabel(label).
		SetText(defaultText).
		SetFieldWidth(0)

	inputField.SetBorder(true).
		SetTitle(title).
		SetTitleColor(tcell.ColorYellow)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(inputField, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	flex.AddItem(inner, 3, 0, true)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	inputField.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage("text_prompt")
		tuiApp.SetFocus(prevFocus)
		if key == tcell.KeyEnter {
			onDone(inputField.GetText())
		}
	})

	pages.AddPage("text_prompt", flex, true, true)
	tuiApp.SetFocus(inputField)
}

func copyToClipboard(text string) error {
	if isWSL() {
		cmd := exec.Command("/mnt/c/Windows/System32/clip.exe")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, text)
		}()
		return cmd.Run()
	}

	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, text)
		}()
		return cmd.Run()
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, text)
		}()
		return cmd.Run()
	}

	return fmt.Errorf("no clipboard utility found (clip.exe, xclip, or xsel)")
}

func onEnterLocal() {
	selectedRow := getSelectedRowIndex(localTable)
	if selectedRow >= 0 && selectedRow < len(localFiles) {
		selected := localFiles[selectedRow]
		if selected.IsDir {
			inSearchModeLocal = false
			if selected.Name == ".." {
				currentLocalDir = filepath.Dir(currentLocalDir)
			} else {
				currentLocalDir = selected.Path
			}
			refreshPanels()
			if len(localFiles) > 0 {
				localTable.Select(1, 0)
			}
		} else {
			content, err := readTextPreview(selected.Path)
			if err != nil {
				showError("Cannot preview binary/empty file: " + err.Error())
			} else {
				showTextPreview(selected.Name, content)
			}
		}
	}
}

func onEnterRemote() {
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	selectedRow := getSelectedRowIndex(remoteTable)
	if selectedRow >= 0 && selectedRow < len(remoteFiles) {
		selected := remoteFiles[selectedRow]
		if selected.IsDir {
			inSearchModeRemote = false
			if selected.Name == ".." {
				if len(remoteFolderStack) > 1 {
					remoteFolderStack = remoteFolderStack[:len(remoteFolderStack)-1]
					currentRemoteFolderID = remoteFolderStack[len(remoteFolderStack)-1].ID
				}
			} else {
				remoteFolderStack = append(remoteFolderStack, FolderInfo{ID: selected.Path, Name: selected.Name})
				currentRemoteFolderID = selected.Path
			}
			refreshPanels()
			if len(remoteFiles) > 0 {
				remoteTable.Select(1, 0)
			}
		} else if selected.MimeType == "application/vnd.google-apps.script" {
			showGasProject(selected.Name, selected.Path)
		} else if strings.HasPrefix(selected.MimeType, "text/") ||
			selected.MimeType == "application/json" ||
			selected.MimeType == "application/javascript" ||
			selected.MimeType == "application/x-javascript" ||
			selected.MimeType == "application/xml" {

			var fileContent []byte
			runTask("Downloading text file content...", func() error {
				var err error
				fileContent, err = app.TuiGetFileContent(selected.Path, authContainer)
				if err != nil {
					return err
				}
				tuiApp.QueueUpdateDraw(func() {
					showTextPreview(selected.Name, string(fileContent))
				})
				return nil
			})
		} else {
			if selected.WebViewLink != "" {
				err := openBrowserFn(selected.WebViewLink)
				if err != nil {
					showError("Failed to open browser: " + err.Error())
				}
			} else {
				showError("No WebViewLink available for this file.")
			}
		}
	}
}

func onCopy() {
	clearSearchModes()
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	if localTable.HasFocus() {
		jobs := getSelectedJobs(true)
		if len(jobs) == 0 {
			return
		}
		startTransferSequence(jobs, true, false)
	} else {
		jobs := getSelectedJobs(false)
		if len(jobs) == 0 {
			return
		}
		startTransferSequence(jobs, false, false)
	}
}

func onMove() {
	clearSearchModes()
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	if localTable.HasFocus() {
		jobs := getSelectedJobs(true)
		if len(jobs) == 0 {
			return
		}
		startTransferSequence(jobs, true, true)
	} else {
		jobs := getSelectedJobs(false)
		if len(jobs) == 0 {
			return
		}
		startTransferSequence(jobs, false, true)
	}
}

func onRename() {
	clearSearchModes()
	var selected FileEntry
	var isLocal bool

	if localTable.HasFocus() {
		isLocal = true
		selectedRow := getSelectedRowIndex(localTable)
		if selectedRow >= 0 && selectedRow < len(localFiles) {
			selected = localFiles[selectedRow]
		}
	} else {
		if !isAuthorized {
			showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
			return
		}
		isLocal = false
		selectedRow := getSelectedRowIndex(remoteTable)
		if selectedRow >= 0 && selectedRow < len(remoteFiles) {
			selected = remoteFiles[selectedRow]
		}
	}

	if selected.Name == "" || selected.Name == ".." {
		return
	}

	promptTextInput(" Rename ", "New Name: ", selected.Name, func(newName string) {
		if newName == "" || newName == selected.Name {
			return
		}

		if isLocal {
			runTask("Renaming local file...", func() error {
				dir := filepath.Dir(selected.Path)
				return os.Rename(selected.Path, filepath.Join(dir, newName))
			})
		} else {
			runTask("Renaming Drive file...", func() error {
				return tuiUpdateDriveMetadataFn(selected.Path, newName, "", nil, authContainer)
			})
		}
	})
}

func onChangeTimestamp() {
	var selected FileEntry
	var isLocal bool

	if localTable.HasFocus() {
		isLocal = true
		selectedRow := getSelectedRowIndex(localTable)
		if selectedRow >= 0 && selectedRow < len(localFiles) {
			selected = localFiles[selectedRow]
		}
	} else {
		if !isAuthorized {
			showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
			return
		}
		isLocal = false
		selectedRow := getSelectedRowIndex(remoteTable)
		if selectedRow >= 0 && selectedRow < len(remoteFiles) {
			selected = remoteFiles[selectedRow]
		}
	}

	if selected.Name == "" || selected.Name == ".." {
		return
	}

	promptTextInput(" Change Timestamp ", "New Time (YYYY-MM-DD HH:MM:SS): ", selected.ModTime, func(newTimeStr string) {
		if newTimeStr == "" {
			return
		}

		parsedTime, err := time.Parse("2006-01-02 15:04:05", newTimeStr)
		if err != nil {
			showError("Invalid time format. Use YYYY-MM-DD HH:MM:SS")
			return
		}

		if isLocal {
			runTask("Updating local timestamp...", func() error {
				return os.Chtimes(selected.Path, time.Now(), parsedTime)
			})
		} else {
			runTask("Updating Drive timestamp...", func() error {
				return tuiUpdateDriveMetadataFn(selected.Path, "", "", &parsedTime, authContainer)
			})
		}
	})
}

func onChangeDescription() {
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	if localTable.HasFocus() {
		showError("Description is only supported on Google Drive files.")
		return
	}

	selectedRow := getSelectedRowIndex(remoteTable)
	if selectedRow < 0 || selectedRow >= len(remoteFiles) {
		return
	}
	selected := remoteFiles[selectedRow]
	if selected.Name == "" || selected.Name == ".." {
		return
	}

	runTask("Fetching description...", func() error {
		r := &utl.RequestParams{
			Method:      "GET",
			APIURL:      "https://www.googleapis.com/drive/v3/files/" + selected.Path + "?fields=description",
			Accesstoken: authContainer.GgsrunCfg.Accesstoken,
			Dtime:       30,
		}
		body, err := requestParamsFetchFn(r)
		if err != nil {
			return err
		}

		var meta struct {
			Description string `json:"description"`
		}
		_ = json.Unmarshal(body, &meta)

		tuiApp.QueueUpdateDraw(func() {
			promptTextInput(" Edit Description ", "Description: ", meta.Description, func(newDesc string) {
				runTask("Updating description...", func() error {
					return tuiUpdateDriveMetadataFn(selected.Path, "", newDesc, nil, authContainer)
				})
			})
		})

		return nil
	})
}

func onConvertMimeType() {
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	if localTable.HasFocus() {
		showError("Conversion in same folder is only supported on Google Drive.")
		return
	}

	selectedRow := getSelectedRowIndex(remoteTable)
	if selectedRow < 0 || selectedRow >= len(remoteFiles) {
		return
	}
	selected := remoteFiles[selectedRow]
	if selected.Name == "" || selected.Name == ".." {
		return
	}

	options := getConvertOptions(selected.MimeType)
	if len(options) == 0 {
		showError("No conversion options available for this file type.")
		return
	}

	prevFocus := tuiApp.GetFocus()
	list := tview.NewList()

	promptTitle := " Convert & Save: " + selected.Name + " "
	keys := []string{}
	for k, v := range options {
		shortcut := rune(49 + len(keys))
		list.AddItem(k, v, shortcut, nil)
		keys = append(keys, v)
	}

	list.SetBorder(true).SetTitle(promptTitle).SetTitleColor(tcell.ColorYellow)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(list, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	flex.AddItem(inner, 10, 0, true)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("convert_choice")
			tuiApp.SetFocus(prevFocus)
			return nil
		}
		return event
	})

	list.SetSelectedFunc(func(itemIndex int, mainText string, secondaryText string, shortcut rune) {
		pages.RemovePage("convert_choice")
		tuiApp.SetFocus(prevFocus)

		targetMime := secondaryText
		runTask("Converting file in place...", func() error {
			return convertDriveFileInPlace(selected.Path, targetMime, authContainer)
		})
	})

	pages.AddPage("convert_choice", flex, true, true)
	tuiApp.SetFocus(list)
}

func onDomainCopy() {
	if !isAuthorized {
		showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
		return
	}
	var selected FileEntry
	var isLocal bool

	if localTable.HasFocus() {
		isLocal = true
		selectedRow := getSelectedRowIndex(localTable)
		if selectedRow >= 0 && selectedRow < len(localFiles) {
			selected = localFiles[selectedRow]
		}
	} else {
		isLocal = false
		selectedRow := getSelectedRowIndex(remoteTable)
		if selectedRow >= 0 && selectedRow < len(remoteFiles) {
			selected = remoteFiles[selectedRow]
		}
	}

	if selected.Name == "" || selected.Name == ".." {
		return
	}

	if isLocal {
		promptTextInput(" Copy Local Item ", "Destination Path: ", selected.Path+"_copy", func(destPath string) {
			if destPath == "" || destPath == selected.Path {
				return
			}
			runTask("Copying local item...", func() error {
				return copyLocal(selected.Path, destPath)
			})
		})
	} else {
		promptTextInput(" Copy Drive File ", "New Name for Copy: ", selected.Name+"_copy", func(newName string) {
			if newName == "" {
				return
			}
			runTask("Copying Drive file...", func() error {
				rMeta := &utl.RequestParams{
					Method:      "GET",
					APIURL:      "https://www.googleapis.com/drive/v3/files/" + selected.Path + "?fields=parents",
					Accesstoken: authContainer.GgsrunCfg.Accesstoken,
					Dtime:       30,
				}
				metaBody, err := requestParamsFetchFn(rMeta)
				if err != nil {
					return err
				}
				var meta struct {
					Parents []string `json:"parents"`
				}
				_ = json.Unmarshal(metaBody, &meta)

				parentID := "root"
				if len(meta.Parents) > 0 {
					parentID = meta.Parents[0]
				}

				_, err = tuiCopyDriveFileFn(selected.Path, newName, parentID, authContainer)
				return err
			})
		})
	}
}

func onDomainMove() {
	var selected FileEntry
	var isLocal bool

	if localTable.HasFocus() {
		isLocal = true
		selectedRow := getSelectedRowIndex(localTable)
		if selectedRow >= 0 && selectedRow < len(localFiles) {
			selected = localFiles[selectedRow]
		}
	} else {
		if !isAuthorized {
			showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
			return
		}
		isLocal = false
		selectedRow := getSelectedRowIndex(remoteTable)
		if selectedRow >= 0 && selectedRow < len(remoteFiles) {
			selected = remoteFiles[selectedRow]
		}
	}

	if selected.Name == "" || selected.Name == ".." {
		return
	}

	if isLocal {
		promptTextInput(" Move Local Item ", "New Path: ", selected.Path, func(destPath string) {
			if destPath == "" || destPath == selected.Path {
				return
			}
			runTask("Moving local item...", func() error {
				err := os.Rename(selected.Path, destPath)
				if err != nil {
					err = copyLocal(selected.Path, destPath)
					if err == nil {
						err = os.RemoveAll(selected.Path)
					}
				}
				return err
			})
		})
	} else {
		promptTextInput(" Move Drive Item ", "New Parent Folder ID: ", currentRemoteFolderID, func(newParentID string) {
			if newParentID == "" || newParentID == currentRemoteFolderID {
				return
			}
			runTask("Moving Drive item...", func() error {
				return moveDriveFile(selected.Path, newParentID, authContainer)
			})
		})
	}
}

func onDelete() {
	clearSearchModes()
	activeTableBeforeAction = lastActiveTable
	var jobs []TransferJob
	var isLocal bool

	if localTable.HasFocus() {
		isLocal = true
		jobs = getSelectedJobs(true)
	} else {
		if !isAuthorized {
			showError("Action unavailable. Please authenticate with 'ggsrun auth' first.")
			return
		}
		isLocal = false
		jobs = getSelectedJobs(false)
	}

	if len(jobs) == 0 {
		return
	}

	prevFocus := tuiApp.GetFocus()

	promptText := fmt.Sprintf("Are you sure you want to delete %d item(s)?", len(jobs))
	if len(jobs) == 1 {
		promptText = "Are you sure you want to delete '" + jobs[0].SourceName + "'?"
	}

	confirmModal := tview.NewModal().
		SetText(promptText).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("delete_confirm")
			if activeTableBeforeAction != nil {
				tuiApp.SetFocus(activeTableBeforeAction)
			} else if lastActiveTable != nil {
				tuiApp.SetFocus(lastActiveTable)
			} else {
				tuiApp.SetFocus(prevFocus)
			}

			if buttonLabel == "Yes" {
				runTask("Deleting items...", func() error {
					errChan := make(chan error, len(jobs))
					for _, job := range jobs {
						go func(j TransferJob) {
							var err error
							if isLocal {
								err = os.RemoveAll(j.SourcePath)
							} else {
								err = deleteRemoteFileFn(authContainer, mainCtx, j.SourcePath)
							}
							errChan <- err
						}(job)
					}

					var firstErr error
					for i := 0; i < len(jobs); i++ {
						if err := <-errChan; err != nil && firstErr == nil {
							firstErr = err
						}
					}

					tuiApp.QueueUpdate(func() {
						selectedLocalPaths = make(map[string]bool)
						selectedRemoteIDs = make(map[string]bool)
					})

					return firstErr
				})
			}
		})

	pages.AddPage("delete_confirm", confirmModal, true, true)
	tuiApp.SetFocus(confirmModal)
}

func promptExit() {
	prevFocus := tuiApp.GetFocus()
	if pages.HasPage("exit_confirm") {
		return
	}

	confirmModal := tview.NewModal().
		SetText("Are you sure you want to exit? (Y/N)").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("exit_confirm")
			if buttonLabel == "Yes" {
				tuiApp.Stop()
			} else {
				tuiApp.SetFocus(prevFocus)
			}
		})

	confirmModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'y', 'Y':
				pages.RemovePage("exit_confirm")
				tuiApp.Stop()
				return nil
			case 'n', 'N':
				pages.RemovePage("exit_confirm")
				tuiApp.SetFocus(prevFocus)
				return nil
			}
		}
		return event
	})

	pages.AddPage("exit_confirm", confirmModal, true, true)
	tuiApp.SetFocus(confirmModal)
}

func showHelp() {
	prevFocus := tuiApp.GetFocus()

	helpText := `            ggsrun TUI File Manager (FD mode) Help
            ======================================

  Tab         : Switch focus between Upper (Local) and Lower (Remote) panels
  Space       : Toggle selection (multi-select)
  Up/Down     : Navigate file list
  Enter       : If folder: Enter directory and refresh panel
                If local text file: View content preview
                If GAS Script (Drive): Open source code explorer & run functions
                If other Drive File: Open in browser (WSL2 optimized)
  F1          : Copy selected item(s) to opposite panel's directory
  F2          : Move selected item(s) to opposite panel's directory
  F3          : Delete selected item(s)
  F5          : Create new directory/folder
  F8          : Search files or folders
  c           : Copy item within same panel (Local-to-Local or Drive-to-Drive)
  m           : Move item within same panel (Local-to-Local or Drive-to-Drive)
  n           : Rename item
  t           : Change timestamp (Last Modified) of item
  d           : Edit description (Drive only)
  x           : Convert MimeType and save as new file in same folder (Drive only)
  e           : Execute selected Google Apps Script (choose exe1/exe2/webapps)
  i           : Show detailed metadata / information of item
  s           : Sort files (choose sort key and order)
  y           : Copy selected item's absolute path (local) or file ID (remote) to clipboard
  r           : Refresh panels (Sync remote/local state)
  q           : Exit FD mode safely

  ----------------------------------------------------------------------
  Press Esc or Enter to close this help window.`

	textView := tview.NewTextView().
		SetText(helpText).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	textView.SetBorder(true).
		SetTitle(" Help / Operation Guide ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderPadding(1, 1, 2, 2)

	helpFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	helpFlex.AddItem(tview.NewBox(), 0, 1, false)

	innerFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	innerFlex.AddItem(tview.NewBox(), 0, 15, false)
	innerFlex.AddItem(textView, 0, 70, true)
	innerFlex.AddItem(tview.NewBox(), 0, 15, false)

	helpFlex.AddItem(innerFlex, 24, 0, true)
	helpFlex.AddItem(tview.NewBox(), 0, 1, false)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
			pages.RemovePage("help")
			tuiApp.SetFocus(prevFocus)
			return nil
		}
		return event
	})

	pages.AddPage("help", helpFlex, true, true)
	tuiApp.SetFocus(textView)
}

func openBrowser(urlStr string) error {
	if isWSL() {
		if _, err := exec.LookPath("wslview"); err == nil {
			return exec.Command("wslview", urlStr).Start()
		}
		if _, err := exec.LookPath("cmd.exe"); err == nil {
			return exec.Command("cmd.exe", "/c", "start", "", urlStr).Start()
		}
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			escapedUrl := strings.ReplaceAll(urlStr, "'", "`'")
			return exec.Command("powershell.exe", "-NoProfile", "-Command", fmt.Sprintf("Start-Process '%s'", escapedUrl)).Start()
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", urlStr)
	case "windows":
		cmd = exec.Command("cmd.exe", "/c", "start", "", urlStr)
	default:
		cmd = exec.Command("xdg-open", urlStr)
	}
	return cmd.Start()
}

func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	if err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "microsoft") || strings.Contains(content, "wsl") {
			return true
		}
	}
	data, err = os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "microsoft") || strings.Contains(content, "wsl") {
			return true
		}
	}
	return false
}

func sortFileEntries(entries []FileEntry, key string, order string) {
	if key == "none" || len(entries) <= 1 {
		return
	}

	// Maintain ".." at the top if present
	hasDotDot := entries[0].Name == ".."
	var startIdx int
	if hasDotDot {
		startIdx = 1
	} else {
		startIdx = 0
	}

	toSort := entries[startIdx:]

	// Separate directories and files so directories always stay grouped on top
	var dirs []FileEntry
	var files []FileEntry
	for _, entry := range toSort {
		if entry.IsDir {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	sortSlice := func(slice []FileEntry) {
		sort.SliceStable(slice, func(i, j int) bool {
			var valI, valJ interface{}
			switch key {
			case "name":
				valI = strings.ToLower(slice[i].Name)
				valJ = strings.ToLower(slice[j].Name)
			case "size":
				valI = slice[i].Size
				valJ = slice[j].Size
			case "date":
				valI = slice[i].ModTime
				valJ = slice[j].ModTime
			case "type":
				valI = strings.ToLower(slice[i].MimeType)
				valJ = strings.ToLower(slice[j].MimeType)
			default:
				valI = strings.ToLower(slice[i].Name)
				valJ = strings.ToLower(slice[j].Name)
			}

			if order == "asc" {
				switch key {
				case "size":
					return valI.(int64) < valJ.(int64)
				default:
					return valI.(string) < valJ.(string)
				}
			} else {
				switch key {
				case "size":
					return valI.(int64) > valJ.(int64)
				default:
					return valI.(string) > valJ.(string)
				}
			}
		})
	}

	sortSlice(dirs)
	sortSlice(files)

	idx := startIdx
	for _, d := range dirs {
		entries[idx] = d
		idx++
	}
	for _, f := range files {
		entries[idx] = f
		idx++
	}
}

func showSortChoice() {
	prevFocus := tuiApp.GetFocus()
	isLocalActive := false
	if localTable != nil && localTable.HasFocus() {
		isLocalActive = true
	}

	list := tview.NewList().
		AddItem("Sort by Name (Ascending)", "A-Z", '1', nil).
		AddItem("Sort by Name (Descending)", "Z-A", '2', nil).
		AddItem("Sort by Size (Ascending)", "Small to Large", '3', nil).
		AddItem("Sort by Size (Descending)", "Large to Small", '4', nil).
		AddItem("Sort by Date (Ascending)", "Oldest first", '5', nil).
		AddItem("Sort by Date (Descending)", "Newest first", '6', nil).
		AddItem("Sort by Type (Ascending)", "MimeType A-Z", '7', nil).
		AddItem("Sort by Type (Descending)", "MimeType Z-A", '8', nil)

	list.SetBorder(true).
		SetTitle(" Sort Options ").
		SetTitleColor(tcell.ColorYellow)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(tview.NewBox(), 0, 15, false)
	inner.AddItem(list, 0, 70, true)
	inner.AddItem(tview.NewBox(), 0, 15, false)

	flex.AddItem(inner, 14, 0, true)
	flex.AddItem(tview.NewBox(), 0, 1, false)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("sort_choice")
			tuiApp.SetFocus(prevFocus)
			return nil
		}
		return event
	})

	list.SetSelectedFunc(func(itemIndex int, mainText string, secondaryText string, shortcut rune) {
		pages.RemovePage("sort_choice")
		tuiApp.SetFocus(prevFocus)

		key := "name"
		order := "asc"

		switch itemIndex {
		case 0:
			key, order = "name", "asc"
		case 1:
			key, order = "name", "desc"
		case 2:
			key, order = "size", "asc"
		case 3:
			key, order = "size", "desc"
		case 4:
			key, order = "date", "asc"
		case 5:
			key, order = "date", "desc"
		case 6:
			key, order = "type", "asc"
		case 7:
			key, order = "type", "desc"
		}

		if isLocalActive {
			localSortKey = key
			localSortOrder = order
		} else {
			remoteSortKey = key
			remoteSortOrder = order
		}

		refreshPanels()
	})

	pages.AddPage("sort_choice", flex, true, true)
	tuiApp.SetFocus(list)
}

func onCreateDirOrFolder() {
	clearSearchModes()
	if lastActiveTable == localTable {
		promptTextInput(" Create New Directory ", "Dir Name: ", "", func(name string) {
			if name == "" {
				return
			}
			newPath := filepath.Join(currentLocalDir, name)
			runTask("Creating directory...", func() error {
				return os.MkdirAll(newPath, 0755)
			})
		})
	} else if lastActiveTable == remoteTable {
		if !isAuthorized {
			showError("Google Drive features are disabled. Please authenticate.")
			return
		}
		promptTextInput(" Create New Folder on Drive ", "Folder Name: ", "", func(name string) {
			if name == "" {
				return
			}
			runTask("Creating Google Drive folder...", func() error {
				_, err := tuiCreateDriveFolderFn(name, currentRemoteFolderID, authContainer)
				return err
			})
		})
	}
}

func onSearch() {
	if lastActiveTable == localTable {
		promptTextInput(" Search Local Directory (Recursive) ", "Search keyword: ", "", func(query string) {
			if query == "" {
				return
			}
			inSearchModeLocal = true
			runTask("Searching local files...", func() error {
				results, err := searchLocalRecursive(currentLocalDir, query)
				if err != nil {
					inSearchModeLocal = false
					return err
				}
				tuiApp.QueueUpdateDraw(func() {
					localFiles = results
					sortFileEntries(localFiles, localSortKey, localSortOrder)
					populateTable(localTable, localFiles, true)
					localTable.SetTitle(" Search Results for '" + query + "' under " + currentLocalDir + " (Press 'r' to return to normal view) ")
					localTable.SetTitleColor(tcell.ColorYellow)
					localTable.SetBorderColor(tcell.ColorYellow)
				})
				return nil
			})
		})
	} else if lastActiveTable == remoteTable {
		if !isAuthorized {
			showError("Google Drive features are disabled. Please authenticate.")
			return
		}
		promptTextInput(" Search Google Drive (All Drives) ", "Search keyword: ", "", func(query string) {
			if query == "" {
				return
			}
			inSearchModeRemote = true
			runTask("Searching Google Drive...", func() error {
				results, err := searchRemoteDriveAllFn(authContainer, mainCtx, query)
				if err != nil {
					inSearchModeRemote = false
					return err
				}
				tuiApp.QueueUpdateDraw(func() {
					remoteFiles = results
					sortFileEntries(remoteFiles, remoteSortKey, remoteSortOrder)
					populateTable(remoteTable, remoteFiles, false)
					remoteTable.SetTitle(" Search Results for '" + query + "' in Drive (Press 'r' to return to normal view) ")
					remoteTable.SetTitleColor(tcell.ColorYellow)
					remoteTable.SetBorderColor(tcell.ColorYellow)
				})
				return nil
			})
		})
	}
}

func searchLocalRecursive(baseDir, query string) ([]FileEntry, error) {
	var results []FileEntry
	queryLower := strings.ToLower(query)
	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == baseDir {
			return nil
		}
		name := d.Name()
		if strings.Contains(strings.ToLower(name), queryLower) {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			isDir := d.IsDir()
			mimeType := ""
			if isDir {
				mimeType = "directory"
			} else {
				mimeType = utl.ExtToMime(filepath.Ext(name))
			}
			modTime := info.ModTime().Format("2006-01-02 15:04:05")
			createdTime := getLocalCreatedTime(info).Format("2006-01-02 15:04:05")
			results = append(results, FileEntry{
				Name:        name,
				Path:        path,
				MimeType:    mimeType,
				ModTime:     modTime,
				CreatedTime: createdTime,
				Size:        info.Size(),
				IsDir:       isDir,
				Permissions: info.Mode().String(),
			})
		}
		return nil
	})
	return results, err
}

func searchRemoteDriveAll(auth *app.AuthContainer, c *cli.Context, query string) ([]FileEntry, error) {
	escapedQuery := strings.ReplaceAll(query, "'", "\\'")
	q := fmt.Sprintf("name contains '%s' and trashed = false", escapedQuery)
	fields := "files(id,name,mimeType,modifiedTime,createdTime,size,webViewLink,owners(displayName),shared,description),nextPageToken"

	var files []FileEntry
	pageToken := ""
	for {
		params := url.Values{}
		params.Set("q", q)
		params.Set("fields", fields)
		params.Set("pageSize", "1000")
		params.Set("supportsAllDrives", "true")
		params.Set("includeItemsFromAllDrives", "true")
		params.Set("corpora", "allDrives")
		if pageToken != "" {
			params.Set("pageToken", pageToken)
		}

		r := &utl.RequestParams{
			Method:      "GET",
			APIURL:      "https://www.googleapis.com/drive/v3/files?" + params.Encode(),
			Accesstoken: auth.GgsrunCfg.Accesstoken,
			Dtime:       30,
		}

		body, err := requestParamsFetchFn(r)
		if err != nil {
			return nil, err
		}

		var res struct {
			Files []struct {
				ID           string     `json:"id"`
				Name         string     `json:"name"`
				MimeType     string     `json:"mimeType"`
				ModifiedTime *time.Time `json:"modifiedTime"`
				CreatedTime  *time.Time `json:"createdTime"`
				Size         string     `json:"size"`
				WebViewLink  string     `json:"webViewLink"`
				Owners       []struct {
					DisplayName string `json:"displayName"`
				} `json:"owners"`
				Shared      bool   `json:"shared"`
				Description string `json:"description"`
			} `json:"files"`
			NextPageToken string `json:"nextPageToken"`
		}

		if err := json.Unmarshal(body, &res); err != nil {
			return nil, err
		}

		for _, f := range res.Files {
			isDir := f.MimeType == "application/vnd.google-apps.folder"
			var size int64
			if f.Size != "" {
				size, _ = strconv.ParseInt(f.Size, 10, 64)
			}
			modTime := ""
			if f.ModifiedTime != nil {
				modTime = f.ModifiedTime.In(time.Local).Format("2006-01-02 15:04:05")
			}
			createdTime := ""
			if f.CreatedTime != nil {
				createdTime = f.CreatedTime.In(time.Local).Format("2006-01-02 15:04:05")
			}
			perm := ""
			if len(f.Owners) > 0 {
				perm = "Owner: " + f.Owners[0].DisplayName
				if f.Shared {
					perm += " (Shared)"
				} else {
					perm += " (Private)"
				}
			}
			files = append(files, FileEntry{
				Name:        f.Name,
				Path:        f.ID,
				MimeType:    f.MimeType,
				ModTime:     modTime,
				CreatedTime: createdTime,
				Size:        size,
				IsDir:       isDir,
				WebViewLink: f.WebViewLink,
				Permissions: perm,
				Description: f.Description,
			})
		}

		pageToken = res.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return files, nil
}

func generateLocalTree(dirPath string, prefix string) ([]string, error) {
	var lines []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		lines = append(lines, prefix+connector+entry.Name())
		if entry.IsDir() {
			subLines, err := generateLocalTree(filepath.Join(dirPath, entry.Name()), nextPrefix)
			if err == nil {
				lines = append(lines, subLines...)
			}
		}
	}
	return lines, nil
}

func generateRemoteTree(auth *app.AuthContainer, c *cli.Context, folderID string, prefix string) ([]string, error) {
	var lines []string
	p := auth.DefDownloadContainerExported(c)
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	fields := "files(id,name,mimeType)"
	fl := p.GetListLoop(q, fields)

	entries := fl.Files
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		isFolder := entry.MimeType == "application/vnd.google-apps.folder"
		displayName := entry.Name
		if isFolder {
			displayName = "[DIR] " + displayName
		}
		lines = append(lines, prefix+connector+displayName)

		if isFolder {
			subLines, err := generateRemoteTreeFn(auth, c, entry.ID, nextPrefix)
			if err == nil {
				lines = append(lines, subLines...)
			}
		}
	}
	return lines, nil
}

func updateAndRenderProgress(msg string) {
	transferProgressMu.Lock()
	defer transferProgressMu.Unlock()

	if strings.HasPrefix(msg, "Progress:") {
		parts := strings.SplitN(msg, ":", 4)
		if len(parts) == 4 {
			filename := parts[1]
			transferred, _ := strconv.ParseInt(parts[2], 10, 64)
			total, _ := strconv.ParseInt(parts[3], 10, 64)
			jp := transferProgress[filename]
			jp.Transferred = transferred
			jp.Total = total
			if jp.Status == "" || jp.Status == "Pending" {
				if strings.Contains(transferHeading, "Uploading") {
					jp.Status = "Uploading"
				} else {
					jp.Status = "Downloading"
				}
			}
			transferProgress[filename] = jp
		}
	} else {
		// Milestones
		if strings.HasPrefix(msg, "Uploaded:") {
			fn := strings.TrimSpace(strings.TrimPrefix(msg, "Uploaded:"))
			jp := transferProgress[fn]
			jp.Status = "Completed"
			jp.Transferred = jp.Total
			transferProgress[fn] = jp
		} else if strings.HasPrefix(msg, "Downloaded:") {
			fn := strings.TrimSpace(strings.TrimPrefix(msg, "Downloaded:"))
			jp := transferProgress[fn]
			jp.Status = "Completed"
			jp.Transferred = jp.Total
			transferProgress[fn] = jp
		} else if strings.HasPrefix(msg, "Downloaded script project:") {
			fn := strings.TrimSpace(strings.TrimPrefix(msg, "Downloaded script project:"))
			jp := transferProgress[fn]
			jp.Status = "Completed"
			jp.Transferred = jp.Total
			transferProgress[fn] = jp
		} else if strings.Contains(msg, "failed") || strings.Contains(msg, "error") || strings.Contains(msg, "Failed") {
			for fn := range transferProgress {
				if strings.Contains(msg, fn) {
					jp := transferProgress[fn]
					jp.Status = "Failed"
					transferProgress[fn] = jp
				}
			}
		}
		transferLogLines = append(transferLogLines, msg)
	}

	if currentLoadingView != nil {
		currentLoadingView.Clear()
		fmt.Fprintln(currentLoadingView, transferHeading)
		fmt.Fprintln(currentLoadingView)

		var keys []string
		for k := range transferProgress {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, fn := range keys {
			jp := transferProgress[fn]
			pct := 0
			if jp.Status == "Completed" {
				pct = 100
			} else if jp.Total > 0 {
				pct = int(float64(jp.Transferred) * 100 / float64(jp.Total))
			}

			barLen := 20
			filledLen := pct * barLen / 100
			bar := ""
			for i := 0; i < barLen; i++ {
				if i < filledLen {
					bar += "="
				} else if i == filledLen && filledLen < barLen {
					bar += ">"
				} else {
					bar += " "
				}
			}

			statusColor := "yellow"
			if jp.Status == "Completed" {
				statusColor = "green"
			} else if jp.Status == "Failed" {
				statusColor = "red"
			}

			sizeInfo := ""
			if jp.Total > 0 {
				sizeInfo = fmt.Sprintf("%s / %s", formatSize(jp.Transferred, false, ""), formatSize(jp.Total, false, ""))
			} else {
				sizeInfo = formatSize(jp.Transferred, false, "")
			}

			fmt.Fprintf(currentLoadingView, "  %-25s : [%s]%3d%%[white] [[blue]%s[white]]  %s\n", fn, statusColor, pct, bar, sizeInfo)
		}

		fmt.Fprintln(currentLoadingView)
		fmt.Fprintln(currentLoadingView, "----------------------------------------------------------------------")
		for _, logLine := range transferLogLines {
			fmt.Fprintln(currentLoadingView, " * "+logLine)
		}
		currentLoadingView.ScrollToEnd()
	}
}

func clearSearchModes() {
	inSearchModeLocal = false
	inSearchModeRemote = false
}


