## 1. Configuration and Dependencies

- [x] 1.1 Pin compatible OpenTelemetry Go API/SDK, OTLP HTTP trace/log/metric exporters, and otelslog versions; verify Go 1.26 and Echo v5 compatibility.
- [x] 1.2 Implement typed observability configuration with Viper defaults and explicit OBSERVABILITY_* bindings, using the contract in design.md; global, traces, logs, metrics and database tracing switches all default to false.
- [x] 1.3 Add startup validation for enabled signals, endpoints, authentication, TLS/HTTP policy, sampling, stream, positive timeouts, ingestion paths and valid header characters without exposing secrets.
- [x] 1.4 Update config.toml.example and web/files/config.default.toml together; verify --generate-config emits the same disabled settings and commented URL examples.

## 2. Export Pipelines and Lifecycle

- [x] 2.1 Implement internal/observability manager with service resources, parent-based sampling, full-URL OTLP/HTTP exporters, Authorization and logs stream-name headers.
- [x] 2.2 Configure bounded batch queues, timeouts/retries, nonblocking overflow, and rate-limited local-only export diagnostics.
- [x] 2.3 Set deployment grace periods (Docker Compose and shipped systemd units) to at least 30s for the default combined HTTP/telemetry shutdown budget; initialize providers for the server and release them on startup failure; coordinate HTTP drain before idempotent bounded flush on SIGINT/SIGTERM.
- [x] 2.4 Preserve disabled behavior and local logging when export is unavailable.

## 3. HTTP Tracing and Correlated Logs

- [x] 3.1 Add Echo v5-compatible middleware for W3C propagation, request context, route-template server spans, final status, and error recording.
- [x] 3.2 Add slog local/OTel fan-out preserving levels, groups and attributes; ensure request logs use the span context.
- [x] 3.3 Implement remote attribute allowlist and audited static-message filtering; review exported attributes/messages; exclude credentials, mailbox identities, bodies, queries, raw SQL, and DSNs.

## 4. Verification

- [x] 4.1 Test malformed boolean/ratio/duration rejection and every OBSERVABILITY_* binding with and without TOML, conflicting OTEL_* settings, disabled defaults, signal toggles, TOML/env precedence, custom paths, invalid configuration, and generated-config parity.
- [x] 4.2 Use fake OTLP receivers to verify protobuf payloads, exact URL paths, headers, resource identity, log stream, and trace/log correlation.
- [x] 4.3 Verify exactly one final error request log in logs-only mode, unchanged disabled local logging, and trusted/untrusted incoming parent context, logs-only/unsampled correlation, recovered panics, root sampling boundaries, independent log export, 404 fallback, and final error response status under Echo v5.
- [x] 4.4 Verify local level filtering, handler groups/attributes, secret redaction, outage behavior, bounded queues, nonrecursive diagnostics, startup cleanup including failed port bind, and drain-before-flush ordering with one shared telemetry timeout.
- [x] 4.5 Run go test ./... and the required project build; confirm no generated assets or secrets enter the change.
- [x] 4.6 Smoke-test traces and correlated logs in a test OpenObserve organization; confirm stream, sampling, unavailable destination behavior, and shutdown flush.

## 5. Documentation During Implementation

- [x] 5.1 Update README.md, DEVELOPMENT.md, FEATURES.md, and DOCUMENTS/setup/README.md with opt-in setup, all config keys/env variables, full endpoint examples, authentication, credential-file permissions, TLS/HTTP policy, incoming context trust, and deployment shutdown grace periods.
- [x] 5.2 Document sampling, potential queue drops, exporter troubleshooting, rollback, and application-logs-only scope.

## 6. Database traces and application metrics

- [x] 6.1 Add disabled-by-default metric/database tracing switches, full metric endpoint and positive export interval.
- [x] 6.2 Propagate HTTP context through all handler database calls and API-key lookup using isolated GORM sessions.
- [x] 6.3 Instrument GORM operations with safe CLIENT spans inheriting HTTP parents and excluding SQL, parameters and raw errors.
- [x] 6.4 Export HTTP/database counters and duration histograms plus connection-pool gauges via periodic OTLP/HTTP.
- [x] 6.5 Test parent correlation, cancellation, shared-session isolation, errors, metric payloads and enabled/disabled modes.
- [x] 6.6 Validate real database traces and metric streams in Docker OpenObserve; update docs and review with Claude CLI.
- [x] 6.7 Run required production build, race tests and strict OpenSpec validation; verify secrets and generated artifacts are excluded.
- [x] 6.8 Commit implementation and documentation with the confirmed Git author identity.
