# Done — migration-vue3-jwt

## Backend JWT + API REST ✅

- [x] `golang-jwt/jwt/v5` integrado
- [x] `internal/auth/jwt.go` — geração e validação de tokens
- [x] Middleware JWT com dual-mode (session legado + JWT novo durante transição)
- [x] `POST /api/v1/auth/login` — auto-detecta admin vs mailbox, retorna access token + cookie refresh httpOnly
- [x] `POST /api/v1/auth/refresh` — rotação de refresh token via cookie httpOnly
- [x] `POST /api/v1/auth/logout` — limpa cookie
- [x] `GET  /api/v1/auth/me` — identidade + claims do token atual
- [x] Rate limiting no login (por IP + por username)
- [x] DTOs em `internal/api/dto/` — request/response tipados
- [x] Formato padronizado de erro (`{ success, error: { code, message, details } }`)
- [x] CRUD completo: domains, mailboxes, aliases, alias-domains, admins, transports
- [x] User self-service: `/user/me`, `/user/forwarding`, `/user/password`, `/user/vacation`
- [x] API Keys: `/settings/apikeys`
- [x] Dashboard stats, logs paginados, maillog paginado
- [x] Swagger/OpenAPI embutido
- [x] CSRF middleware + rate limiting global
- [x] `LogAction` preservado em todas as mutações

## Frontend Vue 3 SPA ✅

- [x] Vue 3 + TypeScript + Vite + Tailwind CSS v4 + Lucide Icons
- [x] Pinia: auth store, estado de autenticação
- [x] Axios com interceptors: Bearer header automático + refresh em 401
- [x] Vue Router com history mode + route guards
- [x] Design system neo-brutalist fiel: cores, bordas 2-4px, sombras hard, tipografia Fira
- [x] `MainLayout.vue`: sidebar 240px (CSS grid) + header 56px
- [x] LoginPage com dual-portal (admin + mailbox)
- [x] Todas as páginas CRUD: Domains, Mailboxes, Aliases, AliasDomains, Admins, Transports
- [x] Logs + Maillog paginados
- [x] API Keys
- [x] User portal: forwarding, senha, vacation
- [x] `AppTable.vue` — componente de tabela reutilizável
- [x] `vue3-toastify` para notificações

## Build + Embedding ✅

- [x] Dockerfile multi-stage: Node builder → Go builder → Alpine final
- [x] Makefile: `make frontend-build` integrado ao `make build`
- [x] `//go:embed web/dist locales` no binário final
- [x] `SPAFileServer` em `internal/server/` — serve assets estáticos + fallback `index.html`
- [x] Rota `GET /*` registrada por último (após `/api/v1/*`)

## Desvio do Plano Original ✅

- [x] **Quasar substituído por Tailwind CSS + Lucide** — Quasar gerava conflito constante com o design neo-brutalist; Tailwind v4 + classes custom resultaram em bundle menor e maior fidelidade visual sem luta contra Material Design defaults
