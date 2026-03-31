package cmd

import (
	"go-postfixadmin/cmd/backup"

	"github.com/spf13/cobra"
)

func init() {
	var (
		cfg       backup.Config
		mysqlHost string
		mysqlPort string
		mysqlUser string
		mysqlPass string
	)

	backupMySQLCmd := &cobra.Command{
		Use:   "backup-mysql",
		Short: "MySQL backup utilities",
		Long:  "Backup and inspect MySQL databases. Config is read from config.toml [backup] section or environment variables.",
		// Reload config after viper has parsed the config file, then apply any
		// explicit CLI flag overrides (Changed == user explicitly passed the flag).
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg = backup.LoadConfig()
			if cmd.Flags().Changed("host") {
				cfg.MySQLHost = mysqlHost
			}
			if cmd.Flags().Changed("port") {
				cfg.MySQLPort = mysqlPort
			}
			if cmd.Flags().Changed("user") {
				cfg.MySQLUser = mysqlUser
			}
			if cmd.Flags().Changed("passwd") {
				cfg.MySQLPass = mysqlPass
			}
			return nil
		},
	}

	pf := backupMySQLCmd.PersistentFlags()
	pf.StringVar(&mysqlHost, "host", "", "MySQL host address (overrides config)")
	pf.StringVar(&mysqlPort, "port", "", "MySQL port (overrides config)")
	pf.StringVar(&mysqlUser, "user", "", "MySQL username (overrides config)")
	pf.StringVar(&mysqlPass, "passwd", "", "MySQL password (overrides config)")

	// --- backup subcommand ---
	var (
		cleanDays  int
		verbose    bool
		doSendmail bool
	)
	backupRunCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup all MySQL databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			backup.InitLog(verbose)
			backup.LogMsg("[*** MYSQL BACKUP JOB ***]")
			backup.ExecBackup(cfg)
			if cleanDays > 0 {
				backup.CleanFiles(cfg.BackupDir, cleanDays)
			}
			if doSendmail {
				backup.SendMail(cfg, "MYSQL BACKUP LOG")
			}
			backup.SaveLog(cfg.LogFile)
			return nil
		},
	}
	backupRunCmd.Flags().IntVar(&cleanDays, "clean", 0, "Remove backup files older than N days (0 = disabled)")
	backupRunCmd.Flags().BoolVar(&verbose, "verbose", false, "Print log output to stdout")
	backupRunCmd.Flags().BoolVar(&doSendmail, "sendmail", false, "Send log by e-mail after backup")

	// --- list subcommand ---
	backupListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all MySQL databases and their sizes",
		RunE: func(cmd *cobra.Command, args []string) error {
			backup.ListDatabases(cfg)
			return nil
		},
	}

	backupMySQLCmd.AddCommand(backupRunCmd, backupListCmd)
	rootCmd.AddCommand(backupMySQLCmd)
}
