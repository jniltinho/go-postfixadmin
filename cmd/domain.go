package cmd

import (
	"log/slog"
	"os"

	"go-postfixadmin/cmd/domain"
	"go-postfixadmin/internal/database"

	"github.com/spf13/cobra"
)

var (
	listDomainsFlag    bool
	addDomainName      string
	deleteDomainName   string
	domainDescription  string
	domainMaxAliases   int
	domainMaxMailboxes int
)

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Domain management utilities",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("Database connection failed", "error", err)
			os.Exit(1)
		}

		if listDomainsFlag {
			domain.ListAllDomains(db)
		} else if addDomainName != "" {
			domain.AddDomain(db, addDomainName, domainDescription, domainMaxAliases, domainMaxMailboxes)
		} else if deleteDomainName != "" {
			domain.DeleteDomain(db, deleteDomainName)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(domainCmd)
	domainCmd.Flags().BoolVarP(&listDomainsFlag, "list", "l", false, "List all domains")
	domainCmd.Flags().StringVarP(&addDomainName, "add", "a", "", "Add a new domain")
	domainCmd.Flags().StringVarP(&deleteDomainName, "delete", "d", "", "Delete a domain and all its data")
	domainCmd.Flags().StringVar(&domainDescription, "description", "", "Domain description")
	domainCmd.Flags().IntVar(&domainMaxAliases, "max-aliases", 10, "Maximum number of aliases (0 = unlimited)")
	domainCmd.Flags().IntVar(&domainMaxMailboxes, "max-mailboxes", 10, "Maximum number of mailboxes (0 = unlimited)")
}
