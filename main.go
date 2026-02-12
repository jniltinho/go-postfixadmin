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
	portFlag := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Go-Postfixadmin version %s\n", Version)
		return
	}

	// Connect to Database
	db, err := utils.ConnectDB()
	if err != nil {
		if *migrateFlag {
			slog.Error("Failed to connect to database for migration", "error", err)
			os.Exit(1)
		}
		slog.Warn("Warning: Database connection failed.", "error", err)
	}

	if *migrateFlag {
		slog.Info("Running database migration...")
		if err := utils.MigrateDB(db); err != nil {
			slog.Error("Database migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migration completed successfully.")
	}

	if !*runFlag {
		if !*migrateFlag {
			flag.Usage()
		}
		return
	}

	slog.Info("Starting Go-Postfixadmin...", "version", Version)
	StartServer(embeddedFiles, *portFlag, db)
}
