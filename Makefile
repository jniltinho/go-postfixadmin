BINARY_NAME=postfixadmin

.PHONY: all build run clean css help

all: css build

build: css
	@echo "Building Go application..."
	rm -f $(BINARY_NAME)
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -v -ldflags="-s -w"
	upx $(BINARY_NAME)

run: build
	@echo "Starting application..."
	./$(BINARY_NAME) --run

css:
	@echo "Building CSS with Tailwind..."
	npx @tailwindcss/cli -i ./public/css/input.css -o ./public/css/style.css

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

build-docker:
	@echo "Building Docker image..."
	docker build --no-cache --progress=plain -t $(BINARY_NAME):latest .

help:
	@echo "Makefile commands:"
	@echo "  build         - Build the Go application"
	@echo "  run           - Build and run the application"
	@echo "  css           - Build the CSS using Tailwind"
	@echo "  watch-css     - Watch for CSS changes"
	@echo "  clean         - Remove binary and generated CSS"
	@echo "  tidy          - Run go mod tidy"
	@echo "  deps          - Install Go and NPM dependencies"
	@echo "  build-docker  - Build the Docker image"
