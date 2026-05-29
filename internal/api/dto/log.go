package dto

import "time"

// LogEntryResponse for admin action logs
type LogEntryResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Username  string    `json:"username"`
	Domain    string    `json:"domain"`
	Action    string    `json:"action"`
	Data      string    `json:"data"`
}

// MailLogEntryResponse for maillog entries
type MailLogEntryResponse struct {
	LogDate    string `json:"logdate"`
	MFrom      string `json:"m_from"`
	MTo        string `json:"m_to"`
	DomainFrom string `json:"domain_from"`
	DomainTo   string `json:"domain_to"`
	HostIP     string `json:"host_ip"`
	HostName   string `json:"host_name"`
	Helo       string `json:"helo"`
	MsgSize    int64  `json:"msgsize"`
}

// PaginatedResponse is a generic wrapper for paginated results
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Filtered   int64 `json:"filtered"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
}

// DashboardStatsResponse for GET /api/v1/dashboard
type DashboardStatsResponse struct {
	Domains    int64              `json:"domains"`
	Mailboxes  int64              `json:"mailboxes"`
	Aliases    int64              `json:"aliases"`
	RecentLogs []LogEntryResponse `json:"recent_logs"`
}
