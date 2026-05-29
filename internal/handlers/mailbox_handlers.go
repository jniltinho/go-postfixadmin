package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// ListMailboxes lista mailboxes com filtro opcional por domínio
func (h *Handler) ListMailboxes(c *echo.Context) error {
	domainFilter := c.QueryParam("domain")
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	SessionUser := middleware.GetUsername(c, middleware.SessionName)

	var mailboxes []models.Mailbox

	if h.DB != nil {
		var err error
		mailboxes, _, err = repositories.GetAllMailboxes(h.DB, SessionUser, isSuperAdmin, domainFilter)
		if err != nil {
			if err.Error() == "access denied to this domain" {
				return c.Render(http.StatusForbidden, "mailboxes/mailboxes.html", map[string]interface{}{
					"Error": "Access denied to this domain",
				})
			}
			return c.Render(http.StatusInternalServerError, "mailboxes/mailboxes.html", map[string]interface{}{
				"Error": "Failed to fetch mailboxes: " + err.Error(),
			})
		}
	}

	var domains []models.Domain
	if h.DB != nil {
		domains, _, _ = repositories.GetActiveDomains(h.DB, SessionUser, isSuperAdmin)
	}

	domainLimitReached := false
	if h.DB != nil && domainFilter != "" {
		if domainRecord, err := repositories.GetDomainByName(h.DB, domainFilter); err == nil {
			if domainRecord.Mailboxes > 0 {
				if count, err := repositories.CountDomainMailboxes(h.DB, domainFilter); err == nil {
					domainLimitReached = count >= int64(domainRecord.Mailboxes)
				}
			}
		}
	}

	return c.Render(http.StatusOK, "mailboxes/mailboxes.html", map[string]interface{}{
		"Mailboxes":          mailboxes,
		"Domains":            domains,
		"DomainFilter":       domainFilter,
		"IsSuperAdmin":       isSuperAdmin,
		"SessionUser":        SessionUser,
		"QuotaMultiplier":    float64(utils.GetQuotaMultiplier()),
		"DomainLimitReached": domainLimitReached,
		"Message":            GetFlash(c, "message"),
		"Error":              GetFlash(c, "error"),
	})
}

// AddMailboxAPI cria um mailbox via AJAX e retorna JSON
func (h *Handler) AddMailboxAPI(c *echo.Context) error {
	localPart := strings.ToLower(strings.TrimSpace(c.FormValue("local_part")))
	domain := strings.TrimSpace(c.FormValue("domain"))
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	SessionUser := middleware.GetUsername(c, middleware.SessionName)

	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Permission check failed"})
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
			return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied to this domain"})
		}
	}

	name := strings.TrimSpace(c.FormValue("name"))
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	smtpActive := c.FormValue("smtp_active") == "true"
	sendWelcomeMail := c.FormValue("send_welcome_mail") == "true"
	emailOther := c.FormValue("email_other")

	quotaMultiplier := utils.GetQuotaMultiplier()
	quota := int64(0)
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed * quotaMultiplier
		}
	}

	if localPart == "" || domain == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Local part and domain are required"})
	}
	if len(localPart) < 4 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Username must be at least 4 characters"})
	}
	if validationErr := ValidatePassword(password); validationErr != "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": validationErr})
	}
	if password != passwordConfirm {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Passwords do not match"})
	}
	localPartRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !localPartRegex.MatchString(localPart) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid local part format. Use only letters, numbers, dots, hyphens, and underscores"})
	}

	username := fmt.Sprintf("%s@%s", localPart, domain)

	if _, err := repositories.GetMailboxByUsername(h.DB, username); err == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Mailbox already exists"})
	}

	domainRecord, err := repositories.GetDomainByName(h.DB, domain)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": T(c, "Domain not found")})
	}
	if domainRecord.Mailboxes > 0 {
		count, err := repositories.CountDomainMailboxes(h.DB, domain)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": T(c, "Failed to check mailbox limit")})
		}
		if count >= int64(domainRecord.Mailboxes) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": TData(c, "Mailbox_LimitReached", map[string]any{
				"Count": fmt.Sprintf("%d", count),
				"Limit": fmt.Sprintf("%d", domainRecord.Mailboxes),
			})})
		}
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to hash password: " + err.Error()})
	}

	now := time.Now()
	newMailbox := models.Mailbox{
		Username:       username,
		Password:       hashedPassword,
		Name:           name,
		Maildir:        fmt.Sprintf("%s/%s/", domain, localPart),
		Quota:          quota,
		LocalPart:      localPart,
		Domain:         domain,
		Created:        now,
		Modified:       now,
		Active:         active,
		EmailOther:     emailOther,
		SMTPActive:     smtpActive,
		TokenValidity:  now.Add(3 * time.Hour),
		PasswordExpiry: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := repositories.CreateMailbox(h.DB, newMailbox, SessionUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to create mailbox: " + err.Error()})
	}

	if sendWelcomeMail {
		lang := "en"
		if cookie, err := c.Cookie("lang"); err == nil {
			lang = cookie.Value
		}
		if err := utils.SendWelcomeEmail(SessionUser, username, lang); err != nil {
			fmt.Printf("Failed to send welcome email to %s: %v\n", username, err)
		}
	}

	SetFlash(c, "message", "Mailbox created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "domain": domain})
}

// GetMailboxAPI retorna os dados de um mailbox em JSON para popular o modal de edição
func (h *Handler) GetMailboxAPI(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))
	SessionUser := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Mailbox not found"})
	}

	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
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

	quotaMultiplier := utils.GetQuotaMultiplier()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"username":   mailbox.Username,
		"name":       mailbox.Name,
		"quota":      mailbox.Quota / quotaMultiplier,
		"active":     mailbox.Active,
		"smtpActive": mailbox.SMTPActive,
		"emailOther": mailbox.EmailOther,
		"domain":     mailbox.Domain,
	})
}

// EditMailboxAPI processa a edição de um mailbox via AJAX e retorna JSON
func (h *Handler) EditMailboxAPI(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))
	SessionUser := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Mailbox not found"})
	}

	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
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

	name := strings.TrimSpace(c.FormValue("name"))
	active := c.FormValue("active") == "true"
	smtpActive := c.FormValue("smtp_active") == "true"
	emailOther := c.FormValue("email_other")
	changePassword := c.FormValue("change_password") == "true"

	quotaMultiplier := utils.GetQuotaMultiplier()
	quota := int64(0)
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed * quotaMultiplier
		}
	}

	if changePassword {
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("password_confirm")
		if validationErr := ValidatePassword(password); validationErr != "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": validationErr})
		}
		if password != passwordConfirm {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Passwords do not match"})
		}
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to hash password: " + err.Error()})
		}
		mailbox.Password = hashedPassword
	}

	mailbox.Name = name
	mailbox.Quota = quota
	mailbox.Active = active
	mailbox.SMTPActive = smtpActive
	mailbox.EmailOther = emailOther
	mailbox.Modified = time.Now()
	mailbox.TokenValidity = time.Now().Add(3 * time.Hour)

	if err := repositories.SaveMailbox(h.DB, mailbox, "edit_mailbox", SessionUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to update mailbox: " + err.Error()})
	}

	SetFlash(c, "message", "Mailbox updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "domain": mailbox.Domain})
}

// DeleteMailbox remove um mailbox e o alias correspondente
func (h *Handler) DeleteMailbox(c *echo.Context) error {
	username, _ := url.PathUnescape(c.Param("username"))

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"success": false, "error": "Mailbox not found"})
	}

	SessionUser := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, SessionUser, isSuperAdmin)
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

	if err := repositories.DeleteMailbox(h.DB, mailbox, SessionUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to delete mailbox: " + err.Error()})
	}

	if viper.GetBool("server.cleanup_maildir") {
		baseDir := "/var/vmail"
		if cleanupErr := utils.CleanupOrphanedMaildir(h.DB, baseDir, mailbox.Domain, mailbox.LocalPart); cleanupErr != nil {
			fmt.Printf("Warning: Failed to clean up orphaned directory for %s: %v\n", username, cleanupErr)
		}
	}

	SetFlash(c, "message", "Mailbox deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": "Mailbox deleted successfully"})
}

// =====================================================
// API v1 Mailbox Handlers (PR 06+)
// JWT + domain scoping enforced via claims
// =====================================================

// ListMailboxesV1 returns mailboxes visible to the authenticated user
func (h *Handler) ListMailboxesV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	domainFilter := c.QueryParam("domain")

	mailboxes, _, err := repositories.GetAllMailboxes(h.DB, claims.Username, claims.Superadmin, domainFilter)
	if err != nil {
		if err.Error() == "access denied to this domain" {
			return dto.Forbidden(c, "access denied to this domain")
		}
		return dto.InternalError(c, "failed to fetch mailboxes")
	}

	var response []dto.MailboxResponse
	for _, m := range mailboxes {
		response = append(response, dto.MailboxResponse{
			Username:       m.Username,
			Domain:         m.Domain,
			Name:           m.Name,
			Quota:          m.Quota,
			Active:         m.Active,
			SMTPActive:     m.SMTPActive,
			EmailOther:     m.EmailOther,
			Created:        m.Created,
			Modified:       m.Modified,
			PasswordExpiry: m.PasswordExpiry,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateMailboxV1 handles POST /api/v1/mailboxes
func (h *Handler) CreateMailboxV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	var req dto.CreateMailboxRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	localPart := strings.ToLower(strings.TrimSpace(req.LocalPart))
	domain := strings.TrimSpace(req.Domain)

	if localPart == "" || domain == "" {
		return dto.ValidationError(c, "local_part and domain are required")
	}

	// Permission check using claims
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied to this domain")
		}
	}

	username := localPart + "@" + domain

	// Check if already exists
	if _, err := repositories.GetMailboxByUsername(h.DB, username); err == nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "mailbox already exists")
	}

	// Validate password if provided
	if req.Password != "" {
		if validationErr := ValidatePassword(req.Password); validationErr != "" {
			return dto.ValidationError(c, validationErr)
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return dto.InternalError(c, "failed to hash password")
	}

	now := time.Now()
	mailbox := models.Mailbox{
		Username:       username,
		Password:       hashedPassword,
		Name:           req.Name,
		Maildir:        domain + "/" + localPart + "/",
		Quota:          req.Quota,
		LocalPart:      localPart,
		Domain:         domain,
		Transport:      "",
		Created:        now,
		Modified:       now,
		Active:         req.Active,
		EmailOther:     req.EmailOther,
		SMTPActive:     req.SMTPActive,
		Token:          "",
		TokenValidity:  now.Add(3 * time.Hour),
		PasswordExpiry: now,
	}

	if err := repositories.CreateMailbox(h.DB, mailbox, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to create mailbox")
	}

	// Optional welcome email
	if req.SendWelcome {
		lang := "pt"
		if cookie, err := c.Cookie("lang"); err == nil {
			lang = cookie.Value
		}
		_ = utils.SendWelcomeEmail(claims.Username, username, lang)
	}

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, map[string]string{
		"username": username,
		"domain":   domain,
	})
}

// GetMailboxV1 handles GET /api/v1/mailboxes/:username
func (h *Handler) GetMailboxV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	username, _ := url.PathUnescape(c.Param("username"))

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return dto.NotFound(c, "mailbox not found")
	}

	// Scoping check
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == mailbox.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	quotaMultiplier := utils.GetQuotaMultiplier()

	resp := dto.MailboxResponse{
		Username:       mailbox.Username,
		Domain:         mailbox.Domain,
		Name:           mailbox.Name,
		Quota:          mailbox.Quota / quotaMultiplier,
		Active:         mailbox.Active,
		SMTPActive:     mailbox.SMTPActive,
		EmailOther:     mailbox.EmailOther,
		Created:        mailbox.Created,
		Modified:       mailbox.Modified,
		PasswordExpiry: mailbox.PasswordExpiry,
	}

	return dto.WriteSuccess(c, resp)
}

// UpdateMailboxV1 handles PUT /api/v1/mailboxes/:username
func (h *Handler) UpdateMailboxV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	username, _ := url.PathUnescape(c.Param("username"))

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return dto.NotFound(c, "mailbox not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == mailbox.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	var req dto.UpdateMailboxRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	quotaMultiplier := utils.GetQuotaMultiplier()

	if req.Name != nil {
		mailbox.Name = *req.Name
	}
	if req.Quota != nil {
		mailbox.Quota = *req.Quota * quotaMultiplier
	}
	if req.Active != nil {
		mailbox.Active = *req.Active
	}
	if req.SMTPActive != nil {
		mailbox.SMTPActive = *req.SMTPActive
	}
	if req.EmailOther != nil {
		mailbox.EmailOther = *req.EmailOther
	}

	if req.ChangePassword != nil && *req.ChangePassword {
		if req.Password == nil || req.PasswordConfirm == nil || *req.Password != *req.PasswordConfirm {
			return dto.ValidationError(c, "passwords do not match")
		}
		if validationErr := ValidatePassword(*req.Password); validationErr != "" {
			return dto.ValidationError(c, validationErr)
		}
		hashed, err := utils.HashPassword(*req.Password)
		if err != nil {
			return dto.InternalError(c, "failed to hash password")
		}
		mailbox.Password = hashed
	}

	mailbox.Modified = time.Now()
	mailbox.TokenValidity = time.Now().Add(3 * time.Hour)

	if err := repositories.SaveMailbox(h.DB, mailbox, "edit_mailbox", claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to update mailbox")
	}

	return dto.WriteSuccess(c, map[string]bool{"updated": true})
}

// DeleteMailboxV1 handles DELETE /api/v1/mailboxes/:username
func (h *Handler) DeleteMailboxV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	username, _ := url.PathUnescape(c.Param("username"))

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return dto.NotFound(c, "mailbox not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == mailbox.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	if err := repositories.DeleteMailbox(h.DB, mailbox, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to delete mailbox")
	}

	if viper.GetBool("server.cleanup_maildir") {
		baseDir := "/var/vmail"
		_ = utils.CleanupOrphanedMaildir(h.DB, baseDir, mailbox.Domain, mailbox.LocalPart)
	}

	return dto.WriteSuccess(c, map[string]string{"deleted": username})
}
