package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd(cfg *config) *cobra.Command {
	root := &cobra.Command{
		Use:           "backup-mysql",
		Short:         "MySQL backup tool",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Persistent flags — available to all subcommands
	pf := root.PersistentFlags()
	pf.StringVar(&cfg.MySQLHost, "host", cfg.MySQLHost, "MySQL host address")
	pf.StringVar(&cfg.MySQLUser, "user", cfg.MySQLUser, "MySQL username")
	pf.StringVar(&cfg.MySQLPass, "passwd", cfg.MySQLPass, "MySQL password")

	root.AddCommand(
		newBackupCmd(cfg),
		newListCmd(cfg),
		newVersionCmd(),
	)
	return root
}

func newBackupCmd(cfg *config) *cobra.Command {
	var (
		clean      int
		debug      bool
		doSendmail bool
	)
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup all MySQL databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			initLog(debug)
			logMsg("[*** MYSQL BACKUP JOB ***]")
			execBackup(*cfg)
			if clean > 0 {
				cleanFiles(cfg.BackupDir, clean)
			}
			if doSendmail {
				sendMail(*cfg, "MYSQL BACKUP LOG")
			}
			saveLog(cfg.LogFile)
			return nil
		},
	}
	cmd.Flags().IntVar(&clean, "clean", 0, "Remove backup files older than X days (0 = disabled)")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug output")
	cmd.Flags().BoolVar(&doSendmail, "sendmail", false, "Send log by e-mail")
	return cmd
}

func newListCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all MySQL databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			listDatabases(*cfg)
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("backup-mysql version %s (Build Date: %s)\n", version, buildDate)
		},
	}
}
