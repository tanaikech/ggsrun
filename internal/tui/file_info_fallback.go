//go:build !linux && !darwin && !windows

package tui

import (
	"os"
	"time"
)

func getLocalCreatedTime(info os.FileInfo) time.Time {
	return info.ModTime()
}

func getPlatformDetails(path string) string {
	return ""
}
