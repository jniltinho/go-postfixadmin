package mailbox

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AddUser adds a new mailbox user to the database
func AddUser(db *gorm.DB, input string, quotaMB int64) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	var username, password string
	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		username = parts[0]
		password = parts[1]
	} else {
		username = input
	}

	if username == "" || !strings.Contains(username, "@") {
		slog.Error("Username must be in email format")
		os.Exit(1)
	}

	parts := strings.SplitN(username, "@", 2)
	localPart := parts[0]
	domain := parts[1]

	// Check if mailbox already exists
	var existing models.Mailbox
	if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
		slog.Error("Mailbox already exists", "username", username)
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

	crypted, err := utils.HashPassword(password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		os.Exit(1)
	}

	now := time.Now()
	mb := models.Mailbox{
		Username:      username,
		Password:      crypted,
		Name:          cases.Title(language.English).String(localPart),
		Maildir:       domain + "/" + localPart + "/",
		LocalPart:     localPart,
		Domain:        domain,
		Created:       now,
		Modified:      now,
		Quota:         quotaMB * utils.GetQuotaMultiplier(),
		Active:        true,
		SMTPActive:    true,
		TokenValidity: now.Add(3 * time.Hour),
	}

	if err := repositories.CreateMailbox(db, mb, "CLI", "127.0.0.1"); err != nil {
		slog.Error("Failed to create mailbox", "error", err)
		os.Exit(1)
	}

	fmt.Printf("User '%s' created successfully.\n", username)
}
