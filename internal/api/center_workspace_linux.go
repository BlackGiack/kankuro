//go:build linux

package api

import (
	"os"
	"syscall"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

func setDirectoryMetadata(entry *contracts.DirectoryEntry, info os.FileInfo) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		links, owner, group := stat.Nlink, stat.Uid, stat.Gid
		entry.LinkCount, entry.Owner, entry.Group = &links, &owner, &group
	}
}
