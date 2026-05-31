# Plan — migration-quasar-tailwind

## Fases

### Fase 0 — Preparação ✅
- [x] Criar branch `feat/remove-quasar-tailwind`
- [x] Fazer backup visual (screenshots das telas principais)
- [x] Documentar todos os ícones únicos usados

### Fase 1 — Infraestrutura ✅
- [x] Instalar `@lucide/vue` e `tailwindcss @tailwindcss/vite`
- [x] Remover dependências Quasar do `package.json`
- [x] Atualizar `vite.config.ts` (remover plugin Quasar)
- [x] Atualizar `main.ts` (remover Quasar plugin + imports de CSS)
- [x] Configurar Tailwind v4 no `style.css` (`@import "tailwindcss"` + `@theme`)
- [x] Deletar `src/css/quasar-variables.sass`

### Fase 2 — Componentes Reutilizáveis ✅
- [x] `Icon.vue`, `BrutalButton.vue`, `BrutalCard.vue`, `BrutalModal.vue` criados
- [x] Registrados globalmente em `main.ts`

### Fase 3 — Layout Principal ✅
- [x] `MainLayout.vue` totalmente reescrito sem Quasar
- [x] Sidebar sempre visível no desktop (grid CSS: 240px + 1fr)

### Fase 4 — Páginas (Ícones + Loading) ✅
- [x] Todas as 11 páginas migradas (zero `<q-*>` restante)
- [x] Substituídos: `q-icon`, `q-page`, `q-spinner`, `q-skeleton`, `q-form`/`q-input`

### Fase 5 — Limpeza de CSS 🔄
- [ ] Remover overrides de Quasar restantes (`:deep(.q-*)`, `.q-field`, `.q-drawer`)
- [ ] MailboxesPage: retomar conversão para `<BrutalModal>` com cuidado
- [ ] Padronizar classes de modal nas páginas com CSS scoped duplicado

### Fase 6 — Validação Visual e Funcional ⏳
- [ ] Comparar lado a lado todas as telas com screenshots de referência
- [ ] Testar CRUD completo: domains, mailboxes, aliases, admins, transports, apikeys
- [ ] Testar estados de loading e erro
- [ ] Testar login/logout + refresh de token
- [ ] `npm run build` limpo + verificar tamanho do bundle

### Fase 7 — Pós-Migração ⏳ (opcional)
- [ ] `BrutalTable.vue` — componente de tabela com slots
- [ ] `ConfirmModal.vue` — modal de confirmação reutilizável
- [ ] Atualizar `internal/routes/routes.go` ("Quasar SPA" → "Vue SPA")
- [ ] Verificar/remover `public/icons.svg` e `HelloWorld.vue` se órfãos

---

## Estimativas

| Fase        | Estimado    | Status          |
|-------------|-------------|-----------------|
| 0 + 1       | 2-3h        | ✅ Concluído    |
| 2           | 2h          | ✅ Concluído    |
| 3 (Layout)  | 4-6h        | ✅ Concluído    |
| 4 (Páginas) | 3-4h        | ✅ Concluído    |
| 5 (CSS)     | 3-4h        | 🔄 Em andamento |
| 6 (Testes)  | 2-3h        | ⏳ Pendente     |
| 7           | 2-4h        | ⏳ Pendente     |

**Total**: 18-26h (2-4 dias úteis).
