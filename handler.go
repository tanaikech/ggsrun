// Package main (handler.go) :
// Handler for ggsrun
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"ggsrun/utl"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/sync/errgroup"
)

// exeAPIWithout : exe1
// Update project and Execution API withour server script.
func exeAPIWithout(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		exe1Function(c).
		executionAPIwithoutServer(c).
		esenderForExe1(c).
		dispResult(c)
	return nil
}

// exeAPIWith : exe2
// No update project. Only execute GAS using Execution API with server script.
func exeAPIWith(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		exe2Function(c).
		dispResult(c)
	return nil
}

// webAppsWith : exe3
// No update project. Only execute GAS using Web Apps with server script.
func webAppsWith(c *cli.Context) error {
	defExecutionContainerWebApps().
		webAppswithServerForExe3(utl.ConvGasToRun(c), c).
		dispResult(c)
	return nil
}

// --- Data Structures for Recursive Transfers ---

type transferNode struct {
	Name     string
	IsDir    bool
	Path     string // Local path or Drive ID
	Size     int64
	MimeType string
	Children []*transferNode
}

type uploadJob struct {
	LocalPath string
	Name      string
	ParentID  string
	Size      int64
}

type downloadJob struct {
	DriveID  string
	Name     string
	SavePath string
	Size     int64
	MimeType string
}

type driveFileObj struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Size     string `json:"size"`
}

// printTransferTree visually outputs the directory structure to the terminal.
func printTransferTree(node *transferNode, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	pterm.Info.Printf("%s%s%s\n", prefix, connector, node.Name)

	newPrefix := prefix + "│   "
	if isLast {
		newPrefix = prefix + "    "
	}

	for i, child := range node.Children {
		printTransferTree(child, newPrefix, i == len(node.Children)-1)
	}
}

// buildDriveTree recursively queries Google Drive API to map the folder structure, with full Shared Drive support.
func buildDriveTree(driveID, name, token string) (*transferNode, error) {
	node := &transferNode{
		Name:  name,
		IsDir: true,
		Path:  driveID,
	}

	query := fmt.Sprintf("'%s' in parents and trashed=false", driveID)
	escapedQuery := url.QueryEscape(query)

	pageToken := ""
	for {
		// CRITICAL: includeItemsFromAllDrives and supportsAllDrives are mandatory for Shared Drives traversal.
		apiURL := "https://www.googleapis.com/drive/v3/files?q=" + escapedQuery + "&fields=nextPageToken,files(id,name,mimeType,size)&pageSize=1000&includeItemsFromAllDrives=true&supportsAllDrives=true"
		if pageToken != "" {
			apiURL += "&pageToken=" + pageToken
		}

		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		var res struct {
			NextPageToken string         `json:"nextPageToken"`
			Files         []driveFileObj `json:"files"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		for _, f := range res.Files {
			if f.MimeType == "application/vnd.google-apps.folder" {
				childNode, err := buildDriveTree(f.ID, f.Name, token)
				if err == nil {
					node.Children = append(node.Children, childNode)
				}
			} else {
				size, _ := strconv.ParseInt(f.Size, 10, 64)
				node.Children = append(node.Children, &transferNode{
					Name:     f.Name,
					IsDir:    false,
					Path:     f.ID,
					Size:     size,
					MimeType: f.MimeType,
				})
			}
		}

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	return node, nil
}

// extractDownloadJobsAndCreateDirs walks the node tree, creating local dirs and flattening file jobs.
func extractDownloadJobsAndCreateDirs(node *transferNode, localParentPath string, jobs *[]downloadJob) error {
	currentPath := filepath.Join(localParentPath, node.Name)
	if node.IsDir {
		if err := os.MkdirAll(currentPath, 0755); err != nil {
			return err
		}
		for _, child := range node.Children {
			extractDownloadJobsAndCreateDirs(child, currentPath, jobs)
		}
	} else {
		*jobs = append(*jobs, downloadJob{
			DriveID:  node.Path,
			Name:     node.Name,
			SavePath: currentPath,
			Size:     node.Size,
			MimeType: node.MimeType,
		})
	}
	return nil
}

// executeDownloadJob safely processes a single download task with fault isolation and explicit API routing.
func executeDownloadJob(job downloadJob, a *AuthContainer, c *cli.Context, progress *mpb.Progress) error {
	var downloadURL string

	// Explicit routing: GAS projects CANNOT be exported via Drive API. They require Apps Script API.
	if job.MimeType == "application/vnd.google-apps.script" {
		downloadURL = "https://script.googleapis.com/v1/projects/" + job.DriveID + "/content"
		ext := "json" // GAS payload structure is natively JSON
		job.SavePath += "." + ext
		job.Name += "." + ext
	} else if strings.Contains(job.MimeType, "application/vnd.google-apps") {
		// Prevent guaranteed 400 Bad Request by skipping inherently un-exportable types
		unexportable := map[string]bool{
			"application/vnd.google-apps.shortcut": true,
			"application/vnd.google-apps.site":     true,
			"application/vnd.google-apps.map":      true,
			"application/vnd.google-apps.form":     true,
			"application/vnd.google-apps.folder":   true,
		}

		if unexportable[job.MimeType] {
			pterm.Warning.Printf("Skipped unexportable Workspace entity '%s' (Type: %s)\n", job.Name, job.MimeType)
			return nil // Isolated skip. Does not crash the errgroup.
		}

		ext := c.String("extension")
		var exportMime string

		if ext != "" {
			switch strings.ToLower(ext) {
			case "xlsx":
				exportMime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			case "docx":
				exportMime = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
			case "pptx":
				exportMime = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
			case "csv":
				exportMime = "text/csv"
			case "json":
				exportMime = "application/json"
			default:
				exportMime = "application/" + ext
			}
		} else {
			switch job.MimeType {
			case "application/vnd.google-apps.spreadsheet":
				ext = "pdf"
				exportMime = "application/pdf"
			case "application/vnd.google-apps.document", "application/vnd.google-apps.presentation", "application/vnd.google-apps.drawing":
				ext = "pdf"
				exportMime = "application/pdf"
			default:
				ext = "pdf"
				exportMime = "application/pdf"
			}
		}

		downloadURL = "https://www.googleapis.com/drive/v3/files/" + job.DriveID + "/export?mimeType=" + exportMime
		job.SavePath += "." + ext
		job.Name += "." + ext
	} else {
		// Native files download via media alt, with Shared Drive support
		downloadURL = "https://www.googleapis.com/drive/v3/files/" + job.DriveID + "?alt=media&supportsAllDrives=true"
	}

	var resp2 *http.Response
	var reqErr error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req2, _ := http.NewRequest("GET", downloadURL, nil)
		req2.Header.Set("Authorization", "Bearer "+a.GgsrunCfg.Accesstoken)
		resp2, reqErr = http.DefaultClient.Do(req2)
		if reqErr != nil {
			return fmt.Errorf("network transport failed for '%s': %w", job.Name, reqErr)
		}

		if resp2.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp2.Body)
			resp2.Body.Close()

			if resp2.StatusCode == 429 || resp2.StatusCode >= 500 {
				if attempt < maxRetries {
					time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
					continue
				}
			}

			pterm.Warning.Printf("Download API failed for '%s' (Status %d): %s\n", job.Name, resp2.StatusCode, strings.TrimSpace(string(bodyBytes)))
			return nil
		}
		break
	}
	defer resp2.Body.Close()

	size := job.Size
	if size == 0 {
		size = resp2.ContentLength
	}
	if size <= 0 {
		size = 0
	}

	bar := progress.AddBar(size,
		mpb.PrependDecorators(decor.Name(job.Name+": ", decor.WCSyncSpaceR), decor.CountersKibiByte("% .2f / % .2f")),
		mpb.AppendDecorators(decor.Percentage(), decor.EwmaETA(decor.ET_STYLE_GO, 90), decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60)),
	)

	defer func() {
		if !bar.Completed() {
			bar.Abort(false)
		}
	}()

	out, err := os.Create(job.SavePath)
	if err != nil {
		return fmt.Errorf("local FS error creating file '%s': %w", job.SavePath, err)
	}

	proxyReader := bar.ProxyReader(resp2.Body)
	written, err := io.Copy(out, proxyReader)

	proxyReader.Close()
	out.Close()

	if err != nil {
		os.Remove(job.SavePath)
		return fmt.Errorf("I/O stream interrupted for '%s': %w", job.Name, err)
	}

	bar.SetTotal(written, true)
	return nil
}

// concurrentDownload : Massively parallel file downloader utilizing a robust Channel-based Worker Pool
func concurrentDownload(c *cli.Context, a *AuthContainer) (*utl.FileInf, error) {
	p := a.defDownloadContainer(c)
	fileIDsStr := c.String("fileid")

	if fileIDsStr == "" || c.Bool("zip") || c.Bool("rawdata") || c.String("query") != "" || c.Bool("showfilelist") || c.String("mimetype") != "" {
		return p.GetFileinf().Downloader(c), nil
	}

	fileIDs := regexp.MustCompile(`\s*,\s*`).Split(fileIDsStr, -1)
	if len(fileIDs) == 0 {
		return p, nil
	}

	filenamesStr := c.String("filename")
	var filenames []string
	if filenamesStr != "" {
		filenames = regexp.MustCompile(`\s*,\s*`).Split(filenamesStr, -1)
	}

	var jobs []downloadJob

	for i, id := range fileIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		// Ensure Shared Drive support for fetching individual file metadata
		req, _ := http.NewRequest("GET", "https://www.googleapis.com/drive/v3/files/"+id+"?fields=id,name,mimeType,size&supportsAllDrives=true", nil)
		req.Header.Set("Authorization", "Bearer "+a.GgsrunCfg.Accesstoken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		var meta driveFileObj
		json.NewDecoder(resp.Body).Decode(&meta)
		resp.Body.Close()

		if meta.MimeType == "application/vnd.google-apps.folder" {
			pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithMargin(10).Println("Drive Folder Detected: " + meta.Name)
			pterm.Info.Println("Fetching folder structure from Google Drive...")
			rootNode, err := buildDriveTree(id, meta.Name, a.GgsrunCfg.Accesstoken)
			if err != nil {
				return nil, err
			}

			pterm.Info.Println("\nTarget Download Structure:")
			printTransferTree(rootNode, "", true)

			localBase := "."
			extractDownloadJobsAndCreateDirs(rootNode, localBase, &jobs)
		} else {
			size, _ := strconv.ParseInt(meta.Size, 10, 64)

			savePath := meta.Name
			if len(filenames) > i && strings.TrimSpace(filenames[i]) != "" {
				savePath = strings.TrimSpace(filenames[i])
			}

			jobs = append(jobs, downloadJob{
				DriveID:  id,
				Name:     meta.Name,
				SavePath: savePath,
				Size:     size,
				MimeType: meta.MimeType,
			})
		}
	}

	if len(jobs) == 0 {
		return p, nil
	}

	workers := c.Int("workers")
	if workers < 1 {
		workers = 5
	}

	var wg sync.WaitGroup
	wg.Add(len(jobs))

	progress := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
		mpb.WithOutput(os.Stderr),
	)

	jobsChan := make(chan downloadJob, len(jobs))
	for _, j := range jobs {
		jobsChan <- j
	}
	close(jobsChan)

	g, ctx := errgroup.WithContext(context.Background())

	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for job := range jobsChan {
				select {
				case <-ctx.Done():
					wg.Done()
					continue
				default:
				}

				err := executeDownloadJob(job, a, c, progress)

				wg.Done()

				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	err := g.Wait()
	progress.Wait()

	if err != nil {
		p.Msgar = append(p.Msgar, "Concurrent Batch Execution Terminated: "+err.Error())
		pterm.Error.Println("Critical Error:", err)
	} else {
		p.Msgar = append(p.Msgar, "Bulk operations processed successfully.")
		pterm.Success.Println("All concurrent jobs executed. Please review logs for any skipped files.")
	}
	return p, nil
}

// downloadFiles : Download files from Google Drive using concurrent parallel architecture.
func downloadFiles(c *cli.Context) error {
	a := defAuthContainer(c).ggsrunIni(c).goauth()
	res, err := concurrentDownload(c, a)
	if err != nil {
		return err
	}
	dispTransferResult(c, res)
	return nil
}

// --- Drive Upload Logic ---

// buildLocalTree recursively scans the local filesystem to construct a node tree.
func buildLocalTree(localPath string) (*transferNode, error) {
	fi, err := os.Stat(localPath)
	if err != nil {
		return nil, err
	}
	node := &transferNode{
		Name:  fi.Name(),
		IsDir: fi.IsDir(),
		Path:  localPath,
		Size:  fi.Size(),
	}
	if node.IsDir {
		entries, err := os.ReadDir(localPath)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			childPath := filepath.Join(localPath, e.Name())
			childNode, err := buildLocalTree(childPath)
			if err == nil && childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}
	return node, nil
}

// createDriveFolder creates a new directory in Google Drive with Shared Drive support.
func createDriveFolder(name, parentID, token string) (string, error) {
	metaMap := map[string]interface{}{
		"name":     name,
		"mimeType": "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		metaMap["parents"] = []string{parentID}
	}
	body, _ := json.Marshal(metaMap)
	// Enable Shared Drive uploads
	req, _ := http.NewRequest("POST", "https://www.googleapis.com/drive/v3/files?supportsAllDrives=true", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if id, ok := res["id"].(string); ok {
		return id, nil
	}
	return "", fmt.Errorf("failed to create remote folder")
}

// extractUploadJobsAndCreateFolders traverses the local tree, mimics it on Drive, and flattens file jobs.
func extractUploadJobsAndCreateFolders(node *transferNode, driveParentID, token string, jobs *[]uploadJob) error {
	if node.IsDir {
		newFolderID, err := createDriveFolder(node.Name, driveParentID, token)
		if err != nil {
			return err
		}
		for _, child := range node.Children {
			extractUploadJobsAndCreateFolders(child, newFolderID, token, jobs)
		}
	} else {
		*jobs = append(*jobs, uploadJob{
			LocalPath: node.Path,
			Name:      node.Name,
			ParentID:  driveParentID,
			Size:      node.Size,
		})
	}
	return nil
}

// executeUploadJob safely processes a single upload task with Shared Drive routing.
func executeUploadJob(job uploadJob, a *AuthContainer, progress *mpb.Progress) error {
	f, err := os.Open(job.LocalPath)
	if err != nil {
		return fmt.Errorf("local FS read error '%s': %w", job.LocalPath, err)
	}
	defer f.Close()

	metaStr := fmt.Sprintf(`{"name":"%s"}`, job.Name)
	if job.ParentID != "" {
		metaStr = fmt.Sprintf(`{"name":"%s", "parents":["%s"]}`, job.Name, job.ParentID)
	}

	var location string
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Include supportsAllDrives=true to ensure seamless uploads to Shared Drives
		req, _ := http.NewRequest("POST", "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&supportsAllDrives=true", strings.NewReader(metaStr))
		req.Header.Set("Authorization", "Bearer "+a.GgsrunCfg.Accesstoken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("network transport init failed for '%s': %w", job.Name, err)
		}

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < maxRetries {
				time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
				continue
			}

			pterm.Warning.Printf("Upload API init failed for '%s' (Status %d): %s\n", job.Name, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
			return nil
		}

		location = resp.Header.Get("Location")
		resp.Body.Close()
		break
	}

	if location == "" {
		pterm.Warning.Printf("Failed to resolve resumable location URI for '%s'. Skipping.\n", job.Name)
		return nil
	}

	bar := progress.AddBar(job.Size,
		mpb.PrependDecorators(decor.Name(job.Name+": ", decor.WCSyncSpaceR), decor.CountersKibiByte("% .2f / % .2f")),
		mpb.AppendDecorators(decor.Percentage(), decor.EwmaETA(decor.ET_STYLE_GO, 90), decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60)),
	)

	defer func() {
		if !bar.Completed() {
			bar.Abort(false)
		}
	}()

	proxyReader := bar.ProxyReader(f)
	defer proxyReader.Close()

	req2, _ := http.NewRequest("PUT", location, proxyReader)
	req2.Header.Set("Content-Length", strconv.FormatInt(job.Size, 10))
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return fmt.Errorf("network transfer broken for '%s': %w", job.Name, err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		pterm.Warning.Printf("Upload API transfer failed for '%s' (Status %d): %s\n", job.Name, resp2.StatusCode, strings.TrimSpace(string(bodyBytes)))
		return nil
	}

	bar.SetTotal(job.Size, true)
	return nil
}

// concurrentUpload : Massively parallel file uploader utilizing a robust Channel-based Worker Pool
func concurrentUpload(c *cli.Context, a *AuthContainer) (*utl.FileInf, error) {
	p := a.defUploadContainer(c)
	filenamesStr := c.String("filename")

	if filenamesStr == "" || c.String("projecttype") != "standalone" || c.String("convertto") != "" || c.Bool("noconvert") || c.String("parentid") != "" {
		return p.Uploader(c), nil
	}

	filenames := regexp.MustCompile(`\s*,\s*`).Split(filenamesStr, -1)
	var jobs []uploadJob

	for _, fname := range filenames {
		fname = strings.TrimSpace(fname)
		if fname == "" {
			continue
		}

		fi, err := os.Stat(fname)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).WithMargin(10).Println("Local Directory Detected: " + fname)
			rootNode, err := buildLocalTree(fname)
			if err != nil {
				return nil, err
			}

			pterm.Info.Println("\nTarget Upload Structure:")
			printTransferTree(rootNode, "", true)

			pterm.Info.Println("\nProvisioning hierarchical folders on Google Drive...")
			err = extractUploadJobsAndCreateFolders(rootNode, c.String("parentfolderid"), a.GgsrunCfg.Accesstoken, &jobs)
			if err != nil {
				return nil, err
			}
		} else {
			jobs = append(jobs, uploadJob{
				LocalPath: fname,
				Name:      fi.Name(),
				ParentID:  c.String("parentfolderid"),
				Size:      fi.Size(),
			})
		}
	}

	if len(jobs) == 0 {
		return p, nil
	}

	workers := c.Int("workers")
	if workers < 1 {
		workers = 5
	}

	var wg sync.WaitGroup
	wg.Add(len(jobs))

	progress := mpb.New(
		mpb.WithWaitGroup(&wg),
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
		mpb.WithOutput(os.Stderr),
	)

	jobsChan := make(chan uploadJob, len(jobs))
	for _, j := range jobs {
		jobsChan <- j
	}
	close(jobsChan)

	g, ctx := errgroup.WithContext(context.Background())

	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for job := range jobsChan {
				select {
				case <-ctx.Done():
					wg.Done()
					continue
				default:
				}

				err := executeUploadJob(job, a, progress)
				wg.Done()

				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	err := g.Wait()
	progress.Wait()

	if err != nil {
		p.Msgar = append(p.Msgar, "Concurrent Batch Execution Terminated: "+err.Error())
		pterm.Error.Println("Critical Error:", err)
	} else {
		p.Msgar = append(p.Msgar, "Bulk operations processed successfully.")
		pterm.Success.Println("All concurrent jobs executed. Please review logs for any skipped files.")
	}
	return p, nil
}

// uploadFiles : Uploads files using concurrent parallel architecture.
func uploadFiles(c *cli.Context) error {
	a := defAuthContainer(c).ggsrunIni(c).goauth()
	res, err := concurrentUpload(c, a)
	if err != nil {
		return err
	}
	dispTransferResult(c, res)
	return nil
}

// updateProject : Updates projects and scripts
func updateProject(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defExecutionContainer().
		projectUpdateControl(c)
	dispTransferResult(c, res)
	return nil
}

// revisionFiles : Retrieves revision IDs and downloads revision files.
func revisionFiles(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetRevisionList(c)
	dispTransferResult(c, res)
	return nil
}

// showFileList : Shows file list on Google Drive
func showFileList(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetFileList(c)
	dispTransferResult(c, res)
	return nil
}

// searchFilesByQueryAndRegex : Search files on Google Drive using search query and regex.
func searchFilesByQueryAndRegex(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		SearchFiles()
	dispTransferResult(c, res)
	return nil
}

// managePermissions : Manage permissions.
func managePermissions(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defPermissionsContainer(c).
		ManagePermissions()
	dispTransferResult(c, res)
	return nil
}

// getDriveInformation : Get drive information.
func getDriveInformation(c *cli.Context) error {
	res := defAuthContainer(c).
		ggsrunIni(c).
		goauth().
		defDownloadContainer(c).
		GetDriveInformation()
	dispTransferResult(c, res)
	return nil
}

// reAuth : Retrieve tokens again.
func reAuth(c *cli.Context) error {
	defAuthContainer(c).
		ggsrunIni(c).
		reAuth()
	pterm.Success.Println("Done.")
	return nil
}

// checkStatus : Health check
func checkStatus(c *cli.Context) error {
	a := defAuthContainer(c).ggsrunIni(c).goauth()
	pterm.Success.Println("Status: Authentication successful!")
	pterm.Info.Printf("Access Token valid. Length: %d characters.\n", len(a.GgsrunCfg.Accesstoken))
	pterm.Info.Printf("Expiration time: %v\n", time.Unix(a.GgsrunCfg.Expiresin, 0).Format(time.RFC3339))
	return nil
}

// sendMCPResponse securely serializes and transmits JSON-RPC results strictly over stdout.
func sendMCPResponse(id interface{}, result interface{}) {
	res := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	b, _ := json.Marshal(res)
	fmt.Println(string(b))
}

// runMCP : MCP Node over stdio
func runMCP(c *cli.Context) error {
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).Println("🤖 ggsrun MCP Server initialized")
	pterm.Info.Println("System: Go 1.26.3 concurrency engine engaged.")
	pterm.Info.Println("Status: Listening on stdin/stdout for MCP JSON-RPC messages...")
	pterm.Warning.Println("NOTE: This server acts as a pure I/O backend for LLM clients.\nNo LLM API keys are required or used by this process.")

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req map[string]interface{}
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		method, _ := req["method"].(string)
		id := req["id"]

		switch method {
		case "initialize":
			sendMCPResponse(id, map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "ggsrun-mcp-server",
					"version": "4.0.1",
				},
			})

		case "notifications/initialized":
			// Acknowledge

		case "tools/list":
			sendMCPResponse(id, map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "searchfiles",
						"description": "Search Google Drive files using query parameters (e.g., name='target' and trashed=false).",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{"type": "string"},
							},
							"required": []string{"query"},
						},
					},
					{
						"name":        "download",
						"description": "Download file(s) or folders from Drive by File ID. Use for retrieving content structure.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"fileid": map[string]interface{}{"type": "string"},
							},
							"required": []string{"fileid"},
						},
					},
					{
						"name":        "upload",
						"description": "Upload a local file or recursive folder to Google Drive.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"filename":       map[string]interface{}{"type": "string"},
								"parentfolderid": map[string]interface{}{"type": "string"},
							},
							"required": []string{"filename"},
						},
					},
					{
						"name":        "exe1",
						"description": "Execute a specific GAS function on a Google Apps Script project. Returns JSON payload.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"scriptid": map[string]interface{}{"type": "string"},
								"function": map[string]interface{}{"type": "string"},
								"value":    map[string]interface{}{"type": "string"},
							},
							"required": []string{"scriptid", "function"},
						},
					},
					{
						"name":        "filelist",
						"description": "List files or search by name exactly. Outputs file IDs.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"searchbyname": map[string]interface{}{"type": "string"},
							},
							"required": []string{"searchbyname"},
						},
					},
				},
			})

		case "tools/call":
			params, _ := req["params"].(map[string]interface{})
			name, _ := params["name"].(string)
			argsMap, _ := params["arguments"].(map[string]interface{})

			var cmdArgs []string
			cmdArgs = append(cmdArgs, name)
			for k, v := range argsMap {
				cmdArgs = append(cmdArgs, "--"+k, fmt.Sprintf("%v", v))
			}
			cmdArgs = append(cmdArgs, "--jsonparser")

			exePath, err := os.Executable()
			if err != nil {
				exePath = "ggsrun"
			}

			cmd := exec.Command(exePath, cmdArgs...)
			cmd.Stderr = os.Stderr
			out, err := cmd.Output()

			resultText := string(out)
			if err != nil {
				resultText = fmt.Sprintf("Execution Error: %v\nOutput Payload: %s", err, resultText)
			}

			sendMCPResponse(id, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": resultText,
					},
				},
			})
		}
	}

	if err := scanner.Err(); err != nil {
		pterm.Error.Printf("MCP Transport breakdown: %v\n", err)
	}

	return nil
}

// dispResult : Display result
func (e *ExecutionContainer) dispResult(c *cli.Context) {
	var dispRes []byte
	if len(e.Msg) > 0 {
		e.FeedBackData.Response.Result.Message = e.Msg
	}
	if c.Bool("jsonparser") {
		dispRes, _ = json.MarshalIndent(e.FeedBackData.Response.Result, "", "  ")
	} else {
		dispRes, _ = json.Marshal(e.FeedBackData.Response.Result)
	}
	if c.Bool("onlyresult") {
		if c.Bool("jsonparser") {
			onlyres, _ := json.MarshalIndent(e.FeedBackData.Response.Result.Result, "", "  ")
			fmt.Printf("%s\n", string(onlyres))
		} else {
			onlyres, _ := json.Marshal(e.FeedBackData.Response.Result.Result)
			fmt.Printf("%s\n", string(onlyres))
		}
	} else {
		fmt.Printf("%v\n", string(dispRes))
	}
}

// dispTransferResult : Display result
func dispTransferResult(c *cli.Context, f *utl.FileInf) {
	var dispRes []byte
	if c.Bool("jsonparser") {
		dispRes, _ = json.MarshalIndent(f, "", "  ")
	} else {
		dispRes, _ = json.Marshal(f)
	}
	fmt.Printf("%s\n", string(dispRes))
}

// commandNotFound :
func commandNotFound(c *cli.Context, command string) {
	pterm.Error.Printf("'%s' is not a %s command. Check '%s --help' or '%s -h'.\n", command, c.App.Name, c.App.Name, c.App.Name)
	os.Exit(2)
}
