# Development

## What This Document Covers

This file covers local development prerequisites, build options, and the most useful Makefile commands.

Related documents:

- [Project README](README.md)
- [Features](FEATURES.md)
- [Quick mail server setup](SETUP_MAILSERVER.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)

## Development Tools

To compile the project locally without Docker, install the following tools:

1. **Go (v1.21 or higher)**: Main language of the project.
   [Download Go](https://go.dev/dl/)
2. **Node.js (v20 or higher)**: Required for CSS processing with Tailwind.
   [Download Node.js](https://nodejs.org/)
3. **Make**: Utility for command automation (native on Linux/macOS).
4. **UPX (Optional)**: Used by the Makefile to compress the final binary.
   Debian/Ubuntu: `sudo apt install upx-ucl`

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
npm install
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
