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
	UserSessionName   = "user_session" // New session for users
	AuthKey           = "authenticated"
	UsernameKey       = "username"
	IsSuperAdminKey   = "is_superadmin"
	LastActivityKey   = "last_activity"
	InactivityTimeout = 30 * time.Minute
)

// baseAuthMiddleware provides a generic authentication middleware generator
func baseAuthMiddleware(sessionName, loginPath string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {

			// Skip check for login page and static assets
			if c.Path() == "/login" || c.Path() == "/static/*" || c.Path() == "/users/login" || c.Path() == "/users/logout" || c.Path() == "/lang/:code" {
				return next(c)
			}

			// Prevent browser caching of protected pages
			c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Response().Header().Set("Pragma", "no-cache")
			c.Response().Header().Set("Expires", "0")

			sess, _ := session.Get(sessionName, c)
			if sess == nil {
				return c.Redirect(http.StatusFound, loginPath)
			}

			// Check if authenticated
			auth, ok := sess.Values[AuthKey].(bool)
			if !ok || !auth {
				return c.Redirect(http.StatusFound, loginPath)
			}

			// Check inactivity
			lastActivity, ok := sess.Values[LastActivityKey].(int64)
			if ok {
				lastActivityTime := time.Unix(lastActivity, 0)
				if time.Since(lastActivityTime) > InactivityTimeout {
					// Session expired
					sess.Options.MaxAge = -1
					sess.Save(c.Request(), c.Response())
					return c.Redirect(http.StatusFound, loginPath+"?expired=true")
				}
			}

			// Update last activity
			sess.Values[LastActivityKey] = time.Now().Unix()
			sess.Save(c.Request(), c.Response())

			return next(c)
		}
	}
}

// AuthMiddleware checks for admin session validity and inactivity
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return baseAuthMiddleware(SessionName, "/login")(next)
}

// UserAuthMiddleware checks for user session validity and inactivity
func UserAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return baseAuthMiddleware(UserSessionName, "/users/login")(next)
}

// SetSession authenticates and sets initial session values
func SetSession(c *echo.Context, sessionName string, username string, isSuperAdmin bool) error {
	sess, _ := session.Get(sessionName, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   viper.GetBool("server.ssl"),
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

// GetUsername retrieves the username from the specified session
func GetUsername(c *echo.Context, sessionName string) string {
	sess, _ := session.Get(sessionName, c)
	if sess == nil {
		return ""
	}
	if username, ok := sess.Values[UsernameKey].(string); ok {
		return username
	}
	return ""
}

// GetIsSuperAdmin retrieves the superadmin flag from the session
func GetIsSuperAdmin(c *echo.Context) bool {
	sess, _ := session.Get(SessionName, c)
	if sess == nil {
		return false
	}
	if isSuper, ok := sess.Values[IsSuperAdminKey].(bool); ok {
		return isSuper
	}
	return false
}

// ClearSession removes a session by its name
func ClearSession(c *echo.Context, sessionName string) error {
	sess, _ := session.Get(sessionName, c)
	// Reset options first to ensure MaxAge=-1 is applied even if Options was nil
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	// Clear all session values
	for key := range sess.Values {
		delete(sess.Values, key)
	}
	return sess.Save(c.Request(), c.Response())
}

// SetFlash stores a flash message in the user session
func SetFlash(c *echo.Context, key, value string) {
	sess, _ := session.Get(UserSessionName, c)
	sess.Values["flash_"+key] = value
	sess.Save(c.Request(), c.Response())
}

// GetFlash retrieves and clears a flash message from the user session
func GetFlash(c *echo.Context, key string) string {
	sess, _ := session.Get(UserSessionName, c)
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
