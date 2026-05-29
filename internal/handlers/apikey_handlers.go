package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"

	"github.com/labstack/echo/v5"
)

// ApiKeyResponse represents the secure serialized format of an API key
type ApiKeyResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Token     string     `json:"token,omitempty"` // populated ONLY on creation
	Preview   string     `json:"preview"`         // masked version e.g. pfa_a1b2...****
	Created   time.Time  `json:"created"`
	ExpiresAt *time.Time `json:"expires_at"`
	Active    bool       `json:"active"`
}

// CreateApiKeyRequest represents the payload to generate a new API key
type CreateApiKeyRequest struct {
	Name      string `json:"name" example:"External Integration Key"`
	DaysValid int    `json:"days_valid" example:"30"` // 0 = never expires
}

// UpdateApiKeyRequest represents the payload to update an existing API key
type UpdateApiKeyRequest struct {
	Name   *string `json:"name,omitempty" example:"Updated Key Name"`
	Active *bool   `json:"active,omitempty" example:"false"`
}

func generateRandomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pfa_" + hex.EncodeToString(b), nil
}

func getMaskedToken(fullToken string) string {
	if len(fullToken) < 12 {
		return "****"
	}
	return fullToken[:8] + "..." + "****"
}

// ListApiKeys godoc
// @Summary      List API Keys
// @Description  Retrieves all API keys for the current authenticated administrator.
// @Tags         Settings
// @Produce      json
// @Success      200 {array} ApiKeyResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      500 {object} dto.APIResponse
// @Router       /settings/apikeys [get]
// @Security     BearerAuth
func (h *Handler) ListApiKeys(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	var keys []models.AdminApiKey
	err := h.DB.Where("username = ?", claims.Username).Order("id desc").Find(&keys).Error
	if err != nil {
		return dto.InternalError(c, "failed to list api keys")
	}

	response := make([]ApiKeyResponse, 0, len(keys))
	for _, k := range keys {
		response = append(response, ApiKeyResponse{
			ID:        k.ID,
			Name:      k.Name,
			Preview:   getMaskedToken(k.Token),
			Created:   k.Created,
			ExpiresAt: k.ExpiresAt,
			Active:    k.Active,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateApiKey godoc
// @Summary      Create API Key
// @Description  Creates a new API key and returns the raw secret token once. Keep this token safe as it cannot be recovered.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Param        request body CreateApiKeyRequest true "API Key Settings"
// @Success      200 {object} ApiKeyResponse
// @Failure      400 {object} dto.APIResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      500 {object} dto.APIResponse
// @Router       /settings/apikeys [post]
// @Security     BearerAuth
func (h *Handler) CreateApiKey(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	var input CreateApiKeyRequest
	if err := c.Bind(&input); err != nil {
		return dto.BadRequest(c, "invalid input")
	}

	if input.Name == "" {
		return dto.BadRequest(c, "name is required")
	}

	rawToken, err := generateRandomToken()
	if err != nil {
		return dto.InternalError(c, "failed to generate secure key")
	}

	var expiresAt *time.Time
	if input.DaysValid > 0 {
		exp := time.Now().AddDate(0, 0, input.DaysValid)
		expiresAt = &exp
	}

	apiKey := models.AdminApiKey{
		Username:  claims.Username,
		Name:      input.Name,
		Token:     rawToken,
		ExpiresAt: expiresAt,
		Active:    true,
		Created:   time.Now(),
	}

	if err := h.DB.Create(&apiKey).Error; err != nil {
		return dto.InternalError(c, "failed to store api key")
	}

	return dto.WriteSuccess(c, ApiKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Token:     rawToken, // RETURNED RAW ONLY HERE
		Preview:   getMaskedToken(rawToken),
		Created:   apiKey.Created,
		ExpiresAt: apiKey.ExpiresAt,
		Active:    apiKey.Active,
	})
}

// UpdateApiKey godoc
// @Summary      Update API Key
// @Description  Renames or changes active/disabled status of a specific API key.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Param        id path int true "API Key ID"
// @Param        request body UpdateApiKeyRequest true "Update settings"
// @Success      200 {object} map[string]string
// @Failure      400 {object} dto.APIResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Failure      404 {object} dto.APIResponse
// @Failure      500 {object} dto.APIResponse
// @Router       /settings/apikeys/{id} [put]
// @Security     BearerAuth
func (h *Handler) UpdateApiKey(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dto.BadRequest(c, "invalid api key id")
	}

	var key models.AdminApiKey
	if err := h.DB.First(&key, id).Error; err != nil {
		return dto.NotFound(c, "api key not found")
	}

	// Verify ownership
	if key.Username != claims.Username {
		return dto.Forbidden(c, "unauthorized access to this api key")
	}

	var input UpdateApiKeyRequest
	if err := c.Bind(&input); err != nil {
		return dto.BadRequest(c, "invalid input")
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		if *input.Name == "" {
			return dto.BadRequest(c, "name cannot be empty")
		}
		updates["name"] = *input.Name
	}
	if input.Active != nil {
		updates["active"] = *input.Active
	}

	if len(updates) > 0 {
		if err := h.DB.Model(&key).Updates(updates).Error; err != nil {
			return dto.InternalError(c, "failed to update api key")
		}
	}

	return dto.WriteSuccess(c, map[string]string{"message": "api key updated successfully"})
}

// DeleteApiKey godoc
// @Summary      Delete API Key
// @Description  Revokes and permanently deletes a specific API key.
// @Tags         Settings
// @Produce      json
// @Param        id path int true "API Key ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} dto.APIResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Failure      404 {object} dto.APIResponse
// @Failure      500 {object} dto.APIResponse
// @Router       /settings/apikeys/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteApiKey(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dto.BadRequest(c, "invalid api key id")
	}

	var key models.AdminApiKey
	if err := h.DB.First(&key, id).Error; err != nil {
		return dto.NotFound(c, "api key not found")
	}

	// Verify ownership
	if key.Username != claims.Username {
		return dto.Forbidden(c, "unauthorized access to this api key")
	}

	if err := h.DB.Delete(&key).Error; err != nil {
		return dto.InternalError(c, "failed to delete api key")
	}

	return dto.WriteSuccess(c, map[string]string{"message": "api key deleted successfully"})
}
