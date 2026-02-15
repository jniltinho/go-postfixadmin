package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "postfixadmin",
		Short: "Go-Postfixadmin CLI",
		Long:  `A command line interface for Go-Postfixadmin application.`,
	}

	// Global flags
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

func init() {
	rootCmd.PersistentFlags().StringVar(&dbUrl, "db-url", "", "Database URL connection string")
	rootCmd.PersistentFlags().StringVar(&dbDriver, "db-driver", "mysql", "Database driver (mysql or postgres)")
}
