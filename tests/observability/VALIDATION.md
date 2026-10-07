# Validation — OpenTelemetry / OpenObserve

Validated on 2026-10-07 against the isolated `postfixadmin-observability` Compose project. Shipped global/log/trace/metric/database tracing switches remain false; only the test stack enables them.

## Checks completed

- `go test ./...`: passed.
- `go vet ./...`: passed.
- `make build-prod`: passed, including rebuilt Vue SPA and UPX-compressed embedded binary.
- Docker Go 1.26: `go test -race -tags integration ./...` passed using the `test-runner` target.
- Strict OpenSpec validation: passed.
- Generated config checked inside the app container: all export and database tracing switches false and full endpoint examples present.
- Fake OTLP HTTP receiver tests verified protobuf payloads, exact custom paths, authorization/stream headers, log/span ID correlation, filtering/redaction, SDK env conflicts, sampling, errors/panics/404s, queue pressure, outage budgets, and HTTP drain ordering.

## Real OpenObserve ingestion

The smoke container created 2 fictional `.test` domains, 4 mailboxes and 2 aliases through the API. Authenticated list/detail requests and unauthorized/invalid-login responses passed. Initial smoke check found 23 logs, 20 spans and 20 correlated log records, with route templates and no fixture passwords/tokens/mailbox addresses in exported telemetry.

Fresh-window probes with 10 API requests per mode verified:

| Mode | Logs | Spans |
| --- | ---: | ---: |
| Global switch disabled | 0 | 0 |
| Logs only | 11 | 0 |
| Traces only | 0 | 11 |
| Both after OpenObserve restart | 12 | 12 |

Counts include the app health check and vary between runs. With OpenObserve stopped, 20 API requests still returned successfully, and local exporter diagnostics appeared at most once per 30 seconds. After restart, ingestion recovered. A second server process failed to bind the occupied port, released providers, and exited with status 1. SIGTERM flushed the final `Shutting down server…` record to OpenObserve before exit; the app was then restarted successfully.

## Browser and retained environment

agent-browser and Chromium ran in Docker. Verified login and the fake dashboard/domain/mailbox/alias data, and authenticated into OpenObserve to query the `postfixadmin` log stream and `default` trace stream. Screenshots are in ignored `tests/observability/runtime/`.

The app, MariaDB and OpenObserve remain running for manual inspection. URLs: app `http://localhost:18080`, OpenObserve `http://localhost:15080`. Credentials remain in the ignored, mode-0600 `.env.observability`; no credentials or runtime data are committed.

Limit: this validates isolated self-hosted OpenObserve and simulated HTTPS/configuration behavior. It does not validate a production OpenObserve Cloud account, production traffic load, mail-server log forwarding.

## Database and metrics follow-up

- Updated proposal approved by Claude CLI after review of mutation cancellation, metrics-only middleware, seconds-based histogram buckets and document consistency. Strict OpenSpec validation passed.
- Request-scoped GORM contexts validated with read cancellation, shared-session isolation and multi-step mutation continuation after client disconnect.
- Real OpenObserve smoke test: 22 logs, 42 spans (including 23 database CLIENT spans), 19 correlated logs. Database parent IDs matched stored HTTP spans. Counts vary with health checks and repeated runs.
- Real metric streams verified: `postfixadmin_http_requests`, `postfixadmin_db_operations`, `http_server_request_duration_count`, `db_client_operation_duration_count`, `db_client_connection_count`. Corresponding histogram bucket/sum streams are present. No fixture credentials, SQL text/parameters or mailbox addresses appeared in exported signals.
- Fake collector tests assert metric protobuf payloads, explicit seconds histogram bounds, disabled/independent modes, database errors/not-found handling and exclusion of operations outside HTTP. Metric outage/retry/shutdown budgets also passed.
- agent-browser displayed live HTTP/database metric charts; screenshot: ignored `runtime/openobserve-metrics.png`.

Fresh-window probes (10 version requests; metric row counts reflect aggregated series):

| Mode | Logs | HTTP spans | HTTP metric rows |
| --- | ---: | ---: | ---: |
| Globally disabled | 0 | 0 | 0 |
| Logs only | 11 | 0 | 0 |
| Traces only | 0 | 11 | 0 |
| Metrics only | 0 | 0 | 1 |
| All signals | 11 | 11 | 1 |

Database span detail and all five metric instruments were checked separately with authenticated fake-data requests. Production OpenObserve Cloud and PostgreSQL were not exercised. Database spans measure GORM callback operations; transaction boundaries and row iteration are outside scope. Mutation queries preserve the previous behavior of finishing after a client disconnect, while GET/HEAD queries now honor cancellation.

Final follow-up: all signals recovered after stopping/restarting OpenObserve; 20 requests remained available during the outage. SIGTERM flushed the shutdown log. Docker Go 1.26 race/integration tests, production build and strict proposal validation passed after the database/metric changes. The stack was restored with all signals enabled and fresh fake-data requests.
