package repositories

import (
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

func GetAliasDomains(db *gorm.DB, username string, isSuperAdmin bool) (aliasDomains []models.AliasDomain, isSuper bool, err error) {
	allowedDomains, isSuper, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, false, err
	}

	query := db.Table("alias_domain").Select("alias_domain.*").
		Joins("JOIN domain ON alias_domain.target_domain = domain.domain").
		Where("domain.active = ?", true).
		Order("alias_domain.alias_domain ASC")

	if !isSuper {
		if len(allowedDomains) == 0 {
			return nil, false, nil
		}
		query = query.Where("alias_domain.target_domain IN ?", allowedDomains)
	}

	if err := query.Find(&aliasDomains).Error; err != nil {
		return nil, isSuper, err
	}
	return aliasDomains, isSuper, nil
}

func GetAliasDomainByName(db *gorm.DB, name string) (models.AliasDomain, error) {
	var aliasDomain models.AliasDomain
	err := db.Where("alias_domain = ?", name).First(&aliasDomain).Error
	return aliasDomain, err
}

func CreateAliasDomain(db *gorm.DB, aliasDomain models.AliasDomain, actorUsername, ip string) error {
	if err := db.Create(&aliasDomain).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, aliasDomain.TargetDomain, "create_alias_domain", aliasDomain.AliasDomain)
}

func SaveAliasDomain(db *gorm.DB, aliasDomain models.AliasDomain, actorUsername, ip string) error {
	aliasDomain.Modified = time.Now()
	if err := db.Save(&aliasDomain).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, aliasDomain.TargetDomain, "edit_alias_domain", aliasDomain.AliasDomain)
}

func DeleteAliasDomain(db *gorm.DB, aliasDomainName, actorUsername, ip string) error {
	var aliasDomain models.AliasDomain
	if err := db.Where("alias_domain = ?", aliasDomainName).First(&aliasDomain).Error; err != nil {
		return err
	}
	if err := db.Delete(&aliasDomain).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, aliasDomain.TargetDomain, "delete_alias_domain", aliasDomainName)
}
