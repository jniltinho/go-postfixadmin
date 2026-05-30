package repositories

import (
	"fmt"
	"strings"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

func GetMailLogs(db *gorm.DB, username string, isSuperAdmin bool, search, orderField, orderDir string, start, length int) (logs []models.Maillog, total int64, filtered int64, err error) {
	allowedDomains, _, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, 0, 0, err
	}

	applyPermission := func(q *gorm.DB) *gorm.DB {
		if isSuperAdmin {
			return q
		}
		if len(allowedDomains) == 0 {
			return q.Where(db.Where("m_from = ?", username).Or("m_to = ?", username))
		}
		return q.Where(db.Where("m_from = ?", username).Or("m_to = ?", username).Or("domain_from IN ?", allowedDomains).Or("domain_to IN ?", allowedDomains))
	}

	applySearch := func(q *gorm.DB) *gorm.DB {
		if search != "" {
			like := "%" + search + "%"
			q = q.Where(
				db.Where("m_from LIKE ?", like).
					Or("m_to LIKE ?", like).
					Or("host_ip LIKE ?", like).
					Or("host_name LIKE ?", like).
					Or("helo LIKE ?", like),
			)
		}
		return q
	}

	applyPermission(db.Model(&models.Maillog{})).Count(&total)
	applySearch(applyPermission(db.Model(&models.Maillog{}))).Count(&filtered)

	if strings.ToLower(orderDir) != "asc" {
		orderDir = "desc"
	}

	applySearch(applyPermission(db.Model(&models.Maillog{}))).
		Order(fmt.Sprintf("%s %s", orderField, orderDir)).
		Offset(start).Limit(length).
		Find(&logs)

	return logs, total, filtered, nil
}
