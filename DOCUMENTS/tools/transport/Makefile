UPX_VERSION      := 5.1.1
UPX_ARCHIVE      := upx-$(UPX_VERSION)-amd64_linux.tar.xz
UPX_DIR          := upx-$(UPX_VERSION)-amd64_linux
UPX_BIN          := /usr/local/bin/upx
UPX_URL          := https://github.com/upx/upx/releases/download/v$(UPX_VERSION)/$(UPX_ARCHIVE)

APP        := transport
BIN        := bin/$(APP)
PREFIX     := transport
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DEB_VERSION := $(shell echo $(VERSION) | sed 's/^v//')
RPM_VERSION := $(shell echo $(VERSION) | sed 's/^v//' | tr '-' '.')


LDFLAGS    := -ldflags "-s -w -X transport/version.VERSION=$(VERSION) -X transport/version.BUILD_DATE=$(BUILD_TIME) -X transport/version.GIT_COMMIT=$(GIT_COMMIT)"

.PHONY: all build build-prod run clean deb rpm help install-upx


all: clean build-prod

build: clean
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS) .

build-prod:
	@echo "Building Go application for production..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS) .
	upx --best --lzma $(BIN)

run:
	@echo "Starting application..."
	./$(BIN)

clean:
	@echo "Cleaning up..."
	rm -f $(BIN)

tidy:
	@echo "Tidying go modules..."
	go mod tidy

deps: tidy
	@echo "Downloading dependencies..."
	go mod download

install-upx:
	@echo "Installing UPX binary..."
	curl -ksSL "$(UPX_URL)" -o "$(UPX_ARCHIVE)"
	tar -xf "$(UPX_ARCHIVE)"
	chmod +x "$(UPX_DIR)/upx"
	sudo mv "$(UPX_DIR)/upx" "$(UPX_BIN)"
	rm -rf "$(UPX_DIR)" "$(UPX_ARCHIVE)"

deb: build-prod
	@echo "Building Debian package..."
	rm -rf build/deb
	mkdir -p build/deb/opt/transport
	mkdir -p build/deb/etc/systemd/system
	mkdir -p build/deb/usr/share/doc/postfix-transport
	mkdir -p build/deb/DEBIAN
	cp $(BIN) build/deb/opt/transport/transport
	chmod 755 build/deb/opt/transport/transport
	cp config.toml.example build/deb/opt/transport/config.toml
	chmod 644 build/deb/opt/transport/config.toml
	cp postfix-transport.service build/deb/etc/systemd/system/
	chmod 644 build/deb/etc/systemd/system/postfix-transport.service
	cp README.md build/deb/usr/share/doc/postfix-transport/
	chmod 644 build/deb/usr/share/doc/postfix-transport/README.md
	@echo "Package: postfix-transport"          > build/deb/DEBIAN/control
	@echo "Version: $(DEB_VERSION)"            >> build/deb/DEBIAN/control
	@echo "Section: mail"                      >> build/deb/DEBIAN/control
	@echo "Priority: optional"                 >> build/deb/DEBIAN/control
	@echo "Architecture: amd64"               >> build/deb/DEBIAN/control
	@echo "Maintainer: jniltinho <jniltinho@gmail.com>" >> build/deb/DEBIAN/control
	@echo "Description: Postfix TCP Transport Lookup Server" >> build/deb/DEBIAN/control
	@echo " High-performance Go replacement for transport.pl with in-memory caching." >> build/deb/DEBIAN/control
	dpkg-deb --build build/deb postfix-transport_$(DEB_VERSION)_amd64.deb
	rm -rf build/deb
	@echo "Debian package created: postfix-transport_$(DEB_VERSION)_amd64.deb"

rpm: build-prod
	@echo "Building RPM package..."
	rm -rf build/rpm
	mkdir -p build/rpm/BUILD build/rpm/RPMS build/rpm/SOURCES build/rpm/SPECS build/rpm/SRPMS
	@echo "Name: postfix-transport"                                           > build/rpm/SPECS/postfix-transport.spec
	@echo "Version: $(RPM_VERSION)"                                          >> build/rpm/SPECS/postfix-transport.spec
	@echo "Release: 1"                                                       >> build/rpm/SPECS/postfix-transport.spec
	@echo "Summary: Postfix TCP Transport Lookup Server"                     >> build/rpm/SPECS/postfix-transport.spec
	@echo "License: MIT"                                                     >> build/rpm/SPECS/postfix-transport.spec
	@echo "URL: https://github.com/jniltinho/go-postfixadmin"               >> build/rpm/SPECS/postfix-transport.spec
	@echo "%description"                                                     >> build/rpm/SPECS/postfix-transport.spec
	@echo "High-performance Go replacement for transport.pl with in-memory caching." >> build/rpm/SPECS/postfix-transport.spec
	@echo "%install"                                                         >> build/rpm/SPECS/postfix-transport.spec
	@echo "mkdir -p %{buildroot}/opt/transport"                             >> build/rpm/SPECS/postfix-transport.spec
	@echo "mkdir -p %{buildroot}/etc/systemd/system"                        >> build/rpm/SPECS/postfix-transport.spec
	@echo "mkdir -p %{buildroot}/usr/share/doc/postfix-transport"           >> build/rpm/SPECS/postfix-transport.spec
	@echo "cp $(CURDIR)/$(BIN) %{buildroot}/opt/transport/transport"        >> build/rpm/SPECS/postfix-transport.spec
	@echo "cp $(CURDIR)/config.toml.example %{buildroot}/opt/transport/config.toml" >> build/rpm/SPECS/postfix-transport.spec
	@echo "cp $(CURDIR)/postfix-transport.service %{buildroot}/etc/systemd/system/" >> build/rpm/SPECS/postfix-transport.spec
	@echo "cp $(CURDIR)/README.md %{buildroot}/usr/share/doc/postfix-transport/" >> build/rpm/SPECS/postfix-transport.spec
	@echo "%files"                                                           >> build/rpm/SPECS/postfix-transport.spec
	@echo "%defattr(-,root,root,-)"                                         >> build/rpm/SPECS/postfix-transport.spec
	@echo "/opt/transport/transport"                                         >> build/rpm/SPECS/postfix-transport.spec
	@echo "%config(noreplace) /opt/transport/config.toml"                   >> build/rpm/SPECS/postfix-transport.spec
	@echo "/etc/systemd/system/postfix-transport.service"                   >> build/rpm/SPECS/postfix-transport.spec
	@echo "%doc /usr/share/doc/postfix-transport/README.md"                 >> build/rpm/SPECS/postfix-transport.spec
	rpmbuild -bb --define "_topdir $(CURDIR)/build/rpm" build/rpm/SPECS/postfix-transport.spec
	find build/rpm/RPMS -name "*.rpm" -exec mv {} . \;
	rm -rf build/rpm
	@echo "RPM package created"

help:
	@echo "Makefile commands:"
	@echo "  build            - Performs standard Go build"
	@echo "  build-prod       - Builds and compresses the binary with UPX"
	@echo "  deb              - Builds .deb package (requires dpkg-deb)"
	@echo "  rpm              - Builds .rpm package (requires rpmbuild)"
	@echo "  run              - Runs the application ./transport"
	@echo "  clean            - Removes the generated binary"
	@echo "  tidy             - Runs 'go mod tidy'"
	@echo "  deps             - Downloads dependencies and runs tidy"
	@echo "  install-upx      - Installs the UPX binary on the system"

