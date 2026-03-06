package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Vacation holds data fetched from the database for a single vacation record.
type Vacation struct {
	Email        string
	Subject      string
	Body         string
	Active       int
	ActiveFrom   time.Time
	ActiveUntil  time.Time
	IntervalTime int
	Maildir      string
}

// IsActive returns true when the vacation is enabled and within its active period.
func (v Vacation) IsActive(now time.Time) bool {
	return v.Active == 1 && now.After(v.ActiveFrom) && now.Before(v.ActiveUntil)
}

// generateSieve builds the Sieve script content for the vacation rule.
func generateSieve(v Vacation) string {
	days := v.IntervalTime
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

// compileSieve compiles a .sieve file using the sievec binary.
func compileSieve(path string) {
	if err := exec.Command("sievec", path).Run(); err != nil {
		log.Println("sievec error:", err)
	}
}

// removeSieve deletes the Sieve file at path if it exists.
func removeSieve(path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	if err := os.Remove(path); err != nil {
		log.Println("erro removendo sieve:", err)
	} else {
		log.Println("vacation removido:", path)
	}
}

// writeSieve creates the directory if needed and writes the Sieve script.
func writeSieve(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
