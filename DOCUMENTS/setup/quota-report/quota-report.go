package main

// To compile:
// CGO_ENABLED=0 go build -ldflags="-s -w" -o quota-report quota-report.go
// upx --best --lzma quota-report
// sudo cp quota-report /usr/local/bin/quota-report
// sudo chmod +x /usr/local/bin/quota-report
// sudo chown vmail:vmail /usr/local/bin/quota-report

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
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

func main() {
	emailFlag := flag.String("email", "", "Send report to this email address via sendmail")
	flag.Parse()

	if *emailFlag == "" {
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

	if *emailFlag == "" {
		t.SetOutputMirror(os.Stdout)
		t.Render()
		return
	}

	sendReportViaEmail(t, *emailFlag)
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

	// Send via sendmail
	subject := "Dovecot Quota Report"

	// Create an HTML message with basic styling
	body := fmt.Sprintf("From: %s\n"+
		"Subject: %s\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\n\n"+
		"<html><head><style>"+
		"table { border-collapse: collapse; font-family: sans-serif; font-size: 14px; } "+
		"th, td { border: 1px solid #dddddd; padding: 8px; text-align: left; } "+
		"th { background-color: #f2f2f2; }"+
		"</style></head><body><h2>Dovecot Quota Report</h2>\n%s\n</body></html>\n",
		fromEmail, subject, htmlTable)

	cmdSendmail := exec.Command("sendmail", "-t", toEmail)
	cmdSendmail.Stdin = bytes.NewBufferString(body)

	if err := cmdSendmail.Run(); err != nil {
		log.Fatalf("Error sending email via sendmail: %v", err)
	}

	fmt.Printf("Report successfully sent to %s\n", toEmail)
}
