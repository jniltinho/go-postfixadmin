## ADDED Requirements

### Requirement: Request tracing compatible with Echo v5
When traces are enabled, the system SHALL create one OpenTelemetry SERVER span per HTTP request using Echo v5-compatible middleware, propagate W3C traceparent/tracestate only when trust_incoming_trace_context is true (default false), and record method, matched route template, final response status, and duration. Root spans SHALL use parent-based trace-ID ratio sampling. When incoming context is trusted, valid remote parent sampling decisions SHALL be preserved. Otherwise the server SHALL create a new root.

#### Scenario: Incoming distributed trace
- **WHEN** an HTTP request carries a valid sampled traceparent and trust_incoming_trace_context is true
- **THEN** its server span belongs to the incoming trace and handlers receive the active span context

#### Scenario: Sampling root spans
- **WHEN** a request has no valid parent and trace_sample_ratio is zero
- **THEN** its root span is not exported

#### Scenario: Error response
- **WHEN** Echo handles a request error and writes its response
- **THEN** the span records the final HTTP status and appropriate error status without executing error handling twice

#### Scenario: Unknown route
- **WHEN** a request has no matched route
- **THEN** its span uses a bounded fallback name without copying the raw path


#### Scenario: Untrusted incoming context
- **WHEN** a client sends a sampled traceparent and trust_incoming_trace_context is false
- **THEN** the server ignores it and applies configured root sampling to a new trace


#### Scenario: Recovered panic
- **WHEN** a request panic is recovered by Echo middleware
- **THEN** exactly one span and one final request log record the resulting 500 response

### Requirement: Structured logs and correlation
When logs are enabled, the system SHALL export slog records through the OpenTelemetry log SDK and otelslog bridge while retaining local logging, configured levels and safe attributes/groups. Local logging SHALL preserve existing attributes/groups. Context-aware request records SHALL carry trace/span correlation. Logs SHALL remain independent of trace sampling. Both providers SHALL carry service.name, build service.version, and deployment.environment.name.

#### Scenario: Correlated request record
- **WHEN** a context-aware slog record is written during a traced request
- **THEN** its remote record includes that request's trace and span IDs and shared service identity

#### Scenario: Background or logs-only record
- **WHEN** a record has no active span
- **THEN** it is still exported without fabricated trace or span IDs

#### Scenario: Local debug filtering
- **WHEN** debug logging is disabled
- **THEN** debug records are filtered consistently in both local and remote outputs


#### Scenario: Logs-only correlation
- **WHEN** only logs are enabled
- **THEN** no incoming trace context is extracted and log records carry no generated or extracted span IDs


#### Scenario: Unsampled span correlation
- **WHEN** traces are enabled and a request span is not sampled
- **THEN** request logs may retain its trace/span IDs while that span is not exported

#### Scenario: Error log in logs-only mode
- **WHEN** only logs are enabled and Echo produces an error response
- **THEN** the independent final-response logging wrapper emits exactly one request log with the final status and no fabricated span IDs

### Requirement: Bounded asynchronous export
The system SHALL export OTLP/HTTP protobuf using bounded batch queues and retries with finite timeouts. Exporter outages and queue overflow SHALL NOT block request handlers or prevent local logging. Diagnostics SHALL be rate-limited and sent only to a local handler.

#### Scenario: OpenObserve unavailable
- **WHEN** the destination times out or returns an export failure
- **THEN** the API remains available, queues stay bounded, and a sanitized local diagnostic reports export degradation

#### Scenario: Queue overflow
- **WHEN** telemetry producers exceed the bounded queue capacity
- **THEN** telemetry may be dropped without blocking request processing

### Requirement: Coordinated telemetry shutdown
The server SHALL initialize providers before serving requests, clean them up on startup failure, and drain HTTP requests before flushing and shutting down providers with a fresh context bounded by shutdown_timeout. Shutdown SHALL be idempotent.

#### Scenario: Signal during active request
- **WHEN** SIGINT or SIGTERM arrives while a request is running
- **THEN** HTTP drain completes or reaches its existing timeout before provider flushing starts
- **AND** final request telemetry is eligible for export within the telemetry shutdown budget

#### Scenario: Export destination stalls at exit
- **WHEN** the provider cannot flush before shutdown_timeout
- **THEN** shutdown returns with a local diagnostic instead of waiting indefinitely

### Requirement: Sensitive data exclusion
Exported spans and logs SHALL exclude credentials, authorization headers, cookies, JWTs, session secrets, passwords, complete DSNs, request/response bodies, query strings, mailbox addresses, raw SQL, and propagated baggage. Remote log sanitization SHALL apply to existing records as well as new instrumentation, using an attribute allowlist and audited static messages. Unknown messages SHALL become a fixed safe event name; arbitrary error text SHALL become a curated class. Email addresses, usernames, client IP and User-Agent SHALL also be omitted.

#### Scenario: Request contains secrets and mailbox identity
- **WHEN** a request includes credentials and a mailbox address in its path or query
- **THEN** exported telemetry uses safe route templates and attributes and contains neither credentials nor the mailbox address

#### Scenario: Existing sensitive log attributes
- **WHEN** an existing slog record contains a sensitive field or message value
- **THEN** remote sanitization removes or redacts the sensitive value before export

### Requirement: Database tracing with request context
The system SHALL propagate request context through GORM sessions and API-key lookups. When database tracing is enabled it SHALL emit CLIENT spans for GORM operations under the HTTP span, preserving sampling without changing the shared session. Reads SHALL retain request cancellation; mutation handlers SHALL preserve trace values with context.WithoutCancel so existing multi-step writes can finish after client disconnect. Export SHALL omit SQL, parameters, table names, DSNs and raw errors.

#### Scenario: Successful query
- **WHEN** a traced HTTP request queries the database
- **THEN** the database span shares its trace ID and has the request span as parent
- **AND** attributes contain fixed operation and database system only

#### Scenario: Query failure
- **WHEN** a database operation returns an unexpected error
- **THEN** its span records database_error without exporting error text

#### Scenario: Not found
- **WHEN** a query returns ErrRecordNotFound
- **THEN** its span does not record an error status

#### Scenario: Database tracing disabled
- **WHEN** HTTP traces are enabled but database_traces_enabled is false
- **THEN** HTTP spans continue without database spans

#### Scenario: Client disconnect during mutation
- **WHEN** a client disconnects between database mutation steps
- **THEN** the mutation database context retains trace values and permits the remaining steps to finish

#### Scenario: Client disconnect during read
- **WHEN** a GET/HEAD client disconnects during a database read
- **THEN** the query receives the cancelled request context without changing the shared GORM context

### Requirement: Independently enabled application metrics
The system SHALL export OTLP/HTTP cumulative counters for HTTP requests and database operations, explicit histograms for their durations in seconds, and pool used/idle connection gauges when metrics_enabled and enabled are true. It SHALL use bounded attributes and limit cardinality to 2000 per instrument, collect independently of trace sampling, and flush under the shared shutdown budget.

#### Scenario: Metrics only
- **WHEN** metric export is enabled without logs or traces
- **THEN** counters, histograms and pool gauges are exported without creating trace spans

#### Scenario: Final HTTP response
- **WHEN** a request succeeds, returns an error or panics
- **THEN** exactly one request count and duration are recorded with final status and route template

#### Scenario: Destination unavailable
- **WHEN** metric export cannot reach the server
- **THEN** HTTP requests remain available and export retries/shutdown stay bounded
