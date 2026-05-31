# Tasks — migration-quasar-tailwind

## Em andamento

### Fase 5 — Limpeza de CSS
- [ ] Remover overrides de Quasar restantes em `style.css` (`:deep(.q-*)`, `.q-field`, `.q-drawer`, `.q-toolbar`)
- [ ] MailboxesPage: retomar conversão para `<BrutalModal>` (revertida por complexidade dos formulários — garantir build verde)
- [ ] Padronizar classes de modal nas páginas que ainda têm CSS scoped duplicado

## Pendente

### Fase 6 — Validação Visual e Funcional
- [ ] Comparar lado a lado todas as telas com screenshots de referência (`DOCUMENTS/screenshots/`)
- [ ] Testar CRUD completo: domains, mailboxes, aliases, admins, transports, apikeys
- [ ] Testar estados de loading e erro em todas as páginas
- [ ] Testar login/logout + refresh de token
- [ ] Testar responsividade (diferentes tamanhos de tela)
- [ ] `npm run build` limpo + verificar redução no tamanho do bundle

### Fase 7 — Pós-Migração
- [ ] `BrutalTable.vue` — componente reutilizável com slots para header/body
- [ ] `ConfirmModal.vue` — modal de confirmação reutilizável
- [ ] Atualizar `internal/routes/routes.go` ("Quasar SPA" → "Vue SPA")
- [ ] Verificar e remover `public/icons.svg` se órfão
- [ ] Verificar e remover `HelloWorld.vue` se órfão
