package handlers

import (
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// Simple IP-based rate limiter for login (fixed window, 5 attempts per minute per IP)
// This is lightweight and sufficient for v1 as per the migration plan.
var loginLimiter = rate.NewLimiter(rate.Every(time.Minute/5), 5) // rough global; we improve per-IP below

// For a small PR we use a very simple per-IP limiter using a map (not production perfect, but good enough)
var ipLimiters = make(map[string]*rate.Limiter)
var ipLimiterLastCleanup = time.Now()

func getIPLimiter(ip string) *rate.Limiter {
	// Very basic cleanup every 10 minutes
	if time.Since(ipLimiterLastCleanup) > 10*time.Minute {
		// naive cleanup - acceptable for this stage
		for k, lim := range ipLimiters {
			if lim.Allow() { // if it has tokens, consider active
				continue
			}
			delete(ipLimiters, k)
		}
		ipLimiterLastCleanup = time.Now()
	}

	if lim, ok := ipLimiters[ip]; ok {
		return lim
	}
	// 5 attempts per minute per IP
	lim := rate.NewLimiter(rate.Every(time.Minute/5), 5)
	ipLimiters[ip] = lim
	return lim
}

// AuthLogin godoc
// @Summary      Login
// @Description  Authenticates as Admin or Mailbox user. Returns JWT access token in body and sets httpOnly refresh cookie.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Credentials"
// @Success      200 {object} dto.LoginResponse "Success"
// @Failure      400 {object} dto.APIResponse "Bad Request"
// @Failure      401 {object} dto.APIResponse "Invalid credentials"
// @Failure      429 {object} dto.APIResponse "Too many requests"
// @Router       /auth/login [post]
func (h *Handler) AuthLogin(c *echo.Context) error {
	// Rate limiting (per IP)
	ip := c.RealIP()
	limiter := getIPLimiter(ip)
	if !limiter.Allow() {
		return dto.TooManyRequests(c, "too many login attempts, please try again later")
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	username := strings.TrimSpace(req.Username)
	password := req.Password

	if username == "" || password == "" {
		return dto.ValidationError(c, "username and password are required")
	}

	if h.DB == nil {
		return dto.WriteError(c, dto.ErrCodeInternal, "database unavailable")
	}

	// Try admin first (more privileged)
	var admin models.Admin
	err := h.DB.Where("username = ? AND active = ?", username, true).First(&admin).Error
	if err == nil {
		match, checkErr := utils.CheckPassword(password, admin.Password)
		if checkErr == nil && match {
			// Success - admin
			domains, isSuper, _ := repositories.GetAllowedDomains(h.DB, admin.Username, admin.Superadmin)

			accessToken, err := auth.GenerateAccessToken(admin.Username, isSuper, domains, "admin")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to generate token"})
			}
			refreshToken, err := auth.GenerateRefreshToken(admin.Username, isSuper, domains, "admin")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to generate refresh token"})
			}

			setRefreshCookie(c, refreshToken)

			return dto.WriteSuccess(c, dto.LoginResponse{
				AccessToken: accessToken,
				TokenType:   "Bearer",
				ExpiresIn:   int(auth.GetAccessTTL().Seconds()),
				User: dto.UserInfo{
					Username:   admin.Username,
					Type:       "admin",
					Superadmin: isSuper,
					Domains:    domains,
				},
			})
		}
	}

	// Try mailbox user
	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err == nil && mailbox.Active {
		match, checkErr := utils.CheckPassword(password, mailbox.Password)
		if checkErr == nil && match {
			// Success - mailbox user
			domains := []string{mailbox.Domain}

			accessToken, err := auth.GenerateAccessToken(mailbox.Username, false, domains, "mailbox")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to generate token"})
			}
			refreshToken, err := auth.GenerateRefreshToken(mailbox.Username, false, domains, "mailbox")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to generate refresh token"})
			}

			setRefreshCookie(c, refreshToken)

			return dto.WriteSuccess(c, dto.LoginResponse{
				AccessToken: accessToken,
				TokenType:   "Bearer",
				ExpiresIn:   int(auth.GetAccessTTL().Seconds()),
				User: dto.UserInfo{
					Username:   mailbox.Username,
					Type:       "mailbox",
					Superadmin: false,
					Domains:    domains,
				},
			})
		}
	}

	// Invalid credentials (same message for both paths to avoid enumeration)
	return dto.WriteError(c, dto.ErrCodeInvalidCredentials, "invalid credentials")
}

// AuthRefresh godoc
// @Summary      Refresh token
// @Description  Uses the httpOnly refresh cookie to issue a new access token (with rotation).
// @Tags         Authentication
// @Produce      json
// @Success      200 {object} dto.RefreshResponse
// @Failure      401 {object} dto.APIResponse
// @Router       /auth/refresh [post]
// @Security     BearerAuth
func (h *Handler) AuthRefresh(c *echo.Context) error {
	refreshCookie, err := c.Cookie("refresh_token")
	if err != nil || refreshCookie.Value == "" {
		return dto.Unauthorized(c, "missing refresh token")
	}

	claims, err := auth.ValidateToken(refreshCookie.Value)
	if err != nil {
		clearRefreshCookie(c)
		return dto.WriteError(c, dto.ErrCodeInvalidToken, "invalid or expired refresh token")
	}

	// Re-issue tokens (rotation) and refresh authorization data from the database.
	superadmin := claims.Superadmin
	domains := claims.Domains

	if claims.Type == "admin" {
		var admin models.Admin
		if err := h.DB.Where("username = ? AND active = ?", claims.Username, true).First(&admin).Error; err != nil {
			clearRefreshCookie(c)
			return dto.Unauthorized(c, "admin account is inactive or no longer exists")
		}

		var err error
		domains, superadmin, err = repositories.GetAllowedDomains(h.DB, admin.Username, admin.Superadmin)
		if err != nil {
			return dto.InternalError(c, "failed to refresh admin permissions")
		}
	} else if claims.Type == "mailbox" {
		mailbox, err := repositories.GetMailboxByUsername(h.DB, claims.Username)
		if err != nil || !mailbox.Active {
			clearRefreshCookie(c)
			return dto.Unauthorized(c, "mailbox account is inactive or no longer exists")
		}
		superadmin = false
		domains = []string{mailbox.Domain}
	}

	newAccess, err := auth.GenerateAccessToken(claims.Username, superadmin, domains, claims.Type)
	if err != nil {
		return dto.InternalError(c, "failed to generate access token")
	}

	newRefresh, err := auth.GenerateRefreshToken(claims.Username, superadmin, domains, claims.Type)
	if err != nil {
		return dto.InternalError(c, "failed to rotate refresh token")
	}

	setRefreshCookie(c, newRefresh)

	return dto.WriteSuccess(c, dto.RefreshResponse{
		AccessToken: newAccess,
		TokenType:   "Bearer",
		ExpiresIn:   int(auth.GetAccessTTL().Seconds()),
	})
}

// AuthLogout godoc
// @Summary      Logout
// @Description  Clears the httpOnly refresh token cookie.
// @Tags         Authentication
// @Produce      json
// @Success      200 {object} dto.APIResponse
// @Router       /auth/logout [post]
// @Security     BearerAuth
func (h *Handler) AuthLogout(c *echo.Context) error {
	clearRefreshCookie(c)
	return dto.WriteSuccess(c, map[string]bool{"success": true})
}

// AuthMe godoc
// @Summary      Get current user
// @Description  Returns information about the logged-in user extracted from the JWT.
// @Tags         Authentication
// @Produce      json
// @Success      200 {object} dto.UserInfo
// @Failure      401 {object} dto.APIResponse
// @Router       /auth/me [get]
// @Security     BearerAuth
func (h *Handler) AuthMe(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	return dto.WriteSuccess(c, dto.MeResponse{
		Username:   claims.Username,
		Type:       claims.Type,
		Superadmin: claims.Superadmin,
		Domains:    claims.Domains,
	})
}

// --- Cookie helpers ---

func setRefreshCookie(c *echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   viper.GetBool("server.ssl_enable"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.GetRefreshTTL().Seconds()),
	}
	c.SetCookie(cookie)
}

func clearRefreshCookie(c *echo.Context) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(cookie)
}
