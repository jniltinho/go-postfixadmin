# Proposal: admin-panel-zimbra

## Why

O ecossistema "Zimbra em Go" recria o Zimbra em Go + Vue 3, leve e rápido. O webmail (go-snappymail) já clona o Zimbra Web Classic. Falta o **painel administrativo**: um clone do **ZimbraAdmin** (console legado em `:7071`) para gerenciar domínios, contas, aliases, listas e configurações. O go-postfixadmin já tem o backend maduro (API Go, GORM, JWT/RBAC, handlers de domain/mailbox/alias/admin/transport/fetchmail) — reaproveitá-lo como base evita reescrever a camada de dados.

**Objetivo desta fase: um único binário fazendo tudo** — webmail + painel admin — para **instalação simples** (baixar um binário, um `config.toml`, subir). O binário embute as duas SPAs Vue 3 (`go:embed`), serve o webmail (IMAP/SMTP passthrough) e o admin (GORM → MariaDB/PostgreSQL) e liga cada serviço por config. A cara é idêntica ao ZimbraAdmin, mas o motor é Vue 3 (bem mais leve que o DWT/AJAX legado).

## What Changes

- **Nova skin/layout de admin `zimbra` no frontend Vue do go-postfixadmin**, clonando o ZimbraAdmin Classic:
  - Barra azul superior "Zimbra Administration"-like (marca em texto, sem logo Zimbra), busca central, menu de usuário `admin@… ▾`, refresh.
  - Árvore de navegação à esquerda: **Home / Monitor / Manage** (Accounts, Aliases, Distribution Lists, Resources, Domains) **/ Configure** (Class of Service, Servers, Global Settings, Zimlets) **/ Tools & Migration / Search**.
  - Home: painel "Overview" (Version, Servers, Accounts, Domains, COS) + "Runtime" (Service status, Active sessions, Queue) + cards de setup, medidos do console real.
  - Content pane: toolbar (New / Edit / Delete / …) + list view estilo Zimbra (colunas, seleção, paginação).
  - Cantos suaves 3px, paleta e tipografia do ZimbraAdmin (mesma família harmony do webmail).
- **Binário único servindo múltiplos serviços por config** — sem processos/instalações separadas:
  - `[webmail]` (IMAP/SMTP passthrough, SPA go-snappymail) — porta padrão `8082`.
  - `[admin]` (GORM → MariaDB/PostgreSQL, SPA ZimbraAdmin-like) — porta `7071` (http/https).
  - Cada serviço liga/desliga por flag no `config.toml`; um só `serve` sobe todos os habilitados.
  - Zero colisão: portas distintas, roteadores/SPAs isolados, mesmo processo.
- **Backend admin reaproveitado e estendido** (base go-postfixadmin, GORM → MariaDB/PostgreSQL):
  - Mapear os endpoints existentes (domains, mailboxes, aliases, admins, transports, RBAC) para as telas do painel.
  - Endpoint de "overview/runtime" agregando contagens (domínios, contas, COS) para a Home.
  - Auth JWT/RBAC já existente reutilizada; o painel admin exige papel admin.
- **Consolidação do ecossistema num binário** (ver design.md): definir o repo/módulo host, importar o webmail e o admin como pacotes internos, um `main` com subcomandos (`serve`, `migrate`, `init`), cada peça em seu diretório sem interferência.
- **Documentação**: arquitetura do binário único, guia de dev (subir tudo + Zimbra por snapshot), e mapeamento telas↔endpoints.

## Capabilities

### New Capabilities

- `admin-panel-ui`: o layout ZimbraAdmin Classic em Vue 3 (barra, árvore de navegação, Home overview/runtime, list views, toolbars), servido em `:7071`, com paridade visual contra o console real (`192.168.56.30:7071`).
- `admin-backend-api`: contrato REST do painel (overview/runtime + CRUD de domains/accounts/aliases/lists/COS) sobre o backend go-postfixadmin (GORM → MariaDB/PostgreSQL), com auth JWT/RBAC.

### Modified Capabilities

<!-- nenhuma — não há specs principais extraídas ainda no go-postfixadmin -->

## Non-goals

- Não recriar o toolkit DWT/AJAX do Zimbra — só o **look** em Vue.
- Sem logo/marca registrada do Zimbra (branding em texto).
- Fora do escopo desta fase: Monitor avançado (gráficos de servidor), Zimlets, Class of Service completo (apenas leitura/listagem inicial), migração, backup/restore, certificados.
- Não reescrever o backend existente do go-postfixadmin; **estender**, não reescrever.
- A consolidação num binário não muda o comportamento de cada serviço — só o empacotamento/instalação. Rodar serviços em processos separados continua possível (habilitar só um bloco no config).

## Impact

- **go-postfixadmin** (repo do backend/admin):
  - **`frontend-admin/` (NOVO diretório irmão)** — painel ZimbraAdmin em Vue 3, **totalmente separado** do `frontend/` atual (deps, build e embed próprios; ver design D7). O `frontend/` (neo-brutalism) **não é alterado**.
  - `web/admin-dist/` (NOVO) — saída de build do painel admin, embutida via `//go:embed` ao lado de `web/dist`.
  - `main.go` — adicionar `web/admin-dist` ao embed; `Makefile` — build dos dois frontends.
  - `internal/handlers/` + `internal/routes/` — endpoint `GET /api/v1/admin/overview`; binding telas↔endpoints `/api/v1/*`; listener/porta `:7071` (http/https) no `internal/server` (multi-listener).
  - `config.toml` — bloco `[admin]` (porta 7071, tls opcional) + `[webmail]`; driver de banco em `[database]`.
  - `openspec/` — esta change (fonte de verdade do roadmap admin).
- **go-snappymail** (este ecossistema): `AGENTS.md` + `docs/dev-environment.md` já referenciam a fase admin; `docs/architecture.md` recebe o diagrama do ecossistema.
- Banco: GORM sobre MariaDB (lab Docker `:3306`) ou PostgreSQL; sem novo schema além do já existente do go-postfixadmin + tabela/consulta de agregados para a Home.
