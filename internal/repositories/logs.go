package repositories

import (
	"fmt"
	"strings"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

func GetLogs(db *gorm.DB, username string, isSuperAdmin bool, filterAdmin, filterDomain, filterAction, search, orderField, orderDir string, start, length int) (logs []models.Log, total int64, filtered int64, err error) {
	allowedDomains, _, err := GetAllowedDomains(db, username, isSuperAdmin)
	if err != nil {
		return nil, 0, 0, err
	}

	applyPermission := func(q *gorm.DB) *gorm.DB {
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				return q.Where("1 = 0")
			}
			return q.Where("domain IN ?", allowedDomains)
		}
		return q
	}

	applyFilters := func(q *gorm.DB) *gorm.DB {
		if filterAdmin != "" {
			q = q.Where("username LIKE ?", "%"+filterAdmin+"%")
		}
		if filterDomain != "" {
			q = q.Where("domain LIKE ?", "%"+filterDomain+"%")
		}
		if filterAction != "" {
			q = q.Where("action LIKE ?", "%"+filterAction+"%")
		}
		if search != "" {
			like := "%" + search + "%"
			q = q.Where(
				db.Where("username LIKE ?", like).
					Or("domain LIKE ?", like).
					Or("action LIKE ?", like).
					Or("data LIKE ?", like),
			)
		}
		return q
	}

	applyPermission(db.Model(&models.Log{})).Count(&total)
	applyFilters(applyPermission(db.Model(&models.Log{}))).Count(&filtered)

	if strings.ToLower(orderDir) != "asc" && strings.ToLower(orderDir) != "desc" {
		orderDir = "desc"
	}

	dataQ := applyFilters(applyPermission(db.Model(&models.Log{}))).
		Order(fmt.Sprintf("%s %s", orderField, orderDir))
	if length > 0 {
		dataQ = dataQ.Offset(start).Limit(length)
	}
	dataQ.Find(&logs)

	return logs, total, filtered, nil
}
