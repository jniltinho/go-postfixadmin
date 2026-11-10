# toast — dependency-free toasts for Vue 3

A small, self-contained toast module: a reactive queue (`toast.ts`) and one host component (`ToastHost.vue`). No
external packages, no global CSS framework required; every color and size is a CSS variable with a neutral default.
To reuse it in another Vue 3 project, copy this folder.

## Use

```vue
<!-- App.vue: once -->
<template>
  <RouterView />
  <ToastHost />
</template>
<script setup lang="ts">
import { ToastHost } from '@/lib/toast'
</script>
```

```ts
import { toast, configureToast } from '@/lib/toast'

toast.success('Saved')
toast.error('Could not save', { duration: 6000 })
const id = toast.info('Uploading…', { duration: 0 }) // sticky until dismissed
toast.dismiss(id) // or toast.dismiss() for all
toast.warning('Disk almost full', { id: 'disk' }) // same id replaces instead of stacking

configureToast({ duration: 3000, max: 5, newestOnTop: true }) // defaults
```

`ToastHost` props: `position` (`top-right` default, `top-left`, `top-center`, `bottom-*`), `closeOnClick` (true),
`pauseOnFocusLoss` (true), `closeLabel` (`"Close"`, for the × button's aria-label).

## Behavior

- Errors use `role="alert"` (assertive), the others `role="status"` (polite).
- The countdown is the progress bar's CSS animation: it pauses on hover or keyboard focus and while the window is in
  the background; when it ends the toast closes. Click (or ×) closes it at once.
- At most `max` toasts; the oldest leave first. Transitions are turned off with `prefers-reduced-motion`.

## Theming

| Variable | Default |
|----------|---------|
| `--toast-{info,success,warning,error}-bg` / `-fg` | `#1d4ed8` / `#15803d` / `#b45309` / `#b91c1c`, white text |
| `--toast-font`, `--toast-font-size`, `--toast-font-weight` | `inherit`, `0.9375rem`, `700` |
| `--toast-radius`, `--toast-shadow` | `0.375rem`, soft shadow |
| `--toast-border-width`, `--toast-border-color` | `0`, `transparent` |
| `--toast-width`, `--toast-gap`, `--toast-offset`, `--toast-z` | `22.5rem`, `0.75rem`, `1rem`, `9999` |
| `--toast-padding`, `--toast-min-height`, `--toast-bar-height` | `0.875rem 1rem 1.125rem`, `3.5rem`, `0.3125rem` |

Go-PostfixAdmin maps them to its theme roles at the end of `src/style.css`, so toasts follow the 20 themes (text color
per status chosen for contrast in each theme).

Tests: `toast.test.ts` (`npm run test`).
