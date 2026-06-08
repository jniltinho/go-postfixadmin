# Installation Guide: Email Server (Ubuntu) + Go-PostfixAdmin with PostgreSQL

Full email server setup on Ubuntu using **Postfix + Dovecot + PostgreSQL + Go-PostfixAdmin**.

See the MariaDB-focused complete guide and quick summary for most overlapping steps (Postfix/Dovecot config is similar; only DB driver + packages + SQL maps differ).

Related: [main README](../../README.md) • [MariaDB quick](README.md) • [Full MariaDB guide](README.md) (adapt the DB sections).

---

## 1. Update the System and Install Dependencies

On Ubuntu, update your packages and install the necessary basic services for PostgreSQL:

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install postfix postfix-pgsql dovecot-core dovecot-imapd dovecot-pop3d dovecot-lmtpd dovecot-pgsql postgresql -y
sudo apt install certbot git curl -y
```

During Postfix installation, the wizard will ask for the configuration type. Select **"Internet Site"** and enter your main domain (e.g., `example.com`).

---

## 2. Configure the PostgreSQL Database

By default, PostgreSQL creates a `postgres` system user. Let's configure the database using this user:

```bash
sudo -u postgres psql
```

Run the commands below to create the database and the user that Postfix, Dovecot, and Go-PostfixAdmin will use:

```sql
CREATE DATABASE postfix;
CREATE USER postfix WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE postfix TO postfix;
\c postfix
GRANT ALL ON SCHEMA public TO postfix;
\q
```

> **Note:** Remember to replace `your_secure_password` with a strong password in all steps of this guide.

---

## 3. Install and Configure Go-PostfixAdmin

Go-PostfixAdmin will manage the database structure (tables, domains, accounts, aliases, etc.).

1. **Get the Application:**

   Instead of building from source or using the `.tar.gz`, you can download and install the pre-compiled `.deb` or `.rpm` packages from the GitHub Releases page automatically:

   **Ubuntu / Debian (.deb):**
   ```bash
   TAG=$(curl -s https://api.github.com/repos/jniltinho/go-postfixadmin/releases/latest|grep tag_name|cut -d '"' -f4|tr -d v)
   wget https://github.com/jniltinho/go-postfixadmin/releases/latest/download/go-postfixadmin_${TAG}_amd64.deb
   sudo dpkg -i go-postfixadmin_${TAG}_amd64.deb
   ```

   **RHEL / CentOS / RockyLinux (.rpm):**
   ```bash
   TAG=$(curl -s https://api.github.com/repos/jniltinho/go-postfixadmin/releases/latest|grep tag_name|cut -d '"' -f4|tr -d v)
   wget https://github.com/jniltinho/go-postfixadmin/releases/latest/download/go-postfixadmin-${TAG}-1.x86_64.rpm
   sudo rpm -i go-postfixadmin-${TAG}-1.x86_64.rpm
   ```
   
   The package automatically creates the folder structure in `/opt/go-postfixadmin`, copies the default `config.toml`, and installs the systemd service in `/etc/systemd/system/postfixadmin.service`.
   
2. **Generate Initial SSL Certificates (Certbot):**
   
   Before configuring secure server routes, generate the primary certificates. Stop any service using port 80 and run:
   ```bash
   sudo certbot certonly --standalone -d mail.example.com
   ```

3. **Configure the Environment (`config.toml`):**
   Edit the default configuration file automatically provided by the package to add the database and session settings:
   
   ```bash
   sudo nano /opt/go-postfixadmin/config.toml
   ```

   Ensure your settings match your environment:
   
   ```toml
   [database]
   host   = "localhost"
   port   = "5432"
   user   = "postfix"
   pass   = "your_secure_password"
   name   = "postfix"
   driver = "postgres"  # mysql or postgres
   debug  = false    # Set to true to enable verbose GORM SQL logs during troubleshooting

   [server]
   # Web Server Configuration. For SSL use port 443
   port           = 443
   cleanup_maildir = false # Clean up orphaned maildirs when deleting a mailbox
   # SSL Settings for standalone secure server
   ssl_enable = true
   ssl_cert   = "/etc/letsencrypt/live/mail.example.com/fullchain.pem"
   ssl_key    = "/etc/letsencrypt/live/mail.example.com/privkey.pem"
   # Secret session key — generate with: openssl rand -hex 32
   session_secret = "your_super_secret_session_key_here"
   # Set to true only during development or in private networks
   swagger_enable = false

   # JWT tokens for the SPA frontend
   jwt_access_ttl  = "15m"   # short-lived access token
   jwt_refresh_ttl = "168h"  # 7-day refresh token (httpOnly cookie)

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
   # TCP transport server — add to Postfix main.cf:
   #   transport_maps = tcp:127.0.0.1:12221
   host          = "127.0.0.1:12221"
   cache         = "10m"
   hostname      = "mail.example.com"
   localdelivery = "smtp:mail.example.com"
   delivery      = "lmtp:unix:private/dovecot-lmtp"

   [features]
   fetchmail = false

   [rbac]
   # Role-based access control. Enable after running "migrate rbac" and
   # assigning roles to existing admins. When false, all permission checks
   # are no-ops and the system behaves exactly as before RBAC was introduced.
   enabled = false

   [smtp]
   server  = "localhost"
   port    = 25
   type    = "plain" # type: plain | tls | starttls
   ```

   Password policy enforced by the backend:
   - Minimum 8 characters
   - At least 1 uppercase letter
   - At least 1 lowercase letter
   - At least 1 number
   - At least 1 special character

   The web interface password generator already follows the same rules.

4. **Run Migrations:**
   Before starting the service, create the necessary tables by running migrations:
   ```bash
   cd /opt/go-postfixadmin
   ./postfixadmin migrate
   ```

---

## 4. Initial Setup via CLI

After the migration, use the CLI to bootstrap the initial data **before starting the web service**.

```bash
cd /opt/go-postfixadmin
```

### Superadmin

```bash
./postfixadmin admin --add-superadmin "admin@example.com:Password1@"
```

### First domain

```bash
./postfixadmin domain --add "example.com" \
  --description "Main domain" \
  --max-aliases 100 \
  --max-mailboxes 50
```

### First mailbox

```bash
./postfixadmin mailbox --add "user@example.com:Password1@"
```

### Dovecot LMTP transport

Required for local mail delivery. Register the `local` domain pointing to Dovecot's LMTP socket:

```bash
./postfixadmin transport --add "local:lmtp:unix:private/dovecot-lmtp"
```

### Start the web service

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now postfixadmin.service
```

*Check logs with: `tail -f /opt/go-postfixadmin/postfixadmin.log`*

> You can now log in to the web interface with the superadmin credentials created above.

---

## 5. RBAC — Role-Based Access Control

RBAC is **optional** and disabled by default (`rbac.enabled = false`). When enabled it enforces fine-grained permissions on every API endpoint and controls UI visibility based on each admin's assigned roles. The system is fully backward-compatible: existing superadmin and domain admin workflows continue to work unchanged.

### 5.1 Create the RBAC tables and seed built-in roles

Run the dedicated migration **after** `migrate` has already created the main tables:

```bash
cd /opt/go-postfixadmin
./postfixadmin migrate rbac
```

This creates four tables (`rbac_roles`, `rbac_permissions`, `rbac_role_permissions`, `rbac_admin_roles`) and seeds six built-in system roles with their default permission sets:

| Role | Permissions |
|------|-------------|
| `superadmin` | Wildcard — full access to all resources and actions |
| `domain_admin` | Full CRUD on assigned domains, mailboxes, aliases, and alias domains; can view (not edit) advanced domain settings; cannot create new domains |
| `mailbox_admin` | Manage mailboxes and aliases within assigned domains |
| `alias_admin` | Manage aliases and alias domains within assigned domains |
| `viewer` | Read-only access to all resources within assigned scope |
| `report_viewer` | Dashboard statistics and log viewer only |

System roles cannot be deleted or have their permissions changed via the API. Custom roles can be created freely through the web UI or REST API.

### 5.2 Migrate existing admins (first-time RBAC enablement)

Existing admins (created before RBAC was enabled) have no role assignments. Use the CLI to automatically grant the `domain_admin` role to every active non-superadmin that has `domain_admins` entries:

```bash
cd /opt/go-postfixadmin
./postfixadmin rbac seed-existing
```

This operation is idempotent — safe to run multiple times. For PostgreSQL there is no SQL script equivalent included; use the CLI command above.

### 5.3 Assign roles manually

```bash
# Global assignment (applies across all domains the admin manages)
./postfixadmin rbac assign alice@example.com domain_admin

# Domain-scoped assignment (permissions apply only to example.com)
./postfixadmin rbac assign bob@example.com mailbox_admin example.com

# Read-only auditor
./postfixadmin rbac assign auditor@example.com viewer
```

Role assignments can also be managed via the web UI under **Settings → Roles** (requires the `settings:write` permission) or via the REST API (`POST /api/v1/rbac/admins/:username/roles`).

### 5.4 Enable RBAC enforcement

Edit `/opt/go-postfixadmin/config.toml`:

```toml
[rbac]
enabled = true
```

Restart the service to apply:

```bash
sudo systemctl restart postfixadmin
```

> **Important:** Superadmin accounts always bypass all permission checks regardless of `rbac.enabled`. Non-superadmin admins that have `domain_admins` entries but no RBAC role assignments automatically receive the `domain_admin` permission set as a backward-compatibility fallback — so enabling RBAC will not lock out existing admins.

---

## 6. Configure Postfix

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

# Virtual mailboxes (PostgreSQL Integration via Go-PostfixAdmin)
virtual_mailbox_base    = /var/vmail
virtual_mailbox_domains = proxy:pgsql:/etc/postfix/sql/pgsql_virtual_domains_maps.cf
virtual_mailbox_maps    = proxy:pgsql:/etc/postfix/sql/pgsql_virtual_mailbox_maps.cf,
                          proxy:pgsql:/etc/postfix/sql/pgsql_virtual_alias_domain_mailbox_maps.cf
virtual_alias_maps      = proxy:pgsql:/etc/postfix/sql/pgsql_virtual_alias_maps.cf,
                          proxy:pgsql:/etc/postfix/sql/pgsql_virtual_alias_domain_maps.cf,
                          proxy:pgsql:/etc/postfix/sql/pgsql_virtual_alias_domain_catchall_maps.cf

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

## 7. SQL Maps for Postfix

Create the SQL query files in the directory below and ensure proper permissions:

```bash
sudo mkdir -p /etc/postfix/sql
sudo chown root:postfix /etc/postfix/sql
```

> **Warning:** In all the files below, replace `your_secure_password` with the password configured in PostgreSQL.

### `/etc/postfix/sql/pgsql_virtual_domains_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT domain FROM domain WHERE domain='%s' AND active=true
```

### `/etc/postfix/sql/pgsql_virtual_mailbox_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox WHERE username='%s' AND active=true
```

### `/etc/postfix/sql/pgsql_virtual_alias_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias WHERE address='%s' AND active=true
```

### `/etc/postfix/sql/pgsql_virtual_alias_domain_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain='%d' AND alias.address=CONCAT('%u','@',alias_domain.target_domain) AND alias.active=true AND alias_domain.active=true
```

### `/etc/postfix/sql/pgsql_virtual_alias_domain_catchall_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain='%d' AND alias.address=CONCAT('@',alias_domain.target_domain) AND alias.active=true AND alias_domain.active=true
```

### `/etc/postfix/sql/pgsql_virtual_alias_domain_mailbox_maps.cf`
```ini
user     = postfix
password = your_secure_password
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox,alias_domain WHERE alias_domain.alias_domain='%d' AND mailbox.username=CONCAT('%u','@',alias_domain.target_domain) AND mailbox.active=true AND alias_domain.active=true
```

Protect the files:
```bash
sudo chmod 640 /etc/postfix/sql/*.cf
sudo chown root:postfix /etc/postfix/sql/*.cf
```

---

## 8. Configure Dovecot

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
driver  = pgsql
connect = host=localhost dbname=postfix user=postfix password=your_secure_password

# The default_pass_scheme must match the hash format of Go-PostfixAdmin
default_pass_scheme = BLF-CRYPT

# Password validation query
password_query = \
  SELECT username AS user, password \
  FROM mailbox WHERE username='%u' AND active=true

# Home directory and quotas query
user_query = \
  SELECT CONCAT('/var/vmail/', maildir) AS home, \
         1001 AS uid, 1001 AS gid, \
         CONCAT('*:bytes=', quota) AS quota_rule \
  FROM mailbox WHERE username='%u' AND active=true

# Real-time used quota update
iterate_query = SELECT username AS user FROM mailbox WHERE active=true
```

Fix the file permissions to protect the passwords:
```bash
sudo chmod 640 /etc/dovecot/dovecot-sql.conf
sudo chown root:dovecot /etc/dovecot/dovecot-sql.conf
```

---

## 9. Create the User and Email Directory

Create the `vmail` (Virtual Mail) system user, which will own all mailbox files:

```bash
sudo groupadd -g 1001 vmail
sudo useradd -g vmail -u 1001 vmail -d /var/vmail -m
sudo chown -R vmail:vmail /var/vmail
```

---

## 10. Configure System Logging (rsyslog)

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

## 11. Restart and Validate Services

After making all configurations, restart the services to apply changes:

```bash
sudo systemctl restart postfix dovecot
sudo systemctl enable postfix dovecot
```

Validate if PostgreSQL support was recognized by Postfix:
```bash
postconf -m | grep pgsql
# The output should list "pgsql"
```

Validate email delivery by strictly observing system logs:
```bash
sudo tail -f /var/log/mail.log
```

---

## 12. Vacation / Auto-Reply with Dovecot Sieve

Go-PostfixAdmin allows users to configure vacation auto-replies through the web UI. The actual delivery of these replies is handled by **Dovecot Sieve** — a mail filtering language built into Dovecot.

The `dovecot-vacation` daemon (included in this project) bridges the two: it reads vacation settings from PostgreSQL and writes/removes `.dovecot.sieve` scripts in each user's Maildir automatically.

### How It Works

```
Cron (*/10 min)
  → dovecot-vacation
  → PostgreSQL: SELECT vacation + mailbox
  → For each user:
      active + within date range → write .dovecot.sieve + compile with sievec
      inactive or out of range  → remove .dovecot.sieve
```

### 12.1 Enable Sieve in Dovecot

Install the Dovecot Sieve plugin:

```bash
sudo apt install dovecot-sieve dovecot-managesieved -y
mkdir -p /var/lib/dovecot/sieve
touch /var/lib/dovecot/sieve/default.sieve
chown -R vmail:vmail /var/lib/dovecot/sieve
```

Add the following to `/etc/dovecot/dovecot.conf` (or append to the existing `plugin {}` block):

```ini
protocols = imap lmtp pop3

mail_plugins = quota sieve

protocol imap {
  mail_plugins = quota imap_sieve
}

protocol lmtp {
  mail_plugins = $mail_plugins sieve
}

plugin {
  # Quota (already configured in section 6)
  quota = maildir:User quota

  # Sieve: path to the active script per user
  sieve = file:~/Maildir/sieve;active=~/Maildir/.dovecot.sieve
  sieve_global_path = /var/lib/dovecot/sieve/default.sieve
  sieve_before = /var/lib/dovecot/sieve/before.sieve
  sieve_after  = /var/lib/dovecot/sieve/after.sieve
  sieve_vacation_send_from_recipient = yes
  sieve_vacation_min_period = 1h
  sieve_vacation_max_period = 30d
}
```

Restart Dovecot to apply:

```bash
sudo systemctl restart dovecot
```

### 12.2 Install the `dovecot-vacation` Binary

The `dovecot-vacation` binary is part of this project. Build and install it:

```bash
git clone https://github.com/jniltinho/go-postfixadmin.git
cd go-postfixadmin/DOCUMENTS/DOVECOT_VACATION/vacation
make deps
make build
sudo cp dovecot-vacation /opt/go-postfixadmin/
```

Ensure the binary can write to Maildirs:

```bash
sudo chown vmail:vmail /opt/go-postfixadmin/dovecot-vacation
sudo chmod 750 /opt/go-postfixadmin/dovecot-vacation
```

### 12.3 Configuration (`config.toml`)

The binary shares the same `config.toml` as Go-PostfixAdmin. Make sure the following keys are present:

```toml
[database]
host   = "localhost"
port   = "5432"
user   = "postfix"
pass   = "your_secure_password"
name   = "postfix"
driver = "postgres"
debug  = false

[server]
mail_base = "/var/vmail"  # Maildir base path (default: /var/vmail)

[vacation]
enabled = true
```

### 12.4 Schedule via Cron

Add the cron entry to run the sync every 10 minutes as the `vmail` user:

```bash
sudo crontab -u vmail -e
```

Add the line:

```cron
*/10 * * * * /opt/go-postfixadmin/dovecot-vacation
```

> **Note:** The binary must be able to read and write files inside `/var/vmail`. Running it as `vmail` is the safest approach.

### 12.5 Generated Sieve Script

For each mailbox with an active vacation, `dovecot-vacation` writes a `.dovecot.sieve` file like this:

```sieve
require ["vacation"];

vacation
  :days 1
  :subject "Estou de Férias"
"Estarei fora do escritório de 05/03/2026 até 18/03/2026.
Em caso de urgência, por favor contate José Nilton.";
```

- `:days` is derived from `interval_time` (stored in seconds) → `interval_time / 86400`
- If `interval_time` is `0` (reply once), `:days` defaults to `1`
- The script is automatically compiled with `sievec` after being written

### 12.6 Verify

Check if Sieve is active for a specific user:

```bash
ls -la /var/vmail/example.com/user/Maildir/.dovecot.sieve
ls -la /var/vmail/example.com/user/Maildir/.dovecot.svbin  # compiled binary
```

Check active vacation records in the database:

```sql
SELECT email, subject, activefrom, activeuntil, interval_time, active
FROM vacation
WHERE active = true;
```

Run the sync manually to test:

```bash
sudo -u vmail /opt/go-postfixadmin/dovecot-vacation
```

---
