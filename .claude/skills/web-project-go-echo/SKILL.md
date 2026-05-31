---
name: web-project-go-echo
description: Scaffold a production-ready Go web project using Echo v5, GORM (SQLite dev / MariaDB prod), Cobra CLI, embedded Vue 3 SPA (Tailwind CSS v4 + tailwindcss-primeui + PrimeVue, Pinia, Lucide, vue3-toastify), thin-border + soft-shadow design system, 3-stage Dockerfile, and optimized Makefile. Use when the user asks to create a new Go + Vue web project or API server.
---

# web-project-go-echo

Scaffold a complete production-ready Go + Vue 3 project. Every file below must be
created exactly as shown. Replace `{PROJECT_NAME}` with the actual module/project name
(e.g. `myapp`) and `{MODULE_PATH}` with the Go module path (e.g. `github.com/user/myapp`).

Ask the user for:
1. **Project name** (snake_case, used for binary + folder)
2. **Go module path** (e.g. `github.com/user/projectname`)
3. **Target directory** (default: `./{project-name}`)

---

## Project Structure

```
{project-name}/
├── cmd/
│   ├── root.go
│   └── server.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   └── database.go
│   ├── handlers/
│   │   └── handlers.go
│   ├── middleware/
│   │   └── middleware.go
│   └── models/
│       └── models.go
├── internal/
│   ├── ...
│   └── server/
│       ├── server.go
│       ├── routes.go
│       └── render.go
├── specs/
│   └── example-feature/
│       ├── FEATURES.md
│       ├── TASK.md
│       ├── PLAN.md
│       ├── DONE.md
│       └── SPEC.md
├── web/
│   ├── dist/            # Vue 3 build output (git-ignored, embedded)
│   │   └── .gitkeep
│   └── files/
│       └── config.default.toml
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   │   └── client.ts
│   │   ├── assets/
│   │   ├── components/
│   │   ├── pages/
│   │   │   ├── LoginPage.vue
│   │   │   └── DashboardPage.vue
│   │   ├── router/
│   │   │   └── index.ts
│   │   ├── stores/
│   │   │   └── auth.ts
│   │   ├── App.vue
│   │   ├── env.d.ts
│   │   ├── main.ts
│   │   └── style.css
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── main.go
├── go.mod
├── Makefile
├── Dockerfile
├── .gitignore
├── config.toml.example
└── README.md
```

---

## Step 1 — Create directory and initialize git

```bash
mkdir -p {project-name} && cd {project-name}
git init
```

---

## Step 2 — `main.go`

```go
// Package main is the entry point for {PROJECT_NAME}.
// It embeds the web/dist SPA assets, web/files config templates,
// and locale files, then delegates to the CLI.
package main

import (
	"embed"

	"{MODULE_PATH}/cmd"
)

//go:embed all:web/dist web/files
var embeddedFiles embed.FS

func main() {
	cmd.Execute(embeddedFiles)
}
```

---

## Step 3 — `cmd/root.go`

```go
// Package cmd provides the CLI for {PROJECT_NAME} using Cobra.
package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"{MODULE_PATH}/internal/config"
	"{MODULE_PATH}/internal/database"
)

var (
	cfgFile        string
	embeddedAssets embed.FS
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "{project-name}",
	Short: "{PROJECT_NAME} — web application server",
	Long:  `{PROJECT_NAME} is a web application built with Go + Echo + Vue 3.`,
}

// Execute passes the embedded FS and runs the CLI.
// It is called from main.go.
func Execute(fs embed.FS) {
	embeddedAssets = fs
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config.toml)")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(migrateCmd)
}

// migrateCmd runs database auto-migrations without starting the server.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		db, err := database.Connect(cfg)
		if err != nil {
			return err
		}
		if err := database.Migrate(db); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		fmt.Println("migrations applied successfully")
		return nil
	},
}

// versionCmd prints the application version.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("{PROJECT_NAME} v0.1.0")
	},
}

// initCmd writes the default config.toml from the embedded template.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file from embedded template",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := embeddedAssets.ReadFile("web/files/config.default.toml")
		if err != nil {
			return fmt.Errorf("reading embedded config template: %w", err)
		}
		target := "config.toml"
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("%s already exists; remove it first", target)
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", target, err)
		}
		fmt.Printf("Created %s from embedded template\n", target)
		return nil
	},
}
```

---

## Step 4 — `cmd/server.go`

```go
package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"{MODULE_PATH}/internal/config"
	"{MODULE_PATH}/internal/database"
	"{MODULE_PATH}/internal/server"
)

// serverCmd starts the HTTP server.
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		db, err := database.Connect(cfg)
		if err != nil {
			return err
		}

		if err := database.Migrate(db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}

		s := server.New(cfg, db, embeddedAssets)
		return s.Start()
	},
}
```

---

## Step 5 — `internal/config/config.go`

```go
// Package config handles loading and validation of application configuration.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application settings loaded from config.toml.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name  string `mapstructure:"name"`
	Debug bool   `mapstructure:"debug"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	TLSEnable bool   `mapstructure:"tls_enable"`
	TLSPort   int    `mapstructure:"tls_port"`
	TLSCert   string `mapstructure:"tls_cert"`
	TLSKey    string `mapstructure:"tls_key"`
	JWTSecret string `mapstructure:"jwt_secret"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`   // sqlite | mysql
	DSN      string `mapstructure:"dsn"`      // file path (sqlite) or DSN (mysql)
}

// Load reads the configuration file and returns a Config.
// If cfgFile is empty it searches for config.toml in the current directory.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("toml")
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("app.name", "{PROJECT_NAME}")
	v.SetDefault("app.debug", false)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.tls_enable", false)
	v.SetDefault("server.tls_port", 8443)
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "{project-name}.db")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}
```

---

## Step 6 — `internal/database/database.go`

```go
// Package database manages database connections and migrations via GORM.
package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"{MODULE_PATH}/internal/config"
	"{MODULE_PATH}/internal/models"
)

// Connect opens a database connection using the driver specified in cfg.
// Supported drivers: "sqlite", "mysql".
func Connect(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{}
	if !cfg.App.Debug {
		gormCfg.Logger = logger.Default.LogMode(logger.Silent)
	}

	var (
		db  *gorm.DB
		err error
	)

	switch cfg.Database.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.DSN), gormCfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.Database.DSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("connecting to database (%s): %w", cfg.Database.Driver, err)
	}

	return db, nil
}

// Migrate runs auto-migration for all registered models.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
	)
}
```

---

## Step 7 — `internal/models/models.go`

```go
// Package models defines GORM database models for {PROJECT_NAME}.
package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an application user.
type User struct {
	ID        uint           `gorm:"primaryKey"         json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:254" json:"email"`
	Password  string         `gorm:"size:255"            json:"-"`
	Active    bool           `gorm:"default:true"        json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"               json:"-"`
}
```

---

## Step 8 — `internal/middleware/middleware.go`

```go
// Package middleware provides Echo middleware for {PROJECT_NAME}.
package middleware

import (
	"net/http"
	"time"

	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	emw "github.com/labstack/echo/v5/middleware"
)

// Register attaches all global middleware to the Echo instance.
func Register(e *echo.Echo, debug bool) {
	e.Use(emw.Recover())
	e.Use(emw.RequestLogger())
	e.Use(emw.RequestID())
	e.Use(emw.Secure())
	e.Use(emw.TimeoutWithConfig(emw.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	if debug {
		e.Use(emw.CORSWithConfig(emw.CORSConfig{
			AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentType, "X-Requested-With"},
			AllowCredentials: true,
		}))
	}
}

// JWT returns Echo JWT middleware configured with the given secret.
func JWT(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(secret),
	})
}

// RateLimiter returns a basic IP-based rate limiter middleware.
func RateLimiter() echo.MiddlewareFunc {
	return emw.RateLimiterWithConfig(emw.RateLimiterConfig{
		Store: emw.NewRateLimiterMemoryStoreWithConfig(emw.RateLimiterMemoryStoreConfig{
			Rate:      10,
			Burst:     30,
			ExpiresIn: time.Minute,
		}),
	})
}
```

---

## Step 9 — `internal/handlers/handlers.go`

```go
// Package handlers provides Echo HTTP request handlers for {PROJECT_NAME}.
package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"{MODULE_PATH}/internal/config"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	cfg *config.Config
	db  *gorm.DB
}

// New creates a new Handler with the given dependencies.
func New(cfg *config.Config, db *gorm.DB) *Handler {
	return &Handler{cfg: cfg, db: db}
}

func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "0.1.0",
	})
}

// AuthLogin validates credentials and issues a signed JWT.
func (h *Handler) AuthLogin(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return apiError(c, http.StatusBadRequest, "bad_request", "invalid request body")
	}
	if req.Username == "" || req.Password == "" {
		return apiError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials")
	}

	// TODO: validate credentials against DB
	// var user models.User
	// if err := h.db.Where("username = ? AND active = true", req.Username).First(&user).Error; err != nil {
	//     return apiError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials")
	// }
	// if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
	//     return apiError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials")
	// }

	claims := jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.cfg.Server.JWTSecret))
	if err != nil {
		return apiError(c, http.StatusInternalServerError, "internal_error", "could not sign token")
	}

	return apiOK(c, map[string]any{
		"access_token": signed,
		"user":         map[string]string{"username": req.Username},
	})
}

// apiError is a standard JSON error response helper.
func apiError(c echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// apiOK is a standard JSON success response helper.
func apiOK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}
```

---

## Step 10 — `internal/server/server.go`

```go
// Package server sets up and runs the Echo HTTP server for {PROJECT_NAME}.
package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	"{MODULE_PATH}/internal/config"
	"{MODULE_PATH}/internal/handlers"
	mw "{MODULE_PATH}/internal/middleware"
)

// Server wraps the Echo instance with application dependencies.
type Server struct {
	echo *echo.Echo
	cfg  *config.Config
}

// New creates and configures a new Server instance.
func New(cfg *config.Config, db *gorm.DB, embeddedFiles embed.FS) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	mw.Register(e, cfg.App.Debug)

	spaFS, err := fs.Sub(embeddedFiles, "web/dist")
	if err != nil {
		slog.Warn("web/dist not found — SPA will not be served. Run 'make build' first.")
		spaFS = nil
	}

	h := handlers.New(cfg, db)
	registerRoutes(e, h, cfg.Server.JWTSecret, spaFS)

	return &Server{echo: e, cfg: cfg}
}

// Start listens on the configured address and shuts down gracefully on SIGINT/SIGTERM.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	fmt.Printf("→ {PROJECT_NAME} listening on http://%s\n", addr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.echo.Start(addr); err != nil {
			s.echo.Logger.Info("shutting down server")
		}
	}()

	if s.cfg.Server.TLSEnable {
		tlsAddr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.TLSPort)
		fmt.Printf("→ {PROJECT_NAME} TLS listening on https://%s\n", tlsAddr)
		go func() {
			if err := s.echo.StartTLS(tlsAddr, s.cfg.Server.TLSCert, s.cfg.Server.TLSKey); err != nil {
				s.echo.Logger.Info("shutting down TLS server")
			}
		}()
	}

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.echo.Shutdown(ctx)
}
```

---

## Step 11 — `internal/server/routes.go`

```go
package server

import (
	"io/fs"
	"mime"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"

	"{MODULE_PATH}/internal/handlers"
	mw "{MODULE_PATH}/internal/middleware"
)

// registerRoutes wires all routes to the Echo instance.
func registerRoutes(e *echo.Echo, h *handlers.Handler, jwtSecret string, spaFS fs.FS) {
	// --- Public routes ---
	e.POST("/api/v1/auth/login", h.AuthLogin)

	// --- Protected API v1 ---
	api := e.Group("/api/v1")
	api.Use(mw.JWT(jwtSecret))
	api.Use(mw.RateLimiter())

	api.GET("/health", h.Health)

	// --- SPA fallback (must be last) ---
	if spaFS != nil {
		e.GET("/*", spaHandler(spaFS))
	}
}

// spaHandler serves the embedded Vue 3 SPA.
// Static files are served with correct MIME type and cache headers.
// All other GET paths fall back to index.html for Vue Router history mode.
func spaHandler(spaFS fs.FS) echo.HandlerFunc {
	return func(c *echo.Context) error {
		urlPath := c.Request().URL.Path
		ext := strings.ToLower(filepath.Ext(urlPath))

		if ext != "" {
			fsPath := strings.TrimPrefix(urlPath, "/")
			data, err := fs.ReadFile(spaFS, fsPath)
			if err != nil {
				return echo.ErrNotFound
			}
			ct := mime.TypeByExtension(ext)
			if ct == "" {
				ct = "application/octet-stream"
			}
			if ext == ".js" || ext == ".mjs" {
				ct = "application/javascript; charset=utf-8"
			}
			if strings.HasPrefix(urlPath, "/assets/") {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Response().Header().Set("Cache-Control", "public, max-age=3600")
			}
			c.Response().Header().Set("Content-Type", ct)
			_, _ = c.Response().Write(data)
			return nil
		}

		indexHTML, err := fs.ReadFile(spaFS, "index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().Header().Set("Cache-Control", "no-cache")
		_, _ = c.Response().Write(indexHTML)
		return nil
	}
}
```

---

## Step 12 — `internal/server/render.go`

```go
// Package server — render.go handles server-side template rendering if needed.
// For a pure SPA project this file is a no-op placeholder.
package server
```

---

## Step 13 — `web/files/config.default.toml`

```toml
# {PROJECT_NAME} — default configuration template
# Copy to config.toml and adjust values before running the server.
# Generated by: {project-name} init

[app]
name  = "{PROJECT_NAME}"
debug = false

[server]
host       = "0.0.0.0"
port       = 8080
tls_enable = false
tls_port   = 8443
tls_cert   = ""
tls_key    = ""
jwt_secret = "change-me-to-a-random-64-char-string"

[database]
# driver = "sqlite" for development, "mysql" for production
driver = "sqlite"
dsn    = "{project-name}.db"

# MySQL/MariaDB DSN example:
# driver = "mysql"
# dsn    = "user:password@tcp(127.0.0.1:3306)/{project-name}?charset=utf8mb4&parseTime=True&loc=Local"
```

---

## Step 14 — `config.toml.example`

Same content as `web/files/config.default.toml`.

---

## Step 15 — `go.mod`

Run after creating the files:

```bash
go mod init {MODULE_PATH}
go get github.com/labstack/echo/v5@latest
go get github.com/labstack/echo-jwt/v5@latest
go get github.com/golang-jwt/jwt/v5@latest
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/sqlite@latest
go get gorm.io/driver/mysql@latest
go mod tidy
```

---

## Step 16 — `Makefile`

```makefile
# ── Variables ──────────────────────────────────────────────────────────────────
APP       := {project-name}
MODULE    := {MODULE_PATH}
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -s -w -X '$(MODULE)/cmd.Version=$(VERSION)'
GOFLAGS   := CGO_ENABLED=1

# Frontend
FRONTEND_DIR := frontend
DIST_DIR     := web/dist

.PHONY: all build build-prod frontend-build dev clean test lint help

# ── Default ────────────────────────────────────────────────────────────────────
all: frontend-build build

# ── Go build ───────────────────────────────────────────────────────────────────

## build: Compile the application (development)
build:
	@echo "→ Building $(APP)..."
	$(GOFLAGS) go build -ldflags "$(LDFLAGS)" -o bin/$(APP) .

## build-prod: Compile and compress the binary for production (requires upx)
build-prod: frontend-build
	@echo "→ Building $(APP) for production..."
	$(GOFLAGS) go build -ldflags "$(LDFLAGS)" -trimpath -o bin/$(APP) .
	@echo "→ Compressing binary with upx..."
	upx --best --lzma bin/$(APP)
	@echo "→ Binary size: $$(du -sh bin/$(APP) | cut -f1)"

## run: Run the server in development mode
run: build
	./bin/$(APP) server --config config.toml

# ── Frontend ───────────────────────────────────────────────────────────────────

## frontend-install: Install npm dependencies
frontend-install:
	@echo "→ Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install

## frontend-build: Build the Vue 3 SPA into web/dist
frontend-build:
	@echo "→ Building frontend..."
	cd $(FRONTEND_DIR) && npm run build
	@echo "→ Frontend built to $(DIST_DIR)"

## frontend-dev: Start the Vite dev server (proxies /api to :8080)
frontend-dev:
	cd $(FRONTEND_DIR) && npm run dev

# ── Development ────────────────────────────────────────────────────────────────

## dev: Run backend + frontend in parallel (requires tmux or two terminals)
dev:
	@echo "→ Start backend:  make run"
	@echo "→ Start frontend: make frontend-dev"

## watch: Auto-rebuild Go on file change (requires entr)
watch:
	find . -name '*.go' | entr -r make run

# ── Quality ────────────────────────────────────────────────────────────────────

## test: Run all tests
test:
	go test ./... -v -race -cover

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## vet: Run go vet
vet:
	go vet ./...

# ── Utilities ──────────────────────────────────────────────────────────────────

## init-config: Generate config.toml from the embedded template
init-config: build
	./bin/$(APP) init

## clean: Remove build artifacts
clean:
	rm -rf bin/ $(DIST_DIR)/*
	@touch $(DIST_DIR)/.gitkeep

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
```

---

## Step 17 — `Dockerfile`

```dockerfile
# ─── Stage 1: Build Vue 3 Frontend ────────────────────────────────────────────
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci --prefer-offline

COPY frontend/ .
RUN npm run build

# ─── Stage 2: Build the Go Application ────────────────────────────────────────
FROM golang:1.26-alpine AS go-builder

RUN apk add --no-cache gcc musl-dev sqlite-dev upx

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /app/frontend/../web/dist ./web/dist

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/bin/{project-name} .

RUN upx --best --lzma /app/bin/{project-name}

# ─── Stage 3: Final Minimal Image ─────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=go-builder /app/bin/{project-name} ./bin/{project-name}

USER app

EXPOSE 8080 8443

ENTRYPOINT ["./bin/{project-name}"]
CMD ["server"]
```

---

## Step 18 — `.gitignore`

```gitignore
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Go test artifacts
*.test
*.out
coverage.txt

# Build cache
vendor/

# Database files (development)
*.db
*.db-shm
*.db-wal

# Configuration (contains secrets)
config.toml

# Frontend build output (regenerated by make frontend-build)
web/dist/*
!web/dist/.gitkeep

# Node
frontend/node_modules/
frontend/dist/

# OS
.DS_Store
Thumbs.db

# Editor
.vscode/
.idea/
*.swp
*.swo
```

---

## Step 19 — `README.md`

```markdown
# {PROJECT_NAME}

> A web application built with Go + Echo v5 + Vue 3.

## Requirements

- Go 1.26+
- Node.js 22+ / npm 11+
- SQLite (development) or MariaDB/MySQL (production)
- `upx` for production builds

## Quick Start

```bash
# 1. Install frontend deps and build
make frontend-install
make frontend-build

# 2. Generate config
make build && ./bin/{project-name} init

# 3. Run the server
make run
# → http://localhost:8080
```

## Development

Run the backend and frontend dev server in parallel:

```bash
# Terminal 1 — Go backend (auto-rebuild with entr)
make watch

# Terminal 2 — Vite dev server (:5173 → proxies /api to :8080)
make frontend-dev
```

## Production Build

```bash
make build-prod
# → bin/{project-name}  (compressed with upx --best --lzma)
```

## Docker

```bash
docker build -t {project-name} .
docker run -p 8080:8080 {project-name}
```

## Configuration

| Key                  | Default            | Description               |
|----------------------|--------------------|---------------------------|
| `app.debug`          | `false`            | Enable debug mode + CORS  |
| `server.port`        | `8080`             | HTTP port                 |
| `server.tls_enable`  | `false`            | Enable HTTPS              |
| `server.tls_port`    | `8443`             | HTTPS port                |
| `database.driver`    | `sqlite`           | `sqlite` or `mysql`       |
| `database.dsn`       | `{project-name}.db`| DB path or connection DSN |

## Project Structure

See the `specs/` folder for feature specifications and implementation plans.

## License

MIT
```

---

## Step 20 — Frontend files

### `frontend/package.json`

```json
{
  "name": "{project-name}-frontend",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.5.0",
    "vue-router": "^4.4.0",
    "pinia": "^2.2.0",
    "axios": "^1.7.0",
    "vue3-toastify": "^0.2.0",
    "@lucide/vue": "^1.0.0",
    "primevue": "^4.0.0",
    "@primevue/themes": "^4.0.0",
    "primeicons": "^7.0.0",
    "tailwindcss-primeui": "^0.5.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "@tailwindcss/vite": "^4.0.0",
    "tailwindcss": "^4.0.0",
    "typescript": "^5.0.0",
    "vue-tsc": "^2.0.0",
    "vite": "^6.0.0"
  }
}
```

---

### `frontend/vite.config.ts`

```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  define: {
    API_BASE: JSON.stringify('/api/v1'),
  },
  resolve: {
    alias: { '@': resolve(__dirname, 'src') }
  },
  build: {
    outDir: '../web/dist',
    emptyOutDir: true
  },
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/lang': { target: 'http://localhost:8080', changeOrigin: true }
    }
  }
})
```

---

### `frontend/tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.d.ts", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

---

### `frontend/index.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>{PROJECT_NAME}</title>
  <link rel="icon" href="/favicon.ico" />
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.ts"></script>
</body>
</html>
```

---

### `frontend/src/style.css`

```css
@import "tailwindcss";
@import "tailwindcss-primeui";

/* ── Brand tokens ──────────────────────────────────────────────────────────── */
@theme {
  --color-brand-primary:    #3B82F6;
  --color-brand-secondary:  #6366EE;
  --color-brand-cta:        #F97316;
  --color-brand-bg:         #F8FAFC;
  --color-brand-text:       #1E293B;

  --font-sans: 'Inter', ui-sans-serif, system-ui, sans-serif;
  --font-mono: 'Fira Code', ui-monospace, monospace;
}

:root {
  --color-brand-primary:   #3B82F6;
  --color-brand-secondary: #6366EE;
  --color-brand-cta:       #F97316;
  --color-brand-bg:        #F8FAFC;
  --color-brand-text:      #1E293B;
}

/* ── Global reset ──────────────────────────────────────────────────────────── */
*, *::before, *::after { box-sizing: border-box; }
body {
  margin: 0;
  font-family: var(--font-sans);
  background: var(--color-brand-bg);
  color: var(--color-brand-text);
  -webkit-font-smoothing: antialiased;
}

/* ── Neo-brutalist primitives (thin border · soft shadow · no radius) ──────── */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  cursor: pointer;
  border: 1px solid currentColor;
  box-shadow: 0 1px 4px rgba(0,0,0,0.10);
  transition: all 0.15s ease;
  border-radius: 0;
}
.btn:hover  { box-shadow: 0 3px 10px rgba(0,0,0,0.14); transform: translateY(-1px); }
.btn:active { transform: translateY(0); box-shadow: 0 1px 3px rgba(0,0,0,0.10); }

.btn-primary { background: var(--color-brand-primary); color: #fff; border-color: var(--color-brand-primary); }
.btn-danger  { background: #ef4444; color: #fff; border-color: #ef4444; }
.btn-ghost   { background: #fff; color: var(--color-brand-text); border-color: #cbd5e1; }

.card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 0;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 0;
  font-size: 14px;
  outline: none;
  background: #fff;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.input:focus { border-color: var(--color-brand-primary); box-shadow: 0 0 0 3px rgba(59,130,246,0.15); }

.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(15,23,42,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 9999; padding: 16px;
}
.modal-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 0;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
  width: 100%; max-width: 520px; max-height: 90vh; overflow: auto;
}
.modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 18px;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
  border-radius: 0;
  font-weight: 900; font-size: 12px;
  letter-spacing: 0.5px; text-transform: uppercase;
}
.modal-body   { padding: 18px; }
.modal-footer {
  padding: 12px 18px;
  border-top: 1px solid #e2e8f0;
  display: flex; gap: 10px; justify-content: flex-end;
  background: #fafafa;
  border-radius: 0;
}

.page-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 20px; flex-wrap: wrap; gap: 12px;
}
.page-title {
  font-size: 16px; font-weight: 900;
  letter-spacing: 1px; text-transform: uppercase;
  color: var(--color-brand-text);
}

.error-banner {
  background: #fef2f2; border: 1px solid #fca5a5; color: #b91c1c;
  border-radius: 0;
  padding: 10px 14px; font-size: 12px; font-weight: 700;
  letter-spacing: 0.3px;
  margin-bottom: 16px; display: flex; align-items: center; gap: 6px;
}

.spinner {
  width: 22px; height: 22px;
  border: 2px solid #e2e8f0; border-top-color: var(--color-brand-primary);
  border-radius: 50%; animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
```

---

### `frontend/src/env.d.ts`

```ts
/// <reference types="vite/client" />

declare const API_BASE: string
```

---

### `frontend/src/main.ts`

```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Vue3Toastify, { toast } from 'vue3-toastify'
import 'vue3-toastify/dist/index.css'

import App from './App.vue'
import router from './router/index'
import './style.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Vue3Toastify, {
  autoClose:   3000,
  position:    'top-right',
  transition:  'slide',
  theme:       'light',
  toastStyle:  {
    border:        '1px solid #e2e8f0',
    boxShadow:     '0 4px 12px rgba(0,0,0,0.10)',
    borderRadius:  '0',
    fontWeight:    '700',
    fontSize:      '12px',
    letterSpacing: '0.3px',
  },
})

app.config.globalProperties.$toast = toast

app.mount('#app')
```

---

### `frontend/src/App.vue`

```vue
<template>
  <router-view />
</template>

<script setup lang="ts">
</script>
```

---

### `frontend/src/router/index.ts`

```ts
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/LoginPage.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/pages/DashboardPage.vue'),
    meta: { requiresAuth: true }
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && auth.isLoggedIn) {
    return { name: 'dashboard' }
  }
})

export default router
```

---

### `frontend/src/stores/auth.ts`

```ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/client'

interface AuthUser {
  username: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('access_token') || '')
  const user  = ref<AuthUser | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  async function login(username: string, password: string) {
    const res = await api.post('/auth/login', { username, password })
    token.value = res.data.data.access_token
    user.value  = res.data.data.user
    localStorage.setItem('access_token', token.value)
    api.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  function logout() {
    token.value = ''
    user.value  = null
    localStorage.removeItem('access_token')
    delete api.defaults.headers.common['Authorization']
  }

  if (token.value) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  return { token, user, isLoggedIn, login, logout }
})
```

---

### `frontend/src/api/client.ts`

```ts
import axios from 'axios'

declare const API_BASE: string

const api = axios.create({
  baseURL: API_BASE,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('access_token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default api
```

---

### `frontend/src/pages/LoginPage.vue`

```vue
<template>
  <div class="login-page">
    <div class="login-wrap">
      <div class="login-brand">
        <div class="brand-icon">
          <Lock :size="28" color="#fff" />
        </div>
        <h1 class="brand-title">{PROJECT_NAME}</h1>
        <p class="brand-sub">Sign in to continue</p>
      </div>

      <div class="card" style="padding: 28px;">
        <p class="card-heading">SIGN IN</p>
        <div v-if="error" class="error-banner">{{ error }}</div>

        <form @submit.prevent="handleLogin">
          <div class="field-wrap">
            <label class="field-label">Username</label>
            <input v-model="form.username" class="input" type="text" required autocomplete="username" />
          </div>
          <div class="field-wrap" style="margin-bottom:20px;">
            <label class="field-label">Password</label>
            <input v-model="form.password" class="input" type="password" required autocomplete="current-password" />
          </div>
          <button class="btn btn-primary" style="width:100%;justify-content:center;" :disabled="loading">
            <span v-if="loading" class="spinner" style="width:16px;height:16px;border-width:2px;"></span>
            <span v-else>Sign In</span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Lock } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue3-toastify'

const router = useRouter()
const auth   = useAuthStore()

const form    = ref({ username: '', password: '' })
const error   = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value   = ''
  loading.value = true
  try {
    await auth.login(form.value.username, form.value.password)
    toast.success('Welcome back!')
    router.push('/dashboard')
  } catch (e) {
    error.value = e.response?.data?.error?.message || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: var(--color-brand-bg);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 8vh;
}
.login-wrap   { width: 100%; max-width: 420px; padding: 0 16px 40px; }
.login-brand  { display: flex; flex-direction: column; align-items: center; margin-bottom: 24px; }
.brand-icon   { width: 52px; height: 52px; background: var(--color-brand-primary); border-radius: 0; display: flex; align-items: center; justify-content: center; margin-bottom: 12px; box-shadow: 0 2px 8px rgba(59,130,246,0.20); }
.brand-title  { font-size: 24px; font-weight: 900; color: var(--color-brand-text); margin: 0 0 4px; letter-spacing: -0.5px; }
.brand-sub    { font-size: 12px; color: #64748b; margin: 0; font-weight: 600; letter-spacing: 0.3px; }
.card-heading { font-size: 11px; font-weight: 900; letter-spacing: 1px; text-transform: uppercase; color: var(--color-brand-text); margin: 0 0 18px; }
.field-wrap   { margin-bottom: 14px; }
.field-label  { display: block; font-size: 11px; font-weight: 700; letter-spacing: 0.8px; text-transform: uppercase; color: #475569; margin-bottom: 5px; }
</style>
```

---

### `frontend/src/pages/DashboardPage.vue`

```vue
<template>
  <div style="padding: 24px;">
    <div class="page-header">
      <div>
        <h1 class="page-title">Dashboard</h1>
        <p style="font-size:12px; color:#64748b; margin:2px 0 0; font-weight:700; letter-spacing:0.3px;">
          Overview
        </p>
      </div>
    </div>

    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px,1fr)); gap:16px;">
      <div class="card" style="padding:20px;">
        <p style="font-size:11px;font-weight:700;letter-spacing:0.8px;text-transform:uppercase;color:#64748b;margin:0 0 8px;">
          Status
        </p>
        <p style="font-size:28px;font-weight:900;margin:0;color:var(--color-brand-primary);">OK</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>
```

---

## Step 21 — `specs/example-feature/` files

Create each file with the standard template:

**SPEC.md**
```markdown
# Spec — example-feature

**Objective**: Describe the technical specification here.

## Overview
...

## Design Decisions
...
```

**FEATURES.md**
```markdown
# Features — example-feature

## Must Have
- [ ] ...

## Nice to Have
- [ ] ...

## Non-Goals
- ...
```

**PLAN.md**
```markdown
# Plan — example-feature

## Phases

### Phase 1
- [ ] ...
```

**TASK.md**
```markdown
# Tasks — example-feature

## In Progress
- [ ] ...

## Pending
- [ ] ...
```

**DONE.md**
```markdown
# Done — example-feature

- [x] ...
```

---

## Step 22 — Final setup commands

```bash
# Go dependencies
go mod init {MODULE_PATH}
go get github.com/labstack/echo/v5@latest
go get github.com/labstack/echo-jwt/v5@latest
go get github.com/golang-jwt/jwt/v5@latest
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/sqlite@latest
go get gorm.io/driver/mysql@latest
go mod tidy

# Frontend — scaffold then add UI packages
npm create vue@latest -- --bare --ts --router --pinia frontend
cd frontend
npm i tailwindcss @tailwindcss/vite @lucide/vue
npm i tailwindcss-primeui primevue @primevue/themes primeicons
npm i axios vue3-toastify
npm run build
cd ..

# Initialize config
make build
./bin/{project-name} init

# Verify
make build
./bin/{project-name} server
```

---

## Step 23 — Git initial commit

```bash
git add .
git commit -m "feat: initial project scaffold

- Echo v5 + GORM (SQLite dev / MariaDB prod) + Cobra CLI
- Vue 3 + Tailwind CSS v4 + tailwindcss-primeui + PrimeVue + Pinia + Lucide + vue3-toastify
- Thin-border + soft-shadow design system
- 3-stage Dockerfile + optimized Makefile (upx)
- SPA embedded in Go binary via //go:embed
- specs/ folder structure

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Notes

- **Go version**: 1.26 (adjust `go.mod` `go` directive accordingly)
- **Echo v5**: uses `github.com/labstack/echo/v5` — middleware import path is `github.com/labstack/echo/v5/middleware`
- **Frontend scaffold**: use `npm create vue@latest -- --bare --ts --router --pinia frontend` then install tailwindcss + primeui packages; overwrite generated files with the templates in Step 20
- **vue3-toastify**: soft config — `position="top-right"`, `transition="slide"`, rounded borders (`8px`), subtle `box-shadow`
- **tailwindcss-primeui**: bridges PrimeVue theming with Tailwind utility classes; import order must be `tailwindcss` first, then `tailwindcss-primeui`
- **SPA fallback**: registered last in `internal/server/routes.go`; serves static files with correct MIME + Cache-Control; falls back to `index.html` for Vue Router history mode paths
- **upx**: must be installed on the build machine (`apt install upx` / `apk add upx`)
- **SQLite**: requires `CGO_ENABLED=1` and `gcc` at build time
- **Production DB**: set `database.driver = "mysql"` and a full DSN in `config.toml`
