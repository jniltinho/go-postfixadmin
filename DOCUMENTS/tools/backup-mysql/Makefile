BINARY    := backup-mysql
VERSION   := 0.1.2
DATE      := $(shell date '+%Y-%m-%d %H:%M')
LDFLAGS   := -ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.buildDate=$(DATE)'"

.PHONY: all build clean install fmt vet

all: build

build:
	go build $(LDFLAGS) -o $(BINARY) .
	upx --best --lzma $(BINARY)

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)

fmt:
	gofmt -w .

vet:
	go vet ./...
