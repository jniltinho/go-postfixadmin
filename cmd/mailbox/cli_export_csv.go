package mailbox

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ExportCSV exports all mailboxes to a CSV file.
// Columns: user,password,domain,name,quota_mb,active
// The password column contains the stored bcrypt hash, compatible with --import-csv --password-crypt.
func ExportCSV(db *gorm.DB, path string) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	var mailboxes []models.Mailbox
	if err := db.Order("domain ASC, username ASC").Find(&mailboxes).Error; err != nil {
		slog.Error("Failed to fetch mailboxes", "error", err)
		os.Exit(1)
	}

	f, err := os.Create(path)
	if err != nil {
		slog.Error("Failed to create CSV file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"user", "password", "domain", "name", "quota_mb", "active"}); err != nil {
		slog.Error("Failed to write CSV header", "error", err)
		os.Exit(1)
	}

	multiplier := utils.GetQuotaMultiplier()

	for _, m := range mailboxes {
		quotaMB := int64(0)
		if multiplier > 0 {
			quotaMB = m.Quota / multiplier
		}
		active := "0"
		if m.Active {
			active = "1"
		}
		row := []string{
			m.LocalPart,
			m.Password,
			m.Domain,
			m.Name,
			strconv.FormatInt(quotaMB, 10),
			active,
		}
		if err := w.Write(row); err != nil {
			slog.Warn("Failed to write row, skipping", "username", m.Username, "error", err)
		}
	}

	fmt.Printf("Exported %d mailboxes to %s\n", len(mailboxes), path)
}
