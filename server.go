package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"

	"go-postfixadmin/internal/handlers"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
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

func StartServer(embeddedFiles embed.FS, port int, db *gorm.DB) {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Template Pre-parsing (from embed.FS)
	t := &Template{
		templates: make(map[string]*template.Template),
	}

	layout := "views/layout.html"
	viewFiles, _ := embeddedFiles.ReadDir("views")

	for _, file := range viewFiles {
		name := file.Name()
		if name == "layout.html" {
			continue
		}

		pagePath := path.Join("views", name)
		if name == "login.html" {
			t.templates[name] = template.Must(template.ParseFS(embeddedFiles, pagePath))
		} else {
			t.templates[name] = template.Must(template.ParseFS(embeddedFiles, layout, pagePath))
		}
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

	e.GET("/login", h.Login)
	e.POST("/login", h.Login)
	e.GET("/dashboard", h.Dashboard)
	e.GET("/domains", h.ListDomains)

	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	addr := fmt.Sprintf(":%d", port)
	slog.Info("Starting server", "address", addr)
	if err := e.Start(addr); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
