package dto

import "time"

// AdminResponse represents an admin in list API responses
type AdminResponse struct {
	Username    string    `json:"username"     example:"admin@example.com"`
	Active      bool      `json:"active"       example:"true"`
	Superadmin  bool      `json:"superadmin"   example:"false"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	DomainCount string    `json:"domain_count,omitempty" example:"3"` // "ALL" for superadmins
}

// AdminDomainOption is an entry in the domain assignment list returned by GET /admins/:username
type AdminDomainOption struct {
	Domain   string `json:"domain"   example:"example.com"`
	Assigned bool   `json:"assigned" example:"true"`
}

// AdminDetailResponse is returned by GET /api/v1/admins/:username
type AdminDetailResponse struct {
	Admin   AdminResponse       `json:"admin"`
	Domains []AdminDomainOption `json:"domains"`
}

// CreateAdminRequest for POST /api/v1/admins
type CreateAdminRequest struct {
	Username   string   `json:"username"   example:"admin@example.com"`
	Password   string   `json:"password"   example:"SecurePass1!"`
	Active     bool     `json:"active"     example:"true"`
	Superadmin bool     `json:"superadmin" example:"false"`
	Domains    []string `json:"domains"`
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
