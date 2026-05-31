# Plan — api-rest

## Melhorias planejadas

### Paginação e Filtros
- [ ] Padronizar resposta de listagem com envelope `{ data, total, page, per_page }`
- [ ] Adicionar filtros por query string nos endpoints de listagem (ex: `?domain=example.com`, `?active=true`)
- [ ] Suporte a ordenação (`?sort=created_at&order=desc`)

### Segurança
- [ ] Rate limiting por usuário autenticado (além do global)
- [ ] Rotação automática de refresh tokens
- [ ] Revogar todos os tokens ao trocar senha

### Documentação
- [ ] Completar anotações Swagger em todos os handlers
- [ ] Exemplos de request/response em cada endpoint

### Novas Funcionalidades
- [ ] Endpoint de broadcast email
- [ ] Endpoint de exportação (CSV/JSON) de mailboxes e aliases

