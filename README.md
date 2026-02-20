# Go-Postfixadmin


Professional Email Administration System built with Go, Echo, and Tailwind CSS.

## ✨ Funcionalidades

*   **Gerenciamento Completo**: Domínios, Caixas de Correio (Mailboxes) e Aliases.
*   **Controle de Acesso (RBAC)**: Diferenciação entre Superadmin e Administradores de Domínio.
*   **Design Moderno**: Interface responsiva e limpa construída com Tailwind CSS.
*   **Segurança**: Hash de senhas forte e proteção contra ataques comuns.
*   **CLI Integrada**: Ferramentas de linha de comando para automação e recuperação de acesso.


## 🛠 Ferramentas de Desenvolvimento

Para compilar o projeto localmente (sem Docker), você precisará instalar as seguintes ferramentas:

1.  **Go (v1.26 ou superior)**: Linguagem principal do projeto.
    *   [Download Go](https://go.dev/dl/)
2.  **Node.js (v20 ou superior)**: Necessário para o processamento do CSS com Tailwind.
    *   [Download Node.js](https://nodejs.org/)
3.  **Make**: Utilitário para automação de comandos (nativo no Linux/macOS).
4.  **UPX (Opcional)**: Utilizado pelo Makefile para compactar o binário final.
    *   `sudo apt install upx-ucl` (Debian/Ubuntu)

---

## 🏗 Como fazer o Build

Este projeto oferece duas formas principais de build: utilizando `make` (local) ou `docker`.

### 1. Build nativo com Makefile

O build local automatiza a geração do CSS e a compilação do binário Go.

#### Instalação de Dependências

Para instalar todas as dependências (Recomendado):

```bash
make deps
```

Caso prefira instalar manualmente:

```bash
go mod download
npm install
```

### Compilação
```bash
# Gerar CSS e compilar o binário
make build

# Para limpar os arquivos gerados
make clean
```

### 2. Build com Docker

Ideal para gerar uma versão final isolada e pronta para produção sem precisar instalar Go ou Node.js na sua máquina.

**Requisitos:** Docker instalado.

```bash
# Gerar a imagem docker profissional (otimizada para ~14MB)
make build-docker
```

Este comando executa um build multi-stage que:
1.  Compila os assets estáticos (Tailwind).
2.  Compila o binário Go (Gera um binário estático).
3.  Compacta o binário com `upx`.
4.  Gera uma imagem final baseada em Alpine Linux.

---

## 🚀 Execução

Após o build, você pode rodar o binário diretamente:

```bash
./postfixadmin server --port=8080
```

Ou via Docker:

```bash
docker run -p 8080:8080 -e DATABASE_URL="seu-dsn" postfixadmin:latest
```

### Exemplos de DATABASE_URL

**MySQL:**
```bash
# Formato padrão
DATABASE_URL="user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

# Para uso com importsql (requer multiStatements=true)
DATABASE_URL="user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true"
```

**PostgreSQL:**
```bash
DATABASE_URL="host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
```

### 3. Deploy com Systemd (Linux)

Para implantar a aplicação de forma nativa em um servidor Linux, você pode utilizar o arquivo de serviço do Systemd incluído no projeto.

O arquivo pré-configurado está localizado em `DOCUMENTS/setup/postfixadmin.service`. Ele espera que a aplicação esteja alocada no diretório `/opt/go-postfixadmin` e lerá as variáveis de ambiente de um arquivo `.env` neste mesmo diretório.

**Instalação do Serviço:**

```bash
# 1. Copie o arquivo para o diretório de serviços do systemd
sudo cp DOCUMENTS/setup/postfixadmin.service /etc/systemd/system/

# 2. Recarregue as configurações do systemd
sudo systemctl daemon-reload

# 3. Ative o serviço para rodar junto com o boot do sistema
sudo systemctl enable postfixadmin.service

# 4. Inicie o serviço
sudo systemctl start postfixadmin.service

# 5. Acompanhe os logs em tempo real
# O serviço direciona a saída para o arquivo postfixadmin.log
tail -f /opt/go-postfixadmin/postfixadmin.log
```

---

## 📝 Flags da CLI

Abaixo estão as flags disponíveis ao executar o binário `./postfixadmin`:

```text
A command line interface for Go-Postfixadmin application.

Usage:
  postfixadmin [command]

Available Commands:
  admin       Admin management utilities
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  importsql   Import SQL file to database
  migrate     Run database migration
  server      Start the administration server
  version     Display version information

Flags:
      --db-driver string   Database driver (mysql or postgres) (default "mysql")
      --db-url string      Database URL connection string
  -h, --help               help for postfixadmin

Use "postfixadmin [command] --help" for more information about a command.
```

### Comandos de Administração (CLI)

O binário também suporta comandos administrativos diretos via subcomando `admin`:

```bash
# Listar todos os administradores
./postfixadmin admin --list-admins

# Listar todos os domínios
./postfixadmin admin --list-domains

# Criar um novo Superadmin (útil para primeiro acesso)
./postfixadmin admin --add-superadmin "admin@example.com:password123"
# Ou deixe a senha em branco para gerar uma aleatória
./postfixadmin admin --add-superadmin "admin@example.com"
```

Outras flags disponíveis para `admin`:
*   `--list-mailboxes`: Listar todas as caixas de correio.
*   `--list-aliases`: Listar todos os aliases.
*   `--domain-admins`: Listar administradores de domínio.


---

## 📝 Comandos úteis do Makefile

| Comando | Descrição |
| :--- | :--- |
| `make build` | Compila o CSS e o binário localmente |
| `make build-docker` | Gera a imagem Docker otimizada |
| `make run` | Compila e inicia o servidor localmente |
| `make watch-css` | Inicia o watcher do Tailwind para desenvolvimento UI |
| `make clean` | Remove o binário e arquivos de CSS gerados |
| `make tidy` | Limpa e organiza as dependências do Go |
| `make deps` | Instala todas as dependências necessárias |

---

## 📸 Screenshots

![Go-Postfixadmin Login Screen](DOCUMENTS/screenshots/postfixadmin_01.png)

Confira mais imagens na pasta [screenshots](DOCUMENTS/screenshots).
