package mailbox

import (
	"log/slog"
	"os"
	"strings"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gorm.io/gorm"
)

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
		unit := utils.GetQuotaMultiplier()
		t.AppendRow(table.Row{m.Username, m.Name, m.Domain, formatQuota(m.Quota * unit), active, m.Modified.Format("2006-01-02 15:04:05")})
	}

	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Mailboxes", strings.Join(os.Args, " ")})
	t.Render()
}
