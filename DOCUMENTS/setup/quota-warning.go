package main

// To compile:
// CGO_ENABLED=0 go build -ldflags="-s -w" -o quota-warning quota-warning.go
// upx --best --lzma quota-warning
// sudo cp quota-warning /usr/local/bin/quota-warning
// sudo chmod +x /usr/local/bin/quota-warning
// sudo chown vmail:vmail /usr/local/bin/quota-warning

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Dovecot passes 2 arguments to the script (excluding the program name itself)
	// 1 = Usage percentage (e.g., 80, 95)
	// 2 = Logged in user's email address
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <percentage> <user_email>\n", os.Args[0])
		os.Exit(1)
	}

	percent := os.Args[1]
	user := os.Args[2]

	// Determine the domain name equivalently to $(hostname -d) in bash
	domain := "example.com"
	out, err := exec.Command("hostname", "-d").Output()
	if err == nil {
		d := strings.TrimSpace(string(out))
		if d != "" {
			domain = d
		}
	} else {
		// Fallback to hostname parsed if hostname -d fails
		host, _ := os.Hostname()
		parts := strings.SplitN(host, ".", 2)
		if len(parts) > 1 {
			domain = parts[1]
		} else {
			domain = host
		}
	}

	from := fmt.Sprintf("postmaster@%s", domain)

	// Build the email payload using RFC 822 format (with \r\n endings)
	emailContent := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Email Quota Warning (%s%%)\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"\r\n"+
		"Dear user,\r\n\r\n"+
		"Your mailbox has reached %s%% of its total storage capacity.\r\n"+
		"Please delete old or unwanted messages to free up space and avoid being blocked from receiving new messages.\r\n\r\n"+
		"Best regards,\r\n"+
		"System Administrator\r\n", from, user, percent, percent)

	err = sendPlain("127.0.0.1:25", from, user, []byte(emailContent))
	if err != nil {
		log.Fatalf("Error sending email via SMTP: %v\n", err)
	}
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
