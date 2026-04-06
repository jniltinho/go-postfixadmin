# Development

## What This Document Covers

This file covers local development prerequisites, build options, and the most useful Makefile commands.

Related documents:

- [Project README](README.md)
- [Features](FEATURES.md)
- [Quick setup summary (MariaDB/MySQL)](DOCUMENTS/setup/SETUP_MAILSERVER_MARIADB.md)
- [Quick setup summary (PostgreSQL)](DOCUMENTS/setup/SETUP_MAILSERVER_POSTGRESQL.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)

## Development Tools

To compile the project locally without Docker, install the following tools:

1. **Go (v1.26 or higher)**: Main language of the project.
   [Download Go](https://go.dev/dl/)
2. **Tailwind CSS standalone binary**: Required for CSS processing. Install via `make install-tailwind` (no Node.js required).
3. **Make**: Utility for command automation (native on Linux/macOS).
4. **UPX (Optional)**: Used by the Makefile to compress the final binary.
   Debian/Ubuntu: `sudo apt install upx-ucl rpm`

## How to Build

This project supports local builds with `make` and containerized builds with Docker.

### Native Build with Makefile

The local build automates CSS generation and Go binary compilation.

#### Dependency Installation

Recommended:

```bash
make deps
```

Manual installation:

```bash
go mod download
make install-tailwind
```

#### Compilation

```bash
# Generate CSS and compile the binary
make build-prod

# Remove generated files
make clean
```

### Build with Docker

Use Docker to produce an isolated, production-ready build without installing Go or Node.js locally.

Requirements: Docker installed.

```bash
make build-docker
```

This process:

1. Compiles static assets
2. Compiles the Go binary
3. Compresses the binary with `upx`
4. Produces a final Alpine-based image

### Quick Start with Docker Compose

The fastest way to start a full local environment with MariaDB and Go-PostfixAdmin.

Requirements: Docker and Docker Compose installed.

```bash
make build-docker
docker compose up
```

This will:

- Start a MariaDB container
- Build and start the Go-PostfixAdmin container
- Wait for the database to become ready
- Run database migrations automatically
- Create an initial superadmin (`admin@example.com` / `adminpassword`)

You can customize ports and environment variables in `docker-compose.yml`.

## Useful Makefile Commands

| Command | Description |
| :--- | :--- |
| `make build-prod` | Build CSS and compile the local binary |
| `make build-docker` | Build the optimized Docker image |
| `make run` | Compile and start the server locally |
| `make watch-css` | Start the Tailwind watcher for UI development |
| `make clean` | Remove generated binary and CSS files |
| `make tidy` | Clean and organize Go dependencies |
| `make deps` | Install required dependencies |
| `make deb` | Build the Debian (.deb) package |
| `make rpm` | Build the RPM (.rpm) package |

## 💻 CLI Flags

Below are the available flags when running the `./postfixadmin` binary:

```text
A command line interface for the Go-PostfixAdmin application.

Usage:
  postfixadmin [command]

Available Commands:
  admin        Admin management utilities
  backup-mysql MySQL backup utilities
  completion   Generate the autocompletion script for the specified shell
  help         Help about any command
  importsql    Import SQL file to database
  mailbox      Mailbox management utilities
  migrate      Run database migration
  readlog      Parse mail.log and store FILTER entries in the maillog table
  server       Start the administration server
  version      Display version information

Flags:
      --config string      config file (default is ./config.toml)
      --db-driver string   Database driver (mysql or postgres)
      --db-url string      Database URL connection string
      --debug              Enable debug output
      --generate-config    Generate a default config.toml file in the current directory
  -h, --help               help for postfixadmin

Use "postfixadmin [command] --help" for more information about a command.
```

## Administration Commands (CLI)

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

- `--list-aliases` (`-s`): List all aliases
- `--list-alias-domains` (`-S`): List all alias domains
- `--domain-admins` (`-A`): List all domain admins
- `--list-logs` (`-L`): List all system logs (last 100)
- `--list-maillog`: List mail filter log entries (last 100 by default)
- `--maillog-domain`: Filter maillog output by domain
- `--maillog-limit`: Number of maillog entries to show (default `100`)
- `--cleanup-maildir` (`-c`): Clean up orphaned maildirs on the server
- `--quota-report` (`-q`): Fetch and display the Dovecot quota report
- `--email` (`-e`): Optionally send the quota report via sendmail to this address (when combined with `-q`)
- `--base-dir`: Base directory for maildirs (default `/var/vmail`)

## Mailbox Commands (CLI)

The `mailbox` subcommand manages mailbox users:

```bash
# List all mailboxes
./postfixadmin mailbox --list
./postfixadmin mailbox -l

# Create a new mailbox user
./postfixadmin mailbox --add "user@example.com:password123"
# Or use the short flag, leaving the password blank to generate a random one
./postfixadmin mailbox -a "user@example.com"
# Set a custom quota (in MB)
./postfixadmin mailbox -a "user@example.com" --quota 500

# Import users from a CSV file
./postfixadmin mailbox --import-csv users.csv
# Import with a custom quota
./postfixadmin mailbox --import-csv users.csv -q 250
# Import with pre-hashed passwords (skip hashing)
./postfixadmin mailbox --import-csv users.csv --password-crypt

# Export all mailboxes to a CSV file
./postfixadmin mailbox --export backup.csv
./postfixadmin mailbox -e backup.csv
```

CSV format (`user,password,domain,name`):

```csv
user,password,domain,name
john,Password&123,example.com,John Doe
jane,Password&123,example.com,Jane Doe
```

When using `--password-crypt`, the `password` column must contain a valid bcrypt hash:

```csv
user,password,domain,name
john,$2y$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012345,example.com,John Doe
jane,$2y$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ678901,example.com,Jane Doe
```

The `name` column is optional — if empty, the local part is used with an uppercase first letter.

The export format (`--export`) includes two additional columns:

```csv
user,password,domain,name,quota_mb,active
john,$2y$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012345,example.com,John Doe,100,1
jane,$2y$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ678901,example.com,Jane Doe,500,1
```

- `quota_mb`: quota in MB (the `quota_mb` and `active` columns are ignored by `--import-csv`, which uses `--quota` instead)

Available flags for `mailbox`:

- `--list` / `-l`: List all mailboxes
- `--add` / `-a`: Create a new mailbox user (format: `email:password`)
- `--quota` / `-q`: Mailbox quota in MB (default: `100`). Applies to `--add` and `--import-csv`. The final stored value is multiplied by `quota.multiplier` in `config.toml`
- `--import-csv`: Import mailbox users from a CSV file. Existing mailboxes and rows with passwords shorter than 8 characters are skipped
- `--password-crypt`: Use with `--import-csv`. Treats the `password` column as an already-hashed value and stores it directly, skipping hashing and length validation. Useful when migrating from another system that exports bcrypt hashes
- `--export` / `-e`: Export all mailboxes to a CSV file (columns: `user,password,domain,name,quota_mb,active`). The `password` column contains the stored bcrypt hash, making the output directly compatible with `--import-csv --password-crypt` for backup and migration workflows

## Backup MySQL Commands (CLI)

The `backup-mysql` subcommand backs up and inspects MySQL/MariaDB databases without requiring a database URL — it calls `mysql` and `mysqldump` directly.

Configuration is read in priority order: `config.toml [backup]` section → environment variables → built-in defaults. CLI flags always override.

```bash
# List all databases and their sizes
./postfixadmin backup-mysql list

# Backup all databases (creates .sql.gz files in backup_dir)
./postfixadmin backup-mysql backup

# Backup with cleanup of files older than 7 days and verbose output
./postfixadmin backup-mysql backup --clean 7 --verbose

# Backup and send the log by e-mail
./postfixadmin backup-mysql backup --sendmail

# Override connection details at runtime
./postfixadmin backup-mysql --host db.example.com --user root --passwd secret backup --clean 30
```

Persistent flags (available to all `backup-mysql` subcommands):

- `--host`: MySQL host address (overrides `backup.mysql_host` in config)
- `--port`: MySQL port (overrides `backup.mysql_port` in config, default `3306`)
- `--user`: MySQL username (overrides `backup.mysql_user` in config)
- `--passwd`: MySQL password (overrides `backup.mysql_pass` in config)

Flags for `backup-mysql backup`:

- `--clean N`: Remove `.sql.gz` backup files older than N days (0 = disabled)
- `--verbose`: Print log output to stdout in addition to the log file
- `--sendmail`: Send the backup log by e-mail after completion

Environment variables (fallback when config key is absent):

| Variable | Description |
| :--- | :--- |
| `MYSQL_HOST` | MySQL host |
| `MYSQL_USER` | MySQL username |
| `MYSQL_PASS` | MySQL password |
| `BACKUP_DIR` | Directory for `.sql.gz` files (default `/usr/local/backup/mysql`) |
| `LOG_FILE` | Path to the log file (default `/var/log/backup_mysql.log`) |
| `SMTP_SERVER` | SMTP host for `--sendmail` |
| `SMTP_PORT` | SMTP port (default `587`) |
| `EMAIL_FROM` | Sender address |
| `EMAIL_TO` | Comma-separated recipient list |
| `EMAIL_CC` | Comma-separated CC list |

Crontab example (daily backup at 02:00, keep 14 days):

```cron
0 2 * * * /opt/go-postfixadmin/postfixadmin --config /opt/go-postfixadmin/config.toml backup-mysql backup --clean 14 --sendmail
```

### `readlog` Command

Parses `FILTER:` entries from `/var/log/mail.log` and stores them in the `maillog` database table. It is recommended to schedule it via cron.

Crontab example (poll every 5 minutes):

```cron
## Readlog Postfixadmin
*/5 * * * * /opt/go-postfixadmin/postfixadmin readlog --once --config /opt/go-postfixadmin/config.toml >/tmp/readlog.log 2>&1
```

The `maillog` table stores timestamp, sender, recipient, sender domain, recipient domain, host IP, hostname, HELO string, and message size.
