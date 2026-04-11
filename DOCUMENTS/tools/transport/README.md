# Postfix TCP Transport Service

A high-performance Go-based TCP server that replaces the traditional `transport.pl` script. Features in-memory caching, TOML configuration via Viper, Cobra CLI, and GORM support for MariaDB/MySQL.

## Architecture

This service acts as a socket-listening daemon implementing the [Postfix TCP Table protocol](http://www.postfix.org/tcp_table.5.html). Postfix delegates transport queries (`get [domain/email]`) to this service, which resolves routes from the database with in-memory caching for high-concurrency support.

### Lookup Priority

**For email queries (`user@domain.com`):**

| Order | Table | Condition | Action |
|-------|-------|-----------|--------|
| 1st | `mailbox` | `username = email AND active = true` | Per-user transport override |
| 2nd | *(fallback)* | Domain-level lookup below | — |

**For domain queries (`domain.com`):**

| Order | Table | Condition | Action |
|-------|-------|-----------|--------|
| 1st | `transport_list` | `domain = ? AND active = true` | Explicit per-domain routing rule |
| 2nd | `domain` | `domain = ? AND active = true` | Default transport field of the domain |
| — | — | `""` or `"virtual"` | Returns `500 not found` → Postfix uses default routing |

## Configuration

Copy the example and adjust for your environment:

```bash
cp config.toml.example /opt/transport/config.toml
```

```toml
[database]
host     = "127.0.0.1"
port     = "3306"
user     = "postfixadmin"
password = "secret"
name     = "postfixadmin"
debug    = false

[transport]
# Hostname used to detect local delivery rewrites
hostname      = "mail.example.com"
host          = "127.0.0.1:12221"
cache         = "10m"
localdelivery = "smtp:mail.example.com"
delivery      = "lmtp:unix:private/dovecot-lmtp"
```

## Installation

```bash
# Build the binary
make build-prod

# Create the service directory
mkdir -p /opt/transport
cp transport /opt/transport/transport
cp config.toml /opt/transport/config.toml

# Install systemd service
cp postfix-transport.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now postfix-transport
```

## Running Manually

```bash
cd /opt/transport
./transport --config /opt/transport/config.toml
```

With debug output:

```bash
cd /opt/transport
./transport -d --config /opt/transport/config.toml
```

## Service Management

```bash
# Status
systemctl status postfix-transport

# Restart
systemctl restart postfix-transport

# Follow logs
tail -f /opt/transport/transport.log
```

## Postfix Integration

Add to `/etc/postfix/main.cf`:

```ini
transport_maps = tcp:127.0.0.1:12221

# Or combine with static hash maps:
# transport_maps = hash:/etc/postfix/transport, tcp:127.0.0.1:12221
```

Reload Postfix after changes:

```bash
systemctl reload postfix
```

## Testing

Query via Postfix tooling:

```bash
postmap -q "user@example.com" tcp:127.0.0.1:12221
postmap -q "example.com" tcp:127.0.0.1:12221
```

Simulate a raw TCP request:

```bash
echo "get user@example.com" | nc 127.0.0.1 12221
echo "get example.com"       | nc 127.0.0.1 12221
```

Expected responses:

| Code | Meaning |
|------|---------|
| `200 <transport>` | Transport found and returned |
| `500 not found` | No routing rule — Postfix uses its default |
| `400 internal error` | Database or encoding error |
