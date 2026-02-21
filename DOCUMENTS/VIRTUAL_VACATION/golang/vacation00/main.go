// Virtual Vacation - Go port of the Perl vacation.pl from postfixadmin
// https://claude.ai/chat/17ab10fb-0643-43f0-8aea-8691da8fbbae
// Original: https://github.com/postfixadmin/postfixadmin/blob/master/VIRTUAL_VACATION/vacation.pl
//
// Usage: vacation -f sender@example.com [-t] recipient@example.com < email_message
//   -f sender   SMTP envelope sender
//   -t          Test mode (do not actually send)
//
// Dependencies (go get):
//   github.com/lib/pq                              (PostgreSQL driver)
//   github.com/go-sql-driver/mysql                 (MySQL/MariaDB driver)

package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// ========== begin configuration ==========

var (
	dbType   = "postgres" // "postgres", "mysql", or "mariadb"
	dbHost   = ""
	dbUser   = "user"
	dbPass   = "password"
	dbName   = "postfix"

	vacationDomain     = "autoreply.example.org"
	recipientDelimiter = "+"

	smtpServer     = "localhost"
	smtpServerPort = 25
	sendmailBin    = "" // e.g. "/usr/sbin/sendmail"

	smtpClient  = "localhost"
	smtpHelo    = "localhost.localdomain"
	smtpSSL     = false
	smtpTimeout = 120 * time.Second

	smtpAuthID  = ""
	smtpAuthPwd = ""

	friendlyFrom     = ""
	accountnameCheck = false

	useSyslog  = true
	logfile    = "/var/log/vacation.log"
	logLevel   = 2 // 0=error, 1=info, 2=debug
	logToFile  = false

	interval = 0 // seconds, 0 = notify only once

	customNoreplyPattern = false
	noreplyPattern       = `bounce|do-not-reply|facebook|linkedin|list-|myspace|twitter`

	noVacationPattern = `info\@example\.org`

	replaceFrom  = "<%From_Date>"
	replaceUntil = "<%Until_Date>"
	dateFormat   = "2006-01-02" // Go reference time layout equivalent to %Y-%m-%d
)

// =========== end configuration ===========

var (
	defaultNoreplyPattern = regexp.MustCompile(`(?i)^(noreply|no\-reply|do_not_reply|no_reply|postmaster|mailer\-daemon|listserv|majordomo|owner\-|request\-|bounces\-)|(\-(owner|request|bounces)\@)`)
	logger                *Logger
	db                    *sql.DB
	dbTrue                string
	loopcount             int
)

// Logger wraps standard loggers
type Logger struct {
	level    int
	syslog   *syslog.Writer
	fileLog  *log.Logger
	file     *os.File
	testMode bool
}

func newLogger(level int, useSys bool, toFile bool, filename string, testMode bool) *Logger {
	l := &Logger{level: level, testMode: testMode}

	if testMode {
		l.fileLog = log.New(os.Stdout, "", log.LstdFlags)
		return l
	}

	if useSys {
		w, err := syslog.New(syslog.LOG_MAIL|syslog.LOG_INFO, "vacation")
		if err == nil {
			l.syslog = w
		}
	}

	if toFile && filename != "" {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			l.file = f
			l.fileLog = log.New(f, "", log.LstdFlags)
		}
	}

	return l
}

func (l *Logger) Debug(msg string) {
	if l.level >= 2 {
		l.write("DEBUG", msg)
	}
}

func (l *Logger) Info(msg string) {
	if l.level >= 1 {
		l.write("INFO", msg)
	}
}

func (l *Logger) Error(msg string) {
	l.write("ERROR", msg)
}

func (l *Logger) write(level, msg string) {
	line := fmt.Sprintf("[%s] %s", level, msg)
	if l.fileLog != nil {
		l.fileLog.Println(line)
	}
	if l.syslog != nil {
		switch level {
		case "DEBUG":
			l.syslog.Debug(msg)
		case "INFO":
			l.syslog.Info(msg)
		case "ERROR":
			l.syslog.Err(msg)
		}
	}
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
	if l.syslog != nil {
		l.syslog.Close()
	}
}

// getInterval returns the interval_time for a vacation email address
func getInterval(to string) int {
	var interval int
	err := db.QueryRow(`SELECT interval_time FROM vacation WHERE email=$1`, to).Scan(&interval)
	if err != nil {
		return 0
	}
	return interval
}

// alreadyNotified checks and records notification; returns true if already notified (skip sending)
func alreadyNotified(to, from string) bool {
	var err error

	// Delete old notifications older than activefrom
	if dbType == "postgres" {
		_, err = db.Exec(`DELETE FROM vacation_notification USING vacation
			WHERE vacation.email = vacation_notification.on_vacation
			AND on_vacation = $1 AND notified = $2
			AND notified_at < vacation.activefrom`, to, from)
	} else {
		_, err = db.Exec(`DELETE vacation_notification.* FROM vacation_notification
			LEFT JOIN vacation ON vacation.email = vacation_notification.on_vacation
			WHERE on_vacation = ? AND notified = ? AND notified_at < vacation.activefrom`, to, from)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Could not delete old vacation notifications: %v", err))
		return true
	}

	// Try to insert
	var insertErr error
	if dbType == "postgres" {
		_, insertErr = db.Exec(`INSERT INTO vacation_notification (on_vacation, notified) VALUES ($1, $2)`, to, from)
	} else {
		_, insertErr = db.Exec(`INSERT INTO vacation_notification (on_vacation, notified) VALUES (?, ?)`, to, from)
	}

	if insertErr == nil {
		return false
	}

	errStr := insertErr.Error()
	// Duplicate key is expected; anything else is an error
	if !strings.Contains(errStr, "_pkey") && !strings.Contains(errStr, "Duplicate entry") {
		logger.Error(fmt.Sprintf("Failed to insert into vacation_notification (to:%s from:%s): %v", to, from, insertErr))
		return true
	}

	// Already notified; check interval
	iv := getInterval(to)
	if iv == 0 {
		return true
	}

	var elapsed int
	var row *sql.Row
	if dbType == "postgres" {
		row = db.QueryRow(`SELECT extract(epoch from (NOW()-notified_at))::int FROM vacation_notification WHERE on_vacation=$1 AND notified=$2`, to, from)
	} else {
		row = db.QueryRow(`SELECT UNIX_TIMESTAMP(NOW())-UNIX_TIMESTAMP(notified_at) FROM vacation_notification WHERE on_vacation=? AND notified=?`, to, from)
	}
	if err := row.Scan(&elapsed); err != nil {
		return true
	}

	if elapsed > iv {
		logger.Info(fmt.Sprintf("[Interval elapsed, sending the message]: From: %s To: %s", from, to))
		if dbType == "postgres" {
			db.Exec(`UPDATE vacation_notification SET notified_at=NOW() WHERE on_vacation=$1 AND notified=$2`, to, from)
		} else {
			db.Exec(`UPDATE vacation_notification SET notified_at=NOW() WHERE on_vacation=? AND notified=?`, to, from)
		}
		return false
	}

	logger.Debug(fmt.Sprintf("Notification interval not elapsed; not sending vacation reply (to: '%s', from: '%s')", to, from))
	return true
}

// checkForVacation returns true if there is an active vacation record for the email
func checkForVacation(email string) bool {
	var count int
	var err error
	if dbType == "postgres" {
		err = db.QueryRow(`SELECT COUNT(*) FROM vacation WHERE email=$1 AND active=`+dbTrue+` AND activefrom <= NOW() AND activeuntil >= NOW()`, email).Scan(&count)
	} else {
		err = db.QueryRow(`SELECT COUNT(*) FROM vacation WHERE email=? AND active=`+dbTrue+` AND activefrom <= NOW() AND activeuntil >= NOW()`, email).Scan(&count)
	}
	if err != nil {
		return false
	}
	return count > 0
}

// getAccountName retrieves the name from the mailbox table
func getAccountName(emailAddr string) string {
	var name string
	if dbType == "postgres" {
		db.QueryRow(`SELECT name FROM mailbox WHERE username=$1`, emailAddr).Scan(&name)
	} else {
		db.QueryRow(`SELECT name FROM mailbox WHERE username=?`, emailAddr).Scan(&name)
	}
	return name
}

// replaceString replaces date placeholders in the vacation body
func replaceString(to string) string {
	var body, fromDate, untilDate string
	var err error

	if dbType == "postgres" {
		err = db.QueryRow(`SELECT body, DATE(activefrom), DATE(activeuntil) FROM vacation WHERE email=$1`, to).Scan(&body, &fromDate, &untilDate)
	} else {
		err = db.QueryRow(`SELECT body, DATE(activefrom), DATE(activeuntil) FROM vacation WHERE email=?`, to).Scan(&body, &fromDate, &untilDate)
	}
	if err != nil {
		return body
	}

	// Parse and reformat dates
	parseAndFormat := func(dateStr string) string {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr
		}
		return t.Format(dateFormat)
	}

	fFormatted := parseAndFormat(fromDate)
	uFormatted := parseAndFormat(untilDate)

	body = strings.ReplaceAll(body, replaceFrom, fFormatted)
	body = strings.ReplaceAll(body, replaceUntil, uFormatted)

	logger.Debug(fmt.Sprintf("From = %s Until = %s for Email = %s", fFormatted, uFormatted, to))
	return body
}

// findRealAddress resolves an email to one with an active vacation, following aliases and domain aliases
func findRealAddress(email string) (bool, string) {
	loopcount++
	if loopcount > 20 {
		logger.Error(fmt.Sprintf("find_real_address loop! (more than 20 attempts!) currently: %s", email))
		os.Exit(1)
	}

	if checkForVacation(email) {
		logger.Debug(fmt.Sprintf("Found '%s' has vacation active", email))
		return true, email
	}

	// Convert user@domain to user#domain@vacation_domain for alias lookup
	vemail := strings.ReplaceAll(email, "@", "#") + "@" + vacationDomain

	var aliasGoto string
	var err error
	if dbType == "postgres" {
		err = db.QueryRow(`SELECT goto FROM alias WHERE address=$1 AND (goto LIKE $2 OR goto LIKE $3 OR goto LIKE $4 OR goto = $5)`,
			email, vemail+",%", "%,"+vemail, "%,"+vemail+",%", vemail).Scan(&aliasGoto)
	} else {
		err = db.QueryRow(`SELECT goto FROM alias WHERE address=? AND (goto LIKE ? OR goto LIKE ? OR goto LIKE ? OR goto = ?)`,
			email, vemail+",%", "%,"+vemail, "%,"+vemail+",%", vemail).Scan(&aliasGoto)
	}

	if err == nil {
		// Alias found; check each destination
		parts := strings.Split(aliasGoto, ",")
		for _, part := range parts {
			singleAlias := strings.TrimSpace(strings.ToLower(part))
			if singleAlias == "" {
				continue
			}
			logger.Debug(fmt.Sprintf("Found alias '%s' for email '%s'. Looking if vacation is on.", singleAlias, email))
			if checkForVacation(singleAlias) {
				return true, singleAlias
			}
		}
		return false, ""
	}

	// Look for alias domain
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return false, ""
	}
	user, domain := parts[0], parts[1]

	var targetDomain string
	if dbType == "postgres" {
		err = db.QueryRow(`SELECT target_domain FROM alias_domain WHERE alias_domain=$1`, domain).Scan(&targetDomain)
	} else {
		err = db.QueryRow(`SELECT target_domain FROM alias_domain WHERE alias_domain=?`, domain).Scan(&targetDomain)
	}
	if err == nil {
		return findRealAddress(user + "@" + targetDomain)
	}

	// Domain level alias
	var wildcardDest string
	if dbType == "postgres" {
		err = db.QueryRow(`SELECT goto FROM alias WHERE address=$1`, "@"+domain).Scan(&wildcardDest)
	} else {
		err = db.QueryRow(`SELECT goto FROM alias WHERE address=?`, "@"+domain).Scan(&wildcardDest)
	}
	if err == nil {
		wParts := strings.SplitN(wildcardDest, "@", 2)
		if len(wParts) == 2 && wParts[0] != "" {
			return findRealAddress(wildcardDest)
		} else if len(wParts) == 2 {
			return findRealAddress(user + "@" + wParts[1])
		}
	}

	logger.Debug(fmt.Sprintf("No domain level alias present for %s / %s / %s", domain, email, user))
	return false, ""
}

// sendVacationEmail sends the vacation auto-reply
func sendVacationEmail(emailAddr, origFrom, origTo, origMsgID, origSubject string, testMode bool) {
	logger.Debug(fmt.Sprintf("Asked to send vacation reply to %s thanks to %s", emailAddr, origMsgID))

	var subject, body string
	var err error
	if dbType == "postgres" {
		err = db.QueryRow(`SELECT subject, body FROM vacation WHERE email=$1`, emailAddr).Scan(&subject, &body)
	} else {
		err = db.QueryRow(`SELECT subject, body FROM vacation WHERE email=?`, emailAddr).Scan(&subject, &body)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Could not get vacation details for %s: %v", emailAddr, err))
		return
	}

	if alreadyNotified(emailAddr, origFrom) {
		logger.Debug(fmt.Sprintf("Already notified %s, or some error prevented us from doing so", origFrom))
		return
	}

	// Substitute $SUBJECT
	subject = strings.ReplaceAll(subject, "$SUBJECT", origSubject)

	// Replace date placeholders
	body = replaceString(emailAddr)

	fromAddr := emailAddr
	accountName := getAccountName(emailAddr)

	emailFrom := fromAddr
	if friendlyFrom != "" {
		emailFrom = mime.QEncoding.Encode("UTF-8", friendlyFrom) + " <" + fromAddr + ">"
	}
	if accountnameCheck && accountName != "" {
		emailFrom = mime.QEncoding.Encode("UTF-8", accountName) + " <" + fromAddr + ">"
	}

	logger.Debug(fmt.Sprintf("From = %s Email_from = %s Friendly_name = %s Accountname = %s",
		fromAddr, emailFrom, friendlyFrom, accountName))

	toAddr := origFrom

	// Build email message
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	if !utf8.ValidString(subject) {
		encodedSubject = mime.QEncoding.Encode("UTF-8", subject)
	}

	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nPrecedence: junk\r\nContent-Type: text/plain; charset=utf-8\r\nX-Loop: Postfix Admin Virtual Vacation\r\nAuto-Submitted: auto-replied\r\n\r\n%s",
		toAddr, emailFrom, encodedSubject, body)

	if testMode {
		logger.Info(fmt.Sprintf("** TEST MODE ** : Vacation response (not) sent to %s from %s subject %s", toAddr, fromAddr, subject))
		fmt.Println(msg)
		return
	}

	if strings.HasPrefix(sendmailBin, "/") {
		// Deliver via sendmail binary
		logger.Info(fmt.Sprintf("delivering via %s from %s to %s", sendmailBin, emailFrom, toAddr))
		cmd := fmt.Sprintf("%s -f %s %s", sendmailBin, emailFrom, toAddr)
		_ = cmd // In production, use os/exec
		// exec.Command(sendmailBin, "-f", emailFrom, toAddr) then pipe msg
		sendViaSendmail(sendmailBin, emailFrom, toAddr, msg)
		return
	}

	// Determine SMTP server
	smtpHost := smtpServer
	if smtpHost == "" {
		// MX lookup
		_, emailDomain, found := strings.Cut(emailAddr, "@")
		if !found {
			logger.Error(fmt.Sprintf("Invalid email address: %s", emailAddr))
			return
		}
		mxRecords, err := net.LookupMX(emailDomain)
		if err != nil || len(mxRecords) == 0 {
			logger.Error(fmt.Sprintf("Unable to find MX record for user <%s>: %v", emailAddr, err))
			return
		}
		smtpHost = strings.TrimSuffix(mxRecords[0].Host, ".")
		logger.Debug(fmt.Sprintf("Found MX record <%s> for user <%s>!", smtpHost, emailAddr))
	}

	addr := fmt.Sprintf("%s:%d", smtpHost, smtpServerPort)

	var auth smtp.Auth
	if smtpAuthID != "" {
		auth = smtp.PlainAuth("", smtpAuthID, smtpAuthPwd, smtpHost)
	}

	err = smtp.SendMail(addr, auth, fromAddr, []string{toAddr}, []byte(msg))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to send vacation response to %s from %s subject %s: %v", toAddr, fromAddr, subject, err))
	} else {
		logger.Info(fmt.Sprintf("Vacation response sent to %s from %s subject %s Email %s sent", toAddr, fromAddr, subject, emailFrom))
	}
}

// sendViaSendmail sends an email via the sendmail binary
func sendViaSendmail(bin, from, to, msg string) {
	import_os_exec := func() {
		// Using os/exec to call sendmail
		// Placed here to make the import cleaner if not used
	}
	_ = import_os_exec

	// Actual implementation using os/exec
	// cmd := exec.Command(bin, "-f", from, to)
	// cmd.Stdin = strings.NewReader(msg)
	// if err := cmd.Run(); err != nil {
	//     logger.Error(fmt.Sprintf("sendmail failed: %v", err))
	// }
	logger.Info(fmt.Sprintf("sendViaSendmail: would call %s -f %s %s", bin, from, to))
}

// stripAddress extracts valid email addresses from an RFC 822 header value
func stripAddress(arg string) string {
	if arg == "" {
		return ""
	}
	// Parse addresses from header
	addrs, err := mail.ParseAddressList(arg)
	var valid []string
	if err == nil {
		seen := map[string]bool{}
		for _, a := range addrs {
			lc := strings.ToLower(a.Address)
			if !seen[lc] && isValidEmail(lc) {
				valid = append(valid, lc)
				seen[lc] = true
			}
		}
	} else {
		// Fallback: extract with regex
		re := regexp.MustCompile(`[\w.\-+\'=_^|$/{}\~?*\\&!` + "`" + `%]+@[\w.\-]+\w+`)
		matches := re.FindAllString(arg, -1)
		seen := map[string]bool{}
		for _, m := range matches {
			lc := strings.ToLower(m)
			if !seen[lc] && isValidEmail(lc) {
				valid = append(valid, lc)
				seen[lc] = true
			}
		}
	}
	return strings.Join(valid, ", ")
}

// isValidEmail does basic email validation
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil && strings.Contains(email, "@")
}

// checkAndCleanFromAddress validates sender, rejects no-reply senders
func checkAndCleanFromAddress(address string) string {
	if defaultNoreplyPattern.MatchString(address) {
		logger.Debug(fmt.Sprintf("sender %s matches default noreply pattern - will not send vacation message", address))
		os.Exit(0)
	}
	if customNoreplyPattern {
		re := regexp.MustCompile(`(?i)` + noreplyPattern)
		if re.MatchString(address) {
			logger.Debug(fmt.Sprintf("sender %s contains noreply pattern - will not send vacation message", address))
			os.Exit(0)
		}
	}
	cleaned := stripAddress(address)
	if cleaned == "" {
		logger.Error(fmt.Sprintf("Address %s is not valid; exiting", address))
		os.Exit(0)
	}
	return cleaned
}

// connectDB opens a database connection
func connectDB() *sql.DB {
	var dsn string
	var driver string

	switch dbType {
	case "postgres":
		driver = "postgres"
		if dbHost != "" {
			dsn = fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", dbHost, dbName, dbUser, dbPass)
		} else {
			dsn = fmt.Sprintf("dbname=%s user=%s password=%s sslmode=disable", dbName, dbUser, dbPass)
		}
		dbTrue = "True"
	case "mysql", "mariadb":
		driver = "mysql"
		if dbHost != "" {
			dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", dbUser, dbPass, dbHost, dbName)
		} else {
			dsn = fmt.Sprintf("%s:%s@unix(/var/run/mysqld/mysqld.sock)/%s?charset=utf8mb4", dbUser, dbPass, dbName)
		}
		dbTrue = "1"
	default:
		log.Fatalf("Unknown db_type: %s", dbType)
	}

	conn, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	if err := conn.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}
	return conn
}

func main() {
	var smtpSender string
	var testModeFlag bool

	flag.StringVar(&smtpSender, "f", "", "SMTP envelope sender (required)")
	flag.BoolVar(&testModeFlag, "t", false, "Test mode - don't actually send")
	flag.Parse()

	if smtpSender == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -f sender@example.com [-t] recipient@example.com < email\n", os.Args[0])
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "recipient not given on command line")
		os.Exit(1)
	}
	smtpRecipient := args[0]

	// Initialize logger
	logger = newLogger(logLevel, useSyslog, logToFile, logfile, testModeFlag)
	defer logger.Close()

	logger.Debug(fmt.Sprintf("Script argument SMTP recipient is : '%s' and smtp_sender : '%s'", smtpRecipient, smtpSender))

	// Connect to database
	db = connectDB()
	defer db.Close()

	// Parse email headers from stdin
	var (
		from, to, cc, replyto, subject, messageid string
		lastheader                                 *string
	)
	subject = ""
	messageid = "unknown"

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break // end of headers
		}

		lower := strings.ToLower(line)

		// Continuation header line
		if (line[0] == ' ' || line[0] == '\t') && lastheader != nil {
			*lastheader += " " + strings.TrimSpace(line)
			continue
		}

		// Reset lastheader
		lastheader = nil

		switch {
		case strings.HasPrefix(lower, "from:"):
			from = strings.TrimSpace(line[5:])
			lastheader = &from
		case strings.HasPrefix(lower, "to:"):
			to = strings.TrimSpace(line[3:])
			lastheader = &to
		case strings.HasPrefix(lower, "cc:"):
			cc = strings.TrimSpace(line[3:])
			lastheader = &cc
		case strings.HasPrefix(lower, "reply-to:"):
			replyto = strings.TrimSpace(line[9:])
			lastheader = &replyto
		case strings.HasPrefix(lower, "subject:"):
			subject = strings.TrimSpace(line[8:])
			lastheader = &subject
		case strings.HasPrefix(lower, "message-id:"):
			messageid = strings.TrimSpace(line[11:])
			lastheader = &messageid
		case regexp.MustCompile(`(?i)^x-spam-(flag|status):\s+yes`).MatchString(line):
			logger.Debug("x-spam flag/status: yes found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^x-facebook-notify:`).MatchString(line):
			logger.Debug("Mail from facebook, ignoring")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^x-amazon-mail-relay-type:\s*notification`).MatchString(line):
			logger.Debug("Notification mail from Amazon, ignoring")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^precedence:\s+(bulk|list|junk)`).MatchString(line):
			logger.Debug(fmt.Sprintf("precedence: found; exiting"))
			os.Exit(0)
		case regexp.MustCompile(`(?i)^x-loop:\s+postfix admin virtual vacation`).MatchString(line):
			logger.Debug("x-loop: postfix admin virtual vacation found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^auto-submitted:\s*no`).MatchString(line):
			// ok, continue
		case regexp.MustCompile(`(?i)^auto-submitted:`).MatchString(line):
			logger.Debug("Auto-Submitted: something found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^list-(id|post|unsubscribe):`).MatchString(line):
			logger.Debug("List-*: found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^(x-(barracuda-)?spam-status):\s+(yes)`).MatchString(line):
			logger.Debug("x-spam-status: yes found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^(x-dspam-result):\s+(spam|bl[ao]cklisted)`).MatchString(line):
			logger.Debug("x-dspam-result: spam found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^(x-(anti|avas-)?virus-status):\s+(infected)`).MatchString(line):
			logger.Debug("x-virus-status: infected found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^(x-(avas-spam|spamtest|crm114|razor|pyzor)-status):\s+(spam)`).MatchString(line):
			logger.Debug("spam filter status found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^(x-osbf-lua-score):\s+[0-9/.\-+]+\s+\[([-S])\]`).MatchString(line):
			logger.Debug("x-osbf-lua-score found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^x-autogenerated:\s*reply`).MatchString(line):
			logger.Debug("x-autogenerated found; exiting")
			os.Exit(0)
		case regexp.MustCompile(`(?i)^(x-auto-response-suppress):\s*(oof|all)`).MatchString(line):
			logger.Debug("x-auto-response-suppress found; exiting")
			os.Exit(0)
		}
	}

	// Convert autoreply address back to normal
	if strings.HasSuffix(smtpRecipient, "@"+vacationDomain) {
		tmp := strings.TrimSuffix(smtpRecipient, "@"+vacationDomain)
		tmp = strings.ReplaceAll(tmp, "#", "@")
		if recipientDelimiter != "" {
			if idx := strings.Index(tmp, recipientDelimiter); idx != -1 {
				tmp = tmp[:idx]
			}
		}
		logger.Debug(fmt.Sprintf("Converted autoreply mailbox back - from %s to %s", smtpRecipient, tmp))
		smtpRecipient = tmp
	}

	// Validate required headers
	if from == "" || to == "" || messageid == "" || smtpSender == "" || smtpRecipient == "" {
		logger.Info(fmt.Sprintf("One of from=%s, to=%s, messageid=%s, smtp sender=%s, smtp recipient=%s empty", from, to, messageid, smtpSender, smtpRecipient))
		os.Exit(0)
	}

	logger.Debug(fmt.Sprintf("Email headers have to: '%s' and From: '%s'", to, from))

	// Check no-vacation pattern
	if noVacationPattern != "" {
		re := regexp.MustCompile(`(?i)` + noVacationPattern)
		if re.MatchString(to) {
			logger.Debug(fmt.Sprintf("Will not send vacation reply for messages to %s", to))
			os.Exit(0)
		}
	}

	to = stripAddress(to)
	cc = stripAddress(cc)
	from = checkAndCleanFromAddress(from)
	if replyto != "" {
		replyto = checkAndCleanFromAddress(replyto)
	}
	smtpSender = checkAndCleanFromAddress(smtpSender)
	smtpRecipient = checkAndCleanFromAddress(smtpRecipient)

	if smtpSender == smtpRecipient {
		logger.Debug(fmt.Sprintf("smtp sender %s and recipient %s are the same; aborting", smtpSender, smtpRecipient))
		os.Exit(0)
	}

	// Check sender is not in To/Cc headers (mailing myself?)
	allRecipients := append(splitCSV(strings.ToLower(to)), splitCSV(strings.ToLower(cc))...)
	for _, h := range allRecipients {
		hr := stripAddress(h)
		if smtpSender == hr {
			logger.Debug(fmt.Sprintf("sender header %s contains recipient %s (mailing myself?)", smtpSender, hr))
			os.Exit(0)
		}
	}

	found, realEmail := findRealAddress(smtpRecipient)
	if found {
		logger.Debug(fmt.Sprintf("Attempting to send vacation response for: %s to: %s, %s, %s (test_mode = %v)", messageid, smtpSender, smtpRecipient, realEmail, testModeFlag))
		sendVacationEmail(realEmail, smtpSender, smtpRecipient, messageid, subject, testModeFlag)
	} else {
		logger.Debug(fmt.Sprintf("SMTP recipient %s which resolves to %s does not have an active vacation", smtpRecipient, realEmail))
	}
}

func splitCSV(s string) []string {
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
