package handlers

import (
	"net/http"

	"go-postfixadmin/internal/models"

	"github.com/labstack/echo/v5"
)

// ListMailboxes lista mailboxes com filtro opcional por domínio
func (h *Handler) ListMailboxes(c *echo.Context) error {
	var mailboxes []models.Mailbox
	domainFilter := c.QueryParam("domain") // Query parameter opcional

	if h.DB != nil {
		query := h.DB.Order("modified DESC")

		// Aplicar filtro se domínio fornecido
		if domainFilter != "" {
			query = query.Where("domain = ?", domainFilter)
		}

		query.Find(&mailboxes)
	}

	return c.Render(http.StatusOK, "mailboxes.html", map[string]interface{}{
		"Mailboxes":    mailboxes,
		"DomainFilter": domainFilter, // Para exibir no template
	})
}
