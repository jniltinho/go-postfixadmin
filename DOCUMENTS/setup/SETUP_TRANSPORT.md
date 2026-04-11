# Transport TCP Server Setup

The `postfixadmin transport server` command starts a TCP server that Postfix queries for transport routing decisions. It replaces the classic Perl `transport.pl` lookup script.

## How It Works

Postfix sends a `GET <subject>` query where `<subject>` is an email address (`user@domain`) or a plain domain. The server resolves the transport using a single SQL query with the following priority:

1. `mailbox.transport` — per-user override
2. `transport_list.transport` — explicit domain rule
3. `domain.transport` — default domain transport

Results are cached in memory (default: 10 minutes) to reduce database load.

---

## 1. Postfix Configuration

Add to `/etc/postfix/main.cf`:

```
transport_maps = tcp:127.0.0.1:12221
```

---

## 2. config.toml

Add or adjust the `[transport]` section in `/opt/go-postfixadmin/config.toml`:

```toml
[transport]
enabled  = true
options  = ["virtual", "local", "relay"]
default  = "virtual"

# Transport TCP server
# Postfix main.cf: transport_maps = tcp:127.0.0.1:12221
host          = "127.0.0.1:12221"
cache         = "10m"
hostname      = "mail.example.com"       # FQDN of this mail server
localdelivery = "smtp:mail.example.com"  # Legacy transport value to rewrite
delivery      = "lmtp:unix:private/dovecot-lmtp" # Rewrite target for local delivery
```

### Field Reference

| Field           | Description                                                                 |
|-----------------|-----------------------------------------------------------------------------|
| `host`          | TCP listen address (default: `127.0.0.1:12221`)                             |
| `cache`         | Cache TTL for resolved transports (e.g. `10m`, `30m`, `1h`)                |
| `hostname`      | FQDN of this mail server — used to detect local delivery                    |
| `localdelivery` | Transport value considered "local" and rewritten (e.g. `smtp:mail.example.com`) |
| `delivery`      | Rewrite target for local delivery (e.g. `lmtp:unix:private/dovecot-lmtp`)  |

---

## 3. systemd Service

Copy the service file and enable it:

```bash
cp /opt/go-postfixadmin/postfixadmin-transport.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now postfixadmin-transport
```

Check status:

```bash
systemctl status postfixadmin-transport
```

View logs:

```bash
tail -f /opt/go-postfixadmin/transport.log
# or via journald
journalctl -u postfixadmin-transport -f
```

---

## 4. Manual Start (Debug)

```bash
cd /opt/go-postfixadmin
./postfixadmin transport server --debug
```

The `--debug` flag enables per-request tracing with colored zerolog output showing the source (`CACHE`, `DB`, `EXTERNAL`), subject, and resolved transport.

---

## 5. Testing

Query the server manually using `postmap`:

```bash
postmap -q "user@example.com" tcp:127.0.0.1:12221
postmap -q "example.com" tcp:127.0.0.1:12221
```

Or with netcat:

```bash
echo "GET user@example.com" | nc 127.0.0.1 12221
```

Expected responses:
- `200 smtp:[relay.example.com]:25` — transport found
- `500 not found` — no transport rule, Postfix uses its default
- `400 internal error` — database error
