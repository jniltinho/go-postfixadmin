package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// FlexibleInt allows deserializing numbers that arrive as string or int in JSON
type FlexibleInt int64

func (fi *FlexibleInt) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(strings.Trim(string(b), "\""))
	if s == "" || s == "-" || s == "unlimited" {
		*fi = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		*fi = 0
		return nil
	}
	*fi = FlexibleInt(v)
	return nil
}

// DoveadmQuota defines the structure of the JSON returned by Dovecot
type DoveadmQuota []struct {
	User    string      `json:"username"`
	Root    string      `json:"root"`
	Type    string      `json:"type"`
	Value   FlexibleInt `json:"value"`
	Limit   FlexibleInt `json:"limit"`
	Percent FlexibleInt `json:"percent"`
}

// ShowQuotaReport fetches and displays the Dovecot quota report, optionally sending it by email
func ShowQuotaReport(email string) {
	if email == "" {
		fmt.Println("Fetching quota information from Dovecot...")
	}

	// Execute: doveadm -f json quota get -A
	cmd := exec.Command("doveadm", "-f", "json", "quota", "get", "-A")
	output, err := cmd.Output()

	if err != nil {
		log.Fatalf("Error executing doveadm: %v. Check if you have root permissions.", err)
	}

	var results DoveadmQuota
	if err := json.Unmarshal(output, &results); err != nil {
		log.Fatalf("Error processing JSON: %v", err)
	}

	// Sort by domain first, then by email username
	sort.Slice(results, func(i, j int) bool {
		userI := results[i].User
		userJ := results[j].User

		partsI := strings.Split(userI, "@")
		partsJ := strings.Split(userJ, "@")

		domainI := ""
		domainJ := ""
		if len(partsI) > 1 {
			domainI = partsI[1]
		}
		if len(partsJ) > 1 {
			domainJ = partsJ[1]
		}

		if domainI != domainJ {
			return domainI < domainJ
		}
		return userI < userJ
	})

	t := table.NewWriter()
	t.AppendHeader(table.Row{"User", "Usage (MB)", "Limit (MB)", "Status"})

	for _, q := range results {
		if q.Type != "STORAGE" {
			continue
		}
		usedMB := float64(q.Value) / 1024
		limitMB := float64(q.Limit) / 1024

		t.AppendRow(table.Row{
			q.User,
			fmt.Sprintf("%.2f", usedMB),
			fmt.Sprintf("%.0f", limitMB),
			fmt.Sprintf("%d%%", int(q.Percent)),
		})
	}
	t.AppendSeparator()

	if email == "" {
		t.SetOutputMirror(os.Stdout)
		t.Render()
		return
	}

	sendReportViaEmail(t, email)
}

func sendReportViaEmail(t table.Writer, toEmail string) {
	// Capture go-pretty output as HTML table for email
	t.SetStyle(table.StyleLight)
	htmlTable := t.RenderHTML()

	// Determine domain name to form the Sender Address (like $(hostname -d))
	domain := "example.com"
	out, err := exec.Command("hostname", "-d").Output()
	if err == nil {
		d := strings.TrimSpace(string(out))
		if d != "" {
			domain = d
		}
	} else {
		host, _ := os.Hostname()
		parts := strings.SplitN(host, ".", 2)
		if len(parts) > 1 {
			domain = parts[1]
		} else {
			domain = host
		}
	}
	fromEmail := fmt.Sprintf("postmaster@%s", domain)

	// Send via SMTP using sendPlain
	subject := "Dovecot Quota Report"

	// Create an HTML message with basic styling
	// Note: With SMTP we need to provide the 'To:' header manually inside the payload.
	body := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
		"<html><head><style>"+
		"table { border-collapse: collapse; font-family: sans-serif; font-size: 14px; } "+
		"th, td { border: 1px solid #dddddd; padding: 8px; text-align: left; } "+
		"th { background-color: #f2f2f2; }"+
		"</style></head><body><h2>Dovecot Quota Report</h2>\r\n%s\r\n</body></html>\r\n",
		fromEmail, toEmail, subject, htmlTable)

	err = sendPlain("127.0.0.1:25", fromEmail, toEmail, []byte(body))
	if err != nil {
		log.Fatalf("Error sending email via SMTP: %v\n", err)
	}

	fmt.Printf("Report successfully sent to %s\n", toEmail)
}

// sendPlain establishes a connection to the SMTP server and sends the email
// without attempting STARTTLS or authentication (typical for local submission).
func sendPlain(addr, from, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial error: %w", err)
	}
	defer c.Close()

	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
