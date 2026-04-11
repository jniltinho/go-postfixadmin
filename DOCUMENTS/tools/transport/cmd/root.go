package cmd

import (
	"os"
	"path/filepath"
	"time"

	"transport/internal/config"
	"transport/internal/database"
	"transport/internal/server"
	"transport/version"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	configPath    string
	address       string
	cacheDuration time.Duration
	debug         bool
)

var rootCmd = &cobra.Command{
	Use:   "transport",
	Short: "Postfix TCP Transport Service in Go",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig(configPath)

		db, err := database.Connect(cfg)
		if err != nil {
			zlog.Fatal().Msgf("Failed to connect to database: %v", err)
		}

		finalAddress := address
		if !cmd.Flags().Changed("address") && cfg.Host != "" {
			finalAddress = cfg.Host
		}

		finalCache := cacheDuration
		if !cmd.Flags().Changed("cache") && cfg.Cache != 0 {
			finalCache = cfg.Cache
		}

		tcpServer := server.New(finalAddress, db, finalCache, debug)

		if err := tcpServer.Start(); err != nil {
			zlog.Fatal().Msgf("Failed to start TCP server: %v", err)
		}
	},
}

func init() {
	// Customizes log output for desired format
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04",
		NoColor:    false,
	})

	rootCmd.Flags().StringVarP(&address, "address", "a", "127.0.0.1:12221", "Address to listen on")
	rootCmd.Flags().DurationVarP(&cacheDuration, "cache", "c", 10*time.Minute, "Cache expiration time")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug Line")
	rootCmd.Flags().StringVar(&configPath, "config", "config.toml", "Path to TOML configuration file")
}

func showVersion() {
	pName := cases.Title(language.English, cases.NoLower).String(filepath.Base(os.Args[0]))
	color.Green("%s Version:[%s] Built:[%s] Commit:[%s]\n", pName, version.VERSION, version.BUILD_DATE, version.GIT_COMMIT)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "-v" || os.Args[1] == "--version") {
		showVersion()
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
