# SETUP_MAILSERVER.md

Quick mail server setup for **Postfix + Dovecot + Go-PostfixAdmin + MySQL**.

## What This Document Covers

This is the short operational setup for a working mail server. For the full walkthrough, use [DOCUMENTS/setup/README.md](DOCUMENTS/setup/README.md).

Related documents:

- [Project README](README.md)
- [Features](FEATURES.md)
- [Complete setup guide](README.md)

---

## Summary

1. [MySQL Database](#1-mysql-database)
2. [Postfix — main.cf](#2-postfix--maincf)
3. [Postfix SQL Files](#3-postfix-sql-files)
4. [Dovecot](#4-dovecot)
5. [User and Emails Directory](#5-create-user-and-emails-directory)
6. [Restart Services](#6-restart-everything)
7. [Points of Attention](#points-of-attention)

---

## 1. MySQL Database

```sql
CREATE DATABASE postfix;
CREATE USER 'postfix'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL ON postfix.* TO 'postfix'@'localhost';
FLUSH PRIVILEGES;
```

```bash
## After creating the database do:
./postfixadmin --generate-config

## Adjust the generated config.toml file with the saved database password ("your_password")
## You can also enable verbose SQL logging during setup/troubleshooting:
## [database]
## debug = true
##
## And then run the migration:
./postfixadmin migrate
## The Go-PostfixAdmin binary creates tables automatically based on config.toml
```

Example `config.toml` database section:

```toml
[database]
host   = "localhost"
port   = "3306"
user   = "postfix"
pass   = "your_password"
name   = "postfix"
driver = "mysql"
debug  = false
```

Password policy enforced by the backend:

- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

The web interface password generator already follows the same rules, so generated passwords are valid for mailbox, admin, and user password updates.

---

## 2. Postfix — `/etc/postfix/main.cf`

```ini
# Domain and hostname
myhostname = mail.example.com
mydomain   = example.com
myorigin   = $mydomain

# Virtual mailboxes
virtual_mailbox_base    = /var/vmail
virtual_mailbox_domains = proxy:mysql:/etc/postfix/sql/mysql_virtual_domains_maps.cf
virtual_mailbox_maps    = proxy:mysql:/etc/postfix/sql/mysql_virtual_mailbox_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_mailbox_maps.cf
virtual_alias_maps      = proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_catchall_maps.cf

# UID/GID of the vmail user (create with: useradd -u 1001 -g 1001 vmail)
virtual_uid_maps = static:1001
virtual_gid_maps = static:1001

# Delivery via Dovecot LMTP
virtual_transport = lmtp:unix:private/dovecot-lmtp

# SASL via Dovecot
smtpd_sasl_type           = dovecot
smtpd_sasl_path           = private/auth
smtpd_sasl_auth_enable    = yes
smtpd_recipient_restrictions = permit_sasl_authenticated, permit_mynetworks, reject_unauth_destination

# TLS
smtpd_tls_cert_file = /etc/letsencrypt/live/mail.example.com/fullchain.pem
smtpd_tls_key_file  = /etc/letsencrypt/live/mail.example.com/privkey.pem
smtpd_use_tls       = yes
smtpd_tls_auth_only = yes
```

---

## 3. Postfix SQL Files

Create the files below in `/etc/postfix/sql/` and then protect them:

```bash
chmod 640 /etc/postfix/sql/*.cf
chown root:postfix /etc/postfix/sql/*.cf
```

### `mysql_virtual_domains_maps.cf`

```ini
user     = postfix
password = your_password
hosts    = localhost
dbname   = postfix
query    = SELECT domain FROM domain WHERE domain='%s' AND active='1'
```

### `mysql_virtual_mailbox_maps.cf`

```ini
user     = postfix
password = your_password
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox WHERE username='%s' AND active='1'
```

### `mysql_virtual_alias_maps.cf`

```ini
user     = postfix
password = your_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias WHERE address='%s' AND active='1'
```

### `mysql_virtual_alias_domain_maps.cf`

```ini
user     = postfix
password = your_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain
           WHERE alias_domain.alias_domain='%d'
           AND alias.address=CONCAT('%u','@',alias_domain.target_domain)
           AND alias.active='1' AND alias_domain.active='1'
```

### `mysql_virtual_alias_domain_catchall_maps.cf`

```ini
user     = postfix
password = your_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain
           WHERE alias_domain.alias_domain='%d'
           AND alias.address=CONCAT('@',alias_domain.target_domain)
           AND alias.active='1' AND alias_domain.active='1'
```

### `mysql_virtual_alias_domain_mailbox_maps.cf`

```ini
user     = postfix
password = your_password
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox,alias_domain
           WHERE alias_domain.alias_domain='%d'
           AND mailbox.username=CONCAT('%u','@',alias_domain.target_domain)
           AND mailbox.active='1' AND alias_domain.active='1'
```

---

## 4. Dovecot

Instead of making changes across multiple fragmented files, you can configure everything mainly using two files.

### `/etc/dovecot/dovecot.conf`

Create a single consolidated configuration file for Dovecot:

```ini
protocols = imap lmtp pop3
disable_plaintext_auth = no

# SSL/TLS certificates
ssl = required
ssl_cert = </etc/letsencrypt/live/mail.example.com/fullchain.pem
ssl_key = </etc/letsencrypt/live/mail.example.com/privkey.pem

# Path to emails
mail_location = maildir:/var/vmail/%d/%n/Maildir

# System user for delivery (vmail mapped with UID/GID 1001)
mail_uid = 1001
mail_gid = 1001

# Authentication mechanisms
auth_mechanisms = plain login

# Password and User databases via SQL
passdb {
  driver = sql
  args   = /etc/dovecot/dovecot-sql.conf
}
userdb {
  driver = sql
  args   = /etc/dovecot/dovecot-sql.conf
}

# Communication sockets with Postfix
service lmtp {
  unix_listener /var/spool/postfix/private/dovecot-lmtp {
    mode  = 0600
    user  = postfix
    group = postfix
  }
}

service auth {
  unix_listener /var/spool/postfix/private/auth {
    mode  = 0660
    user  = postfix
    group = postfix
  }
}

mail_plugins = quota

protocol imap {
  mail_plugins = $mail_plugins imap_quota
}

protocol lmtp {
  mail_plugins = $mail_plugins quota
}

plugin {
  quota = maildir:User quota

  quota_warning = storage=80%% quota-warning 80 %u
  quota_warning2 = storage=95%% quota-warning 95 %u
}

service quota-warning {
  executable = script /usr/local/bin/quota-warning.sh
  user = vmail
  unix_listener quota-warning {
    mode = 0660
    user = vmail
    group = vmail
  }
}

################################
# LOGGING AND OPTIMIZATION
################################
log_path = syslog
syslog_facility = local5
auth_verbose = yes
auth_debug_passwords = no
```

### `/etc/dovecot/dovecot-sql.conf`

Create `/etc/dovecot/dovecot-sql.conf` for the SQL connection and query settings:

```ini
driver  = mysql
connect = host=localhost dbname=postfix user=postfix password=your_password

# Must match $CONF['encrypt'] in PostfixAdmin
default_pass_scheme = BLF-CRYPT

# Password validation query
password_query = \
  SELECT username AS user, password \
  FROM mailbox WHERE username='%u' AND active='1'

# Home directory and quotas query
user_query = \
  SELECT CONCAT('/var/vmail/', maildir) AS home, \
         1001 AS uid, 1001 AS gid, \
         CONCAT('*:bytes=', quota) AS quota_rule \
  FROM mailbox WHERE username='%u' AND active='1'

# Real-time used quota update
iterate_query = SELECT username AS user FROM mailbox WHERE active='1'
```

Fix the file permissions to protect the passwords:
```bash
sudo chmod 640 /etc/dovecot/dovecot-sql.conf
sudo chown root:dovecot /etc/dovecot/dovecot-sql.conf
```

---

## 5. Create user and emails directory

```bash
groupadd -g 1001 vmail
useradd -g vmail -u 1001 vmail -d /var/vmail -m
chown -R vmail:vmail /var/vmail
```

---

## 6. Restart everything

```bash
systemctl restart postfix dovecot rsyslog
```

---

## Points of Attention

| Item | Detail |
|------|---------|
| **Consistent encryption** | The `default_pass_scheme` in Dovecot must match the method configured in PostfixAdmin |
| **UID/GID of vmail** | Must be the same in `virtual_uid_maps`/`virtual_gid_maps` (Postfix) e `mail_uid`/`mail_gid` (Dovecot) |
| **MySQL Support in Postfix** | Verify with `postconf -m \| grep mysql`. If missing, install `postfix-mysql` |
| **Permissions of SQL files** | Keep `640` with owner `root:postfix` for security |
| **TLS** | Always use `smtpd_tls_auth_only = yes` to prevent credential transmission without encryption |
| **Database debug logging** | Set `database.debug = true` in `config.toml` to enable verbose GORM SQL logs during troubleshooting |
| **Password policy** | Password changes and creations are validated by the backend with the same 8+ chars / upper / lower / number / special rule |
