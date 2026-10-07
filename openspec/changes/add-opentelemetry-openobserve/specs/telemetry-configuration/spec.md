## ADDED Requirements

### Requirement: Explicit opt-in per signal
The system SHALL expose `observability.enabled`, `observability.traces_enabled`, `observability.logs_enabled`, `observability.metrics_enabled`, and `observability.database_traces_enabled`, all defaulting to false. The global switch SHALL gate all signals. Settings SHALL take effect on server restart.

#### Scenario: Default configuration
- **WHEN** the server starts without observability settings
- **THEN** it creates no telemetry exporters and makes no telemetry export requests
- **AND** local logging remains available

#### Scenario: Global switch disabled
- **WHEN** enabled is false and any signal switch is true
- **THEN** no traces, logs or metrics are exported

#### Scenario: Only logs enabled
- **WHEN** enabled and logs_enabled are true and traces_enabled and metrics_enabled are false
- **THEN** only the log exporter is initialized and traces_endpoint is not required

#### Scenario: Only traces enabled
- **WHEN** enabled and traces_enabled are true and logs_enabled and metrics_enabled are false
- **THEN** only the trace exporter is initialized and logs_endpoint is not required

#### Scenario: No signal selected
- **WHEN** enabled is true and all signal switches are false
- **THEN** startup fails with a configuration error identifying the missing signal selection

### Requirement: Configurable destinations and identity
The system SHALL configure full traces_endpoint, logs_endpoint and metrics_endpoint URLs, authorization, logs_stream, service_name, environment, trace_sample_ratio, metrics_export_interval, export_timeout, shutdown_timeout, trust_incoming_trace_context, and allow_insecure_http under observability. Defaults SHALL match design.md. Each setting SHALL bind explicitly to its OBSERVABILITY_* environment variable, with environment overriding TOML and TOML overriding defaults. Both shipped TOML templates and generated config SHALL expose matching settings with disabled export defaults.

#### Scenario: Environment override
- **WHEN** OBSERVABILITY_LOGS_ENDPOINT and observability.logs_endpoint are both supplied
- **THEN** the environment URL is used for log export after restart

#### Scenario: Organization or reverse proxy path
- **WHEN** an enabled signal has a full endpoint URL with a custom path
- **THEN** its exporter posts to that exact path without appending a second signal suffix

#### Scenario: Authentication and stream
- **WHEN** logs are exported to the configured destination
- **THEN** Authorization contains the configured value and stream-name contains logs_stream
- **AND** traces use the configured Authorization value at traces_endpoint

#### Scenario: Generated configuration
- **WHEN** an operator invokes --generate-config after implementation
- **THEN** the emitted observability section matches the example template and every enablement switch is false

### Requirement: Safe configuration validation
The system SHALL reject enabled configuration with a missing enabled-signal endpoint, invalid absolute HTTP(S) URL or missing/non-root ingestion path, URL credentials/query/fragment, empty service name or authorization, invalid authorization header characters, empty stream for enabled logs, a nonfinite or out-of-range trace sampling ratio, or nonpositive/invalid timeouts. HTTPS SHALL verify certificates. Plain HTTP SHALL require allow_insecure_http=true. Errors SHALL identify keys without printing secrets.

#### Scenario: Missing logs destination
- **WHEN** global and logs switches are enabled and logs_endpoint is empty
- **THEN** startup fails with an error naming observability.logs_endpoint

#### Scenario: HTTP requires opt-in
- **WHEN** an enabled endpoint uses http and allow_insecure_http is false
- **THEN** startup rejects the endpoint

#### Scenario: Disabled invalid exporter settings
- **WHEN** enabled is false and exporter settings are missing or invalid
- **THEN** exporter validation is skipped and local operation proceeds

#### Scenario: Secret in invalid URL
- **WHEN** an enabled endpoint includes userinfo
- **THEN** startup rejects the configuration without exposing that userinfo in diagnostics

#### Scenario: Invalid endpoint on disabled signal
- **WHEN** global and logs switches are enabled, valid log settings are supplied, traces_enabled is false, and traces_endpoint is invalid
- **THEN** log export starts successfully and the disabled trace endpoint is ignored

#### Scenario: Conflicting SDK environment
- **WHEN** enabled export encounters a standard OTEL_* environment setting that cannot be overridden to honor the project configuration
- **THEN** startup rejects that setting by key without exposing its value

#### Scenario: Missing ingestion path
- **WHEN** an enabled signal endpoint has an empty or root-only path
- **THEN** startup fails with a configuration error for that endpoint key

#### Scenario: Malformed environment value
- **WHEN** an enabled configuration supplies a malformed boolean, ratio or duration through OBSERVABILITY_* variables
- **THEN** startup rejects the setting by key rather than silently converting it to false or zero

#### Scenario: Metrics only
- **WHEN** global and metrics switches are true and log/trace/database trace switches are false
- **THEN** only the metric exporter initializes and disabled signal endpoints are ignored

#### Scenario: Database traces require tracing
- **WHEN** enabled and database_traces_enabled are true but traces_enabled is false
- **THEN** startup rejects database_traces_enabled by key

#### Scenario: Invalid metric interval
- **WHEN** metric export is enabled with a nonpositive metrics_export_interval
- **THEN** startup rejects metrics_export_interval by key
