package handlers

import (
	"go-postfixadmin/internal/middleware"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
)

// SetFlash stores a flash message in the user session.
func SetFlash(c *echo.Context, key, value string) {
	sess, _ := session.Get(middleware.UserSessionName, c)
	sess.Values["flash_"+key] = value
	sess.Save(c.Request(), c.Response())
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
