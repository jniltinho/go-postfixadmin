package models

import (
	"time"
)

// Domain represents the 'domain' table
type Domain struct {
	Domain         string    `gorm:"primaryKey;column:domain;type:varchar(255);not null"`
	Description    string    `gorm:"column:description;type:varchar(255);not null"`
	Aliases        int       `gorm:"column:aliases;type:int(10);not null;default:0"`
	Mailboxes      int       `gorm:"column:mailboxes;type:int(10);not null;default:0"`
	MaxQuota       int64     `gorm:"column:maxquota;type:bigint(20);not null;default:0"`
	Quota          int64     `gorm:"column:quota;type:bigint(20);not null;default:0"`
	Transport      string    `gorm:"column:transport;type:varchar(255);not null"`
	BackupMX       bool      `gorm:"column:backupmx;type:tinyint(1);not null;default:0"`
	Created        time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Modified       time.Time `gorm:"column:modified;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active         bool      `gorm:"column:active;type:tinyint(1);not null;default:1"`
	PasswordExpiry *int      `gorm:"column:password_expiry;type:int(11);default:0"`
}

func (Domain) TableName() string {
	return "domain"
}

// Mailbox represents the 'mailbox' table
type Mailbox struct {
	Username       string    `gorm:"primaryKey;column:username;type:varchar(255);not null"`
	Password       string    `gorm:"column:password;type:varchar(255);not null"`
	Name           string    `gorm:"column:name;type:varchar(255);not null"`
	Maildir        string    `gorm:"column:maildir;type:varchar(255);not null"`
	Quota          int64     `gorm:"column:quota;type:bigint(20);not null;default:0"`
	LocalPart      string    `gorm:"column:local_part;type:varchar(255);not null"`
	Domain         string    `gorm:"column:domain;type:varchar(255);not null;index:domain"`
	Created        time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Modified       time.Time `gorm:"column:modified;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active         bool      `gorm:"column:active;type:tinyint(1);not null;default:1"`
	Phone          string    `gorm:"column:phone;type:varchar(30);not null;default:''"`
	EmailOther     string    `gorm:"column:email_other;type:varchar(255);not null;default:''"`
	Token          string    `gorm:"column:token;type:varchar(255);not null;default:''"`
	TokenValidity  time.Time `gorm:"column:token_validity;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	PasswordExpiry time.Time `gorm:"column:password_expiry;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	TOTPSecret     *string   `gorm:"column:totp_secret;type:varchar(255)"`
	SMTPActive     bool      `gorm:"column:smtp_active;type:tinyint(1);not null;default:1"`
}

func (Mailbox) TableName() string {
	return "mailbox"
}

// Admin represents the 'admin' table
type Admin struct {
	Username      string    `gorm:"primaryKey;column:username;type:varchar(255);not null"`
	Password      string    `gorm:"column:password;type:varchar(255);not null"`
	Created       time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Modified      time.Time `gorm:"column:modified;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active        bool      `gorm:"column:active;type:tinyint(1);not null;default:1"`
	Superadmin    bool      `gorm:"column:superadmin;type:tinyint(1);not null;default:0"`
	Phone         string    `gorm:"column:phone;type:varchar(30);not null;default:''"`
	EmailOther    string    `gorm:"column:email_other;type:varchar(255);not null;default:''"`
	Token         string    `gorm:"column:token;type:varchar(255);not null;default:''"`
	TokenValidity time.Time `gorm:"column:token_validity;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	TOTPSecret    *string   `gorm:"column:totp_secret;type:varchar(255)"`
}

func (Admin) TableName() string {
	return "admin"
}

// Alias represents the 'alias' table
type Alias struct {
	Address  string    `gorm:"primaryKey;column:address;type:varchar(255);not null"`
	Goto     string    `gorm:"column:goto;type:text;not null"`
	Domain   string    `gorm:"column:domain;type:varchar(255);not null;index:domain"`
	Created  time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Modified time.Time `gorm:"column:modified;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active   bool      `gorm:"column:active;type:tinyint(1);not null;default:1"`
}

func (Alias) TableName() string {
	return "alias"
}

// DomainAdmin represents the 'domain_admins' table
type DomainAdmin struct {
	Username string    `gorm:"column:username;type:varchar(255);not null;index:username"`
	Domain   string    `gorm:"column:domain;type:varchar(255);not null"`
	Created  time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active   bool      `gorm:"column:active;type:tinyint(1);not null;default:1"`
	ID       int       `gorm:"primaryKey;column:id;not null;type:int(11);autoIncrement"`
}

func (DomainAdmin) TableName() string {
	return "domain_admins"
}

// Log represents the 'log' table
type Log struct {
	Timestamp time.Time `gorm:"column:timestamp;type:datetime;not null;default:'2000-01-01 00:00:00';index:timestamp;index:domain_timestamp,composite:0"`
	Username  string    `gorm:"column:username;type:varchar(255);not null"`
	Domain    string    `gorm:"column:domain;type:varchar(255);not null;index:domain_timestamp,composite:1"`
	Action    string    `gorm:"column:action;type:varchar(255);not null"`
	Data      string    `gorm:"column:data;type:text;not null"`
	ID        int       `gorm:"primaryKey;column:id;autoIncrement;not null;type:int(11)"`
}

func (Log) TableName() string {
	return "log"
}

// AliasDomain represents the 'alias_domain' table
type AliasDomain struct {
	AliasDomain  string    `gorm:"primaryKey;column:alias_domain;type:varchar(255);not null;default:''"`
	TargetDomain string    `gorm:"column:target_domain;type:varchar(255);not null;default:'';index"`
	Created      time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Modified     time.Time `gorm:"column:modified;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active       bool      `gorm:"column:active;type:tinyint(1);not null;default:1;index"`
}

func (AliasDomain) TableName() string {
	return "alias_domain"
}

// Fetchmail represents the 'fetchmail' table
type Fetchmail struct {
	ID             uint      `gorm:"primaryKey;column:id;autoIncrement;not null;type:int(11) unsigned"`
	Mailbox        string    `gorm:"column:mailbox;type:varchar(255);not null"`
	SrcServer      string    `gorm:"column:src_server;type:varchar(255);not null"`
	SrcAuth        *string   `gorm:"column:src_auth;type:enum('password','kerberos_v5','kerberos','kerberos_v4','gssapi','cram-md5','otp','ntlm','msn','ssh','any')"`
	SrcUser        string    `gorm:"column:src_user;type:varchar(255);not null"`
	SrcPassword    string    `gorm:"column:src_password;type:varchar(255);not null"`
	SrcFolder      string    `gorm:"column:src_folder;type:varchar(255);not null"`
	PollTime       uint      `gorm:"column:poll_time;type:int(11) unsigned;not null;default:10"`
	Fetchall       bool      `gorm:"column:fetchall;type:tinyint(1) unsigned;not null;default:0"`
	Keep           bool      `gorm:"column:keep;type:tinyint(1) unsigned;not null;default:0"`
	Protocol       *string   `gorm:"column:protocol;type:enum('POP3','IMAP','POP2','ETRN','AUTO')"`
	UseSSL         bool      `gorm:"column:usessl;type:tinyint(1) unsigned;not null;default:0"`
	ExtraOptions   *string   `gorm:"column:extra_options;type:text"`
	ReturnedText   *string   `gorm:"column:returned_text;type:text"`
	Mda            *string   `gorm:"column:mda;type:varchar(255)"`
	Date           time.Time `gorm:"column:date;type:timestamp;not null;default:'2000-01-01 03:00:00'"`
	SSLCertCk      bool      `gorm:"column:sslcertck;type:tinyint(1);not null;default:0"`
	SSLCertPath    string    `gorm:"column:sslcertpath;type:varchar(255);charset:utf8mb4;collation:utf8mb4_general_ci;default:''"`
	SSLFingerprint string    `gorm:"column:sslfingerprint;type:varchar(255);default:''"`
	Domain         string    `gorm:"column:domain;type:varchar(255);default:''"`
	Active         bool      `gorm:"column:active;type:tinyint(1);not null;default:0"`
	Created        time.Time `gorm:"column:created;type:timestamp;not null;default:'2000-01-01 03:00:00'"`
	Modified       time.Time `gorm:"column:modified;type:timestamp;not null;default:current_timestamp();autoUpdateTime"`
	SrcPort        int       `gorm:"column:src_port;type:int(11);not null;default:0"`
}

func (Fetchmail) TableName() string {
	return "fetchmail"
}

// Config represents the 'config' table
type Config struct {
	ID    int    `gorm:"primaryKey;column:id;not null;type:int(11)"`
	Name  string `gorm:"column:name;type:varchar(20);not null;default:'';uniqueIndex"`
	Value string `gorm:"column:value;type:varchar(20);not null;default:''"`
}

func (Config) TableName() string {
	return "config"
}

// MailboxAppPassword represents the 'mailbox_app_password' table
type MailboxAppPassword struct {
	ID           int     `gorm:"primaryKey;column:id;not null;type:int(11);autoIncrement"`
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
	ID          int     `gorm:"primaryKey;column:id;not null;type:int(11);autoIncrement"`
	IP          string  `gorm:"column:ip;uniqueIndex:ip_user,composite:0"`
	Username    *string `gorm:"column:username;uniqueIndex:ip_user,composite:1"`
	Description *string `gorm:"column:description"`
}

func (TOTPExceptionAddress) TableName() string {
	return "totp_exception_address"
}

// Vacation represents the 'vacation' table
type Vacation struct {
	Email        string    `gorm:"primaryKey;column:email;type:varchar(255);not null;index:email"`
	Subject      string    `gorm:"column:subject;type:varchar(255);not null"`
	Body         string    `gorm:"column:body;type:text;not null"`
	Cache        string    `gorm:"column:cache;type:text;not null"`
	Domain       string    `gorm:"column:domain;type:varchar(255);not null"`
	Created      time.Time `gorm:"column:created;type:datetime;not null;default:'2000-01-01 00:00:00'"`
	Active       bool      `gorm:"column:active;type:tinyint(1);not null;default:1"`
	Modified     time.Time `gorm:"column:modified;type:timestamp;not null;default:current_timestamp();autoUpdateTime"`
	ActiveFrom   time.Time `gorm:"column:activefrom;type:timestamp;not null;default:'2000-01-01 03:00:00'"`
	ActiveUntil  time.Time `gorm:"column:activeuntil;type:timestamp;not null;default:'2038-01-18 03:00:00'"`
	IntervalTime int       `gorm:"column:interval_time;type:int(11);not null;default:0"`
}

func (Vacation) TableName() string {
	return "vacation"
}

// VacationNotification represents the 'vacation_notification' table
type VacationNotification struct {
	OnVacation string    `gorm:"primaryKey;column:on_vacation;type:varchar(255);not null"`
	Notified   string    `gorm:"primaryKey;column:notified;type:varchar(255);not null;default:''"`
	NotifiedAt time.Time `gorm:"column:notified_at;type:timestamp;not null;default:current_timestamp()"`
	Vacation   Vacation  `gorm:"foreignKey:OnVacation;references:Email;constraint:OnDelete:CASCADE"`
}

func (VacationNotification) TableName() string {
	return "vacation_notification"
}

// DKIM represents the 'dkim' table
type DKIM struct {
	ID          int       `gorm:"primaryKey;column:id;not null;type:int(11);autoIncrement"`
	DomainName  string    `gorm:"column:domain_name;index:idx_domain_name,composite:0"`
	Description *string   `gorm:"column:description;index:idx_domain_name,composite:1"`
	Selector    string    `gorm:"column:selector;default:default"`
	PrivateKey  *string   `gorm:"column:private_key;type:text"`
	PublicKey   *string   `gorm:"column:public_key;type:text"`
	Created     time.Time `gorm:"column:created"`
	Modified    time.Time `gorm:"column:modified"`
	Domain      Domain    `gorm:"foreignKey:DomainName;references:Domain;constraint:OnDelete:CASCADE"`
}

func (DKIM) TableName() string {
	return "dkim"
}

// DKIMSigning represents the 'dkim_signing' table
type DKIMSigning struct {
	ID       int       `gorm:"primaryKey;column:id;not null;type:int(11);autoIncrement"`
	Author   string    `gorm:"column:author;index"`
	DKIMID   int       `gorm:"column:dkim_id;index"`
	Created  time.Time `gorm:"column:created"`
	Modified time.Time `gorm:"column:modified"`
	DKIM     DKIM      `gorm:"foreignKey:DKIMID;references:ID;constraint:OnDelete:CASCADE"`
}

func (DKIMSigning) TableName() string {
	return "dkim_signing"
}

// Maillog represents the 'maillog' table — stores FILTER entries from mail.log
type Maillog struct {
	LogID      int       `gorm:"column:logid;primaryKey;autoIncrement;not null"`
	Tstamp     int64     `gorm:"column:tstamp;not null;default:0;index"`
	LogDate    time.Time `gorm:"column:logdate;type:date;not null"`
	ProcNum    int       `gorm:"column:procnum;not null;default:0"`
	MFrom      string    `gorm:"column:m_from;type:varchar(150);not null;default:''"`
	MTo        string    `gorm:"column:m_to;type:varchar(150);not null;default:''"`
	DomainTo   string    `gorm:"column:domain_to;type:varchar(150);not null;default:''"`
	Helo       string    `gorm:"column:helo;type:varchar(150);not null;default:''"`
	HostIP     string    `gorm:"column:host_ip;type:varchar(150);not null;default:''"`
	HostName   string    `gorm:"column:host_name;type:varchar(150);not null;default:''"`
	MsgSize    int64     `gorm:"column:msgsize;not null;default:0"`
	DomainFrom string    `gorm:"column:domain_from;type:varchar(150);not null;default:''"`
}

func (Maillog) TableName() string { return "maillog" }
