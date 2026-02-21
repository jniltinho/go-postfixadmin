package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "postfixadmin",
		Short: "Go-Postfixadmin CLI",
		Long:  `A command line interface for Go-Postfixadmin application.`,
	}

	// Global flags
	cfgFile  string
	dbUrl    string
	dbDriver string

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
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.toml)")
	rootCmd.PersistentFlags().StringVar(&dbUrl, "db-url", "", "Database URL connection string")
	rootCmd.PersistentFlags().StringVar(&dbDriver, "db-driver", "", "Database driver (mysql or postgres)")

	viper.BindPFlag("database.url", rootCmd.PersistentFlags().Lookup("db-url"))
	viper.BindPFlag("database.driver", rootCmd.PersistentFlags().Lookup("db-driver"))
}
