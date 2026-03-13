<template>
  <div class="page-container">
    <!-- 筛选器 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="指派人">
          <el-select
            v-model="filterForm.assignedTo"
            placeholder="请选择或输入指派人"
            clearable
            filterable
            style="width: 180px"
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
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item label="具体日期">
          <el-date-picker
            v-model="filterForm.specificDate"
            type="date"
            placeholder="选择日期"
            style="width: 180px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <div class="table-header">
        <span v-if="selectedBugs.length > 0">已选择 {{ selectedBugs.length }} 个Bug</span>
        <div class="header-actions">
          <el-button 
            type="primary" 
            size="small" 
            @click="handleViewDetails" 
            :disabled="selectedBugs.length === 0"
          >
            查看详情
          </el-button>
          <el-button 
            type="success" 
            size="small" 
            @click="handleExport" 
            :disabled="selectedBugs.length === 0"
          >
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
            <a href="javascript:void(0)" @click="openZentaoLink(`https://pm.kylin.com/bug-view-${row.id}.html`)" class="bug-title" show-overflow-tooltip>
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
            <el-tag :type="getSeverityType(row.severity)" effect="dark">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="assignedTo" label="指派人" width="120" align="center">
          <template #default="{ row }">
            {{ row.assignedTo?.realname || row.assignedTo?.account || row.assignedTo || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="openedDate" label="创建时间" width="150" align="center">
          <template #default="{ row }">
            {{ formatDate(row.openedDate) }}
          </template>
        </el-table-column>
        <el-table-column label="停留时长" width="120" align="center">
          <template #default="{ row }">
            {{ calculateDuration(row.openedDate) }} 小时
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleViewDetail(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
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
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="Bug详情"
      width="80%"
      destroy-on-close
    >
      <div v-if="currentBug" class="bug-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="ID">{{ currentBug.id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ currentBug.title }}</el-descriptions-item>
          <el-descriptions-item label="产品">{{ (currentBug.product as unknown as { name?: string })?.name }}</el-descriptions-item>
          <el-descriptions-item label="项目">{{ (currentBug.project as unknown as { name?: string })?.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ getStatusLabel(currentBug.status) }}</el-descriptions-item>
          <el-descriptions-item label="严重程度">{{ currentBug.severity }}</el-descriptions-item>
          <el-descriptions-item label="指派人">{{ currentBug.assignedTo?.realname || currentBug.assignedTo?.account || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(currentBug.openedDate) }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            <div v-html="currentBug.steps"></div>
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
import * as XLSX from 'xlsx'
import {
  getBugs,
  getBugStatusOptions,
  getUsers
} from '@/api/zentao'
import type { Bug, User, SelectOption } from '@/types/api'

import * as runtime from '@wailsjs/runtime/runtime'

interface GlobalSelection {
  product: number | null
  project: number | null
  execution: number | null
}

interface FilterForm {
  assignedTo: string
  status: string
  dateRange: [string, string] | []
  specificDate: string
}

interface Pagination {
  page: number
  pageSize: number
  total: number
}

const globalSelection = inject<GlobalSelection>('globalSelection')!

const filterForm = reactive<FilterForm>({
  assignedTo: '',
  status: '',
  dateRange: [],
  specificDate: ''
})

const statusOptions = ref<SelectOption[]>(getBugStatusOptions())
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
      if (account !== filterForm.assignedTo && realname !== filterForm.assignedTo) {
        return false
      }
    }
    
    if (filterForm.status) {
      if (bug.status !== filterForm.status) {
        return false
      }
    }
    
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
      productID: globalSelection.product ?? undefined,
      projectID: globalSelection.project ?? undefined,
      status: filterForm.status,
      startDate: filterForm.dateRange[0] || '',
      endDate: filterForm.dateRange[1] || '',
      specificDate: filterForm.specificDate
    }
    const res = await getBugs(params)
    const data = res.data
    bugList.value = Array.isArray(data) ? data : []
    pagination.total = Array.isArray(data) ? data.length : 0
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
  fetchBugs()
}

const handleReset = (): void => {
  filterForm.assignedTo = ''
  filterForm.status = ''
  filterForm.dateRange = []
  filterForm.specificDate = ''
  pagination.page = 1
  bugList.value = []
  pagination.total = 0
}

const handleSizeChange = (size: number): void => {
  if (!globalSelection.product) {
    return
  }
  pagination.pageSize = size
  pagination.page = 1
  fetchBugs()
}

const handlePageChange = (page: number): void => {
  if (!globalSelection.product) {
    return
  }
  pagination.page = page
  fetchBugs()
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

const calculateDuration = (openedDate: string): string => {
  if (!openedDate) return '-' 
  const openTime = new Date(openedDate).getTime()
  const now = new Date().getTime()
  const durationHours = (now - openTime) / (1000 * 60 * 60)
  return durationHours.toFixed(1)
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

const handleExport = (): void => {
  if (selectedBugs.value.length > 0) {
    const exportData = selectedBugs.value.map((bug: Bug) => ({
      ID: bug.id,
      标题: bug.title,
      链接地址: `https://pm.kylin.com/bug-view-${bug.id}.html`,
      产品: (bug.product as unknown as { name?: string })?.name || '',
      项目: (bug.project as unknown as { name?: string })?.name || '',
      状态: getStatusLabel(bug.status),
      严重程度: bug.severity,
      指派人: bug.assignedTo?.realname || bug.assignedTo?.account || '',
      创建时间: formatDate(bug.openedDate),
      描述: bug.steps || ''
    }))

    const worksheet = XLSX.utils.json_to_sheet(exportData)
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, 'Bug列表')

    try {
      const w = window as unknown as { runtime?: { BrowserOpenURL?: (url: string) => void } }
      if (w.runtime && w.runtime.BrowserOpenURL) {
        const wbout = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' })
        const blob = new Blob([wbout], { type: 'application/octet-stream' })
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `Bug列表_${new Date().toISOString().slice(0, 10)}.xlsx`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(url)
      } else {
        XLSX.writeFile(workbook, `Bug列表_${new Date().toISOString().slice(0, 10)}.xlsx`)
      }
      ElMessage.success(`导出 ${selectedBugs.value.length} 个Bug成功`)
    } catch (error) {
      console.error('导出失败:', error)
      ElMessage.error('导出失败，请重试')
    }
  }
}

const openZentaoLink = (url: string): void => {
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
  fetchUsers()
})
</script>

<style scoped>
/* 页面特定样式 */
.page-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
  overflow: hidden;
}

.filter-card {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.filter-card:hover {
  box-shadow: 0 4px 16px 0 rgba(0, 0, 0, 0.12);
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: end;
  padding: 16px 0;
}

.table-card {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-card:hover {
  box-shadow: 0 4px 16px 0 rgba(0, 0, 0, 0.12);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f2f5;
  background-color: #fafafa;
}

.header-actions {
  display: flex;
  gap: 8px;
}

:deep(.el-table) {
  border-radius: 0;
  flex: 1;
  overflow: auto;
}

:deep(.el-table th) {
  background-color: #fafafa !important;
  font-weight: 600;
  color: #2c3e50;
}

:deep(.el-table tr:hover) {
  background-color: #f5f7fa !important;
}

:deep(.el-table__row) {
  transition: all 0.2s ease;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: 16px 20px;
  border-top: 1px solid #f0f2f5;
  background-color: #fafafa;
}

.bug-detail {
  line-height: 1.6;
  padding: 16px;
}

.bug-detail .el-descriptions__content {
  word-break: break-word;
  line-height: 1.8;
}

.bug-detail .el-descriptions__label {
  font-weight: 600;
  color: #2c3e50;
}

.bug-title {
  color: #409eff;
  text-decoration: none;
  cursor: pointer;
}

.bug-title:hover {
  text-decoration: underline;
}

/* 响应式调整 */
@media screen and (max-width: 768px) {
  .filter-form {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-form .el-form-item {
    margin-right: 0 !important;
  }
  
  .table-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: space-between;
  }
  
  .pagination-wrapper {
    justify-content: center;
  }
}
</style>
