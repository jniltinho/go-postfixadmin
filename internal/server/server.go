package server

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func StartServer(embeddedFiles embed.FS, port int, db *gorm.DB, ssl bool, certFile, keyFile string) {
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())

	// Session Middleware
	secret := viper.GetString("server.session_secret")
	if secret == "" {
		secret = os.Getenv("SESSION_SECRET") // fallback
	}

	if secret == "" {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			secret = "9a048f79e88e35de37dc2c43c1fa002f358f92957a7690e60109cfe8a65178e0"
		} else {
			secret = hex.EncodeToString(bytes)
			slog.Info("Generated random session secret", "secret", secret)
		}
	}
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(secret))))

	// Auth Middleware
	e.Use(middleware.AuthMiddleware)

	// Template Rendering
	t, err := loadTemplates(embeddedFiles)
	if err != nil {
		slog.Error("Failed to load templates", "error", err)
		os.Exit(1)
	}
	e.Renderer = t

	// Handlers
	h := &handlers.Handler{DB: db}

	// Static files from embedded FS
	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		e.Logger.Error("failed to create sub filesystem", "error", err)
		os.Exit(1)
	}

	staticHandler := http.FileServer(http.FS(publicFS))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", staticHandler)))

	// Register Application Routes
	routes.RegisterRoutes(e, h)

	addr := fmt.Sprintf(":%d", port)
	slog.Info("Starting server", "address", addr)

	if ssl {
		if certFile == "" || keyFile == "" {
			slog.Error("SSL enabled but cert or key file not provided")
			os.Exit(1)
		}
		slog.Info("SSL enabled", "cert", certFile, "key", keyFile)

		server := &http.Server{Addr: addr, Handler: e}
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server with SSL", "error", err)
			os.Exit(1)
		}
	} else {
		if err := e.Start(addr); err != nil {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}
}
