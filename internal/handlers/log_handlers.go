package handlers

import (
	"net/http"
	"strconv"

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
