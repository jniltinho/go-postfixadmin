package dto

// UserProfileResponse is returned by GET /api/v1/user/me
type UserProfileResponse struct {
	Username string `json:"username" example:"john@example.com"`
	Name     string `json:"name"     example:"John Doe"`
}

// UserForwardingResponse is returned by GET /api/v1/user/forwarding
type UserForwardingResponse struct {
	Goto string `json:"goto" example:"john@example.com,backup@example.com"`
}

// UpdateForwardingRequest is the body for POST /api/v1/user/forwarding
// Each forwarding address on its own line; leave empty to reset to self.
type UpdateForwardingRequest struct {
	Forwarding string `json:"forwarding" example:"john@example.com\nbackup@example.com"`
}

// UpdatePasswordRequest is the body for POST /api/v1/user/password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" example:"OldPass123!"`
	NewPassword     string `json:"new_password"     example:"NewPass456!"`
	ConfirmPassword string `json:"confirm_password" example:"NewPass456!"`
}

// VacationResponse is returned by GET /api/v1/user/vacation
type VacationResponse struct {
	Active       bool   `json:"active"         example:"true"`
	Subject      string `json:"subject"        example:"Out of Office"`
	Body         string `json:"body"           example:"I am away until Friday."`
	ActiveFrom   string `json:"activefrom"     example:"2026-06-01T09:00"`
	ActiveUntil  string `json:"activeuntil"    example:"2026-06-08T09:00"`
	IntervalTime int    `json:"interval_time"  example:"86400"`
}

// UpdateVacationRequest is the body for POST /api/v1/user/vacation
type UpdateVacationRequest struct {
	Active       bool   `json:"active"        example:"true"`
	Subject      string `json:"subject"       example:"Out of Office"`
	Body         string `json:"body"          example:"I am away until Friday."`
	ActiveFrom   string `json:"activefrom"    example:"2026-06-01T09:00"`
	ActiveUntil  string `json:"activeuntil"   example:"2026-06-08T09:00"`
	IntervalTime int    `json:"interval_time" example:"86400"`
}

// UserSuccessResponse is the generic success response for User Portal mutations
type UserSuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// UserErrorResponse is the error response for User Portal mutations
type UserErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error"   example:"Current password is incorrect"`
}
