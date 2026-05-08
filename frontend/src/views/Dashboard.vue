<template>
  <div class="dashboard-container">
    <div v-if="loading" class="loading-wrapper">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>

    <template v-else-if="data">
      <!-- Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card stat-card--bug">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 100-16 8 8 0 000 16zm-1-5h2v2h-2v-2zm0-8h2v6h-2V7z" />
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-label">Bug</span>
            <span class="stat-value">{{ data.bugs.total }}</span>
            <span class="stat-sub">活跃 <strong>{{ data.bugs.active }}</strong> · 已解决 {{ data.bugs.resolved }} · 已关闭 {{ data.bugs.closed }}</span>
          </div>
        </div>

        <div class="stat-card stat-card--story">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M4 5a2 2 0 012-2h12a2 2 0 012 2v14a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm4 0v4h8V5H8zm0 6v4h8v-4H8zm0 6v2h5v-2H8z" />
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-label">需求</span>
            <span class="stat-value">{{ data.stories.total }}</span>
            <span class="stat-sub">活跃 <strong>{{ data.stories.active }}</strong> · 草稿 {{ data.stories.draft }} · 已关闭 {{ data.stories.closed }}</span>
          </div>
        </div>

        <div class="stat-card stat-card--task">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-label">任务</span>
            <span class="stat-value">{{ data.tasks.total }}</span>
            <span class="stat-sub">进行中 <strong>{{ data.tasks.doing }}</strong> · 待开始 {{ data.tasks.wait }} · 已完成 {{ data.tasks.done }}</span>
          </div>
        </div>

        <div class="stat-card stat-card--time">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm0-2a8 8 0 100-16 8 8 0 000 16zm1-13h-2v6l5.25 3.15.75-1.23-4-2.42V7z" />
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-label">工时汇总</span>
            <span class="stat-value">{{ data.timelog.totalHours.toFixed(1) }}<small>h</small></span>
            <span class="stat-sub">本周 <strong>{{ data.timelog.thisWeekHours.toFixed(1) }}h</strong></span>
          </div>
        </div>
      </div>

      <!-- Recent Lists -->
      <div class="lists-grid">
        <div class="list-card">
          <div class="list-header">
            <h3>最近 Bug</h3>
            <router-link to="/bugs" class="list-link">查看全部 →</router-link>
          </div>
          <div v-if="data.recentBugs.length === 0" class="list-empty">暂无数据</div>
          <ul v-else class="list-body">
            <li v-for="bug in data.recentBugs" :key="bug.id" class="list-item">
              <span class="item-status-dot" :class="'dot--' + bug.status"></span>
              <span class="item-id">#{{ bug.id }}</span>
              <span class="item-title" :title="bug.title">{{ bug.title }}</span>
              <span class="item-status-tag" :class="'tag--' + bug.status">{{ getBugStatusLabel(bug.status) }}</span>
            </li>
          </ul>
        </div>

        <div class="list-card">
          <div class="list-header">
            <h3>最近任务</h3>
            <router-link to="/tasks" class="list-link">查看全部 →</router-link>
          </div>
          <div v-if="data.recentTasks.length === 0" class="list-empty">暂无数据</div>
          <ul v-else class="list-body">
            <li v-for="task in data.recentTasks" :key="task.id" class="list-item">
              <span class="item-status-dot" :class="'dot--' + task.status"></span>
              <span class="item-id">#{{ task.id }}</span>
              <span class="item-title" :title="task.name">{{ task.name }}</span>
              <span class="item-status-tag" :class="'tag--' + task.status">{{ getTaskStatusLabel(task.status) }}</span>
            </li>
          </ul>
        </div>
      </div>
    </template>

    <div v-else class="empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48">
        <path d="M3 13h2v-2H3v2zm0 4h2v-2H3v2zm0-8h2V7H3v2zm4 4h14v-2H7v2zm0 4h14v-2H7v2zM7 7v2h14V7H7z" />
      </svg>
      <p>请先在顶部选择产品以查看仪表盘数据</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, watch } from 'vue'
import { getDashboard } from '@/api/zentao'
import type { DashboardData } from '@/types/api'

interface GlobalSelection {
  product: string
  project: string
}

const globalSelection = inject<GlobalSelection>('globalSelection')!
const loading = ref(false)
const data = ref<DashboardData | null>(null)

const fetchData = async (): Promise<void> => {
  const pid = Number(globalSelection.product)
  if (!pid) {
    data.value = null
    return
  }
  loading.value = true
  try {
    const res = await getDashboard(pid)
    data.value = res.data.data
  } catch (e) {
    console.error('获取仪表盘数据失败:', e)
    data.value = null
  } finally {
    loading.value = false
  }
}

watch(() => globalSelection.product, (val) => {
  if (val) fetchData()
  else data.value = null
}, { immediate: true })

const getBugStatusLabel = (status: string): string => {
  const map: Record<string, string> = { active: '激活', resolved: '已解决', closed: '已关闭' }
  return map[status] || status
}

const getTaskStatusLabel = (status: string): string => {
  const map: Record<string, string> = { wait: '未开始', doing: '进行中', done: '已完成', pause: '已暂停', cancel: '已取消', closed: '已关闭' }
  return map[status] || status
}
</script>

<style scoped>
.dashboard-container {
  max-width: 1200px;
}

/* Loading */
.loading-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--color-text-tertiary);
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--color-border-light);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.stat-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  display: flex;
  align-items: flex-start;
  gap: var(--space-md);
  box-shadow: var(--shadow-sm);
  transition: box-shadow var(--transition-normal);
}

.stat-card:hover {
  box-shadow: var(--shadow-md);
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon svg {
  width: 22px;
  height: 22px;
}

.stat-card--bug .stat-icon { background: var(--color-danger-light); color: var(--color-danger); }
.stat-card--story .stat-icon { background: var(--color-primary-light); color: var(--color-primary); }
.stat-card--task .stat-icon { background: var(--color-warning-light); color: var(--color-warning); }
.stat-card--time .stat-icon { background: var(--color-success-light); color: var(--color-success); }

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.stat-label {
  font-size: 12px;
  color: var(--color-text-tertiary);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
}

.stat-value small {
  font-size: 14px;
  font-weight: 500;
  margin-left: 2px;
  color: var(--color-text-secondary);
}

.stat-sub {
  font-size: 12px;
  color: var(--color-text-tertiary);
  line-height: 1.4;
}

.stat-sub strong {
  color: var(--color-text-secondary);
  font-weight: 600;
}

/* Lists Grid */
.lists-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
}

.list-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  display: flex;
  flex-direction: column;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md) var(--space-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.list-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.list-link {
  font-size: 13px;
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
}

.list-link:hover {
  text-decoration: underline;
}

.list-body {
  list-style: none;
  padding: 0;
  margin: 0;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px var(--space-lg);
  border-bottom: 1px solid var(--color-border-light);
  transition: background var(--transition-fast);
}

.list-item:last-child {
  border-bottom: none;
}

.list-item:hover {
  background: var(--color-bg-hover);
}

.item-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot--active, .dot--doing { background: var(--color-danger); }
.dot--resolved, .dot--done { background: var(--color-success); }
.dot--closed { background: var(--color-info); }
.dot--wait { background: var(--color-warning); }
.dot--draft, .dot--pause { background: var(--color-text-tertiary); }

.item-id {
  font-size: 12px;
  color: var(--color-text-tertiary);
  font-family: monospace;
  flex-shrink: 0;
  min-width: 44px;
}

.item-title {
  flex: 1;
  font-size: 13px;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-status-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 100px;
  font-weight: 500;
  flex-shrink: 0;
  white-space: nowrap;
}

.tag--active, .tag--doing { background: var(--color-danger-light); color: var(--color-danger); }
.tag--resolved, .tag--done { background: var(--color-success-light); color: var(--color-success); }
.tag--closed { background: var(--color-info-light); color: var(--color-info); }
.tag--wait { background: var(--color-warning-light); color: var(--color-warning); }
.tag--draft, .tag--pause { background: #F1F5F9; color: var(--color-text-tertiary); }

.list-empty {
  padding: 40px var(--space-lg);
  text-align: center;
  color: var(--color-text-tertiary);
  font-size: 13px;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  gap: 12px;
  color: var(--color-text-tertiary);
}

.empty-state p {
  font-size: 14px;
}

/* Responsive */
@media screen and (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .lists-grid {
    grid-template-columns: 1fr;
  }
}
</style>
