package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

// Logs renderiza a interface da página de View Logs
func (h *Handler) Logs(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	return c.Render(http.StatusOK, "logs.html", map[string]interface{}{
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  username,
		"Username":     username,
	})
}

// LogsData serve os dados paginados para o DataTables
func (h *Handler) LogsData(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	allowedDomains, _, err := utils.GetAllowedDomains(h.DB, username, isSuperAdmin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check permissions"})
	}

	draw, _ := strconv.Atoi(c.QueryParam("draw"))
	start, _ := strconv.Atoi(c.QueryParam("start"))
	length, _ := strconv.Atoi(c.QueryParam("length"))
	if length == 0 {
		length = 10
	}

	searchValue := c.QueryParam("search[value]")
	orderColumnIdx := c.QueryParam("order[0][column]")
	orderDir := c.QueryParam("order[0][dir]")

	filterAdmin := c.QueryParam("filter_admin")
	filterDomain := c.QueryParam("filter_domain")
	filterAction := c.QueryParam("filter_action")

	var totalRecords int64
	var filteredRecords int64
	var logs []models.Log

	// 1. Base query for Total Count
	totalQuery := h.DB.Model(&models.Log{})
	if !isSuperAdmin {
		if len(allowedDomains) == 0 {
			totalQuery = totalQuery.Where("1 = 0")
		} else {
			totalQuery = totalQuery.Where("domain IN ?", allowedDomains)
		}
	}
	totalQuery.Count(&totalRecords)

	// 2. Query for Filtered Count and Data
	query := h.DB.Model(&models.Log{})
	if !isSuperAdmin {
		if len(allowedDomains) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("domain IN ?", allowedDomains)
		}
	}

	// Custom Filters
	if filterAdmin != "" {
		query = query.Where("username LIKE ?", "%"+filterAdmin+"%")
	}
	if filterDomain != "" {
		query = query.Where("domain LIKE ?", "%"+filterDomain+"%")
	}
	if filterAction != "" {
		query = query.Where("action LIKE ?", "%"+filterAction+"%")
	}

	// Global Search
	if searchValue != "" {
		searchLike := "%" + searchValue + "%"
		query = query.Where(
			h.DB.Where("username LIKE ?", searchLike).
				Or("domain LIKE ?", searchLike).
				Or("action LIKE ?", searchLike).
				Or("data LIKE ?", searchLike),
		)
	}

	// Calculate Filtered records
	query.Count(&filteredRecords)

	// 3. Final Query for Data (recreated because Count mutated the previous one)
	dataQuery := h.DB.Model(&models.Log{})
	if !isSuperAdmin {
		if len(allowedDomains) == 0 {
			dataQuery = dataQuery.Where("1 = 0")
		} else {
			dataQuery = dataQuery.Where("domain IN ?", allowedDomains)
		}
	}
	if filterAdmin != "" {
		dataQuery = dataQuery.Where("username LIKE ?", "%"+filterAdmin+"%")
	}
	if filterDomain != "" {
		dataQuery = dataQuery.Where("domain LIKE ?", "%"+filterDomain+"%")
	}
	if filterAction != "" {
		dataQuery = dataQuery.Where("action LIKE ?", "%"+filterAction+"%")
	}
	if searchValue != "" {
		searchLike := "%" + searchValue + "%"
		dataQuery = dataQuery.Where(
			h.DB.Where("username LIKE ?", searchLike).
				Or("domain LIKE ?", searchLike).
				Or("action LIKE ?", searchLike).
				Or("data LIKE ?", searchLike),
		)
	}

	// Sorting
	columns := []string{"timestamp", "username", "domain", "action", "data"}
	orderField := "timestamp" // default
	if idx, err := strconv.Atoi(orderColumnIdx); err == nil && idx >= 0 && idx < len(columns) {
		orderField = columns[idx]
	}

	if strings.ToLower(orderDir) != "asc" && strings.ToLower(orderDir) != "desc" {
		orderDir = "desc" // default dir
	}

	dataQuery = dataQuery.Order(fmt.Sprintf("%s %s", orderField, orderDir))

	// Pagination
	if length > 0 { // length could be -1 for "All" in datatables sometimes
		dataQuery = dataQuery.Offset(start).Limit(length)
	}

	// Execute Final Query
	dataQuery.Find(&logs)

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
