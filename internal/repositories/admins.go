package repositories

import (
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

func IsSuperAdmin(db *gorm.DB, username string) (bool, error) {
	var admin models.Admin
	if err := db.Select("superadmin").Where("username = ?", username).First(&admin).Error; err != nil {
		return false, err
	}
	return admin.Superadmin, nil
}

func GetAllAdmins(db *gorm.DB, username string, isSuperAdmin bool) ([]models.Admin, error) {
	query := db.Model(&models.Admin{})
	if !isSuperAdmin {
		query = query.Where("username = ?", username)
	}
	var admins []models.Admin
	if err := query.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func CountAdminDomains(db *gorm.DB, username string) (int64, error) {
	var count int64
	err := db.Model(&models.DomainAdmin{}).Where("username = ?", username).Count(&count).Error
	return count, err
}

func GetAdminByUsername(db *gorm.DB, username string) (models.Admin, error) {
	var admin models.Admin
	err := db.First(&admin, "username = ?", username).Error
	return admin, err
}

func GetAdminAssignedDomains(db *gorm.DB, username string) ([]models.DomainAdmin, error) {
	var domainAdmins []models.DomainAdmin
	err := db.Where("username = ?", username).Find(&domainAdmins).Error
	return domainAdmins, err
}

func CreateAdmin(db *gorm.DB, admin models.Admin, domains []string, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		if admin.Superadmin {
			da := models.DomainAdmin{Username: admin.Username, Domain: "ALL", Created: time.Now(), Active: true}
			if err := tx.Create(&da).Error; err != nil {
				return err
			}
		} else {
			for _, d := range domains {
				da := models.DomainAdmin{Username: admin.Username, Domain: d, Created: time.Now(), Active: true}
				if err := tx.Create(&da).Error; err != nil {
					return err
				}
			}
		}
		return utils.LogAction(tx, actorUsername, ip, "ALL", "create_admin", admin.Username)
	})
}

func UpdateAdmin(db *gorm.DB, targetUsername string, updates map[string]interface{}, domains []string, isActorSuper bool, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Admin{}).Where("username = ?", targetUsername).Updates(updates).Error; err != nil {
			return err
		}
		if isActorSuper {
			if err := tx.Where("username = ?", targetUsername).Delete(&models.DomainAdmin{}).Error; err != nil {
				return err
			}
			superadmin, _ := updates["superadmin"].(bool)
			if superadmin {
				da := models.DomainAdmin{Username: targetUsername, Domain: "ALL", Created: time.Now(), Active: true}
				if err := tx.Create(&da).Error; err != nil {
					return err
				}
			} else {
				for _, d := range domains {
					da := models.DomainAdmin{Username: targetUsername, Domain: d, Created: time.Now(), Active: true}
					if err := tx.Create(&da).Error; err != nil {
						return err
					}
				}
			}
		}
		return utils.LogAction(tx, actorUsername, ip, "ALL", "edit_admin", targetUsername)
	})
}

func DeleteAdmin(db *gorm.DB, username, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", username).Delete(&models.DomainAdmin{}).Error; err != nil {
			return err
		}
		if err := tx.Where("username = ?", username).Delete(&models.Admin{}).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, "ALL", "delete_admin", username)
	})
}
