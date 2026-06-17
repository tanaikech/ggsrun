package tui

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ggsrun/internal/app"
	"ggsrun/internal/utl"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

// Helper widgets selectors
func findFileList(p tview.Primitive) *tview.List {
	if list, ok := p.(*tview.List); ok {
		return list
	}
	if flex, ok := p.(*tview.Flex); ok {
		for i := 0; i < flex.GetItemCount(); i++ {
			item := flex.GetItem(i)
			if res := findFileList(item); res != nil {
				return res
			}
		}
	}
	return nil
}

func findInputField(p tview.Primitive) *tview.InputField {
	if input, ok := p.(*tview.InputField); ok {
		return input
	}
	if flex, ok := p.(*tview.Flex); ok {
		for i := 0; i < flex.GetItemCount(); i++ {
			item := flex.GetItem(i)
			if res := findInputField(item); res != nil {
				return res
			}
		}
	}
	return nil
}

func findTextView(p tview.Primitive) *tview.TextView {
	if tv, ok := p.(*tview.TextView); ok {
		return tv
	}
	if flex, ok := p.(*tview.Flex); ok {
		for i := 0; i < flex.GetItemCount(); i++ {
			item := flex.GetItem(i)
			if res := findTextView(item); res != nil {
				return res
			}
		}
	}
	return nil
}

func findModal(p tview.Primitive) *tview.Modal {
	if m, ok := p.(*tview.Modal); ok {
		return m
	}
	if flex, ok := p.(*tview.Flex); ok {
		for i := 0; i < flex.GetItemCount(); i++ {
			item := flex.GetItem(i)
			if res := findModal(item); res != nil {
				return res
			}
		}
	}
	return nil
}

func setupTestTUI(t *testing.T, localFilesData []FileEntry, remoteFilesData []FileEntry) (tcell.SimulationScreen, func()) {
	// Setup mocks
	listLocalFilesFn = func(dir string) ([]FileEntry, error) {
		return localFilesData, nil
	}
	listRemoteFilesFn = func(auth *app.AuthContainer, c *cli.Context, folderID string) ([]FileEntry, error) {
		return remoteFilesData, nil
	}
	osReadFileFn = func(name string) ([]byte, error) {
		return []byte("function main() {\n  Logger.log('Hello');\n}"), nil
	}
	getBoundScriptExportedFn = func(auth *app.AuthContainer, c *cli.Context, scriptID string) *utl.ProjectForAppsScriptApi {
		return &utl.ProjectForAppsScriptApi{
			ScriptId: scriptID,
			Files: []utl.FilesForAppsScriptApi{
				{Name: "Code", Type: "SERVER_JS", Source: "function main() {\n  Logger.log('Hello');\n}"},
			},
		}
	}
	getAuthContainerFn = func(c *cli.Context) *app.AuthContainer {
		if envCfg := os.Getenv("GGSRUN_CFG_PATH"); envCfg != "" {
			if _, err := os.Stat(filepath.Join(envCfg, "ggsrun.cfg")); err == nil {
				return app.GetAuthenticatedAuthContainer(c)
			}
		}
		return &app.AuthContainer{
			InitVal: &app.InitVal{},
			ResMsg:  &app.ResMsg{},
			GgsrunCfg: &app.GgsrunCfg{
				Accesstoken: "mock_access_token",
			},
			Param:  &app.Param{},
			Cs:     &app.Cs{},
			Atoken: &app.Atoken{},
			ChkAt:  &app.ChkAt{},
		}
	}

	simScreen := tcell.NewSimulationScreen("UTF-8")
	err := simScreen.Init()
	if err != nil {
		t.Fatalf("Failed to initialize simulation screen: %v", err)
	}
	testScreen = simScreen

	appObj := cli.NewApp()
	appObj.Version = "5.3.1"
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	cliCtx := cli.NewContext(appObj, set, nil)

	errChan := make(chan error, 1)
	go func() {
		errChan <- RunTUI(cliCtx)
	}()

	// Wait for tuiApp and widgets to be created
	for i := 0; i < 50; i++ {
		if tuiApp != nil && localTable != nil && remoteTable != nil && pages != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if tuiApp == nil || localTable == nil || remoteTable == nil || pages == nil {
		t.Fatal("TUI elements were not fully initialized")
	}

	// Wait for TUI to start event loop and process first draw
	initChan := make(chan struct{})
	tuiApp.QueueUpdate(func() {
		close(initChan)
	})
	select {
	case <-initChan:
	case <-time.After(2 * time.Second):
		t.Fatalf("TUI did not initialize in time")
	}
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		tuiApp.QueueUpdate(func() {
			tuiApp.Stop()
		})
		select {
		case <-errChan:
		case <-time.After(1 * time.Second):
		}
		simScreen.Fini()
		tuiApp = nil
		localTable = nil
		remoteTable = nil
		pages = nil
		
		// Reset mocks
		listLocalFilesFn = listLocalFiles
		listRemoteFilesFn = listRemoteFiles
		getAuthContainerFn = app.GetAuthenticatedAuthContainer
		tuiUploadFn = app.TuiUpload
		tuiDownloadFn = app.TuiDownload
		tuiExecuteGasFn = app.TuiExecuteGas
		tuiUpdateDriveMetadataFn = app.TuiUpdateDriveMetadata
		tuiCopyDriveFileFn = app.TuiCopyDriveFile
		tuiRunExe1Fn = app.TuiRunExe1
		tuiRunExe2Fn = app.TuiRunExe2
		tuiRunWebAppsFn = app.TuiRunWebApps
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
	}

	return simScreen, cleanup
}

func TestTUI_FilerVersion(t *testing.T) {
	appObj := cli.NewApp()
	appObj.Version = "5.3.1"
	if appObj.Version != "5.3.1" {
		t.Errorf("Expected version 5.3.1, got %s", appObj.Version)
	}
}

func TestTUI_NavigationAndHelp(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock/parent", IsDir: true},
		{Name: "file.txt", Path: "/mock/file.txt", IsDir: false, MimeType: "text/plain"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "remote_file.pdf", Path: "id_remote_file", IsDir: false, MimeType: "application/pdf"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Initial focus should be on local table
	if !localTable.HasFocus() {
		t.Error("Expected local table to have focus initially")
	}

	// Switch focus using SetFocus
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
	})
	time.Sleep(50 * time.Millisecond)
	if !remoteTable.HasFocus() {
		t.Error("Expected remote table to have focus after Tab focus switch")
	}

	// Switch back focus
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
	})
	time.Sleep(50 * time.Millisecond)
	if !localTable.HasFocus() {
		t.Error("Expected local table to have focus after switching back")
	}

	// Show help guide by executing input capture directly
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
		}
	})
	time.Sleep(50 * time.Millisecond)
	
	var hasHelp bool
	tuiApp.QueueUpdate(func() {
		hasHelp = pages.HasPage("help")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasHelp {
		t.Error("Expected help page to be visible after pressing 'h'")
	}

	// Close help guide by finding its TextView and simulating Esc
	var helpTextView *tview.TextView
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		helpTextView = findTextView(p)
	})
	time.Sleep(50 * time.Millisecond)

	if helpTextView != nil {
		tuiApp.QueueUpdate(func() {
			handler := helpTextView.GetInputCapture()
			if handler != nil {
				handler(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
			}
		})
	}
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		hasHelp = pages.HasPage("help")
	})
	time.Sleep(50 * time.Millisecond)
	if hasHelp {
		t.Error("Expected help page to be closed after Esc")
	}
}

func TestTUI_LocalPreview(t *testing.T) {
	// Create a temporary text file
	tmpFile, err := os.CreateTemp("", "test_preview_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Hello, Antigravity CLI Filer!"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	mockLocal := []FileEntry{
		{Name: "..", Path: filepath.Dir(tmpFile.Name()), IsDir: true},
		{Name: filepath.Base(tmpFile.Name()), Path: tmpFile.Name(), IsDir: false, MimeType: "text/plain"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Select the temp file (row 2 in table, index 1 in localFiles)
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press Enter to open preview
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasPreview bool
	tuiApp.QueueUpdate(func() {
		hasPreview = pages.HasPage("preview")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasPreview {
		t.Error("Expected preview page to be visible after Enter on local text file")
	}

	// Close preview
	var previewTextView *tview.TextView
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		previewTextView = findTextView(p)
	})
	time.Sleep(50 * time.Millisecond)

	if previewTextView != nil {
		tuiApp.QueueUpdate(func() {
			handler := previewTextView.GetInputCapture()
			if handler != nil {
				handler(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
			}
		})
	}
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		hasPreview = pages.HasPage("preview")
	})
	time.Sleep(50 * time.Millisecond)
	if hasPreview {
		t.Error("Expected preview page to be closed after Esc")
	}
}

func TestTUI_FileDetails(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "local.txt", Path: "/mock/local.txt", IsDir: false, MimeType: "text/plain", ModTime: "2026-06-16 12:00:00", Size: 100},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "remote.txt", Path: "id_remote_txt", IsDir: false, MimeType: "text/plain", ModTime: "2026-06-16 13:00:00", Size: 200},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// 1. Local details
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press 'i'
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasDetails bool
	tuiApp.QueueUpdate(func() {
		hasDetails = pages.HasPage("details")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasDetails {
		t.Error("Expected details dialog to be shown for local file")
	}

	// Close details
	var detailsTV *tview.TextView
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		detailsTV = findTextView(p)
	})
	time.Sleep(50 * time.Millisecond)

	if detailsTV != nil {
		tuiApp.QueueUpdate(func() {
			detailsTV.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		})
	}
	time.Sleep(50 * time.Millisecond)

	// 2. Remote details
	// Switch to remote
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Setup mock FetchAPI to return mock Drive metadata
	var fetchCalled bool
	requestParamsFetchFn = func(r *utl.RequestParams) ([]byte, error) {
		fetchCalled = true
		meta := struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			MimeType     string `json:"mimeType"`
			Size         string `json:"size"`
			ModifiedTime string `json:"modifiedTime"`
			Description  string `json:"description"`
			Owners       []struct {
				DisplayName  string `json:"displayName"`
				EmailAddress string `json:"emailAddress"`
			} `json:"owners"`
		}{
			ID:           "id_remote_txt",
			Name:         "remote.txt",
			MimeType:     "text/plain",
			Size:         "200",
			ModifiedTime: "2026-06-16T13:00:00Z",
			Description:  "Filer Details Test Description",
		}
		return json.Marshal(meta)
	}

	// Press 'i'
	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
		}
	})
	time.Sleep(150 * time.Millisecond) // Give runTask goroutine time to complete

	if !fetchCalled {
		t.Error("Expected requestParamsFetchFn to be called for remote details")
	}

	tuiApp.QueueUpdate(func() {
		hasDetails = pages.HasPage("details")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasDetails {
		t.Error("Expected details dialog to be shown for remote file")
	}

	// Close details
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		detailsTV = findTextView(p)
	})
	time.Sleep(50 * time.Millisecond)

	if detailsTV != nil {
		tuiApp.QueueUpdate(func() {
			detailsTV.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		})
	}
	time.Sleep(50 * time.Millisecond)
}

func TestTUI_SpaceMultiSelect(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "file1.txt", Path: "/mock/file1.txt", IsDir: false, MimeType: "text/plain"},
		{Name: "file2.txt", Path: "/mock/file2.txt", IsDir: false, MimeType: "text/plain"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Select file1.txt (row 2)
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press Space
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
		}
	})
	time.Sleep(50 * time.Millisecond)

	if !selectedLocalPaths["/mock/file1.txt"] {
		t.Error("Expected file1.txt to be selected in selectedLocalPaths")
	}

	// Select file2.txt (row 3)
	tuiApp.QueueUpdateDraw(func() {
		localTable.Select(3, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press Space
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
		}
	})
	time.Sleep(50 * time.Millisecond)

	if !selectedLocalPaths["/mock/file2.txt"] {
		t.Error("Expected file2.txt to be selected in selectedLocalPaths")
	}

	// Check table cell rendering of selection
	var cellText string
	tuiApp.QueueUpdate(func() {
		cellText = localTable.GetCell(2, 0).Text
	})
	time.Sleep(50 * time.Millisecond)
	if !strings.HasPrefix(cellText, "* ") {
		t.Errorf("Expected cell to render with * , got %q", cellText)
	}

	// Toggle file1.txt off
	tuiApp.QueueUpdateDraw(func() {
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
		}
	})
	time.Sleep(50 * time.Millisecond)

	if selectedLocalPaths["/mock/file1.txt"] {
		t.Error("Expected file1.txt to be deselected")
	}
}

func TestTUI_DomainOperations(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "file.txt", Path: "/mock/file.txt", IsDir: false, MimeType: "text/plain", ModTime: "2026-06-16 12:00:00"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "remote.txt", Path: "id_remote_txt", IsDir: false, MimeType: "text/plain", ModTime: "2026-06-16 13:00:00"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// 1. Rename
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasPrompt bool
	tuiApp.QueueUpdate(func() {
		hasPrompt = pages.HasPage("text_prompt")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasPrompt {
		t.Error("Expected rename prompt dialog to be visible")
	}

	// Close prompt
	var promptInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		promptInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)

	if promptInput != nil {
		tuiApp.QueueUpdate(func() {
			promptInput.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(p tview.Primitive) {})
		})
	}
	time.Sleep(50 * time.Millisecond)

	// 2. Change Timestamp
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		hasPrompt = pages.HasPage("text_prompt")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasPrompt {
		t.Error("Expected timestamp change prompt dialog to be visible")
	}

	// Close prompt
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		promptInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)

	if promptInput != nil {
		tuiApp.QueueUpdate(func() {
			promptInput.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(p tview.Primitive) {})
		})
	}
	time.Sleep(50 * time.Millisecond)

	// 3. Edit Description (Drive only)
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Mock fetch description response
	requestParamsFetchFn = func(r *utl.RequestParams) ([]byte, error) {
		return []byte(`{"description": "Old Filer Description"}`), nil
	}

	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		}
	})
	time.Sleep(150 * time.Millisecond) // Goroutine loading description

	tuiApp.QueueUpdate(func() {
		hasPrompt = pages.HasPage("text_prompt")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasPrompt {
		t.Error("Expected description edit prompt dialog to be visible")
	}

	// Close prompt
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		promptInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)

	if promptInput != nil {
		tuiApp.QueueUpdate(func() {
			promptInput.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(p tview.Primitive) {})
		})
	}
	time.Sleep(50 * time.Millisecond)
}

func TestTUI_GASProjectExplorer(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "MyScript", Path: "id_myscript", IsDir: false, MimeType: "application/vnd.google-apps.script"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Switch focus to remote panel
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Mock project load
	getBoundScriptExportedFn = func(auth *app.AuthContainer, c *cli.Context, scriptID string) *utl.ProjectForAppsScriptApi {
		return &utl.ProjectForAppsScriptApi{
			ScriptId: "id_myscript",
			Files: []utl.FilesForAppsScriptApi{
				{Name: "Code", Type: "SERVER_JS", Source: "function main() {\n  Logger.log('Hello');\n}"},
			},
		}
	}

	// Press Enter to open GAS project explorer
	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
		}
	})
	time.Sleep(200 * time.Millisecond) // Goroutine loading

	var hasGasExplorer bool
	tuiApp.QueueUpdate(func() {
		hasGasExplorer = pages.HasPage("gas_project")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasGasExplorer {
		t.Fatal("Expected GAS project page to be visible")
	}

	// Get fileList in gas_project page
	var gasFileList *tview.List
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		gasFileList = findFileList(p)
	})
	time.Sleep(50 * time.Millisecond)

	if gasFileList == nil {
		t.Fatal("Expected to find fileList in gas_project page")
	}

	// Press 'e' on fileList to show execute GAS prompt
	tuiApp.QueueUpdate(func() {
		handler := gasFileList.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasExecPrompt bool
	tuiApp.QueueUpdate(func() {
		hasExecPrompt = pages.HasPage("exec_mode_choice")
	})
	if !hasExecPrompt {
		t.Fatal("Expected exec_mode_choice page to be visible")
	}

	// Get list in exec_mode_choice page
	var modeList *tview.List
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		modeList = findFileList(p)
	})
	time.Sleep(50 * time.Millisecond)

	if modeList == nil {
		t.Fatal("Expected to find list in exec_mode_choice page")
	}

	// Select first option (exe1)
	tuiApp.QueueUpdate(func() {
		modeList.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 1: Script ID text_prompt
	var scriptIDInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		scriptIDInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if scriptIDInput == nil {
		t.Fatal("Expected to find scriptID input field")
	}
	tuiApp.QueueUpdate(func() {
		scriptIDInput.SetText("id_myscript")
		scriptIDInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 2: Function Name text_prompt
	var funcNameInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		funcNameInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if funcNameInput == nil {
		t.Fatal("Expected to find funcName input field")
	}
	tuiApp.QueueUpdate(func() {
		funcNameInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 3: Argument Value text_prompt
	var argValInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		argValInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if argValInput == nil {
		t.Fatal("Expected to find argVal input field")
	}

	// Mock TuiRunExe1 call
	var execGasCalled bool
	tuiRunExe1Fn = func(ctx *cli.Context, a *app.AuthContainer) (string, error) {
		execGasCalled = true
		return `{"result": "Success!"}`, nil
	}

	// Confirm execution by executing input field DoneFunc directly
	tuiApp.QueueUpdate(func() {
		argValInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(200 * time.Millisecond) // Goroutine execution

	if !execGasCalled {
		t.Error("Expected tuiRunExe1Fn to be called")
	}

	var hasResult bool
	tuiApp.QueueUpdate(func() {
		hasResult = pages.HasPage("execution_result")
	})
	time.Sleep(50 * time.Millisecond)
	if !hasResult {
		t.Error("Expected execution result dialog to be visible")
	}

	// Close result
	var resultTextView *tview.TextView
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		resultTextView = findTextView(p)
	})
	time.Sleep(50 * time.Millisecond)

	if resultTextView != nil {
		tuiApp.QueueUpdate(func() {
			handler := resultTextView.GetInputCapture()
			if handler != nil {
				handler(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
			}
		})
	}
	time.Sleep(50 * time.Millisecond)

	// Close GAS project explorer
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		gasFileList = findFileList(p)
	})
	time.Sleep(50 * time.Millisecond)

	if gasFileList != nil {
		tuiApp.QueueUpdate(func() {
			handler := gasFileList.GetInputCapture()
			if handler != nil {
				handler(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
			}
		})
	}
	time.Sleep(50 * time.Millisecond)
}

func TestTUI_MimeTypeConversion(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "MyDoc", Path: "id_mydoc", IsDir: false, MimeType: "application/vnd.google-apps.document"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Switch to remote
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press 'x' to trigger MIME conversion
	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasPrompt bool
	var hasError bool
	var localFocus bool
	var remoteFocus bool
	var selectedIndex int
	var selectedName string
	var optsLen int
	var focusedType string

	tuiApp.QueueUpdate(func() {
		hasPrompt = pages.HasPage("convert_choice")
		hasError = pages.HasPage("error")
		localFocus = localTable.HasFocus()
		remoteFocus = remoteTable.HasFocus()
		selectedIndex = getSelectedRowIndex(remoteTable)
		if selectedIndex >= 0 && selectedIndex < len(remoteFiles) {
			selectedName = remoteFiles[selectedIndex].Name
			optsLen = len(getConvertOptions(remoteFiles[selectedIndex].MimeType))
		}
		if tuiApp.GetFocus() != nil {
			focusedType = fmt.Sprintf("%T", tuiApp.GetFocus())
		} else {
			focusedType = "nil"
		}
	})
	time.Sleep(50 * time.Millisecond)

	t.Logf("TUI State: hasPrompt=%t, hasError=%t, localFocus=%t, remoteFocus=%t, selectedIndex=%d, selectedName=%s, optsLen=%d, focusedType=%s",
		hasPrompt, hasError, localFocus, remoteFocus, selectedIndex, selectedName, optsLen, focusedType)

	if !hasPrompt {
		t.Error("Expected convert prompt to be shown on remote MimeType conversion 'x'")
	}

	// Close prompt by removing conversion page and switching focus
	tuiApp.QueueUpdate(func() {
		pages.RemovePage("convert_choice")
		tuiApp.SetFocus(remoteTable)
	})
	time.Sleep(50 * time.Millisecond)
}

func TestTUI_TransferOperations(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "data.bin", Path: "/mock/data.bin", IsDir: false, MimeType: "application/octet-stream"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "image.png", Path: "id_image_png", IsDir: false, MimeType: "image/png"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Test upload execution
	var uploadCalled bool
	tuiUploadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		uploadCalled = true
		return nil, nil
	}

	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press F5 (copy local to remote)
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone))
		}
	})
	time.Sleep(200 * time.Millisecond) // wait for goroutine

	if !uploadCalled {
		t.Error("Expected tuiUploadFn to be called on F5 upload for bin file")
	}

	// Test download execution
	var downloadCalled bool
	tuiDownloadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		downloadCalled = true
		return nil, nil
	}

	// Focus remote panel
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press F5 (copy remote to local)
	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone))
		}
	})
	time.Sleep(200 * time.Millisecond) // wait for goroutine

	if !downloadCalled {
		t.Error("Expected tuiDownloadFn to be called on F5 download for png file")
	}
}

func TestTUI_LocalScriptExecution(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "local_script.gs", Path: "/mock/local_script.gs", IsDir: false, MimeType: "text/javascript"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Focus local table and select the script file
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press 'e' on local script file
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Verify exec_mode_choice is visible
	var hasChoicePage bool
	tuiApp.QueueUpdate(func() {
		hasChoicePage = pages.HasPage("exec_mode_choice")
	})
	if !hasChoicePage {
		t.Fatal("Expected exec_mode_choice page to be visible for local script")
	}

	// Find the list on exec_mode_choice page
	var modeList *tview.List
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		modeList = findFileList(p)
	})
	time.Sleep(50 * time.Millisecond)
	if modeList == nil {
		t.Fatal("Expected to find list in exec_mode_choice page")
	}

	// Select second option (exe2)
	tuiApp.QueueUpdate(func() {
		modeList.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), func(p tview.Primitive) {})
		modeList.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 1: Script ID text_prompt
	var scriptIDInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		scriptIDInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if scriptIDInput == nil {
		t.Fatal("Expected to find scriptID input field")
	}
	tuiApp.QueueUpdate(func() {
		scriptIDInput.SetText("local_test_script_id")
		scriptIDInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 2: Argument Value text_prompt (Function Name is skipped for exe2)
	var argValInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		argValInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if argValInput == nil {
		t.Fatal("Expected to find argVal input field")
	}

	// Mock TuiRunExe2 call
	var execGasCalled bool
	tuiRunExe2Fn = func(ctx *cli.Context, a *app.AuthContainer) (string, error) {
		execGasCalled = true
		return `{"result": "Success Exe2!"}`, nil
	}

	// Confirm execution
	tuiApp.QueueUpdate(func() {
		argValInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(200 * time.Millisecond) // Goroutine execution

	if !execGasCalled {
		t.Error("Expected tuiRunExe2Fn to be called")
	}
}

func TestTUI_RemoteDirectExecution(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "StandaloneScript", Path: "id_standalone", IsDir: false, MimeType: "application/vnd.google-apps.script"},
		{Name: "NonScriptDoc", Path: "id_doc", IsDir: false, MimeType: "application/vnd.google-apps.document"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// 1. Try to execute a non-standalone document
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(3, 0) // Select NonScriptDoc
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasErrorPage bool
	tuiApp.QueueUpdate(func() {
		hasErrorPage = pages.HasPage("error")
	})
	if !hasErrorPage {
		t.Error("Expected error modal when attempting to execute non-standalone Google Apps Script")
	} else {
		// Close error modal
		tuiApp.QueueUpdate(func() {
			pages.RemovePage("error")
			tuiApp.SetFocus(remoteTable)
		})
		time.Sleep(50 * time.Millisecond)
	}

	// 2. Execute a standalone Google Apps Script
	tuiApp.QueueUpdateDraw(func() {
		remoteTable.Select(2, 0) // Select StandaloneScript
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	var hasChoicePage bool
	tuiApp.QueueUpdate(func() {
		hasChoicePage = pages.HasPage("exec_mode_choice")
	})
	if !hasChoicePage {
		t.Fatal("Expected exec_mode_choice page to be visible for standalone script")
	}

	// Get list
	var modeList *tview.List
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		modeList = findFileList(p)
	})
	time.Sleep(50 * time.Millisecond)

	if modeList == nil {
		name, _ := pages.GetFrontPage()
		t.Fatalf("modeList is nil! Front page is: %s", name)
	}

	// Select third option (webapps)
	tuiApp.QueueUpdate(func() {
		modeList.InputHandler()(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone), func(p tview.Primitive) {})
		modeList.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 1: Web Apps URL text_prompt
	var urlInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		urlInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if urlInput == nil {
		t.Fatal("Expected to find Web Apps URL input field")
	}
	tuiApp.QueueUpdate(func() {
		urlInput.SetText("https://script.google.com/macros/s/12345/exec")
		urlInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 2: Password text_prompt
	var pwdInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		pwdInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if pwdInput == nil {
		t.Fatal("Expected to find Password input field")
	}
	tuiApp.QueueUpdate(func() {
		pwdInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Step 3: Argument Value text_prompt
	var argValInput *tview.InputField
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		argValInput = findInputField(p)
	})
	time.Sleep(50 * time.Millisecond)
	if argValInput == nil {
		t.Fatal("Expected to find Argument Value input field")
	}

	// Mock TuiRunWebApps call
	var execGasCalled bool
	tuiRunWebAppsFn = func(ctx *cli.Context, a *app.AuthContainer) (string, error) {
		execGasCalled = true
		return `{"result": "Success WebApps!"}`, nil
	}

	// Confirm execution
	tuiApp.QueueUpdate(func() {
		argValInput.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(200 * time.Millisecond) // Goroutine execution

	if !execGasCalled {
		t.Error("Expected tuiRunWebAppsFn to be called")
	}
}

func TestTUI_TaskCancellation(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "slow_file.txt", Path: "id_slow", IsDir: false, MimeType: "text/plain"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Mock requestParamsFetchFn to block
	blockChan := make(chan struct{})
	requestParamsFetchFn = func(r *utl.RequestParams) ([]byte, error) {
		<-blockChan
		return []byte(`{}`), nil
	}

	// Focus remote and select slow_file.txt
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press 'i' to trigger details (which runs via runTask)
	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Verify loading page is visible
	var hasLoading bool
	tuiApp.QueueUpdate(func() {
		hasLoading = pages.HasPage("loading")
	})
	if !hasLoading {
		t.Fatal("Expected loading page to be visible")
	}

	// Press Esc to cancel the task
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		// Send Esc to the loading flex page
		p.InputHandler()(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), func(p tview.Primitive) {})
	})
	time.Sleep(100 * time.Millisecond)

	// Verify loading page is gone and error modal is shown
	var hasLoadingAfter bool
	var hasErrorPage bool
	tuiApp.QueueUpdate(func() {
		hasLoadingAfter = pages.HasPage("loading")
		hasErrorPage = pages.HasPage("error")
	})
	if hasLoadingAfter {
		t.Error("Expected loading page to be removed after Esc")
	}
	if !hasErrorPage {
		t.Error("Expected error modal to be visible after cancellation")
	}

	// Close error page
	tuiApp.QueueUpdate(func() {
		pages.RemovePage("error")
		tuiApp.SetFocus(remoteTable)
	})
	time.Sleep(50 * time.Millisecond)

	// Unblock the requestParamsFetchFn goroutine
	close(blockChan)
	time.Sleep(100 * time.Millisecond)
}

func TestTUI_ConvertPromptCancellation(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "text_file.txt", Path: "/mock/text_file.txt", IsDir: false, MimeType: "text/plain"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Focus local and select text_file.txt
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	// Press F5 (copy)
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Verify convert_prompt is visible
	var hasConvertPrompt bool
	tuiApp.QueueUpdate(func() {
		hasConvertPrompt = pages.HasPage("convert_prompt")
	})
	if !hasConvertPrompt {
		t.Fatal("Expected convert_prompt to be visible")
	}

	// Press Esc to cancel the conversion choice
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Verify convert_prompt is gone and focus is back to localTable
	var hasConvertPromptAfter bool
	var localHasFocus bool
	tuiApp.QueueUpdate(func() {
		hasConvertPromptAfter = pages.HasPage("convert_prompt")
		localHasFocus = localTable.HasFocus()
	})
	if hasConvertPromptAfter {
		t.Error("Expected convert_prompt to be removed after Esc")
	}
	if !localHasFocus {
		t.Error("Expected focus to be returned to localTable")
	}
}

func TestTUI_MdFileCopy(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "markdown_file.md", Path: "/mock/markdown_file.md", IsDir: false, MimeType: "text/markdown"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Focus local and select markdown_file.md
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	var uploadCalled bool
	var capturedFlags map[string]string

	tuiUploadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		uploadCalled = true
		capturedFlags = make(map[string]string)
		capturedFlags["noconvert"] = fmt.Sprintf("%t", c.Bool("noconvert"))
		capturedFlags["convertto"] = c.String("convertto")
		return nil, nil
	}

	// Press F5 (copy)
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Select option 0 "Upload as is"
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(200 * time.Millisecond) // wait for goroutine

	if !uploadCalled {
		t.Fatal("Expected tuiUploadFn to be called for Case A")
	}
	if capturedFlags["noconvert"] != "true" {
		t.Errorf("Expected noconvert to be true, got %q", capturedFlags["noconvert"])
	}
	if capturedFlags["convertto"] != "" {
		t.Errorf("Expected convertto to be empty, got %q", capturedFlags["convertto"])
	}

	// --- Case B: Convert to Google Docs ---
	uploadCalled = false
	capturedFlags = nil

	// Press F5 (copy)
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Select option 1 "Google Docs"
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.SetCurrentItem(1)
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(200 * time.Millisecond) // wait for goroutine

	if !uploadCalled {
		t.Fatal("Expected tuiUploadFn to be called for Case B")
	}
	if capturedFlags["noconvert"] == "true" {
		t.Error("Expected noconvert to be false")
	}
	if capturedFlags["convertto"] != "application/vnd.google-apps.document" {
		t.Errorf("Expected convertto to be application/vnd.google-apps.document, got %q", capturedFlags["convertto"])
	}
}

func TestTUI_AtomicMoveOperations(t *testing.T) {
	// Create two real files in a temp dir
	tempDir := t.TempDir()
	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")
	_ = os.WriteFile(file1Path, []byte("content1"), 0644)
	_ = os.WriteFile(file2Path, []byte("content2"), 0644)

	mockLocal := []FileEntry{
		{Name: "..", Path: tempDir, IsDir: true},
		{Name: "file1.txt", Path: file1Path, IsDir: false, MimeType: "text/plain"},
		{Name: "file2.txt", Path: file2Path, IsDir: false, MimeType: "text/plain"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Select both files using Space key
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(localTable)
		localTable.Select(2, 0) // Select file1.txt
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)) // toggle file1
		}
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdateDraw(func() {
		localTable.Select(3, 0) // Select file2.txt
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)) // toggle file2
		}
	})
	time.Sleep(50 * time.Millisecond)

	// --- Case A: One file fails, so no files should be deleted ---
	tuiUploadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		filename := c.String("filename")
		if strings.Contains(filename, "file2.txt") {
			return nil, fmt.Errorf("mock upload error for file2")
		}
		return nil, nil
	}

	// Press F6 (Move)
	tuiApp.QueueUpdate(func() {
		handler := localTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF6, 0, tcell.ModNone))
		}
	})
	time.Sleep(100 * time.Millisecond)

	// Since we upload text files, a convert_prompt is shown. Select option 0 "Upload as is"
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(200 * time.Millisecond) // Wait for first file's convert prompt

	// Second file's convert prompt is shown (due to batch collectNext). Select option 0
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(300 * time.Millisecond) // wait for batch execution to fail

	// Verify error occurred and BOTH files still exist on disk!
	if _, err := os.Stat(file1Path); os.IsNotExist(err) {
		t.Error("Expected file1.txt to still exist on disk because move failed")
	}
	if _, err := os.Stat(file2Path); os.IsNotExist(err) {
		t.Error("Expected file2.txt to still exist on disk because move failed")
	}

	// Close error modal if any
	tuiApp.QueueUpdate(func() {
		pages.RemovePage("error")
		tuiApp.SetFocus(localTable)
	})
	time.Sleep(50 * time.Millisecond)

	// Re-select files for Case B (they got cleared after the runTask completion)
	tuiApp.QueueUpdateDraw(func() {
		localTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)
	tuiApp.QueueUpdate(func() {
		localTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	})
	time.Sleep(50 * time.Millisecond)
	tuiApp.QueueUpdateDraw(func() {
		localTable.Select(3, 0)
	})
	time.Sleep(50 * time.Millisecond)
	tuiApp.QueueUpdate(func() {
		localTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	})
	time.Sleep(50 * time.Millisecond)

	// --- Case B: Both files succeed, so both files must be deleted ---
	tuiUploadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		return nil, nil
	}

	// Press F6 (Move)
	tuiApp.QueueUpdate(func() {
		localTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyF6, 0, tcell.ModNone))
	})
	time.Sleep(100 * time.Millisecond)

	// First file's convert prompt: select option 0
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(200 * time.Millisecond)

	// Second file's convert prompt: select option 0
	tuiApp.QueueUpdate(func() {
		_, p := pages.GetFrontPage()
		list := findFileList(p)
		if list != nil {
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	})
	time.Sleep(300 * time.Millisecond) // Wait for successful execution

	// Verify BOTH files are now deleted from disk!
	if _, err := os.Stat(file1Path); !os.IsNotExist(err) {
		t.Error("Expected file1.txt to be deleted from disk on move success")
	}
	if _, err := os.Stat(file2Path); !os.IsNotExist(err) {
		t.Error("Expected file2.txt to be deleted from disk on move success")
	}
}

func TestTUI_AtomicMoveRemoteToLocal(t *testing.T) {
	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "file1.txt", Path: "id1", IsDir: false, MimeType: "text/plain"},
		{Name: "file2.txt", Path: "id2", IsDir: false, MimeType: "text/plain"},
	}

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	// Select both files using Space key on remoteTable
	tuiApp.QueueUpdateDraw(func() {
		tuiApp.SetFocus(remoteTable)
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
		}
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdateDraw(func() {
		remoteTable.Select(3, 0)
	})
	time.Sleep(50 * time.Millisecond)

	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
		}
	})
	time.Sleep(50 * time.Millisecond)

	// --- Case A: One file fails, so no remote files should be deleted ---
	tuiDownloadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		fileid := c.String("fileid")
		if fileid == "id2" {
			return nil, fmt.Errorf("mock download error for id2")
		}
		return nil, nil
	}

	var deletedRemoteIDs []string
	deleteRemoteFileFn = func(auth *app.AuthContainer, c *cli.Context, fileID string) error {
		deletedRemoteIDs = append(deletedRemoteIDs, fileID)
		return nil
	}

	// Press F6 (Move)
	tuiApp.QueueUpdate(func() {
		handler := remoteTable.GetInputCapture()
		if handler != nil {
			handler(tcell.NewEventKey(tcell.KeyF6, 0, tcell.ModNone))
		}
	})
	time.Sleep(300 * time.Millisecond)

	if len(deletedRemoteIDs) > 0 {
		t.Errorf("Expected no remote files to be deleted on failure, but deleted: %v", deletedRemoteIDs)
	}

	// Close error modal if any
	tuiApp.QueueUpdate(func() {
		pages.RemovePage("error")
		tuiApp.SetFocus(remoteTable)
	})
	time.Sleep(50 * time.Millisecond)

	// Re-select files for Case B
	tuiApp.QueueUpdateDraw(func() {
		remoteTable.Select(2, 0)
	})
	time.Sleep(50 * time.Millisecond)
	tuiApp.QueueUpdate(func() {
		remoteTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	})
	time.Sleep(50 * time.Millisecond)
	tuiApp.QueueUpdateDraw(func() {
		remoteTable.Select(3, 0)
	})
	time.Sleep(50 * time.Millisecond)
	tuiApp.QueueUpdate(func() {
		remoteTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	})
	time.Sleep(50 * time.Millisecond)

	// --- Case B: Both files succeed, so both remote files must be deleted ---
	tuiDownloadFn = func(c *cli.Context, a *app.AuthContainer) (interface{}, error) {
		return nil, nil
	}

	deletedRemoteIDs = nil

	// Press F6 (Move)
	tuiApp.QueueUpdate(func() {
		remoteTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyF6, 0, tcell.ModNone))
	})
	time.Sleep(300 * time.Millisecond)

	if len(deletedRemoteIDs) != 2 {
		t.Errorf("Expected 2 remote files to be deleted, got %d", len(deletedRemoteIDs))
	} else {
		if (deletedRemoteIDs[0] == "id1" && deletedRemoteIDs[1] == "id2") || (deletedRemoteIDs[0] == "id2" && deletedRemoteIDs[1] == "id1") {
			// Success
		} else {
			t.Errorf("Expected deleted file IDs to be id1 and id2, got: %v", deletedRemoteIDs)
		}
	}
}

func TestTUI_GASIntegration(t *testing.T) {
	envCfg := os.Getenv("GGSRUN_CFG_PATH")
	if envCfg == "" {
		t.Skip("Skipping GAS integration test because GGSRUN_CFG_PATH is not set")
	}
	cfgFile := filepath.Join(envCfg, "ggsrun.cfg")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		t.Skip("Skipping GAS integration test because ggsrun.cfg does not exist")
	}

	cfgData, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("Failed to read ggsrun.cfg: %v", err)
	}
	var cfg struct {
		ScriptID string `json:"script_id"`
	}
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		t.Fatalf("Failed to parse ggsrun.cfg: %v", err)
	}
	if cfg.ScriptID == "" {
		t.Fatalf("Script ID is not set in ggsrun.cfg")
	}

	mockLocal := []FileEntry{
		{Name: "..", Path: "/mock", IsDir: true},
		{Name: "local_script.gs", Path: "/mock/local_script.gs", IsDir: false, MimeType: "text/javascript"},
	}
	mockRemote := []FileEntry{
		{Name: "..", Path: "", IsDir: true, MimeType: "application/vnd.google-apps.folder"},
		{Name: "remote_script", Path: cfg.ScriptID, IsDir: false, MimeType: "application/vnd.google-apps.script"},
	}

	oldAuth := getAuthContainerFn
	getAuthContainerFn = app.GetAuthenticatedAuthContainer
	defer func() {
		getAuthContainerFn = oldAuth
	}()

	_, cleanup := setupTestTUI(t, mockLocal, mockRemote)
	defer cleanup()

	tempScriptFile := filepath.Join(t.TempDir(), "test_script.gs")
	scriptContent := "function main() {\n  console.log(\"ok1\");\n  return \"ok2\";\n}"
	if err := os.WriteFile(tempScriptFile, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write temp script: %v", err)
	}

	t.Run("exe1", func(t *testing.T) {
		appObj := cli.NewApp()
		set := flag.NewFlagSet("test-exe1", flag.ContinueOnError)
		set.String("scriptid", cfg.ScriptID, "")
		set.String("function", "main", "")
		set.String("value", "", "")
		set.String("stringscript", scriptContent, "")
		set.Bool("jsonparser", true, "")
		cliCtx := cli.NewContext(appObj, set, nil)

		auth := app.GetAuthenticatedAuthContainer(cliCtx)
		resp, err := app.TuiRunExe1(cliCtx, auth)
		if err != nil {
			t.Fatalf("TuiRunExe1 failed: %v\nResponse: %s", err, resp)
		}
		if !strings.Contains(resp, "ok2") {
			t.Errorf("Expected response to contain 'ok2', got: %s", resp)
		}
	})

	t.Run("exe1_ui", func(t *testing.T) {
		oldReadFile := osReadFileFn
		osReadFileFn = func(name string) ([]byte, error) {
			return []byte("function main() {\n  console.log(\"ok1\");\n  return \"ok2\";\n}"), nil
		}
		defer func() {
			osReadFileFn = oldReadFile
		}()

		var execGasCalled bool
		var gasResult string
		var gasErr error

		oldTuiRunExe1Fn := tuiRunExe1Fn
		tuiRunExe1Fn = func(ctx *cli.Context, a *app.AuthContainer) (string, error) {
			gasResult, gasErr = app.TuiRunExe1(ctx, a)
			execGasCalled = true
			return gasResult, gasErr
		}
		defer func() {
			tuiRunExe1Fn = oldTuiRunExe1Fn
		}()

		tuiApp.QueueUpdateDraw(func() {
			tuiApp.SetFocus(localTable)
			localTable.Select(2, 0)
		})
		time.Sleep(50 * time.Millisecond)

		tuiApp.QueueUpdate(func() {
			localTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
		})
		time.Sleep(100 * time.Millisecond)

		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			list := findFileList(p)
			if list != nil {
				list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			input := findInputField(p)
			if input != nil {
				input.SetText(cfg.ScriptID)
				input.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			input := findInputField(p)
			if input != nil {
				input.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			input := findInputField(p)
			if input != nil {
				input.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		for i := 0; i < 150; i++ {
			if execGasCalled {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if !execGasCalled {
			t.Fatal("Expected TuiRunExe1 to be called via UI flow")
		}
		if gasErr != nil {
			t.Fatalf("GAS execution failed: %v\nResult: %s", gasErr, gasResult)
		}
		if !strings.Contains(gasResult, "ok2") {
			t.Errorf("Expected result to contain 'ok2', got: %s", gasResult)
		}
	})

	t.Run("exe1_remote_ui", func(t *testing.T) {
		var execGasCalled bool
		var gasResult string
		var gasErr error

		oldTuiRunExe1Fn := tuiRunExe1Fn
		tuiRunExe1Fn = func(ctx *cli.Context, a *app.AuthContainer) (string, error) {
			gasResult, gasErr = app.TuiRunExe1(ctx, a)
			execGasCalled = true
			return gasResult, gasErr
		}
		defer func() {
			tuiRunExe1Fn = oldTuiRunExe1Fn
		}()

		// Focus remote panel and select the script file "remote_script" (index 2, including "..")
		tuiApp.QueueUpdateDraw(func() {
			tuiApp.SetFocus(remoteTable)
			remoteTable.Select(2, 0)
		})
		time.Sleep(100 * time.Millisecond)

		// Press Enter to open GAS project explorer
		tuiApp.QueueUpdate(func() {
			remoteTable.GetInputCapture()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
		})
		time.Sleep(800 * time.Millisecond) // Wait for GAS project explorer to fetch and display

		// Press 'e' to execute
		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			p.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone), func(p tview.Primitive) {})
		})
		time.Sleep(100 * time.Millisecond)

		// Select exe1
		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			list := findFileList(p)
			if list != nil {
				list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		// Script ID Prompt (Enter to accept default)
		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			input := findInputField(p)
			if input != nil {
				input.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		// Function Name Prompt (Enter to accept default "main")
		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			input := findInputField(p)
			if input != nil {
				input.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		// Argument Value Prompt (Enter to accept default "")
		tuiApp.QueueUpdate(func() {
			_, p := pages.GetFrontPage()
			input := findInputField(p)
			if input != nil {
				input.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})
			}
		})
		time.Sleep(100 * time.Millisecond)

		for i := 0; i < 150; i++ {
			if execGasCalled {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if !execGasCalled {
			t.Fatal("Expected TuiRunExe1 to be called via Remote UI flow")
		}
		if gasErr != nil {
			t.Fatalf("Remote GAS execution failed: %v\nResult: %s", gasErr, gasResult)
		}
		if !strings.Contains(gasResult, `"API"`) {
			t.Errorf("Expected result to contain API execution response structure, got: %s", gasResult)
		}

		// Close result modal if any to restore TUI state
		tuiApp.QueueUpdateDraw(func() {
			pages.RemovePage("gas_project")
			pages.RemovePage("info")
		})
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("exe2", func(t *testing.T) {
		appObj := cli.NewApp()
		set := flag.NewFlagSet("test-exe2", flag.ContinueOnError)
		set.String("scriptid", cfg.ScriptID, "")
		set.String("value", "", "")
		set.String("stringscript", scriptContent, "")
		set.Bool("jsonparser", true, "")
		cliCtx := cli.NewContext(appObj, set, nil)

		auth := app.GetAuthenticatedAuthContainer(cliCtx)
		resp, err := app.TuiRunExe2(cliCtx, auth)
		if err != nil {
			t.Fatalf("TuiRunExe2 failed: %v\nResponse: %s", err, resp)
		}
		if !strings.Contains(resp, "ok2") {
			t.Errorf("Expected response to contain 'ok2', got: %s", resp)
		}
	})
}

