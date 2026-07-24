# Design: admin-panel-zimbra

## Context

Ecossistema "Zimbra em Go": webmail (go-snappymail, clona Zimbra Web) + admin (este, clona ZimbraAdmin). O go-postfixadmin já é um binário Go único com Vue 3 embutido, GORM (MariaDB/PostgreSQL), JWT/RBAC e handlers de domain/mailbox/alias/admin/transport/fetchmail/rbac. Referência viva do ZimbraAdmin: console legado em `https://192.168.56.30:7071/zimbraAdmin/` (VM `vagrant/zimbra`, snapshot `zimbra-installed`).

Diretriz do dono: **um único binário fazendo tudo**, instalação simples.

## Goals / Non-Goals

**Goals**
- Painel admin com layout idêntico ao ZimbraAdmin Classic, motor Vue 3, em `:7071` (http/https).
- Binário único que serve webmail + admin, ligados por config, sem interferência mútua.
- Backend admin sobre GORM (MariaDB/PostgreSQL) reaproveitando o go-postfixadmin.

**Non-Goals** — ver proposal (sem DWT, sem logo, sem monitor avançado/zimlets/migração nesta fase).

## Decisions

### D1 — Binário host e consolidação
**Decisão:** o **go-postfixadmin** é o binário host (já tem GORM/RBAC/handlers `*V1` maduros, `dto.APIResponse`, Swagger). O webmail (go-snappymail) entra como serviço adicional.
- Alternativa A (rejeitada): novo repo "umbrella" importando os dois — mais setup, duplica CI/release.
- Alternativa B (rejeitada): manter dois binários — contraria o objetivo de instalação simples.
- **Contrato mínimo, decisão de acoplamento adiada mas explícita:** o webmail é IMAP/SMTP passthrough e sua SPA é buildada independente. **Importante (Go):** pacotes `internal/` de outro módulo **não podem** ser importados via `go.mod require` — então as opções válidas de consolidar (decidir **no início** da implementação) são:
  - **(A) reverse-proxy** no host para um go-snappymail rodando embutido no mesmo processo como binário/goroutine, ou como serviço à parte — sem import de código; mais simples e desacoplado.
  - **(B) mover o webmail para dentro da árvore do módulo host** (subdiretório/submódulo) para poder importar seus pacotes, ou o go-snappymail **expor uma API pública** (`pkg/`) com o handler mail montável.
  - **(C) embed apenas da SPA** buildada do webmail + reimplementar o passthrough mínimo no host (evita import de código Go).
  - Preferência inicial: **(A) reverse-proxy** (menor acoplamento, respeita a fronteira `internal/`); reavaliar na implementação.
  - Hoje `main.go` embute só um `web/dist` — a SPA admin exige **um segundo `embed.FS`** (o frontend atual do go-postfixadmin é "neo-brutalism"; o admin ZimbraAdmin é rota/skin nova, não substitui a UI existente).

### D2 — Config multi-serviço (liga/desliga por bloco)
**Compatibilidade:** o repo hoje usa o subcomando `server` e o bloco `[server]` (uma porta, um Echo). Os blocos `[admin]`/`[webmail]` abaixo são **aditivos**: `[admin]` herda o comportamento do `[server]` atual (mesmo binário, mesma porta default se `[webmail]` ausente), e o subcomando continua `server` (não renomear para `serve`). Migração de config documentada; `[server]` legado mapeia para `[admin]`.

```toml
[admin]        # painel ZimbraAdmin
enabled = true
port    = 7071
tls     = false          # true → https com cert/key
skin    = "zimbra"       # layout ZimbraAdmin

[webmail]      # go-snappymail
enabled = true
port    = 8082

[database]     # GORM (compartilhado pelo admin)
driver = "mysql"         # mysql | postgres (ConnectDB)
dsn    = "..."
```
O subcomando `server` inicia todos os blocos `enabled=true`; cada um em sua porta/roteador. Rodar só o admin = `webmail.enabled=false`.

### D3 — Skin `zimbra` do admin (paridade)
Mesma abordagem validada no webmail: tokens de tema + layout clonado, medido do console real. Estrutura de tela do ZimbraAdmin:
- **Top bar** azul "Zimbra Administration"-like (marca texto), busca central com dropdown de tipo, `admin@… ▾`, refresh.
- **Árvore de navegação** (esquerda): Home / Monitor / Manage (Accounts, Aliases, Distribution Lists, Resources, Domains) / Configure (Class of Service, Servers, Global Settings, Zimlets) / Tools & Migration / Search.
- **Home**: "Overview" (Version, Servers, Accounts, Domains, COS) + "Runtime" (Service status, Active sessions, Queue) + cards 1-Iniciado / 2-Configurar domínio / 3-Adicionar contas.
- **Content pane**: toolbar (New/Edit/Delete/…) + list view (colunas, seleção, paginação, outline de seleção).
- Cantos 3px, paleta harmony, tipografia Helvetica/Arial — reaproveitar tokens do webmail.

### D4 — Backend admin (telas ↔ endpoints REAIS)
Reusar os handlers `*V1` existentes sob o prefixo **`/api/v1/*`** com o envelope **`dto.APIResponse`** (Swagger `@BasePath /api/v1`) — **não** criar um namespace `/api/admin/*` paralelo. Adicionar só o agregador da Home.

| Tela | Endpoint real (go-postfixadmin) |
|---|---|
| Home overview | **novo** `GET /api/v1/admin/overview` (agrega counts do banco) |
| Domains | `/api/v1/domains*` |
| Accounts (mailboxes) | `/api/v1/mailboxes*` |
| Aliases | `/api/v1/aliases*` |
| Distribution Lists | ⚠ alias multi-destino ≠ lista de distribuição — modelar depois; nesta fase, listar como aliases com N destinos |
| Admins | `/api/v1/admins*` |
| Transports | `/api/v1/transports*` |

**Overview realista** — só o que o schema PostfixAdmin tem via GORM: `accounts` (mailboxes), `domains`, `aliases`, `admins`. Campos do console legado sem fonte no schema (**COS, servers, active sessions, mail queue**) ficam como stub `—`/"n/a" ou ocultos nesta fase (ver Non-goals). Nada de prometer dado inexistente.

Persistência: GORM → **MariaDB ou PostgreSQL** (`internal/database.ConnectDB` suporta `mysql`/`postgres`). **SQLite não é suportado hoje** — para unit tests usar mock/sqlmock ou MariaDB/Postgres em container; se SQLite for desejável em dev, é uma task à parte (adicionar driver ao ConnectDB).

### D6 — RBAC real (não "papel admin" genérico)
O backend já tem RBAC granular: papéis `superadmin` / `domain_admin` / … e permissões (`domains:read`, `mailboxes:write`, …). A árvore de navegação SHALL respeitar isso: um `domain_admin` vê Manage (dos seus domínios) mas **não** Configure (Servers/Global Settings); nós fora da permissão ficam ocultos ou retornam 403. Mapear cada nó da árvore → permissão exigida.

### D5 — Multi-listener num processo (isolamento real)
O `internal/server.StartServer` atual sobe **um** Echo, **uma** porta, **um** `web/dist` embutido. O binário único exige:
- **Dois listeners Echo** (ou um Echo com dois vhosts) — admin `:7071`, webmail `:8082` — cada um com **seu próprio** middleware (CSRF, rate-limit, auth), **seu** FS de SPA embutido (dois `embed.FS` ou dois `web/dist`), e **graceful shutdown** coordenado (errgroup + signal).
- SPAs em base paths próprios: `/` (webmail) e `/zimbraAdmin/` ou `/admin/` (admin — espelhar o path legado ajuda o clone).
- **Isolamento de cookies/CSRF (crítico):** cookies **não** são isolados por porta no mesmo host — admin e webmail no mesmo `localhost` compartilham o jar. Portanto: **nomes distintos** (`gsn_session` webmail vs `gpa_admin_jwt` admin), **Path** próprio (`/` vs `/admin/`), `SameSite`/`Secure` explícitos, e **CSRF por serviço**. Sem isso, sessão do webmail e JWT do admin colidem.
- Sessão/auth independentes: webmail usa cookie IMAP-session; admin usa JWT/RBAC do banco.
- Sem estado global compartilhado além do processo e do pool GORM (só o admin usa banco).
- Config `enabled=false` num bloco → aquele listener nem abre.

## Risks / Trade-offs

- [Acoplar dois repos num binário] → manter go-snappymail como módulo/SPA versionada; contrato estável (SPA + rotas), não código espalhado.
- [Paridade visual do ZimbraAdmin é grande] → fatiar por tela (Home → Domains → Accounts → …), validar cada uma com o QA agent contra `:7071`, como no webmail.
- [Divergência DWT vs Vue] → clonar look/fluxos, não o toolkit; documentar divergências intencionais.
- [Um binário = deploy acoplado] → aceitável; blocos `enabled` permitem rodar só um serviço quando necessário.

## Migration Plan

Fatiado: (1) config multi-serviço + host servindo webmail+admin vazio; (2) skin/layout ZimbraAdmin (top bar, árvore, Home); (3) telas Domains/Accounts/Aliases sobre endpoints existentes; (4) overview endpoint; (5) validação QA por tela. Rollback = desabilitar `[admin]` no config.

## Infra a criar (gap com o baseline do repo — apontado na validação)

- **`.golangci.yml`** na raiz (não existe) + rodar `golangci-lint` no CI.
- **ESLint + Prettier** no `frontend/` (hoje só `vue-tsc`); typecheck já existe.
- **Workflow CI** de test+lint a cada push (hoje só `release.yml` em tag).
- Ambiente de teste de banco: MariaDB **e** PostgreSQL em container para a matrix (sqlmock para unit).

## Open Questions

- D1: embed da SPA + import do pacote mail do webmail **vs** reverse-proxy — decidir no início da implementação, medindo conflito de dependências no `go.mod`.
- Base path do admin: `/zimbraAdmin/` (espelha o legado) vs `/admin/` — definir na fase de skin.
- i18n: a UI admin é **inglês fixo** (requisito), enquanto o backend tem i18n `.po`; o admin novo fixa `en` (não usa o middleware de locale) — confirmar que não conflita com as respostas de erro localizadas do backend.
- Distribution Lists e COS/servers/queue: modelar (novo schema) ou manter fora de escopo nesta fase.

## Validação cruzada (CLIs externas)

Proposta revisada por múltiplas CLIs de IA (codex, cursor/agent) contra o código real do repo. Achados incorporados: rotas `/api/v1/*` + envelope `dto.APIResponse` (não `/api/admin/*`); import de `internal/` entre módulos Go é inviável → reverse-proxy/embed/submódulo (D1); servidor mono-listener atual precisa virar multi-listener com middleware/SPA/shutdown por serviço (D5); cookies não isolam por porta → nomes/path/SameSite/CSRF por serviço; RBAC granular real (superadmin/domain_admin + permissões) e nós sem permissão (servers/cos/queue) como stub; SQLite não suportado por ConnectDB; falta `.golangci.yml`/ESLint/Prettier/CI de test+lint (infra a criar). Divergências residuais (opencode/agy sem saída útil) não bloqueiam.
