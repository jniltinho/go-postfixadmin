package server

import (
	"fmt"
	"html/template"
	"strings"

	"go-postfixadmin/internal/i18n"
)

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
		"calcPercentage": func(count, total any) string {
			var c, t float64
			switch v := count.(type) {
			case int:
				c = float64(v)
			case int64:
				c = float64(v)
			case float64:
				c = v
			}
			switch v := total.(type) {
			case int:
				t = float64(v)
			case int64:
				t = float64(v)
			case float64:
				t = v
			}
			if t == 0 {
				return "0"
			}
			return fmt.Sprintf("%.0f", (c/t)*100.0)
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
