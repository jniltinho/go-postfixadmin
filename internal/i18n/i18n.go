package i18n

import (
	"embed"
	"log/slog"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

// Init initializes the i18n bundle by loading all TOML locale files from the embedded FS.
func Init(fs embed.FS) {
	bundle = i18n.NewBundle(language.BrazilianPortuguese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Dynamically load all .toml files in the locales directory
	files, err := fs.ReadDir("locales")
	if err != nil {
		slog.Error("Failed to read locales directory", "error", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		filePath := "locales/" + file.Name()
		_, err = bundle.LoadMessageFileFS(fs, filePath)
		if err != nil {
			slog.Warn("Failed to load locale file", "file", filePath, "error", err)
		} else {
			slog.Debug("Loaded locale file", "file", filePath)
		}
	}
}

// Translate returns the localized string for a given message ID and language.
func Translate(lang, messageID string, templateData map[string]any) string {
	if bundle == nil {
		return messageID
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})

	if err != nil {
		// slog.Debug("Translation not found", "msgID", messageID, "lang", lang)
		return messageID
	}

	return msg
}
