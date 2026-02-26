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

	"go-postfixadmin/internal/i18n"

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

	var viewData any = data
	if data == nil {
		viewData = map[string]any{"Lang": lang, "FetchmailEnabled": fetchmailEnabled}
	} else if m, ok := data.(map[string]any); ok {
		m["Lang"] = lang
		m["FetchmailEnabled"] = fetchmailEnabled
		viewData = m
	} else if m, ok := data.(map[string]interface{}); ok {
		m["Lang"] = lang
		m["FetchmailEnabled"] = fetchmailEnabled
		viewData = m
	}

	// Determine layout
	layout := "base"
	if name == "login.html" || name == "users/login.html" {
		layout = name
	} else if len(name) > 6 && name[:6] == "users/" {
		layout = "user_base"
	}

	return tmpl.ExecuteTemplate(w, layout, viewData)
}

// templateFuncMap returns the custom template functions used across all templates.
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"T": func(lang, messageID string) string {
			return i18n.Translate(lang, messageID, nil)
		},
		"TData": func(lang, messageID string, templateData map[string]interface{}) string {
			return i18n.Translate(lang, messageID, templateData)
		},
		"version": func() string { return AppVersion },
		"mul":     func(a, b float64) float64 { return a * b },
		"div":     func(a, b float64) float64 { return a / b },
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
		"unescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
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
	userLayout := "views/users/layout.html"

	err := fs.WalkDir(embeddedFiles, "views", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if filePath == "views/layout.html" || filePath == "views/users/layout.html" {
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
