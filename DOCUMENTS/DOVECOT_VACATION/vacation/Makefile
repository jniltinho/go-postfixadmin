BINARY_NAME=dovecot-vacation

## https://claude.ai/chat/17ab10fb-0643-43f0-8aea-8691da8fbbae

.PHONY: all build deps run clean

all: build

build:
	@echo "Building Go application..."
	rm -f $(BINARY_NAME)
	go mod tidy
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -v -ldflags="-s -w"
	upx $(BINARY_NAME)

deps:
	@echo "Installing dependencies..."
	go get github.com/joho/godotenv
	go mod tidy

run: build
	@echo "Running vacation (test mode)..."
	echo "" | ./$(BINARY_NAME) -f test@example.com -t recipient@example.com

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)


tidy:
	@echo "Tidying go modules..."
	go mod tidy