# Installation Guide: Nginx + SOGo Webmail + MySQL (Ubuntu)

This document describes the installation and configuration process for the SOGo webmail using Nginx as a reverse proxy and MySQL (MariaDB) as the database, integrated with your email server (Go-PostfixAdmin / Postfix / Dovecot).

---

## 1. Prerequisites

* Ubuntu/Debian server with Postfix, Dovecot, and MariaDB already installed and functional.
* A configured domain pointing to your server (e.g., `mail.yourdomain.com`).
* Valid SSL/TLS certificates (e.g., Let's Encrypt).

---

## 2. MariaDB Configuration for SOGo

SOGo needs its own database to store user preferences, contacts, and calendars.

Access the MariaDB console:
```bash
sudo mariadb
```

Create the database and user for SOGo. Then, create a View so that SOGo can read the correctly formatted Go-PostfixAdmin users:
```sql
CREATE DATABASE sogo CHARSET='utf8mb4';
CREATE USER 'sogo'@'localhost' IDENTIFIED BY 'YOUR_SECURE_PASSWORD_HERE';
GRANT ALL PRIVILEGES ON sogo.* TO 'sogo'@'localhost';
-- Allows SOGo to read the postfix mailbox table
GRANT SELECT ON postfix.mailbox TO 'sogo'@'localhost';
FLUSH PRIVILEGES;

USE sogo;
CREATE OR REPLACE VIEW sogo_users_view AS
SELECT
    username AS c_uid,             -- The full email (user@domain.com)
    username AS c_name,            -- Internal name
    name AS c_cn,                  -- Display name (Common Name)
    password AS c_password,        -- Password hash
    username AS mail               -- Email field for searches
FROM postfix.mailbox
WHERE active = 1;

EXIT;
```

---

## 3. SOGo Installation

Add the SOGo key and repository for your Ubuntu/Debian version (make sure to reference the correct repository for your distro, example below is for Ubuntu):

```bash
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-key 74FFC6D72B925A01
sudo add-apt-repository "deb [arch=amd64] https://packages.inverse.ca/SOGo/nightly/5/ubuntu/ $(lsb_release -sc) $(lsb_release -sc)"
sudo apt update
```

Install SOGo and memcached (required for sessions):
```bash
sudo apt install sogo sogo-activesync memcached -y
```

---

## 4. SOGo Configuration (`/etc/sogo/sogo.conf`)

Backup the original configuration:
```bash
sudo cp /etc/sogo/sogo.conf /etc/sogo/sogo.conf.bkp
```

Edit the `/etc/sogo/sogo.conf` file. Clear the original file (which comes with many comments) and use this base configuration adapting it to your reality:

```ini
{
  /* Database configuration (MySQL SOGo) */
  SOGoProfileURL = "mysql://sogo:YOUR_SECURE_PASSWORD_HERE@localhost:3306/sogo/sogo_user_profile";
  OCSFolderInfoURL = "mysql://sogo:YOUR_SECURE_PASSWORD_HERE@localhost:3306/sogo/sogo_folder_info";
  OCSSessionsFolderURL = "mysql://sogo:YOUR_SECURE_PASSWORD_HERE@localhost:3306/sogo/sogo_sessions_folder";
  OCSAdminURL = "mysql://sogo:YOUR_SECURE_PASSWORD_HERE@localhost:3306/sogo/sogo_admin";

  /* Mail */
  SOGoDraftsFolderName = Drafts;
  SOGoSentFolderName = Sent;
  SOGoTrashFolderName = Trash;
  SOGoJunkFolderName = Junk;
  SOGoIMAPServer = "localhost";
  SOGoSMTPServer = "smtp://localhost";
  SOGoMailDomain = yourdomain.com;
  SOGoMailingMechanism = smtp;

  /* Authentication (using the view created in the SOGo database with the Go-PostfixAdmin data) */
  SOGoUserSources = (
    {
      type = sql;
      id = directory;
      viewURL = "mysql://sogo:YOUR_SECURE_PASSWORD_HERE@localhost:3306/sogo/sogo_users_view";
      canAuthenticate = YES;
      isAddressBook = YES;
      /* Must match the format used in Go-Postfixadmin (Dovecot) e.g., blf-crypt (Bcrypt) */
      userPasswordAlgorithm = blf-crypt; 
    }
  );

  /* Web Interface */
  SOGoPageTitle = "SOGo Webmail";
  SOGoVacationEnabled = YES;
  SOGoForwardEnabled = YES;
  SOGoSieveScriptsEnabled = YES;
  SOGoMailAuxiliaryUserAccountsEnabled = YES;

  /* General */
  SOGoLanguage = English;
  SOGoTimeZone = America/Sao_Paulo;
  SOGoSuperUsernames = ("admin@yourdomain.com");
  SOGoMemcachedHost = "127.0.0.1";

  /* Workers */
  WOWorkersCount = 3;
}
```

*Note: Replace `YOUR_SECURE_PASSWORD_HERE`, `POSTFIX_USER_PASSWORD`, and `yourdomain.com` with the real values from your environment.*

After configuring, restart the services:
```bash
sudo systemctl restart memcached sogo
sudo systemctl enable memcached sogo
```

---

## 5. Nginx Installation and Configuration

Install Nginx if it's not already installed:
```bash
sudo apt install nginx -y
```

Create the configuration file for SOGo in Nginx:
```bash
sudo nano /etc/nginx/sites-available/sogo.conf
```

Basic Virtual Host (Reverse Proxy) configuration example with SSL:

```nginx
server {
    listen 80;
    server_name mail.yourdomain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name mail.yourdomain.com;

    # SSL Certificates (Example via Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/mail.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mail.yourdomain.com/privkey.pem;

    root /usr/lib/GNUstep/SOGo/WebServerResources/;

    location = / {
        rewrite ^ https://$server_name/SOGo;
        allow all;
    }

    # SOGo Proxy
    location ^~ /SOGo {
        proxy_pass http://127.0.0.1:20000;
        proxy_redirect http://127.0.0.1:20000 default;
        
        # Required headers
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $host;
        proxy_set_header x-webobjects-server-protocol HTTP/1.0;
        proxy_set_header x-webobjects-remote-host 127.0.0.1;
        proxy_set_header x-webobjects-server-name $server_name;
        proxy_set_header x-webobjects-server-url https://$server_name;
        proxy_set_header x-webobjects-server-port 443;
        
        # Upload limits adjustments
        client_max_body_size 50M;
        client_body_buffer_size 128k;
        break;
    }

    # SOGo Static files
    location /SOGo.woa/WebServerResources/ {
        alias /usr/lib/GNUstep/SOGo/WebServerResources/;
        allow all;
        expires max;
    }

    location /SOGo/WebServerResources/ {
        alias /usr/lib/GNUstep/SOGo/WebServerResources/;
        allow all;
        expires max;
    }

    # ActiveSync (optional, for mobile/Outlook clients)
    location ^~ /Microsoft-Server-ActiveSync {
        proxy_pass http://127.0.0.1:20000/SOGo/Microsoft-Server-ActiveSync;
        proxy_redirect http://127.0.0.1:20000/Microsoft-Server-ActiveSync /;
        
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $server_name;
        proxy_set_header x-webobjects-server-protocol HTTP/1.0;
        proxy_set_header x-webobjects-remote-host 127.0.0.1;
        proxy_set_header x-webobjects-server-name $server_name;
        proxy_set_header x-webobjects-server-url https://$server_name;
        proxy_set_header x-webobjects-server-port 443;
        
        proxy_connect_timeout 75;
        proxy_send_timeout 3600;
        proxy_read_timeout 3600;
        proxy_buffer_size 128k;
        proxy_buffers 64 256k;
        proxy_busy_buffers_size 256k;
        proxy_temp_file_write_size 256k;
        client_max_body_size 0;
        client_body_buffer_size 128k;
    }
}
```

Enable the site in Nginx by setting up the symbolic link and reloading the service:
```bash
sudo ln -s /etc/nginx/sites-available/sogo.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 6. Customizing the SOGo Theme

SOGo relies on the [theming system](https://material.angularjs.org/latest/Theming/01_introduction) of Angular Material to define the colors of the Web interface.

To overwrite the default theme, add the following parameter inside the configuration block in `/etc/sogo/sogo.conf`:

```ini
  SOGoUIAdditionalJSFiles = (js/theme.js);
```

Then, edit or create `theme.js` under `/usr/lib/GNUstep/SOGo/WebServerResources/js/` (or `/usr/lib64/GNUstep/SOGo/WebServerResources/js/` depending on your platform) and restart `sogo`.

### Generating a new stylesheet

If the configuration parameter `SOGoUIxDebugEnabled` is unset or set to `NO` in `/etc/sogo/sogo.conf`, you will need to generate a new `theme-default.css` stylesheet for the new theme:

1. Temporarily set `SOGoUIxDebugEnabled = YES;` in `/etc/sogo/sogo.conf`.
2. Restart the SOGo service:
   ```bash
   sudo systemctl restart sogo
   ```
3. Access the webmail in your browser and verify the theme adjustments.
4. From your favorite browser, open the Developer Tools JavaScript console and type the following:
   ```javascript
   copy([].slice.call(document.styleSheets)
     .map(e => e.ownerNode)
     .filter(e => e.hasAttribute('md-theme-style'))
     .map(e => e.textContent)
     .join('\n')
   )
   ```
5. Overwrite the content of `/usr/lib/GNUstep/SOGo/WebServerResources/css/theme-default.css` with the content of your clipboard.
6. Restore the value of `SOGoUIxDebugEnabled` in `/etc/sogo/sogo.conf` (set to `NO` or remove it).
7. Restart the SOGo service again:
   ```bash
   sudo systemctl restart sogo
   ```

---

## 7. Testing Access

1. In your browser, go to `https://mail.yourdomain.com`.
2. The SOGo login screen should appear.
3. Enter with a complete email address created in your **Go-PostfixAdmin** (e.g., `user@yourdomain.com`) and its respective password.

### Additional Tips

- **SOGo Logs**: Check `/var/log/sogo/sogo.log` in case of 502 errors or mysterious login failures.
- **Go-PostfixAdmin Integration**: Any password change, new mailbox creation, or suspension done through *Go-PostfixAdmin* will automatically impact access to SOGo, since we use the same MariaDB database (`postfix`) for credentials.
