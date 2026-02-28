package utils

import (
	"crypto/tls"
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
	smtpType := viper.GetString("smtp.type")
	if smtpType == "" {
		smtpType = "plain"
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
	switch smtpType {
	case "tls":
		return sendTLS(addr, server, adminUsername, newMailbox, msg)
	case "starttls":
		return sendStartTLS(addr, server, adminUsername, newMailbox, msg)
	default: // "plain" ou vazio
		// Força a conexão sem configuração automática de TLS
		return sendPlain(addr, adminUsername, newMailbox, msg)
	}
}

func sendPlain(addr, from, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial error: %w", err)
	}
	defer c.Close()

	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func sendTLS(addr, server, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         server,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial error: %w", err)
	}

	c, err := smtp.NewClient(conn, server)
	if err != nil {
		return fmt.Errorf("smtp new client error: %w", err)
	}
	defer c.Close()

	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func sendStartTLS(addr, server, from, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial error: %w", err)
	}
	defer c.Close()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         server,
	}

	if err = c.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls error: %w", err)
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
