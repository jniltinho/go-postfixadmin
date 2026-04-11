package cmd

import (
	"log/slog"
	"os"
	"time"

	"go-postfixadmin/cmd/transport"
	"go-postfixadmin/internal/database"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	addTransportEntry     string
	deleteTransportDomain string
	listTransportsFlag    bool
	transportDebugFlag    bool
)

var transportCmd = &cobra.Command{
	Use:   "transport",
	Short: "Transport management utilities",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Database connection failed", "error", err)
			os.Exit(1)
		}

		if listTransportsFlag {
			transport.ListAllTransports(db)
		} else if addTransportEntry != "" {
			transport.AddTransport(db, addTransportEntry)
		} else if deleteTransportDomain != "" {
			transport.DeleteTransport(db, deleteTransportDomain)
		} else {
			cmd.Help()
		}
	},
}

var transportServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Postfix transport map TCP server",
	Run: func(cmd *cobra.Command, args []string) {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04",
			NoColor:    false,
		})

		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			zlog.Fatal().Msgf("Failed to connect to database: %v", err)
		}

		address := viper.GetString("transport.host")
		if address == "" {
			address = "127.0.0.1:12221"
		}

		cacheDurStr := viper.GetString("transport.cache")
		if cacheDurStr == "" {
			cacheDurStr = "10m"
		}
		cacheDur, err := time.ParseDuration(cacheDurStr)
		if err != nil {
			zlog.Fatal().Msgf("Invalid cache duration %q: %v", cacheDurStr, err)
		}

		srv := transport.NewTCPServer(
			address,
			db,
			cacheDur,
			viper.GetString("transport.hostname"),
			viper.GetString("transport.localdelivery"),
			viper.GetString("transport.delivery"),
			debugFlag || transportDebugFlag,
		)
		if err := srv.Start(); err != nil {
			zlog.Fatal().Msgf("Transport server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(transportCmd)
	transportCmd.AddCommand(transportServerCmd)
	transportServerCmd.Flags().BoolVar(&transportDebugFlag, "debug", false, "Enable debug logging for the transport server")
	transportCmd.Flags().BoolVarP(&listTransportsFlag, "list", "l", false, "List all transport entries")
	transportCmd.Flags().StringVarP(&addTransportEntry, "add", "a", "", "Add a transport entry (format: domain:transport, e.g. example.com:smtp:[relay.example.com]:25)")
	transportCmd.Flags().StringVarP(&deleteTransportDomain, "delete", "d", "", "Delete a transport entry by domain")
}
