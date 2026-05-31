# Tasks — migration-vue3-jwt

## Em andamento

_(nenhuma tarefa ativa)_

## Pendente

### Documentação
- [ ] Atualizar `DEVELOPMENT.md` com novo workflow (Vue 3 + Vite + Go dev server paralelo)
- [ ] Atualizar `README` para refletir stack atual (Vue 3 SPA + JWT, sem Quasar)

### i18n (PR 04b)
- [ ] Criar `scripts/sync-i18n.sh` — valida que chaves usadas no frontend existem nos `.po`
- [ ] Makefile target `i18n-check` com falha em CI em caso de drift
- [ ] Centralizar `getLang`/`flashLang` duplicados em `internal/i18n.GetPreferredLang()`
- [ ] Popular `frontend/src/locales/` com strings existentes de `web/static/js/i18n/`

### Open Questions a resolver
- [ ] Login: aceitar `type: "admin"|"user"` explícito ou auto-detectar?
- [ ] Deep-links sem autenticação (ex: `/mailboxes?domain=x`) devem redirecionar preservando o destino?
- [ ] Revogação de refresh tokens em v2 (tabela `refresh_tokens` com `jti`)?
