package handlers

import (
	"fmt"
	"net/http"
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

// DomainDisplay representa um domínio com contadores de aliases e mailboxes
type DomainDisplay struct {
	models.Domain
	AliasCount   int64
	MailboxCount int64
}

// ListDomains lista todos os domínios com contadores de aliases e mailboxes
func (h *Handler) ListDomains(c *echo.Context) error {
	var domains []models.Domain
	var displayDomains []DomainDisplay
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	allowedDomains, _, err := utils.GetAllowedDomains(h.DB, username, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "domains/domains.html", map[string]interface{}{
			"Error": "Failed to check permissions: " + err.Error(),
		})
	}

	if h.DB != nil {
		query := h.DB.Where("domain != ?", "ALL")
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0") // No domains allowed
			} else {
				query = query.Where("domain IN ?", allowedDomains)
			}
		}
		query.Find(&domains)

		for _, d := range domains {
			var aliasCount int64
			var mailboxCount int64

			// Count aliases excluding those that are mailboxes
			h.DB.Model(&models.Alias{}).
				Where("domain = ?", d.Domain).
				Where("address NOT IN (?)", h.DB.Table("mailbox").Select("username")).
				Count(&aliasCount)

			h.DB.Model(&models.Mailbox{}).Where("domain = ?", d.Domain).Count(&mailboxCount)

			displayDomains = append(displayDomains, DomainDisplay{
				Domain:       d,
				AliasCount:   aliasCount,
				MailboxCount: mailboxCount,
			})
		}
	}
	return c.Render(http.StatusOK, "domains/domains.html", map[string]interface{}{
		"Domains":      displayDomains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  username,
		"Message":      middleware.GetFlash(c, "message"),
		"Error":        middleware.GetFlash(c, "error"),
	})
}

// AddDomainAPI processa a criação de um novo domínio via JSON API
func (h *Handler) AddDomainAPI(c *echo.Context) error {
	// Security: Only Superadmins can add domains
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	if !isSuperAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied: Only Superadmins can create domains"})
	}

	// Parse form data
	domainName := strings.TrimSpace(c.FormValue("domain"))
	description := c.FormValue("description")
	active := c.FormValue("active") == "true"
	backupMX := c.FormValue("backupmx") == "true"

	// Parse numeric fields with defaults
	aliases := 10
	if val := c.FormValue("aliases"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			aliases = parsed
		}
	}

	mailboxes := 10
	if val := c.FormValue("mailboxes"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			mailboxes = parsed
		}
	}

	quota := int64(2048) // Default
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed
		}
	}

	var passwordExpiry *int
	if val := c.FormValue("password_expiry"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			passwordExpiry = &parsed
		}
	}

	// Validation: domain is required
	if domainName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Domain name is required"})
	}

	// Validation: basic DNS format
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?(\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?)+$`)
	if !domainRegex.MatchString(domainName) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid domain format. Please enter a valid domain name (e.g., example.com)"})
	}

	// Check if domain already exists
	var existingDomain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&existingDomain).Error; err == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Domain already exists"})
	}

	// Create new domain
	now := time.Now()
	newDomain := models.Domain{
		Domain:         domainName,
		Description:    description,
		Aliases:        aliases,
		Mailboxes:      mailboxes,
		MaxQuota:       0,
		Quota:          quota,
		Transport:      "",
		BackupMX:       backupMX,
		Created:        now,
		Modified:       now,
		Active:         active,
		PasswordExpiry: passwordExpiry,
	}

	if err := h.DB.Create(&newDomain).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to create domain: " + err.Error()})
	}

	// Log Action
	if err := utils.LogAction(h.DB, username, c.RealIP(), domainName, "create_domain", domainName); err != nil {
		fmt.Printf("Failed to log create_domain: %v\n", err)
	}

	middleware.SetFlash(c, "message", "Domain created successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetDomainAPI retorna os dados de um domínio para popular o modal de edição
func (h *Handler) GetDomainAPI(c *echo.Context) error {
	// Security: Only Superadmins can edit domains
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	if !isSuperAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	domainName := c.Param("domain")

	var domain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&domain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Domain not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"domain":  domain,
	})
}

// EditDomainAPI processa a edição de um domínio existente via JSON API
func (h *Handler) EditDomainAPI(c *echo.Context) error {
	// Security: Only Superadmins can edit domains
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	if !isSuperAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	domainName := c.Param("domain")

	// Find existing domain
	var domain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&domain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Domain not found"})
	}

	// Parse form data
	description := c.FormValue("description")
	active := c.FormValue("active") == "true"
	backupMX := c.FormValue("backupmx") == "true"

	aliases := 10
	if val := c.FormValue("aliases"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			aliases = parsed
		}
	}

	mailboxes := 10
	if val := c.FormValue("mailboxes"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			mailboxes = parsed
		}
	}

	quota := domain.Quota
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed
		}
	}

	var passwordExpiry *int
	if val := c.FormValue("password_expiry"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			passwordExpiry = &parsed
		}
	}

	// Check if active state is changing
	activeChanged := domain.Active != active

	// Update domain fields
	domain.Description = description
	domain.Aliases = aliases
	domain.Mailboxes = mailboxes
	domain.Quota = quota
	domain.BackupMX = backupMX
	domain.Modified = time.Now()
	domain.Active = active
	domain.PasswordExpiry = passwordExpiry

	// Use transaction to ensure atomicity
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if activeChanged {
			if err := tx.Model(&models.Mailbox{}).Where("domain = ?", domain.Domain).Update("active", active).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Alias{}).Where("domain = ?", domain.Domain).Update("active", active).Error; err != nil {
				return err
			}
		}

		if err := tx.Save(&domain).Error; err != nil {
			return err
		}

		if err := utils.LogAction(tx, username, c.RealIP(), domainName, "edit_domain", domainName); err != nil {
			fmt.Printf("Failed to log edit_domain: %v\n", err)
		}
		return nil
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to update domain: " + err.Error()})
	}

	middleware.SetFlash(c, "message", "Domain updated successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteDomain remove um domínio e todos os dados associados (aliases e mailboxes)
func (h *Handler) DeleteDomain(c *echo.Context) error {
	// Security: Only Superadmins can delete domains
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	if !isSuperAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied: Only Superadmins can delete domains"})
	}

	domainName := c.Param("domain")

	// Check if domain exists
	var domain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&domain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Domain not found",
		})
	}

	// Use utility function to delete domain and all associated data
	if err := utils.DeleteDomain(h.DB, domainName, username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to delete domain: " + err.Error(),
		})
	}

	middleware.SetFlash(c, "message", "Domain deleted successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Domain deleted successfully",
	})
}
