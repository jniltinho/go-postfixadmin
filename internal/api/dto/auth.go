package dto

// LoginRequest is the request body for POST /api/v1/auth/login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserInfo represents the authenticated user returned in auth responses.
type UserInfo struct {
	Username   string   `json:"username"`
	Type       string   `json:"type"` // "admin" or "mailbox"
	Superadmin bool     `json:"superadmin"`
	Domains    []string `json:"domains"`
}

// LoginResponse is the response for successful login.
type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	User        UserInfo `json:"user"`
}

// RefreshResponse is returned after a successful token refresh.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// MeResponse is the response for GET /api/v1/auth/me
type MeResponse = UserInfo
