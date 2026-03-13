<template>
  <div class="timelog-container">
    <!-- 筛选区域 -->
    <div class="filters">
      <div class="quick-btns">
        <span>快捷选择:</span>
        <el-button 
          v-for="range in quickRanges" 
          :key="range.value"
          :class="{ 'active': selectedRange === range.value }"
          @click="setQuickRange(range.value)"
        >
          {{ range.label }}
        </el-button>
      </div>
      <el-form :inline="true" class="filter-form">
        <el-form-item label="执行/迭代">
          <el-select v-model="filters.executionId" placeholder="全部执行" style="width: 180px;">
            <el-option value="" label="全部执行"></el-option>
            <el-option 
              v-for="execution in executions" 
              :key="execution.id" 
              :value="execution.id" 
              :label="execution.name"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="用户">
          <el-select v-model="filters.assignedTo" placeholder="全部用户" filterable style="width: 160px;">
            <el-option value="" label="全部用户"></el-option>
            <el-option 
              v-for="user in users" 
              :key="user.account" 
              :value="user.account" 
              :label="user.realname || user.account"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="开始日期">
          <el-date-picker
            v-model="filters.dateFrom"
            type="date"
            placeholder="选择开始日期"
            value-format="YYYY-MM-DD"
            style="width: 150px;"
          ></el-date-picker>
        </el-form-item>
        <el-form-item label="结束日期">
          <el-date-picker
            v-model="filters.dateTo"
            type="date"
            placeholder="选择结束日期"
            value-format="YYYY-MM-DD"
            style="width: 150px;"
          ></el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="queryTimelog" :loading="loading" style="width: 120px;">查询统计</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 统计结果 -->
    <div v-if="showResult" class="result-section">
      <!-- 概览卡片 -->
      <div class="summary-cards">
        <el-card class="summary-card">
          <div class="card-value">{{ analysisData.totalHours.toFixed(1) }}</div>
          <div class="card-label">总工时 (小时)</div>
        </el-card>
        <el-card class="summary-card">
          <div class="card-value">{{ analysisData.effortCount || 0 }}</div>
          <div class="card-label">工时记录数</div>
        </el-card>
        <el-card class="summary-card">
          <div class="card-value">{{ (analysisData.byProject || []).length }}</div>
          <div class="card-label">涉及项目</div>
        </el-card>
        <el-card class="summary-card">
          <div class="card-value">{{ avgHours }}</div>
          <div class="card-label">日均工时</div>
        </el-card>
      </div>

      <!-- 图表区域 -->
      <div class="charts-container">
        <el-card class="chart-card full-width">
          <template #header>
            <div class="card-header">
              <span>每日工时</span>
            </div>
          </template>
          <div class="chart-wrapper">
            <canvas ref="dailyChart"></canvas>
          </div>
        </el-card>
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>按项目分布</span>
            </div>
          </template>
          <div class="chart-wrapper">
            <canvas ref="projectChart"></canvas>
          </div>
        </el-card>
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>按任务类型分布</span>
            </div>
          </template>
          <div class="chart-wrapper">
            <canvas ref="typeChart"></canvas>
          </div>
        </el-card>
      </div>

      <!-- 工时流水明细表 -->
      <el-card class="data-table">
        <template #header>
          <div class="table-header">
            <el-input
              v-model="tableSearch"
              placeholder="搜索任务名称/工作内容..."
              prefix-icon="el-icon-search"
              @input="filterTable"
              style="width: 300px"
            ></el-input>
            <span class="table-count">{{ filteredEfforts.length }} 条</span>
          </div>
        </template>
        <el-table
          :data="filteredEfforts"
          style="width: 100%"
          @sort-change="handleSortChange"
        >
          <el-table-column prop="date" label="日期" sortable width="120"></el-table-column>
          <el-table-column prop="taskName" label="任务名称" min-width="200">
            <template #default="scope">
              {{ scope.row.taskName.length > 40 ? scope.row.taskName.substring(0, 40) + '...' : scope.row.taskName }}
            </template>
          </el-table-column>
          <el-table-column prop="taskType" label="类型" width="100"></el-table-column>
          <el-table-column prop="project" label="项目" width="150"></el-table-column>
          <el-table-column prop="execution" label="执行" width="150"></el-table-column>
          <el-table-column prop="account" label="人员" width="120">
            <template #default="scope">
              {{ getUserName(scope.row.account) }}
            </template>
          </el-table-column>
          <el-table-column prop="consumed" label="消耗(h)" sortable width="100">
            <template #default="scope">
              <strong>{{ scope.row.consumed.toFixed(1) }}</strong>
            </template>
          </el-table-column>
          <el-table-column prop="work" label="工作内容" min-width="250">
            <template #default="scope">
              {{ (scope.row.work || '').length > 60 ? (scope.row.work || '').substring(0, 60) + '...' : (scope.row.work || '-') }}
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 加载遮罩 -->
    <el-loading v-if="loading" fullscreen text="正在获取数据..."></el-loading>

    <!-- 空状态 -->
    <div v-if="!showResult && !loading" class="empty-state">
      <el-empty description="请选择筛选条件并点击查询统计"></el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, inject, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { Ref, ComputedRef } from 'vue'
import Chart from 'chart.js/auto'
import type { Chart as ChartType } from 'chart.js/auto'
import {
  getTimelogDashboard,
  getTimelogEfforts,
  getTimelogExecutions,
  getUsers
} from '../api/zentao'
import type {
  Execution,
  User,
  TimelogAnalysis,
  TimelogByDate,
  TimelogByProject,
  TimelogByType,
  TimelogEffort
} from '../types/api'

interface QuickRange {
  label: string
  value: string
}

interface Filters {
  executionId: string
  assignedTo: string
  dateFrom: string
  dateTo: string
}

interface ChartInstances {
  dailyChart?: ChartType<'bar'>
  projectChart?: ChartType<'bar'>
  typeChart?: ChartType<'doughnut'>
}

interface GlobalSelection {
  product: string
  project: string
}

interface SortParams {
  prop: string
  order: string | null
}

const globalSelection = inject<GlobalSelection>('globalSelection') as GlobalSelection

const loading: Ref<boolean> = ref(false)
const showResult: Ref<boolean> = ref(false)
const selectedRange: Ref<string> = ref('thisMonth')
const tableSearch: Ref<string> = ref('')
const filters: Ref<Filters> = ref<Filters>({
  executionId: '',
  assignedTo: '',
  dateFrom: '',
  dateTo: ''
})

const executions: Ref<Execution[]> = ref<Execution[]>([])
const users: Ref<User[]> = ref<User[]>([])
const analysisData: Ref<TimelogAnalysis> = ref<TimelogAnalysis>({
  totalHours: 0,
  effortCount: 0,
  taskCount: 0,
  byDate: [],
  byProject: [],
  byType: [],
  efforts: []
})

const dailyChart: Ref<HTMLCanvasElement | null> = ref<HTMLCanvasElement | null>(null)
const projectChart: Ref<HTMLCanvasElement | null> = ref<HTMLCanvasElement | null>(null)
const typeChart: Ref<HTMLCanvasElement | null> = ref<HTMLCanvasElement | null>(null)
const chartInstances: Ref<ChartInstances> = ref<ChartInstances>({})

const quickRanges: QuickRange[] = [
  { label: '本周', value: 'thisWeek' },
  { label: '上周', value: 'lastWeek' },
  { label: '本月', value: 'thisMonth' },
  { label: '上月', value: 'lastMonth' }
]

const avgHours: ComputedRef<string> = computed(() => {
  const days: number = (analysisData.value.byDate || []).length
  return days > 0 ? (analysisData.value.totalHours / days).toFixed(1) : '0'
})

const filteredEfforts: ComputedRef<TimelogEffort[]> = computed(() => {
  if (!analysisData.value.efforts) return []
  const keyword: string = tableSearch.value.toLowerCase()
  if (!keyword) return analysisData.value.efforts
  return analysisData.value.efforts.filter((effort: TimelogEffort): boolean => {
    return (
      (effort.taskName || '').toLowerCase().includes(keyword) ||
      (effort.work || '').toLowerCase().includes(keyword) ||
      (effort.project || '').toLowerCase().includes(keyword) ||
      (effort.execution || '').toLowerCase().includes(keyword) ||
      (effort.date || '').includes(keyword) ||
      String(effort.taskId).includes(keyword)
    )
  })
})

const setQuickRange = (range: string): void => {
  selectedRange.value = range
  const today: Date = new Date()
  let from: Date, to: Date

  switch (range) {
    case 'thisWeek': {
      const day: number = today.getDay() || 7
      from = new Date(today)
      from.setDate(today.getDate() - day + 1)
      to = new Date(from)
      to.setDate(from.getDate() + 6)
      break
    }
    case 'lastWeek': {
      const day: number = today.getDay() || 7
      from = new Date(today)
      from.setDate(today.getDate() - day - 6)
      to = new Date(from)
      to.setDate(from.getDate() + 6)
      break
    }
    case 'thisMonth': {
      from = new Date(today.getFullYear(), today.getMonth(), 1)
      to = today
      break
    }
    case 'lastMonth': {
      from = new Date(today.getFullYear(), today.getMonth() - 1, 1)
      to = new Date(today.getFullYear(), today.getMonth(), 0)
      break
    }
    default:
      return
  }

  filters.value.dateFrom = formatDateISO(from)
  filters.value.dateTo = formatDateISO(to)
}

const formatDateISO = (date: Date): string => {
  return date.toISOString().split('T')[0]
}

const onProductChange = async (): Promise<void> => {
  filters.value.executionId = ''
  executions.value = []
  if (globalSelection.product) {
    try {
      const response = await getTimelogExecutions({
        projectId: ''
      })
      executions.value = response.data || []
    } catch (error) {
      console.error('获取执行列表失败:', error)
      ElMessage.error('获取执行列表失败')
    }
  } else {
    executions.value = []
  }
}

watch(() => globalSelection.product, (newProduct: string | undefined): void => {
  if (newProduct) {
    onProductChange()
  } else {
    executions.value = []
    filters.value.executionId = ''
  }
}, { immediate: true })

const queryTimelog = async (): Promise<void> => {
  if (!globalSelection.product) {
    ElMessage.warning('请先在顶部选择产品')
    return
  }

  if (!filters.value.dateFrom || !filters.value.dateTo) {
    ElMessage.warning('请选择时间范围')
    return
  }

  const cacheKey: string = `timelog_${globalSelection.product}_${filters.value.executionId}_${filters.value.assignedTo}_${filters.value.dateFrom}_${filters.value.dateTo}`
  
  const cachedData: string | null = localStorage.getItem(cacheKey)
  if (cachedData) {
    const parsedData: { data: TimelogAnalysis; expiry: number } = JSON.parse(cachedData)
    if (parsedData.expiry > Date.now()) {
      analysisData.value = parsedData.data
      showResult.value = true
      await nextTick()
      renderCharts()
      return
    }
  }

  loading.value = true

  try {
    const dashboardResponse = await getTimelogDashboard({
      productId: globalSelection.product ? parseInt(globalSelection.product, 10) : undefined,
      projectId: undefined,
      executionId: filters.value.executionId ? parseInt(filters.value.executionId, 10) : undefined,
      assignedTo: filters.value.assignedTo,
      dateFrom: filters.value.dateFrom,
      dateTo: filters.value.dateTo
    })

    const effortsResponse = await getTimelogEfforts({
      productId: globalSelection.product ? parseInt(globalSelection.product, 10) : undefined,
      projectId: undefined,
      executionId: filters.value.executionId ? parseInt(filters.value.executionId, 10) : undefined,
      assignedTo: filters.value.assignedTo,
      dateFrom: filters.value.dateFrom,
      dateTo: filters.value.dateTo
    })

    const fullData: TimelogAnalysis = {
      ...dashboardResponse.data,
      efforts: effortsResponse.data
    }

    analysisData.value = fullData
    showResult.value = true
    await nextTick()
    renderCharts()

    const cacheData: { data: TimelogAnalysis; expiry: number } = {
      data: fullData,
      expiry: Date.now() + 24 * 60 * 60 * 1000
    }
    localStorage.setItem(cacheKey, JSON.stringify(cacheData))
  } catch (error) {
    console.error('查询工时统计失败:', error)
    ElMessage.error('查询失败: ' + ((error as { response?: { data?: { error?: string } }; message?: string })?.response?.data?.error || (error as Error).message))
  } finally {
    loading.value = false
  }
}

const renderCharts = (): void => {
  renderDailyChart()
  renderProjectChart()
  renderTypeChart()
}

const renderDailyChart = (): void => {
  if (chartInstances.value.dailyChart) {
    chartInstances.value.dailyChart.destroy()
  }

  if (!dailyChart.value) return
  const ctx: CanvasRenderingContext2D = dailyChart.value.getContext('2d') as CanvasRenderingContext2D
  chartInstances.value.dailyChart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: analysisData.value.byDate.map((d: TimelogByDate): string => d.date),
      datasets: [{
        label: '工时 (小时)',
        data: analysisData.value.byDate.map((d: TimelogByDate): number => parseFloat(d.hours.toFixed(1))),
        backgroundColor: '#4e79a7',
        borderRadius: 3
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: (ctx): string => {
              const d: TimelogByDate = analysisData.value.byDate[ctx.dataIndex]
              return `${d.hours.toFixed(1)}h (${d.count}条记录)`
            }
          }
        }
      },
      scales: {
        y: { beginAtZero: true, title: { display: true, text: '小时' } },
        x: { ticks: { maxRotation: 45, minRotation: 0 } }
      }
    }
  })
}

const renderProjectChart = (): void => {
  if (chartInstances.value.projectChart) {
    chartInstances.value.projectChart.destroy()
  }

  if (!projectChart.value) return
  const items: TimelogByProject[] = [...analysisData.value.byProject].sort((a: TimelogByProject, b: TimelogByProject): number => b.hours - a.hours)
  const ctx: CanvasRenderingContext2D = projectChart.value.getContext('2d') as CanvasRenderingContext2D
  chartInstances.value.projectChart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: items.map((i: TimelogByProject): string => i.name),
      datasets: [{
        label: '工时 (小时)',
        data: items.map((i: TimelogByProject): number => parseFloat(i.hours.toFixed(1))),
        backgroundColor: items.map((_: TimelogByProject, idx: number): string => getColor(idx)),
        borderRadius: 4
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: (ctx): string => {
              const item: TimelogByProject = items[ctx.dataIndex]
              return `${ctx.parsed.y}h (${item.count}个任务)`
            }
          }
        }
      },
      scales: {
        y: { beginAtZero: true, title: { display: true, text: '小时' } },
        x: { ticks: { maxRotation: 45, minRotation: 0 } }
      }
    }
  })
}

const renderTypeChart = (): void => {
  if (chartInstances.value.typeChart) {
    chartInstances.value.typeChart.destroy()
  }

  if (!typeChart.value) return
  const items: TimelogByType[] = [...analysisData.value.byType].sort((a: TimelogByType, b: TimelogByType): number => b.hours - a.hours)
  const ctx: CanvasRenderingContext2D = typeChart.value.getContext('2d') as CanvasRenderingContext2D
  chartInstances.value.typeChart = new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: items.map((i: TimelogByType): string => i.name),
      datasets: [{
        data: items.map((i: TimelogByType): number => parseFloat(i.hours.toFixed(1))),
        backgroundColor: items.map((_: TimelogByType, idx: number): string => getColor(idx)),
        borderWidth: 2,
        borderColor: '#fff'
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'right',
          labels: { padding: 12, font: { size: 13 } }
        },
        tooltip: {
          callbacks: {
            label: (ctx): string => {
              const item: TimelogByType = items[ctx.dataIndex]
              const total: number = items.reduce((s: number, i: TimelogByType): number => s + i.hours, 0)
              const pct: string = total > 0 ? ((item.hours / total) * 100).toFixed(1) : '0'
              return `${item.name}: ${item.hours.toFixed(1)}h (${pct}%, ${item.count}个任务)`
            }
          }
        }
      }
    }
  })
}

const getColor = (index: number): string => {
  const colors: string[] = [
    '#4e79a7', '#f28e2b', '#e15759', '#76b7b2', '#59a14f',
    '#edc948', '#b07aa1', '#ff9da7', '#9c755f', '#bab0ac'
  ]
  return colors[index % colors.length]
}

const filterTable = (): void => {
}

const handleSortChange = (sort: SortParams): void => {
  const { prop, order } = sort
  if (!prop) return

  analysisData.value.efforts.sort((a: TimelogEffort, b: TimelogEffort): number => {
    let va: string | number = a[prop as keyof TimelogEffort] as string | number
    let vb: string | number = b[prop as keyof TimelogEffort] as string | number
    if (typeof va === 'string') va = va.toLowerCase()
    if (typeof vb === 'string') vb = vb.toLowerCase()
    if (va < vb) return order === 'ascending' ? -1 : 1
    if (va > vb) return order === 'ascending' ? 1 : -1
    return 0
  })
}

const getUserName = (account: string): string => {
  const user: User | undefined = users.value.find((u: User): boolean => u.account === account)
  return user ? (user.realname || account) : account
}

onMounted(async (): Promise<void> => {
  if (globalSelection.product) {
    onProductChange()
  }
  
  try {
    const userData = await getUsers()
    users.value = userData || []
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
  
  setQuickRange('thisMonth')
})
</script>

<style scoped>
.timelog-container {
  max-width: 1400px;
  margin: 0 auto;
}

.filters {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.quick-btns {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  gap: 10px;
}

.quick-btns span {
  font-weight: 500;
  color: #303133;
}

.quick-btns .el-button.active {
  background-color: #409EFF;
  color: #fff;
  border-color: #409EFF;
}

.filter-form {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  align-items: flex-end;
}

.filter-form .el-form-item {
  margin-bottom: 10px;
}

.summary-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
  margin-bottom: 20px;
}

.summary-card {
  text-align: center;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.card-value {
  font-size: 28px;
  font-weight: 700;
  color: #409EFF;
  margin-bottom: 6px;
}

.card-label {
  font-size: 13px;
  color: #606266;
}

.charts-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

.chart-card {
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.chart-card.full-width {
  grid-column: 1 / -1;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chart-wrapper {
  height: 300px;
  position: relative;
}

.data-table {
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.table-count {
  color: #909399;
  font-size: 14px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #909399;
}



/* 响应式设计 */
@media (max-width: 1200px) {
  .summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .charts-container {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .summary-cards {
    grid-template-columns: 1fr;
  }
  
  .filter-form {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .filter-form .el-form-item {
    margin-right: 0;
    margin-bottom: 10px;
  }
}
</style>