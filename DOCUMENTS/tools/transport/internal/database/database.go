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

// GetTransport resolves the transport for an email address or domain in a
// single query. When email is empty, only domain-level rules apply.
//
// Priority:
//  1. mailbox.transport       WHERE username = email  (per-user override)
//  2. transport_list.transport WHERE domain  = domain (explicit domain rule)
//  3. domain.transport         WHERE domain  = domain (default domain transport)
func (d *Database) GetTransport(email, domain string) (string, error) {
	var result struct {
		MailboxTransport       string `gorm:"column:mailbox_transport"`
		TransportListTransport string `gorm:"column:transport_list_transport"`
		DomainTransport        string `gorm:"column:domain_transport"`
	}

	err := d.DB.Raw(`
		SELECT
			d.transport  AS domain_transport,
			tl.transport AS transport_list_transport,
			m.transport  AS mailbox_transport
		FROM domain d
		LEFT JOIN transport_list tl ON tl.domain   = d.domain   AND tl.active = true
		LEFT JOIN mailbox        m  ON m.username  = ?          AND m.active  = true
		WHERE d.domain = ? AND d.active = true
		LIMIT 1
	`, email, domain).Scan(&result).Error
	if err != nil {
		log.Printf("DB error GetTransport(%s, %s): %v", email, domain, err)
		return "", err
	}

	if result.MailboxTransport != "" && result.MailboxTransport != "virtual" {
		return d.FormatTransport(result.MailboxTransport), nil
	}
	if result.TransportListTransport != "" {
		return d.FormatTransport(result.TransportListTransport), nil
	}
	if result.DomainTransport != "" && result.DomainTransport != "virtual" {
		return d.FormatTransport(result.DomainTransport), nil
	}

	return "", nil
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
