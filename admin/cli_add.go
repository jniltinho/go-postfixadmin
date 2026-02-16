package admin

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AddSuperAdmin adds a new superadmin to the database
func AddSuperAdmin(db *gorm.DB, input string) {
	// Silence DB logger for CLI
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	var username, password string

	// Check if input is in email:password format
	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		username = parts[0]
		password = parts[1]
	} else {
		username = input
	}

	// Validate username (simple check)
	if username == "" || !strings.Contains(username, "@") {
		slog.Error("Username must be in email format")
		os.Exit(1)
	}

	// Check if admin already exists
	var existingAdmin models.Admin
	if err := db.Where("username = ?", username).First(&existingAdmin).Error; err == nil {
		slog.Error("Administrator already exists", "username", username)
		os.Exit(1)
	}

	// Generate random password if not provided
	if password == "" {
		generatedPwd, err := generateRandomPassword(10)
		if err != nil {
			slog.Error("Failed to generate random password", "error", err)
			os.Exit(1)
		}
		password = generatedPwd
		fmt.Printf("Generated Password: %s\n", password)
	}

	if len(password) < 8 {
		slog.Error("Password must be at least 8 characters long")
		os.Exit(1)
	}

	// Hash password
	crypted, err := utils.HashPassword(password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		os.Exit(1)
	}

	// Create Admin
	newAdmin := models.Admin{
		Username:      username,
		Password:      crypted,
		Created:       time.Now(),
		Modified:      time.Now(),
		Active:        true,
		Superadmin:    true,
		TokenValidity: time.Now().Add(3 * time.Hour),
	}

	tx := db.Begin()

	if err := tx.Create(&newAdmin).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to create administrator", "error", err)
		os.Exit(1)
	}

	// Assign ALL domain for superadmin
	da := models.DomainAdmin{
		Username: username,
		Domain:   "ALL",
		Created:  time.Now(),
		Active:   true,
	}
	if err := tx.Create(&da).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to assign 'ALL' domain permission", "error", err)
		os.Exit(1)
	}

	// Log Action
	// CLI doesn't have request IP, using "CLI" or localhost
	if err := utils.LogAction(tx, "CLI", "127.0.0.1", "ALL", "create_admin", username); err != nil {
		slog.Warn("Failed to log action", "error", err)
		// Don't rollback for log failure in CLI
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Superadmin '%s' created successfully.\n", username)
}
