// Package main (handler_download.go) :
// Advanced Google Drive concurrent downloader engine.
package app

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/sync/errgroup"
)

type downloadJob struct {
	DriveID      string
	Name         string
	SavePath     string
	Size         int64
	MimeType     string
	ExportURL    string
	ModifiedTime time.Time
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
		apiURL := "https://www.googleapis.com/drive/v3/files?q=" + escapedQuery + "&fields=nextPageToken,files(id,name,mimeType,size,modifiedTime)&pageSize=1000&includeItemsFromAllDrives=true&supportsAllDrives=true"
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
					Name:         f.Name,
					IsDir:        false,
					Path:         f.ID,
					Size:         size,
					MimeType:     f.MimeType,
					ModifiedTime: f.ModifiedTime,
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
		modTime, _ := time.Parse(time.RFC3339, node.ModifiedTime)
		*jobs = append(*jobs, downloadJob{
			DriveID:      node.Path,
			Name:         node.Name,
			SavePath:     currentPath,
			Size:         node.Size,
			MimeType:     node.MimeType,
			ModifiedTime: modTime,
		})
	}
	return nil
}

// resolveDownloadSavePath computes the final local save path and exact Drive export URL.
func resolveDownloadSavePath(job *downloadJob, c *cli.Context) bool {
	if job.MimeType == "application/vnd.google-apps.script" {
		job.ExportURL = "https://script.googleapis.com/v1/projects/" + job.DriveID + "/content"
		ext := "json"
		job.SavePath += "." + ext
		job.Name += "." + ext
		return true
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
			return false
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

		job.ExportURL = "https://www.googleapis.com/drive/v3/files/" + job.DriveID + "/export?mimeType=" + exportMime
		job.SavePath += "." + ext
		job.Name += "." + ext
		return true
	}

	job.ExportURL = "https://www.googleapis.com/drive/v3/files/" + job.DriveID + "?alt=media&supportsAllDrives=true"
	return true
}

// executeDownloadJob safely processes a single download task with fault isolation and explicit API routing.
func executeDownloadJob(job downloadJob, a *AuthContainer, c *cli.Context, progress *mpb.Progress) error {
	var resp2 *http.Response
	var reqErr error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req2, _ := http.NewRequest("GET", job.ExportURL, nil)
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
func concurrentDownload(c *cli.Context, a *AuthContainer) (interface{}, error) {
	p := a.defDownloadContainer(c)
	fileIDsStr := c.String("fileid")

	if fileIDsStr == "" || c.Bool("zip") || c.Bool("rawdata") || c.String("query") != "" || c.Bool("showfilelist") || c.String("mimetype") != "" {
		return p.GetFileinf().Downloader(c), nil
	}

	fileIDs := regexp.MustCompile(`\s*,\s*`).Split(fileIDsStr, -1)
	if len(fileIDs) == 0 {
		return TransferResult{Message: []string{"No files provided for download."}}, nil
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
		req, _ := http.NewRequest("GET", "https://www.googleapis.com/drive/v3/files/"+id+"?fields=id,name,mimeType,size,modifiedTime&supportsAllDrives=true", nil)
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
			modTime, _ := time.Parse(time.RFC3339, meta.ModifiedTime)

			savePath := meta.Name
			if len(filenames) > i && strings.TrimSpace(filenames[i]) != "" {
				savePath = strings.TrimSpace(filenames[i])
			}

			jobs = append(jobs, downloadJob{
				DriveID:      id,
				Name:         meta.Name,
				SavePath:     savePath,
				Size:         size,
				MimeType:     meta.MimeType,
				ModifiedTime: modTime,
			})
		}
	}

	if len(jobs) == 0 {
		return TransferResult{Message: []string{"No files matched for processing."}}, nil
	}

	// --- Pre-computation Conflict Resolution Matrix ---
	conflictMode := c.String("conflict-mode")
	if conflictMode == "" {
		if c.Bool("overwrite") {
			conflictMode = "overwrite"
		} else if c.Bool("skip") {
			conflictMode = "skip"
		}
	}

	var finalJobs []downloadJob
	var skippedFiles []TransferFileMetadata
	var pendingFiles []TransferFileMetadata

	for _, job := range jobs {
		if !resolveDownloadSavePath(&job, c) {
			continue // Immediately drop unexportable objects
		}

		stat, err := os.Stat(job.SavePath)
		if err == nil {
			mode := conflictMode
			if mode == "" { // Interactive mode fallback or JSON parsing safety
				if c.Bool("jsonparser") {
					// Safe partial failure: Separate this conflicting job and ask the agent to verify it.
					pendingFiles = append(pendingFiles, TransferFileMetadata{
						Name: job.Name, FileID: job.DriveID, MimeType: job.MimeType, Path: job.SavePath, Status: "pending_conflict_locally",
					})
					continue
				}
				mode, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText(fmt.Sprintf("Conflict detected: '%s' exists locally. Action?", job.SavePath)).
					WithOptions([]string{"skip", "overwrite", "rename", "update"}).
					Show()
			}

			switch mode {
			case "skip":
				pterm.Info.Printf("Skipped download: %s\n", job.SavePath)
				skippedFiles = append(skippedFiles, TransferFileMetadata{
					Name: job.Name, FileID: job.DriveID, MimeType: job.MimeType, Path: job.SavePath, Status: "skipped (user chose skip)",
				})
				continue
			case "overwrite":
				// Proceed. Target will be overwritten natively.
			case "rename":
				dir := filepath.Dir(job.SavePath)
				base := filepath.Base(job.SavePath)
				ext := filepath.Ext(base)
				nameWithoutExt := strings.TrimSuffix(base, ext)
				ts := time.Now().Format("20060102_150405")
				job.SavePath = filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, ts, ext))
			case "update":
				if !job.ModifiedTime.IsZero() && job.ModifiedTime.After(stat.ModTime()) {
					// Source is newer. Proceed to overwrite.
				} else {
					pterm.Info.Printf("Skipped (local file is newer or equal): %s\n", job.SavePath)
					skippedFiles = append(skippedFiles, TransferFileMetadata{
						Name: job.Name, FileID: job.DriveID, MimeType: job.MimeType, Path: job.SavePath, Status: "skipped (local is newer or equal)",
					})
					continue
				}
			}
		}
		finalJobs = append(finalJobs, job)
	}

	actionReq := ""
	if len(pendingFiles) > 0 {
		actionReq = "CRITICAL INSTRUCTION FOR AGENT: Some files had naming conflicts and were temporarily skipped. You MUST NOT automatically guess or retry with a conflict-mode. You MUST explicitly present these files to the user and ask how to handle them (skip, overwrite, rename, or update). Once the user decides, execute the tool again ONLY for the files in the 'pendingConflicts' list using the user's chosen 'conflict-mode'."
	}

	if len(finalJobs) == 0 {
		msg := "No jobs require immediate execution. All items were successfully processed, skipped, or are pending conflict resolution."
		pterm.Success.Println(msg)
		return TransferResult{
			Message:          []string{msg},
			Files:            skippedFiles,
			PendingConflicts: pendingFiles,
			ActionRequired:   actionReq,
		}, nil
	}
	jobs = finalJobs

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

	var mu sync.Mutex
	var successFiles []TransferFileMetadata

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

				mu.Lock()
				successFiles = append(successFiles, TransferFileMetadata{
					Name:     job.Name,
					FileID:   job.DriveID,
					MimeType: job.MimeType,
					Size:     job.Size,
					Path:     job.SavePath,
					Status:   "downloaded",
				})
				mu.Unlock()
			}
			return nil
		})
	}

	err := g.Wait()
	progress.Wait()

	resultFiles := append(successFiles, skippedFiles...)

	if err != nil {
		pterm.Error.Println("Critical Error:", err)
		return TransferResult{
			Message:          []string{"Concurrent Batch Execution Terminated: " + err.Error()},
			Files:            resultFiles,
			PendingConflicts: pendingFiles,
			ActionRequired:   actionReq,
		}, err
	}

	pterm.Success.Println("All concurrent jobs executed. Please review logs for any skipped files.")
	return TransferResult{
		Message:          []string{"Bulk operations processed successfully."},
		Files:            resultFiles,
		PendingConflicts: pendingFiles,
		ActionRequired:   actionReq,
	}, nil
}

// downloadFiles : Download files from Google Drive using concurrent parallel architecture.
func downloadFiles(c *cli.Context) error {
	a := defAuthContainer(c).ggsrunIni(c).goauth()
	res, err := concurrentDownload(c, a)
	if err != nil {
		return err
	}
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}
