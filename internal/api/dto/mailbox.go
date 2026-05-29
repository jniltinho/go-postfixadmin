package dto

import "time"

// MailboxResponse represents a mailbox in API responses
type MailboxResponse struct {
	Username       string    `json:"username"`
	Domain         string    `json:"domain"`
	Name           string    `json:"name"`
	Quota          int64     `json:"quota"`
	Active         bool      `json:"active"`
	SMTPActive     bool      `json:"smtp_active"`
	EmailOther     string    `json:"email_other"`
	Created        time.Time `json:"created"`
	Modified       time.Time `json:"modified"`
	PasswordExpiry time.Time `json:"password_expiry"`
}

// CreateMailboxRequest for POST /api/v1/mailboxes
type CreateMailboxRequest struct {
	LocalPart      string `json:"local_part"`
	Domain         string `json:"domain"`
	Name           string `json:"name"`
	Password       string `json:"password"`
	Quota          int64  `json:"quota"`
	Active         bool   `json:"active"`
	SMTPActive     bool   `json:"smtp_active"`
	EmailOther     string `json:"email_other"`
	SendWelcome    bool   `json:"send_welcome"`
}

// UpdateMailboxRequest for PUT /api/v1/mailboxes/:username
type UpdateMailboxRequest struct {
	Name           *string `json:"name,omitempty"`
	Quota          *int64  `json:"quota,omitempty"`
	Active         *bool   `json:"active,omitempty"`
	SMTPActive     *bool   `json:"smtp_active,omitempty"`
	EmailOther     *string `json:"email_other,omitempty"`
	ChangePassword *bool   `json:"change_password,omitempty"`
	Password       *string `json:"password,omitempty"`
	PasswordConfirm *string `json:"password_confirm,omitempty"`
}
