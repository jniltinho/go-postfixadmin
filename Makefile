APP        := postfixadmin
BIN        := bin/$(APP)

PREFIX     := go-postfixadmin/cmd
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")


LDFLAGS    := -ldflags "-s -w -X $(PREFIX).Version=$(VERSION) -X $(PREFIX).BuildDate=$(BUILD_TIME) -X $(PREFIX).GitCommit=$(GIT_COMMIT)"

.PHONY: all build run clean css help

all: css build-prod

build: css
	@echo "Building Go application..."
	rm -f $(BIN)
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)


build-prod: css
	@echo "Building Go application..."
	rm -f $(BIN)
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


help:
	@echo "Makefile commands:"
	@echo "  build         - Build the Go application"
	@echo "  run           - Build and run the application"
	@echo "  css           - Build the CSS using Tailwind"
	@echo "  watch-css     - Watch for CSS changes"
	@echo "  clean         - Remove binary and generated CSS"
	@echo "  tidy          - Run go mod tidy"
	@echo "  deps          - Install Go and NPM dependencies"
	@echo "  certs         - Generate self-signed SSL certificates"
	@echo "  build-docker  - Build the Docker image"
