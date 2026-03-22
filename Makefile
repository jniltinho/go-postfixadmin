## Variables for Tailwind CSS and UPX
TAILWIND_VERSION := v4.2.0
TAILWIND_BIN     := /usr/local/bin/tailwindcss
TAILWIND_URL     := https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64
UPX_VERSION      := 5.1.1
UPX_ARCHIVE      := upx-$(UPX_VERSION)-amd64_linux.tar.xz
UPX_DIR          := upx-$(UPX_VERSION)-amd64_linux
UPX_BIN          := /usr/local/bin/upx
UPX_URL          := https://github.com/upx/upx/releases/download/v$(UPX_VERSION)/$(UPX_ARCHIVE)

## Variables for Go application
APP        := postfixadmin
BIN        := bin/$(APP)
PREFIX     := go-postfixadmin/cmd
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")


LDFLAGS    := -ldflags "-s -w -X $(PREFIX).Version=$(VERSION) -X $(PREFIX).BuildDate=$(BUILD_TIME) -X $(PREFIX).GitCommit=$(GIT_COMMIT)"

.PHONY: all build run clean css help install-tailwind install-upx

all: clean css build-prod

build: clean css
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)


build-prod:
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)
	upx --best --lzma $(BIN)


run:
	@echo "Starting application..."
	./$(BIN) server

css:
	@echo "Building CSS with Tailwind..."
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css --minify

watch-css:
	@echo "Watching CSS changes..."
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css --watch

clean:
	@echo "Cleaning up..."
	rm -f $(BIN)
	rm -f web/static/css/style.css

tidy:
	@echo "Tidying go modules..."
	go mod tidy

deps:
	@echo "Installing dependencies..."
	go mod download

certs:
	@echo "Generating SSL certificates..."
	mkdir -p ssl
	openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
		-keyout ssl/server.key -out ssl/server.crt \
		-subj "/C=BR/ST=SP/L=Sao Paulo/O=Development/CN=localhost"

build-docker:
	@echo "Building Docker image..."
	docker build --no-cache --progress=plain -t jniltinho/go-postfixadmin:latest .

build-docker-prod:
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)
	upx --best --lzma $(BIN)

## For Development and Build/Production
install-tailwind:
	@echo "Installing Tailwind CSS binary..."
	curl -ksSL "$(TAILWIND_URL)" -o tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 "$(TAILWIND_BIN)"

install-upx:
	@echo "Installing UPX binary..."
	curl -ksSL "$(UPX_URL)" -o "$(UPX_ARCHIVE)"
	tar -xf "$(UPX_ARCHIVE)"
	chmod +x "$(UPX_DIR)/upx"
	mv "$(UPX_DIR)/upx" "$(UPX_BIN)"
	rm -rf "$(UPX_DIR)" "$(UPX_ARCHIVE)"



help:
	@echo "Makefile commands:"
	@echo "  build            - Build the Go application"
	@echo "  run              - Build and run the application"
	@echo "  css              - Build the CSS using Tailwind"
	@echo "  watch-css        - Watch for CSS changes"
	@echo "  clean            - Remove binary and generated CSS"
	@echo "  tidy             - Run go mod tidy"
	@echo "  deps             - Install Go dependencies, Tailwind CSS, and UPX"
	@echo "  install-tailwind - Install the Tailwind CSS binary"
	@echo "  install-upx      - Install the UPX binary"
	@echo "  certs            - Generate self-signed SSL certificates"
	@echo "  build-docker     - Build the Docker image"
