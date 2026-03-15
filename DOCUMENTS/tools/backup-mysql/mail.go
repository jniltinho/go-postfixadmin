package main

import (
	"fmt"
	"net/smtp"
	"strings"
)

func sendMail(cfg config, subject string) {
	fullSubject := fmt.Sprintf("%s JOB: backup-mysql HOST: %s", subject, hostname)

	to := append(cfg.EmailTo, cfg.EmailCC...)
	body := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		cfg.EmailFrom,
		strings.Join(cfg.EmailTo, ", "),
		fullSubject,
		logBuf.String(),
	)

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPServer)
	err := smtp.SendMail(cfg.SMTPServer+":"+cfg.SMTPPort, auth, cfg.EmailFrom, to, []byte(body))
	if err != nil {
		logMsg(fmt.Sprintf("mail failed; %s", err))
	}
}
