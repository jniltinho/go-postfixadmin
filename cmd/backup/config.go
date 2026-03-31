package backup

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the backup-mysql command.
// Priority: CLI flag > config.toml [backup] section > environment variable > built-in default.
type Config struct {
	MySQLHost  string
	MySQLPort  string
	MySQLUser  string
	MySQLPass  string
	BackupDir  string
	LogFile    string
	SMTPServer string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	EmailFrom  string
	EmailTo    []string
	EmailCC    []string
}

func viperOr(key, envKey, fallback string) string {
	if v := viper.GetString(key); v != "" {
		return v
	}
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return fallback
}

func viperOrList(key, envKey string, fallback []string) []string {
	if v := viper.GetStringSlice(key); len(v) > 0 {
		return v
	}
	if v := os.Getenv(envKey); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return fallback
}

// LoadConfig loads configuration from viper (config.toml [backup] section) and
// environment variables. CLI flags override the returned values after this call.
func LoadConfig() Config {
	return Config{
		MySQLHost:  viperOr("backup.mysql_host", "MYSQL_HOST", "localhost"),
		MySQLPort:  viperOr("backup.mysql_port", "MYSQL_PORT", "3306"),
		MySQLUser:  viperOr("backup.mysql_user", "MYSQL_USER", "root"),
		MySQLPass:  viperOr("backup.mysql_pass", "MYSQL_PASS", ""),
		BackupDir:  viperOr("backup.backup_dir", "BACKUP_DIR", "/usr/local/backup/mysql"),
		LogFile:    viperOr("backup.log_file", "LOG_FILE", "/var/log/backup_mysql.log"),
		SMTPServer: viperOr("backup.smtp_server", "SMTP_SERVER", ""),
		SMTPPort:   viperOr("backup.smtp_port", "SMTP_PORT", "587"),
		SMTPUser:   viperOr("backup.smtp_user", "SMTP_USER", ""),
		SMTPPass:   viperOr("backup.smtp_pass", "SMTP_PASS", ""),
		EmailFrom:  viperOr("backup.email_from", "EMAIL_FROM", ""),
		EmailTo:    viperOrList("backup.email_to", "EMAIL_TO", nil),
		EmailCC:    viperOrList("backup.email_cc", "EMAIL_CC", nil),
	}
}
