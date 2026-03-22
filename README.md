# Go-PostfixAdmin

Professional Email Administration System built with Go, Echo, and Tailwind CSS.

## What This Document Covers

This is the main entry point for the project documentation. It covers the product overview, build options, runtime basics, and links to the detailed guides.

## Overview

Go-PostfixAdmin provides a web UI and CLI for managing domains, mailboxes, aliases, admins, vacation rules, and operational mail logs in Postfix/Dovecot environments.

For the full feature list and capability breakdown, see [FEATURES.md](FEATURES.md).

## Documentation

- [Features](FEATURES.md)
- [Development guide](DEVELOPMENT.md)
- [Quick mail server setup](SETUP_MAILSERVER.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)
- [Contributing](CONTRIBUTING.md)

---

## 🚀 Execution

After building, you can run the binary directly:

```bash
./postfixadmin server --port=8080
```

Or via Docker:

```bash
docker run -p 8080:8080 -e DB_URL="your-dsn" postfixadmin:latest
```

### Deployment with Systemd (Linux)

To deploy the application natively on a Linux server, you can use the included Systemd service file.

The pre-configured file is located at `DOCUMENTS/setup/postfixadmin.service`. It expects the application to be placed in the `/opt/go-postfixadmin` directory and will read environment variables from a `config.toml` file in this same directory.

**Service Installation:**

```bash
# 1. Copy the file to the systemd services directory
sudo cp DOCUMENTS/setup/postfixadmin.service /etc/systemd/system/

# 2. Reload systemd configurations
sudo systemctl daemon-reload

# 3. Enable the service to run on boot
sudo systemctl enable postfixadmin.service

# 4. Start the service
sudo systemctl start postfixadmin.service

# 5. Monitor logs in real-time
# The service directs output to the postfixadmin.log file
tail -f /opt/go-postfixadmin/postfixadmin.log
```

---

## 💻 CLI Flags

Below are the available flags when running the `./postfixadmin` binary:

```text
A command line interface for the Go-PostfixAdmin application.

Usage:
  postfixadmin [command]

Available Commands:
  admin       Admin management utilities
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  importsql   Import SQL file to database
  migrate     Run database migration
  readlog     Parse mail.log and store FILTER entries in the maillog table
  server      Start the administration server
  version     Display version information

Flags:
      --config string      config file (default is ./config.toml)
      --db-driver string   Database driver (mysql or postgres)
      --db-url string      Database URL connection string
      --debug              Enable debug output
      --generate-config    Generate a default config.toml file in the current directory
  -h, --help               help for postfixadmin

Use "postfixadmin [command] --help" for more information about a command.
```

### Administration Commands (CLI)

The binary also supports direct administrative commands via the `admin` subcommand:

```bash
# List all administrators
./postfixadmin admin --list-admins

# List all domains
./postfixadmin admin --list-domains

# Create a new Superadmin (useful for first access)
./postfixadmin admin --add-superadmin "admin@example.com:password123"
# Or leave the password blank to generate a random one
./postfixadmin admin --add-superadmin "admin@example.com"
```

Other available flags for `admin`:
*   `--list-mailboxes` (`-m`): List all mailboxes.
*   `--list-aliases` (`-s`): List all aliases.
*   `--list-alias-domains` (`-S`): List all alias domains.
*   `--domain-admins` (`-A`): List all domain admins.
*   `--list-logs` (`-L`): List all system logs (last 100).
*   `--list-maillog`: List mail filter log entries (last 100 by default).
*   `--maillog-domain`: Filter maillog output by domain.
*   `--maillog-limit`: Number of maillog entries to show (default `100`).
*   `--cleanup-maildir` (`-c`): Clean up orphaned maildirs on the server.
*   `--quota-report` (`-q`): Fetches and displays Dovecot quota report.
*   `--email` (`-e`): Optionally send the quota report via sendmail to this address (when combined with `-q`).
*   `--base-dir`: Base directory for maildirs (default "/var/vmail").

#### readlog command

Parses `FILTER:` entries from `/var/log/mail.log` and stores them in the `maillog` database table. Runs continuously, polling every 5 minutes:

```bash
./postfixadmin readlog
```

The `maillog` table stores: timestamp, sender, recipient, sender domain, recipient domain, host IP, hostname, HELO string, and message size.

---

## 📸 Screenshots

![Go-PostfixAdmin Login Screen](DOCUMENTS/screenshots/postfixadmin_01.png)

Check out more images in the [screenshots](DOCUMENTS/screenshots) folder.

---

## 📖 Installation and Configuration Guide

For complete step-by-step instructions on how to set up an email server on Ubuntu with Postfix, Dovecot, MariaDB, and integrate it with Go-PostfixAdmin, see our [Complete Setup Guide](DOCUMENTS/setup/README.md).

You can also find our guide for setting up a complete webmail environment with [Nginx, SOGo, and MariaDB here](DOCUMENTS/setup/nginx-sogo-mysql.md).
