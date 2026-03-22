package repositories

import (
	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

func GetAllowedDomains(db *gorm.DB, username string, isSuperAdmin bool) (domains []string, isSuper bool, err error) {
	isSuper = isSuperAdmin
	if !isSuper {
		isSuper, err = IsSuperAdmin(db, username)
		if err != nil {
			return nil, false, err
		}
	}
	if isSuper {
		return nil, true, nil
	}

	var domainAdmins []models.DomainAdmin
	if err := db.Select("domain").Where("username = ? AND active = ?", username, true).Find(&domainAdmins).Error; err != nil {
		return nil, false, err
	}

	for _, da := range domainAdmins {
		if da.Domain == "ALL" {
			return nil, true, nil
		}
		domains = append(domains, da.Domain)
	}

	return domains, false, nil
}
