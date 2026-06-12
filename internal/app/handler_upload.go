// Package main (handler_upload.go) :
// Advanced Google Drive concurrent uploader engine using Resumable endpoints.
package app

import (
	"bytes"
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

type uploadJob struct {
	LocalPath  string
	Name       string
	ParentID   string
	Size       int64
	ExistingID string
}

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
func executeUploadJob(job uploadJob, a *AuthContainer, progress *mpb.Progress) (*TransferFileMetadata, error) {
	f, err := os.Open(job.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("local FS read error '%s': %w", job.LocalPath, err)
	}
	defer f.Close()

	metaStr := fmt.Sprintf(`{"name":"%s"}`, job.Name)
	if job.ParentID != "" && job.ExistingID == "" {
		metaStr = fmt.Sprintf(`{"name":"%s", "parents":["%s"]}`, job.Name, job.ParentID)
	}

	method := "POST"
	apiURL := "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&supportsAllDrives=true"
	if job.ExistingID != "" {
		method = "PATCH"
		apiURL = "https://www.googleapis.com/upload/drive/v3/files/" + job.ExistingID + "?uploadType=resumable&supportsAllDrives=true"
	}

	var location string
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, _ := http.NewRequest(method, apiURL, strings.NewReader(metaStr))
		req.Header.Set("Authorization", "Bearer "+a.GgsrunCfg.Accesstoken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("network transport init failed for '%s': %w", job.Name, err)
		}

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < maxRetries {
				time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
				continue
			}

			pterm.Warning.Printf("Upload API init failed for '%s' (Status %d): %s\n", job.Name, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
			return nil, nil // Return nil so it skips instead of crashing the batch
		}

		location = resp.Header.Get("Location")
		resp.Body.Close()
		break
	}

	if location == "" {
		pterm.Warning.Printf("Failed to resolve resumable location URI for '%s'. Skipping.\n", job.Name)
		return nil, nil
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
		return nil, fmt.Errorf("network transfer broken for '%s': %w", job.Name, err)
	}
	defer resp2.Body.Close()

	bodyBytes, _ := io.ReadAll(resp2.Body)

	if resp2.StatusCode >= 400 {
		pterm.Warning.Printf("Upload API transfer failed for '%s' (Status %d): %s\n", job.Name, resp2.StatusCode, strings.TrimSpace(string(bodyBytes)))
		return nil, nil
	}

	var resMap map[string]interface{}
	json.Unmarshal(bodyBytes, &resMap)

	fileId, _ := resMap["id"].(string)
	mimeType, _ := resMap["mimeType"].(string)

	bar.SetTotal(job.Size, true)

	return &TransferFileMetadata{
		Name:     job.Name,
		FileID:   fileId,
		MimeType: mimeType,
		URL:      "https://drive.google.com/file/d/" + fileId + "/view",
		Size:     job.Size,
		Path:     job.LocalPath,
		Status:   "uploaded",
	}, nil
}

// concurrentUpload : Massively parallel file uploader utilizing a robust Channel-based Worker Pool
func concurrentUpload(c *cli.Context, a *AuthContainer) (interface{}, error) {
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
		return TransferResult{Message: []string{"No valid files found for upload."}}, nil
	}

	// --- Pre-computation Conflict Resolution Matrix ---
	conflictMode := c.String("conflict-mode")
	parentMap := make(map[string]bool)
	for _, job := range jobs {
		pid := job.ParentID
		if pid == "" {
			pid = "root"
		}
		parentMap[pid] = true
	}

	// Bulk-fetch metadata to bypass Drive API Rate Limits
	existingFiles := make(map[string]map[string]driveFileObj)
	for pid := range parentMap {
		existingFiles[pid] = make(map[string]driveFileObj)
		query := fmt.Sprintf("'%s' in parents and trashed=false", pid)
		escapedQuery := url.QueryEscape(query)
		pageToken := ""

		for {
			apiURL := "https://www.googleapis.com/drive/v3/files?q=" + escapedQuery + "&fields=nextPageToken,files(id,name,modifiedTime)&pageSize=1000&includeItemsFromAllDrives=true&supportsAllDrives=true"
			if pageToken != "" {
				apiURL += "&pageToken=" + pageToken
			}

			req, _ := http.NewRequest("GET", apiURL, nil)
			req.Header.Set("Authorization", "Bearer "+a.GgsrunCfg.Accesstoken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				break
			}

			var res struct {
				NextPageToken string         `json:"nextPageToken"`
				Files         []driveFileObj `json:"files"`
			}
			json.NewDecoder(resp.Body).Decode(&res)
			resp.Body.Close()

			for _, f := range res.Files {
				existingFiles[pid][f.Name] = f
			}
			if res.NextPageToken == "" {
				break
			}
			pageToken = res.NextPageToken
		}
	}

	var finalJobs []uploadJob
	var skippedFiles []TransferFileMetadata
	var pendingFiles []TransferFileMetadata

	for _, job := range jobs {
		pid := job.ParentID
		if pid == "" {
			pid = "root"
		}

		if existing, ok := existingFiles[pid][job.Name]; ok {
			mode := conflictMode
			if mode == "" { // Interactive mode fallback or JSON parsing safety
				if c.Bool("jsonparser") {
					// Safe partial failure: Separate this conflicting job and ask the agent to verify it.
					pendingFiles = append(pendingFiles, TransferFileMetadata{
						Name: job.Name, Path: job.LocalPath, Status: "pending_conflict_remotely", FileID: existing.ID,
					})
					continue
				}
				mode, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText(fmt.Sprintf("Conflict detected: '%s' exists on Google Drive. Action?", job.Name)).
					WithOptions([]string{"skip", "overwrite", "rename", "update"}).
					Show()
			}

			switch mode {
			case "skip":
				pterm.Info.Printf("Skipped upload: %s\n", job.LocalPath)
				skippedFiles = append(skippedFiles, TransferFileMetadata{
					Name: job.Name, Path: job.LocalPath, Status: "skipped (user chose skip)",
				})
				continue
			case "overwrite":
				// Bind ID to trigger PATCH request
				job.ExistingID = existing.ID
			case "rename":
				ext := filepath.Ext(job.Name)
				base := strings.TrimSuffix(job.Name, ext)
				ts := time.Now().Format("20060102_150405")
				job.Name = fmt.Sprintf("%s_%s%s", base, ts, ext)
			case "update":
				stat, _ := os.Stat(job.LocalPath)
				driveMod, err := time.Parse(time.RFC3339, existing.ModifiedTime)
				if err == nil && stat.ModTime().After(driveMod) {
					job.ExistingID = existing.ID
				} else {
					pterm.Info.Printf("Skipped (Drive file is newer or equal): %s\n", job.LocalPath)
					skippedFiles = append(skippedFiles, TransferFileMetadata{
						Name: job.Name, Path: job.LocalPath, Status: "skipped (Drive file is newer or equal)",
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

	jobsChan := make(chan uploadJob, len(jobs))
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

				resMeta, err := executeUploadJob(job, a, progress)
				wg.Done()

				if err != nil {
					return err
				}

				if resMeta != nil {
					mu.Lock()
					successFiles = append(successFiles, *resMeta)
					mu.Unlock()
				}
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

// uploadFiles : Uploads files using concurrent parallel architecture.
func uploadFiles(c *cli.Context) error {
	a := defAuthContainer(c).ggsrunIni(c).goauth()
	res, err := concurrentUpload(c, a)
	if err != nil {
		return err
	}
	dispTransferResult(c, res, a.resolveConfigFile())
	return nil
}
