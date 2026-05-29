# Go-PostfixAdmin Frontend

Vue 3 + TypeScript + Vite + Tailwind CSS v4 + Lucide Icons

## Stack

- **Vue 3** (script setup + Pinia + Vue Router)
- **Tailwind CSS v4** (via `@tailwindcss/vite`)
- **Lucide Icons** (`@lucide/vue`)
- **Neo-Brutalist Design System** (custom components + centralized primitives)

## Key Components

Located in `src/components/ui/`:

- `Icon.vue` – Wrapper for Lucide icons
- `BrutalButton.vue`
- `BrutalCard.vue`
- `BrutalModal.vue` – Replaces old custom modals

All are registered globally.

## Development

```bash
npm install
npm run dev     # http://localhost:9000 (proxies to Go backend :8080)
npm run build   # outputs to ../web/dist
```

## Migration Note (2026)

This frontend was migrated from Quasar v2 to a lightweight Tailwind + custom brutalist stack to reduce CSS complexity while preserving the exact visual identity (thick borders, hard shadows, Fira fonts, brand colors).

See `DOCUMENTS/MIGRATION_PLAN_REMOVE_QUASAR_TAILWIND_LUCIDE.md` for full history.
