<template>
  <div class="layout-container">
    <aside class="aside">
      <div class="logo">
        <div class="logo-icon">Z</div>
        <span class="logo-text">禅道 Mini</span>
      </div>
      <nav class="nav-menu">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: $route.path === item.path }"
        >
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path :d="item.icon" />
          </svg>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="sidebar-footer">
        <span class="version-info">v{{ appVersion }}</span>
        <span>©2024-2026 <a href="https://murphyyi.com" target="_blank" rel="noopener noreferrer">murphyyi.com</a></span>
      </div>
    </aside>
    <div class="main-area">
      <header class="header">
        <h1 class="header-title">{{ pageTitle }}</h1>
        <div class="header-actions">
          <ProductSelector
            :model-value="globalSelection"
            @update:model-value="handleSelectionChange"
          />
        </div>
      </header>
      <main class="main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, provide, reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import ProductSelector from '@/components/ProductSelector.vue'

interface GlobalSelection {
  product: string
  project: string
}

interface SelectionChangePayload {
  product: string
  project: string
}

interface MenuItem {
  path: string
  label: string
  icon: string
}

const route = useRoute()
const globalSelection = reactive<GlobalSelection>({ product: '', project: '' })
const appVersion = ref('...')

onMounted(async () => {
  try {
    const res = await fetch('/api/version')
    const json = await res.json()
    if (json.data?.version) appVersion.value = json.data.version
  } catch { appVersion.value = 'dev' }
})

const menuItems: MenuItem[] = [
  { path: '/bugs', label: 'Bug 查询', icon: 'M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 100-16 8 8 0 000 16zm-1-5h2v2h-2v-2zm0-8h2v6h-2V7z' },
  { path: '/stories', label: '需求查询', icon: 'M4 5a2 2 0 012-2h12a2 2 0 012 2v14a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm4 0v4h8V5H8zm0 6v4h8v-4H8zm0 6v2h5v-2H8z' },
  { path: '/tasks', label: '任务查询', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
  { path: '/timelog', label: '工时统计', icon: 'M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 100-16 8 8 0 000 16zm1-13h-2v6l5.25 3.15.75-1.23-4-2.42V7z' },
  { path: '/mcp-guide', label: 'MCP 对接', icon: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
  { path: '/init-guide', label: '重新初始化', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' }
]

const pageTitle = computed<string>(() => {
  const current = menuItems.find(item => item.path === route.path)
  return current?.label || '禅道 Mini'
})

const handleSelectionChange = (selection: SelectionChangePayload): void => {
  globalSelection.product = selection.product
  globalSelection.project = selection.project
}

provide<GlobalSelection>('globalSelection', globalSelection)
</script>

<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  overflow: hidden;
  background-color: var(--color-bg);
}

/* Sidebar */
.aside {
  width: 220px;
  flex-shrink: 0;
  background-color: var(--color-sidebar);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid #334155;
}

.logo-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-heading);
}

.logo-text {
  color: var(--color-text-on-dark);
  font-size: 16px;
  font-weight: 600;
  font-family: var(--font-heading);
}

/* Navigation */
.nav-menu {
  flex: 1;
  padding: 12px 10px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  color: var(--color-text-on-dark);
  text-decoration: none;
  font-size: 14px;
  font-weight: 400;
  transition: all var(--transition-fast);
  opacity: 0.7;
}

.nav-item:hover {
  background-color: var(--color-sidebar-hover);
  opacity: 1;
}

.nav-item.active {
  background-color: var(--color-sidebar-active);
  color: var(--color-text-on-primary);
  opacity: 1;
}

.nav-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

/* Sidebar Footer */
.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid #334155;
  text-align: center;
  font-size: 11px;
  color: var(--color-text-on-dark);
  opacity: 0.5;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.version-info {
  font-family: monospace;
  font-size: 11px;
  opacity: 0.7;
}

.sidebar-footer a {
  color: var(--color-text-on-dark);
  text-decoration: none;
  transition: opacity 0.2s;
}

.sidebar-footer a:hover {
  opacity: 1;
  text-decoration: underline;
}

/* Main Area */
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header */
.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  background-color: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
  font-family: var(--font-heading);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* Content */
.main {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-lg);
  background-color: var(--color-bg);
}

/* Responsive */
@media screen and (max-width: 768px) {
  .aside {
    width: 60px;
  }

  .logo-text {
    display: none;
  }

  .nav-item span {
    display: none;
  }

  .nav-item {
    justify-content: center;
    padding: 12px;
  }

  .header {
    padding: 0 16px;
  }
}
</style>
