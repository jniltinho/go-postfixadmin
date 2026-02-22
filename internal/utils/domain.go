package utils

import (
	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// DeleteDomain removes a domain and all associated data in a single transaction.
// Deletes: alias_domains, aliases, mailboxes, domain_admins, fetchmail, vacation, and the domain itself.
func DeleteDomain(db *gorm.DB, domainName, username, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Delete alias_domain records (source)
		if err := tx.Where("alias_domain = ?", domainName).Delete(&models.AliasDomain{}).Error; err != nil {
			return err
		}

		// Delete alias_domain records (target)
		if err := tx.Where("target_domain = ?", domainName).Delete(&models.AliasDomain{}).Error; err != nil {
			return err
		}

		// Delete aliases
		if err := tx.Where("domain = ?", domainName).Delete(&models.Alias{}).Error; err != nil {
			return err
		}

		// Delete mailboxes
		if err := tx.Where("domain = ?", domainName).Delete(&models.Mailbox{}).Error; err != nil {
			return err
		}

		// Delete domain_admins
		if err := tx.Where("domain = ?", domainName).Delete(&models.DomainAdmin{}).Error; err != nil {
			return err
		}

		// Delete fetchmail
		if err := tx.Where("domain = ?", domainName).Delete(&models.Fetchmail{}).Error; err != nil {
			return err
		}

		// Delete vacation
		if err := tx.Where("domain = ?", domainName).Delete(&models.Vacation{}).Error; err != nil {
			return err
		}

		// Delete the domain itself
		if err := tx.Where("domain = ?", domainName).Delete(&models.Domain{}).Error; err != nil {
			return err
		}

		// Log action inside transaction
		if err := LogAction(tx, username, ip, domainName, "delete_domain", domainName); err != nil {
			return err
		}

		return nil
	})
}
