<template>
  <div class="page-container">
    <div class="settings-section">
      <h3 class="section-title">禅道连接信息</h3>
      <div class="settings-card">
        <div v-if="loading" class="loading-text">加载中...</div>
        <template v-else-if="accountInfo">
          <div class="info-row">
            <span class="info-label">服务器地址</span>
            <span class="info-value">{{ accountInfo.domain }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">登录账号</span>
            <span class="info-value">{{ accountInfo.account }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">连接状态</span>
            <span class="info-value">
              <el-tag :type="accountInfo.connected ? 'success' : 'danger'" size="small">
                {{ accountInfo.connected ? '已连接' : '未连接' }}
              </el-tag>
            </span>
          </div>
          <div class="info-row">
            <span class="info-label">检测延迟</span>
            <span class="info-value">{{ latency }}ms</span>
          </div>
        </template>
        <div v-else class="empty-text">无法获取连接信息</div>
      </div>
    </div>

    <div class="settings-section">
      <h3 class="section-title">显示设置</h3>
      <div class="settings-card">
        <div class="info-row">
          <span class="info-label">深色模式</span>
          <el-switch v-model="darkMode" @change="toggleDarkMode" />
        </div>
      </div>
    </div>

    <div class="settings-section">
      <h3 class="section-title">系统信息</h3>
      <div class="settings-card">
        <div class="info-row">
          <span class="info-label">应用版本</span>
          <span class="info-value">v{{ appVersion }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">前端框架</span>
          <span class="info-value">Vue 3 + Element Plus</span>
        </div>
        <div class="info-row">
          <span class="info-label">后端框架</span>
          <span class="info-value">Go + Hertz</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAccountInfo, testZentaoConnection } from '@/api/zentao'
import api from '@/api/api'
import type { ApiResponse } from '@/types/api'

interface AccountInfo {
  domain: string
  account: string
  connected: boolean
}

const loading = ref(true)
const accountInfo = ref<AccountInfo | null>(null)
const latency = ref(0)
const appVersion = ref('dev')
const darkMode = ref(false)

const fetchAccountInfo = async () => {
  loading.value = true
  try {
    const res = await getAccountInfo()
    if (res.code === 200 && res.data) {
      accountInfo.value = res.data
    }
  } catch { /* ignore */ }

  try {
    const start = Date.now()
    await testZentaoConnection()
    latency.value = Date.now() - start
  } catch { latency.value = -1 }

  try {
    const res = await api.get('/version') as ApiResponse<{ version: string }>
    if (res?.data?.version) appVersion.value = res.data.version
  } catch { /* ignore */ }

  loading.value = false
}

const toggleDarkMode = (val: boolean | string | number) => {
  const isDark = Boolean(val)
  document.documentElement.classList.toggle('dark', isDark)
  localStorage.setItem('zentao-mini-dark', isDark ? '1' : '0')
}

onMounted(() => {
  const saved = localStorage.getItem('zentao-mini-dark')
  darkMode.value = saved === '1'
  document.documentElement.classList.toggle('dark', darkMode.value)
  fetchAccountInfo()
})
</script>

<style scoped>
.settings-section {
  margin-bottom: var(--space-xl);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 var(--space-md) 0;
}

.settings-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  padding: var(--space-lg);
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border-light);
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 14px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.info-value {
  font-size: 14px;
  color: var(--color-text-primary);
  font-weight: 600;
}

.loading-text, .empty-text {
  text-align: center;
  padding: var(--space-lg);
  color: var(--color-text-tertiary);
  font-size: 14px;
}
</style>
