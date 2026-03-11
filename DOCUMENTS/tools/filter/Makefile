BINARY=filter
INSTALL_DIR=/usr/local/bin
FLAGS=-v -ldflags="-s -w"

.PHONY: build install clean

build: clean
	CGO_ENABLED=0 go build $(FLAGS) -o $(BINARY) .
	upx --best --lzma $(BINARY)

install: build
	install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)
