package repositories

import (
	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

func IsSuperAdmin(db *gorm.DB, username string) (bool, error) {
	var admin models.Admin
	if err := db.Select("superadmin").Where("username = ?", username).First(&admin).Error; err != nil {
		return false, err
	}
	return admin.Superadmin, nil
}
