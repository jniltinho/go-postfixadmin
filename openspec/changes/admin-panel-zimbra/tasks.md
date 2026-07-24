# Tasks: admin-panel-zimbra

## 1. Binário único / config multi-serviço

- [ ] 1.1 Blocos `[admin]`, `[webmail]`, `[database]` no config (porta, tls, enabled, driver/dsn)
- [ ] 1.2 Subcomando `server` inicia todos os blocos `enabled=true`, cada um em sua porta/roteador (compat: `[server]` legado → `[admin]`)
- [ ] 1.3 Decidir D1 (import de código do webmail vs. embed da SPA buildada) e integrar go-snappymail no host
- [ ] 1.4 Verificar isolamento: admin `:7071` e webmail `:8082` sem colisão

## 2. Backend admin (GORM → MariaDB/PostgreSQL)

- [ ] 2.1 `GET /api/v1/admin/overview` (counts accounts/domains/aliases/admins que existem no schema; COS/servers/queue/sessions = n/a), permissão admin
- [ ] 2.2 Mapear telas ↔ endpoints existentes (domains/mailboxes/aliases/admins/transports)
- [ ] 2.3 Confirmar JWT/RBAC exigindo papel admin nas rotas do painel; 403 sem papel
- [ ] 2.4 Migração/seed de dev (MariaDB Docker `:3306` ou PostgreSQL; ConnectDB não tem SQLite)

## 3. Skin/layout ZimbraAdmin (Vue 3)

- [ ] 3.1 Tokens de tema `zimbra` do admin (paleta harmony, cantos 3px, tipografia)
- [ ] 3.2 Top bar (marca textual, busca, `admin@… ▾`, refresh)
- [ ] 3.3 Árvore de navegação (Home/Monitor/Manage/Configure/Tools/Search)
- [ ] 3.4 Home (Overview + Runtime + cards de setup) ligada ao `/api/v1/admin/overview`
- [ ] 3.5 List view + toolbar (colunas, seleção com outline, paginação)

## 4. Telas de gerenciamento

- [ ] 4.1 Domains (list + New/Edit/Delete)
- [ ] 4.2 Accounts/Mailboxes (list + CRUD)
- [ ] 4.3 Aliases + Distribution Lists
- [ ] 4.4 Admins (list + papéis RBAC)

## 5. Testes de backend (obrigatório — backend first)

- [ ] 5.1 Testes do `GET /api/v1/admin/overview` (agregados corretos; papel admin exigido; 403 sem papel) — table-driven, `-race`
- [ ] 5.2 Testes dos handlers reusados nas telas: domains/mailboxes/aliases/admins/transports (CRUD, validação, erros) — cobrir casos de sucesso e falha
- [ ] 5.3 Testes de auth/RBAC do painel (JWT válido/expirado/sem papel → 401/403)
- [ ] 5.4 Testes de config multi-serviço (blocos enabled/disabled → portas certas; parsing TOML/env)
- [ ] 5.5 Testes de persistência GORM em MariaDB **e** PostgreSQL (matrix; sqlmock para unit — ConnectDB não tem SQLite), migrações aplicam limpo
- [ ] 5.6 Cobertura mínima acordada e `go test -race ./...` verde no CI

## 6. Lint / qualidade (obrigatório)

- [ ] 6.1 `golangci-lint run` limpo no backend (govet, staticcheck, errcheck, gosec, revive…); config `.golangci.yml`
- [ ] 6.2 `go vet ./...` e `gofmt`/`goimports` sem diffs
- [ ] 6.3 Lint do frontend admin: `eslint` + `vue-tsc --noEmit` (typecheck) sem erros; `prettier` consistente
- [ ] 6.4 CI roda test + lint (backend e frontend) a cada push; falha bloqueia merge

## 7. Validação e docs

- [ ] 7.1 Auditar cada tela com `qa-frontend-cloner` contra `:7071` até 0 P1 (prints em `docs/prints/`)
- [ ] 7.2 Doc de arquitetura do binário único + mapeamento telas↔endpoints
- [ ] 7.3 Atualizar guia de dev (subir binário único: webmail + admin + Zimbra por snapshot)
