import { defineConfig } from 'vite'
import type { UserConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

const config: UserConfig = {
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@wailsjs': resolve(__dirname, 'wailsjs')
    }
  },
  base: './',
  server: {
    port: 6100,
    proxy: {
      '/api': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        secure: false
      },
      '/mcp': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        secure: false
      },
      '/health': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        secure: false
      },
      '/metrics': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        secure: false
      }
    }
  }
}

export default defineConfig(config)
