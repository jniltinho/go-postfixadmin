package models

import "time"

// TerminalHostKey is the SSH host key pinned for the web terminal's target (trust on first use, confirmed by the
// operator); a different key later is refused until a superadmin resets it.
type TerminalHostKey struct {
	Host      string `gorm:"primaryKey;size:255"`
	Key       string `gorm:"type:text;not null"`
	CreatedAt time.Time
}

func (TerminalHostKey) TableName() string { return "terminal_host_key" }
