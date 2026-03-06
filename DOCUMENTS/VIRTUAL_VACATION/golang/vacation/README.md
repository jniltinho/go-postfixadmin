# dovecot-vacation

A Go utility that syncs **PostfixAdmin** vacation settings directly to **Dovecot Sieve** scripts, without relying on an SMTP pipe.

Run via cron, it reads vacation records from MySQL and automatically creates or removes `.dovecot.sieve` files in each user's Maildir.

---

## How it works

```
Cron (*/10 min)
  → dovecot-vacation
  → MySQL: SELECT vacation + mailbox
  → For each user:
      active + within date range → write .dovecot.sieve + compile with sievec
      inactive or out of range  → remove .dovecot.sieve
```

---

## Dependencies

| Module / Tool | Purpose |
|---|---|
| `github.com/go-sql-driver/mysql` | MySQL / MariaDB driver |
| `github.com/spf13/cobra` | CLI |
| `github.com/spf13/viper` | `config.toml` reader |
| `sievec` (system binary) | Compile Sieve scripts |

---

## File structure

```
vacation00/
├── main.go     # CLI wiring (Cobra + Viper)
├── run.go      # Business logic (DB query + Sieve sync)
├── sieve.go    # Vacation struct, Sieve generation and file management
└── Makefile
```

---

## Configuration

The binary reads `/opt/go-postfixadmin/config.toml` by default (shared with Go-Postfixadmin).

```toml
[database]
url = "postfix:password@tcp(127.0.0.1:3306)/postfix?parseTime=True"

[server]
mail_base = "/var/vmail"   # Maildir base path (default: /var/vmail)
```

---

## Build and install

```bash
cd DOCUMENTS/VIRTUAL_VACATION/golang/vacation

make deps    # Install Go dependencies
make build   # Compile and compress with UPX

cp dovecot-vacation /opt/go-postfixadmin/
```

### Makefile targets

| Command | Description |
|---|---|
| `make build` | Compile and compress binary with UPX |
| `make deps` | Install Go dependencies |
| `make clean` | Remove binary |

---

## Crontab

Add to root's crontab (or any user with write access to the Maildirs):

```cron
*/10 * * * * /opt/go-postfixadmin/dovecot-vacation
```

---

## Usage

```bash
# Use default config (/opt/go-postfixadmin/config.toml)
dovecot-vacation

# Use a custom config file
dovecot-vacation --config /path/to/config.toml
```

---

## Generated Sieve script

For each user with an active vacation, the `.dovecot.sieve` file looks like:

```sieve
require ["vacation"];

vacation
  :days 1
  :subject "Out of office"
"I will be back on 10/03. I will reply as soon as possible.";
```

The file is automatically compiled with `sievec` after being written.

---

## Verify in the database

```sql
-- Users with active vacations
SELECT email, subject, activefrom, activeuntil, active
FROM vacation
WHERE active = 1;
```

---

## License

Derived from [PostfixAdmin](https://github.com/postfixadmin/postfixadmin), distributed under the GPL license.
