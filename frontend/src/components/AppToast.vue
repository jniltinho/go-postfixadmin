<template></template>
<style>
/* ─── Neo-brutalist theme tokens ─────────────────────────────────────────── */
:root {
  --toastify-color-info: var(--color-icon-accent);
  --toastify-color-success: var(--color-ok);
  --toastify-color-warning: var(--color-warn);
  --toastify-color-error: var(--color-err);

  --toastify-icon-color-info: var(--color-on-toast-info);
  --toastify-icon-color-success: var(--color-on-toast-ok);
  --toastify-icon-color-warning: var(--color-on-toast-warn);
  --toastify-icon-color-error: var(--color-on-toast-err);

  --toastify-text-color-info: var(--color-on-toast-info);
  --toastify-text-color-success: var(--color-on-toast-ok);
  --toastify-text-color-warning: var(--color-on-toast-warn);
  --toastify-text-color-error: var(--color-on-toast-err);

  --toastify-color-progress-info: color-mix(in srgb, var(--color-on-toast-info) 40%, transparent);
  --toastify-color-progress-success: color-mix(in srgb, var(--color-on-toast-ok) 40%, transparent);
  --toastify-color-progress-warning: color-mix(in srgb, var(--color-on-toast-warn) 40%, transparent);
  --toastify-color-progress-error: color-mix(in srgb, var(--color-on-toast-err) 40%, transparent);
  --toastify-color-progress-colored: color-mix(in srgb, var(--color-on-action) 40%, transparent);

  --toastify-toast-width: 360px;
  --toastify-toast-min-height: 100px;
  --toastify-font-family: var(--font-sans);
  --toastify-z-index: 99999;
}

/* ─── Container ─────────────────────────────────────────────────────────── */
.Toastify__toast-container {
  padding: 0;
}

/* ─── Toast base: square corners + neo-brutalist shadow ─────────────────── */
.Toastify__toast {
  border-radius: 0;
  border: 2px solid var(--color-line);
  box-shadow: 4px 4px 0 color-mix(in srgb, var(--color-shadow) 35%, transparent);
  padding: 0.875rem 1rem;
  font-family: var(--toastify-font-family);
  font-size: 0.9375rem;
  font-weight: 700;
}

.Toastify__toast-body {
  line-height: 1.4;
  gap: 0.625rem;
}

/* ─── Close button ───────────────────────────────────────────────────────── */
.Toastify__close-button {
  color: color-mix(in srgb, var(--color-on-action) 80%, transparent);
  opacity: 1;
}
.Toastify__close-button:hover {
  color: var(--color-on-action);
}

/* ─── Progress bar ───────────────────────────────────────────────────────── */
.Toastify__progress-bar {
  height: 0.375rem;
}
</style>
