package handlers

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"

	"github.com/labstack/echo/v5"
)

// mailServices is the fixed allowlist of units the panel reports; names never come from the request.
var mailServices = []string{"postfix.service", "dovecot.service", "rspamd.service"}

// ListServicesV1 godoc
// @Summary      Mail service status
// @Description  Read-only systemd state of Postfix, Dovecot and Rspamd on the panel's host (systemctl show).
// @Tags         System
// @Produce      json
// @Success      200 {object} dto.ServicesResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Router       /system/services [get]
// @Security     BearerAuth
func (h *Handler) ListServicesV1(c *echo.Context) error {
	// RequirePermission is a no-op with RBAC off: keep mailbox (user portal) tokens out regardless
	if cl := middleware.GetJWTClaims(c); cl == nil || cl.Type != "admin" {
		return dto.WriteError(c, dto.ErrCodeForbidden, "admins only")
	}
	resp := dto.ServicesResponse{Services: []dto.ServiceStatus{}}
	if _, err := exec.LookPath("systemctl"); err != nil {
		resp.Message = "systemctl not available on this host"
		return dto.WriteSuccess(c, resp)
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()
	resp.Available = true
	for _, name := range mailServices {
		out, err := exec.CommandContext(ctx, "systemctl", "show", name,
			"--property=LoadState,ActiveState,SubState,UnitFileState,ActiveEnterTimestamp").Output()
		if err != nil {
			resp.Available = false
			resp.Message = "systemctl show failed: " + err.Error()
			resp.Services = append(resp.Services, dto.ServiceStatus{Name: name})
			continue
		}
		resp.Services = append(resp.Services, parseUnitShow(name, string(out)))
	}
	return dto.WriteSuccess(c, resp)
}

// parseUnitShow reads `systemctl show --property=...` output (key=value lines); a unit not installed on the host
// has LoadState=not-found and is reported with Found=false.
func parseUnitShow(name, out string) dto.ServiceStatus {
	f := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), "="); ok {
			f[k] = v
		}
	}
	return dto.ServiceStatus{
		Name:    name,
		Active:  f["ActiveState"],
		Sub:     f["SubState"],
		Enabled: f["UnitFileState"],
		Since:   f["ActiveEnterTimestamp"],
		Found:   f["LoadState"] != "" && f["LoadState"] != "not-found",
	}
}
