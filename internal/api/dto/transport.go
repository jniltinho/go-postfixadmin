package dto

import "time"

// TransportResponse represents a transport entry in API responses
type TransportResponse struct {
	ID        int       `json:"id"        example:"1"`
	Domain    string    `json:"domain"    example:"example.com"`
	Transport string    `json:"transport" example:"smtp:[relay.example.com]:25"`
	Active    bool      `json:"active"    example:"true"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateTransportRequest for POST /api/v1/transports
type CreateTransportRequest struct {
	Domain    string `json:"domain"    example:"example.com"`
	Transport string `json:"transport" example:"smtp:[relay.example.com]:25"`
	Active    bool   `json:"active"    example:"true"`
}

// UpdateTransportRequest for PUT /api/v1/transports/:id
type UpdateTransportRequest struct {
	Domain    *string `json:"domain,omitempty"    example:"example.com"`
	Transport *string `json:"transport,omitempty" example:"smtp:[relay.example.com]:587"`
	Active    *bool   `json:"active,omitempty"    example:"false"`
}
