package utils

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// vacationRow holds the result of the vacation + mailbox JOIN query.
type vacationRow struct {
	Email        string
	Subject      string
	Body         string
	Active       bool
	ActiveFrom   *time.Time
	ActiveUntil  *time.Time
	IntervalTime int
	Maildir      string
}

// isVacationActive returns true when the vacation is enabled and within its active period.
func isVacationActive(v vacationRow, now time.Time) bool {
	if !v.Active {
		return false
	}

	// If ActiveFrom is NULL/zero or in the past, and ActiveUntil is NULL/zero or in the future
	isAfterStart := v.ActiveFrom == nil || v.ActiveFrom.IsZero() || now.After(*v.ActiveFrom)
	isBeforeEnd := v.ActiveUntil == nil || v.ActiveUntil.IsZero() || now.Before(*v.ActiveUntil)

	return isAfterStart && isBeforeEnd
}

// generateSieve builds the Sieve script content for the vacation rule.
// interval_time is stored in seconds by PostfixAdmin (e.g. 604800 = 7 days).
func generateSieve(v vacationRow) string {
	days := v.IntervalTime / 86400
	if days <= 0 {
		days = 1
	}
	body := strings.ReplaceAll(v.Body, `"`, `'`)
	return fmt.Sprintf(`require ["vacation"];

vacation
  :days %d
  :subject "%s"
"%s";
`, days, v.Subject, body)
}

// writeSieve creates the parent directory if needed and writes the Sieve script.
func writeSieve(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// compileSieve compiles a .sieve file using the sievec binary.
func compileSieve(path string) {
	if err := exec.Command("sievec", path).Run(); err != nil {
		slog.Warn("sievec compilation failed", "path", path, "error", err)
	}
}

// removeSieve deletes the Sieve file at path if it exists.
func removeSieve(maildir string) {
	sieveDir := filepath.Join(maildir, "sieve")
	sieveFile := filepath.Join(sieveDir, "vacation.sieve")
	symlink := filepath.Join(maildir, ".dovecot.sieve")
	if _, err := os.Stat(sieveFile); os.IsNotExist(err) {
		return
	}

	if err := os.Remove(sieveFile); err != nil {
		slog.Warn("failed to remove sieve file", "path", sieveFile, "error", err)
	} else {
		slog.Info("vacation script removed", "path", sieveFile)
	}
	if err := os.Remove(symlink); err != nil {
		slog.Warn("failed to remove symlink", "path", symlink, "error", err)
	} else {
		slog.Info("symlink removed", "path", symlink)
	}
}

const vacationQuery = `
SELECT v.email, v.subject, v.body, v.active,v.activefrom, v.activeuntil, v.interval_time,
  m.maildir FROM vacation v JOIN mailbox m ON m.username = v.email`

// SyncVacationSieve reads all vacation records from the database and writes or
// removes Dovecot Sieve scripts accordingly.
// mailBase is the root directory for virtual mailboxes (e.g. "/var/vmail").
// If empty, it falls back to the config key "server.mail_base" or "/var/vmail".
func SyncVacationSieve(db *gorm.DB, mailBase string) error {
	if mailBase == "" {
		mailBase = viper.GetString("server.mail_base")
	}
	if mailBase == "" {
		mailBase = "/var/vmail"
	}

	rows, err := db.Raw(vacationQuery).Rows()
	if err != nil {
		return fmt.Errorf("vacation query failed: %w", err)
	}
	defer rows.Close()

	now := time.Now()

	for rows.Next() {
		var v vacationRow
		if err := db.ScanRows(rows, &v); err != nil {
			slog.Warn("failed to scan vacation row", "error", err)
			continue
		}

		maildir := filepath.Join(mailBase, v.Maildir, "Maildir")
		sieveDir := filepath.Join(maildir, "sieve")
		sieveFile := filepath.Join(sieveDir, "vacation.sieve")
		symlink := filepath.Join(maildir, ".dovecot.sieve")

		// Check if sieve directory exists
		if _, err := os.Stat(sieveDir); os.IsNotExist(err) {
			os.MkdirAll(sieveDir, 0755)
		}

		if isVacationActive(v, now) {
			if err := writeSieve(sieveFile, generateSieve(v)); err != nil {
				slog.Warn("failed to write sieve script", "email", v.Email, "error", err)
				continue
			}
			compileSieve(sieveFile)
			os.Remove(symlink)
			err = os.Symlink("sieve/vacation.sieve", symlink)
			if err != nil {
				slog.Warn("failed to create symlink", "email", v.Email, "error", err)
			}
			slog.Info("vacation script created", "email", v.Email)
		} else {
			removeSieve(maildir)
		}
	}

	return rows.Err()
}

// SyncSingleVacationSieve reads a single vacation record from the database and writes or
// removes the Dovecot Sieve script accordingly for that specific user.
func SyncSingleVacationSieve(db *gorm.DB, email string, mailBase string) error {
	if mailBase == "" {
		mailBase = viper.GetString("server.mail_base")
	}
	if mailBase == "" {
		mailBase = "/var/vmail"
	}

	// Get maildir from mailbox just in case vacation record is deleted
	var maildir string
	if err := db.Table("mailbox").Select("maildir").Where("username = ?", email).Scan(&maildir).Error; err != nil || maildir == "" {
		slog.Error("SyncSingleVacationSieve: could not find maildir for user", "email", email, "error", err)
		return fmt.Errorf("could not find maildir for user %s", email)
	}

	maildirPath := filepath.Join(mailBase, maildir, "Maildir")
	sieveDir := filepath.Join(maildirPath, "sieve")
	sieveFile := filepath.Join(sieveDir, "vacation.sieve")
	symlink := filepath.Join(maildirPath, ".dovecot.sieve")

	var v vacationRow
	err := db.Raw(vacationQuery+" WHERE v.email = ?", email).Scan(&v).Error

	now := time.Now()

	// If missing, errored, or not active, remove the sieve script
	if err != nil || v.Email == "" || !isVacationActive(v, now) {
		removeSieve(maildirPath)
		return nil
	}

	// Check if sieve directory exists
	if _, err := os.Stat(sieveDir); os.IsNotExist(err) {
		os.MkdirAll(sieveDir, 0755)
	}

	// Write and compile the sieve script
	if err := writeSieve(sieveFile, generateSieve(v)); err != nil {
		slog.Error("SyncSingleVacationSieve: failed to write sieve script", "email", email, "error", err)
		return err
	}
	compileSieve(sieveFile)
	os.Remove(symlink)
	if err := os.Symlink("sieve/vacation.sieve", symlink); err != nil {
		slog.Warn("SyncSingleVacationSieve: failed to create symlink", "email", email, "error", err)
	}
	slog.Info("vacation script created", "email", email)

	return nil
}
