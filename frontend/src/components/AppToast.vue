<template></template>
<style>
/* ─── Neo-brutalist theme tokens ─────────────────────────────────────────── */
:root {
  --toastify-color-info:             #3b82f6;
  --toastify-color-success:          #22c55e;
  --toastify-color-warning:          #f97316;
  --toastify-color-error:            #ef4444;

  --toastify-icon-color-info:        #fff;
  --toastify-icon-color-success:     #fff;
  --toastify-icon-color-warning:     #fff;
  --toastify-icon-color-error:       #fff;

  --toastify-text-color-info:        #fff;
  --toastify-text-color-success:     #fff;
  --toastify-text-color-warning:     #fff;
  --toastify-text-color-error:       #fff;

  --toastify-color-progress-info:    rgba(255,255,255,0.4);
  --toastify-color-progress-success: rgba(255,255,255,0.4);
  --toastify-color-progress-warning: rgba(255,255,255,0.4);
  --toastify-color-progress-error:   rgba(255,255,255,0.4);
  --toastify-color-progress-colored: rgba(255,255,255,0.4);

  --toastify-toast-width:            360px;
  --toastify-toast-min-height:       100px;
  --toastify-font-family:            'Fira Sans', ui-sans-serif, system-ui, sans-serif;
  --toastify-z-index:                99999;
}

/* ─── Container ─────────────────────────────────────────────────────────── */
.Toastify__toast-container {
  padding: 0;
}

/* ─── Toast base: square corners + neo-brutalist shadow ─────────────────── */
.Toastify__toast {
  border-radius: 0;
  border: 2px solid #1e293b;
  box-shadow: 4px 4px 0 rgba(30, 41, 59, 0.35);
  padding: 14px 16px;
  font-family: var(--toastify-font-family);
  font-size: 15px;
  font-weight: 700;
}

.Toastify__toast-body {
  line-height: 1.4;
  gap: 10px;
}

/* ─── Close button ───────────────────────────────────────────────────────── */
.Toastify__close-button {
  color: rgba(255, 255, 255, 0.8);
  opacity: 1;
}
.Toastify__close-button:hover {
  color: #fff;
}

/* ─── Progress bar ───────────────────────────────────────────────────────── */
.Toastify__progress-bar {
  height: 6px;
}
</style>
