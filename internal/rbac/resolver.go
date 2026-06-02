package rbac

import (
	"fmt"
	"slices"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// ResolvePermissions returns the deduplicated permission names and role names
// for the given admin, derived from their assigned RBAC roles.
//
// Superadmin bypass: when superadmin is true the function returns [PermWildcard]
// as the single permission and [RoleSuperadmin] as the role, without touching
// the database. Middleware checks [PermWildcard] first, so no per-endpoint
// enumeration is needed for superadmin accounts.
//
// Backward-compatibility: admins that have entries in domain_admins but no
// explicit RBAC roles in rbac_admin_roles automatically receive the permissions
// of the built-in "domain_admin" role. This ensures existing admins keep
// working after rbac.enabled=true without requiring manual role assignment.
//
// Returns an error only on unexpected database failures; a missing admin or
// empty role list is not an error.
func ResolvePermissions(db *gorm.DB, username string, superadmin bool) (permissions []string, roles []string, err error) {
	if superadmin {
		return []string{PermWildcard}, []string{RoleSuperadmin}, nil
	}

	var adminRoles []models.RBACAdminRole
	if err := db.Preload("Role.Permissions").
		Where("username = ?", username).
		Find(&adminRoles).Error; err != nil {
		return nil, nil, fmt.Errorf("rbac: resolve permissions for %q: %w", username, err)
	}

	// When the admin has explicit RBAC role assignments, use those.
	if len(adminRoles) > 0 {
		permSet := make(map[string]struct{})
		roleSet := make(map[string]struct{})

		for _, ar := range adminRoles {
			roleSet[ar.Role.Name] = struct{}{}
			for _, p := range ar.Role.Permissions {
				permSet[p.Name] = struct{}{}
			}
		}

		permissions = make([]string, 0, len(permSet))
		for p := range permSet {
			permissions = append(permissions, p)
		}
		roles = make([]string, 0, len(roleSet))
		for r := range roleSet {
			roles = append(roles, r)
		}
		return permissions, roles, nil
	}

	// Backward-compat fallback: admin has domain_admins rows but no RBAC roles yet.
	// Derive permissions from the built-in "domain_admin" role so they keep working
	// after rbac.enabled=true without requiring a manual role assignment step.
	var domainCount int64
	db.Model(&models.DomainAdmin{}).
		Where("username = ? AND active = ?", username, true).
		Count(&domainCount)

	if domainCount > 0 {
		var role models.RBACRole
		if err := db.Preload("Permissions").
			Where("name = ?", RoleDomainAdmin).
			First(&role).Error; err == nil {
			perms := make([]string, 0, len(role.Permissions))
			for _, p := range role.Permissions {
				perms = append(perms, p.Name)
			}
			return perms, []string{RoleDomainAdmin}, nil
		}
	}

	return []string{}, []string{}, nil
}

// HasPermission reports whether the given permission slice grants perm.
// [PermWildcard] ("*") in the slice bypasses all checks (superadmin fast-path).
func HasPermission(permissions []string, perm string) bool {
	return slices.Contains(permissions, PermWildcard) || slices.Contains(permissions, perm)
}

// HasRole reports whether the given role slice contains any of the wanted roles.
func HasRole(roles []string, want ...string) bool {
	for _, w := range want {
		if slices.Contains(roles, w) {
			return true
		}
	}
	return false
}
