import { defineConfig } from 'vite'
import type { UserConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import wails from '@wailsio/runtime/plugins/vite'

const config: UserConfig = {
  plugins: [vue(), wails('./bindings')],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  base: './',
  server: {
    port: 6100,
    proxy: {
      // ^ 开头的 key 按 RegExp 匹配（命中=转发）：/mcp 是后端 MCP 端点前缀，
      // 用正则避免吞掉前端路由 /mcp-guide。注意不要加 /health 代理——
      // /health 同时是前端"心跳检测"路由，代理会让整页加载拿到后端 JSON
      '/api': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        secure: false
      },
      '^/mcp($|[/?])': {
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
