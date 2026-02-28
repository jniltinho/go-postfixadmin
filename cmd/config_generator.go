package cmd

import (
	"fmt"
	"os"
	"time"
)

func generateConfig() {
	configContent := `# Go-Postfixadmin Configuration File

[database]
# Format: user:password@tcp(host:port)/dbname?args | driver = "mysql" # mysql or postgres (default: mysql)
url = "postfix:postfixPassword@tcp(localhost:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"

[server]
# Server Port (default 8080)
port = 8080
clean_up_maildir = false # Clean up orphaned maildirs when deleting a mailbox

[ssl]
#enabled = false
#cert = "ssl/server.crt"
#key = "ssl/server.key"
# If both cert and key are provided, SSL will be enabled automatically
# Session Secret Key
# Generate a random 64-character hex string for production use, E.g., using "openssl rand -hex 32"
#session_secret = "9a048f79e88e35de37dc2c43c1fc002f358f92957a7690e60109cfe8a65178e0"

[quota]
enabled      = false
domain_quota = true
multiplier   = 1024000 # Bytes per MB: 1024000 or 1048576

[vacation]
enabled = true

[alias]
edit_alias          = true
alias_control       = true
alias_control_admin = true
special_alias_control = false
alias_domain        = true

[transport]
enabled  = true
options  = ["virtual", "local", "relay"]
default  = "virtual"

[features]
fetchmail = false

[smtp]
server  = "localhost"
port    = 25
subject = "Welcome!"
body    = "Hi,\n\nWelcome to your new account."
type    = "plain" # type: plain | tls | starttls
`

	fileName := fmt.Sprintf("config_%s.toml", time.Now().Format("2006-01-02_150405"))
	err := os.WriteFile(fileName, []byte(configContent), 0644)
	if err != nil {
		fmt.Printf("Error writing %s: %v\n", fileName, err)
		os.Exit(1)
	} else {
		fmt.Printf("Successfully generated %s in the current directory.\n", fileName)
	}
}
