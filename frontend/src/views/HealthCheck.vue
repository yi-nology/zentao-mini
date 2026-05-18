<template>
  <div class="health-page">
    <div class="health-header">
      <h2>心跳检测</h2>
      <div class="header-actions">
        <span class="last-check" v-if="lastCheckTime">上次检测：{{ lastCheckTime }}</span>
        <button class="btn btn-primary" @click="runCheck" :disabled="loading">
          {{ loading ? '检测中...' : '立即检测' }}
        </button>
        <label class="auto-refresh">
          <input type="checkbox" v-model="autoRefresh" />
          自动刷新 (30s)
        </label>
      </div>
    </div>

    <div v-if="loading && !data" class="empty-state">检测中...</div>

    <template v-else-if="data">
      <div class="summary-bar" :class="data.summary.healthy ? 'summary-ok' : 'summary-fail'">
        <div class="summary-item">
          <span class="summary-dot" :class="data.summary.healthy ? 'dot-ok' : 'dot-fail'"></span>
          <span class="summary-text">{{ data.summary.healthy ? '系统正常' : '存在异常' }}</span>
        </div>
        <div class="summary-stats">
          <span class="stat-ok">{{ data.summary.ok }} 正常</span>
          <span class="stat-divider">/</span>
          <span class="stat-fail" v-if="data.summary.fail > 0">{{ data.summary.fail }} 异常</span>
          <span class="stat-total">共 {{ data.summary.total }} 项</span>
        </div>
      </div>

      <div class="checks-grid">
        <div v-if="data.zentao" class="check-card" :class="data.zentao.status === 'ok' ? 'card-ok' : 'card-fail'">
          <div class="card-header">
            <span class="card-icon">🔗</span>
            <span class="card-name">禅道连接</span>
            <span class="card-badge" :class="data.zentao.status === 'ok' ? 'badge-ok' : 'badge-fail'">
              {{ data.zentao.status === 'ok' ? '正常' : '异常' }}
            </span>
          </div>
          <div class="card-body">
            <div class="card-msg">{{ data.zentao.message }}</div>
            <div class="card-latency" v-if="data.zentao.latencyMs">{{ data.zentao.latencyMs }}ms</div>
          </div>
        </div>

        <div v-for="check in data.checks" :key="check.name" class="check-card" :class="check.status === 'ok' ? 'card-ok' : 'card-fail'">
          <div class="card-header">
            <span class="card-icon">{{ iconMap[check.name] || '📊' }}</span>
            <span class="card-name">{{ labelMap[check.name] || check.name }}</span>
            <span class="card-badge" :class="check.status === 'ok' ? 'badge-ok' : 'badge-fail'">
              {{ check.status === 'ok' ? '正常' : '异常' }}
            </span>
          </div>
          <div class="card-body">
            <div class="card-count" v-if="check.status === 'ok' && check.count > 0">
              {{ check.count }} 条数据
            </div>
            <div class="card-count card-count-zero" v-else-if="check.status === 'ok' && check.count === 0">
              暂无数据
            </div>
            <div class="card-msg card-err" v-if="check.status === 'fail'">{{ check.message }}</div>
            <div class="card-latency">{{ check.latencyMs }}ms</div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import api from '@/api/api'

interface CheckItem {
  name: string
  status: string
  count: number
  message: string
  latencyMs: number
}

interface HealthData {
  timestamp: string
  zentao: CheckItem
  checks: CheckItem[]
  summary: { total: number; ok: number; fail: number; healthy: boolean }
}

const data = ref<HealthData | null>(null)
const loading = ref(false)
const lastCheckTime = ref('')
const autoRefresh = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

const iconMap: Record<string, string> = {
  products: '📦',
  projects: '📁',
  bugs: '🐛',
  stories: '📋',
  tasks: '✅',
  users: '👥',
  scheduler: '⏰',
}

const labelMap: Record<string, string> = {
  products: '产品列表',
  projects: '项目列表',
  bugs: 'Bug 数据',
  stories: '需求数据',
  tasks: '任务数据',
  users: '用户数据',
  scheduler: '定时任务',
}

const runCheck = async () => {
  loading.value = true
  try {
    const res = await api.get('/healthz')
    data.value = (res as unknown as { data: HealthData }).data
    lastCheckTime.value = new Date().toLocaleString('zh-CN')
  } catch (e) {
    console.error('健康检查失败', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  runCheck()
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

const startAutoRefresh = () => {
  if (timer) clearInterval(timer)
  timer = setInterval(runCheck, 30000)
}

const stopAutoRefresh = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

import { watch } from 'vue'
watch(autoRefresh, (v) => {
  if (v) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})
</script>

<style scoped>
.health-page {
  max-width: 1000px;
}

.health-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.health-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.last-check {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.auto-refresh {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.summary-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-radius: var(--radius-md);
  margin-bottom: 20px;
}

.summary-ok {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}

.summary-fail {
  background: #fef2f2;
  border: 1px solid #fecaca;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.summary-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.dot-ok {
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.4);
}

.dot-fail {
  background: #ef4444;
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.4);
}

.summary-text {
  font-size: 15px;
  font-weight: 600;
}

.summary-stats {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.stat-ok { color: #16a34a; }
.stat-fail { color: #dc2626; font-weight: 600; }
.stat-divider { color: var(--color-text-tertiary); }
.stat-total { color: var(--color-text-tertiary); }

.checks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.check-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-light);
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.check-card:hover {
  box-shadow: var(--shadow-sm);
}

.card-ok {
  border-left: 3px solid #22c55e;
}

.card-fail {
  border-left: 3px solid #ef4444;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px 8px;
}

.card-icon {
  font-size: 18px;
}

.card-name {
  font-weight: 500;
  font-size: 14px;
  flex: 1;
}

.card-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 100px;
  font-weight: 500;
}

.badge-ok {
  background: #dcfce7;
  color: #16a34a;
}

.badge-fail {
  background: #fee2e2;
  color: #dc2626;
}

.card-body {
  padding: 0 16px 12px;
}

.card-count {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.card-count-zero {
  color: var(--color-text-tertiary);
  font-style: italic;
}

.card-msg {
  font-size: 12px;
  color: var(--color-text-tertiary);
  word-break: break-all;
}

.card-err {
  color: #dc2626;
}

.card-latency {
  font-size: 11px;
  color: var(--color-text-tertiary);
  margin-top: 4px;
  font-family: monospace;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.btn-primary:hover { opacity: 0.9; }
.btn:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: var(--color-text-tertiary);
}
</style>
