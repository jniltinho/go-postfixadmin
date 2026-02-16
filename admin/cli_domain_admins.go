package admin

import (
	"log/slog"
	"os"

	"go-postfixadmin/internal/models"

	"github.com/jedib0t/go-pretty/v6/table"
	"gorm.io/gorm"
)

// ListDomainAdmins lists all domain administrators in the database
func ListDomainAdmins(db *gorm.DB) {
	var domainAdmins []models.DomainAdmin
	if err := db.Order("username DESC").Find(&domainAdmins).Error; err != nil {
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

	t.Render()
}
