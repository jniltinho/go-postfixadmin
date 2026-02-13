package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// ListMailboxes lista mailboxes com filtro opcional por domínio
func (h *Handler) ListMailboxes(c *echo.Context) error {
	var mailboxes []models.Mailbox
	domainFilter := c.QueryParam("domain") // Query parameter opcional

	if h.DB != nil {
		query := h.DB.Order("username ASC")

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

// AddMailboxForm exibe o formulário de adicionar mailbox
func (h *Handler) AddMailboxForm(c *echo.Context) error {
	var domains []models.Domain

	if h.DB != nil {
		h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC").Find(&domains)
	}

	return c.Render(http.StatusOK, "add_mailbox.html", map[string]interface{}{
		"Domains": domains,
	})
}

// AddMailbox processa a criação de um novo mailbox
func (h *Handler) AddMailbox(c *echo.Context) error {
	// Parse form data
	localPart := strings.ToLower(strings.TrimSpace(c.FormValue("local_part")))
	domain := strings.TrimSpace(c.FormValue("domain"))
	name := strings.TrimSpace(c.FormValue("name"))
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	smtpActive := c.FormValue("smtp_active") == "true"

	emailOther := c.FormValue("email_other")

	// Parse quota (MB * quota_multiplier)
	const quotaMultiplier int64 = 1024000
	quota := int64(0)
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed * quotaMultiplier
		}
	}

	// Load domains for re-rendering on error
	var domains []models.Domain
	if h.DB != nil {
		h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC").Find(&domains)
	}

	// Validation: required fields
	if localPart == "" || domain == "" {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":      "Local part and domain are required",
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
		})
	}

	// Validation: local part minimum 4 characters
	if len(localPart) < 4 {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":      "Username must be at least 4 characters",
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
		})
	}

	// Validation: password minimum 8 characters
	if len(password) < 8 {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":      "Password must be at least 8 characters",
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
		})
	}

	// Validation: password confirmation match
	if password != passwordConfirm {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":      "Password and confirmation do not match",
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
		})
	}

	// Validation: local part format (basic email local part validation)
	localPartRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !localPartRegex.MatchString(localPart) {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":      "Invalid local part format. Use only letters, numbers, dots, hyphens, and underscores",
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
		})
	}

	// Construct full email address
	username := fmt.Sprintf("%s@%s", localPart, domain)

	// Check if mailbox already exists
	var existingMailbox models.Mailbox
	if err := h.DB.Where("username = ?", username).First(&existingMailbox).Error; err == nil {
		return c.Render(http.StatusBadRequest, "add_mailbox.html", map[string]interface{}{
			"Error":      "Mailbox already exists",
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "add_mailbox.html", map[string]interface{}{
			"Error":      "Failed to hash password: " + err.Error(),
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
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
			TokenValidity:  defaultDate,
			PasswordExpiry: defaultDate,
		}

		if err := tx.Create(&newMailbox).Error; err != nil {
			return err
		}

		// Create corresponding alias
		if err := createMailboxAlias(tx, username, domain); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Render(http.StatusInternalServerError, "add_mailbox.html", map[string]interface{}{
			"Error":      "Failed to create mailbox: " + err.Error(),
			"Domains":    domains,
			"LocalPart":  localPart,
			"Domain":     domain,
			"Name":       name,
			"Active":     active,
			"SMTPActive": smtpActive,
			"EmailOther": emailOther,
			"Quota":      quota / quotaMultiplier,
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

	return c.Render(http.StatusOK, "edit_mailbox.html", map[string]interface{}{
		"Mailbox": mailbox,
		"QuotaMB": mailbox.Quota / 1024000,
	})
}

// EditMailbox processa a edição de um mailbox existente
func (h *Handler) EditMailbox(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))

	// Find existing mailbox
	var mailbox models.Mailbox
	if err := h.DB.Where("username = ?", username).First(&mailbox).Error; err != nil {
		return c.Render(http.StatusNotFound, "edit_mailbox.html", map[string]interface{}{
			"Error": "Mailbox not found",
		})
	}

	// Parse form data
	name := strings.TrimSpace(c.FormValue("name"))
	active := c.FormValue("active") == "true"
	smtpActive := c.FormValue("smtp_active") == "true"

	emailOther := c.FormValue("email_other")
	changePassword := c.FormValue("change_password") == "true"

	// Parse quota (MB * quota_multiplier)
	const quotaMultiplier int64 = 1024000
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
				"Error":   "Password must be at least 8 characters",
				"Mailbox": mailbox,
			})
		}

		// Validation: password confirmation match
		if password != passwordConfirm {
			return c.Render(http.StatusBadRequest, "edit_mailbox.html", map[string]interface{}{
				"Error":   "Password and confirmation do not match",
				"Mailbox": mailbox,
			})
		}

		// Hash new password
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "edit_mailbox.html", map[string]interface{}{
				"Error":   "Failed to hash password: " + err.Error(),
				"Mailbox": mailbox,
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

	if err := h.DB.Save(&mailbox).Error; err != nil {
		return c.Render(http.StatusInternalServerError, "edit_mailbox.html", map[string]interface{}{
			"Error":   "Failed to update mailbox: " + err.Error(),
			"Mailbox": mailbox,
		})
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

	// Use transaction to ensure atomicity
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Delete the mailbox
		if err := tx.Where("username = ?", username).Delete(&models.Mailbox{}).Error; err != nil {
			return err
		}

		// Delete corresponding alias
		if err := tx.Where("address = ?", username).Delete(&models.Alias{}).Error; err != nil {
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
	password := generateComplexPassword()
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

// generateComplexPassword gera uma senha complexa aleatória
func generateComplexPassword() string {
	const (
		length      = 16
		upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars  = "abcdefghijklmnopqrstuvwxyz"
		digitChars  = "0123456789"
		symbolChars = "!@#$%^&*()-_=+[]{}|;:,.<>?"
		allChars    = upperChars + lowerChars + digitChars + symbolChars
	)

	password := make([]byte, length)

	// Ensure at least one character from each category
	password[0] = upperChars[randomInt(len(upperChars))]
	password[1] = lowerChars[randomInt(len(lowerChars))]
	password[2] = digitChars[randomInt(len(digitChars))]
	password[3] = symbolChars[randomInt(len(symbolChars))]

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = allChars[randomInt(len(allChars))]
	}

	// Shuffle the password to avoid predictable patterns
	for i := range password {
		j := randomInt(len(password))
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

// randomInt gera um número inteiro aleatório entre 0 e max-1
func randomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback to timestamp-based randomness if crypto/rand fails
		return int(time.Now().UnixNano()) % max
	}
	return int(n.Int64())
}
