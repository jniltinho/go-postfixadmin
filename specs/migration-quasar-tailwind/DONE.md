# Done — migration-quasar-tailwind

## Fase 1 — Infraestrutura ✅

- [x] Tailwind CSS v4 + `@tailwindcss/vite` instalados e configurados
- [x] `@lucide/vue` instalado (substituiu Quasar Material Icons)
- [x] `quasar`, `@quasar/extras`, `@quasar/vite-plugin` removidos do `package.json`
- [x] `vite.config.ts` limpo (sem plugin do Quasar)
- [x] `main.ts` limpo (sem `Quasar`, `Dialog`, imports de CSS do Quasar)
- [x] `src/css/quasar-variables.sass` deletado

## Fase 2 — Componentes Reutilizáveis ✅

- [x] `Icon.vue` criado e registrado globalmente
- [x] `BrutalButton.vue` criado e registrado globalmente
- [x] `BrutalCard.vue` criado e registrado globalmente
- [x] `BrutalModal.vue` criado e registrado globalmente

## Fase 3 — MainLayout ✅

- [x] `MainLayout.vue` totalmente reescrito sem nenhum `<q-*>`
- [x] Layout via CSS Grid: sidebar 240px fixo + header 56px + conteúdo
- [x] Sidebar sempre visível no desktop (decisão do usuário)

## Fase 4 — Páginas ✅

- [x] Todas as 11 páginas migradas — zero `<q-*>` restante
- [x] `q-icon` → componentes Lucide em todas as páginas
- [x] `q-spinner` → `.spinner` (CSS puro)
- [x] `q-skeleton` → `.skeleton` (CSS puro com `animate-pulse`)
- [x] `q-form` / `q-input` no LoginPage → inputs nativos HTML

## Fase 5 — CSS (parcial) ✅

- [x] Centralizados em `style.css`: `.spinner`, `.skeleton`, `.error-banner`, `.btn-primary`, `.page-header`, table wrappers, modal primitives
- [x] DomainsPage totalmente convertido para `<BrutalModal>` (Add/Edit/Delete)
- [x] ~80 linhas de CSS duplicado de wrappers removidas do DomainsPage
- [x] `btn-primary`, `error-banner`, `table-card` movidos para o global (sem duplicação scoped)
