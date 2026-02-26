BINARY_NAME=postfixadmin

DATE=`date +%Y-%m-%d\ %H:%M`
VERSION=v1.0.15
PREFIX=go-postfixadmin/cmd
LDFLAGS = -X '${PREFIX}.Version=${VERSION}' -X '${PREFIX}.BuildDate=${DATE}'
FLAGS=-v -ldflags="-s -w ${LDFLAGS}"

.PHONY: all build run clean css help

all: css build-prod

build: css
	@echo "Building Go application..."
	rm -f $(BINARY_NAME)
	CGO_ENABLED=0 go build -o $(BINARY_NAME) ${FLAGS}


build-prod: css
	@echo "Building Go application..."
	rm -f $(BINARY_NAME)
	CGO_ENABLED=0 go build -o $(BINARY_NAME) ${FLAGS}
	upx $(BINARY_NAME)


run: build
	@echo "Starting application..."
	./$(BINARY_NAME) server

css:
	@echo "Building CSS with Tailwind..."
	npx @tailwindcss/cli -i ./public/css/input.css -o ./public/css/style.css --minify

watch-css:
	@echo "Watching CSS changes..."
	npx tailwindcss -i ./public/css/input.css -o ./public/css/style.css --watch

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f public/css/style.css

tidy:
	@echo "Tidying go modules..."
	go mod tidy

deps:
	@echo "Installing dependencies..."
	go mod download
	npm install

certs:
	@echo "Generating SSL certificates..."
	mkdir -p certs
	openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
		-keyout certs/server.key -out certs/server.crt \
		-subj "/C=BR/ST=SP/L=Sao Paulo/O=Development/CN=localhost"

build-docker:
	@echo "Building Docker image..."
	docker build --no-cache --progress=plain -t jniltinho/go-postfixadmin:latest .

build-docker-prod:
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BINARY_NAME) ${FLAGS}


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
