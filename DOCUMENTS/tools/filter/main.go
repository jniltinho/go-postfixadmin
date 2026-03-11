package main

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Exit codes from <sysexits.h>
const (
	exOK       = 0  // Success
	exTempFail = 75 // EX_TEMPFAIL — Postfix will retry later
)

const (
	inspectDir = "/var/spool/filter"
	sendmail   = "/usr/sbin/sendmail"
)

var (
	stress     string
	size       int
	hostIP     string
	hostName   string
	helo       string
	from       string
	recipients []string
)

func main() {
	parseArgs(os.Args[1:])

	// Open syslog
	sl, err := syslog.New(syslog.LOG_MAIL|syslog.LOG_INFO, "postfix/filter")
	if err != nil {
		fmt.Fprintf(os.Stderr, "filter: cannot open syslog: %v\n", err)
		os.Exit(exTempFail)
	}
	defer sl.Close()

	if len(recipients) == 0 {
		sl.Err("FILTER_ERROR: no recipients")
		os.Exit(exTempFail)
	}

	// Ensure inspect dir exists
	if err := os.MkdirAll(inspectDir, 0750); err != nil {
		sl.Err(fmt.Sprintf("FILTER_ERROR: cannot create %s: %v", inspectDir, err))
		os.Exit(exTempFail)
	}

	// Temp file: /var/spool/filter/<timestamp>_<pid>.eml
	tmpFile := fmt.Sprintf("%s/%d_%d.eml", inspectDir, time.Now().Unix(), os.Getpid())
	defer os.Remove(tmpFile)

	// Save stdin to temp file
	f, err := os.Create(tmpFile)
	if err != nil {
		sl.Err(fmt.Sprintf("FILTER_ERROR: cannot create temp file: %v", err))
		os.Exit(exTempFail)
	}

	n, err := io.Copy(f, os.Stdin)
	f.Close()
	if err != nil {
		sl.Err(fmt.Sprintf("FILTER_ERROR: cannot save mail: %v", err))
		os.Exit(exTempFail)
	}

	sl.Info(fmt.Sprintf("FILTER: from=%s to=%s size=%d bytes=%d host-ip=%s host-name=%s helo=%s",
		from, strings.Join(recipients, ","), size, n, hostIP, hostName, helo))

	// Reinject via sendmail
	args := []string{"-i", "-f", from, "--"}
	args = append(args, recipients...)

	input, err := os.Open(tmpFile)
	if err != nil {
		sl.Err(fmt.Sprintf("FILTER_ERROR: cannot open temp file: %v", err))
		os.Exit(exTempFail)
	}
	defer input.Close()

	cmd := exec.Command(sendmail, args...)
	cmd.Stdin = input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		sl.Err(fmt.Sprintf("FILTER_ERROR: sendmail failed: %v", err))
		os.Exit(exTempFail)
	}

	sl.Info(fmt.Sprintf("FILTER_ACCEPT: from=%s recipients=%d", from, len(recipients)))
}

func parseArgs(args []string) {
	rcpt := false
	for i := 0; i < len(args); i++ {
		if rcpt {
			if r := strings.TrimSpace(args[i]); r != "" {
				recipients = append(recipients, r)
			}
			continue
		}
		switch args[i] {
		case "--":
			rcpt = true
		case "--stress":
			if i+1 < len(args) {
				i++
				stress = args[i]
			}
		case "--size":
			if i+1 < len(args) {
				i++
				size, _ = strconv.Atoi(args[i])
			}
		case "--host-ip":
			if i+1 < len(args) {
				i++
				hostIP = args[i]
			}
		case "--host-name":
			if i+1 < len(args) {
				i++
				hostName = args[i]
			}
		case "--helo":
			if i+1 < len(args) {
				i++
				helo = args[i]
			}
		case "--from":
			if i+1 < len(args) {
				i++
				from = strings.ToLower(args[i])
			}
		}
	}
}
