package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/internal/database"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/rbac"
	"go-postfixadmin/internal/repositories"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var rbacCmd = &cobra.Command{
	Use:   "rbac",
	Short: "RBAC role management utilities",
}

// rbacAssignCmd assigns a built-in or custom role to an admin account.
// Usage: postfixadmin rbac assign <username> <role> [domain]
var rbacAssignCmd = &cobra.Command{
	Use:   "assign <username> <role> [domain]",
	Short: "Assign a role to an admin (optionally scoped to a domain)",
	Long: `Assigns the named role to the given admin.

The optional domain argument scopes the assignment: the role's permissions
only apply when the admin is acting on that specific domain. Omit it (or
pass "") for a global assignment covering all domains the admin manages.

Examples:
  postfixadmin rbac assign alice@example.com domain_admin
  postfixadmin rbac assign alice@example.com mailbox_admin example.com
  postfixadmin rbac assign alice@example.com viewer`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		roleName := args[1]
		domain := ""
		if len(args) == 3 {
			domain = args[2]
		}

		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}

		// Verify the admin account exists.
		var admin models.Admin
		if err := db.Where("username = ?", username).First(&admin).Error; err != nil {
			slog.Error("Admin not found", "username", username)
			os.Exit(1)
		}

		// Look up the role by name.
		var role models.RBACRole
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			slog.Error("Role not found — run 'migrate rbac' first", "role", roleName)
			os.Exit(1)
		}

		// Assign (idempotent via unique constraint).
		_, err = repositories.AssignRole(db, username, role.ID, domain)
		if err != nil {
			slog.Error("Failed to assign role", "error", err)
			os.Exit(1)
		}

		slog.Info("Role assigned successfully",
			"username", username,
			"role", roleName,
			"domain", domain,
		)
	},
}

// rbacSeedExistingCmd assigns the domain_admin role to every admin that has
// entries in domain_admins but no RBAC role assignments yet.
// This is the one-shot migration for existing deployments.
var rbacSeedExistingCmd = &cobra.Command{
	Use:   "seed-existing",
	Short: "Assign domain_admin role to all admins that have domain_admins but no RBAC roles",
	Long: `Scans the admin table and, for each active non-superadmin account that has
at least one entry in domain_admins but no row in rbac_admin_roles, inserts a
global domain_admin assignment.

This command is safe to run multiple times — it skips admins that already
have any RBAC role assigned.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}

		// Ensure the domain_admin role exists.
		var domainAdminRole models.RBACRole
		if err := db.Where("name = ?", rbac.RoleDomainAdmin).First(&domainAdminRole).Error; err != nil {
			slog.Error("domain_admin role not found — run 'migrate rbac' first")
			os.Exit(1)
		}

		migrated := seedExistingAdmins(db, domainAdminRole.ID)
		slog.Info("Seed complete", "assigned", migrated)
	},
}

// seedExistingAdmins iterates all active non-superadmin accounts, checks if
// they have domain_admins rows but no rbac_admin_roles rows, and if so
// assigns them the domain_admin role globally. Returns the number assigned.
func seedExistingAdmins(db *gorm.DB, domainAdminRoleID int) int {
	var admins []models.Admin
	db.Where("active = ? AND superadmin = ?", true, false).Find(&admins)

	assigned := 0
	for _, admin := range admins {
		// Skip if already has any RBAC role.
		var count int64
		db.Model(&models.RBACAdminRole{}).Where("username = ?", admin.Username).Count(&count)
		if count > 0 {
			continue
		}

		// Skip if no domain_admins entries.
		var domainCount int64
		db.Model(&models.DomainAdmin{}).Where("username = ? AND active = ?", admin.Username, true).Count(&domainCount)
		if domainCount == 0 {
			continue
		}

		if _, err := repositories.AssignRole(db, admin.Username, domainAdminRoleID, ""); err != nil {
			slog.Warn("Failed to assign role", "username", admin.Username, "error", err)
			continue
		}
		slog.Info("Assigned domain_admin", "username", admin.Username, "domains", domainCount)
		assigned++
	}
	return assigned
}

func init() {
	rbacCmd.AddCommand(rbacAssignCmd)
	rbacCmd.AddCommand(rbacSeedExistingCmd)
	rootCmd.AddCommand(rbacCmd)
}
