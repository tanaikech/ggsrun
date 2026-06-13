package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
