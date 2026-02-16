package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/admin"
	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

var (
	listDomains      bool
	listMailboxes    bool
	listAdmins       bool
	listAliases      bool
	listDomainAdmins bool
	addSuperAdmin    string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin management utilities",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to Database
		db, err := utils.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Database connection failed", "error", err)
			os.Exit(1)
		}

		if listDomains {
			admin.ListAllDomains(db)
		} else if listMailboxes {
			admin.ListAllMailboxes(db)
		} else if listAdmins {
			admin.ListAllAdmins(db)
		} else if listAliases {
			admin.ListAllAliases(db)
		} else if listDomainAdmins {
			admin.ListDomainAdmins(db)
		} else if addSuperAdmin != "" {
			admin.AddSuperAdmin(db, addSuperAdmin)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.Flags().BoolVarP(&listDomains, "list-domains", "l", false, "List all domains")
	adminCmd.Flags().BoolVarP(&listMailboxes, "list-mailboxes", "m", false, "List all mailboxes")
	adminCmd.Flags().BoolVarP(&listAdmins, "list-admins", "a", false, "List all admins")
	adminCmd.Flags().BoolVarP(&listAliases, "list-aliases", "s", false, "List all aliases")
	adminCmd.Flags().BoolVarP(&listDomainAdmins, "domain-admins", "d", false, "List all domain admins")
	adminCmd.Flags().StringVar(&addSuperAdmin, "add-superadmin", "", "Add a new superadmin (format: email:password)")
}
