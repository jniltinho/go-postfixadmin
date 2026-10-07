package observability

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func enabledConfig(t *testing.T) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.SetConfigType("toml")
	err := v.ReadConfig(strings.NewReader(`[observability]
enabled=true
logs_enabled=true
traces_enabled=true
authorization="Basic test"
logs_endpoint="https://example.test/custom/logs"
traces_endpoint="https://example.test/custom/traces"
`))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig(viper.New())
	if err != nil || cfg.Enabled || cfg.LogsEnabled || cfg.TracesEnabled || cfg.MetricsEnabled || cfg.DatabaseTracesEnabled {
		t.Fatalf("default config = %+v, error = %v", cfg, err)
	}
	v := viper.New()
	v.Set("observability.logs_enabled", true)
	v.Set("observability.authorization", "secret\ninvalid")
	v.Set("observability.export_timeout", "invalid")
	cfg, err = LoadConfig(v)
	if err != nil || cfg.Enabled {
		t.Fatalf("global gate: %v", err)
	}
}

func TestConfigEnvironmentBindings(t *testing.T) {
	cases := []struct{ name, value string }{
		{"enabled", "false"}, {"traces_enabled", "false"}, {"logs_enabled", "false"},
		{"metrics_enabled", "false"}, {"database_traces_enabled", "true"},
		{"metrics_endpoint", "https://test.invalid/custom/metrics"}, {"metrics_export_interval", "8s"},
		{"trust_incoming_trace_context", "true"}, {"allow_insecure_http", "true"},
		{"service_name", "test-service"}, {"environment", "test"},
		{"traces_endpoint", "https://test.invalid/custom/traces"},
		{"logs_endpoint", "https://test.invalid/custom/logs"},
		{"authorization", "Basic env-test"}, {"logs_stream", "env_stream"},
		{"trace_sample_ratio", "0.75"}, {"export_timeout", "3s"}, {"shutdown_timeout", "4s"},
	}
	for _, tc := range cases {
		for _, withTOML := range []bool{false, true} {
			name := tc.name + "/defaults"
			if withTOML {
				name = tc.name + "/toml"
			}
			t.Run(name, func(t *testing.T) {
				v := viper.New()
				if withTOML {
					v = enabledConfig(t)
				}
				t.Setenv("OBSERVABILITY_"+strings.ToUpper(tc.name), tc.value)
				BindConfig(v)
				if got := v.GetString("observability." + tc.name); got != tc.value {
					t.Fatalf("environment ignored: got %q", got)
				}
				if _, err := LoadConfig(v); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	v := enabledConfig(t)
	t.Setenv("OBSERVABILITY_EXPORT_TIMEOUT", "3s")
	t.Setenv("OBSERVABILITY_TRACE_SAMPLE_RATIO", "0.75")
	cfg, err := LoadConfig(v)
	if err != nil || cfg.ExportTimeout != 3*time.Second || cfg.TraceSampleRatio != 0.75 {
		t.Fatalf("typed env fields: %+v, %v", cfg, err)
	}
}

func TestConfigValidation(t *testing.T) {
	cases := []struct {
		name, key string
		value     any
	}{
		{"bool", "logs_enabled", "invalid-secret"},
		{"metric interval", "metrics_export_interval", "invalid-secret"},
		{"metric bool", "metrics_enabled", "invalid-secret"},
		{"db bool", "database_traces_enabled", "invalid-secret"},
		{"ratio text", "trace_sample_ratio", "invalid-secret"},
		{"ratio NaN", "trace_sample_ratio", "NaN"},
		{"ratio infinity", "trace_sample_ratio", "+Inf"},
		{"ratio negative", "trace_sample_ratio", -0.1},
		{"ratio above one", "trace_sample_ratio", 1.1},
		{"duration text", "export_timeout", "invalid-secret"},
		{"duration zero", "shutdown_timeout", "0s"},
		{"duration negative", "export_timeout", "-1s"},
		{"missing url", "logs_endpoint", ""},
		{"http", "logs_endpoint", "http://example.test/api/default/v1/logs"},
		{"userinfo", "logs_endpoint", "https://user:invalid-secret@example.test/logs"},
		{"query", "logs_endpoint", "https://example.test/logs?token=invalid-secret"},
		{"fragment", "logs_endpoint", "https://example.test/logs#invalid-secret"},
		{"no path", "logs_endpoint", "https://example.test"},
		{"root path", "logs_endpoint", "https://example.test/"},
		{"relative", "logs_endpoint", "/logs"},
		{"authorization missing", "authorization", ""},
		{"authorization newline", "authorization", "Basic invalid-secret\r\n"},
		{"stream header", "logs_stream", "invalid-secret\n"},
		{"stream empty", "logs_stream", ""},
		{"service empty", "service_name", " "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := enabledConfig(t)
			v.Set("observability."+tc.key, tc.value)
			_, err := LoadConfig(v)
			if err == nil || !strings.Contains(err.Error(), tc.key) || strings.Contains(err.Error(), "invalid-secret") {
				t.Fatalf("validation: %v", err)
			}
		})
	}
	t.Run("no signals", func(t *testing.T) {
		v := enabledConfig(t)
		v.Set("observability.logs_enabled", false)
		v.Set("observability.traces_enabled", false)
		if _, err := LoadConfig(v); err == nil {
			t.Fatal("no signals accepted")
		}
	})
	t.Run("disabled endpoint ignored", func(t *testing.T) {
		v := enabledConfig(t)
		v.Set("observability.traces_enabled", false)
		v.Set("observability.traces_endpoint", "invalid-secret")
		if _, err := LoadConfig(v); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("explicit http", func(t *testing.T) {
		v := enabledConfig(t)
		v.Set("observability.allow_insecure_http", true)
		v.Set("observability.logs_endpoint", "http://localhost:5080/api/default/v1/logs")
		if _, err := LoadConfig(v); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConfigTemplates(t *testing.T) {
	configs := make([]map[string]any, 0, 2)
	for _, path := range []string{"../../config.toml.example", "../../web/files/config.default.toml"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		v := viper.New()
		v.SetConfigType("toml")
		if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
			t.Fatal(err)
		}
		configs = append(configs, v.GetStringMap("observability"))
		cfg, err := LoadConfig(v)
		if err != nil || cfg.Enabled || cfg.LogsEnabled || cfg.TracesEnabled || cfg.MetricsEnabled || cfg.DatabaseTracesEnabled {
			t.Fatalf("%s: %v", path, err)
		}
		for key := range defaults {
			if !v.IsSet("observability." + key) {
				t.Fatalf("%s missing %s", path, key)
			}
		}
	}
	if !reflect.DeepEqual(configs[0], configs[1]) {
		t.Fatal("config templates differ")
	}
}

func TestUnsupportedSDKEnvironment(t *testing.T) {
	for _, key := range []string{"OTEL_RESOURCE_ATTRIBUTES", "OTEL_SERVICE_NAME", "OTEL_EXPORTER_OTLP_CERTIFICATE", "OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "private-value")
			_, err := LoadConfig(enabledConfig(t))
			if err == nil || !strings.Contains(err.Error(), key) || strings.Contains(err.Error(), "private-value") {
				t.Fatalf("SDK env: %v", err)
			}
			if _, err := LoadConfig(viper.New()); err != nil {
				t.Fatal("disabled config rejected SDK env")
			}
		})
	}
}

func TestMetricsOnlyConfiguration(t *testing.T) {
	v := enabledConfig(t)
	v.Set("observability.traces_enabled", false)
	v.Set("observability.logs_enabled", false)
	v.Set("observability.metrics_enabled", true)
	v.Set("observability.metrics_endpoint", "https://observe.test/api/default/v1/metrics")
	if _, err := LoadConfig(v); err != nil {
		t.Fatal(err)
	}
	for key, value := range map[string]any{"metrics_endpoint": "", "metrics_export_interval": "0s", "database_traces_enabled": true} {
		t.Run(key, func(t *testing.T) {
			previous := v.Get("observability." + key)
			v.Set("observability."+key, value)
			defer v.Set("observability."+key, previous)
			if _, err := LoadConfig(v); err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("validation=%v", err)
			}
		})
	}
}
