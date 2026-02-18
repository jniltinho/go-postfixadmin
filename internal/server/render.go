package server

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
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

	// Standalone login pages (no layout)
	if name == "login.html" || name == "users/login.html" {
		return tmpl.ExecuteTemplate(w, name, data)
	}

	// User portal pages use user_base layout
	if len(name) > 6 && name[:6] == "users/" {
		return tmpl.ExecuteTemplate(w, "user_base", data)
	}

	// Admin pages use base layout
	return tmpl.ExecuteTemplate(w, "base", data)
}

// templateFuncMap returns the custom template functions used across all templates.
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
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
		"commaToLines": func(s string) template.HTML {
			parts := strings.Split(s, ",")
			var trimmed []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					trimmed = append(trimmed, template.HTMLEscapeString(p))
				}
			}
			return template.HTML(strings.Join(trimmed, "<br>"))
		},
		"commaToNewlines": func(s string) string {
			parts := strings.Split(s, ",")
			var trimmed []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					trimmed = append(trimmed, p)
				}
			}
			return strings.Join(trimmed, "\n")
		},
	}
}

// loadTemplates parses all view templates from the embedded filesystem.
func loadTemplates(embeddedFiles embed.FS) (*Template, error) {
	t := &Template{
		templates: make(map[string]*template.Template),
	}

	funcMap := templateFuncMap()

	layout := "views/layout.html"
	userLayout := "views/user_layout.html"

	err := fs.WalkDir(embeddedFiles, "views", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if name == "layout.html" || name == "user_layout.html" {
			return nil
		}

		var tmplKey string
		var tmpl *template.Template
		var parseErr error

		if path.Dir(filePath) == "views/users" {
			tmplKey = "users/" + name
			if name == "login.html" {
				tmpl, parseErr = template.New(tmplKey).Funcs(funcMap).ParseFS(embeddedFiles, filePath)
			} else {
				tmpl, parseErr = template.New(tmplKey).Funcs(funcMap).ParseFS(embeddedFiles, userLayout, filePath)
			}
		} else {
			tmplKey = name
			if name == "login.html" {
				tmpl, parseErr = template.New(tmplKey).Funcs(funcMap).ParseFS(embeddedFiles, filePath)
			} else {
				tmpl, parseErr = template.New(tmplKey).Funcs(funcMap).ParseFS(embeddedFiles, layout, filePath)
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
