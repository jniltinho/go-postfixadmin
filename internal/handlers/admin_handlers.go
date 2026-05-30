package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

type AdminData struct {
	models.Admin
	DomainCount string
}

// ListAdmins displays the list of administrators
func (h *Handler) ListAdmins(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, username)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "dashboard.html", map[string]interface{}{"Error": "Permission check failed"})
	}

	admins, err := repositories.GetAllAdmins(h.DB, username, isSuper)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "admins/admins.html", map[string]interface{}{
			"error": "Failed to fetch administrators",
		})
	}

	var adminList []AdminData
	for _, admin := range admins {
		var domainCountStr string
		if admin.Superadmin {
			domainCountStr = "ALL"
		} else {
			count, _ := repositories.CountAdminDomains(h.DB, admin.Username)
			domainCountStr = fmt.Sprintf("%d", count)
		}
		adminList = append(adminList, AdminData{Admin: admin, DomainCount: domainCountStr})
	}

	var domains []models.Domain
	if isSuper {
		domains, _, _ = repositories.GetActiveDomains(h.DB, username, true)
	}

	return c.Render(http.StatusOK, "admins/admins.html", map[string]interface{}{
		"Admins":       adminList,
		"Domains":      domains,
		"IsSuperAdmin": isSuper,
		"SessionUser":  username,
		"Message":      GetFlash(c, "message"),
		"Error":        GetFlash(c, "error"),
	})
}

// AddAdminAPI processes the creation of a new administrator via JSON API
func (h *Handler) AddAdminAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	domains := c.Request().Form["domains"]

	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username is required"})
	}
	if validationErr := ValidatePassword(password); validationErr != "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
	}
	if password != passwordConfirm {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
	}

	if _, err := repositories.GetAdminByUsername(h.DB, username); err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Admin already exists"})
	}

	crypted, err := utils.HashPassword(password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to hash password"})
	}

	newAdmin := models.Admin{
		Username:      username,
		Password:      crypted,
		Created:       time.Now(),
		Modified:      time.Now(),
		Active:        active,
		Superadmin:    superadmin,
		TokenValidity: time.Now().Add(3 * time.Hour),
	}

	if err := repositories.CreateAdmin(h.DB, newAdmin, domains, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create administrator"})
	}

	SetFlash(c, "message", "Admin created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAdmin handles the deletion of an administrator
func (h *Handler) DeleteAdmin(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	username, _ := url.PathUnescape(c.Param("username"))
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username is required"})
	}

	if username == loggedInUser {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "You cannot delete your own administrator account"})
	}

	if err := repositories.DeleteAdmin(h.DB, username, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to delete administrator"})
	}

	SetFlash(c, "message", "Admin deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetAdminAPI fetches a single administrator details for the edit modal
func (h *Handler) GetAdminAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Permission check failed"})
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if !isSuper && loggedInUser != targetUsername {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}
	if targetUsername == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Username required"})
	}

	admin, err := repositories.GetAdminByUsername(h.DB, targetUsername)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Admin not found"})
	}

	allDomains, _, _ := repositories.GetActiveDomains(h.DB, loggedInUser, true)
	domainAdmins, _ := repositories.GetAdminAssignedDomains(h.DB, targetUsername)

	assignedMap := make(map[string]bool)
	for _, da := range domainAdmins {
		assignedMap[da.Domain] = true
	}

	type DomainOption struct {
		Domain   string `json:"domain"`
		Assigned bool   `json:"assigned"`
	}
	var domainOptions []DomainOption
	for _, d := range allDomains {
		domainOptions = append(domainOptions, DomainOption{Domain: d.Domain, Assigned: assignedMap[d.Domain]})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"admin":   admin,
		"domains": domainOptions,
	})
}

// EditAdminAPI processes the update of an administrator via JSON API
func (h *Handler) EditAdminAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Permission check failed"})
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if !isSuper && loggedInUser != targetUsername {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}
	if targetUsername == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username required"})
	}

	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	changePassword := c.FormValue("change_password") == "true"
	domains := c.Request().Form["domains"]

	updates := map[string]interface{}{
		"modified":       time.Now(),
		"token_validity": time.Now().Add(3 * time.Hour),
	}
	if isSuper {
		updates["active"] = active
		updates["superadmin"] = superadmin
	}

	if changePassword && password != "" {
		if validationErr := ValidatePassword(password); validationErr != "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
		}
		if password != passwordConfirm {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
		}
		crypted, err := utils.HashPassword(password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to hash password"})
		}
		updates["password"] = crypted
	}

	if err := repositories.UpdateAdmin(h.DB, targetUsername, updates, domains, isSuper, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update admin"})
	}

	SetFlash(c, "message", "Admin updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// =====================================================
// API v1 Admin Handlers (PR 09+)
// =====================================================

// ListAdminsV1 godoc
// @Summary      List admins
// @Description  Returns all administrators. Superadmins see all; regular admins see only themselves.
// @Tags         Admins
// @Produce      json
// @Success      200  {array}   dto.AdminResponse
// @Failure      401  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /admins [get]
// @Security     BearerAuth
func (h *Handler) ListAdminsV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	admins, err := repositories.GetAllAdmins(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "failed to fetch administrators")
	}

	var response []dto.AdminResponse
	for _, admin := range admins {
		var domainCountStr string
		if admin.Superadmin {
			domainCountStr = "ALL"
		} else {
			count, _ := repositories.CountAdminDomains(h.DB, admin.Username)
			domainCountStr = fmt.Sprintf("%d", count)
		}
		response = append(response, dto.AdminResponse{
			Username:    admin.Username,
			Active:      admin.Active,
			Superadmin:  admin.Superadmin,
			Created:     admin.Created,
			Modified:    admin.Modified,
			DomainCount: domainCountStr,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateAdminV1 godoc
// @Summary      Create admin
// @Description  Creates a new administrator. Only superadmins can call this endpoint. Supply domains[] to assign domain access (ignored for superadmins).
// @Tags         Admins
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateAdminRequest  true  "Admin data"
// @Success      201      {object}  dto.APIResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse  "Superadmin required"
// @Failure      409      {object}  dto.APIResponse  "Admin already exists"
// @Failure      500      {object}  dto.APIResponse
// @Router       /admins [post]
// @Security     BearerAuth
func (h *Handler) CreateAdminV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	// Only superadmins can create other admins
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can create administrators")
	}

	var req dto.CreateAdminRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return dto.ValidationError(c, "username and password are required")
	}

	if validationErr := ValidatePassword(req.Password); validationErr != "" {
		return dto.ValidationError(c, validationErr)
	}

	if _, err := repositories.GetAdminByUsername(h.DB, req.Username); err == nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "admin already exists")
	}

	crypted, err := utils.HashPassword(req.Password)
	if err != nil {
		return dto.InternalError(c, "failed to hash password")
	}

	newAdmin := models.Admin{
		Username:      req.Username,
		Password:      crypted,
		Created:       time.Now(),
		Modified:      time.Now(),
		Active:        req.Active,
		Superadmin:    req.Superadmin,
		TokenValidity: time.Now().Add(3 * time.Hour),
	}

	if err := repositories.CreateAdmin(h.DB, newAdmin, req.Domains, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to create administrator")
	}

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, map[string]string{"username": req.Username})
}

// GetAdminV1 godoc
// @Summary      Get admin
// @Description  Returns details for a single administrator plus the full list of domains with assignment flags. Non-superadmins can only view themselves.
// @Tags         Admins
// @Produce      json
// @Param        username  path      string  true  "Admin username (e.g. admin@example.com)"
// @Success      200       {object}  dto.AdminDetailResponse
// @Failure      401       {object}  dto.APIResponse
// @Failure      403       {object}  dto.APIResponse
// @Failure      404       {object}  dto.APIResponse
// @Failure      500       {object}  dto.APIResponse
// @Router       /admins/{username} [get]
// @Security     BearerAuth
func (h *Handler) GetAdminV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if targetUsername == "" {
		return dto.ValidationError(c, "username required")
	}

	// Non-superadmins can only view themselves
	if !claims.Superadmin && claims.Username != targetUsername {
		return dto.Forbidden(c, "access denied")
	}

	admin, err := repositories.GetAdminByUsername(h.DB, targetUsername)
	if err != nil {
		return dto.NotFound(c, "admin not found")
	}

	allDomains, _, _ := repositories.GetActiveDomains(h.DB, claims.Username, claims.Superadmin)
	domainAdmins, _ := repositories.GetAdminAssignedDomains(h.DB, targetUsername)

	assignedMap := make(map[string]bool)
	for _, da := range domainAdmins {
		assignedMap[da.Domain] = true
	}

	var domainCountStr string
	if admin.Superadmin {
		domainCountStr = "ALL"
	} else {
		count, _ := repositories.CountAdminDomains(h.DB, targetUsername)
		domainCountStr = fmt.Sprintf("%d", count)
	}

	var domainOptions []dto.AdminDomainOption
	for _, d := range allDomains {
		domainOptions = append(domainOptions, dto.AdminDomainOption{Domain: d.Domain, Assigned: assignedMap[d.Domain]})
	}

	return dto.WriteSuccess(c, dto.AdminDetailResponse{
		Admin: dto.AdminResponse{
			Username:    admin.Username,
			Active:      admin.Active,
			Superadmin:  admin.Superadmin,
			Created:     admin.Created,
			Modified:    admin.Modified,
			DomainCount: domainCountStr,
		},
		Domains: domainOptions,
	})
}

// UpdateAdminV1 godoc
// @Summary      Update admin
// @Description  Updates an existing administrator. Superadmins can change active/superadmin flags and domain assignments. Any admin can change their own password via change_password=true.
// @Tags         Admins
// @Accept       json
// @Produce      json
// @Param        username  path      string                   true  "Admin username (e.g. admin@example.com)"
// @Param        request   body      dto.UpdateAdminRequest   true  "Fields to update"
// @Success      200       {object}  dto.APIResponse
// @Failure      400       {object}  dto.APIResponse
// @Failure      401       {object}  dto.APIResponse
// @Failure      403       {object}  dto.APIResponse
// @Failure      500       {object}  dto.APIResponse
// @Router       /admins/{username} [put]
// @Security     BearerAuth
func (h *Handler) UpdateAdminV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if targetUsername == "" {
		return dto.ValidationError(c, "username required")
	}

	// Non-superadmins can only update themselves (limited fields)
	isSuper := claims.Superadmin
	if !isSuper && claims.Username != targetUsername {
		return dto.Forbidden(c, "access denied")
	}

	var req dto.UpdateAdminRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	updates := map[string]interface{}{
		"modified":       time.Now(),
		"token_validity": time.Now().Add(3 * time.Hour),
	}

	if isSuper {
		if req.Active != nil {
			updates["active"] = *req.Active
		}
		if req.Superadmin != nil {
			updates["superadmin"] = *req.Superadmin
		}
	}

	if req.ChangePassword != nil && *req.ChangePassword {
		if req.Password == nil || req.PasswordConfirm == nil || *req.Password != *req.PasswordConfirm {
			return dto.ValidationError(c, "passwords do not match")
		}
		if validationErr := ValidatePassword(*req.Password); validationErr != "" {
			return dto.ValidationError(c, validationErr)
		}
		crypted, err := utils.HashPassword(*req.Password)
		if err != nil {
			return dto.InternalError(c, "failed to hash password")
		}
		updates["password"] = crypted
	}

	domains := req.Domains
	if err := repositories.UpdateAdmin(h.DB, targetUsername, updates, domains, isSuper, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to update admin")
	}

	return dto.WriteSuccess(c, map[string]bool{"updated": true})
}

// DeleteAdminV1 godoc
// @Summary      Delete admin
// @Description  Permanently deletes an administrator. Only superadmins can call this endpoint.
// @Tags         Admins
// @Produce      json
// @Param        username  path      string  true  "Admin username (e.g. admin@example.com)"
// @Success      200       {object}  dto.APIResponse
// @Failure      401       {object}  dto.APIResponse
// @Failure      403       {object}  dto.APIResponse  "Superadmin required"
// @Failure      404       {object}  dto.APIResponse
// @Failure      500       {object}  dto.APIResponse
// @Router       /admins/{username} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteAdminV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if targetUsername == "" {
		return dto.ValidationError(c, "username required")
	}

	if targetUsername == claims.Username {
		return dto.Forbidden(c, "you cannot delete your own administrator account")
	}

	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can delete administrators")
	}

	if err := repositories.DeleteAdmin(h.DB, targetUsername, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to delete administrator")
	}

	return dto.WriteSuccess(c, map[string]bool{"deleted": true})
}
