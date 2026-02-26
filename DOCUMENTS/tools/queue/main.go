package main

// O original esta neste Link
// www.arschkrebs.de/postfix/scripts/mailqfmt.pl
// http://www.arschkrebs.de/postfix/scripts/
// Postfix mailq file reformatter, (C) 2003 by Matthias Andree
// This file is licensed to you under the conditions of the
// GNU General Public License v2.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	POSTQUEUE = "postqueue"
	POSTSUPER = "postsuper"
)

var (
	hashMail = map[string]int{}
	hashTo   = map[string]int{}
	queueOut []string

	// Flags
	flagDelete  = flag.Bool("del", false, "Deletar emails na fila")
	flagMax     = flag.Int("max", 300, "Máximo de mensagens por remetente para deletar")
	flagGetJSON = flag.Bool("getjson", false, "Exibir saída em JSON")
	flagDebug   = flag.Bool("debug", false, "Modo verboso/debug")
)

// Regexes
var (
	reTo      = regexp.MustCompile(`^\s+.*@(.*)`)
	reHeader  = regexp.MustCompile(`^[0-9A-F]+\s*([ !*])\s+(\d+)\s+(\S+\s+\S+\s+\d+\s+\d+:\d+:\d+)\s+(.+)$`)
	reFrom    = regexp.MustCompile(`^(\w|-|_|\.)+@((\w|-|_)+\.)+[a-zA-Z]{2,}$`)
	reQueueID = regexp.MustCompile(`^([0-9A-F]*)[ *!]*(\d+) *([a-zA-Z]{3} [a-zA-Z]{3} [ 0-9]{2} \d{2}:\d{2}:\d{2}) +([^ ]+)`)
	reError   = regexp.MustCompile(`^ *\((.+)\)`)
	reRecip   = regexp.MustCompile(`^ *(.+.+)`)
	reTotal   = regexp.MustCompile(`^-- (\d+) Kbytes in (\d+) Requests\.`)
	reQHeader = regexp.MustCompile(`^-Que.*`)
)

func main() {
	// Run postqueue -p
	out, err := exec.Command(POSTQUEUE, "-p").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar postqueue: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		queueOut = append(queueOut, line)
	}

	if len(queueOut) <= 1 {
		os.Exit(0)
	}

	// Parse queue
	for _, line := range queueOut {
		if strings.HasPrefix(line, "Mail queue is empty") {
			fmt.Println(line)
			continue
		}
		if strings.HasPrefix(line, "-") {
			continue // skip header
		}
		if m := reTo.FindStringSubmatch(line); m != nil {
			hashTo[m[1]]++
		}
		if m := reHeader.FindStringSubmatch(line); m != nil {
			from := m[4]
			if reFrom.MatchString(from) {
				hashMail[from]++
			}
		}
	}

	// Parse flags
	flag.Bool("d", false, "Alias para --del")
	flag.Bool("v", false, "Alias para --debug")
	flag.Bool("verbose", false, "Alias para --debug")
	flag.Int("m", 300, "Alias para --max")
	flag.Parse()

	// Se nenhum argumento foi passado, lista e sai
	if flag.NFlag() == 0 && len(os.Args) == 1 {
		listQueue()
		os.Exit(0)
	}

	if *flagGetJSON {
		listQueueJSON()
	} else if flag.NFlag() == 0 {
		listQueue()
	}

	if *flagDelete {
		deleteQueue()
	}
}

func sortedByValue(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	return keys
}

func listQueue() {
	fmt.Println("LISTA FILA TO DOMINIO")
	fmt.Println("--------------------------------------------")
	total := 0
	for _, k := range sortedByValue(hashTo) {
		total += hashTo[k]
		fmt.Printf("%d - %s \n", hashTo[k], k)
	}
	fmt.Printf("\n%d -> TOTAL NA FILA TO\n", total)

	fmt.Println("\nLISTA FILA FROM")
	fmt.Println("--------------------------------------------")
	total = 0
	for _, k := range sortedByValue(hashMail) {
		total += hashMail[k]
		fmt.Printf("%d - %s \n", hashMail[k], k)
	}
	fmt.Printf("\n%d -> TOTAL NA FILA FROM\n", total)
}

func listQueueJSON() {
	type DomainEntry map[string]interface{}
	type EmailEntry map[string]interface{}

	var domains []DomainEntry
	totalDomain := 0
	for _, k := range sortedByValue(hashTo) {
		totalDomain += hashTo[k]
		domains = append(domains, DomainEntry{"domain": k, "total": hashTo[k]})
	}
	domains = append(domains, DomainEntry{"total_domain": totalDomain})

	var emails []EmailEntry
	totalEmail := 0
	for _, k := range sortedByValue(hashMail) {
		totalEmail += hashMail[k]
		emails = append(emails, EmailEntry{"email": k, "total": hashMail[k]})
	}
	emails = append(emails, EmailEntry{"total_email": totalEmail})

	geral := []map[string]interface{}{
		{"domain": domains, "email": emails},
	}

	jsonBytes, err := json.Marshal(geral)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gerar JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func deleteQueue() {
	max := *flagMax
	fmt.Printf("\nDELETE QUEUE FROM COM MAXIMO DE %d\n", max)
	fmt.Println("--------------------------------------------")

	for _, k := range sortedByValue(hashMail) {
		if hashMail[k] > max {
			deleteMailQueue(k)
		}
	}
}

func parsePostfixDate(dateStr string) (time.Time, error) {
	// Postfix date format: "Mon Jan  2 15:04:05" (no year)
	// We add the current year
	year := time.Now().Year()
	full := fmt.Sprintf("%d %s", year, strings.TrimSpace(dateStr))
	t, err := time.Parse("2006 Mon Jan  2 15:04:05", full)
	if err != nil {
		t, err = time.Parse("2006 Mon Jan 2 15:04:05", full)
	}
	return t, err
}

func deleteMailQueue(delEmail string) {
	// Escape @ for use in regex
	escaped := strings.ReplaceAll(delEmail, "@", `\@`)
	reDelEmail := regexp.MustCompile(escaped)

	Q := map[string]bool{}
	x := 0

	var queueID, sender, recipient, errMsg string
	var delta int64

	for _, line := range queueOut {
		if strings.HasPrefix(line, "Mail queue is empty") {
			fmt.Println("Nada na fila")
			os.Exit(0)
		}

		if m := reQueueID.FindStringSubmatch(line); m != nil {
			queueID = m[1]
			dateStr := m[3]
			sender = m[4]
			errMsg = ""
			recipient = ""

			t, err := parsePostfixDate(dateStr)
			if err == nil {
				delta = int64(time.Since(t).Seconds())
			} else {
				delta = 0
			}
		} else if m := reError.FindStringSubmatch(line); m != nil {
			errMsg = m[1]
			_ = errMsg
		} else if reQHeader.MatchString(line) {
			// do nothing
		} else if reTotal.MatchString(line) {
			// do nothing
		} else if m := reRecip.FindStringSubmatch(line); m != nil {
			recipient = m[1]
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

	// Pipe queue IDs to postsuper -d -
	cmd := exec.Command(POSTSUPER, "-d", "-")
	var stdin bytes.Buffer
	for k := range Q {
		stdin.WriteString(k + "\n")
	}
	cmd.Stdin = &stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "postsuper error: %v\n", err)
		os.Exit(1)
	}

	// Flush queue
	if err := exec.Command(POSTQUEUE, "-f").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer flush na fila: %v\n", err)
	}
}
