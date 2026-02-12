package models

import (
	"time"
)

// Domain represents the 'domain' table
type Domain struct {
	Domain      string    `gorm:"primaryKey;column:domain"`
	Description string    `gorm:"column:description"`
	Aliases     int       `gorm:"column:aliases"`
	Mailboxes   int       `gorm:"column:mailboxes"`
	MaxQuota    int64     `gorm:"column:maxquota"`
	Quota       int64     `gorm:"column:quota"`
	Transport   string    `gorm:"column:transport"`
	BackupMX    bool      `gorm:"column:backupmx"`
	Created     time.Time `gorm:"column:created"`
	Modified    time.Time `gorm:"column:modified"`
	Active      bool      `gorm:"column:active"`
}

func (Domain) TableName() string {
	return "domain"
}

// Mailbox represents the 'mailbox' table
type Mailbox struct {
	Username string    `gorm:"primaryKey;column:username"`
	Password string    `gorm:"column:password"`
	Name     string    `gorm:"column:name"`
	Maildir  string    `gorm:"column:maildir"`
	Quota    int64     `gorm:"column:quota"`
	Domain   string    `gorm:"column:domain"`
	Created  time.Time `gorm:"column:created"`
	Modified time.Time `gorm:"column:modified"`
	Active   bool      `gorm:"column:active"`
}

func (Mailbox) TableName() string {
	return "mailbox"
}

// Admin represents the 'admin' table
type Admin struct {
	Username      string    `gorm:"primaryKey;column:username"`
	Password      string    `gorm:"column:password"`
	Created       time.Time `gorm:"column:created"`
	Modified      time.Time `gorm:"column:modified"`
	Active        bool      `gorm:"column:active"`
	Superadmin    bool      `gorm:"column:superadmin"`
	Phone         string    `gorm:"column:phone"`
	EmailOther    string    `gorm:"column:email_other"`
	Token         string    `gorm:"column:token"`
	TokenValidity time.Time `gorm:"column:token_validity"`
	TOTPSecret    *string   `gorm:"column:totp_secret"`
}

func (Admin) TableName() string {
	return "admin"
}

// Alias represents the 'alias' table
type Alias struct {
	Address  string    `gorm:"primaryKey;column:address"`
	Goto     string    `gorm:"column:goto"`
	Domain   string    `gorm:"column:domain"`
	Created  time.Time `gorm:"column:created"`
	Modified time.Time `gorm:"column:modified"`
	Active   bool      `gorm:"column:active"`
}

func (Alias) TableName() string {
	return "alias"
}
