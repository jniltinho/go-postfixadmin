package mailbox

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ImportCSV imports mailbox users from a CSV file.
// Expected columns: user,password,domain,name
func ImportCSV(db *gorm.DB, path string, quotaMB int64) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	f, err := os.Open(path)
	if err != nil {
		slog.Error("Failed to open CSV file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		slog.Error("Failed to read CSV file", "error", err)
		os.Exit(1)
	}

	if len(records) < 2 {
		slog.Error("CSV file is empty or has no data rows")
		os.Exit(1)
	}

	// Skip header row
	rows := records[1:]

	imported, skipped := 0, 0

	for _, row := range rows {
		if len(row) < 4 {
			slog.Warn("Skipping row with insufficient columns", "row", row)
			skipped++
			continue
		}

		localPart := row[0]
		password := row[1]
		domain := row[2]
		name := row[3]
		username := localPart + "@" + domain

		if name == "" {
			name = cases.Title(language.English).String(localPart)
		}

		// Skip if mailbox already exists
		var existing models.Mailbox
		if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
			slog.Warn("Mailbox already exists, skipping", "username", username)
			skipped++
			continue
		}

		if len(password) < 8 {
			slog.Warn("Password too short, skipping", "username", username)
			skipped++
			continue
		}

		crypted, err := utils.HashPassword(password)
		if err != nil {
			slog.Warn("Failed to hash password, skipping", "username", username, "error", err)
			skipped++
			continue
		}

		now := time.Now()
		mb := models.Mailbox{
			Username:      username,
			Password:      crypted,
			Name:          name,
			Maildir:       domain + "/" + localPart + "/",
			LocalPart:     localPart,
			Domain:        domain,
			Created:       now,
			Modified:      now,
			Quota:         quotaMB * utils.GetQuotaMultiplier(),
			Active:        true,
			SMTPActive:    true,
			TokenValidity: now.Add(3 * time.Hour),
		}

		if err := repositories.CreateMailbox(db, mb, "CLI", "127.0.0.1"); err != nil {
			slog.Warn("Failed to create mailbox, skipping", "username", username, "error", err)
			skipped++
			continue
		}

		fmt.Printf("Created: %s\n", username)
		imported++
	}

	fmt.Printf("\nImport complete: %d created, %d skipped.\n", imported, skipped)
}
