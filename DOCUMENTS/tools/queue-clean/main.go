package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	POSTQUEUE = "postqueue"
	POSTSUPER = "postsuper"
)

// Regexes equivalentes ao Perl
var (
	reQueueID = regexp.MustCompile(`^([0-9A-F]*)[ *!]*(\d+) *([a-zA-Z]{3} [a-zA-Z]{3} [ 0-9]{2} \d{2}:\d{2}:\d{2}) +([^ ]+)`)
	reError   = regexp.MustCompile(`^ *\((.+)\)`)
	reRecip   = regexp.MustCompile(`^ *(.+.+)`)
	reTotal   = regexp.MustCompile(`^-- (\d+) Kbytes in (\d+) Requests\.`)
	reQHeader = regexp.MustCompile(`^-Que.*`)
)

func parsePostfixDate(dateStr string) (time.Time, error) {
	// Formato Postfix não tem ano; usamos o ano atual
	year := time.Now().Year()
	full := fmt.Sprintf("%d %s", year, strings.TrimSpace(dateStr))

	// Tenta com dia de dois dígitos (ex: "Jan 12") ou um dígito com espaço (ex: "Jan  2")
	for _, layout := range []string{
		"2006 Mon Jan  2 15:04:05",
		"2006 Mon Jan 2 15:04:05",
	} {
		if t, err := time.Parse(layout, full); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("não foi possível parsear data: %q", dateStr)
}

func main() {
	// Lê o email a ser deletado pelo primeiro argumento
	delEmail := "MAILER-DAEMON"
	if len(os.Args) > 1 {
		delEmail = os.Args[1]
	} else {
		fmt.Fprintln(os.Stderr, "USO:")
		fmt.Fprintln(os.Stderr, "queue-clean <email_to_delete>")
		fmt.Fprintln(os.Stderr, "email default: MAILER-DAEMON")
	}

	// Escapa o @ para uso no regex (equivalente ao s/\@/\\@/g do Perl)
	escapedEmail := strings.ReplaceAll(delEmail, "@", `\@`)
	reDelEmail, err := regexp.Compile(escapedEmail)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao compilar regex para email %q: %v\n", escapedEmail, err)
		os.Exit(1)
	}

	// Executa postqueue -p
	cmd := exec.Command(POSTQUEUE, "-p")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "postqueue error: %v\n", err)
		os.Exit(1)
	}

	var (
		Q          = map[string]bool{}
		x          = 0
		queueID    string
		sender     string
		recipient  string
		delta      int64
	)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "Mail queue is empty"):
			fmt.Println("Nada na fila")
			os.Exit(0)

		case reQueueID.MatchString(line):
			m := reQueueID.FindStringSubmatch(line)
			queueID = m[1]
			sender = m[4]
			recipient = ""

			t, err := parsePostfixDate(m[3])
			if err == nil {
				delta = int64(time.Since(t).Seconds())
			} else {
				delta = 0
			}

		case reError.MatchString(line):
			// captura erro mas não usa — mantido para equivalência com o Perl
			_ = reError.FindStringSubmatch(line)[1]

		case reTotal.MatchString(line):
			// totalsize / totalrequests — não utilizados neste script

		case reQHeader.MatchString(line):
			// do nothing

		case reRecip.MatchString(line):
			recipient = reRecip.FindStringSubmatch(line)[1]

			if delta > 1 && recipient != "" {
				if reDelEmail.MatchString(sender) || reDelEmail.MatchString(recipient) {
					Q[queueID] = true
					x++
				}
				delta = 0
				sender = ""
				recipient = ""
			} else if delta > 0 && recipient != "" {
				delta = 0
				sender = ""
				recipient = ""
			}
		}
	}

	if x == 0 {
		fmt.Println("nothing to do")
		os.Exit(0)
	}

	// Envia os queue IDs para postsuper -d - via stdin
	psCmd := exec.Command(POSTSUPER, "-d", "-")
	var stdin bytes.Buffer
	for k := range Q {
		stdin.WriteString(k + "\n")
	}
	psCmd.Stdin = &stdin
	psCmd.Stdout = os.Stdout
	psCmd.Stderr = os.Stderr

	if err := psCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "postsuper error: %v\n", err)
		os.Exit(1)
	}
}
