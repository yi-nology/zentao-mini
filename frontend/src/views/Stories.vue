<template>
  <div class="stories-page">
    <!-- 筛选器 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="指派人">
          <el-select
            v-model="filterForm.assignedTo"
            placeholder="请选择指派人"
            clearable
            filterable
            style="width: 150px"
          >
            <el-option
              v-for="item in userOptions"
              :key="item.account"
              :label="item.realname || item.account"
              :value="item.account"
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
        <span v-if="selectedStories.length > 0">已选择 {{ selectedStories.length }} 个需求</span>
        <div class="header-actions">
          <el-button 
            type="primary" 
            size="small" 
            @click="handleViewDetails" 
            :disabled="selectedStories.length === 0"
          >
            查看详情
          </el-button>
          <el-button 
            type="success" 
            size="small" 
            @click="handleExport" 
            :disabled="selectedStories.length === 0"
          >
            导出
          </el-button>
        </div>
      </div>
      <el-table
        v-loading="loading"
        :data="storyList"
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
            <a href="javascript:void(0)" @click="openZentaoLink(`https://pm.kylin.com/story-view-${row.id}.html`)" class="story-title" show-overflow-tooltip>
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
        <el-table-column prop="pri" label="优先级" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.pri)" effect="dark">
              {{ row.pri }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="stage" label="阶段" width="100" align="center">
          <template #default="{ row }">
            {{ getStageLabel(row.stage) }}
          </template>
        </el-table-column>
        <el-table-column prop="assignedTo" label="指派人" width="120" align="center">
          <template #default="{ row }">
            {{ row.assignedTo?.realname || row.assignedTo?.account || row.assignedTo || '-' }}
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
      title="需求详情"
      width="80%"
      destroy-on-close
    >
      <div v-if="currentStory" class="story-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="ID">{{ currentStory.id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ currentStory.title }}</el-descriptions-item>
          <el-descriptions-item label="产品">{{ (currentStory.product as unknown as { name?: string })?.name }}</el-descriptions-item>
          <el-descriptions-item label="项目">{{ (currentStory as unknown as { project?: { name?: string } }).project?.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ getStatusLabel(currentStory.status) }}</el-descriptions-item>
          <el-descriptions-item label="阶段">{{ getStageLabel(currentStory.stage) }}</el-descriptions-item>
          <el-descriptions-item label="优先级">{{ currentStory.pri }}</el-descriptions-item>
          <el-descriptions-item label="指派人">{{ (currentStory.assignedTo as unknown as { realname?: string; account?: string })?.realname || (currentStory.assignedTo as unknown as { realname?: string; account?: string })?.account || currentStory.assignedTo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentStory.openedDate }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            <div v-html="currentStory.spec"></div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, inject, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as XLSX from 'xlsx'
import {
  getStories,
  getUsers
} from '@/api/zentao'
import type { Story, User } from '@/types/api'

import * as runtime from '@wailsjs/runtime/runtime'

interface GlobalSelection {
  product: number | null
  project: number | null
  execution: number | null
}

interface FilterForm {
  assignedTo: string
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
  dateRange: [],
  specificDate: ''
})

const userOptions = ref<User[]>([])

const storyList = ref<Story[]>([])
const loading = ref<boolean>(false)
const selectedStories = ref<Story[]>([])

const detailDialogVisible = ref<boolean>(false)
const currentStory = ref<Story | null>(null)

const pagination = reactive<Pagination>({
  page: 1,
  pageSize: 20,
  total: 0
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

const fetchStories = async (): Promise<void> => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      product: globalSelection.product ?? undefined,
      project: globalSelection.project ?? undefined,
      assignedTo: filterForm.assignedTo,
      startDate: filterForm.dateRange[0] || '',
      endDate: filterForm.dateRange[1] || '',
      specificDate: filterForm.specificDate
    }
    const res = await getStories(params)
    const data = res.data
    storyList.value = Array.isArray(data) ? data : []
    pagination.total = Array.isArray(data) ? data.length : 0
  } catch (error) {
    console.error('获取需求列表失败:', error)
    ElMessage.error('获取需求列表失败')
  } finally {
    loading.value = false
  }
}

watch(() => globalSelection.product, (newProduct: number | null) => {
  if (newProduct) {
    pagination.page = 1
    storyList.value = []
    pagination.total = 0
    fetchStories()
  } else {
    storyList.value = []
    pagination.total = 0
  }
}, { immediate: true })

watch(() => globalSelection.project, () => {
  if (globalSelection.product) {
    pagination.page = 1
    fetchStories()
  }
})

const handleSearch = (): void => {
  if (!globalSelection.product && !globalSelection.project) {
    ElMessage.warning('请先在顶部选择产品或项目')
    return
  }
  pagination.page = 1
  fetchStories()
}

const handleReset = (): void => {
  filterForm.assignedTo = ''
  filterForm.dateRange = []
  filterForm.specificDate = ''
  pagination.page = 1
  storyList.value = []
  pagination.total = 0
}

const handleSizeChange = (size: number): void => {
  if (!globalSelection.product && !globalSelection.project) {
    return
  }
  pagination.pageSize = size
  pagination.page = 1
  fetchStories()
}

const handlePageChange = (page: number): void => {
  if (!globalSelection.product && !globalSelection.project) {
    return
  }
  pagination.page = page
  fetchStories()
}

const getStatusType = (status: string): string => {
  const types: Record<string, string> = {
    draft: 'info',
    active: 'success',
    changed: 'warning',
    closed: 'info'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status: string): string => {
  const labels: Record<string, string> = {
    draft: '草稿',
    active: '激活',
    changed: '已变更',
    closed: '已关闭'
  }
  return labels[status] || status
}

const getPriorityType = (pri: number): string => {
  if (pri === 1) return 'danger'
  if (pri === 2) return 'warning'
  if (pri === 3) return 'primary'
  return 'info'
}

const getStageLabel = (stage: string): string => {
  const labels: Record<string, string> = {
    wait: '等待',
    planned: '已计划',
    projected: '已立项',
    developing: '研发中',
    developed: '研发完毕',
    testing: '测试中',
    tested: '测试完毕',
    verified: '已验收',
    released: '已发布'
  }
  return labels[stage] || stage
}

const handleSelect = (selection: Story[], _row: Story): void => {
  selectedStories.value = selection
}

const handleSelectAll = (selection: Story[]): void => {
  selectedStories.value = selection
}

const handleViewDetails = (): void => {
  if (selectedStories.value.length > 0) {
    currentStory.value = selectedStories.value[0]
    detailDialogVisible.value = true
  }
}

const handleViewDetail = (row: Story): void => {
  currentStory.value = row
  detailDialogVisible.value = true
}

const handleExport = (): void => {
  if (selectedStories.value.length > 0) {
    const exportData = selectedStories.value.map((story: Story) => ({
      ID: story.id,
      标题: story.title,
      产品: (story.product as unknown as { name?: string })?.name || '',
      项目: (story as unknown as { project?: { name?: string } }).project?.name || '',
      状态: getStatusLabel(story.status),
      阶段: getStageLabel(story.stage),
      优先级: story.pri,
      指派人: (story.assignedTo as unknown as { realname?: string; account?: string })?.realname || 
              (story.assignedTo as unknown as { realname?: string; account?: string })?.account || 
              story.assignedTo as string || '',
      创建时间: story.openedDate || '',
      描述: story.spec || ''
    }))

    const worksheet = XLSX.utils.json_to_sheet(exportData)
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, '需求列表')

    try {
      const w = window as unknown as { runtime?: { BrowserOpenURL?: (url: string) => void } }
      if (w.runtime && w.runtime.BrowserOpenURL) {
        const wbout = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' })
        const blob = new Blob([wbout], { type: 'application/octet-stream' })
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `需求列表_${new Date().toISOString().slice(0, 10)}.xlsx`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(url)
      } else {
        XLSX.writeFile(workbook, `需求列表_${new Date().toISOString().slice(0, 10)}.xlsx`)
      }
      ElMessage.success(`导出 ${selectedStories.value.length} 个需求成功`)
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
.stories-page {
  padding: 0;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.table-card {
  min-height: 500px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e4e7ed;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.story-title {
  color: #409eff;
  text-decoration: none;
  cursor: pointer;
}

.story-title:hover {
  text-decoration: underline;
}
</style>
