## ADDED Requirements

### Requirement: Configuration tab access
Settings › Configuration SHALL be available to superadmins only; with RBAC on they also need `settings:read` to view and `settings:write` to save.

#### Scenario: Not a superadmin
- **WHEN** a domain admin calls `GET /api/v1/system/config`
- **THEN** the API answers 403

### Requirement: Editable operational keys
Only `ui.default_theme`, `vacation.enabled`, `smtp.server`, `smtp.port`, `smtp.type`, `server.swagger_enable` (after a restart), `terminal.enabled` and `terminal.idle_timeout` SHALL be editable from the panel. Enabling the terminal SHALL ask for confirmation; its host and port stay read-only.

#### Scenario: Restart needed
- **WHEN** `server.swagger_enable` is saved
- **THEN** the answer says a restart is required

### Requirement: Read-only and secret keys
All other shown keys (database, server security, quota multiplier, RBAC, transport, terminal host/port, backup, observability) SHALL be read-only, and secret values SHALL never be returned.

#### Scenario: Secrets hidden
- **WHEN** a superadmin opens the tab
- **THEN** `database.pass` and the other secrets show "set (hidden)" or "not set" and their values are not in the API response

#### Scenario: Read-only key refused
- **WHEN** a request tries to change `server.port`
- **THEN** the API answers 400 and the file is unchanged

### Requirement: Safe writes that keep the file readable
Saving SHALL validate each value, change only the affected lines (comments and spacing kept), verify that the new file parses with the new values, keep `config.toml.bak`, replace the file atomically with its permissions, apply the values in memory and log `config_update`.

#### Scenario: Locked keys
- **WHEN** an environment variable overrides a key or the panel cannot write config.toml
- **THEN** the key is shown but cannot be changed

#### Scenario: Change the SMTP port
- **WHEN** the superadmin sets `smtp.port` to 587 and saves
- **THEN** only that line changes in config.toml, `config.toml.bak` holds the previous file, new welcome e-mails use port 587 and the log has `config_update smtp.port`

#### Scenario: Invalid value
- **WHEN** `smtp.port` is 70000 or `smtp.server` contains a line break
- **THEN** the API answers 400 and nothing is written
