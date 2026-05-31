# Features — migration-quasar-tailwind

## Must Have

- [x] Remover Quasar completamente (código + dependências + CSS)
- [x] Instalar e configurar **Tailwind CSS v4**
- [x] Instalar **@lucide/vue** e mapear todos os ícones Material → Lucide
- [x] Reescrever `MainLayout.vue` sem Quasar (div + Tailwind + classes semânticas)
- [x] Substituir todos os `<q-icon>` por componentes Lucide
- [x] Substituir `q-spinner` / `q-skeleton` por CSS puro (`.spinner`, `.skeleton`)
- [x] Reescrever formulário do LoginPage (inputs nativos, sem `q-input`/`q-form`)
- [x] Atualizar `vite.config.ts` e `main.ts`
- [ ] Limpar `style.css` — remover todos os overrides de Quasar restantes
- [ ] Validação visual completa (todas as telas idênticas ou melhores)

## Nice to Have

- [ ] `BrutalTable.vue` — componente reutilizável com slots
- [ ] `ConfirmModal.vue` — modal de confirmação reutilizável
- [ ] Suporte a modo escuro via Tailwind
- [ ] Melhorar acessibilidade dos ícones (aria-labels)

## Non-Goals

- Não alterar a API nem o backend
- Não alterar o sistema de autenticação JWT
- Não reescrever a lógica de negócio das páginas (só os wrappers de UI)

## Entregáveis

- `package.json` sem `quasar` / `@quasar/*`
- `style.css` significativamente menor e mais legível
- Visual idêntico ou superior ao anterior
- Build limpo (`npm run build` sem warnings)
