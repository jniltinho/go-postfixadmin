# Installation Guide: Email Server (Ubuntu) + Go-PostfixAdmin

This step-by-step guide teaches you how to prepare a complete email server on Ubuntu using **Postfix**, **Dovecot**, **MariaDB**, and how to manage everything via **Go-PostfixAdmin**.

---

## 1. Update the System and Install Dependencies

On Ubuntu, update your packages and install the necessary basic services:

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install postfix postfix-mysql dovecot-core dovecot-imapd dovecot-pop3d dovecot-lmtpd dovecot-mysql mariadb-server -y
sudo apt install certbot git curl -y
```

During Postfix installation, the wizard will ask for the configuration type. Select **"Internet Site"** and enter your main domain (e.g., `example.com`).

---

## 2. Configure the MariaDB Database

First, run the secure installation script to improve security:

```bash
sudo mariadb-secure-installation
```

During the wizard, you can adjust several security configurations by answering the prompted questions. Here is what you can do during the process:
- Validate password strength to prevent users from using weak credentials.
- Disallow root login remotely, allowing database access only from localhost.
- Delete anonymous user accounts, which enable all users to log in to your database. 
- Remove the default test database that is accessible to all users.

Next, access the MariaDB console:

```bash
sudo mariadb
```

Run the commands below to create the database and the user that Postfix, Dovecot, and Go-PostfixAdmin will use:

```sql
CREATE DATABASE postfix;
CREATE USER 'postfix'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL ON postfix.* TO 'postfix'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

> **Note:** Remember to replace `your_secure_password` with a strong password in all steps of this guide.

---

## 3. Install and Configure Go-PostfixAdmin

Go-PostfixAdmin will manage the database structure (tables, domains, accounts, aliases, etc.).

1. **Get the Application:**

   You can clone the original repository or directly download the latest compiled executable from the Releases page.
   
   *Download binary and Repository:*
   ```bash
   sudo mkdir -p /opt/go-postfixadmin
   cd /opt/go-postfixadmin
   # Download the latest release:
   TAG=$(curl -s https://api.github.com/repos/jniltinho/go-postfixadmin/releases/latest|grep tag_name|cut -d '"' -f4|tr -d v)
   sudo curl -L -o postfixadmin_${TAG}_linux_amd64.tar.gz "https://github.com/jniltinho/go-postfixadmin/releases/download/v${TAG}/postfixadmin_${TAG}_linux_amd64.tar.gz"
   sudo tar -xzvf postfixadmin_*.tar.gz
   ```
   
2. **Generate Initial SSL Certificates (Certbot):**
   
   Before configuring secure server routes, generate the primary certificates. Stop any service using port 80 and run:
   ```bash
   sudo certbot certonly --standalone -d mail.example.com
   ```

3. **Configure the Environment (config.toml):**
   You can generate a default config file using the native CLI command or copy the example file.
   
   *Generating via CLI:*
   ```bash
   cd /opt/go-postfixadmin
   ./postfixadmin --generate-config
   ```

   *Or copying the example:*
   ```bash
   cp download/config.toml.example /opt/go-postfixadmin/config.toml
   ```

   After that, edit the generated (`config*.toml`) or copied file and add the necessary database and session settings to make it work properly:
   
   ```toml
   [database]
   # Format: user:password@tcp(host:port)/dbname?args | driver = "mysql"
   url = "postfix:your_secure_password@tcp(localhost:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
   
   [server]
   # Web Server Configuration. For SSL use port 443
   port = 8080
   cleanup_maildir = false # Clean up orphaned maildirs when deleting a mailbox
   
   [ssl]
   # (Optional) SSL Settings for standalone secure server
   enabled = true
   cert = "/etc/letsencrypt/live/mail.example.com/fullchain.pem"
   key = "/etc/letsencrypt/live/mail.example.com/privkey.pem"
   # Secret Session Key (Generate a 64-character hex string via: openssl rand -hex 32)
   session_secret = "your_super_secret_session_key_here"

   [quota]
   enabled      = false
   domain_quota = true
   # Bytes per MB: 1024000 or 1048576
   multiplier   = 1048576

   [vacation]
   enabled = true

   [alias]
   edit_alias   = true
   alias_domain = true

   [transport]
   enabled  = true
   options  = ["virtual", "local", "relay"]
   default  = "virtual"

   [features]
   fetchmail = false

   [smtp]
   server  = "localhost"
   port    = 25
   subject = "Welcome!"
   body    = "Hi,\n\nWelcome to your new account."
   type    = "plain" # type: plain | tls | starttls
   ```

3. **Run Migrations:**
   Before starting the service, create the necessary tables by running migrations:
   ```bash
   cd /opt/go-postfixadmin
   ./postfixadmin migrate
   ```

4. **Configure the Systemd Service:**
   Copy the service file (provided in `DOCUMENTS/setup/postfixadmin.service`) to systemd:
   ```bash
   wget https://raw.githubusercontent.com/jniltinho/go-postfixadmin/refs/heads/main/DOCUMENTS/setup/postfixadmin.service
   sudo cp postfixadmin.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now postfixadmin.service
   ```
   *Check the logs with: `tail -f /opt/go-postfixadmin/postfixadmin.log`*

---

## 4. Configure Postfix

### General Configuration (`/etc/postfix/main.cf`)

Back up the original file:
```bash
sudo cp /etc/postfix/main.cf /etc/postfix/main.cf.bkp
```

Edit `/etc/postfix/main.cf` and change/add the following entries:

```ini
# Domain and hostname (Adjust to your reality)
myhostname = mail.example.com
mydomain   = example.com
myorigin   = $mydomain

# Virtual mailboxes (MariaDB Integration via Go-PostfixAdmin)
virtual_mailbox_base    = /var/vmail
virtual_mailbox_domains = proxy:mysql:/etc/postfix/sql/mysql_virtual_domains_maps.cf
virtual_mailbox_maps    = proxy:mysql:/etc/postfix/sql/mysql_virtual_mailbox_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_mailbox_maps.cf
virtual_alias_maps      = proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_catchall_maps.cf

# UID/GID of the vmail user (we will create this later)
virtual_uid_maps = static:1001
virtual_gid_maps = static:1001

# Delivery via Dovecot LMTP
virtual_transport = lmtp:unix:private/dovecot-lmtp

# SASL via Dovecot (Authentication)
smtpd_sasl_type           = dovecot
smtpd_sasl_path           = private/auth
smtpd_sasl_auth_enable    = yes
smtpd_recipient_restrictions = permit_sasl_authenticated, permit_mynetworks, reject_unauth_destination

# TLS (Recommended via Let's Encrypt - configure the correct paths)
# smtpd_tls_cert_file = /etc/letsencrypt/live/mail.example.com/fullchain.pem
# smtpd_tls_key_file  = /etc/letsencrypt/live/mail.example.com/privkey.pem
# smtpd_use_tls       = yes
# smtpd_tls_auth_only = yes
```

### Enable Submission Port (`/etc/postfix/master.cf`)

Edit `/etc/postfix/master.cf` to enable the submission port (587) for sending emails securely. Uncomment the `submission` section and modify it to match the configuration below:

```ini
submission inet n       -       y       -       -       smtpd
  -o syslog_name=postfix/submission
  -o smtpd_tls_security_level=encrypt
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_tls_auth_only=yes
  -o smtpd_reject_unlisted_recipient=no
  -o smtpd_client_restrictions=$mua_client_restrictions
  -o smtpd_helo_restrictions=$mua_helo_restrictions
  -o smtpd_sender_restrictions=$mua_sender_restrictions
  -o smtpd_recipient_restrictions=
  -o smtpd_relay_restrictions=permit_sasl_authenticated,reject
  -o milter_macro_daemon_name=ORIGINATING
```

---

## 5. SQL Maps for Postfix

Create the SQL query files in the directory below and ensure proper permissions:

```bash
sudo mkdir -p /etc/postfix/sql
sudo chown root:postfix /etc/postfix/sql
```

> **Warning:** In all the files below, replace `your_secure_password` with the password configured in MariaDB.

### `/etc/postfix/sql/mysql_virtual_domains_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT domain FROM domain WHERE domain='%s' AND active='1'
```

### `/etc/postfix/sql/mysql_virtual_mailbox_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox WHERE username='%s' AND active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias WHERE address='%s' AND active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_domain_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain='%d' AND alias.address=CONCAT('%u','@',alias_domain.target_domain) AND alias.active='1' AND alias_domain.active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_domain_catchall_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain='%d' AND alias.address=CONCAT('@',alias_domain.target_domain) AND alias.active='1' AND alias_domain.active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_domain_mailbox_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox,alias_domain WHERE alias_domain.alias_domain='%d' AND mailbox.username=CONCAT('%u','@',alias_domain.target_domain) AND mailbox.active='1' AND alias_domain.active='1'
```

Protect the files:
```bash
sudo chmod 640 /etc/postfix/sql/*.cf
sudo chown root:postfix /etc/postfix/sql/*.cf
```

---

## 6. Configure Dovecot

Instead of making changes across multiple fragmented files, you can configure everything mainly using two files. 

First, back up the original configuration:
```bash
sudo mv /etc/dovecot/dovecot.conf /etc/dovecot/dovecot.conf.bkp
```

### `/etc/dovecot/dovecot.conf`

Create a new `/etc/dovecot/dovecot.conf` and combine all core settings into it:

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
# LOGGING E OTIMIZAÇÃO
################################
log_path = syslog
syslog_facility = local5
auth_verbose = yes
auth_debug_passwords = no
```

### `/usr/local/bin/quota-warning.sh`

Create the executable script that will be triggered when quotas reach the limits:
```bash
sudo cp DOCUMENTS/setup/quota-warning.sh /usr/local/bin/quota-warning.sh
sudo chmod +x /usr/local/bin/quota-warning.sh
sudo chown vmail:vmail /usr/local/bin/quota-warning.sh
```
Make sure to eventually edit the `From:` header inside `/usr/local/bin/quota-warning.sh` to match your main domain (e.g. `postmaster@example.com`).

### `/etc/dovecot/dovecot-sql.conf`

Create `/etc/dovecot/dovecot-sql.conf` for the SQL connection and query settings:

```ini
driver  = mysql
connect = host=localhost dbname=postfix user=postfix password=your_secure_password

# The default_pass_scheme must match the hash format of Go-PostfixAdmin
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
```

---

## 7. Create the User and Email Directory

Create the `vmail` (Virtual Mail) system user, which will own all mailbox files:

```bash
sudo groupadd -g 1001 vmail
sudo useradd -g vmail -u 1001 vmail -d /var/vmail -m
sudo chown -R vmail:vmail /var/vmail
```

---

## 8. Configure System Logging (rsyslog)

To properly capture logs from Postfix (especially if running chrooted) and Dovecot, you need to configure `rsyslog`.

First, edit the main `rsyslog` configuration file to ensure the traditional timestamp format is enabled:

```bash
sudo nano /etc/rsyslog.conf
```

Find the `#### GLOBAL DIRECTIVES ####` section and make sure the `ActionFileDefaultTemplate` line is present and uncommented:

```ini
###########################
#### GLOBAL DIRECTIVES ####
###########################

# Use traditional timestamp format.
# To enable high precision timestamps, comment out the following line.
$ActionFileDefaultTemplate RSYSLOG_TraditionalFileFormat
```

Next, copy the provided `postfix.conf` to the `rsyslog.d` directory:

```bash
sudo cp /etc/rsyslog.d/postfix.conf /etc/rsyslog.d/postfix.conf.bkp
sudo cp DOCUMENTS/setup/postfix.conf /etc/rsyslog.d/postfix.conf
```

This file tells `rsyslog` to listen for Postfix logs and routes Dovecot logs to specific files:
```ini
# /etc/rsyslog.d/postfix.conf
$AddUnixListenSocket /var/spool/postfix/dev/log
$template MFORMAT, "%TIMESTAMP:::date-rfc3164% %hostname% %syslogtag%%msg:::sp-if-no-1st-sp%%msg:::drop-last-lf%\n"

local5.*        -/var/log/dovecot.log
local5.warning;local5.error;local5.crit -/var/log/dovecot-errors.log
```

Restart the `rsyslog` service to apply the changes:

```bash
sudo systemctl restart rsyslog
```

---

## 9. Restart and Validate Services

After making all configurations, restart the services to apply changes:

```bash
sudo systemctl restart postfix dovecot
sudo systemctl enable postfix dovecot
```

Validate if MariaDB support was recognized by Postfix:
```bash
postconf -m | grep mysql
# The output should list "mysql"
```

Validate email delivery by strictly observing system logs:
```bash
sudo tail -f /var/log/mail.log
```
