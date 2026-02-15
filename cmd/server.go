package cmd

import (
	"log/slog"

	"go-postfixadmin/internal/server"
	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

var port int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the administration server",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to Database
		db, err := utils.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Warn("Warning: Database connection failed.", "error", err)
			db = nil
		}

		slog.Info("Starting Go-Postfixadmin...")
		server.StartServer(EmbeddedFiles, port, db)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&port, "port", 8080, "Port to run the server on")
}
