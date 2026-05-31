# Features — migration-vue3-jwt

## Goals (Must Have)

- [x] Backend com API REST versionada em `/api/v1/`
- [x] Autenticação JWT (access Bearer + refresh httpOnly cookie)
- [x] Vue 3 SPA embarcada no binário Go via `//go:embed`
- [x] Go serve SPA com history-mode fallback (`index.html` para rotas não-API)
- [x] Paridade funcional completa: domains, mailboxes, aliases, admins, alias-domains, transports, logs, maillog
- [x] User self-service: senha, forwarding, vacation + Sieve sync
- [x] i18n en/es/pt_BR
- [x] CLI, transport server, maillog reader, vacation sync: inalterados
- [x] Build atualizado (Makefile + Dockerfile multi-stage com Node)
- [x] Fidelidade visual neo-brutalist (cores, bordas, sombras, tipografia, ícones)
- [ ] Documentação de desenvolvimento atualizada (DEVELOPMENT.md)

## Non-Goals

- Mudanças no schema DB ou modelos GORM (exceto tabela opcional de revogação de tokens)
- Reescrever CLI, backup tools ou subsistemas não-web
- Adicionar novas features de negócio (UI de DKIM, por exemplo)
- WebSockets ou RBAC avançado além do modelo atual (superadmin + domain_admin)
- Suporte a idiomas além dos 3 atuais
- Compatibilidade com API do PostfixAdmin PHP original

## Gaps Identificados no Sistema Anterior

- Sem JSON body binding nos handlers antigos (apenas `c.FormValue`)
- Respostas de erro inconsistentes entre os métodos `*API`
- Flash messages vinculadas às sessões (substituídas por toasts no SPA)
- Sem rate limiting no login
- Lógica de idioma duplicada em vários lugares (`getLang`/`flashLang`)
- Bug latente: `SetFlash` sempre usava `UserSessionName` mesmo em fluxos admin

## Open Questions (em aberto)

- Login deve aceitar campo explícito `type: "admin"|"user"` ou auto-detectar?
- Diretório de saída do build frontend: `web/dist` direto ou `frontend/dist` + copy no Makefile?
- Deep-links sem login (ex: `/mailboxes?domain=example.com`) devem ser preservados?
- Localização de mensagens de erro da API: server-side ou só client-side?
- Tabela de revogação de refresh tokens em v2?
