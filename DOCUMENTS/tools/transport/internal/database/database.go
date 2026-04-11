package database

import (
	"fmt"
	"log"

	"transport/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB  *gorm.DB
	Cfg config.Config
}

func Connect(cfg config.Config) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	logMod := logger.Silent
	if cfg.DBDebug {
		logMod = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMod),
	})
	if err != nil {
		return nil, err
	}

	return &Database{DB: db, Cfg: cfg}, nil
}

// GetEmailTransport queries the mailbox table for a per-user transport override,
// then falls back to the domain-level routing via GetDomainTransport.
//
// Priority:
//  1. mailbox WHERE username = email AND active = true (per-user transport override)
//  2. GetDomainTransport(domain) (domain-level routing)
func (d *Database) GetEmailTransport(email string, domain string) (string, error) {
	// 1. Per-user transport override in mailbox table
	var mb Mailbox
	err := d.DB.Select("transport").Where("username = ? AND active = true", email).First(&mb).Error
	if err == nil && mb.Transport != "" && mb.Transport != "virtual" {
		return d.FormatTransport(mb.Transport), nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("DB error GetEmailTransport(%s): %v", email, err)
		return "", err
	}

	// 2. Fallback to domain-level routing
	return d.GetDomainTransport(domain)
}

// GetDomainTransport queries transport_list first (specific routing rules),
// then falls back to the domain table transport field.
//
// Priority:
//  1. transport_list WHERE domain = ? (explicit per-domain routing rule)
//  2. domain WHERE domain = ? (default transport stored on the domain record)
func (d *Database) GetDomainTransport(domain string) (string, error) {
	// 1. Exact match in transport_list
	var tl TransportList
	if err := d.DB.Select("transport").Where("domain = ? AND active = true", domain).First(&tl).Error; err == nil && tl.Transport != "" {
		return d.FormatTransport(tl.Transport), nil
	}

	// 2. Fallback to domain.transport
	var dmn Domain
	err := d.DB.Select("transport").Where("domain = ? AND active = true", domain).First(&dmn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		log.Printf("DB error GetDomainTransport(%s): %v", domain, err)
		return "", err
	}

	// empty or generic "virtual" means no specific routing
	if dmn.Transport == "" || dmn.Transport == "virtual" {
		return "", nil
	}

	return d.FormatTransport(dmn.Transport), nil
}

// FormatTransport aplica a lógica de reescrita local presente no Perl original.
func (d *Database) FormatTransport(dest string) string {
	local := "smtp:" + d.Cfg.Hostname
	if dest == local {
		return "lmtp:unix:private/dovecot-lmtp"
	}

	if d.Cfg.LocalDelivery != "" && dest == d.Cfg.LocalDelivery {
		return d.Cfg.Delivery
	}

	return dest
}
