package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"time"
)

func generateConfig() {
	// Read the default config from the embedded file (single source of truth)
	configBytes, err := fs.ReadFile(EmbeddedFiles, "web/files/config.default.toml")
	if err != nil {
		fmt.Printf("Error reading embedded default config: %v\n", err)
		fmt.Println("Falling back to minimal config...")

		// Fallback minimal config if embed fails
		configBytes = []byte(`# Go-Postfixadmin Configuration File (fallback)
[database]
host   = "localhost"
port   = "3306"
user   = "postfix"
pass   = "postfixPassword"
name   = "postfix"
driver = "mysql"
debug  = false

[server]
port = 8080
ssl_enable = false
session_secret = "change-me-in-production"
swagger_enable = true
jwt_access_ttl  = "15m"
jwt_refresh_ttl = "168h"
`)
	}

	fileName := fmt.Sprintf("config_%s.toml", time.Now().Format("2006-01-02_150405"))
	err = os.WriteFile(fileName, configBytes, 0644)
	if err != nil {
		fmt.Printf("Error writing %s: %v\n", fileName, err)
		os.Exit(1)
	} else {
		fmt.Printf("Successfully generated %s in the current directory.\n", fileName)
		fmt.Println("This file was generated from the embedded default config (web/files/config.default.toml).")
	}
}
