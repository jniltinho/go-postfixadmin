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
- [Quick mail server setup](DOCUMENTS/setup/SETUP_MAILSERVER.md)
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

---
## 📸 Screenshots

![Go-PostfixAdmin Login Screen](DOCUMENTS/screenshots/postfixadmin_01.png)

Check out more images in the [screenshots](DOCUMENTS/screenshots) folder.
