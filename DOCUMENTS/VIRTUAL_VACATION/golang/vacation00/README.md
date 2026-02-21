# Virtual Vacation - Go

Port em Go do script [vacation.pl](https://github.com/postfixadmin/postfixadmin/blob/master/VIRTUAL_VACATION/vacation.pl) do [PostfixAdmin](https://github.com/postfixadmin/postfixadmin).

Envia respostas automáticas de "ausência" para e-mails recebidos, consultando as configurações de férias no banco de dados do PostfixAdmin.

---

## Dependências

| Go module | Finalidade |
|---|---|
| `github.com/lib/pq` | Driver PostgreSQL |
| `github.com/go-sql-driver/mysql` | Driver MySQL / MariaDB |

---

## Instalação

### 1. Compilar o binário

```bash
cd /opt/vacation
go mod tidy
go build -o vacation vacation.go
chmod 755 vacation
```

### 2. Criar usuário dedicado

```bash
useradd -r -s /sbin/nologin vacation
chown vacation:vacation /opt/vacation/vacation
```

---

## Configuração

Edite as variáveis no topo do `vacation.go` antes de compilar:

```go
var (
    dbType   = "postgres"          // "postgres", "mysql" ou "mariadb"
    dbHost   = ""                  // vazio = socket UNIX
    dbUser   = "user"
    dbPass   = "password"
    dbName   = "postfix"

    vacationDomain     = "autoreply.example.org"
    recipientDelimiter = "+"

    smtpServer     = "localhost"
    smtpServerPort = 25
    sendmailBin    = ""            // ex: "/usr/sbin/sendmail"

    useSyslog  = true
    logfile    = "/var/log/vacation.log"
    logLevel   = 2                 // 0=error, 1=info, 2=debug
)
```

> O `vacationDomain` deve ser **idêntico** ao configurado no PostfixAdmin em **Configuration → vacation_domain**.

---

## Integração com o Postfix

### `/etc/postfix/master.cf`

Adicione o serviço pipe para o vacation:

```
vacation  unix  -  n  n  -  -  pipe
  flags=Rq user=vacation argv=/opt/vacation/vacation -f ${sender} -- ${recipient}
```

### `/etc/postfix/main.cf`

```
# Transport map para o domínio de autoreply
transport_maps = hash:/etc/postfix/transport

# Domínio de autoreply precisa ser aceito pelo Postfix
virtual_mailbox_domains = ..., autoreply.example.org
```

### `/etc/postfix/transport`

Define que e-mails para o domínio de autoreply vão para o pipe:

```
autoreply.example.org    vacation:
```

Após editar, execute:

```bash
postmap /etc/postfix/transport
postfix reload
```

---

## Como o fluxo funciona

O PostfixAdmin cria automaticamente aliases no banco quando um usuário ativa as férias, no formato:

```
usuario@dominio.com  →  usuario#dominio.com@autoreply.example.org
```

Quando um e-mail chega para `usuario@dominio.com`, o Postfix resolve o alias e entrega uma cópia para `usuario#dominio.com@autoreply.example.org`. O transport map redireciona para o pipe `vacation`, que é chamado assim:

```bash
vacation -f remetente@origem.com -- usuario#dominio.com@autoreply.example.org
```

O binário converte `usuario#dominio.com` de volta para `usuario@dominio.com`, consulta o banco de dados, e envia a resposta automática ao remetente original.

```
E-mail chega
    → Postfix
    → alias DB aponta para usuario#dom@autoreply.example.org
    → transport map → pipe vacation
    → vacation consulta DB
    → envia resposta automática
```

---

## Verificar aliases no banco

Quando férias estiver ativo para um usuário, deve aparecer um registro assim:

```sql
SELECT * FROM alias WHERE goto LIKE '%autoreply.example.org%';
```

---

## Uso

```
vacation -f sender@example.com [-t] recipient@example.com < email_message
```

| Flag | Descrição |
|---|---|
| `-f sender` | Endereço do envelope SMTP (obrigatório) |
| `-t` | Modo de teste — não envia o e-mail de verdade |

---

## Testes

### Modo de teste (sem envio real)

```bash
echo -e "From: outro@exemplo.com\nTo: usuario@seudominio.com\nSubject: Teste\nMessage-ID: <123@test>\n" \
  | /opt/vacation/vacation -f outro@exemplo.com -t usuario#seudominio.com@autoreply.example.org
```

### Verificar logs

```bash
tail -f /var/log/mail.log
tail -f /var/log/vacation.log
```

---

## Licença

Derivado do [PostfixAdmin](https://github.com/postfixadmin/postfixadmin), distribuído sob a licença GPL.
