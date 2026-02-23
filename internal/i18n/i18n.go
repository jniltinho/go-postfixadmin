package i18n

import (
	"embed"
	"log/slog"
	"strings"

	"github.com/leonelquinteros/gotext"
)

// locales maps language codes to their parsed PO objects.
var locales map[string]*gotext.Po

// Init initializes the i18n system by loading all PO locale files from the embedded FS.
func Init(fs embed.FS) {
	locales = make(map[string]*gotext.Po)

	// Scan for language directories inside locales/
	dirs, err := fs.ReadDir("locales")
	if err != nil {
		slog.Error("Failed to read locales directory", "error", err)
		return
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		lang := dir.Name()
		poPath := "locales/" + lang + "/default.po"

		data, err := fs.ReadFile(poPath)
		if err != nil {
			slog.Debug("No PO file found for language", "lang", lang, "path", poPath)
			continue
		}

		po := gotext.NewPo()
		po.Parse(data)
		locales[lang] = po
		slog.Debug("Loaded locale", "lang", lang, "path", poPath)
	}

	slog.Info("Loaded locales", "count", len(locales))
}

// Translate returns the localized string for a given message ID and language.
func Translate(lang, messageID string, templateData map[string]any) string {
	if locales == nil {
		return messageID
	}

	// Normalize language code: "pt" -> "pt_BR"
	normalizedLang := normalizeLang(lang)

	po, ok := locales[normalizedLang]
	if !ok {
		// Fallback: try the short code directly
		po, ok = locales[lang]
		if !ok {
			return messageID
		}
	}

	translated := po.Get(messageID) //nolint:govet
	if translated == "" || translated == messageID {
		// If no translation found, return the message ID
		return messageID
	}

	// Handle template data (e.g., {{.Error}} -> actual value)
	for key, val := range templateData {
		placeholder := "{{." + key + "}}"
		if v, ok := val.(string); ok {
			translated = strings.ReplaceAll(translated, placeholder, v)
		}
	}

	return translated
}

// normalizeLang converts short language codes to their full locale codes.
func normalizeLang(lang string) string {
	switch lang {
	case "pt":
		return "pt_BR"
	default:
		return lang
	}
}
