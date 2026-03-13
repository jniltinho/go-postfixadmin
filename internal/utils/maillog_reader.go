package utils

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// filterLineRe matches postfix/filter FILTER lines, e.g.:
// Mar 11 12:21:03 mail postfix/filter[15838]: FILTER: from=... to=... size=2816 bytes=2818 host-ip=... host-name=... helo=...
var filterLineRe = regexp.MustCompile(
	`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+\S+\s+postfix/filter\[(\d+)\]:\s+FILTER:\s+` +
		`from=(\S+)\s+to=(\S+)\s+size=(\d+)\s+bytes=\d+\s+` +
		`host-ip=(\S+)\s+host-name=(\S+)\s+helo=(\S+)`,
)

// parseLogTime converts a syslog timestamp (no year) to time.Time.
// Handles Dec→Jan year rollover.
func parseLogTime(s string) (time.Time, error) {
	now := time.Now()
	t, err := time.ParseInLocation("Jan 2 15:04:05", s, now.Location())
	if err != nil {
		t, err = time.ParseInLocation("Jan  2 15:04:05", s, now.Location())
		if err != nil {
			return time.Time{}, fmt.Errorf("parseLogTime: %w", err)
		}
	}
	year := now.Year()
	// If log month is December but we are in January, the log belongs to last year
	if t.Month() == time.December && now.Month() == time.January {
		year--
	}
	return time.Date(year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location()), nil
}

// domainOf extracts the domain part from an email address.
func domainOf(email string) string {
	if i := strings.LastIndex(email, "@"); i >= 0 {
		return email[i+1:]
	}
	return email
}

// parseLine parses a single log line and returns a Maillog record if it matches.
func parseLine(line string) (*models.Maillog, bool) {
	m := filterLineRe.FindStringSubmatch(line)
	if m == nil {
		return nil, false
	}

	t, err := parseLogTime(m[1])
	if err != nil {
		slog.Warn("maillog: failed to parse timestamp", "value", m[1], "error", err)
		return nil, false
	}

	procNum, _ := strconv.Atoi(m[2])
	msgSize, _ := strconv.ParseInt(m[5], 10, 64)
	mFrom := m[3]
	mTo := m[4]
	domainFrom := domainOf(mFrom)
	domainTo := domainOf(mTo)

	return &models.Maillog{
		Tstamp:     t.Unix(),
		LogDate:    t,
		ProcNum:    procNum,
		MFrom:      mFrom,
		MTo:        mTo,
		DomainTo:   domainTo,
		Helo:       m[8],
		HostIP:     m[6],
		HostName:   m[7],
		MsgSize:    msgSize,
		DomainFrom: domainFrom,
	}, true
}

// readOffset reads the saved file offset from posFile.
// Returns 0 if the file doesn't exist or can't be read.
func readOffset(posFile string) int64 {
	data, err := os.ReadFile(posFile)
	if err != nil {
		return 0
	}
	offset, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0
	}
	return offset
}

// writeOffset saves the current file offset to posFile.
func writeOffset(posFile string, offset int64) {
	if err := os.MkdirAll(filepath.Dir(posFile), 0755); err != nil {
		slog.Warn("maillog: failed to create posfile dir", "error", err)
		return
	}
	if err := os.WriteFile(posFile, []byte(strconv.FormatInt(offset, 10)+"\n"), 0644); err != nil {
		slog.Warn("maillog: failed to write posfile", "path", posFile, "error", err)
	}
}

// PurgeMaillog deletes maillog records older than retentionDays days.
// Returns the number of rows deleted.
func PurgeMaillog(db *gorm.DB, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()
	result := db.Where("tstamp < ?", cutoff).Delete(&models.Maillog{})
	if result.Error != nil {
		return 0, fmt.Errorf("PurgeMaillog: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ReadMailLog reads new lines from logFile since the last saved offset,
// parses FILTER entries and inserts them into the maillog table.
func ReadMailLog(db *gorm.DB, logFile, posFile string) error {
	f, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("ReadMailLog: open %s: %w", logFile, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("ReadMailLog: stat: %w", err)
	}

	offset := readOffset(posFile)

	// Log rotation: file is smaller than stored offset → start from beginning
	if info.Size() < offset {
		slog.Info("maillog: log rotation detected, resetting offset")
		offset = 0
	}

	if offset > 0 {
		if _, err := f.Seek(offset, 0); err != nil {
			return fmt.Errorf("ReadMailLog: seek: %w", err)
		}
	}

	var inserted int
	currentOffset := offset
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Bytes()
		currentOffset += int64(len(raw)) + 1 // +1 for \n
		record, ok := parseLine(string(raw))
		if !ok {
			continue
		}
		if err := db.Create(record).Error; err != nil {
			slog.Warn("maillog: failed to insert record", "from", record.MFrom, "error", err)
		} else {
			inserted++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ReadMailLog: scan: %w", err)
	}

	writeOffset(posFile, currentOffset)

	slog.Info("maillog: batch complete", "inserted", inserted, "offset", currentOffset)
	return nil
}
