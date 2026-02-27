package utils

import (
	"fmt"
	"net/smtp"

	"github.com/spf13/viper"
)

// SendWelcomeEmail envia uma mensagem de boas-vindas para a caixa de correio recém-criada.
// Utiliza configurações da seção [smtp] no config.toml.
func SendWelcomeEmail(adminUsername, newMailbox string) error {
	server := viper.GetString("smtp.server")
	if server == "" {
		server = "127.0.0.1"
	}
	port := viper.GetInt("smtp.port")
	if port == 0 {
		port = 25
	}
	subject := viper.GetString("smtp.subject")
	if subject == "" {
		subject = "Welcome!"
	}
	body := viper.GetString("smtp.body")
	if body == "" {
		body = "Hi,\n\nWelcome to your new account."
	}

	addr := fmt.Sprintf("%s:%d", server, port)

	// Montagem simples dos cabeçalhos e corpo do e-mail no formato RFC 822
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", newMailbox, adminUsername, subject, body))

	// Envia o e-mail através do servidor SMTP configurado
	// Como não há autenticação configurável atualmente, passamos nil no parâmetro smtp.Auth
	err := smtp.SendMail(addr, nil, adminUsername, []string{newMailbox}, msg)
	if err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}
