## Variables for Tailwind CSS and UPX
TAILWIND_VERSION := v4.2.0
TAILWIND_BIN     := /usr/local/bin/tailwindcss
TAILWIND_URL     := https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64
UPX_VERSION      := 5.1.1
UPX_ARCHIVE      := upx-$(UPX_VERSION)-amd64_linux.tar.xz
UPX_DIR          := upx-$(UPX_VERSION)-amd64_linux
UPX_BIN          := /usr/local/bin/upx
UPX_URL          := https://github.com/upx/upx/releases/download/v$(UPX_VERSION)/$(UPX_ARCHIVE)

## Variables for Go application
APP        := postfixadmin
BIN        := bin/$(APP)
PREFIX     := go-postfixadmin/cmd
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")


LDFLAGS    := -ldflags "-s -w -X $(PREFIX).Version=$(VERSION) -X $(PREFIX).BuildDate=$(BUILD_TIME) -X $(PREFIX).GitCommit=$(GIT_COMMIT)"
DEB_VERSION := $(shell echo $(VERSION) | sed 's/^v//')
RPM_VERSION := $(shell echo $(DEB_VERSION) | tr '-' '_')

.PHONY: all build build-prod run clean css help install-tailwind install-upx deb rpm

all: clean css build-prod

build: clean css
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)


build-prod:
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)
	upx --best --lzma $(BIN)


run:
	@echo "Starting application..."
	./$(BIN) server

css:
	@echo "Building CSS with Tailwind..."
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css --minify

watch-css:
	@echo "Watching CSS changes..."
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css --watch

clean:
	@echo "Cleaning up..."
	rm -f $(BIN)
	rm -f web/static/css/style.css

tidy:
	@echo "Tidying go modules..."
	go mod tidy

deps:
	@echo "Installing dependencies..."
	go mod download

certs:
	@echo "Generating SSL certificates..."
	mkdir -p ssl
	openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
		-keyout ssl/server.key -out ssl/server.crt \
		-subj "/C=BR/ST=SP/L=Sao Paulo/O=Development/CN=localhost"

build-docker:
	@echo "Building Docker image..."
	docker build --no-cache --progress=plain -t jniltinho/go-postfixadmin:latest .

build-docker-prod:
	@echo "Building Go application..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)
	upx --best --lzma $(BIN)

deb: build-prod
	@echo "Building Debian package..."
	rm -rf build/deb
	mkdir -p build/deb/opt/go-postfixadmin
	mkdir -p build/deb/etc/systemd/system
	mkdir -p build/deb/usr/share/doc/go-postfixadmin
	mkdir -p build/deb/DEBIAN
	cp $(BIN) build/deb/opt/go-postfixadmin/postfixadmin
	chmod 755 build/deb/opt/go-postfixadmin/postfixadmin
	cp config.toml.example build/deb/opt/go-postfixadmin/config.toml
	chmod 644 build/deb/opt/go-postfixadmin/config.toml
	cp DOCUMENTS/setup/postfixadmin.service build/deb/etc/systemd/system/
	chmod 644 build/deb/etc/systemd/system/postfixadmin.service
	cp DOCUMENTS/setup/README.md build/deb/usr/share/doc/go-postfixadmin/
	chmod 644 build/deb/usr/share/doc/go-postfixadmin/README.md
	@echo "Package: go-postfixadmin" > build/deb/DEBIAN/control
	@echo "Version: $(DEB_VERSION)" >> build/deb/DEBIAN/control
	@echo "Section: mail" >> build/deb/DEBIAN/control
	@echo "Priority: optional" >> build/deb/DEBIAN/control
	@echo "Architecture: amd64" >> build/deb/DEBIAN/control
	@echo "Maintainer: jniltinho <jniltinho@gmail.com>" >> build/deb/DEBIAN/control
	@echo "Description: Go PostfixAdmin Web Interface" >> build/deb/DEBIAN/control
	@echo " A fully featured web interface for configuring Postfix and Dovecot." >> build/deb/DEBIAN/control
	dpkg-deb --build build/deb go-postfixadmin_$(DEB_VERSION)_amd64.deb
	rm -rf build/deb
	@echo "Debian package created: go-postfixadmin_$(DEB_VERSION)_amd64.deb"

rpm: build-prod
	@echo "Building RPM package..."
	rm -rf build/rpm
	mkdir -p build/rpm/BUILD build/rpm/RPMS build/rpm/SOURCES build/rpm/SPECS build/rpm/SRPMS
	@echo "Name: go-postfixadmin" > build/rpm/SPECS/go-postfixadmin.spec
	@echo "Version: $(RPM_VERSION)" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "Release: 1" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "Summary: Go PostfixAdmin Web Interface" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "License: MIT" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "URL: https://github.com/jniltinho/go-postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%description" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "A fully featured web interface for configuring Postfix and Dovecot." >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%install" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "mkdir -p %{buildroot}/opt/go-postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "mkdir -p %{buildroot}/etc/systemd/system" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "mkdir -p %{buildroot}/usr/share/doc/go-postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/$(BIN) %{buildroot}/opt/go-postfixadmin/postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/config.toml.example %{buildroot}/opt/go-postfixadmin/config.toml" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/DOCUMENTS/setup/postfixadmin.service %{buildroot}/etc/systemd/system/" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/DOCUMENTS/setup/README.md %{buildroot}/usr/share/doc/go-postfixadmin/" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%files" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%defattr(-,root,root,-)" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "/opt/go-postfixadmin/postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%config(noreplace) /opt/go-postfixadmin/config.toml" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "/etc/systemd/system/postfixadmin.service" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%doc /usr/share/doc/go-postfixadmin/README.md" >> build/rpm/SPECS/go-postfixadmin.spec
	rpmbuild -bb --define "_topdir $(CURDIR)/build/rpm" build/rpm/SPECS/go-postfixadmin.spec
	find build/rpm/RPMS -name "*.rpm" -exec mv {} . \;
	rm -rf build/rpm
	@echo "RPM package created"

## For Development and Build/Production
install-tailwind:
	@echo "Installing Tailwind CSS binary..."
	curl -ksSL "$(TAILWIND_URL)" -o tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 "$(TAILWIND_BIN)"

install-upx:
	@echo "Installing UPX binary..."
	curl -ksSL "$(UPX_URL)" -o "$(UPX_ARCHIVE)"
	tar -xf "$(UPX_ARCHIVE)"
	chmod +x "$(UPX_DIR)/upx"
	mv "$(UPX_DIR)/upx" "$(UPX_BIN)"
	rm -rf "$(UPX_DIR)" "$(UPX_ARCHIVE)"



help:
	@echo "Makefile commands:"
	@echo "  build            - Build the Go application"
	@echo "  run              - Build and run the application"
	@echo "  css              - Build the CSS using Tailwind"
	@echo "  watch-css        - Watch for CSS changes"
	@echo "  clean            - Remove binary and generated CSS"
	@echo "  tidy             - Run go mod tidy"
	@echo "  deps             - Install Go dependencies, Tailwind CSS, and UPX"
	@echo "  install-tailwind - Install the Tailwind CSS binary"
	@echo "  install-upx      - Install the UPX binary"
	@echo "  certs            - Generate self-signed SSL certificates"
	@echo "  build-docker     - Build the Docker image"
	@echo "  deb              - Build the Debian (.deb) package"
	@echo "  rpm              - Build the RPM (.rpm) package"
