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
	isAfterStart := v.ActiveFrom == nil || v.ActiveFrom.IsZero() || now.After(*v.ActiveFrom)
	isBeforeEnd := v.ActiveUntil == nil || v.ActiveUntil.IsZero() || now.Before(*v.ActiveUntil)
	return isAfterStart && isBeforeEnd
}

// generateSieve builds the Sieve script content for the vacation rule.
// interval_time is stored in seconds (e.g. 604800 = 7 days).
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

// removeSieve deletes vacation.sieve and the .dovecot.sieve symlink for the given maildir.
func removeSieve(maildir string) {
	sieveFile := filepath.Join(maildir, "sieve", "vacation.sieve")
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

// applySieve writes or removes the Dovecot Sieve script for a single vacationRow.
func applySieve(v vacationRow, mailBase string, now time.Time) {
	maildir := filepath.Join(mailBase, v.Maildir, "Maildir")
	sieveFile := filepath.Join(maildir, "sieve", "vacation.sieve")
	symlink := filepath.Join(maildir, ".dovecot.sieve")

	if !isVacationActive(v, now) {
		removeSieve(maildir)
		return
	}

	if err := writeSieve(sieveFile, generateSieve(v)); err != nil {
		slog.Warn("failed to write sieve script", "email", v.Email, "error", err)
		return
	}
	compileSieve(sieveFile)
	os.Remove(symlink)
	if err := os.Symlink("sieve/vacation.sieve", symlink); err != nil {
		slog.Warn("failed to create symlink", "email", v.Email, "error", err)
	}
	slog.Info("vacation script created", "email", v.Email)
}

// resolveMailBase returns mailBase, falling back to config or "/var/vmail".
func resolveMailBase(mailBase string) string {
	if mailBase == "" {
		mailBase = viper.GetString("server.mail_base")
	}
	if mailBase == "" {
		return "/var/vmail"
	}
	return mailBase
}

const vacationQuery = `
SELECT v.email, v.subject, v.body, v.active, v.activefrom, v.activeuntil, v.interval_time,
  m.maildir FROM vacation v JOIN mailbox m ON m.username = v.email`

// SyncVacationSieve reads all vacation records from the database and writes or
// removes Dovecot Sieve scripts accordingly.
// mailBase is the root directory for virtual mailboxes (e.g. "/var/vmail").
// If empty, it falls back to the config key "server.mail_base" or "/var/vmail".
func SyncVacationSieve(db *gorm.DB, mailBase string) error {
	mailBase = resolveMailBase(mailBase)

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
		applySieve(v, mailBase, now)
	}

	return rows.Err()
}

// SyncSingleVacationSieve reads a single vacation record from the database and writes or
// removes the Dovecot Sieve script accordingly for that specific user.
func SyncSingleVacationSieve(db *gorm.DB, email, mailBase string) error {
	mailBase = resolveMailBase(mailBase)

	// Fetch maildir from mailbox to support removal even when vacation record is absent.
	var mailboxMaildir string
	if err := db.Table("mailbox").Select("maildir").Where("username = ?", email).Scan(&mailboxMaildir).Error; err != nil || mailboxMaildir == "" {
		return fmt.Errorf("could not find maildir for user %s", email)
	}

	var v vacationRow
	if err := db.Raw(vacationQuery+" WHERE v.email = ?", email).Scan(&v).Error; err != nil {
		slog.Warn("vacation query failed, removing sieve", "email", email, "error", err)
		removeSieve(filepath.Join(mailBase, mailboxMaildir, "Maildir"))
		return nil
	}

	if v.Email == "" || !isVacationActive(v, time.Now()) {
		removeSieve(filepath.Join(mailBase, mailboxMaildir, "Maildir"))
		return nil
	}

	applySieve(v, mailBase, time.Now())
	return nil
}
