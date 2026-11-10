import { defineConfig } from 'vite-plus'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  define: {
    API_BASE: JSON.stringify('/api/v1'),
  },
  server: {
    port: 9000,
    proxy: {
      // ws: the Settings terminal's WebSocket; Host stays the dev server's, so the panel's Origin check passes
      '/api': { target: 'http://localhost:8080', ws: true },
      '/lang': 'http://localhost:8080',
    },
  },
  fmt: { singleQuote: true, semi: false, printWidth: 120 },
  build: {
    outDir: '../web/dist',
    emptyOutDir: true,
  },
})
