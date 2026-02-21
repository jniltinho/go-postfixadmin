# Guia de Instalação: Servidor de E-mail (Ubuntu) + Go-PostfixAdmin

Este guia passo a passo ensina como preparar um servidor de e-mail completo no Ubuntu utilizando **Postfix**, **Dovecot**, **MySQL** e gerenciar tudo através do **Go-PostfixAdmin**.

---

## 1. Atualizar o Sistema e Instalar Dependências

No Ubuntu, atualize os pacotes e instale os serviços básicos necessários:

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install postfix postfix-mysql dovecot-core dovecot-imapd dovecot-pop3d dovecot-lmtpd dovecot-mysql mysql-server -y
sudo apt install certbot git curl -y
```

Durante a instalação do Postfix, o assistente perguntará o tipo de configuração. Selecione **"Internet Site"** e informe o seu domínio principal (ex: `example.com`).

---

## 2. Configurar o Banco de Dados MySQL

Acesse o console do MySQL:

```bash
sudo mysql
```

Execute os comandos abaixo para criar o banco de dados e o usuário que o Postfix, Dovecot e o Go-PostfixAdmin usarão:

```sql
CREATE DATABASE postfix;
CREATE USER 'postfix'@'localhost' IDENTIFIED BY 'sua_senha_segura';
GRANT ALL ON postfix.* TO 'postfix'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

> **Nota:** Lembre-se de substituir `sua_senha_segura` por uma senha forte em todos os passos deste guia.

---

## 3. Instalar e Configurar o Go-PostfixAdmin

O Go-PostfixAdmin será responsável por gerenciar a estrutura do banco (tabelas, domínios, contas, aliases, etc.).

1. **Obter o Aplicativo:**

   Você pode clonar o repositório orginal ou baixar diretamente o último executável compilado das Releases.
   
   *Baixar o binário e Repositório:*
   ```bash
   sudo mkdir -p /opt/go-postfixadmin
   cd /opt/go-postfixadmin
   # Substitua a URL abaixo pela URL da última release do seu repositório:
   sudo curl -L -O https://github.com/jniltinho/go-postfixadmin/releases/latest/download/postfixadmin_X.X.X_linux_amd64.tar.gz
   sudo git clone https://github.com/jniltinho/go-postfixadmin.git download
   sudo tar -xzvf postfixadmin_*.tar.gz
   ```
   
2. **Gerar Certificados SSL Iniciais (Certbot):**
   
   Antes de configurar as rotas seguras do servidor, gere os certificados primários. Pare qualquer serviço que utilize a porta 80 e rode:
   ```bash
   sudo certbot certonly --standalone -d mail.example.com
   ```

3. **Configurar o Ambiente (.env):**
   Crie o arquivo `/opt/go-postfixadmin/.env` e adicione as variáveis de ambiente necessárias para o correto funcionamento:
   
   ```env
   # Configurações do Banco de Dados
   DATABASE_URL="postfix:sua_senha_segura@tcp(localhost:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
   DB_DRIVER=mysql
   
   # Configurações do Servidor Web
   PORT=8080
   
   # Chave Secreta de Sessão (Gere uma string 64-char via: openssl rand -hex 32)
   SESSION_SECRET=your_super_secret_session_key_here
   
   # (Opcional) Configurações de SSL para servidor standalone seguro
   SSL_CERT="/etc/letsencrypt/live/mail.example.com/fullchain.pem"
   SSL_KEY="/etc/letsencrypt/live/mail.example.com/privkey.pem"
   ```

3. **Executar as Migrations:**
   Antes de iniciar o serviço, crie as tabelas necessárias rodando as migrations:
   ```bash
   cd /opt/go-postfixadmin
   ./postfixadmin migrate
   ```

4. **Configurar o Serviço do Systemd:**
   Copie o arquivo de serviço (fornecido em `DOCUMENTS/setup/postfixadmin.service`) para o systemd:
   ```bash
   sudo cp download/DOCUMENTS/setup/postfixadmin.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now postfixadmin.service
   ```
   *Verifique os logs com: `tail -f /opt/go-postfixadmin/postfixadmin.log`*

---

## 4. Configurar o Postfix (`/etc/postfix/main.cf`)

Faça um backup do arquivo original:
```bash
sudo cp /etc/postfix/main.cf /etc/postfix/main.cf.bkp
```

Edite o `/etc/postfix/main.cf` e altere/adicione as seguintes entradas:

```ini
# Domínio e hostname (Ajuste para a sua realidade)
myhostname = mail.example.com
mydomain   = example.com
myorigin   = $mydomain

# Virtual mailboxes (Integração com MySQL via Go-PostfixAdmin)
virtual_mailbox_base    = /var/vmail
virtual_mailbox_domains = proxy:mysql:/etc/postfix/sql/mysql_virtual_domains_maps.cf
virtual_mailbox_maps    = proxy:mysql:/etc/postfix/sql/mysql_virtual_mailbox_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_mailbox_maps.cf
virtual_alias_maps      = proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_maps.cf,
                          proxy:mysql:/etc/postfix/sql/mysql_virtual_alias_domain_catchall_maps.cf

# UID/GID do usuário vmail (criaremos a seguir)
virtual_uid_maps = static:1001
virtual_gid_maps = static:1001

# Entrega via Dovecot LMTP
virtual_transport = lmtp:unix:private/dovecot-lmtp

# SASL via Dovecot (Autenticação)
smtpd_sasl_type           = dovecot
smtpd_sasl_path           = private/auth
smtpd_sasl_auth_enable    = yes
smtpd_recipient_restrictions = permit_sasl_authenticated, permit_mynetworks, reject_unauth_destination

# TLS (Recomendado via Let's Encrypt - configure os caminhos corretos)
# smtpd_tls_cert_file = /etc/letsencrypt/live/mail.example.com/fullchain.pem
# smtpd_tls_key_file  = /etc/letsencrypt/live/mail.example.com/privkey.pem
# smtpd_use_tls       = yes
# smtpd_tls_auth_only = yes
```

---

## 5. Mapeamentos SQL para o Postfix

Crie os arquivos de consulta SQL no diretório abaixo e garanta as permissões adequadas:

```bash
sudo mkdir -p /etc/postfix/sql
sudo chown root:postfix /etc/postfix/sql
```

> **Atenção:** Em todos os arquivos abaixo, substitua `sua_senha_segura` para a senha configurada no MySQL.

### `/etc/postfix/sql/mysql_virtual_domains_maps.cf`
```ini
user     = postfix
password = sua_senha_segura
hosts    = localhost
dbname   = postfix
query    = SELECT domain FROM domain WHERE domain='%s' AND active='1'
```

### `/etc/postfix/sql/mysql_virtual_mailbox_maps.cf`
```ini
user     = postfix
password = sua_senha_segura
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox WHERE username='%s' AND active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_maps.cf`
```ini
user     = postfix
password = sua_senha_segura
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias WHERE address='%s' AND active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_domain_maps.cf`
```ini
user     = postfix
password = sua_senha_segura
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain='%d' AND alias.address=CONCAT('%u','@',alias_domain.target_domain) AND alias.active='1' AND alias_domain.active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_domain_catchall_maps.cf`
```ini
user     = postfix
password = sua_senha_segura
hosts    = localhost
dbname   = postfix
query    = SELECT goto FROM alias,alias_domain WHERE alias_domain.alias_domain='%d' AND alias.address=CONCAT('@',alias_domain.target_domain) AND alias.active='1' AND alias_domain.active='1'
```

### `/etc/postfix/sql/mysql_virtual_alias_domain_mailbox_maps.cf`
```ini
user     = postfix
password = sua_senha_segura
hosts    = localhost
dbname   = postfix
query    = SELECT maildir FROM mailbox,alias_domain WHERE alias_domain.alias_domain='%d' AND mailbox.username=CONCAT('%u','@',alias_domain.target_domain) AND mailbox.active='1' AND alias_domain.active='1'
```

Proteja os arquivos:
```bash
sudo chmod 640 /etc/postfix/sql/*.cf
sudo chown root:postfix /etc/postfix/sql/*.cf
```

---

## 6. Configurar o Dovecot

### `/etc/dovecot/conf.d/10-ssl.conf` (Recomendado)

Configure o Dovecot para utilizar os mesmos certificados TLS/SSL do Postfix:
```ini
ssl = required
# O sinal '<' antes do caminho diz ao Dovecot para ler o conteúdo do arquivo
ssl_cert = </etc/letsencrypt/live/mail.example.com/fullchain.pem
ssl_key = </etc/letsencrypt/live/mail.example.com/privkey.pem
```

### `/etc/dovecot/dovecot.conf`

Edite ou acrescente ao arquivo principal:
```ini
protocols = imap lmtp pop3

# Caminho para os e-mails
mail_location = maildir:/var/vmail/%d/%n/Maildir

# Usuário de sistema para entrega (vmail mapeado com UID/GID 1001)
mail_uid = 1001
mail_gid = 1001
```

### `/etc/dovecot/conf.d/10-auth.conf`

Configure os mecanismos de autenticação e inclua a integração com MySQL:
```ini
auth_mechanisms = plain login
!include auth-sql.conf.ext
```

### `/etc/dovecot/conf.d/auth-sql.conf.ext`

Habilite a consulta SQL para usuários e senhas:
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

Configuração de conexão e queries para o Dovecot:
```ini
driver  = mysql
connect = host=localhost dbname=postfix user=postfix password=sua_senha_segura

# O default_pass_scheme deve corresponder ao formato de hash do Go-PostfixAdmin
default_pass_scheme = BLF-CRYPT

# Consulta de validação de senha
password_query = \
  SELECT username AS user, password \
  FROM mailbox WHERE username='%u' AND active='1'

# Consulta de diretório home e cotas
user_query = \
  SELECT CONCAT('/var/vmail/', maildir) AS home, \
         1001 AS uid, 1001 AS gid, \
         CONCAT('*:bytes=', quota) AS quota_rule \
  FROM mailbox WHERE username='%u' AND active='1'

# Atualização de quota usada em tempo real
iterate_query = SELECT username AS user FROM mailbox WHERE active='1'
```

### `/etc/dovecot/conf.d/10-master.conf`

Configure os sockets de comunicação com o Postfix:
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

## 7. Criar o Usuário e Diretório de E-mails

Crie o usuário de sistema `vmail` (Virtual Mail) que será dono de todos os arquivos de caixas de correio:

```bash
sudo groupadd -g 1001 vmail
sudo useradd -g vmail -u 1001 vmail -d /var/vmail -m
sudo chown -R vmail:vmail /var/vmail
```

---

## 8. Reiniciar e Validar Serviços

Após realizar todas as configurações, reinicie os serviços para aplicar as mudanças:

```bash
sudo systemctl restart postfix dovecot
sudo systemctl enable postfix dovecot
```

Valide se o suporte ao MySQL foi reconhecido pelo Postfix:
```bash
postconf -m | grep mysql
# A saída deve listar "mysql"
```

Valide a entrega de e-mails observando os logs do sistema:
```bash
sudo tail -f /var/log/mail.log
```
