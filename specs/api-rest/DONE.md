# Done — api-rest

## Auth ✅

- [x] `POST /auth/login` — login admin e mailbox, retorna JWT + cookie refresh
- [x] `POST /auth/refresh` — renova access token via cookie httpOnly
- [x] `POST /auth/logout` — invalida sessão
- [x] `GET  /auth/me` — dados do usuário autenticado

## Recursos Admin ✅

- [x] Domains — CRUD completo (`/domains`, `/domains/:domain`)
- [x] Mailboxes — CRUD completo (`/mailboxes`, `/mailboxes/:username`)
- [x] Aliases — CRUD completo (`/aliases`, `/aliases/:address`)
- [x] Alias Domains — CRUD completo (`/alias-domains`, `/alias-domains/:alias_domain`)
- [x] Admins — CRUD completo (`/admins`, `/admins/:username`)
- [x] Transports — CRUD completo (`/transports`, `/transports/:id`)
- [x] API Keys — CRUD (`/settings/apikeys`, `/settings/apikeys/:id`)

## User Self-Service ✅

- [x] `GET  /user/me` — perfil do usuário
- [x] `GET  /user/forwarding` + `POST /user/forwarding` — forwarding de e-mail
- [x] `POST /user/password` — troca de senha
- [x] `GET  /user/vacation` + `POST /user/vacation` + `DELETE /user/vacation` — férias/autoresponder

## Dashboard e Relatórios ✅

- [x] `GET /dashboard` — estatísticas gerais
- [x] `GET /logs` — logs de admin (paginado, scoped)
- [x] `GET /maillog` — mail log (paginado, scoped)
- [x] `GET /version` — versão da aplicação

## Infraestrutura ✅

- [x] OpenAPI/Swagger embutido (`/swagger/index.html`, `/swagger/doc.json`)
- [x] Middleware JWT com scoping por domínio
- [x] CSRF e rate limiting global
- [x] SPA fallback (`GET /*` → `index.html`)
