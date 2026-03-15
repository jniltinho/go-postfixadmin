package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func cleanFiles(backupDir string, daysToKeep int) {
	cutoff := time.Now().Add(-time.Duration(daysToKeep) * 24 * time.Hour)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		logMsg(fmt.Sprintf("ERROR reading backup directory: %s", err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			fullPath := filepath.Join(backupDir, entry.Name())
			os.Remove(fullPath)
			logMsg(fmt.Sprintf("REMOVING file: %s older than %d days", fullPath, daysToKeep))
		}
	}
}
