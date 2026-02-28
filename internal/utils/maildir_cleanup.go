package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// CleanupOrphanedMaildir checks if a mailbox exists in the database.
// If it does not exist, but the physical directory exists on the server (/var/vmail/domain/user),
// the directory will be deleted.
// baseDir is typically "/var/vmail".
func CleanupOrphanedMaildir(db *gorm.DB, baseDir, domain, localPart string) error {
	username := fmt.Sprintf("%s@%s", localPart, domain)

	// Check if mailbox exists
	var mailbox models.Mailbox
	err := db.Where("username = ?", username).First(&mailbox).Error
	if err == nil {
		// Mailbox exists, we should not delete the folder
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		// Database error occurred
		return fmt.Errorf("error checking mailbox in database: %w", err)
	}

	// Mailbox does not exist, check if physical directory exists
	maildirPath := filepath.Join(baseDir, domain, localPart)

	if _, err := os.Stat(maildirPath); os.IsNotExist(err) {
		// Directory does not exist, nothing to do
		return nil
	}

	// Directory exists, delete it
	if err := os.RemoveAll(maildirPath); err != nil {
		return fmt.Errorf("error removing directory %s: %w", maildirPath, err)
	}

	return nil
}
