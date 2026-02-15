package utils

import (
	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// IsSuperAdmin checks if the user is a superadmin
func IsSuperAdmin(db *gorm.DB, username string) (bool, error) {
	var admin models.Admin
	if err := db.Select("superadmin").Where("username = ?", username).First(&admin).Error; err != nil {
		return false, err
	}
	return admin.Superadmin, nil
}

// GetAllowedDomains returns the list of domains a user is allowed to manage.
// If the user is a superadmin or has "ALL" in domain_admins, isSuperAdmin returns true.
func GetAllowedDomains(db *gorm.DB, username string) (domains []string, isSuperAdmin bool, err error) {
	// Check if user is superadmin
	isSuper, err := IsSuperAdmin(db, username)
	if err != nil {
		return nil, false, err
	}
	if isSuper {
		return nil, true, nil
	}

	// Get domains from domain_admins
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
