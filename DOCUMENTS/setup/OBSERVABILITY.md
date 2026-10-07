# OpenTelemetry and OpenObserve

Go-PostfixAdmin optionally exports application logs, HTTP/database traces, and metrics to OpenObserve using OTLP/HTTP protobuf. Export is disabled by default. Local logs remain available. This does not forward Postfix/Dovecot log files, instrument the transport TCP server and CLI commands.

## Enable export

Copy the `[observability]` section from `config.toml.example` (also included by `--generate-config`) and set the global switch plus the switches for the desired signals:

```toml
[observability]
enabled = true
traces_enabled = true
logs_enabled = true
metrics_enabled = true
database_traces_enabled = true
metrics_export_interval = "30s"
service_name = "go-postfixadmin"
environment = "production"
traces_endpoint = "https://observe.example.com/api/default/v1/traces"
logs_endpoint = "https://observe.example.com/api/default/v1/logs"
metrics_endpoint = "https://observe.example.com/api/default/v1/metrics"
logs_stream = "postfixadmin"
trace_sample_ratio = 0.1
```

Supply `OBSERVABILITY_AUTHORIZATION` through your deployment's secret/environment mechanism. Use the authorization value from OpenObserve ingestion settings, typically `Basic <base64(email:password)>`. Restart the server after changes. URLs must include the complete ingestion path; no suffix is appended. The organization is part of the URL. Signal destinations may use different hosts, but share this authorization value.

Generate the header as follows. Encode `email:password` without a trailing newline; keep the `Basic ` prefix outside the encoded value:

```bash
# Fictional credentials: replace with your OpenObserve ingestion credentials.
export OBSERVABILITY_AUTHORIZATION="Basic $(printf '%s' 'telemetry@example.com:example-password' | base64 | tr -d '\r\n')"
```

Start the server from the same shell after configuring the enabled signals and endpoints:

```bash
./bin/postfixadmin server
```

For Docker Compose, pass the variable into the application service (exporting it on the host alone does not inject it into a container):

```yaml
services:
  app:
    environment:
      OBSERVABILITY_AUTHORIZATION: ${OBSERVABILITY_AUTHORIZATION:?Set OpenObserve authorization}
```

Recreate the application container with `docker compose up -d app`. If using an environment file, store the complete `Basic <encoded-value>` string, restrict the file to mode `0600`, and keep it out of Git. Base64 is an encoding, not encryption; use HTTPS outside the isolated test network. The Docker test fixture generates its own header automatically.

For logs-only mode, disable traces, metrics and database tracing. For traces-only mode, disable logs and metrics; database tracing remains optional and requires traces. For metrics-only mode, disable logs, traces and database tracing. Disabled signal endpoints are ignored. Setting `enabled = false` stops all export on restart. Setting the global switch to true while all signals are false is a configuration error.

## Configuration reference

Environment variables override TOML, which overrides defaults. There are no additional CLI flags or hot reload. All variables are named `OBSERVABILITY_` followed by the uppercase key below; for example `OBSERVABILITY_LOGS_ENDPOINT`, `OBSERVABILITY_TRACES_ENDPOINT`, and `OBSERVABILITY_AUTHORIZATION`.

| TOML key under `observability` | Default | Meaning |
| --- | --- | --- |
| `enabled` | `false` | Global export switch |
| `traces_enabled` | `false` | Trace export switch; also requires global enablement |
| `metrics_enabled` | `false` | Metric export switch; also requires global enablement |
| `database_traces_enabled` | `false` | Database CLIENT spans; requires global and trace enablement |
| `metrics_endpoint` | `""` | Full OTLP/HTTP metric ingestion URL |
| `metrics_export_interval` | `"30s"` | Positive periodic metric export interval |
| `logs_enabled` | `false` | Log export switch; also requires global enablement |
| `trust_incoming_trace_context` | `false` | Accept incoming W3C traceparent/tracestate and parent sampling decisions |
| `service_name` | `"go-postfixadmin"` | Exported service.name |
| `environment` | `"production"` | Exported deployment.environment.name |
| `traces_endpoint` | `""` | Full OTLP/HTTP trace ingestion URL |
| `logs_endpoint` | `""` | Full OTLP/HTTP log ingestion URL |
| `authorization` | `""` | Authorization header; required for enabled export |
| `logs_stream` | `"postfixadmin"` | OpenObserve stream-name header for logs |
| `trace_sample_ratio` | `0.1` | Root trace sampling ratio between 0 and 1 |
| `export_timeout` | `"5s"` | Export request/batch timeout and maximum retry duration |
| `shutdown_timeout` | `"10s"` | Shared provider shutdown/flush budget |
| `allow_insecure_http` | `false` | Explicitly allow unencrypted HTTP; useful inside the isolated test network |

Service version comes from the binary's build version. Trace and log signals use a queue of 2048 records/spans, batches of at most 512, and a 5-second export interval. Retries and queues are bounded; queue pressure or prolonged outages can drop telemetry without blocking requests.

HTTPS verifies certificates using the system trust store. HTTP requires explicit opt-in and prints a local warning. URL credentials, query strings, fragments, and root-only paths are rejected. Authorization and stream values cannot contain control characters. When storing credentials in TOML, restrict the file to mode `0600`; environment credentials are preferred.

Use the project variables rather than SDK environment variables. The pinned SDK always merges resource environment attributes and eagerly reads trace TLS certificate settings. Enabled startup therefore rejects nonempty `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, and `OTEL_EXPORTER_OTLP_{,TRACES_,LOGS_,METRICS_}{CERTIFICATE,CLIENT_CERTIFICATE,CLIENT_KEY}` by variable name without printing values. Other SDK endpoint/header/sampler/batch settings are overridden by explicit project options. Disabled telemetry ignores these SDK settings.

## Logs, tracing, and shutdown

Request telemetry records method, route template, final status, and duration. It omits raw paths, query strings, bodies, cookies, authorization, passwords, mailbox/user identities, SQL, client IP, and User-Agent. Remote application records use audited static messages and an attribute allowlist; unknown messages become `Application event`, and arbitrary fields are omitted. Local log contents and debug filtering remain unchanged.

Context-aware request logs carry trace_id/span_id when tracing is enabled. A request's trace may be unsampled even though its log has IDs; those IDs may have no stored trace. Logs-only mode generates no trace IDs. Incoming remote context is ignored by default to prevent public clients from forcing sampling; enable trust only behind a trusted tracing boundary. A sampled trusted parent is preserved even when the root ratio is zero. Baggage is not collected.

SIGINT/SIGTERM first drains HTTP requests for up to 10 seconds, then flushes providers under the shared telemetry budget. The default combined budget is 20 seconds; allow at least 30 seconds with Docker Compose `stop_grace_period` or systemd `TimeoutStopSec`. Increase deployment grace periods if you increase shutdown_timeout. Fatal server startup errors release providers before Cobra exits.

## Verify and troubleshoot

In OpenObserve, select the organization configured in your URL, open Logs, and select the configured log stream. Search for `service_name = 'go-postfixadmin'`, then inspect a request's trace_id in Traces (default trace stream `default`). Trigger API requests and allow the batch interval to elapse. Confirm route templates such as `/api/v1/mailboxes/:username` appear instead of real mailbox addresses.

A local `Telemetry export degraded` diagnostic is rate-limited and never exported. Check connectivity, the full ingestion path, authorization, organization, certificate trust, and signal switches. A reachable UI alone does not establish ingestion permission. To roll back, set `observability.enabled = false` and restart; no database migration is needed.

## Isolated Docker validation

The [Docker test guide](../../tests/observability/README.md) runs the app, MariaDB, OpenObserve, fake-data/API ingestion checks, Go race tests, and agent-browser in containers. It uses separate Compose resources and generated test credentials, with browser screenshots outside version control.

References: [OpenObserve Go tracing](https://openobserve.ai/docs/ingestion/traces/go/), [OTLP logs](https://openobserve.ai/docs/ingestion/logs/otlp/), [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go).

## Database traces and metrics

`database_traces_enabled = true` adds CLIENT spans for GORM SELECT, INSERT, UPDATE, DELETE, ROW and EXEC callbacks below each HTTP span. Handlers and API-key lookups use request-scoped GORM sessions, preserving parent IDs without mutating the shared database. GET/HEAD preserve request cancellation; mutation handlers retain trace values with `context.WithoutCancel` so a client disconnect does not interrupt existing multi-step writes. API-key lookups retain cancellation before writes begin. Expected not-found results do not set error status. Database failures use the fixed `database_error` class. SQL text, parameters, table names, DSNs and raw errors are never exported. ROW spans measure query acquisition; they do not cover subsequent iteration. Transaction begin/commit/rollback and startup migrations are outside this instrumentation.

Metric export is independent of trace sampling and supports metrics-only mode. The SDK limits each instrument to 2000 attribute combinations. It sends cumulative counters and explicit-bucket latency histograms every `metrics_export_interval`, and flushes on shutdown. Measurements are recorded in memory; HTTP requests do not wait for remote export.

| Instrument | Type | Attributes |
| --- | --- | --- |
| `postfixadmin.http.requests` | Counter | Method, route template, final status |
| `http.server.request.duration` | Histogram, seconds | Method, route template, final status |
| `postfixadmin.db.operations` | Counter | Database system, fixed operation, error class |
| `db.client.operation.duration` | Histogram, seconds | Database system, fixed operation, error class |
| `db.client.connection.count` | Gauge | db.client.connection.state: used or idle |

OpenObserve converts metric names to Prometheus-style stream names, for example `postfixadmin_http_requests` and histogram streams ending in `_count`, `_sum`, `_bucket`. Open Metrics and query these streams; use `rate(postfixadmin_http_requests[5m])` for traffic and `histogram_quantile(0.95, sum by (le) (rate(http_server_request_duration_bucket[5m])))` for aggregate P95 latency. These measurements describe the application, not host CPU/memory or the mail-server processes.
