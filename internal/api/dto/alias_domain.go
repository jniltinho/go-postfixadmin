package dto

import "time"

// AliasDomainResponse represents an alias domain in API responses
type AliasDomainResponse struct {
	AliasDomain  string    `json:"alias_domain"   example:"alias.com"`
	TargetDomain string    `json:"target_domain"  example:"example.com"`
	Active       bool      `json:"active"         example:"true"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// CreateAliasDomainRequest for POST /api/v1/alias-domains
type CreateAliasDomainRequest struct {
	AliasDomain  string `json:"alias_domain"  example:"alias.com"`
	TargetDomain string `json:"target_domain" example:"example.com"`
	Active       bool   `json:"active"        example:"true"`
}

// UpdateAliasDomainRequest for PUT /api/v1/alias-domains/:alias_domain
type UpdateAliasDomainRequest struct {
	TargetDomain *string `json:"target_domain,omitempty" example:"newexample.com"`
	Active       *bool   `json:"active,omitempty"        example:"false"`
}
