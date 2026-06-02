package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/rbac"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// toPermResponse converts a models.RBACPermission to its DTO representation.
func toPermResponse(p models.RBACPermission) dto.RBACPermissionResponse {
	return dto.RBACPermissionResponse{
		ID:       p.ID,
		Name:     p.Name,
		Resource: p.Resource,
		Action:   p.Action,
	}
}

// toRoleResponse converts a models.RBACRole to its DTO representation.
// The permissions slice is populated only when the model has them preloaded.
func toRoleResponse(r models.RBACRole) dto.RBACRoleResponse {
	perms := make([]dto.RBACPermissionResponse, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, toPermResponse(p))
	}
	return dto.RBACRoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		System:      r.System,
		Created:     r.Created,
		Modified:    r.Modified,
		Permissions: perms,
	}
}

// requireSuperadmin is a shared guard for RBAC management endpoints.
// Returns nil when the request carries valid superadmin claims.
func requireSuperadmin(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin && !rbac.HasPermission(claims.Permissions, rbac.PermSettingsWrite) {
		return dto.Forbidden(c, "superadmin or settings:write permission required")
	}
	return nil
}

// ========== Roles ==========

// ListRBACRoles godoc
// @Summary      List RBAC roles
// @Description  Returns all roles (system and custom) with their assigned permissions.
// @Tags         RBAC
// @Produce      json
// @Success      200  {array}   dto.RBACRoleResponse
// @Failure      401  {object}  dto.APIResponse
// @Failure      403  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /rbac/roles [get]
// @Security     BearerAuth
func (h *Handler) ListRBACRoles(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	roles, err := repositories.ListRoles(h.DB)
	if err != nil {
		return dto.InternalError(c, "failed to list roles")
	}

	resp := make([]dto.RBACRoleResponse, 0, len(roles))
	for _, r := range roles {
		resp = append(resp, toRoleResponse(r))
	}
	return dto.WriteSuccess(c, resp)
}

// GetRBACRole godoc
// @Summary      Get RBAC role
// @Description  Returns a single role with its full permission list.
// @Tags         RBAC
// @Produce      json
// @Param        id   path      int  true  "Role ID"
// @Success      200  {object}  dto.RBACRoleResponse
// @Failure      401  {object}  dto.APIResponse
// @Failure      403  {object}  dto.APIResponse
// @Failure      404  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /rbac/roles/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetRBACRole(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dto.BadRequest(c, "invalid role id")
	}

	role, err := repositories.GetRoleByID(h.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.NotFound(c, "role not found")
		}
		return dto.InternalError(c, "failed to get role")
	}

	return dto.WriteSuccess(c, toRoleResponse(role))
}

// CreateRBACRole godoc
// @Summary      Create RBAC role
// @Description  Creates a new custom role. System roles are managed by the seeder and cannot be created here.
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateRoleRequest  true  "Role data"
// @Success      201      {object}  dto.RBACRoleResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse
// @Failure      409      {object}  dto.APIResponse  "Name already in use"
// @Failure      500      {object}  dto.APIResponse
// @Router       /rbac/roles [post]
// @Security     BearerAuth
func (h *Handler) CreateRBACRole(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	var req dto.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return dto.ValidationError(c, "name is required")
	}

	role, err := repositories.CreateRole(h.DB, models.RBACRole{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "role name already exists")
	}

	// Assign permissions if provided.
	if len(req.PermissionIDs) > 0 {
		var perms []models.RBACPermission
		if err := h.DB.Where("id IN ?", req.PermissionIDs).Find(&perms).Error; err == nil {
			_ = h.DB.Model(&role).Association("Permissions").Replace(perms)
		}
		// Reload to include associations in the response.
		_ = h.DB.Preload("Permissions").First(&role, role.ID)
	}

	claims := middleware.GetJWTClaims(c)
	_ = utils.LogAction(h.DB, claims.Username, c.RealIP(), "", "rbac_create_role", req.Name)

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, toRoleResponse(role))
}

// UpdateRBACRole godoc
// @Summary      Update RBAC role
// @Description  Updates the description and permission set of a custom role. System roles cannot be modified.
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Role ID"
// @Param        request  body      dto.UpdateRoleRequest  true  "Fields to update"
// @Success      200      {object}  dto.RBACRoleResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse  "Superadmin required or system role"
// @Failure      404      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /rbac/roles/{id} [put]
// @Security     BearerAuth
func (h *Handler) UpdateRBACRole(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dto.BadRequest(c, "invalid role id")
	}

	var req dto.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	role, err := repositories.UpdateRole(h.DB, id, desc, req.PermissionIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.NotFound(c, "role not found")
		}
		return dto.Forbidden(c, err.Error())
	}

	claims := middleware.GetJWTClaims(c)
	_ = utils.LogAction(h.DB, claims.Username, c.RealIP(), "", "rbac_update_role", strconv.Itoa(id))

	return dto.WriteSuccess(c, toRoleResponse(role))
}

// DeleteRBACRole godoc
// @Summary      Delete RBAC role
// @Description  Permanently deletes a custom role and all its admin assignments. System roles cannot be deleted.
// @Tags         RBAC
// @Produce      json
// @Param        id  path      int  true  "Role ID"
// @Success      200 {object}  dto.APIResponse
// @Failure      401 {object}  dto.APIResponse
// @Failure      403 {object}  dto.APIResponse  "Superadmin required or system role"
// @Failure      404 {object}  dto.APIResponse
// @Failure      500 {object}  dto.APIResponse
// @Router       /rbac/roles/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteRBACRole(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dto.BadRequest(c, "invalid role id")
	}

	if err := repositories.DeleteRole(h.DB, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.NotFound(c, "role not found")
		}
		return dto.Forbidden(c, err.Error())
	}

	claims := middleware.GetJWTClaims(c)
	_ = utils.LogAction(h.DB, claims.Username, c.RealIP(), "", "rbac_delete_role", strconv.Itoa(id))

	return dto.WriteSuccess(c, map[string]bool{"deleted": true})
}

// ========== Permissions ==========

// ListRBACPermissions godoc
// @Summary      List RBAC permissions
// @Description  Returns the full permission catalog sorted by name.
// @Tags         RBAC
// @Produce      json
// @Success      200  {array}   dto.RBACPermissionResponse
// @Failure      401  {object}  dto.APIResponse
// @Failure      403  {object}  dto.APIResponse
// @Failure      500  {object}  dto.APIResponse
// @Router       /rbac/permissions [get]
// @Security     BearerAuth
func (h *Handler) ListRBACPermissions(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	perms, err := repositories.ListPermissions(h.DB)
	if err != nil {
		return dto.InternalError(c, "failed to list permissions")
	}

	resp := make([]dto.RBACPermissionResponse, 0, len(perms))
	for _, p := range perms {
		resp = append(resp, toPermResponse(p))
	}
	return dto.WriteSuccess(c, resp)
}

// ========== Admin role assignments ==========

// ListAdminRoles godoc
// @Summary      List admin roles
// @Description  Returns all RBAC role assignments for the given admin username.
// @Tags         RBAC
// @Produce      json
// @Param        username  path      string  true  "Admin username"
// @Success      200       {array}   dto.RBACAdminRoleResponse
// @Failure      401       {object}  dto.APIResponse
// @Failure      403       {object}  dto.APIResponse
// @Failure      500       {object}  dto.APIResponse
// @Router       /rbac/admins/{username}/roles [get]
// @Security     BearerAuth
func (h *Handler) ListAdminRoles(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	username, _ := url.PathUnescape(c.Param("username"))
	if username == "" {
		return dto.ValidationError(c, "username required")
	}

	assignments, err := repositories.ListAdminRoles(h.DB, username)
	if err != nil {
		return dto.InternalError(c, "failed to list admin roles")
	}

	resp := make([]dto.RBACAdminRoleResponse, 0, len(assignments))
	for _, ar := range assignments {
		resp = append(resp, dto.RBACAdminRoleResponse{
			ID:       ar.ID,
			Username: ar.Username,
			RoleID:   ar.RoleID,
			Domain:   ar.Domain,
			Created:  ar.Created,
			Role:     toRoleResponse(ar.Role),
		})
	}
	return dto.WriteSuccess(c, resp)
}

// AssignAdminRole godoc
// @Summary      Assign role to admin
// @Description  Assigns a role to the given admin, optionally scoped to a specific domain. An empty domain means global scope (all domains the admin manages).
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Param        username  path      string                 true  "Admin username"
// @Param        request   body      dto.AssignRoleRequest  true  "Assignment data"
// @Success      201       {object}  dto.RBACAdminRoleResponse
// @Failure      400       {object}  dto.APIResponse
// @Failure      401       {object}  dto.APIResponse
// @Failure      403       {object}  dto.APIResponse
// @Failure      409       {object}  dto.APIResponse  "Duplicate assignment"
// @Failure      500       {object}  dto.APIResponse
// @Router       /rbac/admins/{username}/roles [post]
// @Security     BearerAuth
func (h *Handler) AssignAdminRole(c *echo.Context) error {
	if err := requireSuperadmin(c); err != nil {
		return err
	}

	username, _ := url.PathUnescape(c.Param("username"))
	if username == "" {
		return dto.ValidationError(c, "username required")
	}

	var req dto.AssignRoleRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}
	if req.RoleID == 0 {
		return dto.ValidationError(c, "role_id is required")
	}

	ar, err := repositories.AssignRole(h.DB, username, req.RoleID, req.Domain)
	if err != nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "role already assigned to this admin for the given domain")
	}

	claims := middleware.GetJWTClaims(c)
	_ = utils.LogAction(h.DB, claims.Username, c.RealIP(), req.Domain, "rbac_assign_role",
		username+" role="+strconv.Itoa(req.RoleID))

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, dto.RBACAdminRoleResponse{
		ID:       ar.ID,
		Username: ar.Username,
		RoleID:   ar.RoleID,
		Domain:   ar.Domain,
		Created:  ar.Created,
		Role:     toRoleResponse(ar.Role),
	})
}

// RemoveAdminRole godoc
// @Summary      Remove role from admin
// @Description  Removes a specific role assignment from the given admin. Superadmins cannot remove the superadmin role from themselves.
// @Tags         RBAC
// @Produce      json
// @Param        username      path      string  true  "Admin username"
// @Param        assignmentId  path      int     true  "Assignment ID (from list endpoint)"
// @Success      200           {object}  dto.APIResponse
// @Failure      400           {object}  dto.APIResponse
// @Failure      401           {object}  dto.APIResponse
// @Failure      403           {object}  dto.APIResponse
// @Failure      404           {object}  dto.APIResponse
// @Failure      500           {object}  dto.APIResponse
// @Router       /rbac/admins/{username}/roles/{assignmentId} [delete]
// @Security     BearerAuth
func (h *Handler) RemoveAdminRole(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin && !rbac.HasPermission(claims.Permissions, rbac.PermSettingsWrite) {
		return dto.Forbidden(c, "superadmin or settings:write permission required")
	}

	username, _ := url.PathUnescape(c.Param("username"))
	if username == "" {
		return dto.ValidationError(c, "username required")
	}

	assignmentID, err := strconv.Atoi(c.Param("assignmentId"))
	if err != nil {
		return dto.BadRequest(c, "invalid assignment id")
	}

	// Prevent a superadmin from removing the superadmin role assignment from
	// themselves — this would silently downgrade their own access.
	if username == claims.Username {
		ar, lookupErr := repositories.ListAdminRoles(h.DB, username)
		if lookupErr == nil {
			for _, a := range ar {
				if a.ID == assignmentID && a.Role.Name == rbac.RoleSuperadmin {
					return dto.Forbidden(c, "you cannot remove the superadmin role from yourself")
				}
			}
		}
	}

	if err := repositories.RemoveAdminRole(h.DB, username, assignmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.NotFound(c, "assignment not found")
		}
		return dto.InternalError(c, "failed to remove role assignment")
	}

	_ = utils.LogAction(h.DB, claims.Username, c.RealIP(), "", "rbac_remove_role",
		username+" assignment="+strconv.Itoa(assignmentID))

	return dto.WriteSuccess(c, map[string]bool{"deleted": true})
}
