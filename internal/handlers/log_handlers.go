package handlers

import (
	"net/http"
	"strconv"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// Logs renderiza a interface da página de View Logs
func (h *Handler) Logs(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	return c.Render(http.StatusOK, "logs/logs.html", map[string]interface{}{
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  username,
		"Username":     username,
	})
}

// LogsData serve os dados paginados para o DataTables
func (h *Handler) LogsData(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	draw, _ := strconv.Atoi(c.QueryParam("draw"))
	start, _ := strconv.Atoi(c.QueryParam("start"))
	length, _ := strconv.Atoi(c.QueryParam("length"))
	if length == 0 {
		length = 10
	}

	searchValue := c.QueryParam("search[value]")
	orderColumnIdx := c.QueryParam("order[0][column]")
	orderDir := c.QueryParam("order[0][dir]")

	columns := []string{"timestamp", "username", "domain", "action", "data"}
	orderField := "timestamp"
	if idx, err := strconv.Atoi(orderColumnIdx); err == nil && idx >= 0 && idx < len(columns) {
		orderField = columns[idx]
	}

	logs, totalRecords, filteredRecords, err := repositories.GetLogs(
		h.DB, username, isSuperAdmin,
		c.QueryParam("filter_admin"),
		c.QueryParam("filter_domain"),
		c.QueryParam("filter_action"),
		searchValue, orderField, orderDir, start, length,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch logs"})
	}

	// Build JSON response format required by DataTables
	type datatableRow struct {
		Timestamp string `json:"timestamp"`
		Username  string `json:"username"`
		Domain    string `json:"domain"`
		Action    string `json:"action"`
		Data      string `json:"data"`
	}

	data := make([]datatableRow, 0, len(logs))
	for _, l := range logs {
		data = append(data, datatableRow{
			Timestamp: l.Timestamp.Format("2006-01-02 15:04:05"),
			Username:  l.Username,
			Domain:    l.Domain,
			Action:    l.Action,
			Data:      l.Data,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"draw":            draw,
		"recordsTotal":    totalRecords,
		"recordsFiltered": filteredRecords,
		"data":            data,
	})
}

// LogsV1 godoc
// @Summary      List admin action logs
// @Description  Returns paginated admin action logs. Superadmins see all; domain admins see only their domains. Supports filtering by admin, domain, action, and free-text search.
// @Tags         Logs
// @Produce      json
// @Param        page      query     int     false  "Page number (default 1)"
// @Param        per_page  query     int     false  "Items per page (default 25)"
// @Param        search    query     string  false  "Free-text search"
// @Param        admin     query     string  false  "Filter by admin username"
// @Param        domain    query     string  false  "Filter by domain"
// @Param        action    query     string  false  "Filter by action"
// @Param        sort      query     string  false  "Sort field (default: timestamp)"
// @Param        order     query     string  false  "Sort direction: asc or desc (default: desc)"
// @Success      200       {object}  dto.PaginatedLogResponse
// @Failure      401       {object}  dto.APIResponse
// @Failure      500       {object}  dto.APIResponse
// @Router       /logs [get]
// @Security     BearerAuth
func (h *Handler) LogsV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage == 0 {
		perPage = 25
	}
	if page < 1 {
		page = 1
	}
	start := (page - 1) * perPage

	search := c.QueryParam("search")
	filterAdmin := c.QueryParam("admin")
	filterDomain := c.QueryParam("domain")
	filterAction := c.QueryParam("action")

	orderField := c.QueryParam("sort")
	if orderField == "" {
		orderField = "timestamp"
	}
	orderDir := c.QueryParam("order")
	if orderDir == "" {
		orderDir = "desc"
	}

	logs, total, filtered, err := repositories.GetLogs(
		h.DB, claims.Username, claims.Superadmin,
		filterAdmin, filterDomain, filterAction,
		search, orderField, orderDir, start, perPage,
	)
	if err != nil {
		return dto.InternalError(c, "failed to fetch logs")
	}

	data := make([]dto.LogEntryResponse, 0, len(logs))
	for _, l := range logs {
		data = append(data, dto.LogEntryResponse{
			Timestamp: l.Timestamp,
			Username:  l.Username,
			Domain:    l.Domain,
			Action:    l.Action,
			Data:      l.Data,
		})
	}

	return dto.WriteSuccess(c, dto.PaginatedResponse[dto.LogEntryResponse]{
		Data:     data,
		Total:    total,
		Filtered: filtered,
		Page:     page,
		PerPage:  perPage,
	})
}

