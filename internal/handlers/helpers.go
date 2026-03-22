package handlers

import (
	"strings"

	"go-postfixadmin/internal/i18n"
	"go-postfixadmin/internal/middleware"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
)

// SetFlash stores a flash message in the user session.
func SetFlash(c *echo.Context, key, value string) {
	sess, _ := session.Get(middleware.UserSessionName, c)
	sess.Values["flash_"+key] = i18n.Translate(flashLang(c), value, nil)
	sess.Save(c.Request(), c.Response())
}

func flashLang(c *echo.Context) string {
	if cookie, err := c.Cookie("lang"); err == nil {
		switch cookie.Value {
		case "en", "pt", "es":
			return cookie.Value
		}
	}

	accept := strings.ToLower(c.Request().Header.Get("Accept-Language"))
	switch {
	case strings.HasPrefix(accept, "en"):
		return "en"
	case strings.HasPrefix(accept, "es"):
		return "es"
	default:
		return "pt"
	}
}

// GetFlash retrieves and clears a flash message from the user session.
func GetFlash(c *echo.Context, key string) string {
	sess, _ := session.Get(middleware.UserSessionName, c)
	if sess == nil {
		return ""
	}

	flashKey := "flash_" + key
	val, ok := sess.Values[flashKey].(string)
	if !ok {
		return ""
	}

	delete(sess.Values, flashKey)
	sess.Save(c.Request(), c.Response())
	return val
}
