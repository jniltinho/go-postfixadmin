package dto

import "time"

// DomainResponse represents a domain in API responses (includes counts for convenience)
type DomainResponse struct {
	Domain         string    `json:"domain"`
	Description    string    `json:"description"`
	Aliases        int       `json:"aliases"`
	Mailboxes      int       `json:"mailboxes"`
	MaxQuota       int64     `json:"maxquota"`
	Quota          int64     `json:"quota"`
	Transport      string    `json:"transport"`
	BackupMX       bool      `json:"backupmx"`
	Active         bool      `json:"active"`
	PasswordExpiry *int      `json:"password_expiry"`
	Created        time.Time `json:"created"`
	Modified       time.Time `json:"modified"`

	// Optional enrichment
	AliasCount   int64 `json:"alias_count,omitempty"`
	MailboxCount int64 `json:"mailbox_count,omitempty"`
}

// CreateDomainRequest for POST /api/v1/domains
type CreateDomainRequest struct {
	Domain         string `json:"domain"`
	Description    string `json:"description"`
	Aliases        int    `json:"aliases"`
	Mailboxes      int    `json:"mailboxes"`
	Quota          int64  `json:"quota"`
	Transport      string `json:"transport"`
	BackupMX       bool   `json:"backupmx"`
	Active         bool   `json:"active"`
	PasswordExpiry *int   `json:"password_expiry"`
}

// UpdateDomainRequest for PUT/PATCH /api/v1/domains/:domain
type UpdateDomainRequest struct {
	Description    *string `json:"description,omitempty"`
	Aliases        *int    `json:"aliases,omitempty"`
	Mailboxes      *int    `json:"mailboxes,omitempty"`
	Quota          *int64  `json:"quota,omitempty"`
	Transport      *string `json:"transport,omitempty"`
	BackupMX       *bool   `json:"backupmx,omitempty"`
	Active         *bool   `json:"active,omitempty"`
	PasswordExpiry *int    `json:"password_expiry,omitempty"`
}
