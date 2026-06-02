package repositories

import (
	"fmt"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// ListRoles returns all roles ordered by name, with their permissions preloaded.
func ListRoles(db *gorm.DB) ([]models.RBACRole, error) {
	var roles []models.RBACRole
	if err := db.Preload("Permissions").Order("name").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("rbac: list roles: %w", err)
	}
	return roles, nil
}

// GetRoleByID returns a single role with its permissions.
// Returns gorm.ErrRecordNotFound when the id does not exist.
func GetRoleByID(db *gorm.DB, id int) (models.RBACRole, error) {
	var role models.RBACRole
	if err := db.Preload("Permissions").First(&role, id).Error; err != nil {
		return models.RBACRole{}, fmt.Errorf("rbac: get role %d: %w", id, err)
	}
	return role, nil
}

// CreateRole inserts a new custom role. System roles are managed by the seeder
// and must not be created through this function.
func CreateRole(db *gorm.DB, role models.RBACRole) (models.RBACRole, error) {
	if err := db.Create(&role).Error; err != nil {
		return models.RBACRole{}, fmt.Errorf("rbac: create role %q: %w", role.Name, err)
	}
	return role, nil
}

// UpdateRole replaces the description and permission set of a non-system role.
// Returns an error when the target role has System=true.
func UpdateRole(db *gorm.DB, id int, description string, permissionIDs []int) (models.RBACRole, error) {
	role, err := GetRoleByID(db, id)
	if err != nil {
		return models.RBACRole{}, err
	}
	if role.System {
		return models.RBACRole{}, fmt.Errorf("rbac: update role %d: cannot modify a system role", id)
	}

	if err := db.Model(&role).Update("description", description).Error; err != nil {
		return models.RBACRole{}, fmt.Errorf("rbac: update role %d description: %w", id, err)
	}

	// Resolve the permission records and replace the association.
	var perms []models.RBACPermission
	if len(permissionIDs) > 0 {
		if err := db.Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
			return models.RBACRole{}, fmt.Errorf("rbac: resolve permissions for role %d: %w", id, err)
		}
	}
	if err := db.Model(&role).Association("Permissions").Replace(perms); err != nil {
		return models.RBACRole{}, fmt.Errorf("rbac: replace permissions for role %d: %w", id, err)
	}

	return GetRoleByID(db, id)
}

// DeleteRole permanently removes a non-system role.
// It clears the many2many permission associations first to avoid FK constraint
// violations on rbac_role_permissions, then deletes the role itself.
// rbac_admin_roles rows are removed via ON DELETE CASCADE.
// Returns an error when the target role has System=true.
func DeleteRole(db *gorm.DB, id int) error {
	role, err := GetRoleByID(db, id)
	if err != nil {
		return err
	}
	if role.System {
		return fmt.Errorf("rbac: delete role %d: cannot delete a system role", id)
	}
	// Clear the join table rows before deleting the role to avoid FK violations.
	if err := db.Model(&role).Association("Permissions").Clear(); err != nil {
		return fmt.Errorf("rbac: clear permissions for role %d: %w", id, err)
	}
	if err := db.Delete(&role).Error; err != nil {
		return fmt.Errorf("rbac: delete role %d: %w", id, err)
	}
	return nil
}

// ListPermissions returns the full permission catalog ordered by name.
func ListPermissions(db *gorm.DB) ([]models.RBACPermission, error) {
	var perms []models.RBACPermission
	if err := db.Order("name").Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("rbac: list permissions: %w", err)
	}
	return perms, nil
}

// ListAdminRoles returns all role assignments for the given admin username,
// with the associated role and its permissions preloaded.
func ListAdminRoles(db *gorm.DB, username string) ([]models.RBACAdminRole, error) {
	var assignments []models.RBACAdminRole
	if err := db.Preload("Role.Permissions").
		Where("username = ?", username).
		Find(&assignments).Error; err != nil {
		return nil, fmt.Errorf("rbac: list admin roles for %q: %w", username, err)
	}
	return assignments, nil
}

// AssignRole creates a new role assignment for the given admin.
// A domain of "" means the role applies globally to all domains the admin manages.
// Returns an error on unique-constraint violation (duplicate assignment).
func AssignRole(db *gorm.DB, username string, roleID int, domain string) (models.RBACAdminRole, error) {
	ar := models.RBACAdminRole{
		Username: username,
		RoleID:   roleID,
		Domain:   domain,
	}
	if err := db.Create(&ar).Error; err != nil {
		return models.RBACAdminRole{}, fmt.Errorf("rbac: assign role %d to %q (domain=%q): %w", roleID, username, domain, err)
	}
	// Reload with associations.
	if err := db.Preload("Role.Permissions").First(&ar, ar.ID).Error; err != nil {
		return models.RBACAdminRole{}, fmt.Errorf("rbac: reload assignment %d: %w", ar.ID, err)
	}
	return ar, nil
}

// RemoveAdminRole deletes a specific role assignment by its primary key.
// It verifies that the assignment belongs to the given username before deleting.
func RemoveAdminRole(db *gorm.DB, username string, assignmentID int) error {
	var ar models.RBACAdminRole
	if err := db.Where("id = ? AND username = ?", assignmentID, username).First(&ar).Error; err != nil {
		return fmt.Errorf("rbac: assignment %d not found for %q: %w", assignmentID, username, err)
	}
	if err := db.Delete(&ar).Error; err != nil {
		return fmt.Errorf("rbac: remove assignment %d: %w", assignmentID, err)
	}
	return nil
}
