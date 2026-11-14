//go:build linux

package api

import (
	"os"
	"syscall"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

func setDirectoryMetadata(entry *contracts.DirectoryEntry, info os.FileInfo) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		links := uint64(stat.Nlink) //nolint:unconvert // Nlink is uint32 on Linux ARM64.
		owner, group := stat.Uid, stat.Gid
		entry.LinkCount, entry.Owner, entry.Group = &links, &owner, &group
	}
}
