package repositories

import (
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

func GetAllDomains(db *gorm.DB, username string, isSuperAdmin bool) ([]models.Domain, error) {
	allowedDomains, _, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	query := db.Where("domain != ?", "ALL")
	if !isSuperAdmin {
		if len(allowedDomains) == 0 {
			return nil, nil
		}
		query = query.Where("domain IN ?", allowedDomains)
	}

	var domains []models.Domain
	if err := query.Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

func GetDomainByName(db *gorm.DB, name string) (models.Domain, error) {
	var domain models.Domain
	err := db.Where("domain = ?", name).First(&domain).Error
	return domain, err
}

func CountDomainAliases(db *gorm.DB, domain string) (int64, error) {
	var count int64
	err := db.Model(&models.Alias{}).
		Where("domain = ?", domain).
		Where("address NOT IN (?)", db.Table("mailbox").Select("username")).
		Count(&count).Error
	return count, err
}

func CountDomainMailboxes(db *gorm.DB, domain string) (int64, error) {
	var count int64
	err := db.Model(&models.Mailbox{}).Where("domain = ?", domain).Count(&count).Error
	return count, err
}

func CreateDomain(db *gorm.DB, domain models.Domain, actorUsername, ip string) error {
	if err := db.Create(&domain).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, domain.Domain, "create_domain", domain.Domain)
}

func UpdateDomain(db *gorm.DB, domain models.Domain, activeChanged bool, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if activeChanged {
			if err := tx.Model(&models.Mailbox{}).Where("domain = ?", domain.Domain).Update("active", domain.Active).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Alias{}).Where("domain = ?", domain.Domain).Update("active", domain.Active).Error; err != nil {
				return err
			}
		}
		domain.Modified = time.Now()
		if err := tx.Save(&domain).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, domain.Domain, "edit_domain", domain.Domain)
	})
}


func GetActiveDomains(db *gorm.DB, username string, isSuperAdmin bool) (domains []models.Domain, isSuper bool, err error) {
	allowedDomains, isSuper, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, false, err
	}

	query := db.Where("domain != ? AND active = ?", "ALL", true).Order("domain ASC")

	if !isSuper {
		if len(allowedDomains) == 0 {
			return nil, false, nil
		}
		query = query.Where("domain IN ?", allowedDomains)
	}

	if err := query.Find(&domains).Error; err != nil {
		return nil, isSuper, err
	}

	return domains, isSuper, nil
}

func DeleteDomain(db *gorm.DB, domainName, username, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("alias_domain = ?", domainName).Delete(&models.AliasDomain{}).Error; err != nil {
			return err
		}

		if err := tx.Where("target_domain = ?", domainName).Delete(&models.AliasDomain{}).Error; err != nil {
			return err
		}

		if err := tx.Where("domain = ?", domainName).Delete(&models.Alias{}).Error; err != nil {
			return err
		}

		if err := tx.Where("domain = ?", domainName).Delete(&models.Mailbox{}).Error; err != nil {
			return err
		}

		if err := tx.Where("domain = ?", domainName).Delete(&models.DomainAdmin{}).Error; err != nil {
			return err
		}

		if err := tx.Where("domain = ?", domainName).Delete(&models.Fetchmail{}).Error; err != nil {
			return err
		}

		if err := tx.Where("domain = ?", domainName).Delete(&models.Vacation{}).Error; err != nil {
			return err
		}

		if err := tx.Where("domain = ?", domainName).Delete(&models.Domain{}).Error; err != nil {
			return err
		}

		if err := utils.LogAction(tx, username, ip, domainName, "delete_domain", domainName); err != nil {
			return err
		}

		return nil
	})
}
