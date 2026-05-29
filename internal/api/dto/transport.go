package dto

import "time"

// TransportResponse represents a transport entry in API responses
type TransportResponse struct {
	ID        int       `json:"id"`
	Domain    string    `json:"domain"`
	Transport string    `json:"transport"`
	Active    bool      `json:"active"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateTransportRequest for POST /api/v1/transports
type CreateTransportRequest struct {
	Domain    string `json:"domain"`
	Transport string `json:"transport"`
	Active    bool   `json:"active"`
}

// UpdateTransportRequest for PUT /api/v1/transports/:id
type UpdateTransportRequest struct {
	Domain    *string `json:"domain,omitempty"`
	Transport *string `json:"transport,omitempty"`
	Active    *bool   `json:"active,omitempty"`
}
