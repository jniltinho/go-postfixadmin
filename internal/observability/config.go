package observability

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Enabled                   bool
	TracesEnabled             bool
	LogsEnabled               bool
	MetricsEnabled            bool
	DatabaseTracesEnabled     bool
	TrustIncomingTraceContext bool
	ServiceName               string
	Environment               string
	TracesEndpoint            string
	LogsEndpoint              string
	MetricsEndpoint           string
	Authorization             string
	LogsStream                string
	TraceSampleRatio          float64
	MetricsExportInterval     time.Duration
	ExportTimeout             time.Duration
	ShutdownTimeout           time.Duration
	AllowInsecureHTTP         bool
}

var defaults = map[string]any{
	"enabled": false, "traces_enabled": false, "logs_enabled": false,
	"metrics_enabled": false, "database_traces_enabled": false, "metrics_endpoint": "",
	"metrics_export_interval":      "30s",
	"trust_incoming_trace_context": false, "allow_insecure_http": false,
	"service_name": "go-postfixadmin", "environment": "production",
	"traces_endpoint": "", "logs_endpoint": "", "authorization": "",
	"logs_stream": "postfixadmin", "trace_sample_ratio": 0.1,
	"export_timeout": "5s", "shutdown_timeout": "10s",
}

func BindConfig(v *viper.Viper) {
	for key, value := range defaults {
		name := "observability." + key
		v.SetDefault(name, value)
		_ = v.BindEnv(name, "OBSERVABILITY_"+strings.ToUpper(key))
	}
}

func LoadConfig(v *viper.Viper) (Config, error) {
	BindConfig(v)
	var cfg Config
	var err error
	cfg.Enabled, err = parseBool(v, "enabled")
	if err != nil || !cfg.Enabled {
		return cfg, err
	}
	for key, field := range map[string]*bool{
		"traces_enabled":               &cfg.TracesEnabled,
		"logs_enabled":                 &cfg.LogsEnabled,
		"metrics_enabled":              &cfg.MetricsEnabled,
		"database_traces_enabled":      &cfg.DatabaseTracesEnabled,
		"trust_incoming_trace_context": &cfg.TrustIncomingTraceContext,
		"allow_insecure_http":          &cfg.AllowInsecureHTTP,
	} {
		*field, err = parseBool(v, key)
		if err != nil {
			return Config{}, err
		}
	}
	cfg.ServiceName = v.GetString("observability.service_name")
	cfg.Environment = v.GetString("observability.environment")
	cfg.TracesEndpoint = v.GetString("observability.traces_endpoint")
	cfg.LogsEndpoint = v.GetString("observability.logs_endpoint")
	cfg.MetricsEndpoint = v.GetString("observability.metrics_endpoint")
	cfg.Authorization = v.GetString("observability.authorization")
	cfg.LogsStream = v.GetString("observability.logs_stream")
	cfg.TraceSampleRatio, err = strconv.ParseFloat(fmt.Sprint(v.Get("observability.trace_sample_ratio")), 64)
	if err != nil {
		return Config{}, configError("trace_sample_ratio")
	}
	for key, field := range map[string]*time.Duration{
		"export_timeout":          &cfg.ExportTimeout,
		"metrics_export_interval": &cfg.MetricsExportInterval,
		"shutdown_timeout":        &cfg.ShutdownTimeout,
	} {
		*field, err = time.ParseDuration(fmt.Sprint(v.Get("observability." + key)))
		if err != nil {
			return Config{}, configError(key)
		}
	}
	return cfg, cfg.Validate()
}

func parseBool(v *viper.Viper, key string) (bool, error) {
	value, err := strconv.ParseBool(fmt.Sprint(v.Get("observability." + key)))
	if err != nil {
		return false, configError(key)
	}
	return value, nil
}

func configError(key string) error {
	return fmt.Errorf("invalid observability.%s", key)
}

func (cfg Config) Validate() error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.DatabaseTracesEnabled && !cfg.TracesEnabled {
		return fmt.Errorf("observability.database_traces_enabled requires traces_enabled")
	}
	if !cfg.TracesEnabled && !cfg.LogsEnabled && !cfg.MetricsEnabled {
		return fmt.Errorf("observability.enabled requires traces_enabled, logs_enabled or metrics_enabled")
	}
	for _, key := range []string{
		"OTEL_RESOURCE_ATTRIBUTES", "OTEL_SERVICE_NAME",
		"OTEL_EXPORTER_OTLP_CERTIFICATE", "OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE", "OTEL_EXPORTER_OTLP_CLIENT_KEY",
		"OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE", "OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY",
		"OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE", "OTEL_EXPORTER_OTLP_LOGS_CLIENT_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY",
		"OTEL_EXPORTER_OTLP_METRICS_CERTIFICATE", "OTEL_EXPORTER_OTLP_METRICS_CLIENT_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_METRICS_CLIENT_KEY",
	} {
		if os.Getenv(key) != "" {
			return fmt.Errorf("unsupported telemetry environment setting %s; use observability configuration", key)
		}
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return configError("service_name")
	}
	if strings.TrimSpace(cfg.Authorization) == "" || !validHeader(cfg.Authorization) {
		return configError("authorization")
	}
	if cfg.LogsEnabled && (strings.TrimSpace(cfg.LogsStream) == "" || !validHeader(cfg.LogsStream)) {
		return configError("logs_stream")
	}
	if math.IsNaN(cfg.TraceSampleRatio) || math.IsInf(cfg.TraceSampleRatio, 0) {
		return configError("trace_sample_ratio")
	}
	if cfg.TraceSampleRatio < 0 || cfg.TraceSampleRatio > 1 {
		return configError("trace_sample_ratio")
	}
	if cfg.MetricsEnabled && cfg.MetricsExportInterval <= 0 {
		return configError("metrics_export_interval")
	}
	if cfg.ExportTimeout <= 0 {
		return configError("export_timeout")
	}
	if cfg.ShutdownTimeout <= 0 {
		return configError("shutdown_timeout")
	}
	for _, signal := range []struct {
		key      string
		enabled  bool
		endpoint string
	}{
		{key: "traces_endpoint", enabled: cfg.TracesEnabled, endpoint: cfg.TracesEndpoint},
		{key: "logs_endpoint", enabled: cfg.LogsEnabled, endpoint: cfg.LogsEndpoint},
		{key: "metrics_endpoint", enabled: cfg.MetricsEnabled, endpoint: cfg.MetricsEndpoint},
	} {
		if signal.enabled && !validEndpoint(signal.endpoint, cfg.AllowInsecureHTTP) {
			return configError(signal.key)
		}
	}
	return nil
}

func validHeader(value string) bool {
	for _, b := range []byte(value) {
		if b < 32 || b == 127 {
			return false
		}
	}
	return true
}

func validEndpoint(raw string, allowHTTP bool) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.Opaque != "" {
		return false
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && allowHTTP) {
		return false
	}
	if u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return false
	}
	return u.Path != "" && u.Path != "/"
}
