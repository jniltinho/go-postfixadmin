package admin

import (
	"fmt"
	"log/slog"
	"os"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// CleanupMaildirs iterates over the physical maildirs and removes any that do not have a corresponding database record.
func CleanupMaildirs(db *gorm.DB, baseDir string) {
	fmt.Printf("Starting cleanup of orphaned maildirs in %s...\n", baseDir)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		slog.Error("Failed to read base directory", "dir", baseDir, "error", err)
		os.Exit(1)
	}

	deletedCount := 0

	for _, domainEntry := range entries {
		if !domainEntry.IsDir() {
			continue
		}

		domain := domainEntry.Name()
		domainPath := fmt.Sprintf("%s/%s", baseDir, domain)

		userEntries, err := os.ReadDir(domainPath)
		if err != nil {
			slog.Warn("Failed to read domain directory", "domain", domain, "error", err)
			continue
		}

		for _, userEntry := range userEntries {
			if !userEntry.IsDir() {
				continue
			}

			localPart := userEntry.Name()
			username := fmt.Sprintf("%s@%s", localPart, domain)
			userPath := fmt.Sprintf("%s/%s", domainPath, localPart)

			var count int64
			if err := db.Model(&models.Mailbox{}).Where("username = ?", username).Count(&count).Error; err != nil {
				slog.Warn("Database error checking mailbox", "username", username, "error", err)
				continue
			}

			if count == 0 {
				fmt.Printf("Deleting orphaned maildir: %s\n", userPath)
				if err := os.RemoveAll(userPath); err != nil {
					slog.Error("Failed to delete directory", "path", userPath, "error", err)
				} else {
					deletedCount++
				}
			}
		}
	}

	fmt.Printf("Cleanup completed. %d orphaned maildirs were removed.\n", deletedCount)
}
