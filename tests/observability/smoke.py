import json
import os
import time
import urllib.error
import urllib.request

app = os.environ["APP_URL"]
observe = os.environ["OBSERVE_URL"]
admin = os.environ["TEST_ADMIN_EMAIL"]
password = os.environ["TEST_ADMIN_PASSWORD"]
auth = os.environ["TEST_OBSERVE_AUTHORIZATION"]
start_us = int(time.time() * 1_000_000) - 10_000_000


def request(base, path, data=None, authorization=None, accepted=(200, 201)):
    headers = {"Content-Type": "application/json"}
    if authorization:
        headers["Authorization"] = authorization
    req = urllib.request.Request(base + path, None if data is None else json.dumps(data).encode(), headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            raw = response.read()
            assert response.status in accepted
    except urllib.error.HTTPError as err:
        raw = err.read()
        if err.code not in accepted:
            raise AssertionError(f"{path}: unexpected HTTP {err.code}") from None
    return json.loads(raw) if raw else {}


for attempt in range(60):
    try:
        request(observe, "/healthz")
        break
    except (OSError, ValueError, AssertionError):
        time.sleep(1)
else:
    raise AssertionError("OpenObserve did not become ready")

login = request(app, "/api/v1/auth/login", {"username": admin, "password": password})
token = "Bearer " + login["data"]["access_token"]
for domain, description in [("example.test", "Fictional Example Company"), ("demo.test", "Fictional Demo Organization")]:
    request(app, "/api/v1/domains", {"domain": domain, "description": description, "aliases": 50, "mailboxes": 50, "quota": 1073741824, "active": True}, token, (201, 409))
    for local, name in [("alice", "Alice Example"), ("bob", "Bob Example")]:
        request(app, "/api/v1/mailboxes", {"local_part": local, "domain": domain, "name": name, "password": password, "quota": 104857600, "active": True, "smtp_active": True, "send_welcome": False}, token, (201, 409))
    request(app, "/api/v1/aliases", {"local_part": "support", "domain": domain, "goto": "alice@" + domain, "active": True}, token, (201, 409))

for resource in ["domains", "mailboxes", "aliases", "dashboard"]:
    result = request(app, "/api/v1/" + resource, authorization=token)
    assert result["success"], resource
request(app, "/api/v1/mailboxes/alice@example.test?secret=do-not-export", authorization=token)
request(app, "/api/v1/mailboxes", accepted=(401,))
request(app, "/api/v1/auth/login", {"username": "unknown@example.test", "password": "NeverExportThis!"}, accepted=(401,))
print("Seeded 2 fictional domains, 4 mailboxes, 2 aliases; authenticated and error API requests passed")


def search(stream, signal="logs"):
    return request(observe, "/api/default/_search?type=" + signal, {"query": {"sql": 'SELECT * FROM "' + stream + '"', "start_time": start_us, "end_time": int(time.time() * 1_000_000), "from": 0, "size": 1000}}, auth).get("hits", [])


for attempt in range(45):
    try:
        logs = search("postfixadmin")
        traces = search("default", "traces")
        db_spans = [row for row in traces if str(row.get("operation_name", row.get("name", ""))).startswith("DB ")]
        if logs and db_spans and any("/api/v1/mailboxes/:username" in json.dumps(row) for row in logs):
            break
    except (AssertionError, urllib.error.URLError):
        pass
    time.sleep(2)
else:
    raise AssertionError("No application logs/traces found in OpenObserve")

for row in logs + traces:
    serialized = json.dumps(row)
    for secret in [password, token, auth, "alice@example.test", "unknown@example.test", "NeverExportThis!", "do-not-export"]:
        assert secret not in serialized, "Sensitive data appeared in exported telemetry"
trace_ids = {row.get("trace_id") for row in traces}
correlated = [row for row in logs if row.get("trace_id") in trace_ids and row.get("trace_id")]
assert correlated, "No correlated trace/log IDs found"
assert any("/api/v1/mailboxes/:username" in json.dumps(row) for row in logs), "Route template not found"
print(f"OpenObserve ingestion verified: {len(logs)} logs, {len(traces)} spans, {len(correlated)} correlated logs")

http_ids = {(row.get("trace_id"), row.get("span_id")) for row in traces if not str(row.get("operation_name", row.get("name", ""))).startswith("DB ")}
assert any((row.get("trace_id"), row.get("reference_parent_span_id", row.get("parent_span_id"))) in http_ids for row in db_spans), "Database spans are not children of HTTP spans"
metric_streams = ["postfixadmin_http_requests", "postfixadmin_db_operations", "http_server_request_duration_count", "db_client_operation_duration_count", "db_client_connection_count"]
for attempt in range(30):
    try:
        metric_rows = {stream: search(stream, "metrics") for stream in metric_streams}
        if all(metric_rows.values()):
            break
    except (AssertionError, urllib.error.URLError):
        pass
    time.sleep(2)
else:
    raise AssertionError("Expected metric streams not found in OpenObserve")
for stream, rows in metric_rows.items():
    for row in rows:
        serialized = json.dumps(row)
        for secret in [password, token, auth, "alice@example.test", "NeverExportThis!", "do-not-export"]:
            assert secret not in serialized, "Sensitive data appeared in exported metrics"
print(f"Database parent correlation verified: {len(db_spans)} CLIENT spans; metrics: " + ", ".join(f"{stream}={len(rows)}" for stream, rows in metric_rows.items()))
