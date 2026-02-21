BINARY_NAME=vacation

## https://claude.ai/chat/17ab10fb-0643-43f0-8aea-8691da8fbbae

.PHONY: all build

all: build


build:
	@echo "Building Go application..."
	rm -f $(BINARY_NAME)
	go mod tidy
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -v -ldflags="-s -w"

