package main

// To compile:
// CGO_ENABLED=0 go build -ldflags="-s -w" -o quota-warning quota-warning.go
// upx --best --lzma quota-warning
// sudo cp quota-warning /usr/local/bin/quota-warning
// sudo chmod +x /usr/local/bin/quota-warning
// sudo chown vmail:vmail /usr/local/bin/quota-warning

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
)

type Translation struct {
	Subject string
	Body    string
}

var translations = map[string]Translation{
	"en": {
		Subject: "Email Quota Warning (%s%%)",
		Body: "Dear user,\r\n\r\n" +
			"Your mailbox has reached %s%% of its total storage capacity.\r\n" +
			"Please delete old or unwanted messages to free up space and avoid being blocked from receiving new messages.\r\n\r\n" +
			"Best regards,\r\n" +
			"System Administrator\r\n",
	},
	"pt-br": {
		Subject: "Aviso de Quota de E-mail (%s%%)",
		Body: "Caro usuário,\r\n\r\n" +
			"Sua caixa de correio atingiu %s%% da sua capacidade total de armazenamento.\r\n" +
			"Por favor, exclua mensagens antigas ou indesejadas para liberar espaço e evitar o bloqueio no recebimento de novas mensagens.\r\n\r\n" +
			"Atenciosamente,\r\n" +
			"Administrador do Sistema\r\n",
	},
	"es": {
		Subject: "Aviso de Cuota de Correo (%s%%)",
		Body: "Estimado usuario,\r\n\r\n" +
			"Su buzón ha alcanzado el %s%% de su capacidad total de almacenamiento.\r\n" +
			"Por favor, elimine mensajes antiguos o no deseados para liberar espacio y evitar ser bloqueado para recibir nuevos mensajes.\r\n\r\n" +
			"Atentamente,\r\n" +
			"Administrador del Sistema\r\n",
	},
}

func main() {
	langPtr := flag.String("lang", "en", "Language for the email (en, pt-br, or es)")
	flag.Parse()

	// Dovecot passes arguments to the script
	// 1 = Usage percentage (e.g., 80, 95)
	// 2 = Logged in user's email address
	args := flag.Args()
	if len(args) != 2 {
		fmt.Printf("Usage: %s [--lang=en|pt-br|es] <percentage> <user_email>\n", os.Args[0])
		os.Exit(1)
	}

	percent := args[0]
	user := args[1]

	lang := strings.ToLower(*langPtr)
	t, ok := translations[lang]
	if !ok {
		t = translations["en"]
	}

	// Determine the domain name equivalently to $(hostname -d) in bash
	domain := "example.com"
	out, err := exec.Command("hostname", "-d").Output()
	if err == nil {
		d := strings.TrimSpace(string(out))
		if d != "" {
			domain = d
		}
	} else {
		// Fallback to hostname parsed if hostname -d fails
		host, _ := os.Hostname()
		parts := strings.SplitN(host, ".", 2)
		if len(parts) > 1 {
			domain = parts[1]
		} else {
			domain = host
		}
	}

	from := fmt.Sprintf("postmaster@%s", domain)

	// Build the email payload using RFC 822 format (with \r\n endings)
	emailContent := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"\r\n"+
		"%s", from, user, fmt.Sprintf(t.Subject, percent), fmt.Sprintf(t.Body, percent))

	err = sendPlain("127.0.0.1:25", from, user, []byte(emailContent))
	if err != nil {
		log.Fatalf("Error sending email via SMTP: %v\n", err)
	}
}

// sendPlain establishes a connection to the SMTP server and sends the email
// without attempting STARTTLS or authentication (typical for local submission).
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
