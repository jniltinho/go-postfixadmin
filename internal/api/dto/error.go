package dto

import (
	"github.com/labstack/echo/v5"
)

// WriteError writes a standardized error response and sets the appropriate HTTP status.
func WriteError(c *echo.Context, code string, message string) error {
	status := GetStatusForErrorCode(code)
	resp := NewErrorResponse(code, message)
	return c.JSON(status, resp)
}

// WriteSuccess writes a standardized success response.
func WriteSuccess(c *echo.Context, data interface{}) error {
	resp := NewSuccessResponse(data)
	return c.JSON(200, resp)
}

// WriteSuccessWithStatus writes a success response with a custom HTTP status.
func WriteSuccessWithStatus(c *echo.Context, status int, data interface{}) error {
	resp := NewSuccessResponse(data)
	return c.JSON(status, resp)
}

// Common convenience error writers
func BadRequest(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeBadRequest, message)
}

func Unauthorized(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeUnauthorized, message)
}

func Forbidden(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeForbidden, message)
}

func NotFound(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeNotFound, message)
}

func ValidationError(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeValidation, message)
}

func InternalError(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeInternal, message)
}

func TooManyRequests(c *echo.Context, message string) error {
	return WriteError(c, ErrCodeTooManyRequests, message)
}
