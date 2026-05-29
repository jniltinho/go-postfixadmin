package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/internal/database"
	"go-postfixadmin/internal/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		// Override with Viper config if not set via flags
		if !cmd.Flags().Changed("port") {
			if vPort := viper.GetInt("server.port"); vPort != 0 {
				port = vPort
			} else if vPort := viper.GetInt("port"); vPort != 0 {
				port = vPort
			}
		}

		if !cmd.Flags().Changed("cert") && viper.GetString("server.ssl_cert") != "" {
			certFile = viper.GetString("server.ssl_cert")
		}

		if !cmd.Flags().Changed("key") && viper.GetString("server.ssl_key") != "" {
			keyFile = viper.GetString("server.ssl_key")
		}

		// Auto-enable SSL if cert and key are provided via config or flags
		if certFile != "" && keyFile != "" && !cmd.Flags().Changed("ssl") {
			ssl = viper.GetBool("server.ssl_enable")
			if !viper.IsSet("server.ssl_enable") {
				ssl = true
			}
		}

		// Connect to Database
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Warn("Warning: Database connection failed.", "error", err)
			db = nil
		}

		slog.Info("Starting Go-Postfixadmin...")
		server.AppVersion = Version

		// JWT foundations (PR 01 of the Vue/Quasar + JWT migration plan) are loaded
		// automatically via viper from [server] jwt_* keys. Full middleware + auth
		// endpoints come in subsequent PRs. See internal/auth/jwt.go and
		// DOCUMENTS/MIGRATION_PLAN_VUE3_QUASAR_JWT.md.
		if err := server.StartServer(EmbeddedFiles, port, db, ssl, certFile, keyFile); err != nil {
				slog.Error("server exited with error", "error", err)
				os.Exit(1)
			}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&port, "port", 8080, "Port to run the server on")
	serverCmd.Flags().BoolVar(&ssl, "ssl", false, "Enable SSL/TLS")
	serverCmd.Flags().StringVar(&certFile, "cert", "", "Path to SSL certificate file")
	serverCmd.Flags().StringVar(&keyFile, "key", "", "Path to SSL key file")
}
