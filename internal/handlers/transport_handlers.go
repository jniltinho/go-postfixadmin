package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

	if err := repositories.CreateTransport(h.DB, newTransport); err != nil {
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

	if err := repositories.UpdateTransport(h.DB, id, updates); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update transport"})
	}

	SetFlash(c, "message", "Transport updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteTransport handles transport deletion
func (h *Handler) DeleteTransport(c *echo.Context) error {
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Invalid ID"})
	}

	if err := repositories.DeleteTransport(h.DB, id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to delete transport"})
	}

	SetFlash(c, "message", "Transport deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}
