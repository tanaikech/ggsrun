package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

type ReleaseResponse struct {
	ID        int64  `json:"id"`
	UploadURL string `json:"upload_url"`
}

func getGithubToken() (string, error) {
	// 1. Check environment variables
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token, nil
	}

	// 2. Try running `gh auth token`
	cmd := exec.Command("gh", "auth", "token")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		token := strings.TrimSpace(stdout.String())
		if token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN/GH_TOKEN env var or authenticate via 'gh auth login'")
}

func main() {
	tag := flag.String("tag", "", "The tag name for the release (e.g., v5.3.11) [Required]")
	title := flag.String("title", "", "The release title (defaults to tag name)")
	notes := flag.String("notes", "", "The release notes body text")
	notesFile := flag.String("notes-file", "", "Path to a file containing the release notes")
	owner := flag.String("owner", "tanaikech", "GitHub repository owner")
	repo := flag.String("repo", "ggsrun", "GitHub repository name")
	assetsPattern := flag.String("assets", "bin/*", "Glob pattern of assets to upload")
	target := flag.String("target", "master", "Target commitish (branch or commit SHA)")
	draft := flag.Bool("draft", false, "Create the release as a draft")
	prerelease := flag.Bool("prerelease", false, "Identify the release as a prerelease")

	flag.Parse()

	if *tag == "" {
		fmt.Fprintln(os.Stderr, "Error: --tag is required.")
		flag.Usage()
		os.Exit(1)
	}

	releaseTitle := *title
	if releaseTitle == "" {
		releaseTitle = *tag
	}

	// Resolve release notes
	releaseBody := *notes
	if *notesFile != "" {
		content, err := os.ReadFile(*notesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading notes-file %s: %v\n", *notesFile, err)
			os.Exit(1)
		}
		releaseBody = string(content)
	}

	// Resolve token
	token, err := getGithubToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Creating release %s on %s/%s...\n", *tag, *owner, *repo)

	reqData := ReleaseRequest{
		TagName:         *tag,
		TargetCommitish: *target,
		Name:            releaseTitle,
		Body:            releaseBody,
		Draft:           *draft,
		Prerelease:      *prerelease,
	}

	payload, err := json.Marshal(reqData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request payload: %v\n", err)
		os.Exit(1)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", *owner, *repo)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating API request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending API request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusCreated {
		fmt.Fprintf(os.Stderr, "Failed to create release (HTTP %d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var relResp ReleaseResponse
	if err := json.Unmarshal(respBody, &relResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling response JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Release created successfully! ID: %d\n", relResp.ID)

	// Upload assets
	files, err := filepath.Glob(*assetsPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving assets glob pattern %s: %v\n", *assetsPattern, err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No assets found matching pattern: %s\n", *assetsPattern)
		return
	}

	for _, file := range files {
		// Skip directories
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			continue
		}

		filename := filepath.Base(file)
		fmt.Printf("Uploading %s...\n", filename)

		fileData, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading asset file %s: %v\n", file, err)
			continue
		}

		uploadURL := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets?name=%s", *owner, *repo, relResp.ID, filename)
		upReq, err := http.NewRequest("POST", uploadURL, bytes.NewBuffer(fileData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating asset upload request for %s: %v\n", filename, err)
			continue
		}
		upReq.Header.Set("Authorization", "Bearer "+token)
		upReq.Header.Set("Accept", "application/vnd.github.v3+json")
		upReq.Header.Set("Content-Type", "application/octet-stream")

		upResp, err := http.DefaultClient.Do(upReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading asset %s: %v\n", filename, err)
			continue
		}
		upResp.Body.Close()

		if upResp.StatusCode != http.StatusCreated {
			fmt.Fprintf(os.Stderr, "Failed to upload asset %s (HTTP %d)\n", filename, upResp.StatusCode)
		} else {
			fmt.Printf("Uploaded %s successfully\n", filename)
		}
	}

	fmt.Println("All assets processed successfully.")
}
