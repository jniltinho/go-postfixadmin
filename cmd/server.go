package cmd

import (
	"log/slog"
	"os"
	"strconv"

	"go-postfixadmin/internal/server"
	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

var (
	port     int
	ssl      bool
	certFile string
	keyFile  string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the administration server",
	Run: func(cmd *cobra.Command, args []string) {
		// Override with ENV if not set via flags
		if !cmd.Flags().Changed("port") {
			if envPort := os.Getenv("APP_PORT"); envPort != "" {
				if p, err := strconv.Atoi(envPort); err == nil {
					port = p
				}
			} else if envPort := os.Getenv("PORT"); envPort != "" {
				if p, err := strconv.Atoi(envPort); err == nil {
					port = p
				}
			}
		}

		if !cmd.Flags().Changed("cert") {
			if envCert := os.Getenv("SSL_CERT"); envCert != "" {
				certFile = envCert
			}
		}

		if !cmd.Flags().Changed("key") {
			if envKey := os.Getenv("SSL_KEY"); envKey != "" {
				keyFile = envKey
			}
		}

		// Auto-enable SSL if cert and key are provided via ENV
		if certFile != "" && keyFile != "" && !cmd.Flags().Changed("ssl") {
			ssl = true
		}

		// Connect to Database
		db, err := utils.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Warn("Warning: Database connection failed.", "error", err)
			db = nil
		}

		slog.Info("Starting Go-Postfixadmin...")
		server.StartServer(EmbeddedFiles, port, db, ssl, certFile, keyFile)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&port, "port", 8080, "Port to run the server on")
	serverCmd.Flags().BoolVar(&ssl, "ssl", false, "Enable SSL/TLS")
	serverCmd.Flags().StringVar(&certFile, "cert", "", "Path to SSL certificate file")
	serverCmd.Flags().StringVar(&keyFile, "key", "", "Path to SSL key file")
}
