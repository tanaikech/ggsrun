// Package main (handler_transfer.go) :
// Shared data structures and generic display logic for Drive transfer operations.
package main

import (
	"fmt"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/urfave/cli"
)

// TransferResult : Standardized JSON payload for MCP Agent to interpret batch results.
type TransferResult struct {
	Message          []string               `json:"message"`
	Files            []TransferFileMetadata `json:"files,omitempty"`
	PendingConflicts []TransferFileMetadata `json:"pendingConflicts,omitempty"`
	ActionRequired   string                 `json:"actionRequired,omitempty"`
}

// TransferFileMetadata : Explicit metadata for individual files processed or skipped.
type TransferFileMetadata struct {
	Name     string `json:"name"`
	FileID   string `json:"fileId,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	URL      string `json:"url,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Path     string `json:"localPath,omitempty"`
	Status   string `json:"status,omitempty"` // "downloaded", "uploaded", "skipped (...)", "pending_conflict"
}

// transferNode : Node for hierarchical structure
type transferNode struct {
	Name         string
	IsDir        bool
	Path         string // Local path or Drive ID
	Size         int64
	MimeType     string
	ModifiedTime string
	Children     []*transferNode
}

// driveFileObj : Google Drive file metadata
type driveFileObj struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MimeType     string `json:"mimeType"`
	Size         string `json:"size"`
	ModifiedTime string `json:"modifiedTime"`
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

// dispTransferResult : Display result generically supporting custom structs or utl.FileInf
func dispTransferResult(c *cli.Context, f interface{}) {
	var dispRes []byte
	if c.Bool("jsonparser") {
		dispRes, _ = json.MarshalIndent(f, "", "  ")
	} else {
		dispRes, _ = json.Marshal(f)
	}
	fmt.Printf("%s\n", string(dispRes))
}
