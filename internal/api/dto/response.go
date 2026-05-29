package dto

import "net/http"

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents a structured error in API responses.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewSuccessResponse creates a successful response with data.
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}

// Common error codes used across the API.
const (
	ErrCodeInvalidCredentials   = "INVALID_CREDENTIALS"
	ErrCodeUnauthorized         = "UNAUTHORIZED"
	ErrCodeForbidden            = "FORBIDDEN"
	ErrCodeValidation           = "VALIDATION_ERROR"
	ErrCodeNotFound             = "NOT_FOUND"
	ErrCodeConflict             = "CONFLICT"
	ErrCodeRateLimited          = "RATE_LIMITED"
	ErrCodeInternal             = "INTERNAL_ERROR"
	ErrCodeBadRequest           = "BAD_REQUEST"
	ErrCodeTokenExpired         = "TOKEN_EXPIRED"
	ErrCodeInvalidToken         = "INVALID_TOKEN"
	ErrCodeTooManyRequests      = "TOO_MANY_REQUESTS"
)

// HTTP status mapping for common error codes (useful for handlers)
var ErrorStatusMap = map[string]int{
	ErrCodeInvalidCredentials: http.StatusUnauthorized,
	ErrCodeUnauthorized:       http.StatusUnauthorized,
	ErrCodeForbidden:          http.StatusForbidden,
	ErrCodeValidation:         http.StatusBadRequest,
	ErrCodeNotFound:           http.StatusNotFound,
	ErrCodeConflict:           http.StatusConflict,
	ErrCodeRateLimited:        http.StatusTooManyRequests,
	ErrCodeInternal:           http.StatusInternalServerError,
	ErrCodeBadRequest:         http.StatusBadRequest,
	ErrCodeTokenExpired:       http.StatusUnauthorized,
	ErrCodeInvalidToken:       http.StatusUnauthorized,
	ErrCodeTooManyRequests:    http.StatusTooManyRequests,
}

// GetStatusForErrorCode returns the recommended HTTP status for an error code.
func GetStatusForErrorCode(code string) int {
	if status, ok := ErrorStatusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}
