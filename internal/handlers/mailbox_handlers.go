package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// ListMailboxes lista mailboxes com filtro opcional por domínio
func (h *Handler) ListMailboxes(c *echo.Context) error {
	domainFilter := c.QueryParam("domain") // Query parameter opcional
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	SessionUser := middleware.GetUsername(c)

	var mailboxes []models.Mailbox

	if h.DB != nil {
		var err error
		mailboxes, _, err = utils.GetAllMailboxes(h.DB, SessionUser, isSuperAdmin, domainFilter)
		if err != nil {
			if err.Error() == "access denied to this domain" {
				return c.Render(http.StatusForbidden, "mailboxes.html", map[string]interface{}{
					"Error": "Access denied to this domain",
				})
			}
			return c.Render(http.StatusInternalServerError, "mailboxes.html", map[string]interface{}{
				"Error": "Failed to fetch mailboxes: " + err.Error(),
			})
		}
	}

	// Fetch domains for the filter dropdown
	var domains []models.Domain
	if h.DB != nil {
		domains, _, _ = utils.GetActiveDomains(h.DB, SessionUser, isSuperAdmin)
	}

	quotaMultiplier := utils.GetQuotaMultiplier()

	return c.Render(http.StatusOK, "mailboxes.html", map[string]interface{}{
		"Mailboxes":       mailboxes,
		"Domains":         domains,
		"DomainFilter":    domainFilter, // Para exibir no template
		"IsSuperAdmin":    isSuperAdmin,
		"SessionUser":     SessionUser,
		"QuotaMultiplier": float64(quotaMultiplier),
	})
}

// AddMailboxForm exibe o formulário de adicionar mailbox
func (h *Handler) AddMailboxForm(c *echo.Context) error {
	var domains []models.Domain
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	SessionUser := middleware.GetUsername(c)

	if h.DB != nil {
		// Security: Filter domains
		var err error
		domains, _, err = utils.GetActiveDomains(h.DB, SessionUser, isSuperAdmin)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "mailboxes.html", map[string]interface{}{"Error": "Permission check failed"})
		}
	}

	return c.Render(http.StatusOK, "add_mailbox.html", map[string]interface{}{
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  SessionUser,
	})
}

// AddMailbox processa a criação de um novo mailbox
func (h *Handler) AddMailbox(c *echo.Context) error {
	// Parse form data
	localPart := strings.ToLower(strings.TrimSpace(c.FormValue("local_part")))
	domain := strings.TrimSpace(c.FormValue("domain"))
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	SessionUser := middleware.GetUsername(c)

	// Security: Validate domain access
	allowedDomains, _, err := utils.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "add_mailbox.html", map[string]interface{}{"Error": "Permission check failed"})
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.Render(http.StatusForbidden, "add_mailbox.html", map[string]interface{}{"Error": "Access denied to this domain"})
		}
	}
	name := strings.TrimSpace(c.FormValue("name"))
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	smtpActive := c.FormValue("smtp_active") == "true"

	emailOther := c.FormValue("email_other")

	// Parse quota (MB * quota_multiplier)
	quotaMultiplier := utils.GetQuotaMultiplier()

	quota := int64(0)
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed * quotaMultiplier
		}
	}

	// Load domains for re-rendering on error
	var domains []models.Domain
	if h.DB != nil {
		domains, _, _ = utils.GetActiveDomains(h.DB, SessionUser, isSuperAdmin)
	}

	// Validation: required fields
	if localPart == "" || domain == "" {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":        "Local part and domain are required",
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
			"SessionUser":  SessionUser,
		})
	}

	// Validation: local part minimum 4 characters
	if len(localPart) < 4 {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":        "Username must be at least 4 characters",
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
		})
	}

	// Validation: password minimum 8 characters
	if len(password) < 8 {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":        "Password must be at least 8 characters",
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
		})
	}

	// Validation: password confirmation match
	if password != passwordConfirm {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":        "Password and confirmation do not match",
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
		})
	}

	// Validation: local part format (basic email local part validation)
	localPartRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !localPartRegex.MatchString(localPart) {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":        "Invalid local part format. Use only letters, numbers, dots, hyphens, and underscores",
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
		})
	}

	// Construct full email address
	username := fmt.Sprintf("%s@%s", localPart, domain)

	// Check if mailbox already exists
	var existingMailbox models.Mailbox
	if err := h.DB.Where("username = ?", username).First(&existingMailbox).Error; err == nil {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":        "Mailbox already exists",
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "add_mailbox.html", map[string]interface{}{
			"Error":        "Failed to hash password: " + err.Error(),
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
		})
	}

	// Generate maildir
	maildir := generateMaildir(domain, localPart)

	// Create mailbox and alias in a transaction
	now := time.Now()
	defaultDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Create mailbox
		newMailbox := models.Mailbox{
			Username:       username,
			Password:       hashedPassword,
			Name:           name,
			Maildir:        maildir,
			Quota:          quota,
			LocalPart:      localPart,
			Domain:         domain,
			Created:        now,
			Modified:       now,
			Active:         active,
			EmailOther:     emailOther,
			SMTPActive:     smtpActive,
			TokenValidity:  time.Now().Add(3 * time.Hour),
			PasswordExpiry: defaultDate,
		}

		if err := tx.Create(&newMailbox).Error; err != nil {
			return err
		}

		// Create corresponding alias
		if err := createMailboxAlias(tx, username, domain); err != nil {
			return err
		}

		// Log Action inside transaction
		if err := utils.LogAction(tx, SessionUser, c.RealIP(), domain, "create_mailbox", username); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Render(http.StatusInternalServerError, "add_mailbox.html", map[string]interface{}{
			"Error":        "Failed to create mailbox: " + err.Error(),
			"Domains":      domains,
			"LocalPart":    localPart,
			"Domain":       domain,
			"Name":         name,
			"Active":       active,
			"SMTPActive":   smtpActive,
			"EmailOther":   emailOther,
			"Quota":        quota / quotaMultiplier,
			"IsSuperAdmin": isSuperAdmin,
			"SessionUser":  SessionUser,
		})
	}

	// Redirect to mailboxes list filtered by domain
	return c.Redirect(http.StatusFound, fmt.Sprintf("/mailboxes?domain=%s", domain))
}

// EditMailboxForm exibe o formulário de edição de mailbox
func (h *Handler) EditMailboxForm(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))

	var mailbox models.Mailbox
	if err := h.DB.Where("username = ?", username).First(&mailbox).Error; err != nil {
		return c.Render(http.StatusNotFound, "edit_mailbox.html", map[string]interface{}{
			"Error": "Mailbox not found",
		})
	}

	// Security: Check permission
	SessionUser := middleware.GetUsername(c)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	allowedDomains, _, err := utils.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "edit_mailbox.html", map[string]interface{}{"Error": "Permission check failed"})
	}
	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == mailbox.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.Render(http.StatusForbidden, "mailboxes.html", map[string]interface{}{"Error": "Access denied"})
		}
	}

	quotaMultiplier := utils.GetQuotaMultiplier()

	return c.Render(http.StatusOK, "edit_mailbox.html", map[string]interface{}{
		"Mailbox":      mailbox,
		"QuotaMB":      mailbox.Quota / quotaMultiplier,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  SessionUser,
	})
}

// EditMailbox processa a edição de um mailbox existente
func (h *Handler) EditMailbox(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))
	SessionUser := middleware.GetUsername(c)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	// Find existing mailbox
	var mailbox models.Mailbox
	if err := h.DB.Where("username = ?", username).First(&mailbox).Error; err != nil {
		return c.Render(http.StatusNotFound, "edit_mailbox.html", map[string]interface{}{
			"Error": "Mailbox not found",
		})
	}

	// Security: Check permission
	allowedDomains, _, err := utils.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "edit_mailbox.html", map[string]interface{}{"Error": "Permission check failed"})
	}
	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == mailbox.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.Render(http.StatusForbidden, "mailboxes.html", map[string]interface{}{"Error": "Access denied"})
		}
	}

	// Parse form data
	name := strings.TrimSpace(c.FormValue("name"))
	active := c.FormValue("active") == "true"
	smtpActive := c.FormValue("smtp_active") == "true"

	emailOther := c.FormValue("email_other")
	changePassword := c.FormValue("change_password") == "true"

	// Parse quota (MB * quota_multiplier)
	quotaMultiplier := utils.GetQuotaMultiplier()

	quota := int64(0)
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed * quotaMultiplier
		}
	}

	// Handle optional password change
	if changePassword {
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("password_confirm")

		// Validation: password minimum 8 characters
		if len(password) < 8 {
			return c.Render(http.StatusBadRequest, "edit_mailbox.html", map[string]interface{}{
				"Error":        "Password must be at least 8 characters",
				"Mailbox":      mailbox,
				"IsSuperAdmin": isSuperAdmin,
				"SessionUser":  SessionUser,
			})
		}

		// Validation: password confirmation match
		if password != passwordConfirm {
			return c.Render(http.StatusBadRequest, "edit_mailbox.html", map[string]interface{}{
				"Error":        "Password and confirmation do not match",
				"Mailbox":      mailbox,
				"IsSuperAdmin": isSuperAdmin,
				"SessionUser":  SessionUser,
			})
		}

		// Hash new password
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "edit_mailbox.html", map[string]interface{}{
				"Error":        "Failed to hash password: " + err.Error(),
				"Mailbox":      mailbox,
				"IsSuperAdmin": isSuperAdmin,
				"SessionUser":  SessionUser,
			})
		}

		mailbox.Password = hashedPassword
	}

	// Update mailbox fields
	mailbox.Name = name
	mailbox.Quota = quota
	mailbox.Active = active
	mailbox.SMTPActive = smtpActive
	mailbox.EmailOther = emailOther
	mailbox.Modified = time.Now()
	mailbox.TokenValidity = time.Now().Add(3 * time.Hour)

	if err := h.DB.Save(&mailbox).Error; err != nil {
		return c.Render(http.StatusInternalServerError, "edit_mailbox.html", map[string]interface{}{
			"Error":        "Failed to update mailbox: " + err.Error(),
			"Mailbox":      mailbox,
			"IsSuperAdmin": isSuperAdmin,
			"SessionUser":  SessionUser,
		})
	}

	// Log Action
	if err := utils.LogAction(h.DB, SessionUser, c.RealIP(), mailbox.Domain, "edit_mailbox", username); err != nil {
		fmt.Printf("Failed to log edit_mailbox: %v\n", err)
	}

	// Redirect to mailboxes list filtered by domain
	return c.Redirect(http.StatusFound, fmt.Sprintf("/mailboxes?domain=%s", mailbox.Domain))
}

// DeleteMailbox remove um mailbox e o alias correspondente
func (h *Handler) DeleteMailbox(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))

	// Check if mailbox exists
	var mailbox models.Mailbox
	if err := h.DB.Where("username = ?", username).First(&mailbox).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Mailbox not found",
		})
	}

	// Security: Check permission
	SessionUser := middleware.GetUsername(c)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	allowedDomains, _, err := utils.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Permission check failed"})
	}
	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == mailbox.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
		}
	}

	// Use transaction to ensure atomicity
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Delete corresponding alias
		if err := tx.Where("address = ?", username).Delete(&models.Alias{}).Error; err != nil {
			return err
		}

		// Delete Vacation
		if err := tx.Where("username = ?", username).Delete(&models.Vacation{}).Error; err != nil {
			return err
		}

		// Delete the mailbox
		if err := tx.Where("username = ?", username).Delete(&models.Mailbox{}).Error; err != nil {
			return err
		}

		// Log Action inside transaction
		if err := utils.LogAction(tx, SessionUser, c.RealIP(), mailbox.Domain, "delete_mailbox", username); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to delete mailbox: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Mailbox deleted successfully",
	})
}

// GeneratePassword API endpoint para gerar senha complexa
func (h *Handler) GeneratePassword(c *echo.Context) error {
	password := utils.GenerateComplexPassword()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"password": password,
	})
}

// Helper Functions

// generateMaildir gera o caminho do maildir no formato domain/localpart/
func generateMaildir(domain, localPart string) string {
	return fmt.Sprintf("%s/%s/", domain, localPart)
}

// createMailboxAlias cria um alias automático para o mailbox
func createMailboxAlias(tx *gorm.DB, username, domain string) error {
	now := time.Now()
	alias := models.Alias{
		Address:  username,
		Goto:     username,
		Domain:   domain,
		Created:  now,
		Modified: now,
		Active:   true,
	}
	return tx.Create(&alias).Error
}
