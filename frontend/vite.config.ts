import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { quasar, transformAssetUrls } from '@quasar/vite-plugin'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue({
      template: { transformAssetUrls }
    }),
    quasar({
      sassVariables: 'src/css/quasar-variables.sass'
    })
  ],
  define: {
    API_BASE: JSON.stringify('/api/v1'),
  },
  server: {
    port: 9000,
    proxy: {
      '/api': 'http://localhost:8080',
      '/lang': 'http://localhost:8080'
    }
  },
  build: {
    outDir: '../web/dist',
    emptyOutDir: true
  }
})
