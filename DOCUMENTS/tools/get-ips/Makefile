BINARY_NAME=get-ips
BUILD_DIR=./bin

.PHONY: all build run clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "-s -w" .

run:
	go run main.go

clean:
	@rm -rf $(BUILD_DIR)
