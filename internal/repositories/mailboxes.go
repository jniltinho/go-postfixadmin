package repositories

import (
	"fmt"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

func GetMailboxByUsername(db *gorm.DB, username string) (models.Mailbox, error) {
	var mailbox models.Mailbox
	err := db.Where("username = ?", username).First(&mailbox).Error
	return mailbox, err
}

func CreateMailbox(db *gorm.DB, mailbox models.Mailbox, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&mailbox).Error; err != nil {
			return err
		}
		alias := models.Alias{
			Address:  mailbox.Username,
			Goto:     mailbox.Username,
			Domain:   mailbox.Domain,
			Created:  mailbox.Created,
			Modified: mailbox.Modified,
			Active:   true,
		}
		if err := tx.Create(&alias).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, mailbox.Domain, "create_mailbox", mailbox.Username)
	})
}

func SaveMailbox(db *gorm.DB, mailbox models.Mailbox, action, actorUsername, ip string) error {
	if err := db.Save(&mailbox).Error; err != nil {
		return err
	}
	return utils.LogAction(db, actorUsername, ip, mailbox.Domain, action, mailbox.Username)
}

func GetDashboardCounts(db *gorm.DB, username string, isSuperAdmin bool) (domainCount, mailboxCount int64, err error) {
	allowedDomains, _, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return 0, 0, err
	}

	domainQ := db.Model(&models.Domain{}).Where("active = ? AND domain != ?", true, "ALL")
	mailboxQ := db.Model(&models.Mailbox{}).Where("active = ?", true)

	if !isSuperAdmin {
		if len(allowedDomains) == 0 {
			return 0, 0, nil
		}
		domainQ = domainQ.Where("domain IN ?", allowedDomains)
		mailboxQ = mailboxQ.Where("domain IN ?", allowedDomains)
	}

	domainQ.Count(&domainCount)
	mailboxQ.Count(&mailboxCount)
	return domainCount, mailboxCount, nil
}

func GetRecentLogs(db *gorm.DB, username string, isSuperAdmin bool, since time.Time) ([]models.Log, error) {
	allowedDomains, _, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	logQ := db.Order("timestamp desc").Where("timestamp >= ?", since).Limit(500)

	if !isSuperAdmin {
		if len(allowedDomains) == 0 {
			logQ = logQ.Where("1 = 0")
		} else {
			logQ = logQ.Where("domain IN ?", allowedDomains)
		}
	}

	var logs []models.Log
	logQ.Find(&logs)
	return logs, nil
}

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

func DeleteMailbox(db *gorm.DB, mailbox models.Mailbox, actorUsername, actorIP string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("address = ?", mailbox.Username).Delete(&models.Alias{}).Error; err != nil {
			return err
		}

		if err := tx.Where("email = ?", mailbox.Username).Delete(&models.Vacation{}).Error; err != nil {
			return err
		}

		if err := tx.Where("username = ?", mailbox.Username).Delete(&models.Mailbox{}).Error; err != nil {
			return err
		}

		if err := utils.LogAction(tx, actorUsername, actorIP, mailbox.Domain, "delete_mailbox", mailbox.Username); err != nil {
			return err
		}

		return nil
	})
}
