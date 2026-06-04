# Development

Local development, build, CLI reference, and Makefile targets for Go-PostfixAdmin.

Related: [Project README](README.md) • [Features](FEATURES.md) • [Complete setup](DOCUMENTS/setup/README.md) • [Transport setup](DOCUMENTS/setup/SETUP_TRANSPORT.md)

## Development Tools

To compile the project locally without Docker, install the following tools:

1. **Go 1.26 or higher**: Main language of the project.
   [Download Go](https://go.dev/dl/)
2. **Node.js 18 or higher**: Required to build the Vue 3 frontend.
   [Download Node.js](https://nodejs.org/)
3. **Make**: Utility for command automation (native on Linux/macOS).
4. **UPX (Optional)**: Used by `make build-prod` to compress the final binary.
   Debian/Ubuntu: `sudo apt install upx-ucl`

## How to Build

This project supports local builds with `make` and containerized builds with Docker.

### Native Build with Makefile

#### Dependency Installation

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies (run once, or after package.json changes)
cd frontend && npm install
```

#### Compilation

```bash
# Build the Vue 3 frontend and compile the Go binary
make build

# Build + compress binary with UPX (for production/distribution)
make build-prod

# Remove generated binary and web/dist
make clean
```

The `make build` target:
1. Runs `cd frontend && npm install && npm run build` → produces `web/dist/`
2. Compiles the Go binary with the frontend embedded via `//go:embed`

### Build with Docker

Use Docker to produce an isolated, production-ready build without installing Go or Node.js locally.

Requirements: Docker installed.

```bash
make build-docker
```

This process:

1. Builds the Vue 3 frontend assets
2. Compiles the Go binary
3. Compresses the binary with UPX
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

### Frontend Development (Hot Reload)

To work on the Vue 3 frontend with hot module replacement:

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server proxies API calls to the Go backend. Start the backend separately:

```bash
./bin/postfixadmin server --port=8080
```

### Generating Swagger Documentation

After adding or changing swag annotations in handler files, regenerate the docs:

```bash
make swagger
```

Then rebuild so the updated docs are embedded in the binary:

```bash
make build
```

The Swagger UI is served at `http://localhost:8080/swagger/` when `server.swagger_enable = true` in `config.toml`.

## Useful Makefile Commands

```bash
make help          # Show available targets
make build         # Frontend (Vue 3) + Go binary (with embed)
make build-prod    # Same + UPX compression (for releases)
make frontend      # Only build Vue 3 → web/dist/
make swagger       # Regenerate OpenAPI docs (then `make build` to embed)
make run           # Build + start server
make clean         # Remove bin/ and web/dist/
make deps / tidy   # Go module maintenance
make deb / rpm     # Build Debian/RPM packages (with embedded config + services)
make build-docker  # Multi-stage Docker image (Alpine + UPX)
make install-upx   # Install UPX (used by build-prod)
```

Run `make help` for the current list. Note: `frontend` step requires Node 18+ and runs `npm install && npm run build` from `frontend/`.

## CLI Overview

Run `./bin/postfixadmin --help` (or the installed binary) for the latest. Current output (as of this workspace):

```text
A command line interface for Go-Postfixadmin application.

Usage:
  postfixadmin [flags]
  postfixadmin [command]

Available Commands:
  admin        Admin management utilities
  backup-mysql MySQL backup utilities
  completion   Generate the autocompletion script for the specified shell
  domain       Domain management utilities
  help         Help about any command
  importsql    Import SQL file to database
  mailbox      Mailbox management utilities
  migrate      Run database migration
  rbac         RBAC role management utilities
  readlog      Read mail.log and import FILTER entries into the maillog table
  server       Start the administration server
  transport    Transport management utilities
  version      Print version information

Flags:
      --config string      config file (default is ./config.toml)
      --db-driver string   Database driver (mysql or postgres)
      --db-url string      Database URL connection string
      --debug              Enable debug output
      --generate-config    Generate a default config.toml file in the current directory
  -h, --help               help for postfixadmin
      --vacation           Sync vacation auto-replies to Dovecot Sieve scripts (for crontab)

Use "postfixadmin [command] --help" for more information about a command.
```

See subcommand help (e.g. `postfixadmin admin --help`, `postfixadmin mailbox --help`, `postfixadmin migrate --help`) for details on each area. The `version` command prints build info.

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

## Domain Commands (CLI)

The `domain` subcommand manages mail domains.

```bash
# List all domains
./postfixadmin domain --list
./postfixadmin domain -l

# Add a new domain
./postfixadmin domain --add "example.com"
./postfixadmin domain -a "example.com" --description "My domain" --max-aliases 50 --max-mailboxes 100

# Delete a domain and all its associated data (mailboxes, aliases, alias domains, fetchmail, vacation)
./postfixadmin domain --delete "example.com"
./postfixadmin domain -d "example.com"
```

Available flags for `domain`:

- `--list` / `-l`: List all domains
- `--add` / `-a`: Create a new domain
- `--delete` / `-d`: Delete a domain and all its associated data
- `--description`: Domain description (applies to `--add`)
- `--max-aliases`: Maximum number of aliases (default: `10`, `0` = unlimited)
- `--max-mailboxes`: Maximum number of mailboxes (default: `10`, `0` = unlimited)

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

## Transport Commands (CLI)

The `transport` subcommand manages transport routing entries and runs the Postfix transport map TCP server.

```bash
# List all transport entries
./postfixadmin transport --list
./postfixadmin transport -l

# Add a transport entry (format: domain:transport)
./postfixadmin transport --add "example.com:smtp:[relay.example.com]:25"
./postfixadmin transport -a "example.com:relay:[mx.provider.com]"

# Delete a transport entry by domain
./postfixadmin transport --delete "example.com"
./postfixadmin transport -d "example.com"

# Start the Postfix transport map TCP server
./postfixadmin transport server
./postfixadmin transport server --debug
```

The `transport server` subcommand starts a TCP server that answers Postfix `transport_maps` lookups. Configure in `config.toml` under `[transport]` — see [SETUP_TRANSPORT.md](DOCUMENTS/setup/SETUP_TRANSPORT.md) for full setup instructions.

Available flags for `transport server`:

- `--debug`: Enable per-request tracing with colored zerolog output (source, subject, resolved transport)

## `readlog` Command

Parses `FILTER:` entries from `/var/log/mail.log` and stores them in the `maillog` database table. It is recommended to schedule it via cron.

Crontab example (poll every 5 minutes):

```cron
## Readlog Postfixadmin
*/5 * * * * /opt/go-postfixadmin/postfixadmin readlog --once --config /opt/go-postfixadmin/config.toml >/tmp/readlog.log 2>&1
```

The `maillog` table stores timestamp, sender, recipient, sender domain, recipient domain, host IP, hostname, HELO string, and message size.
