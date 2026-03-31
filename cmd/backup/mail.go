package backup

import (
	"fmt"
	"net/smtp"
	"strings"
)

// SendMail sends the accumulated log buffer via SMTP.
func SendMail(cfg Config, subject string) {
	fullSubject := fmt.Sprintf("%s JOB: backup-mysql HOST: %s", subject, Hostname)

	to := append(cfg.EmailTo, cfg.EmailCC...)
	body := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		cfg.EmailFrom,
		strings.Join(cfg.EmailTo, ", "),
		fullSubject,
		LogBuf.String(),
	)

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPServer)
	err := smtp.SendMail(cfg.SMTPServer+":"+cfg.SMTPPort, auth, cfg.EmailFrom, to, []byte(body))
	if err != nil {
		LogMsg(fmt.Sprintf("mail failed: %s", err))
	}
}
