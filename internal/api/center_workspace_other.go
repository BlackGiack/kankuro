//go:build !linux

package api

import (
	"os"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

func setDirectoryMetadata(_ *contracts.DirectoryEntry, _ os.FileInfo) {}
