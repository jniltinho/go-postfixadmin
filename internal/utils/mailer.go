package utils

import (
	"fmt"
	"net/smtp"
)

// SendWelcomeEmail envia uma mensagem de boas-vindas para a caixa de correio recém-criada.
// Utiliza uma conexão SMTP local (localhost:25) sem autenticação e sem SSL, conforme solicitado.
func SendWelcomeEmail(adminUsername, newMailbox string) error {
	addr := "127.0.0.1:25"

	subject := "Welcome!"
	body := "Hi,\n\nWelcome to your new account."

	// Montagem simples dos cabeçalhos e corpo do e-mail no formato RFC 822
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", newMailbox, adminUsername, subject, body))

	// Envia o e-mail através do servidor SMTP local
	// Como não há autenticação, passamos nil no parâmetro smtp.Auth
	err := smtp.SendMail(addr, nil, adminUsername, []string{newMailbox}, msg)
	if err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}
