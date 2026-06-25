<template>
  <div class="page-container">
    <div class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="指派人">
          <el-select
            v-model="filterForm.assignedTo"
            placeholder="请选择或输入指派人"
            clearable
            filterable
            style="width: 160px"
          >
            <el-option
              v-for="item in assignedToOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="filterForm.status"
            placeholder="请选择状态"
            clearable
            style="width: 120px"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select
            v-model="filterForm.type"
            placeholder="请选择类型"
            clearable
            filterable
            style="width: 140px"
          >
            <el-option
              v-for="item in typeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-card">
      <div class="table-header">
        <span v-if="selectedBugs.length > 0">已选择 {{ selectedBugs.length }} 个Bug</span>
        <div class="header-actions">
          <el-button type="primary" size="small" @click="handleViewDetails" :disabled="selectedBugs.length === 0">
            查看详情
          </el-button>
          <el-button type="success" size="small" @click="handleExport" :disabled="selectedBugs.length === 0">
            导出
          </el-button>
        </div>
      </div>
      <el-table
        v-loading="loading"
        :data="filteredBugList"
        border
        stripe
        style="width: 100%"
        @select="handleSelect"
        @select-all="handleSelectAll"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <a href="javascript:void(0)" @click="openZentaoLink(buildZentaoUrl(`bug-view-${row.id}.html`))" class="bug-title">
              {{ row.title }}
            </a>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="severity" label="严重程度" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)">
              {{ getSeverityLabel(row.severity) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.type" type="info" size="small">{{ getTypeLabel(row.type) }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="assignedTo" label="指派人" width="100" align="center">
          <template #default="{ row }">
            {{ row.assignedTo?.realname || row.assignedTo?.account || row.assignedTo || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="openedDate" label="创建时间" width="150" align="center">
          <template #default="{ row }">
            {{ formatDate(row.openedDate) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">
              查看
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <el-dialog v-model="detailDialogVisible" title="Bug详情" width="80%" destroy-on-close>
      <div v-if="currentBug" class="bug-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="ID">{{ currentBug.id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ currentBug.title }}</el-descriptions-item>
          <el-descriptions-item label="产品">{{ productMap[currentBug.product] || currentBug.product }}</el-descriptions-item>
          <el-descriptions-item label="项目">{{ currentBug.project }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ getStatusLabel(currentBug.status) }}</el-descriptions-item>
          <el-descriptions-item label="严重程度">{{ currentBug.severity }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentBug.type ? getTypeLabel(currentBug.type) : '-' }}</el-descriptions-item>
          <el-descriptions-item label="指派人">{{ currentBug.assignedTo?.realname || currentBug.assignedTo?.account || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(currentBug.openedDate) }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            <div v-html="sanitizeHtml(currentBug.steps)"></div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, inject, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { sanitizeHtml } from '@/utils/sanitize'
import { getBugs, getBugStatusOptions, getUsers, getProducts } from '@/api/zentao'
import { useZentaoConfig } from '@/composables/useZentaoConfig'
import type { Bug, User, SelectOption } from '@/types/api'
import * as runtime from '@wailsjs/runtime/runtime'
import { useRoute, useRouter } from 'vue-router'

interface GlobalSelection {
  product: number | null
  project: number | null
  execution: number | null
}

interface FilterForm {
  assignedTo: string
  status: string
  type: string
  dateRange: [string, string] | []
  specificDate: string
}

interface Pagination {
  page: number
  pageSize: number
  total: number
}

const globalSelection = inject<GlobalSelection>('globalSelection')!
const route = useRoute()
const router = useRouter()
const { buildUrl: buildZentaoUrl } = useZentaoConfig()

const filterForm = reactive<FilterForm>({
  assignedTo: '',
  status: '',
  type: '',
  dateRange: [],
  specificDate: ''
})

const statusOptions = ref<SelectOption[]>(getBugStatusOptions())
const productMap = ref<Record<number, string>>({})

const fetchProductNames = async (): Promise<void> => {
  try {
    const res = await getProducts()
    const products = res.data || []
    const map: Record<number, string> = {}
    products.forEach(p => { map[p.id] = p.name })
    productMap.value = map
  } catch { /* ignore */ }
}
const typeOptions = computed(() => {
  const types = new Map<string, { value: string; label: string }>()
  bugList.value.forEach((bug: Bug) => {
    if (bug.type && !types.has(bug.type)) {
      types.set(bug.type, {
        value: bug.type,
        label: getTypeLabel(bug.type)
      })
    }
  })
  return Array.from(types.values()).sort((a, b) => a.label.localeCompare(b.label))
})
const userOptions = ref<User[]>([])
const bugList = ref<Bug[]>([])
const loading = ref<boolean>(false)
const selectedBugs = ref<Bug[]>([])
const detailDialogVisible = ref<boolean>(false)
const currentBug = ref<Bug | null>(null)

const pagination = reactive<Pagination>({
  page: 1,
  pageSize: 20,
  total: 0
})

const assignedToOptions = computed(() => {
  const assignees = new Map<string, { value: string; label: string }>()
  userOptions.value.forEach((user: User) => {
    if (user.account) {
      assignees.set(user.account, {
        value: user.account,
        label: user.realname || user.account
      })
    }
  })
  return Array.from(assignees.values()).sort((a, b) => a.label.localeCompare(b.label))
})

const filteredBugList = computed(() => {
  return bugList.value.filter((bug: Bug) => {
    if (filterForm.assignedTo) {
      const assigned = bug.assignedTo
      if (!assigned) return false
      const account = typeof assigned === 'object' ? assigned.account : assigned
      const realname = typeof assigned === 'object' ? assigned.realname : assigned
      if (account !== filterForm.assignedTo && realname !== filterForm.assignedTo) return false
    }
    if (filterForm.status && bug.status !== filterForm.status) return false
    if (filterForm.type && bug.type !== filterForm.type) return false
    return true
  })
})

const fetchUsers = async (): Promise<void> => {
  try {
    const users = await getUsers()
    userOptions.value = users || []
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败，请刷新页面重试')
  }
}

const fetchBugs = async (): Promise<void> => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      productId: globalSelection.product ?? undefined,
      projectId: globalSelection.project ?? undefined,
      status: filterForm.status,
      startDate: filterForm.dateRange[0] || '',
      endDate: filterForm.dateRange[1] || '',
      specificDate: filterForm.specificDate
    }
    const res = await getBugs(params)
    const paginatedData = res.data
    bugList.value = paginatedData.list || []
    pagination.total = paginatedData.total || 0
  } catch (error) {
    console.error('获取 Bug 列表失败:', error)
    ElMessage.error('获取 Bug 列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = (): void => {
  if (!globalSelection.product) {
    ElMessage.warning('请先在顶部选择产品')
    return
  }
  pagination.page = 1
  syncRoute()
  fetchBugs()
}

const handleReset = (): void => {
  filterForm.assignedTo = ''
  filterForm.status = ''
  filterForm.type = ''
  filterForm.dateRange = []
  filterForm.specificDate = ''
  pagination.page = 1
  syncRoute()
  fetchBugs()
}

const handleSizeChange = (size: number): void => {
  if (!globalSelection.product) return
  pagination.pageSize = size
  pagination.page = 1
  syncRoute()
  fetchBugs()
}

const handlePageChange = (page: number): void => {
  if (!globalSelection.product) return
  pagination.page = page
  syncRoute()
  fetchBugs()
}

const syncRoute = (): void => {
  const q: Record<string, string> = {}
  if (filterForm.assignedTo) q.assignedTo = filterForm.assignedTo
  if (filterForm.status) q.status = filterForm.status
  if (filterForm.type) q.type = filterForm.type
  if (filterForm.dateRange[0]) q.startDate = filterForm.dateRange[0]
  if (filterForm.dateRange[1]) q.endDate = filterForm.dateRange[1]
  if (filterForm.specificDate) q.specificDate = filterForm.specificDate
  if (pagination.page > 1) q.page = String(pagination.page)
  if (pagination.pageSize !== 20) q.pageSize = String(pagination.pageSize)
  router.replace({ query: q })
}

watch(() => globalSelection.product, (newProduct: number | null) => {
  if (newProduct) {
    pagination.page = 1
    bugList.value = []
    pagination.total = 0
    fetchBugs()
  } else {
    bugList.value = []
    pagination.total = 0
  }
}, { immediate: true })

watch(() => globalSelection.project, () => {
  if (globalSelection.product) {
    pagination.page = 1
    fetchBugs()
  }
})

const getStatusType = (status: string): string => {
  const types: Record<string, string> = {
    active: 'danger',
    resolved: 'success',
    closed: 'info'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status: string): string => {
  const labels: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭'
  }
  return labels[status] || status
}

const getSeverityType = (severity: number): string => {
  if (severity === 1) return 'danger'
  if (severity === 2) return 'warning'
  if (severity === 3) return 'primary'
  return 'info'
}

const getSeverityLabel = (severity: number): string => {
  const labels: Record<number, string> = { 1: '致命', 2: '严重', 3: '一般', 4: '轻微', 5: '建议' }
  return labels[severity] || String(severity)
}

const getTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    codeerror: '代码错误',
    configerror: '配置错误',
    security: '安全问题',
    performance: '性能问题',
    standard: '标准规范',
    designdefect: '设计缺陷',
    ui: '界面问题',
    install: '安装部署',
    automation: '自动化',
    other: '其他'
  }
  return labels[type] || type
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const handleSelect = (selection: Bug[], _row: Bug): void => {
  selectedBugs.value = selection
}

const handleSelectAll = (selection: Bug[]): void => {
  selectedBugs.value = selection
}

const handleViewDetails = (): void => {
  if (selectedBugs.value.length > 0) {
    currentBug.value = selectedBugs.value[0]
    detailDialogVisible.value = true
  }
}

const handleViewDetail = (row: Bug): void => {
  currentBug.value = row
  detailDialogVisible.value = true
}

const handleExport = async (): Promise<void> => {
  if (selectedBugs.value.length === 0) return
  const XLSX = await import('xlsx')
  const exportData = selectedBugs.value.map((bug: Bug) => ({
      ID: bug.id,
      标题: bug.title,
      链接地址: buildZentaoUrl(`bug-view-${bug.id}.html`),
      产品: productMap.value[bug.product] || String(bug.product),
      项目: String(bug.project),
      状态: getStatusLabel(bug.status),
      严重程度: getSeverityLabel(bug.severity),
      类型: getTypeLabel(bug.type),
      指派人: bug.assignedTo?.realname || bug.assignedTo?.account || '',
      创建时间: formatDate(bug.openedDate),
      描述: bug.steps || ''
    }))

    const worksheet = XLSX.utils.json_to_sheet(exportData)
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, 'Bug列表')

    try {
      XLSX.writeFile(workbook, `Bug列表_${new Date().toISOString().slice(0, 10)}.xlsx`)
      ElMessage.success(`导出 ${selectedBugs.value.length} 个Bug成功`)
    } catch (error) {
      console.error('导出失败:', error)
      ElMessage.error('导出失败，请重试')
    }
}

const openZentaoLink = (url: string): void => {
  if (!url) {
    ElMessage.warning('禅道地址未配置，请检查系统设置')
    return
  }
  try {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (url: string) => void } }
    if (w.runtime && w.runtime.BrowserOpenURL) {
      runtime.BrowserOpenURL(url)
    } else {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  } catch (error) {
    console.error('打开链接失败:', error)
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}

onMounted(() => {
  const q = route.query
  if (q.assignedTo) filterForm.assignedTo = String(q.assignedTo)
  if (q.status) filterForm.status = String(q.status)
  if (q.type) filterForm.type = String(q.type)
  if (q.startDate || q.endDate) filterForm.dateRange = [String(q.startDate || ''), String(q.endDate || '')] as [string, string]
  if (q.specificDate) filterForm.specificDate = String(q.specificDate)
  if (q.page) pagination.page = Number(q.page) || 1
  if (q.pageSize) pagination.pageSize = Number(q.pageSize) || 20

  fetchUsers()
  fetchProductNames()
})
</script>

<style scoped>
.bug-title {
  color: var(--color-primary);
  text-decoration: none;
  cursor: pointer;
  transition: color var(--transition-fast);
}

.bug-title:hover {
  text-decoration: underline;
  color: var(--color-primary-hover);
}

.bug-detail {
  line-height: 1.6;
  padding: 8px;
}

.bug-detail :deep(.el-descriptions__content) {
  word-break: break-word;
  line-height: 1.8;
}

.bug-detail :deep(.el-descriptions__label) {
  font-weight: 600;
  color: var(--color-text-primary);
}
</style>
