//go:build windows

package tui

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func getLocalCreatedTime(info os.FileInfo) time.Time {
	if sys := info.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Win32FileAttributeData); ok {
			return time.Unix(0, stat.CreationTime.Nanoseconds())
		}
	}
	return info.ModTime()
}

func getPlatformDetails(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	if sys := info.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Win32FileAttributeData); ok {
			return fmt.Sprintf("\n  Attributes    : 0x%X\n  Created       : %s\n  Last Access   : %s",
				stat.FileAttributes,
				time.Unix(0, stat.CreationTime.Nanoseconds()).Format("2006-01-02 15:04:05"),
				time.Unix(0, stat.LastAccessTime.Nanoseconds()).Format("2006-01-02 15:04:05"))
		}
	}
	return ""
}
