package i18n

import (
	"embed"
	"log/slog"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

// Init initializes the i18n bundle by loading TOML locale files from the embedded FS.
func Init(fs embed.FS) {
	bundle = i18n.NewBundle(language.BrazilianPortuguese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load Portuguese
	_, err := bundle.LoadMessageFileFS(fs, "locales/pt_br.toml")
	if err != nil {
		slog.Warn("Failed to load pt_br locale", "error", err)
	}

	// Load English
	_, err = bundle.LoadMessageFileFS(fs, "locales/en.toml")
	if err != nil {
		slog.Warn("Failed to load en locale", "error", err)
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
