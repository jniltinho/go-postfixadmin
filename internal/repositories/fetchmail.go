package repositories

import (
	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

func CreateFetchmail(db *gorm.DB, fetchmail models.Fetchmail) error {
	return db.Create(&fetchmail).Error
}
