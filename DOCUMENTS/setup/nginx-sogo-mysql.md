# Guia de Instalação: Nginx + SOGo Webmail + MySQL (Ubuntu/Debian)

Este documento descreve o processo de instalação e configuração do webmail SOGo utilizando Nginx como proxy reverso e MySQL (MariaDB) como banco de dados, integrado ao seu servidor de email (Go-PostfixAdmin / Postfix / Dovecot).

---

## 1. Pré-requisitos

* Servidor Ubuntu/Debian com Postfix, Dovecot e MariaDB já instalados e funcionais.
* Um domínio configurado apontando para o seu servidor (ex: `mail.seudominio.com.br`).
* Certificados SSL/TLS válidos (ex: Let's Encrypt).

---

## 2. Configuração do MariaDB para o SOGo

O SOGo precisa de um banco de dados próprio para armazenar preferências de usuários, contatos e calendários.

Acesse o console do MariaDB:
```bash
sudo mariadb
```

Crie o banco de dados e o usuário para o SOGo. Em seguida, crie uma View (Visão) para que o SOGo consiga ler os usuários do Go-PostfixAdmin formatados corretamente:
```sql
CREATE DATABASE sogo CHARSET='utf8mb4';
CREATE USER 'sogo'@'localhost' IDENTIFIED BY 'SUA_SENHA_SEGURA_AQUI';
GRANT ALL PRIVILEGES ON sogo.* TO 'sogo'@'localhost';
-- Permite que o SOGo leia a tabela de caixas de e-mail do postfix
GRANT SELECT ON postfix.mailbox TO 'sogo'@'localhost';
FLUSH PRIVILEGES;

USE sogo;
CREATE OR REPLACE VIEW sogo_users_view AS
SELECT
    username AS c_uid,             -- O e-mail completo (user@domain.com)
    username AS c_name,            -- Nome interno
    name AS c_cn,                  -- Nome de exibição (Common Name)
    password AS c_password,        -- Hash da senha
    username AS mail               -- Campo de e-mail para buscas
FROM postfix.mailbox
WHERE active = 1;

EXIT;
```

---

## 3. Instalação do SOGo

Adicione a chave e o repositório do SOGo para a sua versão do Ubuntu/Debian (certifique-se de referenciar o repositório correto da sua distro, exemplo abaixo para Ubuntu):

```bash
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-key 74FFC6D72B925A01
sudo add-apt-repository "deb [arch=amd64] https://packages.inverse.ca/SOGo/nightly/5/ubuntu/ $(lsb_release -sc) $(lsb_release -sc)"
sudo apt update
```

Instale o SOGo e o memcached (necessário para as sessões):
```bash
sudo apt install sogo sogo-activesync memcached -y
```

---

## 4. Configuração do SOGo (`/etc/sogo/sogo.conf`)

Faça backup da configuração original:
```bash
sudo cp /etc/sogo/sogo.conf /etc/sogo/sogo.conf.bkp
```

Edite o arquivo `/etc/sogo/sogo.conf`. Limpe o arquivo original (que vem com muitos comentários) e use esta configuração base adaptando para a sua realidade:

```ini
{
  /* Database configuration (MySQL SOGo) */
  SOGoProfileURL = "mysql://sogo:SUA_SENHA_SEGURA_AQUI@localhost:3306/sogo/sogo_user_profile";
  OCSFolderInfoURL = "mysql://sogo:SUA_SENHA_SEGURA_AQUI@localhost:3306/sogo/sogo_folder_info";
  OCSSessionsFolderURL = "mysql://sogo:SUA_SENHA_SEGURA_AQUI@localhost:3306/sogo/sogo_sessions_folder";
  OCSAdminURL = "mysql://sogo:UA_SENHA_SEGURA_AQUI@localhost:3306/sogo/sogo_admin";

  /* Mail */
  SOGoDraftsFolderName = Drafts;
  SOGoSentFolderName = Sent;
  SOGoTrashFolderName = Trash;
  SOGoJunkFolderName = Junk;
  SOGoIMAPServer = "localhost";
  SOGoSMTPServer = "smtp://localhost";
  SOGoMailDomain = seudominio.com.br;
  SOGoMailingMechanism = smtp;

  /* Authentication (usando a view criada no banco do SOGo com os dados do Go-PostfixAdmin) */
  SOGoUserSources = (
    {
      type = sql;
      id = directory;
      viewURL = "mysql://sogo:SUA_SENHA_SEGURA_AQUI@localhost:3306/sogo/sogo_users_view";
      canAuthenticate = YES;
      isAddressBook = YES;
      /* Deve bater com o formato usado no Go-Postfixadmin (Dovecot) ex: blf-crypt (Bcrypt) */
      userPasswordAlgorithm = blf-crypt; 
    }
  );

  /* Web Interface */
  SOGoPageTitle = "SOGo Webmail";
  SOGoVacationEnabled = YES;
  SOGoForwardEnabled = YES;
  SOGoSieveScriptsEnabled = YES;
  SOGoMailAuxiliaryUserAccountsEnabled = YES;

  /* General */
  SOGoLanguage = Portuguese;
  SOGoTimeZone = America/Sao_Paulo;
  SOGoSuperUsernames = ("admin@seudominio.com.br");
  SOGoMemcachedHost = "127.0.0.1";

  /* Workers */
  WOWorkersCount = 3;
}
```

*Atenção: Substitua `SUA_SENHA_SEGURA_AQUI`, `SENHA_DO_USUARIO_POSTFIX`, e `seudominio.com.br` pelos valores reais do seu ambiente.*

Após configurar, reinicie os serviços:
```bash
sudo systemctl restart memcached sogo
sudo systemctl enable memcached sogo
```

---

## 5. Instalação e Configuração do Nginx

Instale o Nginx se ainda não estiver instalado:
```bash
sudo apt install nginx -y
```

Crie o arquivo de configuração para o SOGo no Nginx:
```bash
sudo nano /etc/nginx/sites-available/sogo.conf
```

Exemplo básico de configuração de Virtual Host (Proxy Reverso) com SSL:

```nginx
server {
    listen 80;
    server_name mail.seudominio.com.br;
    
    # Redirecionar HTTP para HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name mail.seudominio.com.br;

    # Certificados SSL (Exemplo via Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/mail.seudominio.com.br/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mail.seudominio.com.br/privkey.pem;

    root /usr/lib/GNUstep/SOGo/WebServerResources/;

    location = / {
        rewrite ^ https://$server_name/SOGo;
        allow all;
    }

    # Proxy para o SOGo
    location ^~ /SOGo {
        proxy_pass http://127.0.0.1:20000;
        proxy_redirect http://127.0.0.1:20000 default;
        
        # Headers necessários
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $host;
        proxy_set_header x-webobjects-server-protocol HTTP/1.0;
        proxy_set_header x-webobjects-remote-host 127.0.0.1;
        proxy_set_header x-webobjects-server-name $server_name;
        proxy_set_header x-webobjects-server-url https://$server_name;
        proxy_set_header x-webobjects-server-port 443;
        
        # Ajustes de limite de envios
        client_max_body_size 50M;
        client_body_buffer_size 128k;
        break;
    }

    # Arquivos estáticos SOGo
    location /SOGo.woa/WebServerResources/ {
        alias /usr/lib/GNUstep/SOGo/WebServerResources/;
        allow all;
        expires max;
    }

    location /SOGo/WebServerResources/ {
        alias /usr/lib/GNUstep/SOGo/WebServerResources/;
        allow all;
        expires max;
    }

    # ActiveSync (opcional, para clientes mobile/Outlook)
    location ^~ /Microsoft-Server-ActiveSync {
        proxy_pass http://127.0.0.1:20000/SOGo/Microsoft-Server-ActiveSync;
        proxy_redirect http://127.0.0.1:20000/Microsoft-Server-ActiveSync /;
        
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $server_name;
        proxy_set_header x-webobjects-server-protocol HTTP/1.0;
        proxy_set_header x-webobjects-remote-host 127.0.0.1;
        proxy_set_header x-webobjects-server-name $server_name;
        proxy_set_header x-webobjects-server-url https://$server_name;
        proxy_set_header x-webobjects-server-port 443;
        
        proxy_connect_timeout 75;
        proxy_send_timeout 3600;
        proxy_read_timeout 3600;
        proxy_buffer_size 128k;
        proxy_buffers 64 256k;
        proxy_busy_buffers_size 256k;
        proxy_temp_file_write_size 256k;
        client_max_body_size 0;
        client_body_buffer_size 128k;
    }
}
```

Habilite o site no Nginx configurando o link simbólico e recarregando o serviço:
```bash
sudo ln -s /etc/nginx/sites-available/sogo.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 6. Testando o Acesso

1. No seu navegador, acesso `https://mail.seudominio.com.br`.
2. A tela de login do SOGo deverá aparecer.
3. Entre com um endereço completo criado no seu **Go-PostfixAdmin** (ex: `usuario@seudominio.com.br`) e com a respectiva senha.

### Dicas Adicionais

- **Logs do SOGo**: Verifique `/var/log/sogo/sogo.log` caso ocorram erros 502 ou falhas de login que pareçam misteriosas.
- **Integração Go-PostfixAdmin**: Qualquer alteração de senha, criação de nova caixa ou suspensão feita pelo *Go-PostfixAdmin* vai automaticamente impactar o acesso ao SOGo, já que utilizamos o mesmo banco de dados MariaDB (`postfix`) para as credenciais.
