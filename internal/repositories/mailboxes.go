package repositories

import (
	"fmt"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

func GetAllMailboxes(db *gorm.DB, username string, isSuperAdmin bool, domainFilter string) (mailboxes []models.Mailbox, isSuper bool, err error) {
	allowedDomains, isSuper, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, false, err
	}

	query := db.Table("mailbox").Select("mailbox.*").
		Joins("JOIN domain ON mailbox.domain = domain.domain").
		Where("domain.active = ?", true).
		Order("mailbox.domain ASC")

	if !isSuper {
		if len(allowedDomains) == 0 {
			return nil, false, nil
		}
		query = query.Where("mailbox.domain IN ?", allowedDomains)
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
				return nil, false, fmt.Errorf("access denied to this domain")
			}
		}
		query = query.Where("mailbox.domain = ?", domainFilter)
	}

	if err := query.Find(&mailboxes).Error; err != nil {
		return nil, isSuper, err
	}

	return mailboxes, isSuper, nil
}
