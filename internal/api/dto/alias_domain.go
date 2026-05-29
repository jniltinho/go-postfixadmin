package dto

import "time"

// AliasDomainResponse represents an alias domain in API responses
type AliasDomainResponse struct {
	AliasDomain  string    `json:"alias_domain"`
	TargetDomain string    `json:"target_domain"`
	Active       bool      `json:"active"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// CreateAliasDomainRequest for POST /api/v1/alias-domains
type CreateAliasDomainRequest struct {
	AliasDomain  string `json:"alias_domain"`
	TargetDomain string `json:"target_domain"`
	Active       bool   `json:"active"`
}

// UpdateAliasDomainRequest for PUT /api/v1/alias-domains/:alias_domain
type UpdateAliasDomainRequest struct {
	TargetDomain *string `json:"target_domain,omitempty"`
	Active       *bool   `json:"active,omitempty"`
}
