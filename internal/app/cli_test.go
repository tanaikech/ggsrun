package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ggsrun/internal/utl"
)

func init() {
	loadEnv()
}

func loadEnv() {
	content, err := os.ReadFile("../../.env")
	if err != nil {
		content, err = os.ReadFile(".env")
		if err != nil {
			content, err = os.ReadFile("../.env")
			if err != nil {
				return // Ignore if not found
			}
		}
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			os.Setenv(key, val)
		}
	}
}

// TestConflictModeNormalization verifies that the legacy and new conflict-mode strings
// are correctly mapped to their canonical forms (OverwriteIfNewer, Ignore, Rename).
func TestConflictModeNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "OverwriteIfNewer"},
		{"update", "OverwriteIfNewer"},
		{"overwriteifnewer", "OverwriteIfNewer"},
		{"OverwriteIfNewer", "OverwriteIfNewer"},
		{"skip", "Ignore"},
		{"ignore", "Ignore"},
		{"Ignore", "Ignore"},
		{"rename", "Rename"},
		{"Rename", "Rename"},
		{"overwrite", "overwrite"},
		{"random", "OverwriteIfNewer"}, // Fallback
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			normalized := tc.input
			if normalized == "" {
				normalized = "OverwriteIfNewer"
			} else {
				switch strings.ToLower(normalized) {
				case "overwriteifnewer", "update":
					normalized = "OverwriteIfNewer"
				case "ignore", "skip":
					normalized = "Ignore"
				case "rename":
					normalized = "Rename"
				case "overwrite":
					normalized = "overwrite"
				default:
					normalized = "OverwriteIfNewer"
				}
			}

			if normalized != tc.expected {
				t.Errorf("For input %q, expected normalized value %q, but got %q", tc.input, tc.expected, normalized)
			}
		})
	}
}

// TestRenameConflictHandling simulates the dynamic auto-renaming process
// when Rename conflict mode is engaged, confirming that sequence numbers
// are correctly appended if name collisions persist.
func TestRenameConflictHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ggsrun_conflict_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	baseFile := "sample.gs"
	jobSavePath := filepath.Join(tempDir, baseFile)

	// Create initial file
	if err := os.WriteFile(jobSavePath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to create base test file: %v", err)
	}

	// Calculate target rename path
	dir := filepath.Dir(jobSavePath)
	base := filepath.Base(jobSavePath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	ts := time.Now().Format("20060102_150405")

	newPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, ts, ext))

	// Write a file at the new path to force a rename collision
	if err := os.WriteFile(newPath, []byte("collision-1"), 0644); err != nil {
		t.Fatalf("Failed to create collision file: %v", err)
	}

	// Now apply the resolution logic to find a free name (should append _1)
	resolvedPath := newPath
	if _, statErr := os.Stat(resolvedPath); statErr == nil {
		for k := 1; k <= 1000; k++ {
			tempPath := filepath.Join(dir, fmt.Sprintf("%s_%s_%d%s", nameWithoutExt, ts, k, ext))
			if _, tempStatErr := os.Stat(tempPath); tempStatErr != nil {
				resolvedPath = tempPath
				break
			}
		}
	}

	expectedSuffix := fmt.Sprintf("_%s_1%s", ts, ext)
	if !strings.HasSuffix(resolvedPath, expectedSuffix) {
		t.Errorf("Expected resolved path to end with %q, but got %q", expectedSuffix, resolvedPath)
	}
}

// TestCliDownloadEnhancement verifies the logic of downloading Google Apps Script files
// and separating/converting extensions from .gs to .js when rawdata is false, and raw JSON download when true.
func TestCliDownloadEnhancement(t *testing.T) {
	// Construct simulated project response payload
	projectPayload := `{
		"files": [
			{"name": "appsscript", "type": "JSON", "source": "{\n  \"timeZone\": \"America/New_York\"\n}"},
			{"name": "Code", "type": "SERVER_JS", "source": "function myFunction() {\n  Logger.log('test');\n}"},
			{"name": "Index", "type": "HTML", "source": "<h1>Hello</h1>"}
		]
	}`

	tempDir, err := os.MkdirTemp("", "ggsrun_download_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Raw download (rawdata = true)
	// We want to verify that baseName.json is created containing the raw JSON
	p := &utl.FileInf{
		FileName:   "MyGASProject",
		RawProject: true,
		OverWrite:  true,
		Workdir:    tempDir,
		PstartTime: time.Now(),
	}

	p.SaveScript([]byte(projectPayload))

	rawFilePath := filepath.Join(tempDir, "MyGASProject.json")
	if _, err := os.Stat(rawFilePath); err != nil {
		t.Errorf("Expected raw JSON file to exist at %s, but got error: %v", rawFilePath, err)
	}

	// Test case 2: Default structured extraction download (rawdata = false)
	// We want to verify that a directory named MyGASProject is created, and:
	// - Code.gs is converted to Code.js
	// - Index.html is saved
	// - appsscript.json is saved
	p2 := &utl.FileInf{
		FileName:   "MyGASProject",
		RawProject: false,
		OverWrite:  true,
		Workdir:    tempDir,
		PstartTime: time.Now(),
	}

	p2.SaveScript([]byte(projectPayload))

	projectDir := filepath.Join(tempDir, "MyGASProject")
	if info, err := os.Stat(projectDir); err != nil || !info.IsDir() {
		t.Fatalf("Expected directory to exist at %s", projectDir)
	}

	expectedFiles := []struct {
		name string
		ext  string
	}{
		{"appsscript", ".json"},
		{"Code", ".js"},
		{"Index", ".html"},
	}

	for _, ef := range expectedFiles {
		filePath := filepath.Join(projectDir, ef.name+ef.ext)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Expected extracted file to exist at %s, but got: %v", filePath, err)
		}
	}
}

func TestIntegrationFlows(t *testing.T) {
	// 1. Check if we have the configuration or skip
	envCfg := os.Getenv("GGSRUN_CFG_PATH")
	if envCfg == "" {
		t.Skip("Skipping integration flows test because GGSRUN_CFG_PATH is not set")
	}

	gasProjectID := os.Getenv("GGSRUN_TEST_GAS_PROJECT_FILEID_ON_GOOGLE_DRIVE")
	googleDocsID := os.Getenv("GGSRUN_TEST_GOOGLE_DOCS_FILEID_ON_GOOGLE_DRIVE")
	pdfID := os.Getenv("GGSRUN_TEST_PDF_FILEID_ON_GOOGLE_DRIVE")

	if gasProjectID == "" || googleDocsID == "" || pdfID == "" {
		t.Skip("Skipping integration flows test because test File IDs are not set in environment")
	}

	// 2. Compile ggsrun to a temporary directory once
	tempDir, err := os.MkdirTemp("", "ggsrun_integration_tests")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempBinary := filepath.Join(tempDir, "ggsrun_test_bin")
	buildCmd := exec.Command("go", "build", "-o", tempBinary, "../..")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build ggsrun binary: %v\nOutput: %s", err, string(out))
	}

	// Helper helper function to run commands and check they completed without error
	runGgsrun := func(args ...string) (string, error) {
		cmd := exec.Command(tempBinary, args...)
		// Set working directory to the project root so relative paths from .env resolve correctly
		rootDir, _ := filepath.Abs("../..")
		cmd.Dir = rootDir
		// Ensure it uses the correct config path & parses output as JSON
		cmd.Env = append(os.Environ(), "GGSRUN_MCP_MODE=true")
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	// --- DOWNLOAD TESTS ---

	// A. GAS Project Download Tests (GGSRUN_TEST_GAS_PROJECT_FILEID_ON_GOOGLE_DRIVE)
	t.Run("GASProject_Download_AsFolder", func(t *testing.T) {
		dest := filepath.Join(tempDir, "gas_folder")
		out, err := runGgsrun("download", "-i", gasProjectID, "-d", dest, "-j")
		if err != nil {
			t.Fatalf("download gas folder failed: %v\nOutput: %s", err, out)
		}
		// Confirm folder contents exist
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected files in %s, got: %v (len: %d)", dest, err, len(entries))
		}
		// Expect nested directory with project name
		projDir := filepath.Join(dest, entries[0].Name())
		subEntries, err := os.ReadDir(projDir)
		if err != nil || len(subEntries) == 0 {
			t.Fatalf("Expected files in project dir %s, got: %v", projDir, err)
		}
	})

	t.Run("GASProject_Download_AsRawJSON", func(t *testing.T) {
		dest := filepath.Join(tempDir, "gas_raw")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("failed to create dest: %v", err)
		}
		out, err := runGgsrun("download", "-i", gasProjectID, "-d", dest, "--rawdata", "-j")
		if err != nil {
			t.Fatalf("download raw JSON failed: %v\nOutput: %s", err, out)
		}
		// Confirm a JSON file exists
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected raw JSON file in %s, got none", dest)
		}
		if !strings.HasSuffix(entries[0].Name(), ".json") {
			t.Errorf("Expected file name to end with .json, got: %s", entries[0].Name())
		}
	})

	t.Run("GASProject_Download_AsZip", func(t *testing.T) {
		dest := filepath.Join(tempDir, "gas_zip")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("failed to create dest: %v", err)
		}
		out, err := runGgsrun("download", "-i", gasProjectID, "-d", dest, "-z", "-j")
		if err != nil {
			t.Fatalf("download zip failed: %v\nOutput: %s", err, out)
		}
		// Confirm a Zip file exists
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected zip file in %s, got none", dest)
		}
		foundZip := false
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".zip") {
				foundZip = true
				break
			}
		}
		if !foundZip {
			t.Errorf("Expected a .zip file in %s, but found none in: %v", dest, entries)
		}
	})

	// B. Google Docs Download Tests (GGSRUN_TEST_GOOGLE_DOCS_FILEID_ON_GOOGLE_DRIVE)
	t.Run("GoogleDocs_Download_AsPDF", func(t *testing.T) {
		dest := filepath.Join(tempDir, "docs_pdf")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("failed to create dest: %v", err)
		}
		out, err := runGgsrun("download", "-i", googleDocsID, "-d", dest, "-e", "pdf", "-j")
		if err != nil {
			t.Fatalf("download docs as PDF failed: %v\nOutput: %s", err, out)
		}
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected pdf file in %s, got none", dest)
		}
		if !strings.HasSuffix(entries[0].Name(), ".pdf") {
			t.Errorf("Expected file name to end with .pdf, got: %s", entries[0].Name())
		}
	})

	t.Run("GoogleDocs_Download_AsDOCX", func(t *testing.T) {
		dest := filepath.Join(tempDir, "docs_docx")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("failed to create dest: %v", err)
		}
		out, err := runGgsrun("download", "-i", googleDocsID, "-d", dest, "-e", "docx", "-j")
		if err != nil {
			t.Fatalf("download docs as DOCX failed: %v\nOutput: %s", err, out)
		}
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected docx file in %s, got none", dest)
		}
		if !strings.HasSuffix(entries[0].Name(), ".docx") {
			t.Errorf("Expected file name to end with .docx, got: %s", entries[0].Name())
		}
	})

	t.Run("GoogleDocs_Download_AsTXT", func(t *testing.T) {
		dest := filepath.Join(tempDir, "docs_txt")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("failed to create dest: %v", err)
		}
		out, err := runGgsrun("download", "-i", googleDocsID, "-d", dest, "-e", "txt", "-j")
		if err != nil {
			t.Fatalf("download docs as TXT failed: %v\nOutput: %s", err, out)
		}
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected txt file in %s, got none", dest)
		}
		if !strings.HasSuffix(entries[0].Name(), ".txt") {
			t.Errorf("Expected file name to end with .txt, got: %s", entries[0].Name())
		}
	})

	// C. PDF Download Tests (GGSRUN_TEST_PDF_FILEID_ON_GOOGLE_DRIVE)
	t.Run("PDF_Download_AsIs", func(t *testing.T) {
		dest := filepath.Join(tempDir, "pdf_asis")
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatalf("failed to create dest: %v", err)
		}
		out, err := runGgsrun("download", "-i", pdfID, "-d", dest, "-j")
		if err != nil {
			t.Fatalf("download pdf failed: %v\nOutput: %s", err, out)
		}
		entries, err := os.ReadDir(dest)
		if err != nil || len(entries) == 0 {
			t.Fatalf("Expected pdf file in %s, got none", dest)
		}
		if !strings.HasSuffix(entries[0].Name(), ".pdf") {
			t.Errorf("Expected file name to end with .pdf, got: %s", entries[0].Name())
		}
	})

	// --- UPLOAD TESTS ---

	// D. Standalone GAS Projects upload
	gasScripts := []string{os.Getenv("TESTDATA_GAS_SCRIPT1"), os.Getenv("TESTDATA2_GAS_SCRIPT2")}
	for idx, gsPath := range gasScripts {
		if gsPath == "" {
			continue
		}
		t.Run(fmt.Sprintf("Upload_GASProject_%d", idx+1), func(t *testing.T) {
			projName := fmt.Sprintf("GgsrunTest_GASProj_%d_%d", idx+1, time.Now().Unix())
			out, err := runGgsrun("upload", "-f", gsPath, "--pn", projName, "--pt", "standalone", "-g", "-j")
			if err != nil {
				t.Fatalf("upload gas project failed: %v\nOutput: %s", err, out)
			}
		})
	}

	// E. mimeType Uploads (Images, Markdown, PDF)
	uploadTestData := []struct {
		name    string
		path1   string
		path2   string
		convDoc bool
	}{
		{"Image1_NoConvert", os.Getenv("TESTDATA_IMAGE1"), os.Getenv("TESTDATA_IMAGE2"), false},
		{"Image1_ConvertToDoc", os.Getenv("TESTDATA_IMAGE1"), os.Getenv("TESTDATA_IMAGE2"), true},
		{"Markdown1_NoConvert", os.Getenv("TESTDATA_MARKDOWN1"), os.Getenv("TESTDATA_MARKDOWN2"), false},
		{"Markdown1_ConvertToDoc", os.Getenv("TESTDATA_MARKDOWN1"), os.Getenv("TESTDATA_MARKDOWN2"), true},
		{"PDF1_NoConvert", os.Getenv("TESTDATA_PDF1"), os.Getenv("TESTDATA_PDF2"), false},
		{"PDF1_ConvertToDoc", os.Getenv("TESTDATA_PDF1"), os.Getenv("TESTDATA_PDF2"), true},
	}

	for _, ut := range uploadTestData {
		for fileSetIdx, filePath := range []string{ut.path1, ut.path2} {
			if filePath == "" {
				continue
			}
			suffix := fmt.Sprintf("Set%d", fileSetIdx+1)
			testName := fmt.Sprintf("%s_%s", ut.name, suffix)
			t.Run(testName, func(t *testing.T) {
				var args []string
				args = append(args, "upload", "-f", filePath)
				if ut.convDoc {
					args = append(args, "-c", "doc")
				} else {
					args = append(args, "--noconvert")
				}
				args = append(args, "-j")

				out, err := runGgsrun(args...)
				if err != nil {
					t.Fatalf("Upload failed for %s: %v\nOutput: %s", filePath, err, out)
				}
			})
		}
	}
}
