package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
)

const (
	SessionName       = "session"
	UserSessionName   = "user_session" // New session for users
	AuthKey           = "authenticated"
	UserAuthKey       = "user_authenticated" // Key for user auth
	UsernameKey       = "username"
	UserUsernameKey   = "user_username" // Key for user username
	LastActivityKey   = "last_activity"
	InactivityTimeout = 60 * time.Minute
)

// AuthMiddleware checks for admin session validity and inactivity
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Skip check for login page and static assets
		if c.Path() == "/login" || c.Path() == "/static/*" || c.Path() == "/users/login" || c.Path() == "/users/logout" {
			return next(c)
		}

		sess, _ := session.Get(SessionName, c)
		if sess == nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		// Check if authenticated
		auth, ok := sess.Values[AuthKey].(bool)
		if !ok || !auth {
			return c.Redirect(http.StatusFound, "/login")
		}

		// Check inactivity
		lastActivity, ok := sess.Values[LastActivityKey].(int64)
		if ok {
			lastActivityTime := time.Unix(lastActivity, 0)
			if time.Since(lastActivityTime) > InactivityTimeout {
				// Session expired
				sess.Options.MaxAge = -1
				sess.Save(c.Request(), c.Response())
				return c.Redirect(http.StatusFound, "/login?expired=true")
			}
		}

		// Update last activity
		sess.Values[LastActivityKey] = time.Now().Unix()
		sess.Save(c.Request(), c.Response())

		return next(c)
	}
}

// UserAuthMiddleware checks for user session validity and inactivity
func UserAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Skip check for login page and static assets
		if c.Path() == "/users/login" || c.Path() == "/static/*" {
			return next(c)
		}

		sess, _ := session.Get(UserSessionName, c)
		if sess == nil {
			return c.Redirect(http.StatusFound, "/users/login")
		}

		// Check if authenticated
		auth, ok := sess.Values[UserAuthKey].(bool)
		if !ok || !auth {
			return c.Redirect(http.StatusFound, "/users/login")
		}

		// Check inactivity
		lastActivity, ok := sess.Values[LastActivityKey].(int64)
		if ok {
			lastActivityTime := time.Unix(lastActivity, 0)
			if time.Since(lastActivityTime) > InactivityTimeout {
				// Session expired
				sess.Options.MaxAge = -1
				sess.Save(c.Request(), c.Response())
				return c.Redirect(http.StatusFound, "/users/login?expired=true")
			}
		}

		// Update last activity
		sess.Values[LastActivityKey] = time.Now().Unix()
		sess.Save(c.Request(), c.Response())

		return next(c)
	}
}

// SetSession authenticates the admin and sets initial session values
func SetSession(c *echo.Context, username string) error {
	sess, _ := session.Get(SessionName, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	sess.Values[AuthKey] = true
	sess.Values[UsernameKey] = username
	sess.Values[LastActivityKey] = time.Now().Unix()
	return sess.Save(c.Request(), c.Response())
}

// SetUserSession authenticates the user and sets initial session values
func SetUserSession(c *echo.Context, username string) error {
	sess, _ := session.Get(UserSessionName, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	sess.Values[UserAuthKey] = true
	sess.Values[UserUsernameKey] = username
	sess.Values[LastActivityKey] = time.Now().Unix()
	return sess.Save(c.Request(), c.Response())
}

// GetUsername retrieves the admin username from the session
func GetUsername(c *echo.Context) string {
	sess, _ := session.Get(SessionName, c)
	if sess == nil {
		return ""
	}
	if username, ok := sess.Values[UsernameKey].(string); ok {
		return username
	}
	return ""
}

// GetUser retrieves the user username from the session
func GetUser(c *echo.Context) string {
	sess, _ := session.Get(UserSessionName, c)
	if sess == nil {
		return ""
	}
	if username, ok := sess.Values[UserUsernameKey].(string); ok {
		return username
	}
	return ""
}

// ClearSession removes the admin session
func ClearSession(c *echo.Context) error {
	sess, _ := session.Get(SessionName, c)
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}

// ClearUserSession removes the user session
func ClearUserSession(c *echo.Context) error {
	sess, _ := session.Get(UserSessionName, c)
	sess.Options.MaxAge = -1
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
