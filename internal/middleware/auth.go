package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

const (
	SessionName       = "session"
	UserSessionName   = "user_session"
	AuthKey           = "authenticated"
	UsernameKey       = "username"
	IsSuperAdminKey   = "is_superadmin"
	LastActivityKey   = "last_activity"
	InactivityTimeout = 30 * time.Minute
)

// SetSession authenticates and sets initial session values
func SetSession(c *echo.Context, sessionName string, username string, isSuperAdmin bool) error {
	sess, _ := session.Get(sessionName, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   viper.GetBool("server.ssl_enable"),
		SameSite: http.SameSiteLaxMode,
	}
	sess.Values[AuthKey] = true
	sess.Values[UsernameKey] = username
	if sessionName == SessionName {
		sess.Values[IsSuperAdminKey] = isSuperAdmin
	}
	sess.Values[LastActivityKey] = time.Now().Unix()
	return sess.Save(c.Request(), c.Response())
}

// ClearSession removes a session by its name
func ClearSession(c *echo.Context, sessionName string) error {
	sess, _ := session.Get(sessionName, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	for key := range sess.Values {
		delete(sess.Values, key)
	}
	return sess.Save(c.Request(), c.Response())
}

// GetUsername retrieves the username from JWT (preferred) or session fallback.
func GetUsername(c *echo.Context, sessionName string) string {
	if username, _, found := getIdentityFromJWT(c); found && username != "" {
		return username
	}
	sess, _ := session.Get(sessionName, c)
	if sess == nil {
		return ""
	}
	if username, ok := sess.Values[UsernameKey].(string); ok {
		return username
	}
	return ""
}

// GetIsSuperAdmin retrieves the superadmin flag from JWT (preferred) or session fallback.
func GetIsSuperAdmin(c *echo.Context) bool {
	if _, isSuper, found := getIdentityFromJWT(c); found {
		return isSuper
	}
	sess, _ := session.Get(SessionName, c)
	if sess == nil {
		return false
	}
	if isSuper, ok := sess.Values[IsSuperAdminKey].(bool); ok {
		return isSuper
	}
	return false
}
