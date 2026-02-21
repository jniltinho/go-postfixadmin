package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
==================================================
CONFIG (igual vacation.pl)
https://chatgpt.com/c/6999f953-3190-832b-83ed-dac6ae2b8fb7
==================================================
*/

var (
	dbDSN           = "user:password@tcp(127.0.0.1:3306)/postfix"
	dateFormat      = "02/01/2006"
	maxAliasLoop    = 20
	defaultInterval = 24 // fallback horas
)

var db *sql.DB

/*
==================================================
MAIN
Suporta:
vacation -f sender@example.com -- recipient@example.com
==================================================
*/

func main() {

	envelopeSender := flag.String("f", "", "Envelope sender")
	flag.Parse()

	args := flag.Args()

	if *envelopeSender == "" || len(args) != 1 {
		log.Fatal("Usage: vacation -f sender@example.com -- recipient@example.com")
	}

	sender := strings.ToLower(*envelopeSender)
	recipient := strings.ToLower(args[0])

	var err error
	db, err = sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	headers := readHeaders()

	if skipMessage(headers, sender, recipient) {
		return
	}

	ok, realAddr := findRealAddress(recipient, 0)
	if !ok {
		return
	}

	if !shouldNotify(realAddr, sender) {
		return
	}

	subject := headers["Subject"]

	err = sendVacation(realAddr, sender, subject)
	if err != nil {
		log.Println("Send error:", err)
		return
	}

	updateNotification(realAddr, sender)
}

/*
==================================================
READ STDIN (igual perl)
==================================================
*/

func readHeaders() map[string]string {

	headers := make(map[string]string)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}

func skipMessage(headers map[string]string, sender, recipient string) bool {

	if sender == "" || sender == recipient {
		return true
	}

	if _, ok := headers["Auto-Submitted"]; ok {
		return true
	}
	if _, ok := headers["Precedence"]; ok {
		return true
	}
	if _, ok := headers["X-Loop"]; ok {
		return true
	}

	return false
}

/*
==================================================
ALIAS RESOLUTION (igual perl)
==================================================
*/

func findRealAddress(email string, depth int) (bool, string) {

	if depth > maxAliasLoop {
		return false, ""
	}

	active, _ := checkVacation(email)
	if active {
		return true, email
	}

	rows, err := db.Query(`SELECT goto FROM alias WHERE address=?`, email)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var g string
			rows.Scan(&g)
			for _, addr := range strings.Split(g, ",") {
				addr = strings.TrimSpace(strings.ToLower(addr))
				ok, real := findRealAddress(addr, depth+1)
				if ok {
					return true, real
				}
			}
		}
	}

	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		var target string
		err = db.QueryRow(
			`SELECT target_domain FROM alias_domain WHERE alias_domain=?`,
			parts[1],
		).Scan(&target)

		if err == nil {
			return findRealAddress(parts[0]+"@"+target, depth+1)
		}
	}

	return false, ""
}

/*
==================================================
VACATION CHECK
==================================================
*/

func checkVacation(email string) (bool, error) {

	var e string
	err := db.QueryRow(`
		SELECT email FROM vacation
		WHERE email=? AND active=1
		AND activefrom <= NOW()
		AND activeuntil >= NOW()
	`, email).Scan(&e)

	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func shouldNotify(email, sender string) bool {

	var interval int
	err := db.QueryRow(
		`SELECT interval_time FROM vacation WHERE email=?`,
		email,
	).Scan(&interval)

	if err != nil {
		interval = defaultInterval
	}

	var last time.Time
	err = db.QueryRow(`
		SELECT sent FROM vacation_notification
		WHERE on_vacation=? AND notified=?
	`, email, sender).Scan(&last)

	if err == sql.ErrNoRows {
		return true
	}

	return time.Since(last).Hours() >= float64(interval)
}

func updateNotification(email, sender string) {
	db.Exec(`
		REPLACE INTO vacation_notification
		(on_vacation, notified, sent)
		VALUES (?, ?, NOW())
	`, email, sender)
}

/*
==================================================
SEND MAIL (igual perl)
==================================================
*/

func sendVacation(email, sender, origSubject string) error {

	var subject, body string
	var fromDate, untilDate time.Time

	err := db.QueryRow(`
		SELECT subject, body, activefrom, activeuntil
		FROM vacation WHERE email=?
	`, email).Scan(&subject, &body, &fromDate, &untilDate)

	if err != nil {
		return err
	}

	body = strings.ReplaceAll(body, "<%From_Date>", fromDate.Format(dateFormat))
	body = strings.ReplaceAll(body, "<%Until_Date>", untilDate.Format(dateFormat))
	subject = strings.ReplaceAll(subject, "$SUBJECT", origSubject)

	msg := buildMessage(email, sender, subject, body)

	domain := strings.Split(email, "@")[1]
	mx, err := net.LookupMX(domain)
	if err != nil || len(mx) == 0 {
		return err
	}

	return smtp.SendMail(mx[0].Host+":25", nil, email, []string{sender}, msg)
}

func buildMessage(from, to, subject, body string) []byte {

	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "Auto-Submitted: auto-replied\r\n"
	msg += "Precedence: bulk\r\n"
	msg += "X-Loop: PostfixAdmin Vacation\r\n"
	msg += "Content-Type: text/plain; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += body

	return []byte(msg)
}

/*
==================================================
UTIL
==================================================
*/

func stripAddress(input string) string {
	re := regexp.MustCompile(`[\w\.\-\+]+@[\w\.\-]+\w+`)
	return strings.ToLower(re.FindString(input))
}
