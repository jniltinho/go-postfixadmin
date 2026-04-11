package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/cmd/transport"
	"go-postfixadmin/internal/database"

	"github.com/spf13/cobra"
)

var (
	addTransportEntry     string
	deleteTransportDomain string
	listTransportsFlag    bool
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

func init() {
	rootCmd.AddCommand(transportCmd)
	transportCmd.Flags().BoolVarP(&listTransportsFlag, "list", "l", false, "List all transport entries")
	transportCmd.Flags().StringVarP(&addTransportEntry, "add", "a", "", "Add a transport entry (format: domain:transport, e.g. example.com:smtp:[relay.example.com]:25)")
	transportCmd.Flags().StringVarP(&deleteTransportDomain, "delete", "d", "", "Delete a transport entry by domain")
}
