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
