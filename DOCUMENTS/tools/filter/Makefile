BINARY=filter
BINARY_C=filter2
BINARY_PL=filter.pl
INSTALL_DIR=/usr/local/bin
FLAGS=-v -ldflags="-s -w"
CFLAGS=-O2 -s

.PHONY: build build-c install install-c install-pl clean

build: clean
	CGO_ENABLED=0 go build $(FLAGS) -o $(BINARY) .
	upx --best --lzma $(BINARY)

build-c:
	gcc $(CFLAGS) -o $(BINARY_C) filter.c

install: build
	install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)

install-c: build-c
	install -m 755 $(BINARY_C) $(INSTALL_DIR)/$(BINARY_C)

install-pl:
	install -m 755 $(BINARY_PL) $(INSTALL_DIR)/$(BINARY_PL)

clean:
	rm -f $(BINARY) $(BINARY_C)
