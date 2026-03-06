package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const query = `
SELECT v.email, v.subject, v.body, v.active,
  v.activefrom, v.activeuntil, v.interval_time,
  m.maildir FROM vacation v JOIN mailbox m ON m.username = v.email`

func run(cmd *cobra.Command, args []string) error {
	dsn := viper.GetString("database.url")
	if dsn == "" {
		return fmt.Errorf("database.url não definido no config.toml")
	}

	mailBase := viper.GetString("server.mail_base")
	if mailBase == "" {
		mailBase = "/var/vmail"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("erro abrindo banco: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("erro na query: %w", err)
	}
	defer rows.Close()

	now := time.Now()

	for rows.Next() {
		var v Vacation
		if err := rows.Scan(&v.Email, &v.Subject, &v.Body, &v.Active, &v.ActiveFrom, &v.ActiveUntil, &v.IntervalTime, &v.Maildir); err != nil {
			log.Println("scan error:", err)
			continue
		}

		maildirPath := filepath.Join(mailBase, v.Maildir, "Maildir")
		sievePath := filepath.Join(maildirPath, ".dovecot.sieve")

		if v.IsActive(now) {
			if err := writeSieve(sievePath, generateSieve(v)); err != nil {
				log.Println("erro escrevendo sieve:", err)
				continue
			}
			compileSieve(sievePath)
			log.Println("vacation criado:", v.Email)
		} else {
			removeSieve(sievePath)
		}
	}

	return rows.Err()
}
