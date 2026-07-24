# Spec: admin-backend-api

Contrato REST do painel admin sobre o backend go-postfixadmin (GORM → MariaDB/PostgreSQL), auth JWT/RBAC. Reaproveita endpoints existentes; adiciona o agregador da Home.

## ADDED Requirements

### Requirement: Config multi-serviço num binário
O binário SHALL ler blocos `[admin]`, `[webmail]` e `[database]` do `config.toml` e iniciar por um único `serve` todos os serviços com `enabled=true`, cada um em sua porta.

#### Scenario: Só admin
- **WHEN** `[admin] enabled=true` e `[webmail] enabled=false`
- **THEN** o processo serve apenas `:7071` e não abre a porta do webmail

#### Scenario: Ambos
- **WHEN** ambos `enabled=true`
- **THEN** um processo serve `:7071` (admin) e `:8082` (webmail) sem interferência

### Requirement: Overview endpoint
O backend SHALL expor `GET /api/v1/admin/overview` (prefixo `/api/v1`, envelope `dto.APIResponse`) retornando os contadores **que existem no schema** — accounts (mailboxes), domains, aliases, admins — para a Home, exigindo permissão de admin. Campos do console legado sem fonte no schema (COS, servers, active sessions, mail queue) SHALL ser omitidos ou retornados como `null`/`"n/a"`, nunca inventados.

#### Scenario: Agregados reais
- **WHEN** um admin autenticado chama `/api/v1/admin/overview`
- **THEN** recebe `dto.APIResponse` com counts de accounts/domains/aliases/admins derivados via GORM; campos sem fonte vêm nulos/"n/a"

### Requirement: CRUD via endpoints v1 existentes
As telas SHALL operar sobre os endpoints já existentes sob `/api/v1/*` (envelope `dto.APIResponse`, handlers `*V1`): domains, mailboxes (accounts), aliases, admins, transports — sob JWT/RBAC. **Não** criar namespace `/api/admin/*` paralelo.

#### Scenario: Criar domínio
- **WHEN** o admin cria um domínio no painel via `/api/v1/domains`
- **THEN** o backend persiste via GORM e a lista reflete o novo domínio no envelope padrão

### Requirement: Persistência GORM MariaDB/PostgreSQL
O admin SHALL persistir em MariaDB ou PostgreSQL via GORM, selecionável por `[database] driver` (`mysql`/`postgres` — os que `ConnectDB` suporta hoje). Unit tests usam sqlmock ou container; SQLite-para-dev, se desejado, é task à parte (adicionar driver).

#### Scenario: Troca de driver
- **WHEN** `[database] driver` muda entre `mysql` e `postgres` com DSN válido
- **THEN** o painel funciona sem alteração de código (só config/migração)

### Requirement: Auth admin (JWT/RBAC granular real)
O painel SHALL usar o RBAC existente com papéis reais (`superadmin`/`domain_admin`/…) e permissões granulares (`domains:read`, `mailboxes:write`, …). Cada rota/nó exige a permissão correspondente; sem ela → 403. Um `domain_admin` acessa Manage dos seus domínios mas não Configure (Servers/Global Settings).

#### Scenario: Acesso negado por permissão
- **WHEN** um `domain_admin` chama uma rota que exige `servers:read` (Global Settings)
- **THEN** o backend responde 403 e o nó correspondente fica oculto na árvore
