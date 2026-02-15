package utils

import (
	"fmt"
	"time"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// LogAction logs an administrative action to the database
func LogAction(db *gorm.DB, username, ip, domain, action, data string) error {
	logEntry := models.Log{
		Timestamp: time.Now(),
		Username:  fmt.Sprintf("%s (%s)", username, ip),
		Domain:    domain,
		Action:    action,
		Data:      data,
	}

	return db.Create(&logEntry).Error
}
