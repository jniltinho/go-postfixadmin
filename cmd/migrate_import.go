package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/internal/database"
	"go-postfixadmin/internal/rbac"
	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migration",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Failed to connect to database for migration", "error", err)
			os.Exit(1)
		}

		slog.Info("Running database migration...")
		if err := database.MigrateDB(db); err != nil {
			slog.Error("Database migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migration completed successfully.")
	},
}

// migrateRBACCmd creates the RBAC tables and seeds the built-in roles and
// permissions. It is safe to run multiple times — all inserts are idempotent.
var migrateRBACCmd = &cobra.Command{
	Use:   "rbac",
	Short: "Create RBAC tables and seed built-in roles and permissions",
	Long: `Creates the four RBAC tables (rbac_roles, rbac_permissions,
rbac_role_permissions, rbac_admin_roles) and populates them with the six
system roles and the 22-entry permission catalog.

This command is idempotent: running it multiple times is safe.
Enable RBAC enforcement by setting rbac.enabled=true in your config.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}

		slog.Info("Creating RBAC tables...")
		if err := database.MigrateRBAC(db); err != nil {
			slog.Error("RBAC table migration failed", "error", err)
			os.Exit(1)
		}

		slog.Info("Seeding built-in roles and permissions...")
		if err := rbac.Seed(db); err != nil {
			slog.Error("RBAC seed failed", "error", err)
			os.Exit(1)
		}

		slog.Info("RBAC migration and seed completed successfully.")
	},
}

var importCmd = &cobra.Command{
	Use:   "importsql",
	Short: "Import SQL file to database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sqlFile := args[0]
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Failed to connect to database for import", "error", err)
			os.Exit(1)
		}

		slog.Info("Importing SQL file...", "file", sqlFile)
		if err := utils.ImportSQL(db, sqlFile); err != nil {
			slog.Error("Failed to import SQL file", "error", err)
			os.Exit(1)
		}
		slog.Info("SQL file imported successfully.")
	},
}

func init() {
	migrateCmd.AddCommand(migrateRBACCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(importCmd)
}
