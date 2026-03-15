# backup-mysql

Automatic MySQL backup tool written in Go.
Dumps each database to a `.sql.gz` file using `mysqldump` + `gzip`,
with support for old file cleanup and e-mail log delivery.

## Requirements

- Go 1.22+
- `mysql` and `mysqldump` installed on the system

## Installation

```bash
make install
```

The binary will be installed at `/usr/local/bin/backup-mysql`.

## Configuration

Priority order: **CLI flag > environment variable > built-in default**.

### Environment variables

| Variable      | Default                       | Description                          |
|---------------|-------------------------------|--------------------------------------|
| `MYSQL_HOST`  | `localhost`                   | MySQL host address                   |
| `MYSQL_USER`  | `root`                        | MySQL username                       |
| `MYSQL_PASS`  | `root`                        | MySQL password                       |
| `BACKUP_DIR`  | `/usr/local/backup/mysql`     | Directory where backups are stored   |
| `LOG_FILE`    | `/var/log/backup_mysql.log`   | Log file path                        |
| `SMTP_SERVER` | `smtp.dominio.com.br`         | SMTP server                          |
| `SMTP_PORT`   | `587`                         | SMTP port                            |
| `SMTP_USER`   | `email@dominio.com.br`        | SMTP username                        |
| `SMTP_PASS`   | `senha_email`                 | SMTP password                        |
| `EMAIL_FROM`  | `email@dominio.com.br`        | Sender e-mail address                |
| `EMAIL_TO`    | `email1@dominio_x.com.br`     | Recipients (comma-separated)         |
| `EMAIL_CC`    | `admin@dominio_y.com.br`      | CC recipients (comma-separated)      |

## Commands

### `list` — list all databases

```bash
backup-mysql list
backup-mysql --host=192.168.1.10 --user=admin --passwd=secret list
```

### `backup` — backup all databases

```bash
# Basic backup
backup-mysql backup

# With terminal output
backup-mysql backup --debug

# Remove files older than 5 days after backup
backup-mysql backup --clean=5

# Send log by e-mail when done
backup-mysql backup --sendmail

# All options combined
backup-mysql backup --debug --clean=5 --sendmail

# Override credentials via flags
backup-mysql --host=192.168.1.10 --user=admin --passwd=secret backup --clean=7
```

### `version` — show version and build date

```bash
backup-mysql version
```

## Flags

### Global (apply to all subcommands)

| Flag        | Default     | Description        |
|-------------|-------------|--------------------|
| `--host`    | `localhost` | MySQL host address |
| `--user`    | `root`      | MySQL username     |
| `--passwd`  | `root`      | MySQL password     |

### `backup`

| Flag         | Default | Description                                     |
|--------------|---------|-------------------------------------------------|
| `--clean`    | `0`     | Remove `.gz` files older than N days (0 = off) |
| `--debug`    | `false` | Print log to terminal                           |
| `--sendmail` | `false` | Send log by e-mail when done                    |

## Crontab

Cron runs in a minimal environment — it does **not** load `~/.bashrc` or
`~/.profile`, so shell `export` statements from those files are ignored.

### Option 1 — declare variables directly in crontab

Variables declared before the job lines are applied to all jobs in the file:

```cron
MYSQL_HOST=localhost
MYSQL_USER=admin
MYSQL_PASS=s3cr3t
BACKUP_DIR=/mnt/backups/mysql
SMTP_SERVER=smtp.gmail.com
SMTP_USER=alerts@company.com
SMTP_PASS=app_password
EMAIL_FROM=alerts@company.com
EMAIL_TO=dba@company.com,ops@company.com

# Daily backup at 01:05, keeping only the last 5 days
05 01 * * * /usr/local/bin/backup-mysql backup --clean=5
```

### Option 2 — env file with `env`

Store variables in `/etc/backup-mysql.env`:

```bash
MYSQL_HOST=localhost
MYSQL_USER=admin
MYSQL_PASS=s3cr3t
BACKUP_DIR=/mnt/backups/mysql
```

Reference it in crontab:

```cron
05 01 * * * env $(cat /etc/backup-mysql.env | xargs) /usr/local/bin/backup-mysql backup --clean=5
```

### Option 3 — wrapper script

Create `/usr/local/bin/backup-mysql-run`:

```bash
#!/bin/bash
set -a
source /etc/backup-mysql.env
set +a
exec /usr/local/bin/backup-mysql "$@"
```

```bash
chmod +x /usr/local/bin/backup-mysql-run
```

```cron
05 01 * * * /usr/local/bin/backup-mysql-run backup --clean=5
```

## Build

```bash
make build          # compile and compress with upx
make install        # install to /usr/local/bin
make clean          # remove binary
make fmt            # format source code
make vet            # static analysis
```

To set a custom version:

```bash
make build VERSION=1.2.0
```

## License

LGPLv3 — <http://www.gnu.org/licenses/lgpl.html>
