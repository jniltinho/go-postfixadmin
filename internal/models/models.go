package models

import (
	"time"
)

// Domain represents the 'domain' table
type Domain struct {
	Domain         string    `gorm:"primaryKey;column:domain"`
	Description    string    `gorm:"column:description"`
	Aliases        int       `gorm:"column:aliases"`
	Mailboxes      int       `gorm:"column:mailboxes"`
	MaxQuota       int64     `gorm:"column:maxquota"`
	Quota          int64     `gorm:"column:quota"`
	Transport      string    `gorm:"column:transport"`
	BackupMX       bool      `gorm:"column:backupmx"`
	Created        time.Time `gorm:"column:created;default:'2000-01-01 00:00:00'"`
	Modified       time.Time `gorm:"column:modified;default:'2000-01-01 00:00:00'"`
	Active         bool      `gorm:"column:active"`
	PasswordExpiry *int      `gorm:"column:password_expiry"`
}

func (Domain) TableName() string {
	return "domain"
}

// Mailbox represents the 'mailbox' table
type Mailbox struct {
	Username       string    `gorm:"primaryKey;column:username"`
	Password       string    `gorm:"column:password;not null"`
	Name           string    `gorm:"column:name"`
	Maildir        string    `gorm:"column:maildir;not null"`
	Quota          int64     `gorm:"column:quota;default:0;not null"`
	LocalPart      string    `gorm:"column:local_part;not null"`
	Domain         string    `gorm:"column:domain;index:domain;not null"`
	Created        time.Time `gorm:"column:created;default:'2000-01-01 00:00:00'"`
	Modified       time.Time `gorm:"column:modified;default:'2000-01-01 00:00:00'"`
	Active         bool      `gorm:"column:active;default:true;not null"`
	EmailOther     string    `gorm:"column:email_other"`
	Phone          string    `gorm:"column:phone;default:'+0000000000000'"`
	Token          string    `gorm:"column:token"`
	TokenValidity  time.Time `gorm:"column:token_validity;default:'2000-01-01 00:00:00'"`
	PasswordExpiry time.Time `gorm:"column:password_expiry;default:'2000-01-01 00:00:00'"`
	TOTPSecret     *string   `gorm:"column:totp_secret;default:null"`
	SMTPActive     bool      `gorm:"column:smtp_active;default:true"`
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
	TokenValidity time.Time `gorm:"column:token_validity;"`
	TOTPSecret    *string   `gorm:"column:totp_secret"`
}

func (Admin) TableName() string {
	return "admin"
}

// Alias represents the 'alias' table
type Alias struct {
	Address  string    `gorm:"primaryKey;column:address"`
	Goto     string    `gorm:"column:goto;type:text"`
	Domain   string    `gorm:"column:domain;index"`
	Created  time.Time `gorm:"column:created"`
	Modified time.Time `gorm:"column:modified"`
	Active   bool      `gorm:"column:active"`
}

func (Alias) TableName() string {
	return "alias"
}

// DomainAdmin represents the 'domain_admins' table
type DomainAdmin struct {
	ID       int       `gorm:"primaryKey;column:id;autoIncrement"`
	Username string    `gorm:"column:username;index:username"`
	Domain   string    `gorm:"column:domain;index:domain"`
	Created  time.Time `gorm:"column:created;default:'2000-01-01 00:00:00'"`
	Active   bool      `gorm:"column:active"`
}

func (DomainAdmin) TableName() string {
	return "domain_admins"
}

// Log represents the 'log' table
type Log struct {
	Timestamp time.Time `gorm:"column:timestamp;default:'2000-01-01 00:00:00'"`
	Username  string    `gorm:"column:username"`
	Domain    string    `gorm:"column:domain"`
	Action    string    `gorm:"column:action"`
	Data      string    `gorm:"column:data"`
	ID        int       `gorm:"primaryKey;column:id;autoIncrement"`
}

func (Log) TableName() string {
	return "log"
}

// AliasDomain represents the 'alias_domain' table
type AliasDomain struct {
	AliasDomain  string    `gorm:"primaryKey;column:alias_domain"`
	TargetDomain string    `gorm:"column:target_domain;index"`
	Created      time.Time `gorm:"column:created"`
	Modified     time.Time `gorm:"column:modified"`
	Active       bool      `gorm:"column:active;index"`
}

func (AliasDomain) TableName() string {
	return "alias_domain"
}

// Fetchmail represents the 'fetchmail' table
type Fetchmail struct {
	ID             int       `gorm:"primaryKey;column:id;autoIncrement"`
	Mailbox        string    `gorm:"column:mailbox"`
	SrcServer      string    `gorm:"column:src_server"`
	SrcAuth        *string   `gorm:"column:src_auth"`
	SrcUser        string    `gorm:"column:src_user"`
	SrcPassword    string    `gorm:"column:src_password"`
	SrcFolder      string    `gorm:"column:src_folder"`
	PollTime       int       `gorm:"column:poll_time;default:10"`
	Fetchall       bool      `gorm:"column:fetchall;default:false"`
	Keep           bool      `gorm:"column:keep;default:false"`
	Protocol       *string   `gorm:"column:protocol"`
	UseSSL         bool      `gorm:"column:usessl;default:false"`
	ExtraOptions   *string   `gorm:"column:extra_options;type:text"`
	ReturnedText   *string   `gorm:"column:returned_text;type:text"`
	Mda            *string   `gorm:"column:mda"`
	Date           time.Time `gorm:"column:date;type:timestamp;default:'2000-01-01 00:00:00';not null"`
	SSLCertCk      bool      `gorm:"column:sslcertck;default:false"`
	SSLCertPath    *string   `gorm:"column:sslcertpath"`
	SSLFingerprint *string   `gorm:"column:sslfingerprint"`
	Domain         *string   `gorm:"column:domain"`
	Active         bool      `gorm:"column:active;default:false"`
	Created        time.Time `gorm:"column:created;default:'2000-01-01 00:00:00'"`
	Modified       time.Time `gorm:"column:modified;type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime;not null"`
	SrcPort        int       `gorm:"column:src_port;default:0"`
}

func (Fetchmail) TableName() string {
	return "fetchmail"
}

// Config represents the 'config' table
type Config struct {
	ID    int    `gorm:"primaryKey;column:id;autoIncrement"`
	Name  string `gorm:"column:name;unique"`
	Value string `gorm:"column:value"`
}

func (Config) TableName() string {
	return "config"
}

// MailboxAppPassword represents the 'mailbox_app_password' table
type MailboxAppPassword struct {
	ID           int     `gorm:"primaryKey;column:id;autoIncrement"`
	Username     *string `gorm:"column:username"`
	Description  *string `gorm:"column:description"`
	PasswordHash *string `gorm:"column:password_hash"`
}

func (MailboxAppPassword) TableName() string {
	return "mailbox_app_password"
}

// Quota represents the 'quota' table
type Quota struct {
	Username string `gorm:"primaryKey;column:username"`
	Path     string `gorm:"primaryKey;column:path;not null"`
	Current  int64  `gorm:"column:current;default:0;not null"`
}

func (Quota) TableName() string {
	return "quota"
}

// Quota2 represents the 'quota2' table
type Quota2 struct {
	Username string `gorm:"primaryKey;column:username"`
	Bytes    int64  `gorm:"column:bytes;default:0;not null"`
	Messages int    `gorm:"column:messages;default:0;not null"`
}

func (Quota2) TableName() string {
	return "quota2"
}

// TOTPExceptionAddress represents the 'totp_exception_address' table
type TOTPExceptionAddress struct {
	ID          int     `gorm:"primaryKey;column:id;autoIncrement"`
	IP          string  `gorm:"column:ip;index"`
	Username    *string `gorm:"column:username;index"`
	Description *string `gorm:"column:description"`
}

func (TOTPExceptionAddress) TableName() string {
	return "totp_exception_address"
}

// Vacation represents the 'vacation' table
type Vacation struct {
	Email        string    `gorm:"primaryKey;column:email"`
	Subject      string    `gorm:"column:subject"`
	Body         string    `gorm:"column:body;type:text"`
	Cache        string    `gorm:"column:cache;type:text"`
	Domain       string    `gorm:"column:domain;not null"`
	Created      time.Time `gorm:"column:created;default:'2000-01-01 00:00:00'"`
	Active       bool      `gorm:"column:active;default:true;not null"`
	Modified     time.Time `gorm:"column:modified;type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime;not null"`
	ActiveFrom   time.Time `gorm:"column:activefrom;type:timestamp;default:'2000-01-01 00:00:00';not null"`
	ActiveUntil  time.Time `gorm:"column:activeuntil;type:timestamp;default:'2000-01-01 00:00:00';not null"`
	IntervalTime int       `gorm:"column:interval_time;default:0"`
}

func (Vacation) TableName() string {
	return "vacation"
}

// VacationNotification represents the 'vacation_notification' table
type VacationNotification struct {
	OnVacation string    `gorm:"primaryKey;column:on_vacation"`
	Notified   string    `gorm:"primaryKey;column:notified"`
	NotifiedAt time.Time `gorm:"column:notified_at;type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime;not null"`
}

func (VacationNotification) TableName() string {
	return "vacation_notification"
}

// DKIM represents the 'dkim' table
type DKIM struct {
	ID          int       `gorm:"primaryKey;column:id;autoIncrement"`
	DomainName  string    `gorm:"column:domain_name;index"`
	Description *string   `gorm:"column:description"`
	Selector    string    `gorm:"column:selector;default:default"`
	PrivateKey  *string   `gorm:"column:private_key;type:text"`
	PublicKey   *string   `gorm:"column:public_key;type:text"`
	Created     time.Time `gorm:"column:created"`
	Modified    time.Time `gorm:"column:modified"`
}

func (DKIM) TableName() string {
	return "dkim"
}

// DKIMSigning represents the 'dkim_signing' table
type DKIMSigning struct {
	ID       int       `gorm:"primaryKey;column:id;autoIncrement"`
	Author   string    `gorm:"column:author;index"`
	DKIMID   int       `gorm:"column:dkim_id;index"`
	Created  time.Time `gorm:"column:created"`
	Modified time.Time `gorm:"column:modified"`
}

func (DKIMSigning) TableName() string {
	return "dkim_signing"
}
