# Go-Postfixadmin

Professional Email Administration System built with Go, Echo, and Tailwind CSS.

## ✨ Features

*   **Complete Management**: Domains, Mailboxes, and Aliases.
*   **Role-Based Access Control (RBAC)**: Differentiation between Superadmins and Domain Admins.
*   **Modern Design**: Clean and responsive interface built with Tailwind CSS.
*   **Security**: Strong password hashing and protection against common attacks.
*   **Password Policy Enforcement**: Backend validation requires at least 8 characters, including uppercase, lowercase, number, and special character.
*   **Frontend Password Generation**: Password generation is handled entirely in JavaScript and produces values compatible with the backend password policy.
*   **Integrated CLI**: Command-line tools for automation and access recovery.
*   **Welcome Emails**: Optional automatic welcome emails sent upon mailbox creation.
*   **Auto-Reply (Vacation)**: Integrated Dovecot Sieve script generation for automatic vacation responses.
*   **Mail Log Viewer**: Built-in admin panel page to browse Postfix filter log entries (`/maillog`) with server-side DataTables pagination, search, and sorting.
*   **Mail Log Reader (`readlog`)**: CLI daemon that parses `FILTER:` entries from `/var/log/mail.log` and stores them in the `maillog` database table (runs every 5 minutes via internal ticker).
*   **Internationalization (i18n)**: Multi-language support (PT, EN, ES) powered by [gotext](https://github.com/leonelquinteros/gotext) with GNU Gettext `.po` files and dynamic locale loading.


## 🛠 Development Tools

To compile the project locally (without Docker), you will need to install the following tools:

1.  **Go (v1.21 or higher)**: Main language of the project.
    *   [Download Go](https://go.dev/dl/)
2.  **Node.js (v20 or higher)**: Required for CSS processing with Tailwind.
    *   [Download Node.js](https://nodejs.org/)
3.  **Make**: Utility for command automation (native on Linux/macOS).
4.  **UPX (Optional)**: Used by the Makefile to compress the final binary.
    *   `sudo apt install upx-ucl` (Debian/Ubuntu)

---

## 🏗 How to Build

This project offers two main ways to build: using `make` (local) or `docker`.

### 1. Native Build with Makefile

The local build automates CSS generation and Go binary compilation.

#### Dependency Installation

To install all dependencies (Recommended):

```bash
make deps
```

If you prefer to install manually:

```bash
go mod download
npm install
```

### Compilation
```bash
# Generate CSS and compile the binary
make build-prod

# To clean generated files
make clean
```

### 2. Build with Docker

Ideal for generating an isolated, production-ready final version without needing to install Go or Node.js on your machine.

**Requirements:** Docker installed.

```bash
# Generate the professional docker image (optimized to ~14MB)
make build-docker
```

This command runs a multi-stage build that:
1.  Compiles static assets (Tailwind).
2.  Compiles the Go binary (Generates a static binary).
3.  Compresses the binary with `upx`.
4.  Generates a final image based on Alpine Linux.

### 3. Quick Start with Docker Compose

The fastest way to get a full environment running (MariaDB + Go-Postfixadmin).

**Requirements:** Docker and Docker Compose installed.

```bash
make build-docker
docker-compose up
```

This will:
- Start a MariaDB container.
- Build and start the Go-Postfixadmin container.
- **Automatically** wait for the DB to be ready.
- **Automatically** run database migrations.
- **Automatically** create an initial superadmin (default: `admin@example.com` / `adminpassword`).

You can customize the environment variables and port mappings in the `docker-compose.yml` file.

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

### DB_URL Examples

**MariaDB:**
```bash
# Standard format
DB_URL="user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

**PostgreSQL:**
```bash
DB_URL="host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=America/Sao_Paulo"
```

### Database Debug Logging

You can enable verbose GORM SQL logging in `config.toml`:

```toml
[database]
url = "postfix:postfixPassword@tcp(mysql:3306)/postfix?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local"
debug = true
```

This flag enables full database query logging independently of the global `--debug` CLI flag.

### Password Rules

The backend validates all mailbox, admin, and user password changes using the same rules:

- At least 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

The UI password generator already follows the same policy, so generated passwords are accepted by the backend without additional adjustment.

### 4. Deployment with Systemd (Linux)

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
A command line interface for Go-Postfixadmin application.

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

## 💻 Useful Makefile Commands

| Command | Description |
| :--- | :--- |
| `make build-prod` | Compiles CSS and the local binary |
| `make build-docker` | Generates the optimized Docker image |
| `make run` | Compiles and starts the server locally |
| `make watch-css` | Starts the Tailwind watcher for UI development |
| `make clean` | Removes the generated binary and CSS files |
| `make tidy` | Cleans and organizes Go dependencies |
| `make deps` | Installs all required dependencies |

---

## 📸 Screenshots

![Go-Postfixadmin Login Screen](DOCUMENTS/screenshots/postfixadmin_01.png)

Check out more images in the [screenshots](DOCUMENTS/screenshots) folder.

---

## 📖 Installation and Configuration Guide

For complete step-by-step instructions on how to set up an email server on Ubuntu with Postfix, Dovecot, MariaDB, and integrate it with Go-PostfixAdmin, see our [Complete Setup Guide](DOCUMENTS/setup/README.md).

You can also find our guide for setting up a complete webmail environment with [Nginx, SOGo, and MariaDB here](DOCUMENTS/setup/nginx-sogo-mysql.md).
