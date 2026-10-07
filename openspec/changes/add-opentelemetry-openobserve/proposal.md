## Why

Go-PostfixAdmin needs centralized application logs, HTTP/database traces and metrics to diagnose errors and latency across the API. OpenObserve can receive all three signals using OpenTelemetry, but operators need configurable ingestion URLs, credentials, and an explicit opt-in before sending telemetry outside the server.

## What Changes

- Add optional OpenTelemetry Go SDK pipelines for HTTP/database traces, metrics and application logs using OTLP/HTTP protobuf and OpenObserve-compatible endpoints.
- Add an `[observability]` configuration section with separate `traces_endpoint`, `logs_endpoint` and `metrics_endpoint` URLs, global and per-signal enablement (all default to false), service identity, authentication, log stream, sampling, and export/shutdown timeouts.
- Instrument the Echo v5 HTTP server with opt-in trusted W3C trace context propagation (incoming context is ignored by default) and route-based span names; correlate context-aware `log/slog` records with trace/span IDs.
- Propagate request context through GORM and API-key lookups; export database operations without SQL or parameters.
- Add HTTP/database counters and latency histograms plus pool connection gauges, independently enabled and disabled by default.
- Preserve local logging and existing debug levels while exporting logs asynchronously with bounded buffering.
- Flush telemetry after HTTP requests drain on shutdown and keep the application available when the destination is unavailable.
- Document configuration in both config templates, generated configuration, and deployment/development guides during implementation.

Implementation includes an isolated Docker environment with fake data and OpenObserve ingestion validation.

## Capabilities

### New Capabilities

- `telemetry-configuration`: Opt-in configuration of OpenObserve ingestion URLs, authentication, service identity, and exporter limits through TOML and environment variables.
- `opentelemetry-export`: HTTP/database tracing, metrics and correlated application log export with safe lifecycle management.

### Modified Capabilities

None. There are no existing OpenSpec capabilities in this repository.

## Impact

- New `internal/observability/` package; integration in `cmd/root.go` and `internal/server/server.go`; targeted context-aware logging in HTTP paths; request-context propagation in internal/handlers/* and internal/middleware/jwt.go. Existing repository signatures remain unchanged.
- Configuration examples: `config.toml.example` and `web/files/config.default.toml`, which supplies `--generate-config`.
- Docker deployment/test stack: docker-compose.yml, docker-compose.observability.yml, Dockerfile and tests/observability/; shipped systemd grace period in DOCUMENTS/setup/postfixadmin.service.
- Documentation: `README.md`, `DEVELOPMENT.md`, `FEATURES.md`, and `DOCUMENTS/setup/README.md` and `DOCUMENTS/setup/OBSERVABILITY.md`.
- Dependencies: OpenTelemetry API/SDK, OTLP HTTP trace/log/metric exporters, and the `otelslog` bridge. Select compatible pinned releases when implementing.
- No database migrations, new REST endpoints, permission changes, or frontend changes. Keep the transport TCP server's existing zerolog logging.

## References

- [OpenObserve Go trace ingestion](https://openobserve.ai/docs/ingestion/traces/go/)
- [OpenTelemetry Go API and SDK](https://github.com/open-telemetry/opentelemetry-go)
- [OpenObserve OTLP ingestion](https://openobserve.ai/docs/ingestion/logs/otlp/)
- [OpenObserve OTLP logs API](https://openobserve.ai/docs/reference/api/ingestion/logs/otlp/)
