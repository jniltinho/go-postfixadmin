package transport

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ListAllTransports lists all transport entries in the database.
func ListAllTransports(db *gorm.DB) {
	var transports []models.TransportList
	if err := db.Order("domain ASC").Find(&transports).Error; err != nil {
		slog.Error("Failed to fetch transports", "error", err)
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Domain", "Transport", "Active", "Created", "Modified"})

	for _, tr := range transports {
		active := "No"
		if tr.Active {
			active = "Yes"
		}
		t.AppendRow(table.Row{
			tr.ID,
			tr.Domain,
			tr.Transport,
			active,
			tr.Created.Format("2006-01-02 15:04:05"),
			tr.Modified.Format("2006-01-02 15:04:05"),
		})
	}

	style := table.StyleDefault
	style.Format.Footer = text.FormatDefault
	t.SetStyle(style)
	t.AppendFooter(table.Row{"List All Transports", strings.Join(os.Args, " ")})
	t.Render()
}

// AddTransport creates a new transport entry.
// input must be in the format "domain:transport", e.g. "example.com:smtp:[relay.example.com]:25"
func AddTransport(db *gorm.DB, input string) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	colonIdx := strings.Index(input, ":")
	if colonIdx < 1 {
		slog.Error("Input must be in the format domain:transport (e.g. example.com:smtp:[relay]:25)")
		os.Exit(1)
	}

	domain := strings.TrimSpace(input[:colonIdx])
	transport := strings.TrimSpace(input[colonIdx+1:])

	if domain == "" || transport == "" {
		slog.Error("Both domain and transport are required")
		os.Exit(1)
	}

	var existing models.TransportList
	if err := db.Where("domain = ?", domain).First(&existing).Error; err == nil {
		slog.Error("Transport entry already exists for domain", "domain", domain, "transport", existing.Transport)
		os.Exit(1)
	}

	now := time.Now()
	entry := models.TransportList{
		Domain:    domain,
		Transport: transport,
		Active:    true,
		Created:   now,
		Modified:  now,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, "CLI", "127.0.0.1", domain, "create_transport", transport)
	}); err != nil {
		slog.Error("Failed to create transport entry", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Transport created: %s -> %s\n", domain, transport)
}

// DeleteTransport removes a transport entry by domain name.
func DeleteTransport(db *gorm.DB, domain string) {
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	domain = strings.TrimSpace(domain)
	if domain == "" {
		slog.Error("Domain is required")
		os.Exit(1)
	}

	var entry models.TransportList
	if err := db.Where("domain = ?", domain).First(&entry).Error; err != nil {
		slog.Error("Transport entry not found", "domain", domain)
		os.Exit(1)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("domain = ?", domain).Delete(&models.TransportList{}).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, "CLI", "127.0.0.1", domain, "delete_transport", entry.Transport)
	}); err != nil {
		slog.Error("Failed to delete transport entry", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Transport deleted: %s -> %s\n", domain, entry.Transport)
}
