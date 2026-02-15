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
	AuthKey           = "authenticated"
	UsernameKey       = "username" // New key for storing username
	LastActivityKey   = "last_activity"
	InactivityTimeout = 60 * time.Minute
)

// AuthMiddleware checks for session validity and inactivity
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Skip check for login page and static assets
		if c.Path() == "/login" || c.Path() == "/static/*" {
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

// SetSession authenticates the user and sets initial session values
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

// GetUsername retrieves the username from the session
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

// ClearSession removes the session
func ClearSession(c *echo.Context) error {
	sess, _ := session.Get(SessionName, c)
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}
