#!/usr/bin/env python3
import base64
from pathlib import Path
import secrets

root = Path(__file__).resolve().parents[2]
env_path = root / ".env.observability"
if env_path.exists():
    print("Existing isolated test credentials preserved in .env.observability")
    raise SystemExit(0)
values = {
    "TEST_DB_PASSWORD": secrets.token_hex(16),
    "TEST_ADMIN_EMAIL": "admin@example.test",
    "TEST_ADMIN_PASSWORD": "Test!" + secrets.token_hex(12),
    "TEST_OBSERVE_EMAIL": "telemetry@example.test",
    "TEST_OBSERVE_PASSWORD": "Observe!" + secrets.token_hex(12),
    "APP_PORT": "18080",
    "OBSERVE_PORT": "15080",
}
auth = base64.b64encode((values["TEST_OBSERVE_EMAIL"] + ":" + values["TEST_OBSERVE_PASSWORD"]).encode()).decode()
values["TEST_OBSERVE_AUTHORIZATION"] = "Basic " + auth
runtime = root / "tests/observability/runtime"
runtime.mkdir(parents=True, exist_ok=True)
config = (root / "config.toml.example").read_text()
config = config.replace('session_secret = "9a048f79e88e35de37dc2c43c1fc002f358f92957a7690e60109cfe8a65178e0"', 'session_secret = "' + secrets.token_hex(32) + '"')
(runtime / "config.toml").write_text(config)
(runtime / "config.toml").chmod(0o600)
env_path.write_text("\n".join(f"{key}='{value}'" for key, value in values.items()) + "\n")
env_path.chmod(0o600)
print("Generated isolated test credentials in .env.observability (0600)")
