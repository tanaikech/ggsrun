//go:build linux

package tui

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func getLocalCreatedTime(info os.FileInfo) time.Time {
	if sys := info.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
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
		if stat, ok := sys.(*syscall.Stat_t); ok {
			return fmt.Sprintf("\n  Device ID     : %d\n  Inode Number  : %d\n  Link Count    : %d\n  Owner UID     : %d\n  Group GID     : %d\n  Block Size    : %d\n  Blocks        : %d",
				stat.Dev, stat.Ino, stat.Nlink, stat.Uid, stat.Gid, stat.Blksize, stat.Blocks)
		}
	}
	return ""
}
