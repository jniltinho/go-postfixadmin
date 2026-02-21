package cmd

import (
	"fmt"
	"os"
)

func generateConfig() {
	configContent := `# Go-Postfixadmin Configuration File

[database]
# Format: user:password@tcp(host:port)/dbname?args
url = "postfix:postfixPassword@tcp(localhost:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
# driver = "mysql" # mysql or postgres (default: mysql)

[server]
# Server Port (default 8080)
port = 8080

# Session Secret Key
# Generate a random 64-character hex string for production use
# E.g., using openssl rand -hex 32
# session_secret = "9a048f79e88e35de37dc2c43c1fc002f358f92957a7690e60109cfe8a65178e0"

[ssl]
# If both cert and key are provided, SSL will be enabled automatically
# enabled = false
# cert = "ssl/server.crt"
# key = "ssl/server.key"
`
	err := os.WriteFile("config.toml", []byte(configContent), 0644)
	if err != nil {
		fmt.Println("Error writing config.toml:", err)
		os.Exit(1)
	} else {
		fmt.Println("Successfully generated config.toml in the current directory.")
	}
}
