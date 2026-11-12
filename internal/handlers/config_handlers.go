package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/rbac"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// configField describes one config.toml key shown in Settings > Configuration. Only Editable keys can be changed
// from the panel; everything else is shown read-only, and Secret values never leave the server.
type configField struct {
	Key      string // section.key
	Kind     string // bool, int, string, enum, duration, list
	Editable bool
	Secret   bool
	Restart  bool     // takes effect after a restart
	Env      string   // environment variable that overrides the file (then not editable)
	Options  []string // enum values
	Min, Max int64    // int range; duration range in seconds
	Pattern  *regexp.Regexp
}

var (
	hostPattern  = regexp.MustCompile(`^[A-Za-z0-9.:\[\]-]{1,253}$`)
	themePattern = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)
)

// configSchema: the editable keys are operational settings read at request time (or flagged Restart); database,
// security, delivery routing, RBAC, the terminal target, backup and telemetry stay read-only.
var configSchema = []configField{
	{Key: "ui.default_theme", Kind: "string", Editable: true, Env: "UI_DEFAULT_THEME", Pattern: themePattern},
	{Key: "vacation.enabled", Kind: "bool", Editable: true},
	{Key: "smtp.server", Kind: "string", Editable: true, Pattern: hostPattern},
	{Key: "smtp.port", Kind: "int", Editable: true, Min: 1, Max: 65535},
	{Key: "smtp.type", Kind: "enum", Editable: true, Options: []string{"plain", "tls", "starttls"}},
	{Key: "server.swagger_enable", Kind: "bool", Editable: true, Restart: true},
	{Key: "terminal.idle_timeout", Kind: "duration", Editable: true, Min: 60, Max: 8 * 3600},
	// on/off only: opening a shell still needs a superadmin, the host's SSH credentials and its pinned key; host and
	// port stay read-only so the panel cannot be pointed at another machine
	{Key: "terminal.enabled", Kind: "bool", Editable: true},

	{Key: "server.port", Kind: "int"},
	{Key: "server.ssl_enable", Kind: "bool"},
	{Key: "server.ssl_cert", Kind: "string"},
	{Key: "server.ssl_key", Kind: "string"}, // a path, not the key itself
	{Key: "server.cleanup_maildir", Kind: "bool"},
	{Key: "server.jwt_access_ttl", Kind: "duration"},
	{Key: "server.jwt_refresh_ttl", Kind: "duration"},
	{Key: "server.session_idle_timeout", Kind: "duration"},
	{Key: "server.trusted_proxies", Kind: "list"},
	{Key: "server.session_secret", Kind: "string", Secret: true},
	{Key: "server.jwt_secret", Kind: "string", Secret: true},
	{Key: "database.driver", Kind: "string"},
	{Key: "database.host", Kind: "string"},
	{Key: "database.port", Kind: "string"},
	{Key: "database.name", Kind: "string"},
	{Key: "database.user", Kind: "string"},
	{Key: "database.pass", Kind: "string", Secret: true},
	{Key: "database.url", Kind: "string", Secret: true},
	{Key: "database.debug", Kind: "bool"},
	{Key: "quota.multiplier", Kind: "int"},
	{Key: "rbac.enabled", Kind: "bool"},
	{Key: "transport.host", Kind: "string"},
	{Key: "transport.cache", Kind: "duration"},
	{Key: "transport.hostname", Kind: "string"},
	{Key: "transport.localdelivery", Kind: "string"},
	{Key: "transport.delivery", Kind: "string"},
	{Key: "terminal.host", Kind: "string"},
	{Key: "terminal.port", Kind: "int"},
	{Key: "backup.backup_dir", Kind: "string"},
	{Key: "backup.mysql_pass", Kind: "string", Secret: true},
	{Key: "backup.smtp_pass", Kind: "string", Secret: true},
	{Key: "observability.enabled", Kind: "bool"},
	{Key: "observability.authorization", Kind: "string", Secret: true},
}

var configMu sync.Mutex

func schemaField(key string) (configField, bool) {
	i := slices.IndexFunc(configSchema, func(f configField) bool { return f.Key == key })
	if i < 0 {
		return configField{}, false
	}
	return configSchema[i], true
}

// editable reports whether the panel may change f now (an environment variable would win over the file).
func (f configField) editable() bool {
	return f.Editable && (f.Env == "" || os.Getenv(f.Env) == "")
}

// configWritable reports whether config.toml exists and can be rewritten in place (file and directory).
func configWritable() (string, bool) {
	path := viper.ConfigFileUsed()
	if path == "" {
		return "", false
	}
	if st, err := os.Stat(path); err != nil || st.IsDir() {
		return path, false
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return path, false
	}
	f.Close()
	probe, err := os.CreateTemp(filepath.Dir(path), ".config-probe-*")
	if err != nil {
		return path, false
	}
	probe.Close()
	os.Remove(probe.Name())
	return path, true
}

// displayValue is the value in effect, not just the file's: TLS is switched on by ssl_cert + ssl_key when
// ssl_enable is not set (cmd/server.go), and the JWT lifetimes have defaults in internal/auth.
func displayValue(f configField) any {
	if f.Secret {
		return nil
	}
	switch f.Key {
	case "server.ssl_enable":
		if !viper.IsSet(f.Key) {
			return viper.GetString("server.ssl_cert") != "" && viper.GetString("server.ssl_key") != ""
		}
	case "server.jwt_access_ttl":
		return shortDuration(auth.GetAccessTTL())
	case "server.jwt_refresh_ttl":
		return shortDuration(auth.GetRefreshTTL())
	case "server.session_idle_timeout":
		return shortDuration(auth.GetIdleTimeout())
	}
	switch f.Kind {
	case "bool":
		return viper.GetBool(f.Key)
	case "int":
		return viper.GetInt64(f.Key)
	case "list":
		return viper.GetStringSlice(f.Key)
	default:
		return viper.GetString(f.Key)
	}
}

// shortDuration writes 15m and 168h instead of 15m0s and 168h0m0s.
func shortDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	return s
}

// GetConfigV1 godoc
// @Summary      Panel configuration
// @Description  config.toml values shown in Settings > Configuration: which keys are editable, read-only or secret (secret values are never returned). Superadmins only.
// @Tags         System
// @Produce      json
// @Success      200 {object} dto.ConfigResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Router       /system/config [get]
// @Security     BearerAuth
func (h *Handler) GetConfigV1(c *echo.Context) error {
	if !superadminWith(middleware.GetJWTClaims(c), rbac.PermSettingsRead) {
		return dto.WriteError(c, dto.ErrCodeForbidden, "superadmins only")
	}
	path, writable := configWritable()
	resp := dto.ConfigResponse{File: path, Writable: writable, Fields: []dto.ConfigField{}}
	for _, f := range configSchema {
		item := dto.ConfigField{Key: f.Key, Kind: f.Kind, Editable: f.editable() && writable, Config: f.Editable, Secret: f.Secret,
			Restart: f.Restart, Options: f.Options, Value: displayValue(f)}
		if f.Secret {
			set := viper.GetString(f.Key) != ""
			item.Set = &set
		}
		if f.Env != "" && os.Getenv(f.Env) != "" {
			item.Env = f.Env
		}
		if f.Kind == "int" || f.Kind == "duration" {
			item.Min, item.Max = f.Min, f.Max
		}
		resp.Fields = append(resp.Fields, item)
	}
	return dto.WriteSuccess(c, resp)
}

// parseConfigValue validates v for f and returns the value for viper and its TOML literal.
func parseConfigValue(f configField, v any) (any, string, error) {
	switch f.Kind {
	case "bool":
		b, ok := v.(bool)
		if !ok {
			return nil, "", fmt.Errorf("%s must be true or false", f.Key)
		}
		return b, strconv.FormatBool(b), nil
	case "int":
		n, ok := v.(float64) // JSON numbers
		if !ok || n != float64(int64(n)) || int64(n) < f.Min || int64(n) > f.Max {
			return nil, "", fmt.Errorf("%s must be a whole number between %d and %d", f.Key, f.Min, f.Max)
		}
		return int64(n), strconv.FormatInt(int64(n), 10), nil
	case "enum":
		s, ok := v.(string)
		if !ok || !slices.Contains(f.Options, s) {
			return nil, "", fmt.Errorf("%s must be one of %s", f.Key, strings.Join(f.Options, ", "))
		}
		return s, strconv.Quote(s), nil
	case "duration":
		s, ok := v.(string)
		d, err := time.ParseDuration(s)
		if !ok || err != nil || d < time.Duration(f.Min)*time.Second || d > time.Duration(f.Max)*time.Second {
			return nil, "", fmt.Errorf("%s must be a duration between %s and %s (e.g. 30m)", f.Key,
				time.Duration(f.Min)*time.Second, time.Duration(f.Max)*time.Second)
		}
		return s, strconv.Quote(s), nil
	case "string":
		s, ok := v.(string)
		if !ok || (f.Pattern != nil && !f.Pattern.MatchString(s)) {
			return nil, "", fmt.Errorf("%s has an invalid value", f.Key)
		}
		return s, strconv.Quote(s), nil
	}
	return nil, "", fmt.Errorf("%s cannot be changed", f.Key)
}

// UpdateConfigV1 godoc
// @Summary      Update panel configuration
// @Description  Changes editable config.toml keys (comments and formatting are kept, the previous file is saved as config.toml.YYYYMMDD-HHMMSS.bak). Read-only and unknown keys are refused. Superadmins only.
// @Tags         System
// @Accept       json
// @Produce      json
// @Param        body body dto.ConfigUpdateRequest true "keys to change"
// @Success      200 {object} dto.ConfigUpdateResponse
// @Failure      400 {object} dto.APIResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Router       /system/config [put]
// @Security     BearerAuth
func (h *Handler) UpdateConfigV1(c *echo.Context) error {
	cl := middleware.GetJWTClaims(c)
	if !superadminWith(cl, rbac.PermSettingsWrite) {
		return dto.WriteError(c, dto.ErrCodeForbidden, "superadmins only")
	}
	var req dto.ConfigUpdateRequest
	if err := c.Bind(&req); err != nil || len(req.Changes) == 0 {
		return dto.BadRequest(c, "no changes")
	}
	path, writable := configWritable()
	if !writable {
		return dto.WriteError(c, dto.ErrCodeForbidden, "config.toml is not writable by the panel")
	}
	keys := make([]string, 0, len(req.Changes))
	for k := range req.Changes {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	configMu.Lock()
	defer configMu.Unlock()
	raw, err := os.ReadFile(path)
	if err != nil {
		return dto.WriteError(c, dto.ErrCodeInternal, "could not read config.toml")
	}
	doc := string(raw)
	values := map[string]any{}
	restart := false
	for _, k := range keys {
		f, ok := schemaField(k)
		if !ok || !f.editable() {
			return dto.ValidationError(c, k+" cannot be changed from the panel")
		}
		val, lit, err := parseConfigValue(f, req.Changes[k])
		if err != nil {
			return dto.ValidationError(c, err.Error())
		}
		section, key, _ := strings.Cut(k, ".")
		if doc, err = utils.SetTOMLValue(doc, section, key, lit); err != nil {
			return dto.ValidationError(c, err.Error())
		}
		values[k] = val
		restart = restart || f.Restart
	}
	// the new file must parse and carry exactly the new values before it replaces the old one
	check := viper.New()
	check.SetConfigType("toml")
	if err := check.ReadConfig(strings.NewReader(doc)); err != nil {
		return dto.WriteError(c, dto.ErrCodeInternal, "the updated config.toml would not parse; nothing was changed")
	}
	for k, v := range values {
		if fmt.Sprint(check.Get(k)) != fmt.Sprint(v) {
			return dto.WriteError(c, dto.ErrCodeInternal, "could not update "+k+"; nothing was changed")
		}
	}
	if err := writeFileAtomic(path, []byte(doc), raw); err != nil {
		return dto.WriteError(c, dto.ErrCodeInternal, "could not write config.toml")
	}
	for k, v := range values {
		viper.Set(k, v) // effective now for keys read at request time
	}
	_ = utils.LogAction(h.DB, cl.Username, c.RealIP(), "", "config_update", strings.Join(keys, ", "))
	return dto.WriteSuccess(c, dto.ConfigUpdateResponse{Updated: keys, RestartRequired: restart})
}

// writeFileAtomic keeps the previous content as path.YYYYMMDD-HHMMSS.bak (never overwriting an older backup) and
// replaces path through a temp file in the same directory, with the original permissions.
func writeFileAtomic(path string, data, previous []byte) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if err := writeBackup(path, previous, st.Mode().Perm()); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.toml")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), st.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func writeBackup(path string, data []byte, perm os.FileMode) error {
	stamp := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("%s.%s.bak", path, stamp)
	for i := 2; ; i++ {
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
		if errors.Is(err, fs.ErrExist) {
			name = fmt.Sprintf("%s.%s-%d.bak", path, stamp, i)
			continue
		}
		if err != nil {
			return err
		}
		if _, err := f.Write(data); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	}
}
