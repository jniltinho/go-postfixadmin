## Context

`cmd/root.go` configures Viper and the default slog text handler. `internal/server/server.go` creates Echo v5, registers request logging, and handles SIGINT/SIGTERM with HTTP shutdown. The example TOML and embedded default TOML are maintained together. The existing `specs/` documents remain in place; this proposal introduces the standard OpenSpec change structure.

The OpenObserve Go example demonstrates OTLP/HTTP traces with an organization-specific URL path and authorization header. The OTLP logs documentation defines a separate logs endpoint and the `stream-name` header. Use the official Go SDK rather than copying the example's Gin integration or hardcoded credentials.

## Goals / Non-Goals

**Goals:**

- Export HTTP/database traces, application metrics and structured logs to self-hosted or cloud OpenObserve.
- Make the server URLs configurable without rebuilding the binary.
- Correlate request logs with spans while retaining local logging.
- Bound export resource use and shutdown duration.

**Non-Goals:**

- Profiling and browser/RUM instrumentation.
- Forwarding Postfix/Dovecot log files, changing the existing mail-log UI, or instrumenting the transport TCP server and all CLI commands.
- New frontend settings or ingestion/query APIs.

## Decisions

### 1. Separate full OTLP/HTTP URLs

Use `otlptracehttp`, `otlploghttp` and `otlpmetrichttp` with full endpoint URLs (using `WithEndpointURL` or an equivalent explicit host/path configuration supported by the selected release). Preserve scheme, host, optional port, and path exactly; never append another `/v1/logs` or `/v1/traces`. This supports organization paths, reverse proxies, and separate destinations. OTLP/gRPC and automatic failover across multiple hosts are outside this proposal.

### 2. Configuration contract

Add the following section to both configuration templates during implementation. Empty endpoints prevent accidental export; the commented URLs show a self-hosted example. Operators obtain cloud URLs and authorization values from their OpenObserve ingestion settings.

```toml
[observability]
enabled = false
service_name = "go-postfixadmin"
environment = "production"
traces_enabled = false
logs_enabled = false
metrics_enabled = false
database_traces_enabled = false
metrics_export_interval = "30s"
trust_incoming_trace_context = false
traces_endpoint = ""
logs_endpoint = ""
metrics_endpoint = ""
# metrics_endpoint = "https://observe.example.com/api/default/v1/metrics"
# traces_endpoint = "https://observe.example.com/api/default/v1/traces"
# logs_endpoint = "https://observe.example.com/api/default/v1/logs"
authorization = "" # Prefer OBSERVABILITY_AUTHORIZATION via environment
logs_stream = "postfixadmin"
trace_sample_ratio = 0.1
export_timeout = "5s"
shutdown_timeout = "10s"
allow_insecure_http = false
```

Read fields using fully qualified Viper keys (for example `viper.GetString("observability.logs_endpoint")`), rather than `Sub` or `UnmarshalKey`, so environment bindings are honored. For booleans, ratios and durations, retrieve the resolved raw value with `viper.Get` and convert with error-returning parsing (`strconv`, `time.ParseDuration`, or equivalent), accepting correctly typed TOML/default values. Reject malformed strings rather than using zero-valued fallback from `GetBool`, `GetFloat64` or `GetDuration`.

Bind each key explicitly to its `OBSERVABILITY_<UPPERCASE_KEY>` environment variable using Viper, including `OBSERVABILITY_TRACES_ENDPOINT`, `OBSERVABILITY_LOGS_ENDPOINT`, and `OBSERVABILITY_AUTHORIZATION`. Precedence is environment > TOML > defaults. Do not add CLI flags or imply standard `OTEL_*` variables override this contract; configure SDK options explicitly and document project variable names. Construct resources without `resource.Default()` or `WithFromEnv`. Explicitly configure exporter headers, scheme, path, protobuf encoding, compression, TLS verification, timeouts and batch limits. Do not mutate process environment. Test conflicting SDK environment variables; the pinned SDK consumes resource and TLS certificate environment settings before or in addition to explicit options. Enabled startup therefore rejects `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, and `OTEL_EXPORTER_OTLP_{,TRACES_,LOGS_,METRICS_}{CERTIFICATE,CLIENT_CERTIFICATE,CLIENT_KEY}` with key-only diagnostics. Other SDK endpoint/header/sampler/batch settings are overridden by explicit project options.

The global switch gates all three signals. To send traces, set both `enabled = true` and `traces_enabled = true`; to send logs, set both `enabled = true` and `logs_enabled = true`. Global, per-signal and database trace switches all default to false. Database tracing requires trace enablement; metrics require their endpoint and a positive export interval. Changing configuration takes effect on server restart; runtime hot reload is outside scope.

When enabled, require at least one enabled signal, a nonempty service name, an absolute URL for every enabled signal, and a nonempty authorization value for every enabled exporter. Validate finite sampling ratio in `[0,1]`, positive duration strings, and nonempty log stream when logs are enabled. Disabled signal endpoints are ignored. HTTPS uses certificate verification; HTTP requires `allow_insecure_http = true`. Require a non-root ingestion path and a valid HTTP header value for authorization (no CR/LF or control characters). Reject URL userinfo, query strings, and fragments; credentials belong in headers. Do not introduce a skip-certificate-verification option.

Send `Authorization` exactly as configured (for example `Basic <base64(email:password)>`) without logging it; add `stream-name` to the logs exporter. Organization is already part of each configured URL. Templates must contain no real tokens. Recommend environment credentials; document permission 0600 when an operator stores authorization in TOML. An explicit HTTP opt-in emits a local warning that transport is unencrypted, without printing the destination or credentials. Invalid enabled configuration returns a startup error with the key name, never its secret value; disabled telemetry ignores exporter settings and creates no exporters.

### 3. Providers and lifecycle

Create a server-scoped manager in `internal/observability/` owning trace/log/metric providers and an idempotent shutdown function. Initialize after configuration and before HTTP serving. Attach `service.name`, build-derived `service.version`, and `deployment.environment.name` to the resource; no mailbox identity or credentials.

Use batch processors with explicit bounded queues (initial defaults: trace queue 2048/batch 512, log queue 2048/batch 512, flush interval 5s), exporter timeout from config, and bounded retry behavior. Queue overflow may drop telemetry and must not block request handlers. Remote reachability is not a startup prerequisite: configuration errors fail startup, network errors degrade export and retain local logging. Exporter diagnostics go to a local-only handler with rate limiting to avoid recursive export and log floods.

Keep ownership in `StartServer`: release providers before returning any startup/serve error to Cobra, whose `os.Exit` does not run defers. The final Cobra fatal diagnostic remains local after shutdown.

On SIGINT/SIGTERM, finish the existing HTTP drain before flushing providers, using a fresh context and `shutdown_timeout`. Wait for the HTTP drain goroutine explicitly: a defer on `ListenAndServe` returning alone can flush too early. Clean up providers on startup errors and repeated shutdown calls. Use one shared telemetry timeout for all providers, not a fresh full timeout per signal. The existing 10s HTTP drain plus the default 10s telemetry budget totals at most 20s, excluding a small scheduling margin. During implementation set Docker Compose `stop_grace_period` and shipped systemd `TimeoutStopSec` to at least 30s; document the sum and operator adjustments when changing timeouts. On timeout report locally and exit without waiting indefinitely.

### 4. Echo v5 tracing

Implement project middleware against Echo v5's `*echo.Context`, with the Go SDK directly or a verified compatible instrumentation module. Do not assume the current `otelecho` version supports v5. When `trust_incoming_trace_context = true`, extract W3C `traceparent`/`tracestate`; otherwise ignore remote trace context and start a new root, preventing public clients from forcing sampling. Do not propagate baggage. Then create one SERVER span per request, and set the updated context on the request before handlers and request logging run.

Use method and matched route template for span names/attributes, including final status and duration. Use a fixed fallback for unmatched routes. The installed request middleware combines optional tracing with final request logging and runs outside recover → handlers. It replaces the legacy RequestLogger only when observability is enabled. For returned errors, leave the span open and finish it in a wrapper around Echo's configured HTTPErrorHandler, after delegating to the original handler exactly once; finish success spans when the chain returns. Whenever the global switch and at least one signal switch are enabled, install a final-response logging wrapper independently of the tracing middleware. The wrapper writes the final local context-aware request log in every enabled mode, including metrics-only and traces-only. The error wrapper writes it even in logs-only mode, and finalizes a span only if tracing is enabled. Omit the legacy RequestLogger in enabled mode to avoid duplicate/inaccurate early records. Cover recovered panics and use a single span-finalization guard. When the global switch is disabled, leave the existing local request-logging/error-handler behavior unchanged. Use parent-based trace-ID ratio sampling with `trace_sample_ratio` for root spans. When incoming context is trusted, preserve a valid remote parent's sampling decision. A zero root ratio does not override a sampled trusted parent. In logs-only mode do not install tracing/extraction middleware: logs have no generated or extracted span IDs. With traces enabled, nonrecording spans may still provide IDs in logs even if the trace is not exported; document that such IDs may have no matching stored trace.

### 5. slog bridge and data minimization

Fan out the existing local handler and an `otelslog` handler attached to the managed log provider. Retain the existing level filter and handler attributes/groups. Use context variants in request logging and relevant HTTP error logs so the bridge associates trace/span IDs; background records legitimately have no span. Logs remain independently enabled and are not discarded solely because a trace is unsampled.

Record HTTP method, route template, status, duration, and safe operational error classes. Exclude Authorization, cookies, JWTs, session secrets, passwords, full DSNs, bodies, query strings, mailbox addresses, and raw SQL. Use a remote attribute allowlist: service identity, trace/span IDs, HTTP method/route template/status/duration, and curated operation/error class. Omit client IP, User-Agent, arbitrary attributes/groups, email addresses and usernames. Preserve only allowed grouped attributes. Review existing slog calls: allow only audited static message templates; replace unknown/freeform messages with a fixed safe event name, and convert errors to curated classes without forwarding `err.Error()`. This is deterministic filtering rather than a claim that regex can redact arbitrary secret text. Retain local behavior and local handler groups/attributes. Do not use raw URL paths as span names, which expose mailbox identifiers. Do not copy `baggage` into records or attributes.

### 6. Database context and safe callbacks

Propagate HTTP context into every handler/repository call using a request-scoped GORM session and into API-key authentication. GET/HEAD use the original request context and preserve cancellation. Mutation handlers use context.WithoutCancel(request.Context()), preserving trace values while retaining the previous behavior that multi-step writes finish even if a client disconnects. API-key authentication retains cancellation because it runs before mutations. No new transaction/cancellation policy is imposed on existing writes. Test cancellation between mutation steps. Install server-owned GORM before/after callbacks for SELECT, INSERT, UPDATE, DELETE, ROW and EXEC. The final-response middleware is installed whenever any signal is enabled, including metrics-only mode. It marks the request context for database measurements and records the HTTP counter/histogram under the same exactly-once guard used for logs/spans. Only requests marked by this middleware are measured; database CLIENT spans require database_traces_enabled and a valid HTTP parent, inheriting its sampling decision. Restore the original statement context after each operation, including nested queries. Export fixed operation/system names and curated errors only; omit SQL text, parameters, tables, DSNs, identities and raw errors entirely. Treat ErrRecordNotFound as a normal result. ROW covers acquisition rather than result iteration. Transaction boundaries and startup migrations remain outside instrumentation.

### 7. Independent OTLP metrics

Use a server-owned SDK MeterProvider and periodic OTLP/HTTP protobuf exporter at metrics_endpoint, sharing explicit authenticated transport and shutdown budget. Override SDK temporality/aggregation selectors with cumulative counters and explicit histograms. metrics_export_interval defaults to 30s; export timeout and bounded retry follow the common settings. Set provider-level cardinality limit 2000 applied to each instrument using SDK WithCardinalityLimit. Explicit HTTP histogram boundaries in seconds: 0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10. Database boundaries: 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10. Configure them on histogram instruments and assert exported buckets. Record postfixadmin.http.requests and http.server.request.duration with bounded method, route template, status; record postfixadmin.db.operations and db.client.operation.duration with fixed system, operation and error class. Export db.client.connection.count observable gauge from database/sql pool statistics with used/idle state only. Metrics remain independent of traces, logs and sampling; unavailable exports cannot block request handlers. Tests must cover metrics-only mode, HTTP errors/panics exactly once, database correlation and failed queries, and no sensitive data in metric labels or exemplars. Validate the real OpenObserve metric streams in Docker.

## Risks / Trade-offs

- SDK and bridge packages can evolve independently: pin compatible releases and verify Go 1.26 and Echo v5 builds.
- An async queue can drop data during outages: availability takes priority, with local diagnostics and bounded memory.
- Sampling reduces trace volume: configure a higher ratio for troubleshooting; logs export independently.
- Existing records may contain sensitive values: review and test remote redaction before enabling export.
- Shutdown currently runs in a goroutine: explicitly coordinate request drain and provider flush to avoid losing final request records.

## Rollout and Validation

Ship disabled defaults and synchronized config templates. Enable against a test OpenObserve organization using credentials supplied through environment. Verify HTTP/database parent correlation, all metric streams and a correlated request log, service identity, configured stream, and custom URL paths. Test an unavailable endpoint, queue pressure, and signal-driven shutdown. Roll back by setting `observability.enabled = false`; no database rollback is needed.

Use local fake OTLP HTTP endpoints for automated assertions without credentials or external services. Run `go test ./...` and the required project build during implementation; Swagger regeneration is only needed if handler annotations change. Proposal validation is separate from runtime acceptance.
