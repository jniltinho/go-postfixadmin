package server

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// Template stores pre-parsed templates for each route.
type Template struct {
	templates map[string]*template.Template
}

// Render executes the pre-parsed template set.
func (t *Template) Render(c *echo.Context, w io.Writer, name string, data any) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Template not found: "+name)
	}

	// Determine client language
	lang := "pt"

	// 1. Check for cookie
	if cookie, err := c.Cookie("lang"); err == nil && (cookie.Value == "en" || cookie.Value == "pt" || cookie.Value == "es") {
		lang = cookie.Value
	} else {
		// 2. Fallback to Accept-Language header
		accept := c.Request().Header.Get("Accept-Language")
		if strings.HasPrefix(strings.ToLower(accept), "en") {
			lang = "en"
		}
	}

	fetchmailEnabled := viper.GetBool("features.fetchmail")

	currentPath := c.Request().URL.Path

	var viewData any = data
	if data == nil {
		viewData = map[string]any{"Lang": lang, "FetchmailEnabled": fetchmailEnabled, "CurrentPath": currentPath}
	} else if m, ok := data.(map[string]any); ok {
		m["Lang"] = lang
		m["FetchmailEnabled"] = fetchmailEnabled
		m["CurrentPath"] = currentPath
		viewData = m
	} else if m, ok := data.(map[string]interface{}); ok {
		m["Lang"] = lang
		m["FetchmailEnabled"] = fetchmailEnabled
		m["CurrentPath"] = currentPath
		viewData = m
	}

	// Determine layout
	layout := "base"
	if name == "auth/login.html" {
		layout = "login.html"
	} else if name == "users/login.html" {
		layout = "users/login.html"
	} else if strings.HasPrefix(name, "users/") {
		layout = "user_base"
	}

	return tmpl.ExecuteTemplate(w, layout, viewData)
}

// loadTemplates parses all view templates from the embedded filesystem.
// Files named form_*.html are treated as partials: they are not registered as
// standalone templates but are automatically included in every page template set.
func loadTemplates(embeddedFiles embed.FS) (*Template, error) {
	t := &Template{
		templates: make(map[string]*template.Template),
	}

	funcMap := templateFuncMap()

	layout := "web/templates/layout/layout.html"
	userLayout := "web/templates/users/layout.html"

	// Collect partial templates (form_*.html) from all directories.
	var partials []string
	fs.WalkDir(embeddedFiles, "web/templates", func(filePath string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasPrefix(d.Name(), "form_") {
			partials = append(partials, filePath)
		}
		return nil
	})

	err := fs.WalkDir(embeddedFiles, "web/templates", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if filePath == "web/templates/layout/layout.html" || filePath == "web/templates/users/layout.html" {
			return nil
		}
		// Partials are included in other templates, not registered standalone.
		if strings.HasPrefix(name, "form_") {
			return nil
		}

		// Use the path relative to web/templates/ as the key
		tmplKey := strings.TrimPrefix(filePath, "web/templates/")

		var tmpl *template.Template
		var parseErr error

		if strings.HasPrefix(tmplKey, "users/") {
			if name == "login.html" {
				tmpl, parseErr = template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, filePath)
			} else {
				tmpl, parseErr = template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, userLayout, filePath)
			}
		} else {
			if name == "login.html" {
				tmpl, parseErr = template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, filePath)
			} else {
				parseFiles := append([]string{layout}, append(partials, filePath)...)
				tmpl, parseErr = template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, parseFiles...)
			}
		}

		if parseErr != nil {
			return fmt.Errorf("failed to parse template %s: %w", filePath, parseErr)
		}

		t.templates[tmplKey] = tmpl
		return nil
	})

	if err != nil {
		return nil, err
	}

	return t, nil
}
