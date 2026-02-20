# SETUP_MAILSERVER.md

> Guia completo de configuração de servidor de e-mail com **Postfix + Dovecot + Go-PostfixAdmin + MySQL**

---

## Sumário

1. [Banco de Dados MySQL](#1-banco-de-dados-mysql)
2. [Postfix — main.cf](#2-postfix--maincf)
3. [Arquivos SQL do Postfix](#3-arquivos-sql-do-postfix)
4. [Dovecot](#4-dovecot)
5. [Usuário e Diretório de E-mails](#5-criar-usuário-e-diretório-de-e-mails)
6. [Reiniciar Serviços](#6-reiniciar-tudo)
7. [Pontos de Atenção](#pontos-de-atenção)

---

## 1. Banco de Dados MySQL

```sql
CREATE DATABASE postfix;
CREATE USER 'postfix'@'localhost' IDENTIFIED BY 'sua_senha';
GRANT ALL ON postfix.* TO 'postfix'@'localhost';
FLUSH PRIVILEGES;
```

```bash
## Após criar o banco faça:
export DATABASE_URL="postfix:sua_senha@tcp(localhost:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
./postfixadmin migrate
## No Binário do Go-PostfixAdmin, ele cria as tabelas automaticamente.
```

---

## 2. Postfix — `/etc/postfix/main.cf`

```ini
# Domínio e hostname
myhostname = mail.example.com
mydomain   = example.com
myorigin   = $mydomain

# Virtual mailboxes
virtual_mailbox_base    = /var/vmail
virtual_mailbox_domains = proxy:mysql:/etc/postfix/sql/mysql_virtual_domains_maps.cf
virtual_mailbox_maps    = proxy:mysql:/etc/postfix/sql/mysql_virtual_mailbox_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_mailbox_maps.cf
virtual_alias_maps      = proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_catchall_maps.cf

# UID/GID do usuário vmail (crie com: useradd -u 1001 -g 1001 vmail)
virtual_uid_maps = static:1001
virtual_gid_maps = static:1001

# Entrega via Dovecot LMTP
virtual_transport = lmtp:unix:private/dovecot-lmtp

# SASL via Dovecot
smtpd_sasl_type           = dovecot
smtpd_sasl_path           = private/auth
smtpd_sasl_auth_enable    = yes
smtpd_recipient_restrictions = permit_sasl_authenticated, permit_mynetworks, reject_unauth_destination

# TLS
smtpd_tls_cert_file = /etc/letsencrypt/live/mail.example.com/fullchain.pem
smtpd_tls_key_file  = /etc/letsencrypt/live/mail.example.com/privkey.pem
smtpd_use_tls       = yes
smtpd_tls_auth_only = yes
```

---

## 3. Arquivos SQL do Postfix

Crie os arquivos abaixo em `/etc/postfix/sql/` e em seguida proteja-os:

```bash
chmod 640 /etc/postfix/sql/*.cf
chown root:postfix /etc/postfix/sql/*.cf
```

### `mysql_virtual_domains_maps.cf`

```ini
user     = postfix
password = sua_senha
hosts    = localhost
dbname   = postfix
query    = SELECT domain FROM domain WHERE domain='%s' AND active='1'
```

### `mysql_virtual_mailbox_maps.cf`

```ini
user     = postfix
password = sua_senha
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox WHERE username='%s' AND active='1'
```

### `mysql_virtual_alias_maps.cf`

```ini
user     = postfix
password = sua_senha
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias WHERE address='%s' AND active='1'
```

### `mysql_virtual_alias_domain_maps.cf`

```ini
user     = postfix
password = sua_senha
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain
           WHERE alias_domain.alias_domain='%d'
           AND alias.address=CONCAT('%u','@',alias_domain.target_domain)
           AND alias.active='1' AND alias_domain.active='1'
```

### `mysql_virtual_alias_domain_catchall_maps.cf`

```ini
user     = postfix
password = sua_senha
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain
           WHERE alias_domain.alias_domain='%d'
           AND alias.address=CONCAT('@',alias_domain.target_domain)
           AND alias.active='1' AND alias_domain.active='1'
```

### `mysql_virtual_alias_domain_mailbox_maps.cf`

```ini
user     = postfix
password = sua_senha
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox,alias_domain
           WHERE alias_domain.alias_domain='%d'
           AND mailbox.username=CONCAT('%u','@',alias_domain.target_domain)
           AND mailbox.active='1' AND alias_domain.active='1'
```

---

## 4. Dovecot

### `/etc/dovecot/dovecot.conf`

```ini
protocols = imap lmtp

mail_location = maildir:/var/vmail/%d/%n/Maildir

# Usuário de sistema para entrega
mail_uid = 1001
mail_gid = 1001
```

### `/etc/dovecot/conf.d/10-auth.conf`

```ini
auth_mechanisms = plain login
!include auth-sql.conf.ext
```

### `/etc/dovecot/conf.d/auth-sql.conf.ext`

```ini
passdb {
  driver = sql
  args   = /etc/dovecot/dovecot-sql.conf.ext
}
userdb {
  driver = sql
  args   = /etc/dovecot/dovecot-sql.conf.ext
}
```

### `/etc/dovecot/dovecot-sql.conf.ext`

```ini
driver  = mysql
connect = host=localhost dbname=postfix user=postfix password=sua_senha

# Deve bater com $CONF['encrypt'] no PostfixAdmin
default_pass_scheme = BLF-CRYPT

password_query = \
  SELECT username AS user, password \
  FROM mailbox WHERE username='%u' AND active='1'

user_query = \
  SELECT CONCAT('/var/vmail/', maildir) AS home, \
         1001 AS uid, 1001 AS gid, \
         CONCAT('*:bytes=', quota) AS quota_rule \
  FROM mailbox WHERE username='%u' AND active='1'

# Para atualizar quota usada em tempo real
iterate_query = SELECT username AS user FROM mailbox WHERE active='1'
```

### `/etc/dovecot/conf.d/10-master.conf` — Sockets para o Postfix

```ini
service lmtp {
  unix_listener /var/spool/postfix/private/dovecot-lmtp {
    mode  = 0600
    user  = postfix
    group = postfix
  }
}

service auth {
  unix_listener /var/spool/postfix/private/auth {
    mode  = 0660
    user  = postfix
    group = postfix
  }
}
```

---

## 5. Criar usuário e diretório de e-mails

```bash
groupadd -g 1001 vmail
useradd -g vmail -u 1001 vmail -d /var/vmail -m
```

---

## 6. Reiniciar tudo

```bash
systemctl restart postfix dovecot
```

---

## Pontos de Atenção

| Item | Detalhe |
|------|---------|
| **Criptografia consistente** | O `default_pass_scheme` no Dovecot deve bater com o método configurado no PostfixAdmin |
| **UID/GID do vmail** | Deve ser o mesmo em `virtual_uid_maps`/`virtual_gid_maps` (Postfix) e `mail_uid`/`mail_gid` (Dovecot) |
| **Suporte MySQL no Postfix** | Verifique com `postconf -m \| grep mysql`. Se não houver, instale `postfix-mysql` |
| **Permissões dos arquivos SQL** | Mantenha `640` com owner `root:postfix` para segurança |
| **TLS** | Sempre use `smtpd_tls_auth_only = yes` para evitar transmissão de credenciais sem criptografia |
