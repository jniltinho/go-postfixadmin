package handlers

import (
	"net/http"
	"strconv"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// MailLog renderiza a página de visualização do maillog
func (h *Handler) MailLog(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	return c.Render(http.StatusOK, "logs/maillog.html", map[string]interface{}{
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  username,
		"Username":     username,
	})
}

// MailLogData serve os dados paginados do maillog para o DataTables
func (h *Handler) MailLogData(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	draw, _ := strconv.Atoi(c.QueryParam("draw"))
	start, _ := strconv.Atoi(c.QueryParam("start"))
	length, _ := strconv.Atoi(c.QueryParam("length"))
	if length == 0 {
		length = 15
	}

	searchValue := c.QueryParam("search[value]")
	orderColumnIdx := c.QueryParam("order[0][column]")
	orderDir := c.QueryParam("order[0][dir]")

	columns := []string{"tstamp", "m_from", "m_to", "domain_from", "domain_to", "host_ip", "host_name", "helo", "msgsize"}
	orderField := "tstamp"
	if idx, err := strconv.Atoi(orderColumnIdx); err == nil && idx >= 0 && idx < len(columns) {
		orderField = columns[idx]
	}

	logs, totalRecords, filteredRecords, err := repositories.GetMailLogs(
		h.DB, username, isSuperAdmin,
		searchValue, orderField, orderDir, start, length,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch maillogs"})
	}

	type row struct {
		LogDate    string `json:"logdate"`
		MFrom      string `json:"m_from"`
		MTo        string `json:"m_to"`
		DomainFrom string `json:"domain_from"`
		DomainTo   string `json:"domain_to"`
		HostIP     string `json:"host_ip"`
		HostName   string `json:"host_name"`
		Helo       string `json:"helo"`
		MsgSize    int64  `json:"msgsize"`
	}

	data := make([]row, 0, len(logs))
	for _, l := range logs {
		data = append(data, row{
			LogDate:    time.Unix(l.Tstamp, 0).Format("2006-01-02 15:04:05"),
			MFrom:      l.MFrom,
			MTo:        l.MTo,
			DomainFrom: l.DomainFrom,
			DomainTo:   l.DomainTo,
			HostIP:     l.HostIP,
			HostName:   l.HostName,
			Helo:       l.Helo,
			MsgSize:    l.MsgSize,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"draw":            draw,
		"recordsTotal":    totalRecords,
		"recordsFiltered": filteredRecords,
		"data":            data,
	})
}
