package admin

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go-postfixadmin/internal/models"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gorm.io/gorm"
)

// FormatQuota formats a byte value into a human-readable string (e.g., "1.5 GB")
func FormatQuota(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024000 // Using 1024000 as per user preference (base 1000ish or config specific)
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := fmt.Sprintf("%.1f", float64(bytes)/float64(div))
	return fmt.Sprintf("%s %cB", strings.TrimSuffix(val, ".0"), "KMGTPE"[exp])
}

// ListAllDomains lists all domains in the database
func ListAllDomains(db *gorm.DB) {
	var domains []models.Domain
	if err := db.Where("domain != ?", "ALL").Order("domain ASC").Find(&domains).Error; err != nil {
		slog.Error("Failed to fetch domains", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Domain", "Description", "Aliases", "Mailboxes", "Quota", "Active", "Modified"})

	for _, d := range domains {
		active := "No"
		if d.Active {
			active = "Yes"
		}
		// Domain quota stored in MB/units, apply 1024000 * 1024000 for display?
		// Wait, previous code had: formatQuota(d.Quota * 1024000 * 1024000)
		// Let's keep that logic.
		t.AppendRow(table.Row{d.Domain, d.Description, d.Aliases, d.Mailboxes, FormatQuota(d.Quota * 1024000 * 1024000), active, d.Modified.Format("2006-01-02 15:04:05")})
	}
	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Domains", strings.Join(os.Args, " ")})
	t.Render()
}

// ListAllMailboxes lists all mailboxes in the database
func ListAllMailboxes(db *gorm.DB) {
	var mailboxes []models.Mailbox
	if err := db.Order("domain ASC, username ASC").Find(&mailboxes).Error; err != nil {
		slog.Error("Failed to fetch mailboxes", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Username", "Name", "Domain", "Quota", "Active", "Modified"})

	for _, m := range mailboxes {
		active := "No"
		if m.Active {
			active = "Yes"
		}
		// Previous logic: formatQuota(m.Quota * 1024000)
		t.AppendRow(table.Row{m.Username, m.Name, m.Domain, FormatQuota(m.Quota * 1024000), active, m.Modified.Format("2006-01-02 15:04:05")})

	}
	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Mailboxes", strings.Join(os.Args, " ")})
	t.Render()
}

// ListAllAdmins lists all administrators in the database
func ListAllAdmins(db *gorm.DB) {
	var admins []models.Admin
	if err := db.Order("username ASC").Find(&admins).Error; err != nil {
		slog.Error("Failed to fetch admins", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Username", "Superadmin", "Active", "Modified"})

	for _, a := range admins {
		active := "No"
		if a.Active {
			active = "Yes"
		}
		super := "No"
		if a.Superadmin {
			super = "Yes"
		}
		t.AppendRow(table.Row{a.Username, super, active, a.Modified.Format("2006-01-02 15:04:05")})
	}
	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Admins", strings.Join(os.Args, " ")})
	t.Render()
}

// ListAllAliases lists all aliases in the database
func ListAllAliases(db *gorm.DB) {
	var aliases []models.Alias
	if err := db.Order("domain ASC, address ASC").Find(&aliases).Error; err != nil {
		slog.Error("Failed to fetch aliases", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Address", "Goto", "Domain", "Active", "Modified"})

	for _, a := range aliases {
		active := "No"
		if a.Active {
			active = "Yes"
		}
		t.AppendRow(table.Row{a.Address, a.Goto, a.Domain, active, a.Modified.Format("2006-01-02 15:04:05")})
	}
	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Aliases", strings.Join(os.Args, " ")})
	t.Render()
}

// ListDomainAdmins lists all domain administrators in the database
func ListDomainAdmins(db *gorm.DB) {
	var domainAdmins []models.DomainAdmin
	if err := db.Order("domain ASC").Find(&domainAdmins).Error; err != nil {
		slog.Error("Failed to fetch domain admins", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Username", "Domain", "Active", "Created"})

	for _, da := range domainAdmins {
		active := "No"
		if da.Active {
			active = "Yes"
		}
		t.AppendRow(table.Row{da.Username, da.Domain, active, da.Created.Format("2006-01-02 15:04:05")})
	}
	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Domain Admins", strings.Join(os.Args, " ")})
	t.Render()
}

// ListLogs lists all system logs in the database
func ListLogs(db *gorm.DB) {
	var logs []models.Log
	if err := db.Order("id DESC").Limit(100).Find(&logs).Error; err != nil {
		slog.Error("Failed to fetch logs", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Date", "Username", "Domain", "Action", "Data"})

	for _, l := range logs {
		t.AppendRow(table.Row{l.Timestamp.Format("2006-01-02 15:04:05"), l.Username, l.Domain, strings.ToUpper(l.Action), l.Data})
	}
	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List System Logs (Last 100)", strings.Join(os.Args, " ")})
	t.Render()
}
