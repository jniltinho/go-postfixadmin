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

// GetAllMailboxes returns the list of active mailboxes filtered by the user's allowed domains.
// Superadmins get all active mailboxes, regular admins only get mailboxes from their domains.
func GetAllMailboxes(db *gorm.DB, username string) (mailboxes []models.Mailbox, isSuperAdmin bool, err error) {
	allowedDomains, isSuper, err := GetAllowedDomains(db, username)
	if err != nil {
		return nil, false, err
	}

	query := db.Select("username").Where("active = ?", true).Order("domain ASC")
	if !isSuper {
		if len(allowedDomains) == 0 {
			return nil, false, nil
		}
		query = query.Where("domain IN ?", allowedDomains)
	}

	if err := query.Find(&mailboxes).Error; err != nil {
		return nil, isSuper, err
	}

	return mailboxes, isSuper, nil
}
