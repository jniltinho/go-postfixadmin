BINARY_NAME=vacation

## https://chatgpt.com/c/6999f953-3190-832b-83ed-dac6ae2b8fb7


.PHONY: all build

all: build


build:
	@echo "Building Go application..."
	rm -f $(BINARY_NAME)
	go mod tidy
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -v -ldflags="-s -w"

