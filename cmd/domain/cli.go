package domain

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ListAllDomains lists all domains in the database.
func ListAllDomains(db *gorm.DB) {
	var domains []models.Domain
	if err := db.Where("domain != ?", "ALL").Order("domain ASC").Find(&domains).Error; err != nil {
		slog.Error("Failed to fetch domains", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Domain", "Description", "Aliases", "Mailboxes", "Active", "Modified"})

	for _, d := range domains {
		active := "No"
		if d.Active {
			active = "Yes"
		}
		t.AppendRow(table.Row{
			d.Domain,
			d.Description,
			d.Aliases,
			d.Mailboxes,
			active,
			d.Modified.Format("2006-01-02 15:04:05"),
		})
	}

	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Domains", strings.Join(os.Args, " ")})
	t.Render()
}

// AddDomain creates a new domain entry.
func AddDomain(db *gorm.DB, domainName, description string, aliases, mailboxes int) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	domainName = strings.TrimSpace(domainName)
	if domainName == "" {
		slog.Error("Domain name is required")
		os.Exit(1)
	}

	if _, err := repositories.GetDomainByName(db, domainName); err == nil {
		slog.Error("Domain already exists", "domain", domainName)
		os.Exit(1)
	}

	now := time.Now()
	d := models.Domain{
		Domain:      domainName,
		Description: description,
		Aliases:     aliases,
		Mailboxes:   mailboxes,
		Active:      true,
		Created:     now,
		Modified:    now,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&d).Error; err != nil {
			return err
		}
		// Ensure the ALL domain admin entry exists
		da := models.DomainAdmin{
			Username: "admin",
			Domain:   domainName,
			Created:  now,
			Active:   true,
		}
		tx.Where(models.DomainAdmin{Domain: domainName}).FirstOrCreate(&da)
		return utils.LogAction(tx, "CLI", "127.0.0.1", domainName, "create_domain", domainName)
	}); err != nil {
		slog.Error("Failed to create domain", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Domain created: %s\n", domainName)
}

// DeleteDomain removes a domain and all associated data.
func DeleteDomain(db *gorm.DB, domainName string) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	domainName = strings.TrimSpace(domainName)
	if domainName == "" {
		slog.Error("Domain name is required")
		os.Exit(1)
	}

	if _, err := repositories.GetDomainByName(db, domainName); err != nil {
		slog.Error("Domain not found", "domain", domainName)
		os.Exit(1)
	}

	if err := repositories.DeleteDomain(db, domainName, "CLI", "127.0.0.1"); err != nil {
		slog.Error("Failed to delete domain", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Domain deleted: %s\n", domainName)
}
