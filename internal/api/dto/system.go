package dto

// ServiceStatus is the systemd state of one mail service unit.
type ServiceStatus struct {
	Name    string `json:"name" example:"postfix.service"`
	Active  string `json:"active" example:"active"`
	Sub     string `json:"sub" example:"running"`
	Enabled string `json:"enabled" example:"enabled"`
	Since   string `json:"since" example:"Sat 2026-11-07 20:15:02 -03"`
	Found   bool   `json:"found"`
}

// ServicesResponse lists the mail services of the panel's host. Available is false when systemd cannot be queried
// (e.g. inside a container), with the reason in Message.
type ServicesResponse struct {
	Available bool            `json:"available"`
	Message   string          `json:"message,omitempty"`
	Services  []ServiceStatus `json:"services"`
}

// TerminalInfoResponse describes the web SSH terminal (Settings > Terminal).
type TerminalInfoResponse struct {
	Enabled     bool   `json:"enabled"`
	Target      string `json:"target" example:"127.0.0.1:22"`
	HostKey     string `json:"host_key,omitempty" example:"ssh-ed25519 SHA256:abc..."`
	IdleMinutes int    `json:"idle_minutes" example:"30"`
}

// ConfigField is one config.toml key in Settings > Configuration. Value is omitted for secrets; Set tells whether a
// secret has a value. Env names the environment variable that overrides the file.
type ConfigField struct {
	Key      string   `json:"key" example:"smtp.port"`
	Kind     string   `json:"kind" example:"int"`
	Value    any      `json:"value,omitempty"`
	Editable bool     `json:"editable"`               // can be changed now
	Config   bool     `json:"configurable,omitempty"` // editable kind of key (may be locked by file permissions or env)
	Secret   bool     `json:"secret,omitempty"`
	Set      *bool    `json:"set,omitempty"`
	Restart  bool     `json:"restart,omitempty"`
	Env      string   `json:"env,omitempty"`
	Options  []string `json:"options,omitempty"`
	Min      int64    `json:"min,omitempty"`
	Max      int64    `json:"max,omitempty"`
}

// ConfigResponse lists the configuration shown in Settings; Writable is false when the panel cannot rewrite the file.
type ConfigResponse struct {
	File     string        `json:"file"`
	Writable bool          `json:"writable"`
	Fields   []ConfigField `json:"fields"`
}

// ConfigUpdateRequest maps "section.key" to its new value (bool, number or string).
type ConfigUpdateRequest struct {
	Changes map[string]any `json:"changes"`
}

// ConfigUpdateResponse lists the keys written; RestartRequired when one of them is read only at startup.
type ConfigUpdateResponse struct {
	Updated         []string `json:"updated"`
	RestartRequired bool     `json:"restart_required"`
}
