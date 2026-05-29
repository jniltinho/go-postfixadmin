package dto

import "time"

// AliasResponse represents an alias in API responses
type AliasResponse struct {
	Address  string    `json:"address"`
	Goto     string    `json:"goto"`
	Domain   string    `json:"domain"`
	Active   bool      `json:"active"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CreateAliasRequest for POST /api/v1/aliases
type CreateAliasRequest struct {
	LocalPart string `json:"local_part"`
	Domain    string `json:"domain"`
	Goto      string `json:"goto"` // Can be multiline or comma separated
	Active    bool   `json:"active"`
}

// UpdateAliasRequest for PUT /api/v1/aliases/:address
type UpdateAliasRequest struct {
	Goto   *string `json:"goto,omitempty"`
	Active *bool   `json:"active,omitempty"`
}
