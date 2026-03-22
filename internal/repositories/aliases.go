package repositories

import (
	"fmt"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

func GetAllAliases(db *gorm.DB, username string, isSuperAdmin bool, domainFilter string) (aliases []models.Alias, isSuper bool, err error) {
	allowedDomains, isSuper, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, false, err
	}

	query := db.Table("alias").Select("alias.*").
		Joins("JOIN domain ON alias.domain = domain.domain").
		Where("domain.active = ?", true).
		Order("alias.address ASC").
		Where("address NOT IN (?)", db.Table("mailbox").Select("username"))

	if !isSuper {
		if len(allowedDomains) == 0 {
			return nil, false, nil
		}
		query = query.Where("alias.domain IN ?", allowedDomains)
	}

	if domainFilter != "" {
		if !isSuper {
			allowed := false
			for _, d := range allowedDomains {
				if d == domainFilter {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, isSuper, fmt.Errorf("access denied to this domain")
			}
		}
		query = query.Where("alias.domain = ?", domainFilter)
	}

	if err := query.Find(&aliases).Error; err != nil {
		return nil, isSuper, err
	}
	return aliases, isSuper, nil
}

func GetAliasByAddress(db *gorm.DB, address string) (models.Alias, error) {
	var alias models.Alias
	err := db.Where("address = ?", address).First(&alias).Error
	return alias, err
}

func CreateAlias(db *gorm.DB, alias models.Alias, actorUsername, ip string) error {
	if err := db.Create(&alias).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, alias.Domain, "create_alias", alias.Address)
}

func SaveAlias(db *gorm.DB, alias models.Alias, actorUsername, ip string) error {
	alias.Modified = time.Now()
	if err := db.Save(&alias).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, alias.Domain, "edit_alias", alias.Address)
}

func DeleteAlias(db *gorm.DB, address, actorUsername, ip string) error {
	var alias models.Alias
	if err := db.Select("domain").Where("address = ?", address).First(&alias).Error; err != nil {
		return err
	}
	if result := db.Where("address = ?", address).Delete(&models.Alias{}); result.Error != nil {
		return result.Error
	}
	return utils.LogAction(db, actorUsername, ip, alias.Domain, "delete_alias", address)
}

func UpdateUserForwarding(db *gorm.DB, username, gotoStr, domain, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var alias models.Alias
		if err := tx.First(&alias, "address = ?", username).Error; err != nil {
			alias = models.Alias{
				Address:  username,
				Goto:     gotoStr,
				Domain:   domain,
				Created:  time.Now(),
				Modified: time.Now(),
				Active:   true,
			}
		} else {
			alias.Goto = gotoStr
			alias.Modified = time.Now()
		}
		if err := tx.Save(&alias).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, domain, "USER_EDIT_ALIAS", gotoStr)
	})
}
