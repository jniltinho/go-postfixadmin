package repositories

import (
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

func GetVacationByEmail(db *gorm.DB, email string) (models.Vacation, error) {
	var vacation models.Vacation
	err := db.First(&vacation, "email = ?", email).Error
	return vacation, err
}

func UpsertVacation(db *gorm.DB, vacation models.Vacation, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&vacation).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, vacation.Domain, "USER_UPDATE_VACATION", vacation.Email)
	})
}

func DeleteVacation(db *gorm.DB, email, domain, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ?", email).Delete(&models.Vacation{}).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, domain, "USER_DELETE_VACATION", email)
	})
}
