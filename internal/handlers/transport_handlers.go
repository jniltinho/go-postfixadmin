package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// ListTransports displays the list of transports
func (h *Handler) ListTransports(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper := middleware.GetIsSuperAdmin(c)
	if !isSuper {
		SetFlash(c, "error", "Access denied: Only Superadmins can manage transports")
		return c.Redirect(http.StatusFound, "/dashboard")
	}

	transports, err := repositories.GetAllTransports(h.DB)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "transports/list.html", map[string]interface{}{
			"Error": "Failed to fetch transports",
		})
	}

	return c.Render(http.StatusOK, "transports/list.html", map[string]interface{}{
		"Transports":   transports,
		"SessionUser":  loggedInUser,
		"IsSuperAdmin": isSuper,
		"Message":      GetFlash(c, "message"),
		"Error":        GetFlash(c, "error"),
	})
}

// AddTransportAPI processes the creation of a new transport via JSON API
func (h *Handler) AddTransportAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}
	transportName := c.FormValue("transport")
	domainName := strings.TrimSpace(c.FormValue("domain"))
	active := c.FormValue("active") == "true"

	if transportName == "" || domainName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Domain and Transport name are required"})
	}

	newTransport := models.TransportList{
		Domain:    domainName,
		Transport: transportName,
		Created:   time.Now(),
		Modified:  time.Now(),
		Active:    active,
	}

	if err := repositories.CreateTransport(h.DB, newTransport, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create transport"})
	}

	SetFlash(c, "message", "Transport created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetTransportAPI fetches a single transport details for the edit modal
func (h *Handler) GetTransportAPI(c *echo.Context) error {
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid ID"})
	}

	transport, err := repositories.GetTransportByID(h.DB, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Transport not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"transport": transport,
	})
}

// EditTransportAPI update an existing transport
func (h *Handler) EditTransportAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Invalid ID"})
	}

	transportName := c.FormValue("transport")
	domainName := strings.TrimSpace(c.FormValue("domain"))
	active := c.FormValue("active") == "true"

	if transportName == "" || domainName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Domain and Transport name are required"})
	}

	updates := map[string]interface{}{
		"domain":    domainName,
		"transport": transportName,
		"active":    active,
		"modified":  time.Now(),
	}

	if err := repositories.UpdateTransport(h.DB, id, updates, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update transport"})
	}

	SetFlash(c, "message", "Transport updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteTransport handles transport deletion
func (h *Handler) DeleteTransport(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Invalid ID"})
	}

	if err := repositories.DeleteTransport(h.DB, id, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to delete transport"})
	}

	SetFlash(c, "message", "Transport deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// =====================================================
// API v1 Transport Handlers (PR 10+)
// Note: Transports are superadmin-only
// =====================================================

// ListTransportsV1 returns all transports (superadmin only)
func (h *Handler) ListTransportsV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can manage transports")
	}

	transports, err := repositories.GetAllTransports(h.DB)
	if err != nil {
		return dto.InternalError(c, "failed to fetch transports")
	}

	var response []dto.TransportResponse
	for _, t := range transports {
		response = append(response, dto.TransportResponse{
			ID:        t.ID,
			Domain:    t.Domain,
			Transport: t.Transport,
			Active:    t.Active,
			Created:   t.Created,
			Modified:  t.Modified,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateTransportV1 handles POST /api/v1/transports
func (h *Handler) CreateTransportV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can manage transports")
	}

	var req dto.CreateTransportRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	if req.Transport == "" || req.Domain == "" {
		return dto.ValidationError(c, "domain and transport name are required")
	}

	newTransport := models.TransportList{
		Domain:    req.Domain,
		Transport: req.Transport,
		Created:   time.Now(),
		Modified:  time.Now(),
		Active:    req.Active,
	}

	if err := repositories.CreateTransport(h.DB, newTransport, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to create transport")
	}

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, map[string]int{"id": 0}) // ID will be auto-generated
}

// GetTransportV1 handles GET /api/v1/transports/:id
func (h *Handler) GetTransportV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can manage transports")
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return dto.ValidationError(c, "invalid ID")
	}

	transport, err := repositories.GetTransportByID(h.DB, id)
	if err != nil {
		return dto.NotFound(c, "transport not found")
	}

	return dto.WriteSuccess(c, transport)
}

// UpdateTransportV1 handles PUT /api/v1/transports/:id
func (h *Handler) UpdateTransportV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can manage transports")
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return dto.ValidationError(c, "invalid ID")
	}

	var req dto.UpdateTransportRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	updates := map[string]interface{}{
		"modified": time.Now(),
	}

	if req.Domain != nil {
		updates["domain"] = *req.Domain
	}
	if req.Transport != nil {
		updates["transport"] = *req.Transport
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if err := repositories.UpdateTransport(h.DB, id, updates, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to update transport")
	}

	return dto.WriteSuccess(c, map[string]bool{"updated": true})
}

// DeleteTransportV1 handles DELETE /api/v1/transports/:id
func (h *Handler) DeleteTransportV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can manage transports")
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return dto.ValidationError(c, "invalid ID")
	}

	if err := repositories.DeleteTransport(h.DB, id, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to delete transport")
	}

	return dto.WriteSuccess(c, map[string]bool{"deleted": true})
}
