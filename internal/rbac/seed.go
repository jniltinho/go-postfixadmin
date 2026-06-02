// Package rbac provides role-based access control for go-postfixadmin.
// It exposes a permission resolver, a database seeder, and constants for
// the built-in system roles and permission names.
//
// Roles marked as system=true are seeded at startup and cannot be deleted
// or modified through the management API.
package rbac

import (
	"fmt"
	"time"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Built-in system role names. These constants are used by the resolver and
// the middleware to apply special-case logic (e.g. superadmin bypass).
const (
	RoleSuperadmin   = "superadmin"
	RoleDomainAdmin  = "domain_admin"
	RoleMailboxAdmin = "mailbox_admin"
	RoleAliasAdmin   = "alias_admin"
	RoleViewer       = "viewer"
	RoleReportViewer = "report_viewer"
)

// Permission name constants in the format "<resource>:<action>".
// These are the canonical strings embedded in JWT claims and checked by
// [RequirePermission] middleware.
const (
	PermDomainsRead        = "domains:read"
	PermDomainsWrite       = "domains:write"
	PermDomainsDelete      = "domains:delete"
	PermMailboxesRead      = "mailboxes:read"
	PermMailboxesWrite     = "mailboxes:write"
	PermMailboxesDelete    = "mailboxes:delete"
	PermAliasesRead        = "aliases:read"
	PermAliasesWrite       = "aliases:write"
	PermAliasesDelete      = "aliases:delete"
	PermAliasDomainsRead   = "alias_domains:read"
	PermAliasDomainsWrite  = "alias_domains:write"
	PermAliasDomainsDelete = "alias_domains:delete"
	PermAdminsRead         = "admins:read"
	PermAdminsWrite        = "admins:write"
	PermAdminsDelete       = "admins:delete"
	PermTransportsRead     = "transports:read"
	PermTransportsWrite    = "transports:write"
	PermTransportsDelete   = "transports:delete"
	PermLogsRead           = "logs:read"
	PermDashboardRead      = "dashboard:read"
	PermSettingsRead       = "settings:read"
	PermSettingsWrite      = "settings:write"

	// PermProfileRead and PermProfileWrite allow an admin to view and edit their
	// own account without requiring the broader admins:read / admins:write
	// permissions that are reserved for superadmin-level user management.
	PermProfileRead  = "profile:read"
	PermProfileWrite = "profile:write"

	// PermAPIKeysRead and PermAPIKeysWrite control access to the personal API key
	// management endpoints (/settings/apikeys). Separated from settings:read/write
	// so that API key access can be granted independently of other settings.
	PermAPIKeysRead  = "apikeys:read"
	PermAPIKeysWrite = "apikeys:write"

	// Domain sub-permissions for fine-grained UI control.
	//
	// domain_adset: controls the Advanced Settings section in the Edit Domain modal.
	//   read  → section visible but all fields read-only (cannot be changed)
	//   write → section visible and fully editable
	//
	// domain_add: controls the Add Domain button.
	//   read  → button hidden / creation disabled
	//   write → button visible and creation allowed
	PermDomainAdsetRead  = "domain_adset:read"
	PermDomainAdsetWrite = "domain_adset:write"
	PermDomainAddRead    = "domain_add:read"
	PermDomainAddWrite   = "domain_add:write"

	// PermWildcard is embedded in JWT claims for superadmin users, allowing
	// the middleware to bypass all permission checks with a single comparison.
	PermWildcard = "*"
)

// systemRoles defines the built-in roles that are seeded on every startup.
var systemRoles = []models.RBACRole{
	{Name: RoleSuperadmin, Description: "Full system access — all resources and actions", System: true},
	{Name: RoleDomainAdmin, Description: "Full CRUD on assigned domains and their mailboxes and aliases", System: true},
	{Name: RoleMailboxAdmin, Description: "Manage mailboxes and aliases within assigned domains", System: true},
	{Name: RoleAliasAdmin, Description: "Manage aliases and alias domains within assigned domains", System: true},
	{Name: RoleViewer, Description: "Read-only access to all resources within assigned scope", System: true},
	{Name: RoleReportViewer, Description: "Access to dashboard statistics and log viewers only", System: true},
}

// allPermissions is the complete permission catalog. Every entry maps a human-readable
// name to its resource and action components.
var allPermissions = []models.RBACPermission{
	{Name: PermDomainsRead, Resource: "domains", Action: "read"},
	{Name: PermDomainsWrite, Resource: "domains", Action: "write"},
	{Name: PermDomainsDelete, Resource: "domains", Action: "delete"},
	{Name: PermMailboxesRead, Resource: "mailboxes", Action: "read"},
	{Name: PermMailboxesWrite, Resource: "mailboxes", Action: "write"},
	{Name: PermMailboxesDelete, Resource: "mailboxes", Action: "delete"},
	{Name: PermAliasesRead, Resource: "aliases", Action: "read"},
	{Name: PermAliasesWrite, Resource: "aliases", Action: "write"},
	{Name: PermAliasesDelete, Resource: "aliases", Action: "delete"},
	{Name: PermAliasDomainsRead, Resource: "alias_domains", Action: "read"},
	{Name: PermAliasDomainsWrite, Resource: "alias_domains", Action: "write"},
	{Name: PermAliasDomainsDelete, Resource: "alias_domains", Action: "delete"},
	{Name: PermAdminsRead, Resource: "admins", Action: "read"},
	{Name: PermAdminsWrite, Resource: "admins", Action: "write"},
	{Name: PermAdminsDelete, Resource: "admins", Action: "delete"},
	{Name: PermTransportsRead, Resource: "transports", Action: "read"},
	{Name: PermTransportsWrite, Resource: "transports", Action: "write"},
	{Name: PermTransportsDelete, Resource: "transports", Action: "delete"},
	{Name: PermLogsRead, Resource: "logs", Action: "read"},
	{Name: PermDashboardRead, Resource: "dashboard", Action: "read"},
	{Name: PermSettingsRead, Resource: "settings", Action: "read"},
	{Name: PermSettingsWrite, Resource: "settings", Action: "write"},
	{Name: PermProfileRead, Resource: "profile", Action: "read"},
	{Name: PermProfileWrite, Resource: "profile", Action: "write"},
	{Name: PermAPIKeysRead, Resource: "apikeys", Action: "read"},
	{Name: PermAPIKeysWrite, Resource: "apikeys", Action: "write"},
	{Name: PermDomainAdsetRead, Resource: "domain_adset", Action: "read"},
	{Name: PermDomainAdsetWrite, Resource: "domain_adset", Action: "write"},
	{Name: PermDomainAddRead, Resource: "domain_add", Action: "read"},
	{Name: PermDomainAddWrite, Resource: "domain_add", Action: "write"},
}

// rolePermissions maps each system role name to its assigned permission names.
var rolePermissions = map[string][]string{
	RoleSuperadmin: {
		PermDomainsRead, PermDomainsWrite, PermDomainsDelete,
		PermMailboxesRead, PermMailboxesWrite, PermMailboxesDelete,
		PermAliasesRead, PermAliasesWrite, PermAliasesDelete,
		PermAliasDomainsRead, PermAliasDomainsWrite, PermAliasDomainsDelete,
		PermAdminsRead, PermAdminsWrite, PermAdminsDelete,
		PermTransportsRead, PermTransportsWrite, PermTransportsDelete,
		PermLogsRead, PermDashboardRead,
		PermSettingsRead, PermSettingsWrite,
		PermAPIKeysRead, PermAPIKeysWrite,
		PermProfileRead, PermProfileWrite,
		PermDomainAdsetRead, PermDomainAdsetWrite,
		PermDomainAddRead, PermDomainAddWrite,
	},
	RoleDomainAdmin: {
		PermDomainsRead, PermDomainsWrite, PermDomainsDelete,
		PermMailboxesRead, PermMailboxesWrite, PermMailboxesDelete,
		PermAliasesRead, PermAliasesWrite, PermAliasesDelete,
		PermAliasDomainsRead, PermAliasDomainsWrite, PermAliasDomainsDelete,
		PermLogsRead, PermDashboardRead,
		PermSettingsRead, PermSettingsWrite,
		PermAPIKeysRead, PermAPIKeysWrite,
		PermProfileRead, PermProfileWrite,
		PermDomainAdsetRead, // can view advanced settings but not edit
		PermDomainAddRead,   // cannot add new domains
	},
	RoleMailboxAdmin: {
		PermDomainsRead,
		PermMailboxesRead, PermMailboxesWrite, PermMailboxesDelete,
		PermAliasesRead, PermAliasesWrite, PermAliasesDelete,
		PermLogsRead, PermDashboardRead,
		PermSettingsRead, PermSettingsWrite,
		PermAPIKeysRead, PermAPIKeysWrite,
		PermProfileRead, PermProfileWrite,
	},
	RoleAliasAdmin: {
		PermDomainsRead,
		PermAliasesRead, PermAliasesWrite, PermAliasesDelete,
		PermAliasDomainsRead, PermAliasDomainsWrite, PermAliasDomainsDelete,
		PermLogsRead, PermDashboardRead,
		PermSettingsRead, PermSettingsWrite,
		PermAPIKeysRead, PermAPIKeysWrite,
		PermProfileRead, PermProfileWrite,
	},
	RoleViewer: {
		PermDomainsRead, PermMailboxesRead, PermAliasesRead,
		PermAliasDomainsRead, PermAdminsRead, PermTransportsRead,
		PermLogsRead, PermDashboardRead, PermSettingsRead,
		PermAPIKeysRead,
		PermProfileRead,
		PermDomainAdsetRead, // can view advanced settings read-only
		PermDomainAddRead,   // cannot add new domains
	},
	RoleReportViewer: {
		PermDashboardRead, PermLogsRead,
		PermProfileRead,
	},
}

// Seed inserts the built-in roles, the full permission catalog, and the
// role-to-permission assignments. All inserts use ON CONFLICT DO NOTHING so
// the operation is idempotent and safe to run on every startup.
//
// Seed must be called after AutoMigrate (or the equivalent manual migration)
// has created the rbac_* tables.
func Seed(db *gorm.DB) error {
	now := time.Now()

	// 1. Upsert system roles (update description if name already exists).
	for i := range systemRoles {
		systemRoles[i].Created = now
		systemRoles[i].Modified = now
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"description", "modified"}),
	}).Create(&systemRoles).Error; err != nil {
		return fmt.Errorf("rbac seed: upsert roles: %w", err)
	}

	// 2. Upsert the permission catalog (no-op on conflict).
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&allPermissions).Error; err != nil {
		return fmt.Errorf("rbac seed: upsert permissions: %w", err)
	}

	// 3. Wire role-permission associations using the persisted IDs.
	return seedRolePermissions(db)
}

// seedRolePermissions looks up each system role and its expected permissions
// by name, then sets the many-to-many association using GORM's Replace strategy.
// Existing associations that are no longer in the definition are removed.
func seedRolePermissions(db *gorm.DB) error {
	// Build a name→model map for quick lookup.
	var roles []models.RBACRole
	if err := db.Where("system = ?", true).Find(&roles).Error; err != nil {
		return fmt.Errorf("rbac seed: load system roles: %w", err)
	}
	roleByName := make(map[string]models.RBACRole, len(roles))
	for _, r := range roles {
		roleByName[r.Name] = r
	}

	var perms []models.RBACPermission
	if err := db.Find(&perms).Error; err != nil {
		return fmt.Errorf("rbac seed: load permissions: %w", err)
	}
	permByName := make(map[string]models.RBACPermission, len(perms))
	for _, p := range perms {
		permByName[p.Name] = p
	}

	for roleName, permNames := range rolePermissions {
		role, ok := roleByName[roleName]
		if !ok {
			continue
		}

		assigned := make([]models.RBACPermission, 0, len(permNames))
		for _, pn := range permNames {
			if p, exists := permByName[pn]; exists {
				assigned = append(assigned, p)
			}
		}

		if err := db.Model(&role).Association("Permissions").Replace(assigned); err != nil {
			return fmt.Errorf("rbac seed: assign permissions to role %q: %w", roleName, err)
		}
	}

	return nil
}
