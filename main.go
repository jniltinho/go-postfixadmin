package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"go-postfixadmin/internal/utils"
)

const Version = "1.0.0"

//go:embed views public
var embeddedFiles embed.FS

func main() {
	// CLI Flags
	versionFlag := flag.Bool("version", false, "Display version information")
	runFlag := flag.Bool("run", false, "Start the administration server")
	migrateFlag := flag.Bool("migrate", false, "Run database migration")
	importSQLFlag := flag.String("import-sql", "", "Import SQL file to database")
	portFlag := flag.Int("port", 8080, "Port to run the server on")
	dbUrl := flag.String("db-url", "", "Database URL connection string")
	dbDriver := flag.String("db-driver", "mysql", "Database driver (mysql or postgres)")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Go-Postfixadmin version %s\n", Version)
		return
	}

	// Connect to Database
	db, err := utils.ConnectDB(*dbUrl, *dbDriver)
	if err != nil {
		if *migrateFlag || *importSQLFlag != "" {
			slog.Error("Failed to connect to database for operation", "error", err)
			os.Exit(1)
		}
		slog.Warn("Warning: Database connection failed.", "error", err)
		db = nil
	}

	if *migrateFlag {
		slog.Info("Running database migration...")
		if err := utils.MigrateDB(db); err != nil {
			slog.Error("Database migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migration completed successfully.")
	}

	if *importSQLFlag != "" {
		slog.Info("Importing SQL file...", "file", *importSQLFlag)
		if err := utils.ImportSQL(db, *importSQLFlag); err != nil {
			slog.Error("Failed to import SQL file", "error", err)
			os.Exit(1)
		}
		slog.Info("SQL file imported successfully.")
	}

	if !*runFlag {
		if !*migrateFlag && *importSQLFlag == "" {
			flag.Usage()
		}
		return
	}

	slog.Info("Starting Go-Postfixadmin...", "version", Version)
	StartServer(embeddedFiles, *portFlag, db)
}
