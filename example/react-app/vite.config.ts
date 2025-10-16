import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: [['babel-plugin-react-compiler']],
      },
    }),
  ],
  server: {
    proxy: {
      '/rooms': {
        target: 'https://127.0.0.1:4433',
        changeOrigin: true,
        secure: false, // 忽略 HTTPS 证书验证
      },
      '/frames': {
        target: 'https://127.0.0.1:4433',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
