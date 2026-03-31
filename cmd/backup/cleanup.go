package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanFiles removes .gz backup files in backupDir older than daysToKeep days.
func CleanFiles(backupDir string, daysToKeep int) {
	cutoff := time.Now().Add(-time.Duration(daysToKeep) * 24 * time.Hour)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		LogMsg(fmt.Sprintf("ERROR reading backup directory: %s", err))
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
			LogMsg(fmt.Sprintf("REMOVING file: %s older than %d days", fullPath, daysToKeep))
		}
	}
}
