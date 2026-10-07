import json
import os
import sys
import time
import urllib.error
import urllib.request

app = os.environ["APP_URL"]
observe = os.environ["OBSERVE_URL"]
auth = os.environ["TEST_OBSERVE_AUTHORIZATION"]
mode = sys.argv[1]


def call(base, path, data=None, authorization=None):
    headers = {"Content-Type": "application/json"}
    if authorization:
        headers["Authorization"] = authorization
    request = urllib.request.Request(base + path, None if data is None else json.dumps(data).encode(), headers)
    with urllib.request.urlopen(request, timeout=5) as response:
        return json.loads(response.read())


if mode == "shutdown":
    for attempt in range(10):
        data = {"query": {"sql": 'SELECT * FROM "postfixadmin"', "start_time": int(time.time() * 1_000_000) - 30_000_000, "end_time": int(time.time() * 1_000_000), "from": 0, "size": 1000}}
        rows = call(observe, "/api/default/_search?type=logs", data, auth).get("hits", [])
        if any(row.get("body") == "Shutting down server…" for row in rows):
            print("SIGTERM final shutdown log was flushed to OpenObserve")
            raise SystemExit(0)
        time.sleep(1)
    raise AssertionError("Final shutdown record was not ingested")


if mode == "availability":
    for _ in range(20):
        assert "version" in call(app, "/api/v1/version")
    print("API stayed available during the OpenObserve outage (20 requests)")
    raise SystemExit(0)

start = int(time.time() * 1_000_000)
for _ in range(10):
    assert "version" in call(app, "/api/v1/version")
time.sleep(7)
end = int(time.time() * 1_000_000)


def count(stream, signal):
    data = {"query": {"sql": 'SELECT * FROM "' + stream + '"', "start_time": start, "end_time": end, "from": 0, "size": 1000}}
    try:
        return len(call(observe, "/api/default/_search?type=" + signal, data, auth).get("hits", []))
    except urllib.error.HTTPError as err:
        if err.code == 400:
            return 0
        raise


metrics = count("postfixadmin_http_requests", "metrics")
logs, traces = count("postfixadmin", "logs"), count("default", "traces")
expected = {"disabled": (False, False, False), "logs-only": (True, False, False), "traces-only": (False, True, False), "metrics-only": (False, False, True), "all": (True, True, True)}[mode]
assert (logs > 0, traces > 0, metrics > 0) == expected, f"Unexpected signal counts in {mode}: logs={logs}, traces={traces}, metrics={metrics}"
print(f"{mode} verified in OpenObserve: logs={logs}, traces={traces}, metrics={metrics}")
