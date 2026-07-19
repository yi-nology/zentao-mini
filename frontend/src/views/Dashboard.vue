<template>
  <div class="dashboard-container">
    <div class="dashboard-toolbar">
      <div class="time-range">
        <span class="time-range-label">时间范围：</span>
        <el-radio-group v-model="timeRange" size="small">
          <el-radio-button label="7d">近 7 天</el-radio-button>
          <el-radio-button label="30d">近 30 天</el-radio-button>
          <el-radio-button label="90d">近 90 天</el-radio-button>
          <el-radio-button label="all">全部</el-radio-button>
          <el-radio-button label="custom">自定义</el-radio-button>
        </el-radio-group>
        <el-date-picker
          v-if="timeRange === 'custom'"
          v-model="customDateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          size="small"
          style="width: 240px; margin-left: 8px"
        />
        <span class="time-range-current">{{ timeRangeLabel }}</span>
      </div>
    </div>

    <div v-if="loading" class="loading-wrapper">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="error" class="error-banner">
      <span>加载仪表盘数据失败：{{ error }}</span>
      <el-button type="primary" size="small" @click="fetchData">重试</el-button>
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
            <span class="stat-value">{{ data.bugs?.total ?? 0 }}</span>
            <span class="stat-sub">活跃 <strong>{{ data.bugs?.active ?? 0 }}</strong> · 已解决 {{ data.bugs?.resolved ?? 0 }} · 已关闭 {{ data.bugs?.closed ?? 0 }}</span>
            <span v-if="bugTypeSummary" class="stat-sub">{{ bugTypeSummary }}</span>
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
            <span class="stat-value">{{ data.stories?.total ?? 0 }}</span>
            <span class="stat-sub">活跃 <strong>{{ data.stories?.active ?? 0 }}</strong> · 草稿 {{ data.stories?.draft ?? 0 }} · 已关闭 {{ data.stories?.closed ?? 0 }}</span>
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
            <span class="stat-value">{{ data.tasks?.total ?? 0 }}</span>
            <span class="stat-sub">进行中 <strong>{{ data.tasks?.doing ?? 0 }}</strong> · 待开始 {{ data.tasks?.wait ?? 0 }} · 已完成 {{ data.tasks?.done ?? 0 }}</span>
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
            <span class="stat-value">{{ (data.timelog?.totalHours ?? 0).toFixed(1) }}<small>h</small></span>
            <span class="stat-sub">本周 <strong>{{ (data.timelog?.thisWeekHours ?? 0).toFixed(1) }}h</strong></span>
          </div>
        </div>
      </div>

      <!-- Charts -->
      <div class="charts-grid">
        <div class="chart-card">
          <div class="chart-card-header">
            <h3>Bug 严重程度分布</h3>
            <el-button v-if="hasSeverityData" link size="small" @click="downloadChart(0, 'bug-severity')">下载图片</el-button>
          </div>
          <div v-if="hasSeverityData" class="chart-wrapper"><canvas ref="severityChartRef" /></div>
          <div v-else class="chart-empty">暂无数据</div>
        </div>
        <div class="chart-card">
          <div class="chart-card-header">
            <h3>Bug 类型分布</h3>
            <el-button v-if="hasTypeData" link size="small" @click="downloadChart(1, 'bug-type')">下载图片</el-button>
          </div>
          <div v-if="hasTypeData" class="chart-wrapper"><canvas ref="typeChartRef" /></div>
          <div v-else class="chart-empty">暂无数据</div>
        </div>
        <div class="chart-card">
          <div class="chart-card-header">
            <h3>任务状态分布</h3>
            <el-button v-if="hasTaskData" link size="small" @click="downloadChart(2, 'task-status')">下载图片</el-button>
          </div>
          <div v-if="hasTaskData" class="chart-wrapper"><canvas ref="taskChartRef" /></div>
          <div v-else class="chart-empty">暂无数据</div>
        </div>
      </div>

      <!-- Recent Lists -->
      <div class="lists-grid">
        <div class="list-card">
          <div class="list-header">
            <h3>最近 Bug</h3>
            <router-link to="/bugs" class="list-link">查看全部 →</router-link>
          </div>
          <div v-if="!data.recentBugs || data.recentBugs.length === 0" class="list-empty">暂无数据</div>
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
          <div v-if="!data.recentTasks || data.recentTasks.length === 0" class="list-empty">暂无数据</div>
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
import { ref, inject, watch, computed, onBeforeUnmount, nextTick } from 'vue'
import { getDashboard } from '@/api/zentao'
import { ElMessage } from 'element-plus'
import Chart from 'chart.js/auto'
import type { Chart as ChartType } from 'chart.js/auto'
import type { DashboardData } from '@/types/api'

interface GlobalSelection {
  product: number | null
  project: number | null
}

const globalSelection = inject<GlobalSelection>('globalSelection')!
const loading = ref(false)
const error = ref('')
const data = ref<DashboardData | null>(null)

// 时间范围筛选：近 7 天 / 近 30 天 / 近 90 天 / 全部 / 自定义
type TimeRangeKey = '7d' | '30d' | '90d' | 'all' | 'custom'
const timeRange = ref<TimeRangeKey>('all')
const customDateRange = ref<[string, string] | null>(null)

const formatDate = (d: Date): string => {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

// 计算实际起止日期（YYYY-MM-DD）
const timeRangeParams = computed<{ startDate?: string; endDate?: string }>(() => {
  if (timeRange.value === 'all') return {}
  if (timeRange.value === 'custom') {
    if (!customDateRange.value || !customDateRange.value[0] || !customDateRange.value[1]) return {}
    return {
      startDate: customDateRange.value[0],
      endDate: customDateRange.value[1]
    }
  }
  const days = timeRange.value === '7d' ? 7 : timeRange.value === '30d' ? 30 : 90
  const end = new Date()
  const start = new Date()
  start.setDate(end.getDate() - days + 1)
  return { startDate: formatDate(start), endDate: formatDate(end) }
})

// 时间范围标签描述（显示在卡片上方）
const timeRangeLabel = computed(() => {
  const p = timeRangeParams.value
  if (!p.startDate) return '全部时间'
  return `${p.startDate} 至 ${p.endDate}`
})

const severityChartRef = ref<HTMLCanvasElement | null>(null)
const typeChartRef = ref<HTMLCanvasElement | null>(null)
const taskChartRef = ref<HTMLCanvasElement | null>(null)
let charts: ChartType[] = []

const bugTypeSummary = computed(() => {
  const byType = data.value?.bugs?.byType
  if (!byType) return ''
  return Object.entries(byType)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 4)
    .map(([type, count]) => `${type} ${count}`)
    .join(' · ')
})

const destroyCharts = () => {
  charts.forEach(c => c.destroy())
  charts = []
}

const severityLabelMap: Record<string, string> = { '1': '致命', '2': '严重', '3': '一般', '4': '轻微', '5': '建议' }

const hasSeverityData = computed(() => {
  const bySeverity = data.value?.bugs?.bySeverity
  return bySeverity && Object.keys(bySeverity).length > 0
})

const hasTypeData = computed(() => {
  const byType = data.value?.bugs?.byType
  return byType && Object.keys(byType).length > 0
})

const hasTaskData = computed(() => {
  const tasks = data.value?.tasks
  if (!tasks) return false
  return (tasks.wait || 0) + (tasks.doing || 0) + (tasks.done || 0) + (tasks.closed || 0) > 0
})

const renderCharts = () => {
  destroyCharts()
  if (!data.value) return

  const bugs = data.value.bugs
  const tasks = data.value.tasks

  // Bug severity pie
  if (severityChartRef.value && bugs?.bySeverity && Object.keys(bugs.bySeverity).length > 0) {
    const labels = Object.keys(bugs.bySeverity).map(k => severityLabelMap[k] || `等级${k}`)
    const values = Object.values(bugs.bySeverity)
    charts.push(new Chart(severityChartRef.value, {
      type: 'doughnut',
      data: {
        labels,
        datasets: [{ data: values, backgroundColor: ['#EF4444', '#F59E0B', '#3B82F6', '#94A3B8'], borderWidth: 0 }]
      },
      options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom', labels: { padding: 16, usePointStyle: true, pointStyleWidth: 10 } } } }
    }))
  }

  // Bug type pie
  if (typeChartRef.value && bugs?.byType && Object.keys(bugs.byType).length > 0) {
    const labels = Object.keys(bugs.byType)
    const values = Object.values(bugs.byType)
    const colors = ['#EF4444', '#F59E0B', '#3B82F6', '#22C55E', '#8B5CF6', '#EC4899', '#06B6D4', '#F97316', '#6366F1', '#14B8A6']
    charts.push(new Chart(typeChartRef.value, {
      type: 'doughnut',
      data: {
        labels,
        datasets: [{ data: values, backgroundColor: colors.slice(0, labels.length), borderWidth: 0 }]
      },
      options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom', labels: { padding: 16, usePointStyle: true, pointStyleWidth: 10 } } } }
    }))
  }

  // Task status doughnut
  if (taskChartRef.value && tasks) {
    const entries = [
      { key: 'wait', label: '未开始' },
      { key: 'doing', label: '进行中' },
      { key: 'done', label: '已完成' },
      { key: 'closed', label: '已关闭' }
    ].filter(e => (tasks as Record<string, number>)[e.key] > 0)

    if (entries.length > 0) {
      charts.push(new Chart(taskChartRef.value, {
        type: 'doughnut',
        data: {
          labels: entries.map(e => e.label),
          datasets: [{ data: entries.map(e => (tasks as Record<string, number>)[e.key]), backgroundColor: ['#94A3B8', '#3B82F6', '#22C55E', '#6B7280'], borderWidth: 0 }]
        },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom', labels: { padding: 16, usePointStyle: true, pointStyleWidth: 10 } } } }
      }))
    }
  }
}

const fetchData = async (): Promise<void> => {
  const pid = globalSelection.product
  if (!pid) {
    data.value = null
    return
  }
  loading.value = true
  error.value = ''
  try {
    const res = await getDashboard(pid, timeRangeParams.value)
    data.value = res.data
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '未知错误'
    console.error('获取仪表盘数据失败:', e)
    error.value = msg
    data.value = null
    ElMessage.error('获取仪表盘数据失败')
  } finally {
    loading.value = false
  }
}

// 时间范围变化时重新加载
watch([timeRange, customDateRange], () => {
  if (globalSelection.product) fetchData()
})

// 下载图表为 PNG（chart.toBase64Image()）
const downloadChart = (index: number, name: string): void => {
  const chart = charts[index]
  if (!chart) return
  // chart.js 提供 toBase64Image()
  const base64 = (chart as any).toBase64Image('image/png', 1)
  const link = document.createElement('a')
  link.download = `${name}-${Date.now()}.png`
  link.href = base64
  link.click()
}

watch(() => globalSelection.product, (val) => {
  destroyCharts()
  if (val) fetchData()
  else data.value = null
}, { immediate: true })

watch(data, async (newData) => {
  if (!newData) return
  await nextTick()
  renderCharts()
}, { flush: 'post' })

onBeforeUnmount(() => { destroyCharts() })

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

.dashboard-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
.time-range {
  display: flex;
  align-items: center;
  gap: 8px;
}
.time-range-label {
  font-size: 13px;
  color: var(--color-text-secondary);
}
.time-range-current {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-left: 8px;
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

/* Error Banner */
.error-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: var(--space-lg);
  margin-bottom: var(--space-lg);
  background: var(--color-danger-light);
  color: var(--color-danger);
  border-radius: var(--radius-md);
  font-size: 14px;
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

/* Charts Grid */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
}

.chart-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  padding: var(--space-lg);
}

.chart-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-md);
}

.chart-card-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.chart-card h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 var(--space-md) 0;
}

.chart-wrapper {
  position: relative;
  height: 220px;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 220px;
  color: var(--color-text-tertiary);
  font-size: 13px;
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
.tag--draft, .tag--pause { background: var(--color-info-light); color: var(--color-text-tertiary); }

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
  .charts-grid {
    grid-template-columns: 1fr 1fr;
  }
}

@media screen and (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .charts-grid {
    grid-template-columns: 1fr;
  }
  .lists-grid {
    grid-template-columns: 1fr;
  }
}
</style>
