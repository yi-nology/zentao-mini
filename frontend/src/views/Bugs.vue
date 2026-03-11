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
          <el-descriptions-item label="产品">{{ currentBug.product?.name }}</el-descriptions-item>
          <el-descriptions-item label="项目">{{ currentBug.project?.name }}</el-descriptions-item>
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

<script setup>
import { ref, reactive, onMounted, computed, inject, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as XLSX from 'xlsx'
import {
  getBugs,
  getBugStatusOptions,
  getUsers
} from '@/api/zentao'

// 导入Wails运行时
import * as runtime from '@wailsjs/runtime/runtime'

const globalSelection = inject('globalSelection')

// 筛选表单
const filterForm = reactive({
  assignedTo: '',
  status: '',
  dateRange: [],
  specificDate: ''
})

// 选项数据
const productOptions = ref([])
const statusOptions = ref(getBugStatusOptions())
const userOptions = ref([])

// 表格数据
const bugList = ref([])
const loading = ref(false)
const selectedBugs = ref([])

// 详情弹窗
const detailDialogVisible = ref(false)
const currentBug = ref(null)

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 从用户列表中获取指派人选项
const assignedToOptions = computed(() => {
  const assignees = new Map()
  userOptions.value.forEach(user => {
    if (user.account) {
      assignees.set(user.account, {
        value: user.account,
        label: user.realname || user.account
      })
    }
  })
  return Array.from(assignees.values()).sort((a, b) => a.label.localeCompare(b.label))
})

// 根据筛选条件过滤后的 Bug 列表
const filteredBugList = computed(() => {
  return bugList.value.filter(bug => {
    // 指派人筛选
    if (filterForm.assignedTo) {
      const assigned = bug.assignedTo
      if (!assigned) return false
      const account = typeof assigned === 'object' ? assigned.account : assigned
      const realname = typeof assigned === 'object' ? assigned.realname : assigned
      if (account !== filterForm.assignedTo && realname !== filterForm.assignedTo) {
        return false
      }
    }
    
    // 状态筛选
    if (filterForm.status) {
      if (bug.status !== filterForm.status) {
        return false
      }
    }
    
    return true
  })
})

// 获取用户列表
const fetchUsers = async () => {
  try {
    const res = await getUsers()
    if (res.data?.users) {
      userOptions.value = res.data.users
    }
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败，请刷新页面重试')
  }
}

// 获取 Bug 列表
const fetchBugs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      productID: globalSelection.product,
      projectID: globalSelection.project,
      status: filterForm.status,
      startDate: filterForm.dateRange[0] || '',
      endDate: filterForm.dateRange[1] || '',
      specificDate: filterForm.specificDate
    }
    const res = await getBugs(params)
    const data = res.data || []
    bugList.value = Array.isArray(data) ? data : (data.list || [])
    pagination.total = Array.isArray(data) ? data.length : (data.total || 0)
  } catch (error) {
    console.error('获取 Bug 列表失败:', error)
    ElMessage.error('获取 Bug 列表失败')
  } finally {
    loading.value = false
  }
}

// 查询
const handleSearch = () => {
  if (!globalSelection.product) {
    ElMessage.warning('请先在顶部选择产品')
    return
  }
  pagination.page = 1
  fetchBugs()
}

// 重置
const handleReset = () => {
  filterForm.assignedTo = ''
  filterForm.status = ''
  filterForm.dateRange = []
  filterForm.specificDate = ''
  pagination.page = 1
  bugList.value = []
  pagination.total = 0
}

// 分页大小变化
const handleSizeChange = (size) => {
  if (!globalSelection.product) {
    return
  }
  pagination.pageSize = size
  pagination.page = 1
  fetchBugs()
}

// 页码变化
const handlePageChange = (page) => {
  if (!globalSelection.product) {
    return
  }
  pagination.page = page
  fetchBugs()
}

// 监听全局选择器变化
watch(() => globalSelection.product, (newProduct) => {
  if (newProduct) {
    // 重置分页和数据
    pagination.page = 1
    bugList.value = []
    pagination.total = 0
    // 自动查询
    fetchBugs()
  } else {
    // 清空数据
    bugList.value = []
    pagination.total = 0
  }
}, { immediate: true })

// 监听项目变化
watch(() => globalSelection.project, (newProject) => {
  if (globalSelection.product) {
    pagination.page = 1
    fetchBugs()
  }
})

// 获取状态标签类型
const getStatusType = (status) => {
  const types = {
    active: 'danger',
    resolved: 'success',
    closed: 'info'
  }
  return types[status] || 'info'
}

// 获取状态标签文本
const getStatusLabel = (status) => {
  const labels = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭'
  }
  return labels[status] || status
}

// 获取严重程度标签类型
const getSeverityType = (severity) => {
  if (severity === 1) return 'danger'
  if (severity === 2) return 'warning'
  if (severity === 3) return 'primary'
  return 'info'
}

// 格式化日期
const formatDate = (dateStr) => {
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

// 计算停留时长（小时）
const calculateDuration = (openedDate) => {
  if (!openedDate) return '-' 
  const openTime = new Date(openedDate).getTime()
  const now = new Date().getTime()
  const durationHours = (now - openTime) / (1000 * 60 * 60)
  return durationHours.toFixed(1)
}

// 选择处理
const handleSelect = (selection, row) => {
  selectedBugs.value = selection
}

// 全选处理
const handleSelectAll = (selection) => {
  selectedBugs.value = selection
}

// 查看详情（批量）
const handleViewDetails = () => {
  if (selectedBugs.value.length > 0) {
    // 显示第一个选中的Bug详情
    currentBug.value = selectedBugs.value[0]
    detailDialogVisible.value = true
  }
}

// 查看详情（单行）
const handleViewDetail = (row) => {
  currentBug.value = row
  detailDialogVisible.value = true
}

// 导出
const handleExport = () => {
  if (selectedBugs.value.length > 0) {
    // 准备导出数据
    const exportData = selectedBugs.value.map(bug => ({
      ID: bug.id,
      标题: bug.title,
      链接地址: `https://pm.kylin.com/bug-view-${bug.id}.html`,
      产品: bug.product?.name || '',
      项目: bug.project?.name || '',
      状态: getStatusLabel(bug.status),
      严重程度: bug.severity,
      指派人: bug.assignedTo?.realname || bug.assignedTo?.account || '',
      创建时间: formatDate(bug.openedDate),
      描述: bug.steps || ''
    }))

    // 创建工作簿
    const worksheet = XLSX.utils.json_to_sheet(exportData)
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, 'Bug列表')

    // 导出文件
    try {
      // 检查是否在Wails环境中
      if (window.runtime && window.runtime.BrowserOpenURL) {
        // 在Wails环境中，使用Blob和a标签下载
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
        // 在浏览器环境中使用XLSX.writeFile
        XLSX.writeFile(workbook, `Bug列表_${new Date().toISOString().slice(0, 10)}.xlsx`)
      }
      ElMessage.success(`导出 ${selectedBugs.value.length} 个Bug成功`)
    } catch (error) {
      console.error('导出失败:', error)
      ElMessage.error('导出失败，请重试')
    }
  }
}

// 打开禅道链接（兼容浏览器和Wails环境）
const openZentaoLink = (url) => {
  try {
    // 检查是否在Wails环境中
    if (window.runtime && window.runtime.BrowserOpenURL) {
      // 在Wails环境中使用BrowserOpenURL
      runtime.BrowserOpenURL(url)
    } else {
      // 在浏览器环境中使用window.open
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  } catch (error) {
    console.error('打开链接失败:', error)
    // 降级到window.open
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
