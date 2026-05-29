package dto

import "time"

// AliasResponse represents an alias in API responses
type AliasResponse struct {
	Address  string    `json:"address"  example:"info@example.com"`
	Goto     string    `json:"goto"     example:"admin@example.com,backup@example.com"`
	Domain   string    `json:"domain"   example:"example.com"`
	Active   bool      `json:"active"   example:"true"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CreateAliasRequest for POST /api/v1/aliases
// goto accepts multiple recipients separated by newlines or commas
type CreateAliasRequest struct {
	LocalPart string `json:"local_part" example:"info"`
	Domain    string `json:"domain"     example:"example.com"`
	Goto      string `json:"goto"       example:"admin@example.com\nbackup@example.com"`
	Active    bool   `json:"active"     example:"true"`
}

// UpdateAliasRequest for PUT /api/v1/aliases/:address
type UpdateAliasRequest struct {
	Goto   *string `json:"goto,omitempty"   example:"admin@example.com"`
	Active *bool   `json:"active,omitempty" example:"true"`
}
