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
    proxy: {
      '/api': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        secure: false
      }
    }
  }
}

export default defineConfig(config)
