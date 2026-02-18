package server

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"

	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"
)

// Template stores pre-parsed templates for each route
type Template struct {
	templates map[string]*template.Template
}

// Render executes the pre-parsed template set.
func (t *Template) Render(c *echo.Context, w io.Writer, name string, data any) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Template not found: "+name)
	}

	if name != "login.html" {
		return tmpl.ExecuteTemplate(w, "base", data)
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func StartServer(embeddedFiles embed.FS, port int, db *gorm.DB, ssl bool, certFile, keyFile string) {
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())

	// Session Middleware
	// Using a hardcoded secret for simplicity. In production, use os.Getenv("SESSION_SECRET")
	secret := os.Getenv("SESSION_SECRET")
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

	// Template Pre-parsing (from embed.FS)
	t := &Template{
		templates: make(map[string]*template.Template),
	}

	layout := "views/layout.html"
	viewFiles, _ := embeddedFiles.ReadDir("views")

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b float64) float64 { return a / b },
		"float64": func(i any) float64 {
			switch v := i.(type) {
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case float64:
				return v
			default:
				return 0
			}
		},
	}

	for _, file := range viewFiles {
		name := file.Name()
		if name == "layout.html" {
			continue
		}

		pagePath := path.Join("views", name)

		var tmpl *template.Template
		var err error

		if name == "login.html" {
			tmpl, err = template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, pagePath)
		} else {
			// For pages using layout, we need to parse layout first or together.
			// template.ParseFS uses the filenames as template names.
			// "base" is likely defined in layout.html.
			tmpl, err = template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, layout, pagePath)
		}

		if err != nil {
			slog.Error("Failed to parse template", "name", name, "error", err)
			os.Exit(1)
		}
		t.templates[name] = tmpl
	}

	e.Renderer = t

	// Handlers
	h := &handlers.Handler{DB: db}

	// Routes
	// Serve static files from embedded FS (public subdirectory)
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
