package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

// AddFetchmailGET renders the form to create a new fetchmail entry
func (h *Handler) AddFetchmailGET(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	mailboxes, isSuper, err := utils.GetAllMailboxes(h.DB, username, middleware.GetIsSuperAdmin(c), "")
	if err != nil {
		slog.Error("Failed to fetch mailboxes", "error", err)
	}

	renderData := map[string]interface{}{
		"Mailboxes":    mailboxes,
		"Active":       true,
		"PollTime":     10,
		"SessionUser":  username,
		"IsSuperAdmin": isSuper,
	}

	return c.Render(http.StatusOK, "add_fetchmail.html", renderData)
}

// AddFetchmailPOST processes the form submission to create a new fetchmail entry
func (h *Handler) AddFetchmailPOST(c *echo.Context) error {
	mailbox := c.FormValue("mailbox")
	srcServer := c.FormValue("src_server")
	srcAuth := c.FormValue("src_auth")
	srcUser := c.FormValue("src_user")
	srcPassword := c.FormValue("src_password")
	srcFolder := c.FormValue("src_folder")
	protocol := c.FormValue("protocol")

	// Parse Ints
	pollTime, _ := strconv.Atoi(c.FormValue("poll_time"))
	if pollTime == 0 {
		pollTime = 10
	}
	srcPort, _ := strconv.Atoi(c.FormValue("src_port"))

	// Parse Bools
	fetchall := c.FormValue("fetchall") == "true"
	keep := c.FormValue("keep") == "true"
	usessl := c.FormValue("usessl") == "true"
	sslcertck := c.FormValue("sslcertck") == "true"
	active := c.FormValue("active") == "true"

	// Basic validation
	if mailbox == "" {
		return renderFetchmailFormWithError(c, h, "O campo 'Conta' é obrigatório.")
	}

	// For domain, we extract it from the mailbox (assuming mailbox is an email address user@domain.com)
	var domainStr *string
	username := middleware.GetUsername(c, middleware.SessionName) // Admin username

	// Create Fetchmail object
	newFetchmail := models.Fetchmail{
		Mailbox:     mailbox,
		SrcServer:   srcServer,
		SrcUser:     srcUser,
		SrcPassword: srcPassword,
		SrcFolder:   srcFolder,
		PollTime:    pollTime,
		Fetchall:    fetchall,
		Keep:        keep,
		UseSSL:      usessl,
		SSLCertCk:   sslcertck,
		Active:      active,
		SrcPort:     srcPort,
		Date:        time.Now(),
		Created:     time.Now(),
		Modified:    time.Now(),
	}

	if srcAuth != "" {
		newFetchmail.SrcAuth = &srcAuth
	}
	if protocol != "" {
		newFetchmail.Protocol = &protocol
	}

	// Try extracting domain from mailbox
	if mailbox != "" {
		// basic extraction
		for i, char := range mailbox {
			if char == '@' {
				domain := mailbox[i+1:]
				domainStr = &domain
				break
			}
		}
	}
	if domainStr != nil {
		newFetchmail.Domain = domainStr
	}

	// Save to database
	if err := h.DB.Create(&newFetchmail).Error; err != nil {
		slog.Error("Failed to create fetchmail entry", "error", err, "username", username)
		return renderFetchmailFormWithError(c, h, "Falha ao salvar registro no banco de dados. Tente novamente.")
	}

	// Redirect or render success. Redirecting to same page for now (form reset) or list
	// Assuming there might be a list page in future, redirecting back to form with success message is also fine.
	// For simple implementation, redirect to dashboard or mailboxes, taking user back to form for now.
	return c.Redirect(http.StatusFound, "/fetchmail/add")
}

// Helper to re-render form with error state
func renderFetchmailFormWithError(c *echo.Context, h *Handler, errorMsg string) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	mailboxes, isSuper, _ := utils.GetAllMailboxes(h.DB, username, middleware.GetIsSuperAdmin(c), "")

	pollTime, _ := strconv.Atoi(c.FormValue("poll_time"))
	srcPort, _ := strconv.Atoi(c.FormValue("src_port"))

	renderData := map[string]interface{}{
		"Error":        errorMsg,
		"Mailboxes":    mailboxes,
		"Mailbox":      c.FormValue("mailbox"),
		"SrcServer":    c.FormValue("src_server"),
		"SrcAuth":      c.FormValue("src_auth"),
		"SrcUser":      c.FormValue("src_user"),
		"SrcPassword":  c.FormValue("src_password"),
		"SrcFolder":    c.FormValue("src_folder"),
		"PollTime":     pollTime,
		"Fetchall":     c.FormValue("fetchall") == "true",
		"Keep":         c.FormValue("keep") == "true",
		"Protocol":     c.FormValue("protocol"),
		"UseSSL":       c.FormValue("usessl") == "true",
		"SSLCertCk":    c.FormValue("sslcertck") == "true",
		"Active":       c.FormValue("active") == "true",
		"SrcPort":      srcPort,
		"SessionUser":  username,
		"IsSuperAdmin": isSuper,
	}

	return c.Render(http.StatusBadRequest, "add_fetchmail.html", renderData)
}
