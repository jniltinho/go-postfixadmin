package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migration",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := utils.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Failed to connect to database for migration", "error", err)
			os.Exit(1)
		}

		slog.Info("Running database migration...")
		if err := utils.MigrateDB(db); err != nil {
			slog.Error("Database migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migration completed successfully.")
	},
}

var importCmd = &cobra.Command{
	Use:   "importsql",
	Short: "Import SQL file to database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sqlFile := args[0]
		db, err := utils.ConnectDB(dbUrl, dbDriver)
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
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(importCmd)
}
