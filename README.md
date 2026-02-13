# Go-Postfixadmin

Professional Email Administration System built with Go, Echo, and Tailwind CSS.

## 🛠 Ferramentas de Desenvolvimento

Para compilar o projeto localmente (sem Docker), você precisará instalar as seguintes ferramentas:

1.  **Go (v1.25.4 ou superior)**: Linguagem principal do projeto.
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
./postfixadmin --run --port 8080
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

# Para uso com --import-sql (requer multiStatements=true)
DATABASE_URL="user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true"
```

**PostgreSQL:**
```bash
DATABASE_URL="host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
```

---

## � Flags da CLI

Abaixo estão as flags disponíveis ao executar o binário `./postfixadmin`:

```text
Usage of ./postfixadmin:
  -db-driver string
        Database driver (mysql or postgres) (default "mysql")
  -db-url string
        Database URL connection string
  -import-sql string
        Import SQL file to database
  -migrate
        Run database migration
  -port int
        Port to run the server on (default 8080)
  -run
        Start the administration server
  -version
        Display version information
```

---

## �📝 Comandos úteis do Makefile

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
