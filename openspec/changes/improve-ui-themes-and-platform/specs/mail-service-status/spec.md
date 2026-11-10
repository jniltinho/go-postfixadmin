## ADDED Requirements

### Requirement: Read-only mail service status
Settings › Services SHALL show, for `postfix.service`, `dovecot.service` and `rspamd.service`, the systemd active state, sub-state, boot (unit file) state and the time it entered the active state, read with `systemctl show`. The unit list SHALL be a fixed allowlist, never taken from the request. The API (`GET /api/v1/system/services`) SHALL require the `settings:read` permission when RBAC is enabled.

#### Scenario: Services running
- **WHEN** the units are active on the panel's host
- **THEN** each row shows `active/running` (or `active/exited`) with a check icon, `enabled` and its start time

#### Scenario: Unit not installed
- **WHEN** `systemctl show` reports `LoadState=not-found` for a unit
- **THEN** that row shows "not installed" instead of an error

#### Scenario: No systemd
- **WHEN** `systemctl` is not available (e.g. the panel runs in a container)
- **THEN** the tab says the status is not available on this host and the API still answers 200 with `available: false`

#### Scenario: Permission
- **WHEN** RBAC is enabled and the admin lacks `settings:read`
- **THEN** `GET /api/v1/system/services` answers 403
