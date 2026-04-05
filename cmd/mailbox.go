package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/cmd/mailbox"
	"go-postfixadmin/internal/database"

	"github.com/spf13/cobra"
)

var (
	addMailboxUser      string
	listMailboxesFlag   bool
	importCSVFile       string
	exportCSVFile       string
	mailboxQuotaMB      int64
	importPasswordCrypt bool
)

var mailboxCmd = &cobra.Command{
	Use:   "mailbox",
	Short: "Mailbox management utilities",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Database connection failed", "error", err)
			os.Exit(1)
		}

		if addMailboxUser != "" {
			mailbox.AddUser(db, addMailboxUser, mailboxQuotaMB)
		} else if listMailboxesFlag {
			mailbox.ListAllMailboxes(db)
		} else if importCSVFile != "" {
			mailbox.ImportCSV(db, importCSVFile, mailboxQuotaMB, importPasswordCrypt)
		} else if exportCSVFile != "" {
			mailbox.ExportCSV(db, exportCSVFile)
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
	mailboxCmd.Flags().BoolVar(&importPasswordCrypt, "password-crypt", false, "Treat the password column in the CSV as an already-hashed value (skip hashing and length validation)")
	mailboxCmd.Flags().StringVarP(&exportCSVFile, "export", "e", "", "Export all mailboxes to a CSV file (columns: user,password,domain,name,quota_mb,active)")
}
