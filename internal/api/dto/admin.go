package dto

import "time"

// AdminResponse represents an admin in API responses
type AdminResponse struct {
	Username      string    `json:"username"`
	Active        bool      `json:"active"`
	Superadmin    bool      `json:"superadmin"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
	DomainCount   string    `json:"domain_count,omitempty"` // "ALL" or number
}

// CreateAdminRequest for POST /api/v1/admins
type CreateAdminRequest struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	Active        bool     `json:"active"`
	Superadmin    bool     `json:"superadmin"`
	Domains       []string `json:"domains"` // only relevant if not superadmin
}

// UpdateAdminRequest for PUT /api/v1/admins/:username
type UpdateAdminRequest struct {
	Password        *string  `json:"password,omitempty"`
	PasswordConfirm *string  `json:"password_confirm,omitempty"`
	Active          *bool    `json:"active,omitempty"`
	Superadmin      *bool    `json:"superadmin,omitempty"`
	Domains         []string `json:"domains,omitempty"`
	ChangePassword  *bool    `json:"change_password,omitempty"`
}
