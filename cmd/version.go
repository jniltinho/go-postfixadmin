package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "1.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Go-Postfixadmin version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
