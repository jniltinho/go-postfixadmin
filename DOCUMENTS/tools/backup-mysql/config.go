package main

import (
	"os"
	"strings"
)

// config holds all runtime configuration.
// Priority: CLI flag > environment variable > built-in default.
type config struct {
	MySQLHost  string
	MySQLUser  string
	MySQLPass  string
	BackupDir  string
	LogFile    string
	SMTPServer string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	EmailFrom  string
	EmailTo    []string // comma-separated in env: EMAIL_TO
	EmailCC    []string // comma-separated in env: EMAIL_CC
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func loadConfig() config {
	return config{
		MySQLHost:  envOr("MYSQL_HOST", "localhost"),
		MySQLUser:  envOr("MYSQL_USER", "root"),
		MySQLPass:  envOr("MYSQL_PASS", "root"),
		BackupDir:  envOr("BACKUP_DIR", "/usr/local/backup/mysql"),
		LogFile:    envOr("LOG_FILE", "/var/log/backup_mysql.log"),
		SMTPServer: envOr("SMTP_SERVER", "smtp.dominio.com.br"),
		SMTPPort:   envOr("SMTP_PORT", "587"),
		SMTPUser:   envOr("SMTP_USER", "email@dominio.com.br"),
		SMTPPass:   envOr("SMTP_PASS", "senha_email"),
		EmailFrom:  envOr("EMAIL_FROM", "email@dominio.com.br"),
		EmailTo:    envList("EMAIL_TO", []string{"email1@dominio_x.com.br"}),
		EmailCC:    envList("EMAIL_CC", []string{"admin@dominio_y.com.br"}),
	}
}
