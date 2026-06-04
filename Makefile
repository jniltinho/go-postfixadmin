## Variables for UPX
UPX_VERSION := 5.1.1
UPX_ARCHIVE  := upx-$(UPX_VERSION)-amd64_linux.tar.xz
UPX_DIR      := upx-$(UPX_VERSION)-amd64_linux
UPX_BIN      := /usr/local/bin/upx
UPX_URL      := https://github.com/upx/upx/releases/download/v$(UPX_VERSION)/$(UPX_ARCHIVE)

## Variables for Go application
APP        := postfixadmin
BIN        := bin/$(APP)
PREFIX     := go-postfixadmin/cmd
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS    := -trimpath -ldflags "-s -w -X $(PREFIX).Version=$(VERSION) -X $(PREFIX).BuildDate=$(BUILD_TIME) -X $(PREFIX).GitCommit=$(GIT_COMMIT)"
DEB_VERSION := $(shell echo $(VERSION) | sed 's/^v//')
RPM_VERSION := $(shell echo $(DEB_VERSION) | tr '-' '_')

.PHONY: all build build-prod run clean frontend swagger help install-upx deb rpm

all: clean frontend build

build: clean frontend
	@echo "Building Go application (with embedded frontend)..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)


build-prod: frontend
	@echo "Building Go application (with embedded frontend)..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)
	upx --best --lzma $(BIN)


run:
	@echo "Starting application..."
	./$(BIN) server

frontend:
	@echo "Building frontend (Vue 3 + Tailwind CSS)..."
	@if [ ! -d "frontend" ]; then \
		echo "Error: frontend/ directory not found"; \
		exit 1; \
	fi
	cd frontend && npm install && npm run build
	@echo "Frontend built successfully into web/dist"

swagger:
	@echo "Generating Swagger documentation..."
	go run github.com/swaggo/swag/cmd/swag@latest init -g main.go --parseDependency --parseInternal
	@echo "Swagger docs generated in docs/"
	@echo "Don't forget to run 'make build' afterwards so the docs are embedded."

clean:
	@echo "Cleaning up..."
	rm -f $(BIN)
	rm -rf web/dist

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

build-docker-prod: frontend
	@echo "Building Go application (with embedded frontend)..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS)
	upx --best --lzma $(BIN)

deb: build-prod
	@echo "Building Debian package..."
	rm -rf build/deb
	mkdir -p build/deb/opt/go-postfixadmin/logs
	mkdir -p build/deb/etc/systemd/system
	mkdir -p build/deb/usr/share/doc/go-postfixadmin
	mkdir -p build/deb/DEBIAN
	cp $(BIN) build/deb/opt/go-postfixadmin/postfixadmin
	chmod 755 build/deb/opt/go-postfixadmin/postfixadmin
	cp config.toml.example build/deb/opt/go-postfixadmin/config.toml
	chmod 644 build/deb/opt/go-postfixadmin/config.toml
	cp DOCUMENTS/setup/postfixadmin.service build/deb/etc/systemd/system/
	chmod 644 build/deb/etc/systemd/system/postfixadmin.service
	cp DOCUMENTS/setup/postfixadmin-transport.service build/deb/etc/systemd/system/
	chmod 644 build/deb/etc/systemd/system/postfixadmin-transport.service
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
	@echo "mkdir -p %{buildroot}/opt/go-postfixadmin/logs" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "mkdir -p %{buildroot}/etc/systemd/system" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "mkdir -p %{buildroot}/usr/share/doc/go-postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/$(BIN) %{buildroot}/opt/go-postfixadmin/postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/config.toml.example %{buildroot}/opt/go-postfixadmin/config.toml" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/DOCUMENTS/setup/postfixadmin.service %{buildroot}/etc/systemd/system/" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/DOCUMENTS/setup/postfixadmin-transport.service %{buildroot}/etc/systemd/system/" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "cp $(CURDIR)/DOCUMENTS/setup/README.md %{buildroot}/usr/share/doc/go-postfixadmin/" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%files" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%defattr(-,root,root,-)" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "/opt/go-postfixadmin/postfixadmin" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%config(noreplace) /opt/go-postfixadmin/config.toml" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%dir /opt/go-postfixadmin/logs" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "/etc/systemd/system/postfixadmin.service" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "/etc/systemd/system/postfixadmin-transport.service" >> build/rpm/SPECS/go-postfixadmin.spec
	@echo "%doc /usr/share/doc/go-postfixadmin/README.md" >> build/rpm/SPECS/go-postfixadmin.spec
	rpmbuild -bb --define "_topdir $(CURDIR)/build/rpm" build/rpm/SPECS/go-postfixadmin.spec
	find build/rpm/RPMS -name "*.rpm" -exec mv {} . \;
	rm -rf build/rpm
	@echo "RPM package created"

install-upx:
	@echo "Installing UPX binary..."
	curl -ksSL "$(UPX_URL)" -o "$(UPX_ARCHIVE)"
	tar -xf "$(UPX_ARCHIVE)"
	chmod +x "$(UPX_DIR)/upx"
	mv "$(UPX_DIR)/upx" "$(UPX_BIN)"
	rm -rf "$(UPX_DIR)" "$(UPX_ARCHIVE)"

help:
	@echo "Makefile commands:"
	@echo "  build            - Full build: Frontend (Vue 3) + Go binary (embedded)"
	@echo "  build-prod       - Full build + UPX compression (for releases)"
	@echo "  frontend         - Build only Vue 3 SPA -> web/dist/"
	@echo "  swagger          - Regenerate Swagger/OpenAPI from annotations"
	@echo "  clean            - Remove bin/postfixadmin and web/dist/"
	@echo "  run              - Build (if needed) and start server"
	@echo "  deps / tidy      - Go module download / cleanup"
	@echo "  deb / rpm        - Build Debian / RPM packages (include services + config)"
	@echo "  build-docker     - Multi-stage Docker image (tagged jniltinho/go-postfixadmin:latest)"
	@echo "  install-upx      - Install UPX compressor"
	@echo ""
	@echo "Swagger UI: http://localhost:8080/swagger/ (when server.swagger_enable=true)"
	@echo "Tip: 'make swagger && make build' to embed updated docs. See DEVELOPMENT.md for full list."
