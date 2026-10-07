# Docker observability validation

This isolated Compose project runs Go-PostfixAdmin, MariaDB and a pinned OpenObserve image. Application telemetry is enabled only in this test stack; shipped application defaults remain disabled. Test data uses reserved `.test` domains and sends no welcome email.

## Start and populate

From the repository root:

```bash
python3 tests/observability/setup.py
docker compose --env-file .env.observability -f docker-compose.observability.yml up -d --build db openobserve app
docker compose --env-file .env.observability -f docker-compose.observability.yml --profile test run --rm smoke
```

The setup script generates random credentials in `.env.observability` and a runtime config with a random session secret under `tests/observability/runtime/`. Both paths are ignored by Git, and credential files use mode 0600. Re-running setup preserves existing credentials and data.

The smoke container seeds 2 domains, 4 mailboxes and 2 aliases through the API, exercises authenticated and rejected requests, and queries OpenObserve to verify ingestion, route templates, trace/log ID correlation, and exclusion of test credentials and mailbox addresses. It can be re-run: conflicts for existing fixture records are accepted.

Access:

- Application: <http://localhost:18080>; use TEST_ADMIN_EMAIL and TEST_ADMIN_PASSWORD from the generated environment file.
- OpenObserve: <http://localhost:15080>; use TEST_OBSERVE_EMAIL and TEST_OBSERVE_PASSWORD from that file. Select organization `default`, Logs stream `postfixadmin`, or Traces stream `default`.

Ports bind to loopback only. Override APP_PORT and OBSERVE_PORT in the environment file if needed. Containers use HTTP inside the private Docker network, with explicit test-only allow_insecure_http enablement.

## Go tests in Docker

```bash
docker compose --env-file .env.observability -f docker-compose.observability.yml --profile test build go-tests
docker compose --env-file .env.observability -f docker-compose.observability.yml --profile test run --rm --no-deps go-tests
```

This uses Go 1.26, rebuilds the embedded SPA, and runs `go test -race -tags integration ./...`. Local fake OTLP receivers check custom paths, protobuf payloads, headers, sampling, final error/panic statuses, safe log export, outage/queue behavior, and bounded shutdown. No external account is needed.

## Browser in Docker

```bash
docker compose --env-file .env.observability -f docker-compose.observability.yml --profile browser up -d --build browser
docker compose --env-file .env.observability -f docker-compose.observability.yml exec browser agent-browser open http://postfixadmin.test:8080
docker compose --env-file .env.observability -f docker-compose.observability.yml exec browser agent-browser snapshot -i
```

The browser container installs agent-browser and Chromium. Use snapshot refs to log in with its TEST_ADMIN_EMAIL/TEST_ADMIN_PASSWORD environment variables, then visit dashboard, domains, mailboxes and aliases. The `postfixadmin.test` network alias avoids Chromium's HTTPS upgrade for the `.app` hostname. To inspect OpenObserve, navigate to `http://openobserve:5080` and use TEST_OBSERVE_EMAIL/TEST_OBSERVE_PASSWORD. Browser commands use the named session `postfixadmin-observability`; screenshots under `/artifacts` appear in ignored `tests/observability/runtime/`.

To save a screenshot and close the browser:

```bash
docker compose --env-file .env.observability -f docker-compose.observability.yml exec browser agent-browser screenshot /artifacts/app.png
docker compose --env-file .env.observability -f docker-compose.observability.yml exec browser agent-browser close
```

## Switches, outage, and shutdown

The fixture stack accepts TEST_TELEMETRY_ENABLED, TEST_TRACES_ENABLED, TEST_LOGS_ENABLED, TEST_METRICS_ENABLED and TEST_DB_TRACES_ENABLED overrides. The smoke container also provides `python /tests/probe.py disabled`, `logs-only`, `traces-only`, `metrics-only`, `all`, and `availability` probes, invoked with `run --rm --no-deps smoke`; signal probes assert ingestion over a fresh time window after sending requests. Recreate the app to change its environment; a restart alone does not update container environment.

```bash
TEST_TELEMETRY_ENABLED=false docker compose --env-file .env.observability -f docker-compose.observability.yml up -d --no-build app
# Restore all signals:
docker compose --env-file .env.observability -f docker-compose.observability.yml up -d --no-build app
```

For an outage check, stop only openobserve, send requests to the app, and confirm HTTP availability and rate-limited local export diagnostics. Restart openobserve and allow a batch interval for ingestion to recover. `docker compose ... stop app` exercises SIGTERM with 30 seconds of grace; compare final request telemetry in OpenObserve before bringing app back up.

Stop or resume without deleting data:

```bash
docker compose --env-file .env.observability -f docker-compose.observability.yml stop
docker compose --env-file .env.observability -f docker-compose.observability.yml up -d app
```

`down` removes this project's containers/network while retaining data volumes. Add `--volumes` only when intentionally discarding all fake database and OpenObserve records. To rotate fixture credentials, discard this stack's data and regenerate the ignored runtime files together.

To run metrics only, recreate app with `TEST_TRACES_ENABLED=false TEST_DB_TRACES_ENABLED=false TEST_LOGS_ENABLED=false TEST_METRICS_ENABLED=true`, then run the `metrics-only` probe. For logs-only or traces-only, disable metric export as well. Database tracing requires traces enabled. The default fixture enables every signal; shipped configuration defaults remain disabled. Metric streams use normalized names (`postfixadmin_http_requests`, `postfixadmin_db_operations`, `http_server_request_duration_count`, `db_client_operation_duration_count`, `db_client_connection_count`).
