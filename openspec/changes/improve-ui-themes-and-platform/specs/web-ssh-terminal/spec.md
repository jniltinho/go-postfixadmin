## ADDED Requirements

### Requirement: Opt-in, superadmin-only web terminal
Settings › Terminal SHALL open an SSH shell on `terminal.host`:`terminal.port` in the browser only when `terminal.enabled` is true (default false). Only superadmins (with RBAC on, also holding `settings:write`) SHALL see the tab and be able to open a terminal; the server SHALL check this itself.

#### Scenario: Disabled by default
- **WHEN** `terminal.enabled` is not set
- **THEN** the WebSocket endpoint answers 404 and the tab explains how to enable it

#### Scenario: Not a superadmin
- **WHEN** a domain admin or a mailbox token opens the WebSocket
- **THEN** the server answers `forbidden` before dialing anything

#### Scenario: Cross-origin
- **WHEN** the WebSocket upgrade's `Origin` does not match the panel's `Host`
- **THEN** it is refused with 403

### Requirement: Credentials are not kept
The access token and the SSH password, key and passphrase SHALL travel only in the WebSocket's first message, never in the URL, and SHALL be dropped by the server after the SSH handshake and cleared from the form once the shell opens.

#### Scenario: Shell opened
- **WHEN** the shell is ready
- **THEN** the password/key fields are empty and nothing of the credential is stored in the database or logs

### Requirement: Host key pinned on first use
On the first connection the panel SHALL show the server's host key type and SHA-256 fingerprint and SHALL pin the key only after the operator confirms that exact fingerprint. Later connections SHALL accept only the pinned key; a different key SHALL be refused until a superadmin forgets the pinned one.

#### Scenario: First connection
- **WHEN** no key is pinned for the target
- **THEN** the operator sees `ssh-ed25519 SHA256:…` and nothing is sent to the server until they trust it

#### Scenario: Key changed
- **WHEN** the server presents a different host key than the pinned one
- **THEN** the connection is refused with "host key changed" and `terminal_open_failed … hostkey_changed` is logged

### Requirement: Real terminal behavior
The server SHALL request an `xterm-256color` PTY and the browser SHALL render it with xterm.js, resize the PTY with the panel and close the session after `terminal.idle_timeout` without keystrokes. Terminal colors SHALL come from the active theme's `--term-*` roles and follow theme changes live.

#### Scenario: 256 colors
- **WHEN** `echo $TERM; tput colors` runs in the terminal
- **THEN** it prints `xterm-256color` and `256`

#### Scenario: Theme change while connected
- **WHEN** the operator switches to the Matrix theme with a terminal open
- **THEN** the open terminal turns to the green phosphor palette without reconnecting

### Requirement: Audit and limits
Opening, failing to open, closing (with duration), trusting and forgetting a host key SHALL be written to the admin log. An admin SHALL have at most 4 open terminals.

#### Scenario: Session logged
- **WHEN** a superadmin opens and closes a terminal
- **THEN** the log has `terminal_open user@host:port` and `terminal_close user@host:port <duration>`
