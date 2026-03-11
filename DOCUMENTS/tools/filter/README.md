# filter-nilton

A simple Postfix after-queue content filter written in Go, based on the [Postfix FILTER_README](https://www.postfix.org/FILTER_README.html) simple filter approach.

This is a minimal pass-through filter intended as a foundation for future improvements (spam/virus scanning, header rewriting, etc.).

## How It Works

Postfix delivers the message to the filter via the `pipe(8)` transport. The filter:

1. Saves the message from stdin to a temporary file in `/var/spool/filter/`
2. Logs the key metadata to syslog (`LOG_MAIL`)
3. Reinjects the message via `/usr/sbin/sendmail`
4. Removes the temporary file

On any error the filter exits with `75` (`EX_TEMPFAIL`) so Postfix queues the message for retry.

## Requirements

- Go 1.21+
- Linux
- `/usr/sbin/sendmail` available (Postfix `sendmail` wrapper)
- A dedicated system user to run the filter (e.g. `filter`)

## Build and Install

```bash
cd filter-nilton
make build
sudo make install
```

The binary is installed to `/usr/local/bin/filter`.

## System Setup

Create the working directory and the filter user:

```bash
useradd -r -s /sbin/nologin filter
mkdir -p /var/spool/filter
chown filter:filter /var/spool/filter
chmod 750 /var/spool/filter
```

## Postfix Configuration

### master.cf

Add the following entries:

```
# Simple Go filter — normal load
smtp      inet  n       -       y       -       -       smtpd
  -o content_filter=filter:dummy


filter    unix  -       n       n       -       0       pipe
  flags=Rq user=filter argv=/usr/local/bin/filter --stress 0 --size ${size} --host-ip ${client_address} --host-name ${client_hostname} --helo ${client_helo} --from ${sender} -- ${recipient}

```

Reload Postfix after changes:

```bash
postfix reload
```

## Logging

All messages are logged to syslog (`LOG_MAIL` facility) with the ident `postfix/filter`.

```bash
grep filter /var/log/mail.log
```

Log tags:

| Tag | Description |
|-----|-------------|
| `FILTER:` | Message received, key metadata |
| `FILTER_ACCEPT:` | Message reinjected successfully |
| `FILTER_ERROR:` | Error — Postfix will retry |

## Exit Codes

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | `EX_OK` | Success |
| 75 | `EX_TEMPFAIL` | Temporary failure — Postfix retries |

## Future Improvements

- ClamAV virus scanning
- SpamAssassin integration
- Per-domain configuration (database or config file)
- Header rewriting
- Custom routing based on spam score
