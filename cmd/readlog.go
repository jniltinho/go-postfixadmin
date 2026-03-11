package cmd

import (
	"log/slog"
	"os"
	"time"

	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
)

const (
	defaultLogFile  = "/var/log/mail.log"
	defaultPosFile  = "/opt/go-postfixadmin/.maillog.pos"
	defaultInterval = 5 * time.Minute
)

var readlogCmd = &cobra.Command{
	Use:   "readlog",
	Short: "Read mail.log and import FILTER entries into the maillog table",
	Long: `Reads /var/log/mail.log periodically, extracts postfix/filter FILTER lines
and persists them into the maillog database table.

Uses a position file to track the last read offset, so only new lines are
processed on each run. Detects log rotation automatically.

Examples:
  # Run as daemon (every 5 minutes):
  ./postfixadmin readlog

  # Run once (for crontab use):
  ./postfixadmin readlog --once

  # Custom interval and log path:
  ./postfixadmin readlog --interval 1m --logfile /var/log/mail.log`,
	Run: func(cmd *cobra.Command, args []string) {
		logFile, _ := cmd.Flags().GetString("logfile")
		posFile, _ := cmd.Flags().GetString("posfile")
		interval, _ := cmd.Flags().GetDuration("interval")
		once, _ := cmd.Flags().GetBool("once")

		db, err := utils.ConnectDB(dbUrl, dbDriver)
		if err != nil {
			slog.Error("readlog: database connection failed", "error", err)
			os.Exit(1)
		}

		slog.Info("readlog started", "logfile", logFile, "posfile", posFile, "interval", interval, "once", once)

		if err := utils.ReadMailLog(db, logFile, posFile); err != nil {
			slog.Error("readlog: error processing log", "error", err)
		}

		if once {
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := utils.ReadMailLog(db, logFile, posFile); err != nil {
				slog.Error("readlog: error processing log", "error", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(readlogCmd)
	readlogCmd.Flags().String("logfile", defaultLogFile, "Path to the mail log file")
	readlogCmd.Flags().String("posfile", defaultPosFile, "Path to the position state file")
	readlogCmd.Flags().Duration("interval", defaultInterval, "Polling interval (e.g. 5m, 30s)")
	readlogCmd.Flags().Bool("once", false, "Run once and exit (for crontab use)")
}
