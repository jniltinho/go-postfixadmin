package repositories

import (
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

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
