package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/cmd/mailbox"
	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

var (
	addMailboxUser    string
	listMailboxesFlag bool
	importCSVFile     string
	mailboxQuotaMB    int64
)

var mailboxCmd = &cobra.Command{
	Use:   "mailbox",
	Short: "Mailbox management utilities",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := utils.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Database connection failed", "error", err)
			os.Exit(1)
		}

		if addMailboxUser != "" {
			mailbox.AddUser(db, addMailboxUser, mailboxQuotaMB)
		} else if listMailboxesFlag {
			mailbox.ListAllMailboxes(db)
		} else if importCSVFile != "" {
			mailbox.ImportCSV(db, importCSVFile, mailboxQuotaMB)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(mailboxCmd)
	mailboxCmd.Flags().StringVarP(&addMailboxUser, "add", "a", "", "Add a new mailbox user (format: email:password)")
	mailboxCmd.Flags().BoolVarP(&listMailboxesFlag, "list", "l", false, "List all mailboxes")
	mailboxCmd.Flags().StringVar(&importCSVFile, "import-csv", "", "Import mailbox users from a CSV file (columns: user,password,domain,name)")
	mailboxCmd.Flags().Int64VarP(&mailboxQuotaMB, "quota", "q", 100, "Mailbox quota in MB")
}
