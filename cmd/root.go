package cmd

import (
	"embed"
	"fmt"
	"log/slog"
	"os"

	"go-postfixadmin/internal/database"
	"go-postfixadmin/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "postfixadmin",
		Short: "Go-Postfixadmin CLI",
		Long:  `A command line interface for Go-Postfixadmin application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if generateConfigFlag {
				generateConfig()
			} else if vacationSync {
				db, err := database.ConnectDB(dbUrl, dbDriver)
				if err != nil {
					slog.Error("database connection failed", "error", err)
					os.Exit(1)
				}
				if err := utils.SyncVacationSieve(db, ""); err != nil {
					slog.Error("vacation sync failed", "error", err)
					os.Exit(1)
				}
			} else {
				cmd.Help()
			}
		},
	}

	// Global flags
	cfgFile            string
	dbUrl              string
	dbDriver           string
	generateConfigFlag bool
	vacationSync       bool
	debugFlag          bool

	// Shared resources
	EmbeddedFiles embed.FS
)

func Execute(files embed.FS) {
	EmbeddedFiles = files
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/postfixadmin")
		viper.AddConfigPath("$HOME/.postfixadmin")
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// Successfully read config
	}

	// Setup logger level after config is parsed
	level := slog.LevelInfo
	if viper.GetBool("debug") {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolVar(&generateConfigFlag, "generate-config", false, "Generate a default config.toml file in the current directory")
	rootCmd.Flags().BoolVar(&vacationSync, "vacation", false, "Sync vacation auto-replies to Dovecot Sieve scripts (for crontab)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.toml)")
	rootCmd.PersistentFlags().StringVar(&dbUrl, "db-url", "", "Database URL connection string")
	rootCmd.PersistentFlags().StringVar(&dbDriver, "db-driver", "", "Database driver (mysql or postgres)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")

	viper.BindPFlag("database.url", rootCmd.PersistentFlags().Lookup("db-url"))
	viper.BindPFlag("database.driver", rootCmd.PersistentFlags().Lookup("db-driver"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}
