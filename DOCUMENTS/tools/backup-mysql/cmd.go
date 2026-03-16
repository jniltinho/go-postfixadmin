package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newRootCmd(cfg *config) *cobra.Command {
	var envFile string

	root := &cobra.Command{
		Use:           "backup-mysql",
		Short:         "MySQL backup tool",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			*cfg = loadConfig(envFile)
			// Re-apply any flags explicitly set by the user
			cmd.Flags().Visit(func(f *pflag.Flag) {
				switch f.Name {
				case "host":
					cfg.MySQLHost = f.Value.String()
				case "user":
					cfg.MySQLUser = f.Value.String()
				case "passwd":
					cfg.MySQLPass = f.Value.String()
				}
			})
			return nil
		},
	}

	// Persistent flags — available to all subcommands
	pf := root.PersistentFlags()
	pf.StringVar(&envFile, "env", "", "Path to .env file (default: .env in current directory)")
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
