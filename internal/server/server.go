package server

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var AppVersion string = "1.0.0"

func StartServer(embeddedFiles embed.FS, port int, db *gorm.DB, ssl bool, certFile, keyFile string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	e := echo.New()
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.RequestLogger())

	secret := viper.GetString("server.session_secret")
	if secret == "" {
		secret = os.Getenv("SESSION_SECRET")
	}
	if secret == "" {
		slog.Warn("session_secret not configured — sessions will be invalidated on server restart!")
		b := make([]byte, 32)
		if _, err := rand.Read(b); err == nil {
			secret = hex.EncodeToString(b)
		} else {
			secret = "9a048f79e88e35de37dc2c43c1fa002f358f92957a7690e60109cfe8a65178e0"
		}
	}
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(secret))))

	spaFS, err := fs.Sub(embeddedFiles, "web/dist")
	if err != nil {
		slog.Warn("web/dist not found — SPA will not be served. Run 'make build' first.")
		spaFS = nil
	}

	h := &handlers.Handler{DB: db}
	routes.RegisterRoutes(e, h, spaFS)

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           e,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down server…")
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = srv.Shutdown(shutCtx)
	}()

	slog.Info("Starting server", "address", addr)
	if ssl {
		if certFile == "" || keyFile == "" {
			return fmt.Errorf("SSL enabled but cert or key file not provided")
		}
		slog.Info("SSL enabled", "cert", certFile, "key", keyFile)
		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
