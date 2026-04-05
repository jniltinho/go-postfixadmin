package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/cmd/admin"
	"go-postfixadmin/internal/database"

	"github.com/spf13/cobra"
)

var (
	listDomains      bool
	listAdmins       bool
	listAliases      bool
	listAliasDomains bool
	listDomainAdmins bool
	listLogs         bool
	listMaillog      bool
	maillogDomain    string
	maillogLimit     int
	addSuperAdmin    string
	cleanupMaildirs  bool
	quotaReport      bool
	reportEmail      string
	baseDir          string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin management utilities",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to Database
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Database connection failed", "error", err)
			os.Exit(1)
		}

		if listDomains {
			admin.ListAllDomains(db)
		} else if listAdmins {
			admin.ListAllAdmins(db)
		} else if listAliases {
			admin.ListAllAliases(db)
		} else if listAliasDomains {
			admin.ListAllAliasDomains(db)
		} else if listDomainAdmins {
			admin.ListDomainAdmins(db)
		} else if listLogs {
			admin.ListLogs(db)
		} else if listMaillog {
			admin.ListMaillog(db, maillogDomain, maillogLimit)
		} else if cleanupMaildirs {
			admin.CleanupMaildirs(db, baseDir)
		} else if quotaReport {
			admin.ShowQuotaReport(reportEmail)
		} else if addSuperAdmin != "" {
			admin.AddSuperAdmin(db, addSuperAdmin)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.Flags().BoolVarP(&listDomains, "list-domains", "d", false, "List all domains")
	adminCmd.Flags().BoolVarP(&listAdmins, "list-admins", "a", false, "List all admins")
	adminCmd.Flags().BoolVarP(&listAliases, "list-aliases", "s", false, "List all aliases")
	adminCmd.Flags().BoolVarP(&listAliasDomains, "list-alias-domains", "S", false, "List all alias domains")
	adminCmd.Flags().BoolVarP(&listDomainAdmins, "domain-admins", "A", false, "List all domain admins")
	adminCmd.Flags().BoolVarP(&listLogs, "list-logs", "L", false, "List all system logs")
	adminCmd.Flags().BoolVar(&listMaillog, "list-maillog", false, "List mail filter log entries")
	adminCmd.Flags().StringVar(&maillogDomain, "maillog-domain", "", "Filter maillog by domain")
	adminCmd.Flags().IntVar(&maillogLimit, "maillog-limit", 100, "Number of maillog entries to show")
	adminCmd.Flags().BoolVarP(&cleanupMaildirs, "cleanup-maildir", "c", false, "Clean up orphaned maildirs on the server")
	adminCmd.Flags().BoolVarP(&quotaReport, "quota-report", "q", false, "Show Dovecot quota report")
	adminCmd.Flags().StringVarP(&reportEmail, "email", "e", "", "Send quota report via sendmail to this address")
	adminCmd.Flags().StringVar(&addSuperAdmin, "add-superadmin", "", "Add a new superadmin (format: email:password)")
	adminCmd.Flags().StringVar(&baseDir, "base-dir", "/var/vmail", "Base directory for maildirs")
}
