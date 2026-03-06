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
	ActiveFrom   time.Time
	ActiveUntil  time.Time
	IntervalTime int
	Maildir      string
}

// isVacationActive returns true when the vacation is enabled and within its active period.
func isVacationActive(v vacationRow, now time.Time) bool {
	return v.Active && now.After(v.ActiveFrom) && now.Before(v.ActiveUntil)
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
func removeSieve(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}
	if err := os.Remove(path); err != nil {
		slog.Warn("failed to remove sieve file", "path", path, "error", err)
	} else {
		slog.Info("vacation script removed", "path", path)
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
		if err := rows.Scan(
			&v.Email, &v.Subject, &v.Body, &v.Active,
			&v.ActiveFrom, &v.ActiveUntil, &v.IntervalTime, &v.Maildir,
		); err != nil {
			slog.Warn("failed to scan vacation row", "error", err)
			continue
		}

		maildirPath := filepath.Join(mailBase, v.Maildir, "Maildir")
		sievePath := filepath.Join(maildirPath, ".dovecot.sieve")

		if isVacationActive(v, now) {
			if err := writeSieve(sievePath, generateSieve(v)); err != nil {
				slog.Warn("failed to write sieve script", "email", v.Email, "error", err)
				continue
			}
			compileSieve(sievePath)
			slog.Info("vacation script created", "email", v.Email)
		} else {
			removeSieve(sievePath)
		}
	}

	return rows.Err()
}
